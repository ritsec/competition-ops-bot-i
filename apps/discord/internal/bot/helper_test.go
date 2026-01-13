package bot

import (
	"context"
	"net/http"
	"testing"

	. "gopkg.in/check.v1"

	"github.com/bwmarrin/discordgo"
	_ "github.com/mattn/go-sqlite3"
	"github.com/ritsec/competition-ops-bot-i/ent"
	"github.com/ritsec/competition-ops-bot-i/ent/enttest"

	"github.com/ewohltman/discordgo-mock/mockconstants"
	"github.com/ewohltman/discordgo-mock/mockguild"
	"github.com/ewohltman/discordgo-mock/mockrest"
	"github.com/ewohltman/discordgo-mock/mockrole"
	"github.com/ewohltman/discordgo-mock/mocksession"
	"github.com/ewohltman/discordgo-mock/mockstate"
)

func Test(t *testing.T) { TestingT(t) }

type eventSuite struct {
	session *discordgo.Session
	client  *ent.Client
	ctx     context.Context
	bot     *Bot
}

var _ = Suite(&eventSuite{})

func (e *eventSuite) SetUpTest(c *C) {
	// Set up Ent sqlite3
	e.client = enttest.Open(c, "sqlite3", "file:ent?mode=memory&_fk=1")
	e.ctx = context.Background()

	// Create instance of bot to use its methods for setup
	// and testing.
	e.bot = &Bot{
		Session:   e.session,
		Client:    e.client,
		ClientCtx: e.ctx,
	}

	// Set up DiscordGo state
	state, err := e.mockServer()
	if err != nil {
		c.Fatal(err)
	}

	// Set up DiscordGo session
	e.session, err = mocksession.New(
		mocksession.WithClient(&http.Client{
			Transport: mockrest.NewTransport(state),
		}),
	)
	if err != nil {
		c.Fatal(err)
	}

}

func (e *eventSuite) TearDownTest(c *C) {
	// Tear down Ent sqlite3
	err := e.client.Close()
	c.Assert(err, IsNil)

	// Tear down DiscordGo session
	err = e.session.Close()
	c.Assert(err, IsNil)
}

// mockServer returns a pointer to a DiscordGo state with test server data
// that will be used to create the mock session. When creating the mock data,
// it will also populate the Ent database accordingly.
func (e *eventSuite) mockServer() (*discordgo.State, error) {
	// Create mock common Blue role and Ent Role
	roleBlueCommon := mockrole.New(
		mockrole.WithID(mockconstants.TestRole),
		mockrole.WithName("Blue Team"),
		mockrole.WithPermissions(discordgo.PermissionViewChannel),
	)

	// Create mock Blue Team 1 role and Ent Role
	roleBlue := mockrole.New(
		mockrole.WithID(mockconstants.TestRole),
		mockrole.WithName("Blue Team 1"),
		mockrole.WithPermissions(discordgo.PermissionViewChannel),
	)

	return mockstate.New(
		mockstate.WithGuilds(
			mockguild.New(
				mockguild.WithID(mockconstants.TestGuild),
				mockguild.WithName(mockconstants.TestGuild),
				mockguild.WithRoles(roleBlue, roleBlueCommon),
			),
		),
	)
}
