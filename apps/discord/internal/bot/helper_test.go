package bot

import (
	"context"
	"net/http"

	. "gopkg.in/check.v1"

	"github.com/bwmarrin/discordgo"
	_ "github.com/mattn/go-sqlite3"
	"github.com/ritsec/competition-ops-bot-i/ent"
	"github.com/ritsec/competition-ops-bot-i/ent/enttest"

	"github.com/ewohltman/discordgo-mock/mockconstants"
	"github.com/ewohltman/discordgo-mock/mockguild"
	"github.com/ewohltman/discordgo-mock/mockmember"
	"github.com/ewohltman/discordgo-mock/mockrest"
	"github.com/ewohltman/discordgo-mock/mockrole"
	"github.com/ewohltman/discordgo-mock/mocksession"
	"github.com/ewohltman/discordgo-mock/mockstate"
	"github.com/ewohltman/discordgo-mock/mockuser"
)

type handlerSuite struct {
	session   *discordgo.Session
	client    *ent.Client
	ctx       context.Context
	member    *discordgo.Member
	rolesBlue []string
}

var _ = Suite(&handlerSuite{})

func (h *handlerSuite) SetUpTest(c *C) {
	// Set up Ent sqlite3
	h.client = enttest.Open(c, "sqlite3", "file:ent?mode=memory&_fk=1")
	h.ctx = context.Background()

	// Set up DiscordGo state
	state, err := h.mockServer()
	if err != nil {
		c.Fatal(err)
	}

	// Set up DiscordGo session
	h.session, err = mocksession.New(
		mocksession.WithClient(&http.Client{
			Transport: mockrest.NewTransport(state),
		}),
	)
	if err != nil {
		c.Fatal(err)
	}

}

func (h *handlerSuite) TearDownTest(c *C) {
	// Tear down Ent sqlite3
	err := h.client.Close()
	if err != nil {
		c.Fatal(err)
	}

	// Tear down DiscordGo session
	err = h.session.Close()
	if err != nil {
		c.Fatal(err)
	}
}

// mockServer returns a pointer to a DiscordGo state with test server data
// that will be used to create the mock session. When creating the mock data,
// it will also populate the Ent database accordingly.
func (h *handlerSuite) mockServer() (*discordgo.State, error) {
	// Create mock common Blue role and Ent Role
	roleBlueCommon := mockrole.New(
		mockrole.WithID("blueteam"),
		mockrole.WithName("Blue Team"),
		mockrole.WithPermissions(discordgo.PermissionViewChannel),
	)
	_, err := h.client.Role.Create().
		SetID(roleBlueCommon.ID).
		SetName(roleBlueCommon.Name).
		Save(h.ctx)
	if err != nil {
		return nil, err
	}

	// Create mock Blue Team 1 role and Ent Role
	roleBlueTeam := mockrole.New(
		mockrole.WithID("blueteam1"),
		mockrole.WithName("Blue Team 1"),
		mockrole.WithPermissions(discordgo.PermissionViewChannel),
	)
	_, err = h.client.Role.Create().
		SetID(roleBlueTeam.ID).
		SetName(roleBlueTeam.Name).
		Save(h.ctx)
	if err != nil {
		return nil, err
	}

	// Fill Blue roles array
	h.rolesBlue = []string{
		roleBlueCommon.ID,
		roleBlueTeam.ID,
	}

	// Create test Guild Member. The username is not set to mockconstants.TestUser
	// to simulate the scenario where COBI cannot know user ID prior to
	// the member joining.
	testUser := mockuser.New(
		mockuser.WithID(mockconstants.TestUser),
		mockuser.WithUsername("fakeusername"),
	)
	testMember := mockmember.New(
		mockmember.WithUser(testUser),
		mockmember.WithGuildID(mockconstants.TestGuild),
	)
	h.member = testMember
	return mockstate.New(
		mockstate.WithGuilds(
			mockguild.New(
				mockguild.WithID(mockconstants.TestGuild),
				mockguild.WithName(mockconstants.TestGuild),
				mockguild.WithRoles(roleBlueTeam, roleBlueCommon),
				mockguild.WithMembers(testMember),
			),
		),
	)
}
