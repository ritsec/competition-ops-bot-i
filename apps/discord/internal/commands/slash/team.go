package slash

import (
	"context"
	"fmt"
	"reflect"
	"strconv"

	"github.com/bwmarrin/discordgo"
	"github.com/ritsec/competition-ops-bot-i/ent"
	"github.com/ritsec/competition-ops-bot-i/internal/database"
	"github.com/ritsec/competition-ops-bot-i/internal/logging"
	"github.com/ritsec/competition-ops-bot-i/internal/utils"
)

type Blue struct {
	// Blue
	School    string `csv:"School"`
	TeamNum   string `csv:"Team Number"`
	Teammate1 string `csv:"Teammate 1"`
	Teammate2 string `csv:"Teammate 2"`
	Teammate3 string `csv:"Teammate 3"`
	Teammate4 string `csv:"Teammate 4"`
	Teammate5 string `csv:"Teammate 5"`
}

type Red struct {
	// Red
	Members string `csv:"Members"`
	Leads   string `csv:"Leads"`
}

type Black struct {
	// Black
	Infra   string `csv:"Infra"`
	Linux   string `csv:"Linux"`
	Windows string `csv:"Windows"`
	Scoring string `csv:"Scoring"`
	Logging string `csv:"Logging"`
	Store   string `csv:"Store"`
	CTF     string `csv:"CTF"`
	KotH    string `csv:"KotH"`
	Leads   string `csv:"Leads"`
}

func Team(db *ent.Client) (*discordgo.ApplicationCommand, func(s *discordgo.Session, i *discordgo.InteractionCreate)) {
	return &discordgo.ApplicationCommand{
			Name:                     "team",
			Description:              "Submit a CSV of team members",
			DefaultMemberPermissions: &utils.Admin,
			Options: []*discordgo.ApplicationCommandOption{
				{
					Type:        discordgo.ApplicationCommandOptionString,
					Name:        "option",
					Description: "Option of team",
					Required:    true,
					Choices: []*discordgo.ApplicationCommandOptionChoice{
						{
							Name:  "Blue",
							Value: "Blue",
						},
						{
							Name:  "Red",
							Value: "Red",
						},
						{
							Name:  "Black",
							Value: "Black",
						},
						{
							Name:  "White",
							Value: "White",
						},
					},
				},
				{
					Type:        discordgo.ApplicationCommandOptionAttachment,
					Name:        "attachment",
					Description: "Attach CSV",
					Required:    true,
				},
			},
		},
		func(s *discordgo.Session, i *discordgo.InteractionCreate) {
			ctx, span := logging.StartSpan(
				context.Background(),
				"commands.slash.team:Team",
			)
			defer span.End()

			options := optionsByName(i)

			// Get team option from interaction
			teamType := options["option"].StringValue()

			// Get file URL
			fileID := options["attachment"].Value.(string)
			fileURL := i.ApplicationCommandData().Resolved.Attachments[fileID].URL
			logging.Info(ctx, fmt.Sprintf("Received %s", fileURL))

			if !startStatus(ctx, s, i, "Downloading and parsing file...") {
				return
			}

			switch teamType {
			case "Blue":
				importTeam(ctx, s, i, db, fileURL, handleBlue)
			case "Red":
				importTeam(ctx, s, i, db, fileURL, handleRed)
			case "Black":
				importTeam(ctx, s, i, db, fileURL, handleBlack)
			default:
				setStatus(ctx, s, i, fmt.Sprintf("Importing %s team members isn't supported yet.", teamType))
			}
		}
}

// importTeam parses the CSV at fileURL and passes its entries to handle. All
// entries are written in a single transaction so a bad row doesn't leave a
// partial import.
func importTeam[T any](ctx context.Context, s *discordgo.Session, i *discordgo.InteractionCreate, db *ent.Client, fileURL string, handle func(context.Context, *ent.Client, []*T) error) {
	// Populate array of entries from CSV
	entries, err := utils.HandleCSV[T](fileURL)
	if err != nil {
		reportError(ctx, s, i, "Could not parse CSV", err)
		return
	}

	setStatus(ctx, s, i, "Updating database...")
	err = database.WithTx(ctx, db, func(tx *ent.Client) error {
		return handle(ctx, tx, entries)
	})
	if err != nil {
		reportError(ctx, s, i, "Could not add team data", err)
		return
	}

	setStatus(ctx, s, i, "Successfully added team data!")
}

// addMember creates a user with the given username and adds them to the
// team. Empty usernames and users that already exist are skipped.
func addMember(ctx context.Context, db *ent.Client, t *ent.Team, username string) error {
	if username == "" {
		return nil
	}

	// Check if user already exists
	_, err := database.GetUserFromUsername(ctx, db, username)
	if err == nil {
		return nil
	}
	if !ent.IsNotFound(err) {
		return err
	}

	// Create user and add to team
	u, err := database.CreateUser(ctx, db, username)
	if err != nil {
		return err
	}
	return db.Team.UpdateOne(t).AddUser(u).Exec(ctx)
}

// Update DB for Blue teamers
func handleBlue(ctx context.Context, db *ent.Client, entries []*Blue) error {
	for _, entry := range entries {
		// Team number
		num, err := strconv.Atoi(entry.TeamNum)
		if err != nil {
			return fmt.Errorf("invalid team number %q: %w", entry.TeamNum, err)
		}

		// Check/create Blue team
		t, err := database.GetBlue(ctx, db, num)
		if err != nil {
			return err
		}

		// Handle team members
		for _, username := range []string{
			entry.Teammate1,
			entry.Teammate2,
			entry.Teammate3,
			entry.Teammate4,
			entry.Teammate5,
		} {
			if err := addMember(ctx, db, t, username); err != nil {
				return err
			}
		}
	}

	return nil
}

// Update DB for Red teamers
func handleRed(ctx context.Context, db *ent.Client, entries []*Red) error {
	// Check if Red team exists
	t, err := database.GetRed(ctx, db)
	if err != nil {
		return err
	}

	var leads []string
	for _, entry := range entries {
		// Create user from Members column and add them to Red team
		if err := addMember(ctx, db, t, entry.Members); err != nil {
			return err
		}

		// Add user to list of leads
		if entry.Leads != "" {
			leads = append(leads, entry.Leads)
		}
	}

	return database.AddLeads(ctx, db, leads)
}

// Update DB for Black teamers
func handleBlack(ctx context.Context, db *ent.Client, entries []*Black) error {
	// Get Black teams. Subteams with missing roles are still returned, so
	// only stop if no teams came back at all.
	teams, err := database.GetBlack(ctx, db)
	if teams == nil {
		return err
	}
	if err != nil {
		logging.Warn(ctx, fmt.Sprintf("Some Black Team roles are missing: %v", err))
	}

	var leads []string
	for _, entry := range entries {
		// Dereference *Entry
		entryVal := reflect.ValueOf(entry).Elem()
		entryType := entryVal.Type()

		// Iterate over each entry field
		for i := 0; i < entryVal.NumField(); i++ {
			// Get subteam name and the username from the individual entry
			subteam := entryType.Field(i).Name
			username := entryVal.Field(i).String()

			// Check if the 'Leads' field has the username.
			// We don't want to create a duplicate user, so add them
			// to the leads array and continue.
			if subteam == "Leads" {
				leads = append(leads, username)
				continue
			}

			// Get the corresponding pointer to team entity
			t, ok := teams[subteam]
			if !ok {
				return fmt.Errorf("unknown Black subteam %q", subteam)
			}

			if err := addMember(ctx, db, t, username); err != nil {
				return err
			}
		}
	}

	return database.AddLeads(ctx, db, leads)
}
