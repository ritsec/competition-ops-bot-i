package utils

import (
	"fmt"
	"strings"

	"github.com/bwmarrin/discordgo"
	"github.com/ritsec/competition-ops-bot-i/ent"
)

func MessageCreds(num int, creds *ent.Credential) *discordgo.MessageEmbed {
	compServices := fmt.Sprintf("For your convenience, [Authentik](https://auth.ists.io), "+
		"[Compsole](https://compsole.ritsec.cloud), "+
		"Scorify (https://scoring.ists.space), "+
		"[Store](https://store.ists.space), and "+
		"[CTFd](https://ctf.ists.space) share a username and password:\n%s", formatServiceCred(creds.Authentik))

	return &discordgo.MessageEmbed{
		Title:       "🔐 Credentials 🔐",
		Description: "Here are your teams credentials. Keep them safe!",
		Color:       0x5865F2,
		Fields: []*discordgo.MessageEmbedField{
			{
				Name:  "Competition Services",
				Value: compServices,
			},
			{
				Name:  "Wazuh",
				Value: formatServiceCred(creds.Wazuh),
			},
			{
				Name:  "pfSense",
				Value: formatServiceCred(creds.Pfsense),
			},
			{
				Name:  "Hosts",
				Value: formatServiceCred(creds.Default),
			},
			{
				Name:  "Kali",
				Value: formatServiceCred(creds.Kali),
			},
		},
	}
}

func formatServiceCred(cred string) string {
	creds := strings.Split(cred, ":")
	if len(creds) == 1 {
		return fmt.Sprintf(
			"```text\nPassword: %s\n```",
			creds[0],
		)
	} else {
		return fmt.Sprintf(
			"```text\nUsername: %s\nPassword: %s\n```",
			creds[0], creds[1],
		)
	}
}
