package bot

import (
	"fmt"
	"log"

	"github.com/google/uuid"
	"github.com/ritsec/competition-ops-bot-i/ent"
	"github.com/ritsec/competition-ops-bot-i/ent/role"
	"github.com/ritsec/competition-ops-bot-i/ent/team"
)

// map of teams to the name of their default roles
var defaultRole = map[string]string{
	"blue":  "Blue Team",
	"red":   "Red Team",
	"black": "Black Team",
}

var blackSubteams = map[string]team.Subteam{
	"Infra":   team.SubteamInfra,
	"Linux":   team.SubteamLinux,
	"Windows": team.SubteamWindows,
	"Scoring": team.SubteamScoring,
	"Logging": team.SubteamLogging,
	"Store":   team.SubteamStore,
	"CTF":     team.SubteamCtf,
	"KotH":    team.SubteamKoth,
}

// addRoles adds an array of roles to a team via edges
func (b *Bot) addRoles(team *ent.Team, roles ...string) error {
	for _, roleStr := range roles {
		r, err := b.Client.Role.
			Query().
			Where(role.Name(roleStr)).
			Only(b.ClientCtx)
		if err != nil {
			return err
		}

		team.Update().AddRole(r).Save(b.ClientCtx)
	}
	return nil
}

// createUser is a helper function to create a user with the given username
// and a default UUID.
func (b *Bot) createUser(username string) (*ent.User, error) {
	u, err := b.Client.User.
		Create().
		SetUID(uuid.New().String()). // Set temporary uuid to be changed on join event
		SetUsername(username).
		Save(b.ClientCtx)

	return u, err
}

// getBlue handles requests to get/create Blue Teams
func (b *Bot) getBlue(i int) (*ent.Team, error) {
	// Check if team already exists
	t, err := b.Client.Team.
		Query().
		Where(team.Number(i)).
		Only(b.ClientCtx)

	if err != nil { // Create team if it doesn't exist
		log.Printf("creating team %d", i)

		t, err = b.Client.Team.
			Create().
			SetType("blue").
			SetNumber(i).
			Save(b.ClientCtx)
		if err != nil {
			return nil, err
		}

		// Add default and individual team roles
		teamRole := fmt.Sprintf("Blue Team %d", i)

		if err := b.addRoles(t, defaultRole["blue"], teamRole); err != nil {
			return nil, err
		}
	}

	return t, err
}

// getRed handles requests to get/create Red Team
func (b *Bot) getRed() (*ent.Team, error) {
	// Check if Red team exists
	t, err := b.Client.Team.
		Query().
		Where(team.TypeEQ("red")).
		Only(b.ClientCtx)

	if err != nil { // Create Red team if it doesn't exist
		log.Println("creating Red team")

		t, err = b.Client.Team.
			Create().
			SetType("red").
			Save(b.ClientCtx)

		if err := b.addRoles(t, defaultRole["red"]); err != nil {
			return nil, err
		}
	}

	return t, err
}

// getBlack handles requests to get/create Black Teams
func (b *Bot) getBlack() (map[string]*ent.Team, error) {
	teams := make(map[string]*ent.Team)

	for name, subteam := range blackSubteams {
		// Check if subteam exists
		t, err := b.Client.Team.
			Query().
			Where(team.And(team.TypeEQ("black"), team.SubteamEQ(subteam))).
			Only(b.ClientCtx)
		if err != nil { // Create if it doesn't exist
			t, err = b.Client.Team.
				Create().
				SetType("black").
				SetSubteam(subteam).
				Save(b.ClientCtx)
			if err != nil {
				return nil, err
			}

			if err := b.addRoles(t, defaultRole["black"], name); err != nil {
				return nil, err
			}

			teams[name] = t
		}
	}

	return teams, nil
}
