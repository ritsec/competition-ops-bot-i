package bot

import (
	"fmt"
	"log"

	"github.com/bwmarrin/discordgo"
)

// ptr is a helper function to cleanly return a pointer to a value
func ptr[T any](value T) *T {
	return &value
}

var (
	// Define role parameters for universal Blue Team role
	bluePermissions int64 = 1<<6 | 1<<11 | 1<<14 | 1<<15 | 1<<16 | 1<<26 | 1<<35 | 1<<38
	blueTeam              = discordgo.RoleParams{
		Name:        "Blue Team",
		Color:       ptr(2123412),
		Hoist:       ptr(true),
		Permissions: &bluePermissions,
		Mentionable: ptr(true),
	}
)

func (b *Bot) Build() (*discordgo.ApplicationCommand, func(s *discordgo.Session, i *discordgo.InteractionCreate)) {
	return &discordgo.ApplicationCommand{
			Name:                     "build",
			Description:              "Build server teams",
			DefaultMemberPermissions: &Admin,
			Options: []*discordgo.ApplicationCommandOption{
				{
					Type:        discordgo.ApplicationCommandOptionString,
					Name:        "option",
					Description: "Option of which team to build",
					Required:    true,
					Choices: []*discordgo.ApplicationCommandOptionChoice{
						{
							Name:  "Blue",
							Value: "Blue",
						},
					},
				},
			},
		},
		func(s *discordgo.Session, i *discordgo.InteractionCreate) {
			choice := i.ApplicationCommandData().Options[0].StringValue()

			initialMessage(s, i, "Building server...")

			switch choice {
			case "Blue":
				if err := b.buildBlue(); err != nil {
					log.Println(err)
				}
			}

			updateMessage(s, i, "Finished building requested resources!")
		}
}

// buildBlue will build the universal Blue Team role, Blue Team N roles, shared channels and individual channels
func (b *Bot) buildBlue() error {
	// Create universal blue role
	_, err := b.Session.GuildRoleCreate(guildID, &blueTeam)
	if err != nil {
		return err
	}

	for i := 1; i <= numBlueTeams; i++ {
		name := fmt.Sprintf("Blue Team %d", i)
		_, err := b.Session.GuildRoleCreate(guildID, &discordgo.RoleParams{
			Name:        name,
			Color:       ptr(2123412),
			Mentionable: ptr(true),
		})
		if err != nil {
			return err
		}
	}

	return nil
}
