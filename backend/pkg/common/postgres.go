package common

import (
	"database/sql"
	"errors"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/lib/pq"
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

func InitDB(conf DBConfig) (DBAll, error) {
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

type DBAll interface {
	Expose() *sql.DB
	CreateUser(*User) (int32, error)
	LoadUserByName(string) (User, error)
	LoadUserByID(int32) (User, error)
	GetUserAccounts(int32) ([]Account, error)
	CreateAccount(Account) (int32, error)
	GetAccountByID(int32) (*Account, error)
	CreatePayment(*PaymentRequest) error
	UpdatePaymentStatus(string, string) error
	CommitTransaction(*Transaction) error
}

type DB struct {
	db *sql.DB
}

func (db *DB) CreatePayment(pr *PaymentRequest) error {
	// TODO: check affected row count == 1
	_, err := db.db.Exec(`
	INSERT INTO payments.transfer (
		system_id,
		app_id,
		source_account_id,
		target_account_id,
		amount,
		status
		)
		VALUES ($1, $2, $3, $4, $5, "PENDING")
		`, pr.SystemID, pr.AppID, pr.SourceAccountID, pr.TargetAccountID, pr.Amount)

	return err
}

func (db *DB) UpdatePaymentStatus(sysId, status string) error {
	// TODO: check affected row count == 1
	_, err := db.db.Exec(`
		UPDATE payments.transfer
		SET status = $1
		WHERE system_id = $2
	`, status, sysId)

	return err
}

// return the underlying sql.DB for query prototyping
func (db *DB) Expose() *sql.DB {
	return db.db
}

// Creates the user in the db, returning the new user id
func (db *DB) CreateUser(user *User) (int32, error) {
	userID := int32(0)
	err := db.db.QueryRow(`
		INSERT INTO accounts."user" (username, roles) VALUES ($1, $2) RETURNING id
	`, user.Username, pq.Array(user.Roles)).Scan(&userID)
	return userID, err
}

// load user by name from db.
func (db *DB) LoadUserByName(username string) (User, error) {
	// searches case insensitively, returns userame casing as in db

	log.Printf("Try load user %s from db...", username)
	log.Printf("db is %+v", db)
	log.Printf("db.db is %+v", db.db)
	var user User
	err := db.db.QueryRow(`
		SELECT id, username, roles FROM accounts."user" WHERE LOWER(username) = LOWER($1)
	`, username).Scan(&user.ID, &user.Username, pq.Array(&user.Roles))
	return user, err
}

func (db *DB) LoadUserByID(userID int32) (User, error) {
	log.Printf("Try load user if %d from db...", userID)
	var user User
	err := db.db.QueryRow(`
		SELECT id, username, roles FROM accounts."user" WHERE id = $1
	`, userID).Scan(&user.ID, &user.Username, pq.Array(&user.Roles))
	return user, err
}

// get all accounts for the user from db
func (db *DB) GetUserAccounts(userID int32) ([]Account, error) {

	// TODO: redis
	// TOFO: squirrel / sqlx

	var accounts []Account

	rows, err := db.db.Query(`
		SELECT id, name, balance FROM accounts.account WHERE user_id = $1
	`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var acc Account
		err := rows.Scan(&acc.AccountID, &acc.Name, &acc.Balance)
		if err != nil {
			return nil, err
		}
		accounts = append(accounts, acc)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	log.Printf("user %d accounts\n%+v", userID, accounts)

	return accounts, nil
}

// get single account matching id.
func (db *DB) GetAccountByID(accountID int32) (*Account, error) {

	// TODO: redis
	// TOFO: squirrel / sqlx

	acc := Account{}

	err := db.db.QueryRow(`
		SELECT id, user_id, balance from accounts.account WHERE id = $1
	`, accountID).Scan(&acc.AccountID, &acc.UserID, &acc.Balance)
	if err != nil {
		return nil, err
	}
	return &acc, nil
}

func (db *DB) CreateAccount(a Account) (int32, error) {
	var newAccID int32
	err := db.db.QueryRow(`
		INSERT INTO accounts.account (user_id, name, balance)
		VALUES ($1, $2, $3)
		RETURNING id
		`, a.UserID, a.Name, a.Balance).Scan(&newAccID)
	return newAccID, err
}

var (
	ErrTxProcessed     = errors.New("transaction already processed")
	ErrAccountNotExist = errors.New("account doesn't exist")
)

func (db *DB) CommitTransaction(transaction *Transaction) error {
	tx, err := db.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// TODO: redis
	var exists bool
	err = tx.QueryRow(`
        SELECT EXISTS (
            SELECT 1 FROM transactions.transaction WHERE id = $1
        )`, transaction.TxID).Scan(&exists)
	if err != nil {
		return err
	}
	if exists {
		log.Println("Transaction already processed:", transaction.TxID)
		return ErrTxProcessed
	}

	var balance int64
	// FOR UPDATE = pessimistic lock
	err = tx.QueryRow(`
        SELECT balance FROM accounts WHERE id = $1 FOR UPDATE
    `, transaction.AccountID).Scan(&balance)
	if err != nil {
		log.Printf("account not found: %+v", transaction)
		return ErrAccountNotExist
	}

	newBalance := balance + transaction.Amount
	_, err = tx.Exec(`
        UPDATE accounts SET balance = $1 WHERE id = $2
    `, newBalance, transaction.AccountID)
	if err != nil {
		return err
	}

	_, err = tx.Exec(`
        INSERT INTO transactions (id, account_id, amount) VALUES ($1, $2, $3)
    `, transaction.TxID, transaction.AccountID, transaction.Amount)
	if err != nil {
		return err
	}

	return tx.Commit()
}
