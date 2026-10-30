package slash

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/bwmarrin/discordgo"
	"github.com/ritsec/competition-ops-bot-i/ent"
	"github.com/ritsec/competition-ops-bot-i/ent/team"
	"github.com/ritsec/competition-ops-bot-i/internal/database"
	"github.com/ritsec/competition-ops-bot-i/internal/logging"
	"github.com/ritsec/competition-ops-bot-i/internal/utils"
)

type Cred struct {
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

func Creds(db *ent.Client) (*discordgo.ApplicationCommand, func(s *discordgo.Session, i *discordgo.InteractionCreate)) {
	return &discordgo.ApplicationCommand{
			Name:                     "creds",
			Description:              "Submit a CSV of credentials",
			DefaultMemberPermissions: &utils.Admin,
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
			ctx, span := logging.StartSpan(
				context.Background(),
				"commands.slash.creds:Creds",
			)
			defer span.End()

			// Both options are optional, so the user may have picked neither
			options := i.ApplicationCommandData().Options
			if len(options) == 0 {
				startStatus(ctx, s, i, "Please choose either process or send.")
				return
			}

			switch options[0].Name {
			case "process":
				processCreds(ctx, s, i, db, options[0])
			case "send":
				if !utils.ValidateAction(s, i) {
					return
				}

				if err := sendCreds(ctx, s, db); err != nil {
					logging.Error(ctx, fmt.Sprintf("Could not send credentials: %v", err))
				}
			}
		}
}

// processCreds downloads the attached CSV and stores its credentials
func processCreds(ctx context.Context, s *discordgo.Session, i *discordgo.InteractionCreate, db *ent.Client, option *discordgo.ApplicationCommandInteractionDataOption) {
	// Get file URL
	fileID := option.Value.(string)
	fileURL := i.ApplicationCommandData().Resolved.Attachments[fileID].URL
	logging.Info(ctx, fmt.Sprintf("Received %s", fileURL))

	if !startStatus(ctx, s, i, "Downloading and parsing file...") {
		return
	}

	// Populate array of entries from CSV
	entries, err := utils.HandleCSV[Cred](fileURL)
	if err != nil {
		reportError(ctx, s, i, "Could not parse CSV", err)
		return
	}

	setStatus(ctx, s, i, "Adding credentials to database...")
	if err := handleCreds(ctx, db, entries); err != nil {
		reportError(ctx, s, i, "Could not add credentials", err)
		return
	}

	setStatus(ctx, s, i, "Successfully added credentials!")
}

// handleCreds adds or updates the credentials for each entry. Entries are
// written in a single transaction so a bad row doesn't leave a partial import.
func handleCreds(ctx context.Context, db *ent.Client, entries []*Cred) error {
	return database.WithTx(ctx, db, func(tx *ent.Client) error {
		for _, entry := range entries {
			if err := upsertCred(ctx, tx, entry); err != nil {
				return err
			}
		}
		return nil
	})
}

// upsertCred creates or updates the credential for the team in entry
func upsertCred(ctx context.Context, db *ent.Client, entry *Cred) error {
	teamType := team.Type(strings.ToLower(entry.TeamType))
	if err := team.TypeValidator(teamType); err != nil {
		return fmt.Errorf("invalid team type %q: %w", entry.TeamType, err)
	}

	query := db.Team.Query().Where(team.TypeEQ(teamType))

	// Only Blue teams are numbered, other team types are unique by type
	if teamType == team.TypeBlue {
		teamNum, err := strconv.Atoi(entry.TeamNum)
		if err != nil {
			return fmt.Errorf("invalid team number %q: %w", entry.TeamNum, err)
		}
		query = query.Where(team.Number(teamNum))
	}

	t, err := query.Only(ctx)
	if err != nil {
		return fmt.Errorf("could not find %s team %s: %w", teamType, entry.TeamNum, err)
	}

	// Check if credentials entry exists
	c, err := t.QueryCredential().Only(ctx)
	switch {
	case ent.IsNotFound(err):
		return db.Credential.
			Create().
			SetTeam(t).
			SetCompsole(entry.Compsole).
			SetScorify(entry.Scorify).
			SetAuthentik(entry.Authentik).
			SetStore(entry.Store).
			SetCtfd(entry.CTFd).
			SetWazuh(entry.Wazuh).
			SetPfsense(entry.PfSense).
			SetDefault(entry.Default).
			SetKali(entry.Kali).
			Exec(ctx)
	case err != nil:
		return err
	default:
		return c.Update().
			SetCompsole(entry.Compsole).
			SetScorify(entry.Scorify).
			SetAuthentik(entry.Authentik).
			SetStore(entry.Store).
			SetCtfd(entry.CTFd).
			SetWazuh(entry.Wazuh).
			SetPfsense(entry.PfSense).
			SetDefault(entry.Default).
			SetKali(entry.Kali).
			Exec(ctx)
	}
}

// sendCreds sends each Blue team its credentials in its team channel
func sendCreds(ctx context.Context, s *discordgo.Session, db *ent.Client) error {
	teams, err := db.Team.Query().
		Where(team.TypeEQ(team.TypeBlue)).
		All(ctx)
	if err != nil {
		return err
	}

	for _, t := range teams {
		// Get team's credential
		creds, err := t.QueryCredential().Only(ctx)
		if err != nil {
			logging.Warn(ctx, fmt.Sprintf("Could not get credentials for Blue Team %d: %v", t.Number, err))
			continue
		}

		// Create message embed
		embed := utils.MessageCreds(t.Number, creds)

		// Get team's channel
		channel, err := t.QueryChannel().Only(ctx)
		if err != nil {
			logging.Warn(ctx, fmt.Sprintf("Could not get channel for Blue Team %d: %v", t.Number, err))
			continue
		}

		// Send embed to channel ID
		if _, err := s.ChannelMessageSendEmbed(channel.ID, embed); err != nil {
			logging.Warn(ctx, fmt.Sprintf("Could not send credentials to Blue Team %d: %v", t.Number, err))
		}
	}

	return nil
}
