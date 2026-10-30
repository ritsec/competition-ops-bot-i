package database

import (
	"context"
	"fmt"
	"log"
	"os"

	_ "github.com/go-sql-driver/mysql"
	"github.com/ritsec/competition-ops-bot-i/ent"
)

var (
	dbUser = os.Getenv("DB_USER")
	dbPass = os.Getenv("DB_PASS")
	dbHost = os.Getenv("DB_HOST")
	dbPort = os.Getenv("DB_PORT")
	dbName = os.Getenv("DB_DATABASE")
)

// Connect opens the MySQL connection, runs migrations, and returns the Ent
// client. The caller owns the client and is responsible for closing it.
func Connect() *ent.Client {
	connectionString := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?parseTime=True", dbUser, dbPass, dbHost, dbPort, dbName)
	client, err := ent.Open("mysql", connectionString)
	if err != nil {
		log.Fatalf("failed opening connection to mysql: %v", err)
	}

	// Run the auto migration tool. This is a one-time startup task, so a
	// background context is appropriate here.
	if err := client.Schema.Create(context.Background()); err != nil {
		log.Fatalf("failed creating schema resources: %v", err)
	}
	log.Println("connected to mysql")

	return client
}
