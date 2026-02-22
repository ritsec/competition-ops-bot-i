package bot

import (
	"log"
	"strconv"
	"strings"

	"github.com/bwmarrin/discordgo"
	"github.com/ritsec/competition-ops-bot-i/ent"
	"github.com/ritsec/competition-ops-bot-i/ent/team"
	"github.com/ritsec/competition-ops-bot-i/internal/utils"
)

type Creds struct {
	TeamType  string `csv:"Team Type"`
	TeamNum   string `csv:"Team Number"`
	Compsole  string `csv:"Compsole (compsole.ritsec.cloud)"`
	Scorify   string `csv:"Scorify (scoring.ists.space)"`
	Authentik string `csv:"Authentik (auth.ists.io)"`
	Store     string `csv:"Store (store.ists.space)"`
	CTFd      string `csv:"CTFd (ctf.ists.space)"`
	Wazuh     string `csv:"Wazuh"`
	PfSense   string `csv:"pfSense"`
	Default   string `csv:"Default Password Linux & Windows"`
	Kali      string `csv:"Kali KoTH Access"`
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
					Choices: []*discordgo.ApplicationCommandOptionChoice{
						{
							Name:  "All",
							Value: "All",
						},
					},
				},
			},
		},
		func(s *discordgo.Session, i *discordgo.InteractionCreate) {
			choice := i.ApplicationCommandData().Options[0].Name

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
				verified := validateAction(s, i)
				if verified {
					if err := b.sendCreds(); err != nil {
						log.Fatal(err)
					}
				}

			}
		}
}

func (b *Bot) handleCreds(entries []*Creds) error {

	for _, entry := range entries {

		var teamType team.Type = team.Type(strings.ToLower(entry.TeamType))
		var t *ent.Team
		var err error

		if teamType == team.TypeBlue {
			// Team number
			teamNum, err := strconv.Atoi(entry.TeamNum)
			if err != nil {
				return err
			}

			t, err = b.Client.Team.
				Query().
				Where(team.And(team.TypeEQ(teamType), team.Number(teamNum))).
				Only(b.ClientCtx)
		} else {
			// The query for teams besides Blue omits the team number field
			t, err = b.Client.Team.
				Query().
				Where(team.TypeEQ(teamType)).
				Only(b.ClientCtx)
		}

		// Get team object from team number
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
				SetStore(entry.Store).
				SetCtfd(entry.CTFd).
				SetWazuh(entry.Wazuh).
				SetPfsense(entry.PfSense).
				SetDefault(entry.Default).
				SetKali(entry.Kali).
				Save(b.ClientCtx)
			if err != nil {
				return err
			}

			// Add credential to team
			t, err = t.Update().
				AddCredential(c).
				Save(b.ClientCtx)
			if err != nil {
				return err
			}
		} else {
			c, err = c.Update().
				SetCompsole(entry.Compsole).
				SetScorify(entry.Scorify).
				SetAuthentik(entry.Authentik).
				SetStore(entry.Store).
				SetCtfd(entry.CTFd).
				SetWazuh(entry.Wazuh).
				SetPfsense(entry.PfSense).
				SetDefault(entry.Default).
				SetKali(entry.Kali).
				Save(b.ClientCtx)
			if err != nil {
				return err
			}
		}

	}
	return nil
}

func (b *Bot) sendCreds() error {
	teams, err := b.Client.Team.Query().
		Where(team.TypeEQ(team.TypeBlue)).
		All(b.ClientCtx)
	if err != nil {
		return err
	}

	for _, team := range teams {
		// Get team's credential
		creds, err := team.QueryCredential().Only(b.ClientCtx)
		if err != nil {
			log.Printf("Could not get credentials for Blue Team %d", team.Number)
			continue
		}

		// Create message embed
		embed := utils.MessageCreds(team.Number, creds)

		// Get team's channel
		channel, err := team.QueryChannel().Only(b.ClientCtx)
		if err != nil {
			log.Printf("Could not get channel for Blue Team %d", team.Number)
			continue
		}

		// Send embed to channel ID
		_, err = b.Session.ChannelMessageSendEmbed(channel.ID, embed)

	}

	return nil
}
