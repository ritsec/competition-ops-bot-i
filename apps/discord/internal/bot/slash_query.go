package bot

import (
	"fmt"
	"log"
	"strings"

	"github.com/bwmarrin/discordgo"
)

func (b *Bot) Query() (*discordgo.ApplicationCommand, func(s *discordgo.Session, i *discordgo.InteractionCreate)) {
	return &discordgo.ApplicationCommand{
			Name:                     "query",
			Description:              "Query COBI data",
			DefaultMemberPermissions: &Admin,
			Options: []*discordgo.ApplicationCommandOption{
				{
					Type:        discordgo.ApplicationCommandOptionString,
					Name:        "option",
					Description: "Option of what to query",
					Required:    true,
					Choices: []*discordgo.ApplicationCommandOptionChoice{
						{
							Name:  "Keys",
							Value: "Keys",
						},
					},
				},
			},
		},
		func(s *discordgo.Session, i *discordgo.InteractionCreate) {
			choice := i.ApplicationCommandData().Options[0].StringValue()

			initialMessage(s, i, "Querying database...")
			switch choice {
			case "Keys":
				var content strings.Builder

				keys, err := b.Client.Key.Query().All(b.ClientCtx)
				if err != nil {
					log.Fatal(err)
				}

				for _, keyArray := range keys {
					for _, key := range keyArray.Keys {
						fmt.Fprintf(&content, "- \"%s\"\n", key)
					}
				}

				file := &discordgo.File{
					Name:        "ssh_keys.txt",
					ContentType: "text/plain",
					Reader:      strings.NewReader(content.String()),
				}

				_, err = s.ChannelMessageSendComplex(i.ChannelID, &discordgo.MessageSend{
					Content: "Here are the registered SSH public keys:",
					Files:   []*discordgo.File{file},
				})
				if err != nil {
					updateMessage(s, i, "Failed to send file :( Please check channel or bot permissions")
					return
				}
				updateMessage(s, i, "Query complete!")
			}
		}
}
