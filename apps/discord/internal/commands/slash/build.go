package slash

import (
	"context"
	"fmt"

	"github.com/bwmarrin/discordgo"
	"github.com/ritsec/competition-ops-bot-i/ent"
	"github.com/ritsec/competition-ops-bot-i/internal/config"
	"github.com/ritsec/competition-ops-bot-i/internal/database"
	"github.com/ritsec/competition-ops-bot-i/internal/logging"
	"github.com/ritsec/competition-ops-bot-i/internal/utils"
)

/*
Prior to running this command, ensure:
1. Roles have been refreshed and all other teams but Blue team has been created
2. No current Blue Team resources exist

I know this is all over the place but I just need to get this working quick
*/

// ptr is a helper function to cleanly return a pointer to a value
func ptr[T any](value T) *T {
	return &value
}

var (
	// Define role parameters for universal Blue Team role
	bluePermissions int64 = 1<<6 | 1<<11 | 1<<14 | 1<<15 | 1<<16 | 1<<26 | 1<<35 | 1<<38
	blueTeamParams        = discordgo.RoleParams{
		Name:        "Blue Team",
		Color:       ptr(2123412),
		Hoist:       ptr(true),
		Permissions: &bluePermissions,
		Mentionable: ptr(true),
	}
)

func Server(db *ent.Client) (*discordgo.ApplicationCommand, func(s *discordgo.Session, i *discordgo.InteractionCreate)) {
	return &discordgo.ApplicationCommand{
			Name:                     "server",
			Description:              "Modify server structure",
			DefaultMemberPermissions: &utils.Admin,
			Options: []*discordgo.ApplicationCommandOption{
				{
					Type:        discordgo.ApplicationCommandOptionSubCommandGroup,
					Name:        "create",
					Description: "Create server resources for a team",
					Options: []*discordgo.ApplicationCommandOption{
						{
							Type:        discordgo.ApplicationCommandOptionSubCommand,
							Name:        "blue",
							Description: "Create Blue team resources",
						},
					},
				},
			},
		},
		func(s *discordgo.Session, i *discordgo.InteractionCreate) {
			ctx, span := logging.StartSpan(
				context.Background(),
				"commands.slash.build:Server",
			)
			defer span.End()

			choice := i.ApplicationCommandData().Options[0].Options[0].Name
			if !startStatus(ctx, s, i, "Building server...") {
				return
			}

			switch choice {
			case "blue":
				if err := buildBlue(ctx, s, db); err != nil {
					reportError(ctx, s, i, "Could not build Blue Team resources", err)
					return
				}
			}

			setStatus(ctx, s, i, "Finished building requested resources!")
		}
}

