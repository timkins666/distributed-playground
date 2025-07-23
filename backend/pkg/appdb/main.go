package appdb

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/lib/pq"
	cmn "github.com/timkins666/distributed-playground/backend/pkg/common"
)

type DBConfig struct {
	User           string
	Password       string
	DBName         string
	Host           string
	ConnectTimeout int
}

func (c *DBConfig) ConnectionString() string {
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

func InitDB(conf DBConfig) (*DB, error) {
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
			return &DB{db: db}, nil
		}

		log.Println("Waiting for connection...")
		continue
	}

	return nil, err
}

type DB struct {
	db *sql.DB
}

func (db *DB) Expose() *sql.DB {
	// return the underlying sql.DB
	return db.db
}

func (db *DB) CreateUser(user *cmn.User) error {
	// Creates the user in the db. User.ID will be populated in the struct.
	return db.db.QueryRow(`
		INSERT INTO accounts."user" (username, roles) VALUES ($1, $2) RETURNING id
	`, user.Username, pq.Array(user.Roles)).Scan(&user.ID)
}

func (db *DB) LoadUserByName(username string) (cmn.User, error) {
	// load user by name from db.
	// searches case insensitively, returns userame casing as in db

	log.Printf("Try load user %s from db...", username)
	log.Printf("db is %+v", db)
	log.Printf("db.db is %+v", db.db)
	var user cmn.User
	err := db.db.QueryRow(`
		SELECT id, username, roles FROM accounts."user" WHERE LOWER(username) = LOWER($1)
	`, username).Scan(&user.ID, &user.Username, pq.Array(&user.Roles))
	return user, err
}

func (db *DB) LoadUserByID(userID int) (cmn.User, error) {
	log.Printf("Try load user if %d from db...", userID)
	var user cmn.User
	err := db.db.QueryRow(`
		SELECT id, username, roles FROM accounts."user" WHERE id = $1
	`, userID).Scan(&user.ID, &user.Username, pq.Array(&user.Roles))
	return user, err
}

func (db *DB) GetUserAccounts(userID int) ([]cmn.Account, error) {
	// get accounts from the user from db

	// TODO: redis
	// TOFO: squirrel / sqlx

	var accounts []cmn.Account

	rows, err := db.db.Query(`
		SELECT id, user_id, balance FROM accounts.account WHERE user_id = $1
	`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var acc cmn.Account
		err := rows.Scan(&acc.AccountID, &acc.UserID, &acc.Balance)
		if err != nil {
			return nil, err
		}
		accounts = append(accounts, acc)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return accounts, nil
}

func (db *DB) GetAccountByID(accountID int) (*cmn.Account, error) {
	// get account matching id.

	// TODO: redis
	// TOFO: squirrel / sqlx

	acc := cmn.Account{}

	err := db.db.QueryRow(`
		SELECT id, user_id, balance from accounts.account WHERE id = $1
	`, accountID).Scan(&acc.AccountID, &acc.UserID, &acc.Balance)
	if err != nil {
		return nil, err
	}
	return &acc, nil
}

func (db *DB) CreateAccount(a cmn.Account) (int, error) {
	var newAccID int
	err := db.db.QueryRow(`
		INSERT INTO accounts.account (user_id, name)
		VALUES ($1, $2)
		RETURNING id
		`, a.UserID, a.Name).Scan(&newAccID)
	return newAccID, err
}
