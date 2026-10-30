package slash

import (
	"context"
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/bwmarrin/discordgo"
	"github.com/ritsec/competition-ops-bot-i/ent"
	"github.com/ritsec/competition-ops-bot-i/ent/schema"
	"github.com/ritsec/competition-ops-bot-i/ent/team"
	"github.com/ritsec/competition-ops-bot-i/internal/database"
	"github.com/ritsec/competition-ops-bot-i/internal/logging"
	"github.com/ritsec/competition-ops-bot-i/internal/utils"
)

// Placeholders in the service table that are filled in per team
const (
	teamNumberPlaceholder   = "##"              // e.g. "team##" becomes "team01"
	teamPasswordPlaceholder = "<TEAM PASSWORD>" // replaced by the team's password
)

// Column positions in the credentials sheet. The sheet holds two tables side
// by side sharing one header row: teams in A-B, and services in D-F.
const (
	colTeam            = 0 // "team01"
	colTeamPassword    = 1
	colServiceName     = 3 // "Router", "vpn.irsec.club", ...
	colServiceUsername = 4
	colServicePassword = 5
)

// teamIDRegex matches the team column, capturing the number as written
var teamIDRegex = regexp.MustCompile(`(?i)^team(\d+)$`)

// teamCred is one row of the team table
type teamCred struct {
	ID       string // Team number as written in the sheet, e.g. "01"
	Number   int
	Password string
}