// buildBlue will build the universal Blue Team role, Blue Team N roles, shared channels and individual channels
func buildBlue(ctx context.Context, s *discordgo.Session, db *ent.Client) error {
	roleID := func(name string) string {
		return database.GetRoleID(ctx, db, name)
	}

	// Check if role already exists
	if roleID("Blue Team") == "" {
		//Create universal Blue role
		blueTeamRole, err := s.GuildRoleCreate(config.GuildID, &blueTeamParams)
		if err != nil {
			return err
		}

		err = db.Role.
			Create().
			SetID(blueTeamRole.ID).
			SetName(blueTeamRole.Name).
			Exec(ctx)
		if err != nil {
			return err
		}

		// Create universal Blue category
		blueTeamCategoryData := discordgo.GuildChannelCreateData{
			Name: "Blue Team",
			Type: discordgo.ChannelTypeGuildCategory,
			PermissionOverwrites: []*discordgo.PermissionOverwrite{
				{ // @everyone
					ID:    config.GuildID,
					Type:  discordgo.PermissionOverwriteTypeRole,
					Allow: 0,
					Deny:  discordgo.PermissionViewChannel,
				},
				{ // Red Team
					ID:    roleID("Red Team"),
					Type:  discordgo.PermissionOverwriteTypeRole,
					Allow: 0,
					Deny:  discordgo.PermissionViewChannel,
				},
				{ // Blue Team
					ID:    blueTeamRole.ID, // hardcoded placeholder for development purposes
					Type:  discordgo.PermissionOverwriteTypeRole,
					Allow: discordgo.PermissionViewChannel,
					Deny:  0,
				},
				{ // Black Team
					ID:    roleID("Black Team"),
					Type:  discordgo.PermissionOverwriteTypeRole,
					Allow: discordgo.PermissionViewChannel,
					Deny:  0,
				},
				{ // White Team
					ID:    roleID("White Team"),
					Type:  discordgo.PermissionOverwriteTypeRole,
					Allow: discordgo.PermissionViewChannel,
					Deny:  0,
				},
				{ // COBI
					ID:    roleID("COBI"),
					Type:  discordgo.PermissionOverwriteTypeRole,
					Allow: discordgo.PermissionViewChannel,
					Deny:  0,
				},
			},
		}
		blueTeamCategory, err := s.GuildChannelCreateComplex(config.GuildID, blueTeamCategoryData)
		if err != nil {
			return err
		}

		// Create universal Blue channels
		err = createTextChannelInCategory(s, blueTeamCategory.ID,
			"blue-announcements",
			"blue-general",
			"blue-questions",
			"blue-shitposting",
		)
		if err != nil {
			return err
		}
	}

	// Iterate over the specified number of Blue Teams
	logging.Info(ctx, fmt.Sprintf("Creating resources for %d Blue Teams", config.NumBlueTeams))
	for i := 1; i <= config.NumBlueTeams; i++ {
		name := fmt.Sprintf("Blue Team %d", i)

		// Check if role already exists
		if roleID(name) == "" {
			textChannel := fmt.Sprintf("blue-%d-chat", i)
			voiceChannel := fmt.Sprintf("blue-%d-voice", i)

			// Create individual Blue roles
			blueTeamNRole, err := s.GuildRoleCreate(config.GuildID, &discordgo.RoleParams{
				Name:        name,
				Color:       ptr(2123412),
				Mentionable: ptr(true),
			})
			if err != nil {
				return err
			}

			err = db.Role.
				Create().
				SetID(blueTeamNRole.ID).
				SetName(blueTeamNRole.Name).
				Exec(ctx)
			if err != nil {
				return err
			}

			blueTeamNCategoryData := discordgo.GuildChannelCreateData{
				Name: name,
				Type: discordgo.ChannelTypeGuildCategory,
				PermissionOverwrites: []*discordgo.PermissionOverwrite{
					{ // @everyone
						ID:    config.GuildID,
						Type:  discordgo.PermissionOverwriteTypeRole,
						Allow: 0,
						Deny:  discordgo.PermissionViewChannel,
					},
					{ // Blue Team N
						ID:    blueTeamNRole.ID,
						Type:  discordgo.PermissionOverwriteTypeRole,
						Allow: discordgo.PermissionViewChannel,
						Deny:  0,
					},
					{ // Black Team
						ID:    roleID("Black Team"),
						Type:  discordgo.PermissionOverwriteTypeRole,
						Allow: discordgo.PermissionViewChannel,
						Deny:  0,
					},
					{ // Leads
						ID:    roleID("Leads"),
						Type:  discordgo.PermissionOverwriteTypeRole,
						Allow: discordgo.PermissionViewChannel,
						Deny:  0,
					},
					{ // COBI
						ID:    roleID("COBI"),
						Type:  discordgo.PermissionOverwriteTypeRole,
						Allow: discordgo.PermissionViewChannel,
						Deny:  0,
					},
				},
			}

			blueTeamNCategory, err := s.GuildChannelCreateComplex(config.GuildID, blueTeamNCategoryData)
			if err != nil {
				return err
			}

			// Create text channels
			err = createTextChannelInCategory(s, blueTeamNCategory.ID,
				"injects-backup",
				textChannel,
			)
			if err != nil {
				return err
			}

			// Create voice channels
			err = createVoiceChannelInCategory(s, blueTeamNCategory.ID,
				voiceChannel,
			)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

// createTextChannelInCategory will take a category ID and a collection of channel names,
// creating text channels under the category
func createTextChannelInCategory(s *discordgo.Session, catID string, channels ...string) error {
	for _, channelName := range channels {
		_, err := s.GuildChannelCreateComplex(config.GuildID, discordgo.GuildChannelCreateData{
			Name:     channelName,
			Type:     discordgo.ChannelTypeGuildText,
			ParentID: catID,
		})
		if err != nil {
			return err
		}
	}

	return nil
}

// createVoiceChannelInCategory will take a category ID and a collection of channel names,
// creating voice channels under the category
func createVoiceChannelInCategory(s *discordgo.Session, catID string, channels ...string) error {
	for _, channelName := range channels {
		_, err := s.GuildChannelCreateComplex(config.GuildID, discordgo.GuildChannelCreateData{
			Name:     channelName,
			Type:     discordgo.ChannelTypeGuildVoice,
			ParentID: catID,
		})
		if err != nil {
			return err
		}
	}

	return nil
}
