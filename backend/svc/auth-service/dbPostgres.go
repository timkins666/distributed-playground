package main

import (
	"database/sql"
	"log"

	"github.com/lib/pq"
	cmn "github.com/timkins666/distributed-playground/backend/pkg/common"
)

type dbPostgres struct {
	db *sql.DB
}

// load user by name from db. searches case insensitively, returns userame casing as in db.
func (db *dbPostgres) getUserByName(username string) (*cmn.User, error) {

	log.Printf("Try load user %s from db...", username)
	log.Printf("db is %+v", db)
	log.Printf("db.db is %+v", db.db)
	var user cmn.User
	err := db.db.QueryRow(`
		SELECT id, username, roles FROM accounts."user" WHERE LOWER(username) = LOWER($1)
	`, username).Scan(&user.ID, &user.Username, pq.Array(&user.Roles))
	return &user, err
}

// creates the user in the db, returning the new user id
func (db *dbPostgres) createUser(user *cmn.User) (int32, error) {
	userID := int32(0)
	err := db.db.QueryRow(`
		INSERT INTO accounts."user" (username, roles) VALUES ($1, $2) RETURNING id
	`, user.Username, pq.Array(user.Roles)).Scan(&userID)
	return userID, err
}
