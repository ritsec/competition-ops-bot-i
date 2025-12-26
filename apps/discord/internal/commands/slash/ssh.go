package slash

import (
	"log"

	"github.com/bwmarrin/discordgo"
)

func SSH() (*discordgo.ApplicationCommand, func(s *discordgo.Session, i *discordgo.InteractionCreate)) {
	return &discordgo.ApplicationCommand{
			Name:        "ssh",
			Description: "Register an SSH key for Black Team",
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
			ssOption := i.ApplicationCommandData().Options[0].StringValue()

			if ssOption == "Add" {
				log.Println("received /SSH Add slash command")
			} else {
				return
			}
		}
}
