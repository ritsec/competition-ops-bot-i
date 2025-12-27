package bot

import (
	"strings"

	"github.com/bwmarrin/discordgo"
	"github.com/ritsec/competition-ops-bot-i/ent/user"
)

func (b *Bot) Join(s *discordgo.Session, m *discordgo.GuildMemberAdd) {
	username := strings.ToLower(m.User.Username)

	_, err := b.Client.User.
		Query().
		Where(user.Username(username)).
		Only(b.ClientCtx)
	if err != nil {
		panic(err) // TODO: Give a user a role to see a channel where they can request roles
	}
}
