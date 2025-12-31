package bot

import (
	"log"

	"github.com/bwmarrin/discordgo"
)

func (b *Bot) SSH() (*discordgo.ApplicationCommand, func(s *discordgo.Session, i *discordgo.InteractionCreate)) {
	return &discordgo.ApplicationCommand{
			Name:                     "ssh",
			Description:              "Register an SSH key for Black Team",
			DefaultMemberPermissions: &BlackTeam,
			Options: []*discordgo.ApplicationCommandOption{
				{
					Name:        "add",
					Description: "Add SSH key",
					Type:        discordgo.ApplicationCommandOptionSubCommandGroup,
					Options: []*discordgo.ApplicationCommandOption{
						{
							Type:        discordgo.ApplicationCommandOptionSubCommand,
							Name:        "key",
							Description: "SSH Key input",
							Options: []*discordgo.ApplicationCommandOption{
								{
									Type:        discordgo.ApplicationCommandOptionString,
									Name:        "value",
									Description: "SSH Key",
									Required:    true,
								},
							},
						},
					},
				},
			},
		},
		func(s *discordgo.Session, i *discordgo.InteractionCreate) {
			// Get public key from SSH key add subcommand
			key := i.ApplicationCommandData().Options[0].Options[0].Options[0].StringValue()
			log.Println(key)

			// Get Ent user object
			uid := i.Member.User.ID
			u, err := b.getUser(uid)
			if err != nil {
				log.Fatal(err)
			}

			// Add key to user's 'key' field
			if err := b.addKey(u, key); err != nil {
				log.Fatal(err)
			}
		}
}
