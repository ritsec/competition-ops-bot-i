package bot

import (
	"github.com/bwmarrin/discordgo"
	"github.com/ritsec/competition-ops-bot-i/ent/role"
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
						}, // TODO: Option to remove entry
					},
				},
			},
		},
		func(s *discordgo.Session, i *discordgo.InteractionCreate) {
			choice := i.ApplicationCommandData().Options[0].StringValue()

			initialMessage(s, i, "Refreshing server role data...")
			if choice == "Roles" {
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
			}
			updateMessage(s, i, "Successfully refreshed server role data!")
		}
}
