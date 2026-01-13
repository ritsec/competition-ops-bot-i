package bot

import (
	. "gopkg.in/check.v1"

	"github.com/bwmarrin/discordgo"
	"github.com/ewohltman/discordgo-mock/mockconstants"
	_ "github.com/mattn/go-sqlite3"
	"github.com/ritsec/competition-ops-bot-i/ent/role"
	"github.com/ritsec/competition-ops-bot-i/ent/user"
)

func (h *handlerSuite) TestJoinBlue(c *C) {
	// Create instance of bot to use its methods for setup
	// and testing
	bot := &Bot{
		Session:   h.session,
		Client:    h.client,
		ClientCtx: h.ctx,
	}

	// Create mock CSV entry array
	mockCSV := []*Blue{
		{
			School:    "foo",
			TeamNum:   "1",
			Teammate1: h.member.User.Username,
		},
	}

	// Populate database
	err := bot.handleBlue(mockCSV)
	if err != nil {
		c.Fatal(err)
	}

	// Simulate event of member join
	err = bot.Join(h.session, &discordgo.GuildMemberAdd{
		Member: h.member,
	})
	c.Assert(err, IsNil)

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
