// db/postgres.go
package db

import (
	"database/sql"
	"fmt"
	"os"

	_ "github.com/lib/pq"
)

func ConnectPostgres() error {
	connStr := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=require",
		os.Getenv("PG_HOST"),
		os.Getenv("PG_PORT"),
		os.Getenv("PG_USER"),
		os.Getenv("PG_PASSWORD"),
		os.Getenv("PG_DBNAME"))

	var err error
	PostgresDB, err = sql.Open("postgres", connStr)
	if err != nil {
		return fmt.Errorf("failed to open postgres connection: %w", err)
	}

	if err = PostgresDB.Ping(); err != nil {
		return fmt.Errorf("failed to ping postgres: %w", err)
	}

	return nil
}

func ClosePostgres() error {
	if PostgresDB != nil {
		return PostgresDB.Close()
	}
	return nil
}
