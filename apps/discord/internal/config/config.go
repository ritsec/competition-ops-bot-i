package config

import (
	"os"
	"strconv"
)

var (
	Token           = os.Getenv("DISCORD_TOKEN")
	GuildID         = os.Getenv("DISCORD_GUILD")
	AppID           = os.Getenv("DISCORD_APP")
	NumBlueTeams, _ = strconv.Atoi(os.Getenv("BLUE_TEAMS"))

	// For the future, COBI should handle server setup as well.
)
