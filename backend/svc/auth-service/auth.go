package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"slices"
	"strings"

	"github.com/lib/pq"
	cmn "github.com/timkins666/distributed-playground/backend/pkg/common"
)

func createUser(username string, db *sql.DB) *cmn.User {
	// create new user
	// any username beginning with s or S will be a customer and an admin.
	// admin user will be admin only.
	// any other user will be customer only.
	roles := []string{}
	if username == "admin" {
		roles = append(roles, "admin")
	} else {
		roles = append(roles, "customer")
	}

	if strings.ToLower(username)[0:1] == "s" {
		roles = append(roles, "admin")
	}

	user := cmn.User{
		Username: username,
		Roles:    roles,
	}

	log.Printf("Creating user %+v", user)
	err := db.QueryRow(`
	INSERT INTO accounts."user" (username, roles) VALUES ($1, $2) RETURNING id
	`, user.Username, pq.Array(user.Roles)).Scan(&user.ID)

	if err != nil {
		log.Println("ERROR CREATING USER: ", err)
		return nil
	}

	log.Printf("Creating user with id %d", user.ID)
	return &user
}

func getOrCreateUser(username string, db *sql.DB) *cmn.User {
	// fakes getting existing user info. don't care about passwords, that's not why we're here.

	log.Printf("Try load user %s from db...", username)
	var user cmn.User
	err := db.QueryRow(`
		SELECT id, username, roles FROM accounts."user" WHERE username = $1
	`, username).Scan(&user.ID, &user.Username, pq.Array(&user.Roles))

	if user.ID > 0 {
		log.Printf("User found %+v", user)
		return &user
	}

	log.Println("User not found")
	if err == sql.ErrNoRows {
		return createUser(username, db)
	}

	log.Println("ERROR: ", err)
	return nil
}

func loginHandler(db *sql.DB) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		var req cmn.LoginRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "invalid request", http.StatusBadRequest)
			return
		}

		if strings.ToLower(req.Username[0:1]) == "x" {
			log.Println("x users not allowed")
			http.Error(w, "no x users", http.StatusUnauthorized)
			return
		}

		user := getOrCreateUser(req.Username, db)

		if user == nil {
			log.Println("Failed creating user :p")
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		token, err := cmn.CreateUserToken(user)
		if err != nil {
			log.Println("Error creating token: ", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]any{"username": user.Username, "roles": user.Roles, "token": token})
	}
}

func adminHandler(w http.ResponseWriter, r *http.Request) {
	// for testing tokens, will go away soon
	user, err := cmn.GetUserFromClaims(r)
	if err != nil || !user.Valid() {
		w.WriteHeader(http.StatusUnauthorized)
		fmt.Println(w, "Nope")
		return
	}

	if !slices.Contains(user.Roles, "admin") {
		log.Printf("User %s does not have admin role (has %s)", user.Username, user.Roles)
		w.WriteHeader(http.StatusUnauthorized)
		fmt.Println(w, "Not admin")
		return
	}

	w.WriteHeader(http.StatusOK)
}

func main() {
	db, err := cmn.InitDB(10)
	if err != nil {
		log.Panicf("Couldn't connect to db: %s", err)
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/login", loginHandler(db))
	mux.HandleFunc("/admin", adminHandler)

	port := ":" + os.Getenv("SERVE_PORT")
	log.Printf("Auth service running on %s", port)
	log.Fatal(http.ListenAndServe(port, mux))
}
