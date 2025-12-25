package main

import (
	"context"
	"log"
	"os"
	"os/signal"

	"github.com/ritsec/competition-ops-bot-i/internal/bot"
)

func main() {
	//
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	// Create the Bot class to hold the global session
	bot := &bot.Bot{}
	bot.Start()
	log.Println("session created")

	// Start the bot session
	if err := bot.Session.Open(); err != nil {
		panic(err)
	}
	log.Println("session started")

	// Wait for signal to stop bot
	<-ctx.Done()
	if err := bot.Session.Close(); err != nil {
		panic(err)
	}
	log.Println("session stopped")
}
