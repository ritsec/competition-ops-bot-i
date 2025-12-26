package bot

import (
	"log"
	"os"

	"github.com/bwmarrin/discordgo"
	"github.com/ritsec/competition-ops-bot-i/ent"
	"github.com/ritsec/competition-ops-bot-i/internal/commands/slash"
	"github.com/ritsec/competition-ops-bot-i/internal/connections"
)

type Bot struct {
	Session *discordgo.Session
	Client  *ent.Client
}

var (
	token   = os.Getenv("DISCORD_TOKEN")
	guildID = os.Getenv("DISCORD_GUILD")
	appID   = os.Getenv("DISCORD_APP")

	// SlashCommands is a map of all slash commands
	slashCommands map[string]func() (*discordgo.ApplicationCommand, func(s *discordgo.Session, i *discordgo.InteractionCreate)) = make(map[string]func() (*discordgo.ApplicationCommand, func(s *discordgo.Session, i *discordgo.InteractionCreate)))

	// SlashCommandHandlers is a map of all slash command handlers
	slashCommandHandlers map[string]func(s *discordgo.Session, i *discordgo.InteractionCreate) = make(map[string]func(s *discordgo.Session, i *discordgo.InteractionCreate))
)

// Start will create and set the global session for the bot class
func (b *Bot) Start() {
	session, err := discordgo.New("Bot " + token)
	if err != nil {
		panic(err)
	}

	session.Identify.Intents = discordgo.MakeIntent(discordgo.IntentsAll)
	session.State.MaxMessageCount = 50
	b.Session = session

	// Populate the SlashCommands map
	slashCommands["ssh"] = slash.SSH

	// Register slash commands
	b.registerSlashCommands()

	// Add all handlers
	b.Session.AddHandler(func(
		s *discordgo.Session,
		i *discordgo.InteractionCreate,
	) {
		switch i.Type {
		case discordgo.InteractionApplicationCommand:
			data := i.ApplicationCommandData()

			if command, ok := slashCommandHandlers[data.Name]; ok {
				command(s, i)
			}
		}
	})
}

// InitDB will connect to the MySQL database and set the global Client
func (b *Bot) InitDB() {
	b.Client = connections.Connect()
}

func (b *Bot) registerSlashCommands() {
	for _, slashFunc := range slashCommands {
		command, handler := slashFunc()
		_, err := b.Session.ApplicationCommandCreate(appID, guildID, command)
		if err != nil {
			log.Fatalf("failed registering slash command %s", command.Name)
		}

		slashCommandHandlers[command.Name] = handler
	}
}
