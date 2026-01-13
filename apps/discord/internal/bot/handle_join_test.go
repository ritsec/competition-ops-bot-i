package bot

import (
	. "gopkg.in/check.v1"

	"github.com/bwmarrin/discordgo"
	"github.com/ewohltman/discordgo-mock/mockconstants"
	"github.com/ewohltman/discordgo-mock/mockmember"
	"github.com/ewohltman/discordgo-mock/mockuser"
	_ "github.com/mattn/go-sqlite3"
)

func (e *eventSuite) TestJoinBlue(c *C) {
	// Create test Guild Member
	testUser := mockuser.New(
		mockuser.WithID(mockconstants.TestUser),
		mockuser.WithUsername(mockconstants.TestUser),
		mockuser.WithBotFlag(true),
	)
	testMember := mockmember.New(
		mockmember.WithUser(testUser),
		mockmember.WithGuildID(mockconstants.TestGuild),
	)

	// Create mock CSV entry array
	mockCSV := []*Blue{
		{
			School:    "foo",
			TeamNum:   "1",
			Teammate1: testUser.Username,
		},
	}

	// Populate database
	err := e.bot.handleBlue(mockCSV)
	if err != nil {
		c.Fatal(err)
	}

	err = e.bot.Join(e.session, &discordgo.GuildMemberAdd{
		Member: testMember,
	})
	c.Assert(err, IsNil)

}
