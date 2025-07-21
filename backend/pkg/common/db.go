package common

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"time"

	_ "github.com/lib/pq"
)

func InitDB() (*sql.DB, error) {
	pgHost := os.Getenv("POSTGRES_HOST")
	if pgHost == "" {
		log.Fatalf("POSTGRES_HOST not found")
	}

	// password would obvs be in a secrets manager
	connStr := fmt.Sprintf("postgres://me:me@%s/mydb?sslmode=disable", pgHost)
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return nil, fmt.Errorf("sql.Open: %w", err)
	}

	i := 10
	for {
		err := db.Ping()
		if err == nil {
			return db, nil
		}
		i--
		if i == 0 {
			return nil, fmt.Errorf("db.Ping: %w", err)
		}

		log.Println("Waiting for postgres...")
		time.Sleep(time.Second)
	}
}
