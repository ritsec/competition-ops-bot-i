package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"strconv"

	"github.com/bwmarrin/discordgo"
	"github.com/ritsec/competition-ops-bot-i/internal/bot"
	_ "github.com/ritsec/competition-ops-bot-i/internal/logging"
)

var (
	token           = os.Getenv("DISCORD_TOKEN")
	guildID         = os.Getenv("DISCORD_GUILD")
	appID           = os.Getenv("DISCORD_APP")
	numBlueTeams, _ = strconv.Atoi(os.Getenv("BLUE_TEAMS"))
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	s, err := discordgo.New("Bot " + token)
	if err != nil {
		panic(err)
	}

	s.Identify.Intents = discordgo.MakeIntent(discordgo.IntentsAll)
	s.State.MaxMessageCount = 50

	// Initialize MySQL DB connection before any handler can run
	db := bot.InitDB()

	// Register commands and handlers
	bot.Init(s, ctx, db)
	log.Println("session created")

	// Start the bot session
	if err := s.Open(); err != nil {
		panic(err)
	}
	log.Println("session started")

	// Wait for signal to close session and Ent client
	<-ctx.Done()
	if err := s.Close(); err != nil {
		panic(err)
	}
	if err := db.Close(); err != nil {
		panic(err)
	}
	log.Println("session stopped")
}
