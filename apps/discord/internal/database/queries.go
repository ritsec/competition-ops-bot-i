package database

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/ritsec/competition-ops-bot-i/ent"
	"github.com/ritsec/competition-ops-bot-i/ent/role"
	"github.com/ritsec/competition-ops-bot-i/ent/team"
	"github.com/ritsec/competition-ops-bot-i/ent/user"
)

// map of teams to the name of their default roles
var defaultRole = map[team.Type]string{
	team.TypeBlue:  "Blue Team",
	team.TypeRed:   "Red Team",
	team.TypeBlack: "Black Team",
}

// BlackSubteams maps Black subteam names to the Ent schema enum
var BlackSubteams = map[string]team.Subteam{
	"Infra":   team.SubteamInfra,
	"Linux":   team.SubteamLinux,
	"Windows": team.SubteamWindows,
	"Scoring": team.SubteamScoring,
	"Logging": team.SubteamLogging,
	"Store":   team.SubteamStore,
	"CTF":     team.SubteamCtf,
	"KotH":    team.SubteamKoth,
}

// WithTx runs fn inside a transaction, committing if it succeeds and rolling
// back if it returns an error
func WithTx(ctx context.Context, db *ent.Client, fn func(tx *ent.Client) error) error {
	tx, err := db.Tx(ctx)
	if err != nil {
		return err
	}

	if err := fn(tx.Client()); err != nil {
		// The original error is more useful to report than a rollback error
		_ = tx.Rollback()
		return err
	}

	return tx.Commit()
}

// AddRoles links the named roles to a team, skipping any already linked
func AddRoles(ctx context.Context, db *ent.Client, t *ent.Team, roles ...string) error {
	for _, name := range roles {
		r, err := db.Role.
			Query().
			Where(role.Name(name)).
			Only(ctx)
		if err != nil {
			return fmt.Errorf("could not find role %q: %w", name, err)
		}

		// Team and role are many-to-many, so adding an existing link would fail
		linked, err := t.QueryRole().Where(role.ID(r.ID)).Exist(ctx)
		if err != nil {
			return err
		}
		if linked {
			continue
		}

		if err := db.Team.UpdateOne(t).AddRole(r).Exec(ctx); err != nil {
			return err
		}
	}
	return nil
}

// CreateUser is a helper function to create a user with the given username
// and a default UUID.
func CreateUser(ctx context.Context, db *ent.Client, username string) (*ent.User, error) {
	return db.User.
		Create().
		SetUID(uuid.New().String()). // Set temporary uuid to be changed on join event
		SetUsername(strings.ToLower(username)).
		Save(ctx)
}

// GetUserFromUID returns an Ent user object based on the provided Discord UID
func GetUserFromUID(ctx context.Context, db *ent.Client, uid string) (*ent.User, error) {
	return db.User.
		Query().
		Where(user.UID(uid)).
		Only(ctx)
}

// GetUserFromUsername returns an Ent user object based on the provided Discord username
// TODO: Find a better way to check if user has already been processed
func GetUserFromUsername(ctx context.Context, db *ent.Client, username string) (*ent.User, error) {
	return db.User.
		Query().
		Where(user.Username(strings.ToLower(username))).
		Only(ctx)
}

// GetBlue handles requests to get/create Blue Teams.
// This assumes that there are roles on the server correlating
// to the number of blue teams specified in the config.
func GetBlue(ctx context.Context, db *ent.Client, num int) (*ent.Team, error) {
	// Check if team already exists
	t, err := db.Team.
		Query().
		Where(team.TypeEQ(team.TypeBlue), team.Number(num)).
		Only(ctx)
	if ent.IsNotFound(err) { // Create team if it doesn't exist
		t, err = db.Team.
			Create().
			SetType(team.TypeBlue).
			SetNumber(num).
			Save(ctx)
	}
	if err != nil {
		return nil, err
	}

	// Add default and individual team roles
	teamRole := fmt.Sprintf("Blue Team %d", num)
	if err := AddRoles(ctx, db, t, defaultRole[team.TypeBlue], teamRole); err != nil {
		return nil, err
	}

	return t, nil
}

// GetRed handles requests to get/create Red Team
func GetRed(ctx context.Context, db *ent.Client) (*ent.Team, error) {
	// Check if Red team exists
	t, err := db.Team.
		Query().
		Where(team.TypeEQ(team.TypeRed)).
		Only(ctx)
	if ent.IsNotFound(err) { // Create Red team if it doesn't exist
		t, err = db.Team.
			Create().
			SetType(team.TypeRed).
			Save(ctx)
	}
	if err != nil {
		return nil, err
	}

	if err := AddRoles(ctx, db, t, defaultRole[team.TypeRed]); err != nil {
		return nil, err
	}

	return t, nil
}

// GetBlack handles requests to get/create Black Teams, returning a map
// of the subteam name to the Ent team object. Subteams whose roles are
// missing from the server are still returned, along with an error
// describing the missing roles.
func GetBlack(ctx context.Context, db *ent.Client) (map[string]*ent.Team, error) {
	teams := make(map[string]*ent.Team)
	var roleErrs []error

	for name, subteam := range BlackSubteams {
		// Check if subteam exists
		t, err := db.Team.
			Query().
			Where(team.TypeEQ(team.TypeBlack), team.SubteamEQ(subteam)).
			Only(ctx)
		if ent.IsNotFound(err) { // Create if it doesn't exist
			t, err = db.Team.
				Create().
				SetType(team.TypeBlack).
				SetSubteam(subteam).
				Save(ctx)
		}
		if err != nil {
			return nil, err
		}
		teams[name] = t

		// Not every server has all the ISTS roles (e.g. quals), so a missing
		// role shouldn't stop the subteam from being usable
		if err := AddRoles(ctx, db, t, defaultRole[team.TypeBlack], name); err != nil {
			roleErrs = append(roleErrs, err)
		}
	}

	return teams, errors.Join(roleErrs...)
}

// AddLeads takes an array of Discord usernames and sets their
// 'lead' boolean column to true. This is used during the role assignment
// during their join event.
func AddLeads(ctx context.Context, db *ent.Client, leads []string) error {
	for _, lead := range leads {
		if lead == "" {
			continue
		}

		// Get user
		u, err := GetUserFromUsername(ctx, db, lead)
		if err != nil {
			return fmt.Errorf("could not find lead %q: %w", lead, err)
		}

		// Set the user's lead value to true
		if err := db.User.UpdateOne(u).SetLead(true).Exec(ctx); err != nil {
			return err
		}
	}

	return nil
}

// AddKey adds an SSH key to the user's 'key' field
func AddKey(ctx context.Context, db *ent.Client, u *ent.User, key string) error {
	// Check if user already has SSH key in database
	k, err := u.QueryKey().Only(ctx)
	switch {
	case ent.IsNotFound(err): // If it doesn't exist, create the key edge
		k, err = db.Key.
			Create().
			SetKeys([]string{key}).
			Save(ctx)
		if err != nil {
			return err
		}

		// Add the key to the user's edge
		return db.User.UpdateOne(u).AddKey(k).Exec(ctx)
	case err != nil:
		return err
	default: // If it exists, append the key to the list
		return db.Key.UpdateOne(k).AppendKeys([]string{key}).Exec(ctx)
	}
}

// GetRoleID returns the Discord ID of the role with the given name, or an
// empty string if it doesn't exist
func GetRoleID(ctx context.Context, db *ent.Client, name string) string {
	r, err := db.Role.Query().
		Where(role.Name(name)).
		Only(ctx)
	if err != nil {
		return ""
	}

	return r.ID
}
