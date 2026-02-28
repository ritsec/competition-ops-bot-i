package bot

import (
	"fmt"
	"log"

	"github.com/bwmarrin/discordgo"
	"github.com/ritsec/competition-ops-bot-i/ent/team"
)

func (b *Bot) TestFunc() (*discordgo.ApplicationCommand, func(s *discordgo.Session, i *discordgo.InteractionCreate)) {
	return &discordgo.ApplicationCommand{
			Name:                     "test",
			Description:              "Test COBI function",
			DefaultMemberPermissions: &Admin,
			Options: []*discordgo.ApplicationCommandOption{
				{
					Type:        discordgo.ApplicationCommandOptionString,
					Name:        "option",
					Description: "Option of what to test",
					Required:    true,
					Choices: []*discordgo.ApplicationCommandOptionChoice{
						{
							Name:  "Message",
							Value: "Message",
						},
					},
				},
			},
		},
		func(s *discordgo.Session, i *discordgo.InteractionCreate) {
			choice := i.ApplicationCommandData().Options[0].StringValue()

			switch choice {
			case "Message":
				teams, err := b.Client.Team.Query().
					Where(team.TypeEQ(team.TypeBlue)).
					All(b.ClientCtx)
				if err != nil {
					log.Println(err)
				}

				for _, team := range teams {
					// Get team's channel
					channel, err := team.QueryChannel().Only(b.ClientCtx)
					if err != nil {
						log.Printf("Could not get channel for Blue Team %d", team.Number)
						continue
					}

					message := fmt.Sprintf("Hello team %d! This is a test to make sure COBI can reach everyone. Please disregard this message, or say hi to COBI if you so please.", team.Number)
					embed := &discordgo.MessageEmbed{
						Title:       "Test",
						Description: message,
					}

					// Send embed to channel ID
					_, err = b.Session.ChannelMessageSendEmbed(channel.ID, embed)
				}
			}
		}
}
