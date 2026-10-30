package slash

import (
	"context"
	"errors"
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/bwmarrin/discordgo"
	"github.com/ritsec/competition-ops-bot-i/ent"
	"github.com/ritsec/competition-ops-bot-i/ent/team"
	"github.com/ritsec/competition-ops-bot-i/internal/config"
	"github.com/ritsec/competition-ops-bot-i/internal/database"
	"github.com/ritsec/competition-ops-bot-i/internal/logging"
	"github.com/ritsec/competition-ops-bot-i/internal/utils"
)

// blueChannelRegex matches Blue Team chat channels, capturing the team number
var blueChannelRegex = regexp.MustCompile(`^blue-(\d+)-chat$`)

func Refresh(db *ent.Client) (*discordgo.ApplicationCommand, func(s *discordgo.Session, i *discordgo.InteractionCreate)) {
	return &discordgo.ApplicationCommand{
			Name:                     "refresh",
			Description:              "Refresh COBI data",
			DefaultMemberPermissions: &utils.Admin,
			Options: []*discordgo.ApplicationCommandOption{
				{
					Type:        discordgo.ApplicationCommandOptionString,
					Name:        "option",
					Description: "Option of what to refresh",
					Required:    true,
					Choices: []*discordgo.ApplicationCommandOptionChoice{
						{
							Name:  "Roles",
							Value: "Roles",
						},
						{
							Name:  "Teams",
							Value: "Teams",
						},
						{
							Name:  "Channels",
							Value: "Channels",
						},
						{
							Name:  "All",
							Value: "All",
						},
					},
				},
			},
		},
		func(s *discordgo.Session, i *discordgo.InteractionCreate) {
			ctx, span := logging.StartSpan(
				context.Background(),
				"commands.slash.refresh:Refresh",
			)
			defer span.End()

			choice := i.ApplicationCommandData().Options[0].StringValue()
			all := choice == "All"

			if !startStatus(ctx, s, i, "Refreshing server data...") {
				return
			}

			// Run each requested step, continuing past failures so one missing
			// role doesn't block the rest of the refresh
			var failures []string
			run := func(name string, fn func() error) {
				setStatus(ctx, s, i, fmt.Sprintf("Refreshing %s...", name))
				if err := fn(); err != nil {
					logging.Error(ctx, fmt.Sprintf("Could not refresh %s: %v", name, err))
					failures = append(failures, fmt.Sprintf("- %s: %v", name, err))
				}
			}

			// Order matters: teams need roles, and channels need teams
			if all || choice == "Roles" {
				run("roles", func() error { return refreshRoles(ctx, s, db) })
			}
			if all || choice == "Teams" {
				run("teams", func() error { return createTeams(ctx, db) })
			}
			if all || choice == "Channels" {
				run("channels", func() error { return initChannels(ctx, s, db) })
			}

			if len(failures) > 0 {
				setStatus(ctx, s, i, "Finished refreshing with errors:\n"+strings.Join(failures, "\n"))
				return
			}
			setStatus(ctx, s, i, "Successfully refreshed server data!")
		}
}

// refreshRoles stores every role in the Guild, updating names of roles that
// already exist
func refreshRoles(ctx context.Context, s *discordgo.Session, db *ent.Client) error {
	roles, err := s.GuildRoles(config.GuildID)
	if err != nil {
		return err
	}

	for _, r := range roles {
		existing, err := db.Role.Get(ctx, r.ID)
		switch {
		case ent.IsNotFound(err): // Create Role if it doesn't exist
			err = db.Role.
				Create().
				SetID(r.ID).
				SetName(r.Name).
				Exec(ctx)
		case err != nil:
			return err
		case existing.Name != r.Name: // Role was renamed in Discord
			err = db.Role.UpdateOne(existing).SetName(r.Name).Exec(ctx)
		}
		if err != nil {
			return fmt.Errorf("role %q: %w", r.Name, err)
		}
	}

	return nil
}

// createTeams creates any missing Blue, Red, and Black teams and links them
// to their roles
func createTeams(ctx context.Context, db *ent.Client) error {
	var errs []error

	// Blue Teams
	for n := 1; n <= config.NumBlueTeams; n++ {
		if _, err := database.GetBlue(ctx, db, n); err != nil {
			errs = append(errs, fmt.Errorf("blue team %d: %w", n, err))
		}
	}

	// Red Team
	if _, err := database.GetRed(ctx, db); err != nil {
		errs = append(errs, fmt.Errorf("red team: %w", err))
	}

	// Black Team
	if _, err := database.GetBlack(ctx, db); err != nil {
		errs = append(errs, fmt.Errorf("black team: %w", err))
	}

	return errors.Join(errs...)
}

// initChannels will parse the Blue Team channels of the Guild and
// add the channel as a relation to the corresponding team in SQL.
func initChannels(ctx context.Context, s *discordgo.Session, db *ent.Client) error {
	channels, err := s.GuildChannels(config.GuildID)
	if err != nil {
		return err
	}

	for _, channel := range channels {
		match := blueChannelRegex.FindStringSubmatch(channel.Name)
		if match == nil {
			continue
		}

		// Get team number
		num, err := strconv.Atoi(match[1])
		if err != nil {
			return err
		}

		// Get corresponding Blue Team
		t, err := db.Team.Query().
			Where(team.TypeEQ(team.TypeBlue), team.Number(num)).
			Only(ctx)
		if ent.IsNotFound(err) {
			logging.Warn(ctx, fmt.Sprintf("Could not find Blue Team %d, skipping...", num))
			continue
		}
		if err != nil {
			return err
		}

		exists, err := t.QueryChannel().Exist(ctx)
		if err != nil {
			return err
		}
		if exists {
			continue // TODO: Update value
		}

		// Create channel entry
		c, err := db.Channel.Create().
			SetID(channel.ID).
			SetName(channel.Name).
			Save(ctx)
		if err != nil {
			return err
		}

		// Add channel to team
		if err := db.Team.UpdateOne(t).AddChannel(c).Exec(ctx); err != nil {
			return err
		}
	}

	return nil
}
