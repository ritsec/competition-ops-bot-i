package bot

import (
	"context"
	"fmt"

	"github.com/bwmarrin/discordgo"
	"github.com/ritsec/competition-ops-bot-i/ent"
	"github.com/ritsec/competition-ops-bot-i/ent/role"
	"github.com/ritsec/competition-ops-bot-i/internal/database"
	"github.com/ritsec/competition-ops-bot-i/internal/logging"
)

// Join returns a handler that links a joining member to their database entry
// and gives them their team roles
func Join(db *ent.Client) func(s *discordgo.Session, m *discordgo.GuildMemberAdd) {
	return func(s *discordgo.Session, m *discordgo.GuildMemberAdd) {
		ctx, span := logging.StartSpan(
			context.Background(),
			"bot.handle_join:Join",
		)
		defer span.End()

		// Query Ent for user
		u, err := database.GetUserFromUsername(ctx, db, m.User.Username)
		if err != nil {
			logging.Info(ctx, fmt.Sprintf("User %s is not in database", m.User.Username))
			return // TODO: Give a user a role to see a channel where they can request roles
		}

		// Update UID of user to their Discord UID
		u, err = db.User.UpdateOne(u).SetUID(m.User.ID).Save(ctx)
		if err != nil {
			logging.Error(ctx, fmt.Sprintf("Could not set UID for %s: %v", m.User.Username, err))
			return
		}

		// Get user's team
		t, err := u.QueryTeam().Only(ctx)
		if err != nil {
			logging.Warn(ctx, fmt.Sprintf("Could not get team for %s: %v", m.User.Username, err))
			return
		}

		// Get roles from team
		roles, err := t.QueryRole().All(ctx)
		if err != nil {
			logging.Error(ctx, fmt.Sprintf("Could not get roles for %s: %v", m.User.Username, err))
			return
		}

		// Give lead role if user is a lead
		if u.Lead {
			lead, err := db.Role.
				Query().
				Where(role.Name("Leads")).
				Only(ctx)
			if err != nil {
				logging.Error(ctx, fmt.Sprintf("Could not get Leads role: %v", err))
			} else {
				roles = append(roles, lead)
			}
		}

		// Assign roles
		for _, r := range roles {
			if err := s.GuildMemberRoleAdd(m.GuildID, m.User.ID, r.ID); err != nil {
				logging.Error(ctx, fmt.Sprintf("Could not give role %s to %s: %v", r.Name, m.User.Username, err))
			}
		}
	}
}
