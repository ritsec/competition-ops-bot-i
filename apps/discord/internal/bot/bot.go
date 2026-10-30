package bot

import (
	"github.com/bwmarrin/discordgo"
	"github.com/ritsec/competition-ops-bot-i/ent"
	"github.com/ritsec/competition-ops-bot-i/internal/commands"
	"github.com/ritsec/competition-ops-bot-i/internal/database"
	"github.com/ritsec/competition-ops-bot-i/internal/logging"
)

// InitDB will connect to the MySQL database and return the Ent client
func InitDB() *ent.Client {
	return database.Connect()
}

func Init(s *discordgo.Session, ctx logging.Context, db *ent.Client) {
	commands.RegisterSlashCommands(s, ctx, db)
	s.AddHandler(Join(db))
}