func Creds(db *ent.Client) (*discordgo.ApplicationCommand, func(s *discordgo.Session, i *discordgo.InteractionCreate)) {
	return &discordgo.ApplicationCommand{
			Name:                     "creds",
			Description:              "Submit a CSV of credentials",
			DefaultMemberPermissions: &utils.Admin,
			Options: []*discordgo.ApplicationCommandOption{
				// TODO: Add option to query credentials
				{
					Type:        discordgo.ApplicationCommandOptionAttachment,
					Name:        "process",
					Description: "Process credentials CSV.",
				},
				{
					Type:        discordgo.ApplicationCommandOptionString,
					Name:        "send",
					Description: "Send credentials to team channels.",
					Choices: []*discordgo.ApplicationCommandOptionChoice{
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
				"commands.slash.creds:Creds",
			)
			defer span.End()

			// Both options are optional, so the user may have picked neither
			options := i.ApplicationCommandData().Options
			if len(options) == 0 {
				startStatus(ctx, s, i, "Please choose either process or send.")
				return
			}

			switch options[0].Name {
			case "process":
				processCreds(ctx, s, i, db, options[0])
			case "send":
				if !utils.ValidateAction(s, i) {
					return
				}

				if err := sendCreds(ctx, s, db); err != nil {
					logging.Error(ctx, fmt.Sprintf("Could not send credentials: %v", err))
				}
			}
		}
}

// processCreds downloads the attached CSV and stores its credentials
func processCreds(ctx context.Context, s *discordgo.Session, i *discordgo.InteractionCreate, db *ent.Client, option *discordgo.ApplicationCommandInteractionDataOption) {
	// Get file URL
	fileID := option.Value.(string)
	fileURL := i.ApplicationCommandData().Resolved.Attachments[fileID].URL
	logging.Info(ctx, fmt.Sprintf("Received %s", fileURL))

	if !startStatus(ctx, s, i, "Downloading and parsing file...") {
		return
	}

	rows, err := utils.ReadCSV(fileURL)
	if err != nil {
		reportError(ctx, s, i, "Could not download CSV", err)
		return
	}

	teams, services, err := parseCredsSheet(rows)
	if err != nil {
		reportError(ctx, s, i, "Could not parse CSV", err)
		return
	}

	setStatus(ctx, s, i, "Adding credentials to database...")
	if err := handleCreds(ctx, db, teams, services); err != nil {
		reportError(ctx, s, i, "Could not add credentials", err)
		return
	}

	setStatus(ctx, s, i, fmt.Sprintf("Successfully added credentials for %d teams and %d services!", len(teams), len(services)))
}

// parseCredsSheet reads the team and service tables from the credentials
// sheet. The first row is the header and is skipped.
func parseCredsSheet(rows [][]string) ([]teamCred, []schema.ServiceCredential, error) {
	if len(rows) < 2 {
		return nil, nil, fmt.Errorf("sheet has no data rows")
	}

	// Rows can be shorter than the header when trailing cells are empty
	cell := func(row []string, col int) string {
		if col < len(row) {
			return row[col]
		}
		return ""
	}

	var teams []teamCred
	var services []schema.ServiceCredential

	for idx, row := range rows[1:] {
		rowNum := idx + 2 // 1-based, after the header

		// Team table
		if id := strings.TrimSpace(cell(row, colTeam)); id != "" {
			match := teamIDRegex.FindStringSubmatch(id)
			if match == nil {
				return nil, nil, fmt.Errorf("row %d: team %q should look like team01", rowNum, id)
			}
			num, err := strconv.Atoi(match[1])
			if err != nil {
				return nil, nil, fmt.Errorf("row %d: invalid team number %q: %w", rowNum, match[1], err)
			}

			password := cell(row, colTeamPassword)
			if password == "" {
				return nil, nil, fmt.Errorf("row %d: %s has no password", rowNum, id)
			}

			teams = append(teams, teamCred{ID: match[1], Number: num, Password: password})
		}

		// Service table
		if name := strings.TrimSpace(cell(row, colServiceName)); name != "" {
			services = append(services, schema.ServiceCredential{
				Name:     name,
				Username: strings.TrimSpace(cell(row, colServiceUsername)),
				Password: cell(row, colServicePassword),
			})
		}
	}

	if len(teams) == 0 {
		return nil, nil, fmt.Errorf("no teams found in column A")
	}
	if len(services) == 0 {
		return nil, nil, fmt.Errorf("no services found in column D")
	}

	return teams, services, nil
}

// servicesFor returns the service logins with the team's number and password
// filled in
func (t teamCred) servicesFor(services []schema.ServiceCredential) []schema.ServiceCredential {
	replacer := strings.NewReplacer(
		teamNumberPlaceholder, t.ID,
		teamPasswordPlaceholder, t.Password,
	)

	filled := make([]schema.ServiceCredential, len(services))
	for i, service := range services {
		filled[i] = schema.ServiceCredential{
			Name:     service.Name,
			Username: replacer.Replace(service.Username),
			Password: replacer.Replace(service.Password),
		}
	}
	return filled
}

// handleCreds adds or updates the credentials for each team. Teams are
// written in a single transaction so a bad row doesn't leave a partial import.
func handleCreds(ctx context.Context, db *ent.Client, teams []teamCred, services []schema.ServiceCredential) error {
	return database.WithTx(ctx, db, func(tx *ent.Client) error {
		for _, t := range teams {
			if err := upsertCred(ctx, tx, t, t.servicesFor(services)); err != nil {
				return err
			}
		}
		return nil
	})
}

// upsertCred creates or updates the credential for a Blue team
func upsertCred(ctx context.Context, db *ent.Client, tc teamCred, services []schema.ServiceCredential) error {
	t, err := db.Team.Query().
		Where(team.TypeEQ(team.TypeBlue), team.Number(tc.Number)).
		Only(ctx)
	if ent.IsNotFound(err) {
		return fmt.Errorf("Blue Team %d doesn't exist yet, run /refresh Teams first", tc.Number)
	}
	if err != nil {
		return err
	}

	// Check if credentials entry exists
	c, err := t.QueryCredential().Only(ctx)
	switch {
	case ent.IsNotFound(err):
		return db.Credential.
			Create().
			SetTeam(t).
			SetPassword(tc.Password).
			SetServices(services).
			Exec(ctx)
	case err != nil:
		return err
	default:
		return db.Credential.UpdateOne(c).
			SetPassword(tc.Password).
			SetServices(services).
			Exec(ctx)
	}
}

// sendCreds sends each Blue team its credentials in its team channel
func sendCreds(ctx context.Context, s *discordgo.Session, db *ent.Client) error {
	teams, err := db.Team.Query().
		Where(team.TypeEQ(team.TypeBlue)).
		All(ctx)
	if err != nil {
		return err
	}

	for _, t := range teams {
		// Get team's credential
		creds, err := t.QueryCredential().Only(ctx)
		if err != nil {
			logging.Warn(ctx, fmt.Sprintf("Could not get credentials for Blue Team %d: %v", t.Number, err))
			continue
		}

		// Create message embed
		embed := utils.MessageCreds(t.Number, creds)

		// Get team's channel
		channel, err := t.QueryChannel().Only(ctx)
		if err != nil {
			logging.Warn(ctx, fmt.Sprintf("Could not get channel for Blue Team %d: %v", t.Number, err))
			continue
		}

		// Send embed to channel ID
		if _, err := s.ChannelMessageSendEmbed(channel.ID, embed); err != nil {
			logging.Warn(ctx, fmt.Sprintf("Could not send credentials to Blue Team %d: %v", t.Number, err))
		}
	}

	return nil
}
