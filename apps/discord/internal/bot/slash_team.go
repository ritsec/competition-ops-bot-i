package bot

import (
	"log"
	"strconv"
	"strings"

	"github.com/bwmarrin/discordgo"
	"github.com/ritsec/competition-ops-bot-i/internal/utils"
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

	// Black
	Infra   string `csv:"Infra"`
	Linux   string `csv:"Linux"`
	Windows string `csv:"Windows"`
	Scoring string `csv:"Scoring"`
	Logging string `csv:"Logging"`
	Store   string `csv:"Store"`
	CTF     string `csv:"CTF"`
	KotH    string `csv:"KotH"`
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
			entries, err := utils.HandleCSV[Entry](fileURL)
			if err != nil {
				log.Fatal(err)
			}

			// Handle according to command option
			updateMessage(s, i, "Updating database...")
			switch team {
			case "Blue": // TODO: handle errors
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
		// Team number
		num, err := strconv.Atoi(entry.TeamNum)
		if err != nil {
			return err
		}

		// Check/create Blue team
		t, err := b.getBlue(num)
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
			if username != "" {
				// Create user and add to team
				u, err := b.createUser(username)
				if err != nil {
					return err
				}
				t.Update().AddUser(u).Save(b.ClientCtx)
			}
		}
		log.Println(entry)
	}

	return nil
}

// Query DB for Red teamers
func (b *Bot) handleRed(entries []*Entry) error {

	// Check if Red team exists
	t, err := b.getRed()
	if err != nil {
		return err
	}

	var leads []string
	for _, entry := range entries {
		username := entry.Members

		// Create user from Members column and add them to Red team
		if username != "" {
			u, err := b.createUser(username)
			if err != nil {
				return err
			}
			t.Update().AddUser(u).Save(b.ClientCtx)
		}
		// Add user to list of leads
		if entry.Leads != "" {
			leads = append(leads, entry.Leads) // Add to leads array
		}
	}
	t.Update().SetLead(strings.Join(leads, ",")) // Set leads field of team to be a comma separated list

	return nil
}

// Update DB for Black teamers
func (b *Bot) handleBlack(entries []*Entry) error {

	// Get Black teams
	teams, err := b.getBlack()
	if err != nil {
		return err
	}

	// var leads []string
	var username string
	for _, entry := range entries {
		if entry.Infra != "" {
			username = entry.Infra

			u, err := b.createUser(username)
			if err != nil {
				return err
			}
			teams["Infra"].Update().AddUser(u).Save(b.ClientCtx)
		}
	}
}
