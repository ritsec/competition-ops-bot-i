package bot

import (
	"log"
	"strings"

	"github.com/bwmarrin/discordgo"
	"github.com/ritsec/competition-ops-bot-i/ent/user"
)

func (b *Bot) Join(s *discordgo.Session, m *discordgo.GuildMemberAdd) {
	username := strings.ToLower(m.User.Username)

	// Query Ent for user
	u, err := b.Client.User.
		Query().
		Where(user.Username(username)).
		Only(b.ClientCtx)
	if err != nil {
		log.Printf("user %s is not in database", username)
		return // TODO: Give a user a role to see a channel where they can request roles
	}

	// Get user's team
	t, err := u.QueryTeam().All(b.ClientCtx)
	if err != nil {
		panic(err)
	}

	// Get roles from team
	roles, err := t[0].QueryRole().All(b.ClientCtx)
	if err != nil {
		panic(err)
	}

	// Assign roles from team
	for _, r := range roles {
		err := b.Session.GuildMemberRoleAdd(m.GuildID, m.User.ID, r.ID)
		if err != nil {
			log.Println(err)
		}
	}
}
