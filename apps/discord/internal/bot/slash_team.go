package bot

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/gocarina/gocsv"
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

func (b *Bot) Team() (*discordgo.ApplicationCommand, func(s *discordgo.Session, i *discordgo.InteractionCreate)) {
	return &discordgo.ApplicationCommand{
			Name:        "team",
			Description: "Submit a CSV of team members",
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
			entries, err := fileHandler(fileURL)
			if err != nil {
				log.Fatal(err)
			}

			// Handle according to command option
			switch team {
			case "Blue":
				b.handleBlue(entries)
			}
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
		// Handle team members
		for _, username := range []string{
			entry.Teammate1,
			entry.Teammate2,
			entry.Teammate3,
			entry.Teammate4,
			entry.Teammate5,
		} {
			// Create user and add to team
			_, err := b.Client.User.
				Create().
				SetUsername(username).
				AddTeam(t).
				Save(b.ClientCtx)
			if err != nil {
				return err
			}
		}
		log.Println(entry)
	}

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

	var entries []*Entry
	if unmarshalError := gocsv.UnmarshalFile(temp, &entries); unmarshalError != nil {
		panic(unmarshalError)
	}

	return entries, nil
}
