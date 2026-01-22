package bot

import (
	"github.com/bwmarrin/discordgo"
	"golang.org/x/crypto/ssh"
)

func (b *Bot) SSH() (*discordgo.ApplicationCommand, func(s *discordgo.Session, i *discordgo.InteractionCreate)) {
	return &discordgo.ApplicationCommand{
			Name:                     "ssh",
			Description:              "Register an SSH key for Black Team",
			DefaultMemberPermissions: &BlackTeam,
			Options: []*discordgo.ApplicationCommandOption{
				{
					Name:        "key",
					Description: "Add SSH key",
					Type:        discordgo.ApplicationCommandOptionSubCommandGroup,
					Options: []*discordgo.ApplicationCommandOption{
						{
							Type:        discordgo.ApplicationCommandOptionSubCommand,
							Name:        "add",
							Description: "SSH Key input",
							Options: []*discordgo.ApplicationCommandOption{
								{
									Type:        discordgo.ApplicationCommandOptionString,
									Name:        "value",
									Description: "SSH public key",
									Required:    true,
								},
								{
									Type:        discordgo.ApplicationCommandOptionString,
									Name:        "name",
									Description: "SSH key name",
									Required:    true,
								},
							},
						},
						{
							Type:        discordgo.ApplicationCommandOptionSubCommand,
							Name:        "remove",
							Description: "Remove SSH key entry",
						},
					},
				},
			},
		},
		func(s *discordgo.Session, i *discordgo.InteractionCreate) {
			// Get public key from SSH key add subcommand
			key := i.ApplicationCommandData().Options[0].Options[0].Options[0].StringValue()

			initialMessage(s, i, "Checking key...")

			_, _, _, _, err := ssh.ParseAuthorizedKey([]byte(key))
			if err != nil {
				updateMessage(s, i, "That doesn't look like a public key. Try again.")
				return
			}
			// Get Ent user object
			uid := i.Member.User.ID
			u, err := b.getUserFromUID(uid)
			if err != nil {
				updateMessage(s, i, "Hmmm, I can't find your user data. Trying to find you based on your username...")
				u, err = b.getUserFromUsername(i.Member.User.Username)
				if err != nil {
					updateMessage(s, i, "Couldn't find your user entry. Please reach out to a moderator.")
					return
				}
			}

			// Add key to user's 'key' field
			if err := b.addKey(u, key); err != nil {
				updateMessage(s, i, "Failed adding key to user profile :(")
				return
			}
			updateMessage(s, i, "Successfully added key to database!")
		}
}
