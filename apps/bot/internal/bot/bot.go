package bot

import (
	"os"

	"github.com/bwmarrin/discordgo"
)

type Bot struct {
	Session *discordgo.Session
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
