package connections

import (
	"context"
	"fmt"
	"log"
	"os"

	_ "github.com/go-sql-driver/mysql"
	_ "github.com/mattn/go-sqlite3"
	"github.com/ritsec/competition-ops-bot-i/ent"
)

var (
	user     = os.Getenv("DB_USER")
	pass     = os.Getenv("DB_PASS")
	host     = os.Getenv("DB_HOST")
	port     = os.Getenv("DB_PORT")
	database = os.Getenv("DB_DATABASE")
)

func Connect() (*ent.Client, context.Context) {
	connectionString := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?parseTime=True", user, pass, host, port, database)
	client, err := ent.Open("mysql", connectionString)
	if err != nil {
		log.Fatalf("failed opening connection to mysql: %v", err)
	}

	ctx := context.Background()
	// Run the auto migration tool.
	if err := client.Schema.Create(ctx); err != nil {
		log.Fatalf("failed creating schema resources: %v", err)
	}
	log.Println("connected to mysql")

	return client, ctx
}
