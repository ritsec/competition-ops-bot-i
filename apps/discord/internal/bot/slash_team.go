package bot

import (
	"log"

	"github.com/bwmarrin/discordgo"
)

func Team() (*discordgo.ApplicationCommand, func(s *discordgo.Session, i *discordgo.InteractionCreate)) {
	return &discordgo.ApplicationCommand{
			Name:        "team",
			Description: "Submit a CSV of team members",
			Options: []*discordgo.ApplicationCommandOption{
				{
					Type:        discordgo.ApplicationCommandOptionString,
					Name:        "option",
					Description: "Option of team",
					Required:    true,
					Choices: []*discordgo.ApplicationCommandOptionChoice{
						{
							Name:  "Blue",
							Value: "Blue",
						},
						{
							Name:  "Red",
							Value: "Red",
						},
						{
							Name:  "Black",
							Value: "Black",
						},
						{
							Name:  "White",
							Value: "White",
						},
					},
				},
				{
					Type:        discordgo.ApplicationCommandOptionAttachment,
					Name:        "attachment",
					Description: "Attach CSV",
					Required:    true,
				},
			},
		},
		func(s *discordgo.Session, i *discordgo.InteractionCreate) {
			choice := i.ApplicationCommandData().Options[0].StringValue()

			if choice == "Red" {
				log.Println("received /Team Red command")
			} else {
				return
			}
		}
}
