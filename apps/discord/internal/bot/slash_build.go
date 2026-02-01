package bot

import (
	"fmt"
	"log"

	"github.com/bwmarrin/discordgo"
)

/*
Prior to running this command, ensure:
1. Roles have been refreshed
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

func (b *Bot) Server() (*discordgo.ApplicationCommand, func(s *discordgo.Session, i *discordgo.InteractionCreate)) {
	return &discordgo.ApplicationCommand{
			Name:                     "server",
			Description:              "Modify server structure",
			DefaultMemberPermissions: &Admin,
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
							// Choices: []*discordgo.ApplicationCommandOptionChoice{
							// 	{
							// 		Name:  "Blue",
							// 		Value: "blue",
							// 	},
							// },
						},
					},
				},
			},
		},
		func(s *discordgo.Session, i *discordgo.InteractionCreate) {
			choice := i.ApplicationCommandData().Options[0].Options[0].Name
			initialMessage(s, i, "Building server...")

			switch choice {
			case "blue":
				if err := b.buildBlue(); err != nil {
					log.Println(err)
				}
			}

			updateMessage(s, i, "Finished building requested resources!")
		}
}

// buildBlue will build the universal Blue Team role, Blue Team N roles, shared channels and individual channels
func (b *Bot) buildBlue() error {
	// Create universal Blue role
	// blueTeamRole, err := b.Session.GuildRoleCreate(guildID, &blueTeamParams)
	// if err != nil {
	// 	return err
	// }

	// Create universal Blue category
	blueTeamCategoryData := discordgo.GuildChannelCreateData{
		Name: "Blue Team",
		Type: discordgo.ChannelTypeGuildCategory,
		PermissionOverwrites: []*discordgo.PermissionOverwrite{
			{ // @everyone
				ID:    guildID,
				Type:  discordgo.PermissionOverwriteTypeRole,
				Allow: 0,
				Deny:  discordgo.PermissionViewChannel,
			},
			{ // Red Team
				ID:    b.getRole("Red Team"),
				Type:  discordgo.PermissionOverwriteTypeRole,
				Allow: 0,
				Deny:  discordgo.PermissionViewChannel,
			},
			{ // Blue Team
				ID:    "1467593670873321737", // hardcoded placeholder for development purposes
				Type:  discordgo.PermissionOverwriteTypeRole,
				Allow: discordgo.PermissionViewChannel,
				Deny:  0,
			},
			{ // Black Team
				ID:    b.getRole("Black Team"),
				Type:  discordgo.PermissionOverwriteTypeRole,
				Allow: discordgo.PermissionViewChannel,
				Deny:  0,
			},
			{ // White Team
				ID:    b.getRole("White Team"),
				Type:  discordgo.PermissionOverwriteTypeRole,
				Allow: discordgo.PermissionViewChannel,
				Deny:  0,
			},
			{ // COBI
				ID:    b.getRole("COBI"),
				Type:  discordgo.PermissionOverwriteTypeRole,
				Allow: discordgo.PermissionViewChannel,
				Deny:  0,
			},
		},
	}
	blueTeamCategory, err := b.Session.GuildChannelCreateComplex(guildID, blueTeamCategoryData)
	if err != nil {
		return err
	}

	// Create universal Blue channels
	b.createTextChannelInCategory(blueTeamCategory.ID,
		"blue-announcements",
		"blue-general",
		"blue-questions",
		"blue-shitposting",
	)

	// Iterate over the specified number of Blue Teams
	log.Printf("Creating resources for %d Blue Teams", numBlueTeams)
	for i := 1; i <= numBlueTeams; i++ {
		name := fmt.Sprintf("Blue Team %d", i)
		textChannel := fmt.Sprintf("blue-%d-chat", i)
		voiceChannel := fmt.Sprintf("blue-%d-voice", i)

		// Create individual Blue roles
		blueTeamNRole, err := b.Session.GuildRoleCreate(guildID, &discordgo.RoleParams{
			Name:        name,
			Color:       ptr(2123412),
			Mentionable: ptr(true),
		})
		if err != nil {
			return err
		}

		blueTeamNCategoryData := discordgo.GuildChannelCreateData{
			Name: name,
			Type: discordgo.ChannelTypeGuildCategory,
			PermissionOverwrites: []*discordgo.PermissionOverwrite{
				{ // @everyone
					ID:    guildID,
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
					ID:    b.getRole("Black Team"),
					Type:  discordgo.PermissionOverwriteTypeRole,
					Allow: discordgo.PermissionViewChannel,
					Deny:  0,
				},
				{ // Leads
					ID:    b.getRole("Leads"),
					Type:  discordgo.PermissionOverwriteTypeRole,
					Allow: discordgo.PermissionViewChannel,
					Deny:  0,
				},
				{ // COBI
					ID:    b.getRole("COBI"),
					Type:  discordgo.PermissionOverwriteTypeRole,
					Allow: discordgo.PermissionViewChannel,
					Deny:  0,
				},
			},
		}

		blueTeamNCategory, err := b.Session.GuildChannelCreateComplex(guildID, blueTeamNCategoryData)
		if err != nil {
			return err
		}

		// Create text channels
		b.createTextChannelInCategory(blueTeamNCategory.ID,
			"injects-backup",
			textChannel,
		)

		// Create voice channels
		b.createVoiceChannelInCategory(blueTeamNCategory.ID,
			voiceChannel,
		)
	}

	return nil
}

// createTextChannelInCategory will take a category ID and a collection of channel names,
// creating text channels under the category
func (b *Bot) createTextChannelInCategory(catID string, channels ...string) error {
	for _, channelName := range channels {
		_, err := b.Session.GuildChannelCreateComplex(guildID, discordgo.GuildChannelCreateData{
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
// creating text channels under the category
func (b *Bot) createVoiceChannelInCategory(catID string, channels ...string) error {
	for _, channelName := range channels {
		_, err := b.Session.GuildChannelCreateComplex(guildID, discordgo.GuildChannelCreateData{
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
