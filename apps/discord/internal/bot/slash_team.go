package bot

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/gocarina/gocsv"
	"github.com/ritsec/competition-ops-bot-i/ent"
	"github.com/ritsec/competition-ops-bot-i/ent/role"
	"github.com/ritsec/competition-ops-bot-i/ent/team"
)

type Entry struct {
	// Blue
	School    string `csv:"School"`
	TeamNum   string `csv:"Team Number"`
	Teammate1 string `csv:"Teammate 1"`
	Teammate2 string `csv:"Teammate 2"`
	Teammate3 string `csv:"Teammate 3"`
	Teammate4 string `csv:"Teammate 4"`
	Teammate5 string `csv:"Teammate 5"`

	// Red
	Members string `csv:"Members"`
	Leads   string `csv:"Leads"`
}

var defaultRole = map[string]string{
	"blue": "Blue Team",
	"red":  "Red Team",
}

func (b *Bot) Team() (*discordgo.ApplicationCommand, func(s *discordgo.Session, i *discordgo.InteractionCreate)) {
	return &discordgo.ApplicationCommand{
			Name:                     "team",
			Description:              "Submit a CSV of team members",
			DefaultMemberPermissions: &Admin,
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
			// Get team option from interaction
			team := i.ApplicationCommandData().Options[0].StringValue()

			// Get file URL
			fileID := i.ApplicationCommandData().Options[1].Value.(string)
			fileURL := i.ApplicationCommandData().Resolved.Attachments[fileID].URL
			log.Printf("received %s\n", fileURL)

			// Populate array of entries from CSV
			initialMessage(s, i, "Downloading and parsing file...")
			entries, err := fileHandler(fileURL)
			if err != nil {
				log.Fatal(err)
			}

			// Handle according to command option
			updateMessage(s, i, "Updating database...")
			switch team {
			case "Blue":
				b.handleBlue(entries)
			case "Red":
				b.handleRed(entries)
			}

			updateMessage(s, i, "Successfully added team data!")
		}
}

// Query DB for Blue teamers
func (b *Bot) handleBlue(entries []*Entry) error {

	for _, entry := range entries {
		num, err := strconv.Atoi(entry.TeamNum)
		if err != nil {
			return err
		}

		// Check if team already exists
		t, err := b.Client.Team.
			Query().
			Where(team.Number(num)).
			Only(b.ClientCtx)
		if err != nil { // Create team if it doesn't exist
			log.Printf("creating team %d", num)

			t, err = b.Client.Team.
				Create().
				SetType("blue").
				SetNumber(num).
				Save(b.ClientCtx)
			if err != nil {
				return err
			}

		}
		teamRole := fmt.Sprintf("Blue Team %d", num)
		roles := []string{
			defaultRole["blue"],
			teamRole,
		}
		if err := b.addRoles(t, roles...); err != nil {
			log.Fatal(err)
		}

		// Handle team members
		for _, username := range []string{
			entry.Teammate1,
			entry.Teammate2,
			entry.Teammate3,
			entry.Teammate4,
			entry.Teammate5,
		} {
			// Create user and add to team
			u, err := b.Client.User.
				Create().
				SetUsername(username).
				Save(b.ClientCtx)
			if err != nil {
				return err
			}
			t.Update().AddUser(u).Save(b.ClientCtx)
		}
		log.Println(entry)
	}

	return nil
}

// Query DB for Red teamers
func (b *Bot) handleRed(entries []*Entry) error {

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
	}

	roles := []string{
		defaultRole["red"],
	}
	if err := b.addRoles(t, roles...); err != nil {
		log.Fatal(err)
	}

	var leads []string
	for _, entry := range entries {
		// Create user from Members column and add them to Red team
		if entry.Members != "" {
			u, err := b.Client.User.
				Create().
				SetUsername(entry.Members).
				Save(b.ClientCtx)
			if err != nil {
				return err
			}
			t.Update().AddUser(u).Save(b.ClientCtx)
		}
		// Create user from Leads column and add them to Red team
		if entry.Leads != "" {
			u, err := b.Client.User.
				Create().
				SetUsername(entry.Leads).
				Save(b.ClientCtx)
			if err != nil {
				return err
			}
			t.Update().AddUser(u).Save(b.ClientCtx)
			leads = append(leads, entry.Leads) // Add to leads array
		}
	}
	t.Update().SetLead(strings.Join(leads, ",")) // Set leads field of team to be a comma separated list

	return nil
}

func fileHandler(URL string) ([]*Entry, error) {
	// Create temp file
	filename := fmt.Sprintf("download-%s.csv", time.Now())
	temp, err := os.CreateTemp("/tmp", filename)
	if err != nil {
		return nil, err
	}
	defer os.Remove(temp.Name())
	defer temp.Close()

	// Download file
	resp, err := http.Get(URL)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("error getting file")
	}
	defer resp.Body.Close()

	// Copy file content to temp file
	_, err = io.Copy(temp, resp.Body)
	if err != nil {
		return nil, err
	}

	// Set offset to beginning of file
	if _, err := temp.Seek(0, 0); err != nil {
		return nil, err
	}

	// Unmarshal file based on Entry struct
	var entries []*Entry
	if unmarshalError := gocsv.UnmarshalFile(temp, &entries); unmarshalError != nil {
		panic(unmarshalError)
	}

	return entries, nil
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
