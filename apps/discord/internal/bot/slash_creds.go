package bot

import (
	"log"
	"strconv"

	"github.com/bwmarrin/discordgo"
	"github.com/ritsec/competition-ops-bot-i/ent/team"
	"github.com/ritsec/competition-ops-bot-i/internal/utils"
)

type Creds struct {
	TeamNum   string `csv:"Team Number"`
	Compsole  string `csv:"Compsole"`
	Scorify   string `csv:"Scorify"`
	Authentik string `csv:"Authentik"`
}

func (b *Bot) Creds() (*discordgo.ApplicationCommand, func(s *discordgo.Session, i *discordgo.InteractionCreate)) {
	return &discordgo.ApplicationCommand{
			Name:                     "creds",
			Description:              "Submit a CSV of credentials",
			DefaultMemberPermissions: &Admin,
			Options: []*discordgo.ApplicationCommandOption{
				// TODO: Add option to query credentials
				{
					Type:        discordgo.ApplicationCommandOptionAttachment,
					Name:        "process",
					Description: "Process credentials CSV.",
				},
				{
					Type:        discordgo.ApplicationCommandOptionString,
					Name:        "send",
					Description: "Send credentials to team channels.",
				},
			},
		},
		func(s *discordgo.Session, i *discordgo.InteractionCreate) {
			choice := i.ApplicationCommandData().Options[0].Name
			log.Println(choice)

			switch choice {
			case "process":
				// Get file URL
				fileID := i.ApplicationCommandData().Options[0].Value.(string)
				fileURL := i.ApplicationCommandData().Resolved.Attachments[fileID].URL
				log.Printf("received %s\n", fileURL)

				// Populate array of entries from CSV
				initialMessage(s, i, "Downloading and parsing file...")

				entries, err := utils.HandleCSV[Creds](fileURL)
				if err != nil {
					log.Fatal(err)
				}

				updateMessage(s, i, "Adding credentials to databases...")
				if err := b.handleCreds(entries); err != nil {
					log.Fatal(err)
				}

				updateMessage(s, i, "Successfully added team data!")
			case "send":

			}
		}
}

func (b *Bot) handleCreds(entries []*Creds) error {

	for _, entry := range entries {
		// Team number
		num, err := strconv.Atoi(entry.TeamNum)
		if err != nil {
			return err
		}

		// Get team object from team number
		t, err := b.Client.Team.
			Query().
			Where(team.And(team.TypeEQ(team.TypeBlue), team.Number(num))).
			Only(b.ClientCtx)
		if err != nil {
			return err
		}
		log.Println(t)

		// Check if credentials entry exists
		c, err := t.QueryCredential().Only(b.ClientCtx)
		if err != nil {
			// Create credentials entry
			c, err = b.Client.Credential.
				Create().
				SetCompsole(entry.Compsole).
				SetScorify(entry.Scorify).
				SetAuthentik(entry.Authentik).
				Save(b.ClientCtx)
			if err != nil {
				return err
			}
		}

		// Add credential to team
		t, err = t.Update().
			AddCredential(c).
			Save(b.ClientCtx)
		if err != nil {
			return err
		}
	}
	return nil
}
