package commands

import (
	"fmt"

	"github.com/bwmarrin/discordgo"
	"github.com/ritsec/competition-ops-bot-i/ent"
	"github.com/ritsec/competition-ops-bot-i/internal/commands/slash"
	"github.com/ritsec/competition-ops-bot-i/internal/config"
	"github.com/ritsec/competition-ops-bot-i/internal/logging"
	"github.com/ritsec/competition-ops-bot-i/internal/utils"
)

type slashFactory func(db *ent.Client) (*discordgo.ApplicationCommand, func(s *discordgo.Session, i *discordgo.InteractionCreate))

var (
	slashCommands = map[string]slashFactory{
		"ssh":     slash.SSH,
		"team":    slash.Team,
		"refresh": slash.Refresh,
		"query":   slash.Query,
		"creds":   slash.Creds,
		"test":    slash.Test,
		// "server": slash.Server,
	}

	// slashCommandHandlers maps a command name to its handler
	slashCommandHandlers = make(map[string]func(s *discordgo.Session, i *discordgo.InteractionCreate))
)

func RegisterSlashCommands(s *discordgo.Session, ctx logging.Context, db *ent.Client) {
	ctx, span := logging.StartSpan(
		ctx,
		"commands.register_commands:registerSlashCommands",
	)
	defer span.End()

	for _, slashFunc := range slashCommands {
		command, handler := slashFunc(db)

		// Create the command in the bot
		_, err := s.ApplicationCommandCreate(config.AppID, config.GuildID, command)
		if err != nil {
			logging.Error(ctx, fmt.Sprintf("Could not register command: %v", command.Name))
		}
		slashCommandHandlers[command.Name] = handler
	}

	// Single entry point that routes every interaction to the right handler
	s.AddHandler(func(s *discordgo.Session, i *discordgo.InteractionCreate) {
		switch i.Type {
		case discordgo.InteractionApplicationCommand:
			if handler, ok := slashCommandHandlers[i.ApplicationCommandData().Name]; ok {
				handler(s, i)
			}
		case discordgo.InteractionModalSubmit:
			utils.HandleComponent(s, i, i.ModalSubmitData().CustomID)
		case discordgo.InteractionMessageComponent:
			utils.HandleComponent(s, i, i.MessageComponentData().CustomID)
		}
	})
}

func RegisterComponents(s *discordgo.Session, ctx logging.Context) {
	ctx, span := logging.StartSpan(
		ctx,
		"commands.register_commands:registerComponents",
	)
	defer span.End()

}
