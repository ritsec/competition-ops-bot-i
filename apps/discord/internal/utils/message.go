package utils

import (
	"fmt"

	"github.com/bwmarrin/discordgo"
	"github.com/ritsec/competition-ops-bot-i/ent"
)

func MessageCreds(num int, creds *ent.Credential) *discordgo.MessageEmbed {
	return &discordgo.MessageEmbed{
		Title:       "🔐 Credentials 🔐",
		Description: "Here are your teams credentials. Keep them safe!",
		Color:       0x5865F2,
		Fields: []*discordgo.MessageEmbedField{
			{
				Name:  "Compsole",
				Value: formatServiceCred(num, creds.Compsole),
			},
			{
				Name:  "Scorify",
				Value: formatServiceCred(num, creds.Scorify),
			},
			{
				Name:  "Authentik",
				Value: formatServiceCred(num, creds.Authentik),
			},
		},
	}
}

func formatServiceCred(num int, password string) string {
	return fmt.Sprintf(
		"```text\nUsername: team%d\nPassword: %s\n```",
		num,
		password,
	)
}
