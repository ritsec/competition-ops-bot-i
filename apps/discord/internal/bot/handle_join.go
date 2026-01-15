package bot

import (
	"log"
	"strings"

	"github.com/bwmarrin/discordgo"
	"github.com/ritsec/competition-ops-bot-i/ent/role"
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

	// Update UID of user to their Discord UID
	u.Update().SetUID(m.User.ID).Save(b.ClientCtx)

	// Get user's team
	t, err := u.QueryTeam().All(b.ClientCtx)
	if err != nil {
		return
	}

	// Get roles from team
	roles, err := t[0].QueryRole().All(b.ClientCtx)
	if err != nil {
		return
	}

	// Assign roles from team
	for _, r := range roles {
		err := b.Session.GuildMemberRoleAdd(m.GuildID, m.User.ID, r.ID)
		if err != nil {
			return
		}
	}

	// Give lead role if user is a lead
	if u.Lead {
		lead, err := b.Client.Role.
			Query().
			Where(role.Name("Leads")).
			Only(b.ClientCtx)
		if err != nil {
			return
		}

		err = b.Session.GuildMemberRoleAdd(m.GuildID, m.User.ID, lead.ID)
		if err != nil {
			return
		}
	}
}
