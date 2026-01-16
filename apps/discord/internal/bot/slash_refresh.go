package bot

import (
	"log"
	"regexp"
	"strconv"

	"github.com/bwmarrin/discordgo"
	"github.com/ritsec/competition-ops-bot-i/ent/role"
	"github.com/ritsec/competition-ops-bot-i/ent/team"
)

func (b *Bot) Refresh() (*discordgo.ApplicationCommand, func(s *discordgo.Session, i *discordgo.InteractionCreate)) {
	return &discordgo.ApplicationCommand{
			Name:                     "refresh",
			Description:              "Refresh COBI data",
			DefaultMemberPermissions: &Admin,
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
			choice := i.ApplicationCommandData().Options[0].StringValue()

			initialMessage(s, i, "Refreshing server data...")

			switch choice {
			case "Roles":
				roles, err := s.GuildRoles(guildID)
				if err != nil {
					panic(err)
				}
				for _, roleObj := range roles {
					// Check if Role already exists
					_, err := b.Client.Role.
						Query().
						Where(role.Name(roleObj.Name)).
						Only(b.ClientCtx)
					if err != nil { // Create Role if it doesn't exist
						_, err = b.Client.Role.
							Create().
							SetID(roleObj.ID).
							SetName(roleObj.Name).
							Save(b.ClientCtx)
						if err != nil {
							panic(err)
						}
					} // TODO: Update role
				}
				updateMessage(s, i, "Successfully refreshed server role data!")
			case "Channels":
				if err := b.initChannels(); err != nil {
					log.Fatal(err)
				}
				updateMessage(s, i, "Successfully refreshed server channel data!")
			case "Teams":
				if err := b.createTeams(); err != nil {
					log.Fatal(err)
				}
				updateMessage(s, i, "Successfully refreshed server team data!")
			}
		}
}

func (b *Bot) createTeams() error {
	// Blue Teams
	for i := 1; i <= NUM_BLUE_TEAMS; i++ {
		// Create/check for team
		_, err := b.getBlue(i)
		if err != nil {
			return err
		}
	}

	// Red Team
	_, err := b.getRed()
	if err != nil {
		return err
	}

	// Black Team
	_, err = b.getBlack()
	if err != nil {
		return err
	}

	return nil
}

// initChannels will parse the Blue Team channels of the Guild and
// add the channel as a relation to the corresponding team in SQL.
func (b *Bot) initChannels() error {
	channels, err := b.Session.GuildChannels(guildID)
	if err != nil {
		return err
	}

	blueRegex := regexp.MustCompile(`^team([1-9]|1[0-9]|2[01])-chat$`)

	for _, channel := range channels {
		match := blueRegex.FindStringSubmatch(channel.Name)

		if match == nil {
			continue
		}

		// Get team number
		num, err := strconv.Atoi(match[1])
		if err != nil {
			return err
		}

		// Get corresponding Blue Team
		t, err := b.Client.Team.Query().
			Where(team.And(team.TypeEQ(team.TypeBlue), team.Number(num))).
			Only(b.ClientCtx)
		if err != nil {
			log.Printf("Could not find team %d, skipping...", num)
			continue
		}

		if t.QueryChannel() != nil {
			continue // TODO: Update value
		}

		// Create channel entry
		c, err := b.Client.Channel.Create().
			SetID(channel.ID).
			SetName(channel.Name).
			Save(b.ClientCtx)
		if err != nil {
			return err
		}

		// Add channel to team
		_, err = t.Update().AddChannel(c).Save(b.ClientCtx)
		if err != nil {
			return err
		}
	}

	return nil
}
