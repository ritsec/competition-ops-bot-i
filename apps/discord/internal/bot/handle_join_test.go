package bot

import (
	"testing"

	. "gopkg.in/check.v1"

	"github.com/bwmarrin/discordgo"
	"github.com/ewohltman/discordgo-mock/mockconstants"
	_ "github.com/mattn/go-sqlite3"
	"github.com/ritsec/competition-ops-bot-i/ent/role"
	"github.com/ritsec/competition-ops-bot-i/ent/user"
	"github.com/ritsec/competition-ops-bot-i/internal/database"
)

func Test(t *testing.T) { TestingT(t) }

func (h *handlerSuite) TestJoinBlue(c *C) {
	// Populate database with Blue Team 1 and the mock member
	t, err := database.GetBlue(h.ctx, h.client, 1)
	if err != nil {
		c.Fatal(err)
	}
	u, err := database.CreateUser(h.ctx, h.client, h.member.User.Username)
	if err != nil {
		c.Fatal(err)
	}
	if err := h.client.Team.UpdateOne(t).AddUser(u).Exec(h.ctx); err != nil {
		c.Fatal(err)
	}

	// Simulate event of member join
	Join(h.client)(h.session, &discordgo.GuildMemberAdd{
		Member: h.member,
	})

	// Check that user's UID field is changed to their Discord UID
	uid, err := h.client.User.Query().
		Where(user.Username(h.member.User.Username)).
		Select(user.FieldUID).
		String(h.ctx)
	c.Check(uid, Equals, mockconstants.TestUser)

	// Get Ent user object
	user, err := h.client.User.Query().
		Where(user.UID(mockconstants.TestUser)).
		Only(h.ctx)

	// Check that user has received team roles
	roles, err := user.QueryTeam().QueryRole().
		Select(role.FieldID).
		Strings(h.ctx)
	c.Check(roles, DeepEquals, h.rolesBlue)
}
