package bot

import (
	"os"

	"github.com/bwmarrin/discordgo"
	"github.com/ritsec/competition-ops-bot-i/ent"
	"github.com/ritsec/competition-ops-bot-i/internal/connections"
)

type Bot struct {
	Session *discordgo.Session
	Client  *ent.Client
}

// Start will create and set the global session for the bot class
func (b *Bot) Start() {
	session, err := discordgo.New("Bot " + os.Getenv("DISCORD_TOKEN"))
	if err != nil {
		panic(err)
	}

	session.Identify.Intents = discordgo.MakeIntent(discordgo.IntentsAll)

	session.State.MaxMessageCount = 50

	b.Session = session
}

// InitDB will connect to the MySQL database and set the global Client
func (b *Bot) InitDB() {
	b.Client = connections.Connect()
}
