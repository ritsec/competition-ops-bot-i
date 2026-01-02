package bot

import (
	"log"

	"github.com/bwmarrin/discordgo"
	"github.com/ritsec/competition-ops-bot-i/ent/role"
)

var (
	// TODO: Find a better way to specify this. Env variables the best quick option.
	// For the future, COBI should handle server setup as well.
	NUM_BLUE_TEAMS = 18
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
							Name:  "All",
							Value: "All",
						},
					},
				},
			},
		},
		func(s *discordgo.Session, i *discordgo.InteractionCreate) {
			choice := i.ApplicationCommandData().Options[0].StringValue()

			initialMessage(s, i, "Refreshing server role data...")

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
