package slash

import (
	"context"

	"github.com/bwmarrin/discordgo"
	"github.com/ritsec/competition-ops-bot-i/ent"
	"github.com/ritsec/competition-ops-bot-i/internal/database"
	"github.com/ritsec/competition-ops-bot-i/internal/logging"
	"github.com/ritsec/competition-ops-bot-i/internal/utils"
	"golang.org/x/crypto/ssh"
)

func SSH(db *ent.Client) (*discordgo.ApplicationCommand, func(s *discordgo.Session, i *discordgo.InteractionCreate)) {
	return &discordgo.ApplicationCommand{
			Name:                     "ssh",
			Description:              "Register an SSH key for Black Team",
			DefaultMemberPermissions: &utils.BlackTeam,
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
									Name:        "key",
									Description: "SSH public key",
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
			ctx, span := logging.StartSpan(
				context.Background(),
				"commands.slash.ssh:SSH",
			)
			defer span.End()

			// /ssh key <subcommand>
			subcommand := i.ApplicationCommandData().Options[0].Options[0]

			if subcommand.Name != "add" {
				startStatus(ctx, s, i, "Removing keys isn't supported yet.")
				return
			}

			// Get public key from SSH key add subcommand
			key := subcommand.Options[0].StringValue()

			if !startStatus(ctx, s, i, "Checking key...") {
				return
			}

			if _, _, _, _, err := ssh.ParseAuthorizedKey([]byte(key)); err != nil {
				setStatus(ctx, s, i, "That doesn't look like a public key. Try again.")
				return
			}

			// Get Ent user object
			u, err := database.GetUserFromUID(ctx, db, i.Member.User.ID)
			if err != nil {
				setStatus(ctx, s, i, "Hmmm, I can't find your user data. Trying to find you based on your username...")
				u, err = database.GetUserFromUsername(ctx, db, i.Member.User.Username)
				if err != nil {
					setStatus(ctx, s, i, "Couldn't find your user entry. Please reach out to a moderator.")
					return
				}
			}

			// Add key to user's 'key' field
			if err := database.AddKey(ctx, db, u, key); err != nil {
				reportError(ctx, s, i, "Failed adding key to user profile :(", err)
				return
			}
			setStatus(ctx, s, i, "Successfully added key to database!")
		}
}
