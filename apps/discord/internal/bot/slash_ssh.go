package bot

import (
	"log"

	"github.com/bwmarrin/discordgo"
)

func SSH() (*discordgo.ApplicationCommand, func(s *discordgo.Session, i *discordgo.InteractionCreate)) {
	return &discordgo.ApplicationCommand{
			Name:                     "ssh",
			Description:              "Register an SSH key for Black Team",
			DefaultMemberPermissions: &BlackTeam,
			Options: []*discordgo.ApplicationCommandOption{
				{
					Type:        discordgo.ApplicationCommandOptionString,
					Name:        "option",
					Description: "Option of add or remove",
					Required:    true,
					Choices: []*discordgo.ApplicationCommandOptionChoice{
						{
							Name:  "Add",
							Value: "Add",
						}, // TODO: Option to remove entry
					},
				},
			},
		},
		func(s *discordgo.Session, i *discordgo.InteractionCreate) {
			choice := i.ApplicationCommandData().Options[0].StringValue()

			if choice == "Add" {
				log.Println("received /SSH Add command")
			} else {
				return
			}
		}
}
