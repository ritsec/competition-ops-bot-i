package slash

import (
	"context"
	"fmt"
	"strings"

	"github.com/bwmarrin/discordgo"
	"github.com/ritsec/competition-ops-bot-i/ent"
	"github.com/ritsec/competition-ops-bot-i/internal/logging"
	"github.com/ritsec/competition-ops-bot-i/internal/utils"
)

func Query(db *ent.Client) (*discordgo.ApplicationCommand, func(s *discordgo.Session, i *discordgo.InteractionCreate)) {
	return &discordgo.ApplicationCommand{
			Name:                     "query",
			Description:              "Query COBI data",
			DefaultMemberPermissions: &utils.Admin,
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
			ctx, span := logging.StartSpan(
				context.Background(),
				"commands.slash.query:Query",
			)
			defer span.End()

			choice := i.ApplicationCommandData().Options[0].StringValue()

			if !startStatus(ctx, s, i, "Querying database...") {
				return
			}

			switch choice {
			case "Keys":
				var content strings.Builder

				keys, err := db.Key.Query().All(ctx)
				if err != nil {
					reportError(ctx, s, i, "Could not query keys", err)
					return
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
					logging.Error(ctx, fmt.Sprintf("Could not send keys file: %v", err))
					setStatus(ctx, s, i, "Failed to send file :( Please check channel or bot permissions")
					return
				}
				setStatus(ctx, s, i, "Query complete!")
			}
		}
}
