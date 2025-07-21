package common

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"time"

	_ "github.com/lib/pq"
)

func InitDB(connectTimeout int32) (*sql.DB, error) {
	// Initialise DB connection and wait `connectTimeout` seconds for successful ping

	pgHost := os.Getenv("POSTGRES_HOST")
	if pgHost == "" {
		log.Fatalf("POSTGRES_HOST not found")
	}

	// password would obvs be in a secrets manager
	connStr := fmt.Sprintf("postgres://postgres:postgres@%s/banking?sslmode=disable", pgHost)
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return nil, fmt.Errorf("sql.Open: %w", err)
	}

	for i := connectTimeout; i > 0; i-- {
		err = db.Ping()
		if err == nil {
			log.Println("Conneted to postgres (:")
			return db, nil
		}

		log.Println("Waiting for connection...")
		time.Sleep(time.Second)
	}
	return db, fmt.Errorf("db.Ping: %w", err)
}
