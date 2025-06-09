package db

import (
	"database/sql"
	"log"
	"os"
)

var (
	Store      *ValkeyChatStore
	PostgresDB *sql.DB
)

func DB_init() {
	valkey_endpoint := os.Getenv("VALKEY_ENDPOINT")
	valkey_password := os.Getenv("VALKEY_PASSWORD")
	Store = NewValkeyChatStore(valkey_endpoint, valkey_password, 1) //this `1` is for the room information partation

	if err := ConnectPostgres(); err != nil {
		log.Fatalf("Failed to connect to PostgreSQL: %v", err)
	}
}
