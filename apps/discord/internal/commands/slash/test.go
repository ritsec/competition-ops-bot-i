package slash

import (
	"context"
	"fmt"

	"github.com/bwmarrin/discordgo"
	"github.com/ritsec/competition-ops-bot-i/ent"
	"github.com/ritsec/competition-ops-bot-i/ent/team"
	"github.com/ritsec/competition-ops-bot-i/internal/logging"
	"github.com/ritsec/competition-ops-bot-i/internal/utils"
)

func Test(db *ent.Client) (*discordgo.ApplicationCommand, func(s *discordgo.Session, i *discordgo.InteractionCreate)) {
	return &discordgo.ApplicationCommand{
			Name:                     "test",
			Description:              "Test COBI function",
			DefaultMemberPermissions: &utils.Admin,
			Options: []*discordgo.ApplicationCommandOption{
				{
					Type:        discordgo.ApplicationCommandOptionString,
					Name:        "option",
					Description: "Option of what to test",
					Required:    true,
					Choices: []*discordgo.ApplicationCommandOptionChoice{
						{
							Name:  "Message",
							Value: "Message",
						},
					},
				},
			},
		},
		func(s *discordgo.Session, i *discordgo.InteractionCreate) {
			ctx, span := logging.StartSpan(
				context.Background(),
				"commands.slash.test:Test",
			)
			defer span.End()

			choice := i.ApplicationCommandData().Options[0].StringValue()

			if !startStatus(ctx, s, i, "Running test...") {
				return
			}

			switch choice {
			case "Message":
				if err := testMessage(ctx, s, db); err != nil {
					reportError(ctx, s, i, "Could not message teams", err)
					return
				}
				setStatus(ctx, s, i, "Sent test message to all Blue Teams!")
			}
		}
}

// testMessage sends a test message to every Blue team's channel to make sure
// COBI can reach them
func testMessage(ctx context.Context, s *discordgo.Session, db *ent.Client) error {
	teams, err := db.Team.Query().
		Where(team.TypeEQ(team.TypeBlue)).
		All(ctx)
	if err != nil {
		return err
	}

	for _, t := range teams {
		// Get team's channel
		channel, err := t.QueryChannel().Only(ctx)
		if err != nil {
			logging.Warn(ctx, fmt.Sprintf("Could not get channel for Blue Team %d: %v", t.Number, err))
			continue
		}

		message := fmt.Sprintf("Hello team %d! This is a test to make sure COBI can reach everyone. Please disregard this message, or say hi to COBI if you so please.", t.Number)
		embed := &discordgo.MessageEmbed{
			Title:       "Test",
			Description: message,
		}

		// Send embed to channel ID
		if _, err := s.ChannelMessageSendEmbed(channel.ID, embed); err != nil {
			logging.Warn(ctx, fmt.Sprintf("Could not send test message to Blue Team %d: %v", t.Number, err))
		}
	}

	return nil
}
