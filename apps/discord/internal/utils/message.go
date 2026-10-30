package utils

import (
	"fmt"

	"github.com/bwmarrin/discordgo"
	"github.com/ritsec/competition-ops-bot-i/ent"
)

// maxEmbedFields is the most fields Discord allows in one embed
const maxEmbedFields = 25

// MessageCreds builds the embed sent to a team with a field for each of
// their service logins
func MessageCreds(num int, creds *ent.Credential) *discordgo.MessageEmbed {
	var fields []*discordgo.MessageEmbedField
	for _, service := range creds.Services {
		if len(fields) == maxEmbedFields {
			break
		}

		fields = append(fields, &discordgo.MessageEmbedField{
			Name:  service.Name,
			Value: formatServiceCred(service.Username, service.Password),
		})
	}

	return &discordgo.MessageEmbed{
		Title:       "🔐 Credentials 🔐",
		Description: fmt.Sprintf("Here are Team %d's credentials. Keep them safe!", num),
		Color:       0x5865F2,
		Fields:      fields,
	}
}

func formatServiceCred(username, password string) string {
	if username == "" {
		return fmt.Sprintf(
			"```text\nPassword: %s\n```",
			password,
		)
	}
	return fmt.Sprintf(
		"```text\nUsername: %s\nPassword: %s\n```",
		username, password,
	)
}
