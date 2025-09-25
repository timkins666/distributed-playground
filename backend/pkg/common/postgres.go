package common

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"time"
)

type DBConfig struct {
	User           string
	Password       string
	DBName         string
	Host           string
	ConnectTimeout int32
}

func (c DBConfig) ConnectionString() string {
	return fmt.Sprintf(
		"postgres://%s:%s@%s/%s?sslmode=disable",
		c.User,
		c.Password,
		c.Host,
		c.DBName,
	)
}

var DefaultConfig DBConfig = DBConfig{
	User:           "postgres",
	Password:       "postgres",
	DBName:         "banking",
	Host:           os.Getenv("POSTGRES_HOST"),
	ConnectTimeout: 10,
}

func InitPostgres(conf DBConfig) (*sql.DB, error) {
	// Initialise Postgres connection

	connStr := conf.ConnectionString()
	var (
		db     *sql.DB
		err    error
		isOpen bool
	)

	for i := conf.ConnectTimeout; i > 0; i-- {
		if i != conf.ConnectTimeout {
			time.Sleep(time.Second)
		}

		if !isOpen {
			db, err = sql.Open("postgres", connStr)
			if err == nil {
				isOpen = true
			} else {
				continue
			}
		}

		err = db.Ping()
		if err == nil {
			log.Println("Connected to postgres (:")
			return db, nil
		}

		log.Println("Waiting for connection...")
		continue
	}

	return nil, err
}
