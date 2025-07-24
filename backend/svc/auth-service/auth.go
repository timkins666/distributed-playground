package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	cmn "github.com/timkins666/distributed-playground/backend/pkg/common"
)

type app struct {
	cancelCtx context.Context
	db        *cmn.DB
	log       *log.Logger
}

func main() {
	cancelCtx, stop := cmn.GetCancelContext()
	defer stop()

	log.Println("sleep for debug connect time")
	time.Sleep(10 * time.Second)

	db, err := cmn.InitDB(cmn.DefaultConfig)

	if err != nil || db == nil {
		log.Panicln("Failed to initialise postgres")
	}

	app := app{
		cancelCtx: cancelCtx,
		db:        db,
		log:       cmn.AppLogger(),
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/login", loginHandler)
	// mux.HandleFunc("/admin", adminHandler)

	port := ":" + os.Getenv("SERVE_PORT")
	log.Printf("Auth service running on %s", port)
	log.Fatal(http.ListenAndServe(port,
		cmn.SetContextValuesMiddleware(
			map[cmn.ContextKey]any{cmn.AppKey: app})(mux)))
}

func loginHandler(w http.ResponseWriter, r *http.Request) {
	app, _ := r.Context().Value(cmn.AppKey).(app)

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

	user, err := getOrCreateUser(req.Username, app)

	if err != nil {
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
	_ = json.NewEncoder(w).Encode(map[string]any{"username": user.Username, "roles": user.Roles, "token": token})
}

func createUser(username string, app app) (cmn.User, error) {
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
	err := app.db.CreateUser(&user)

	if err != nil {
		return cmn.User{}, err
	}
	if user.ID <= 0 {
		return cmn.User{}, errors.New("user not saved correctly")
	}

	log.Printf("Created user with id %d", user.ID)
	return user, nil
}

func getOrCreateUser(username string, app app) (cmn.User, error) {
	// fakes getting existing user info. don't care about passwords, that's not why we're here.
	user, err := app.db.LoadUserByName(username)
	if err == nil {
		return user, nil
	}

	if err == sql.ErrNoRows {
		log.Println("User not found, creating")
		return createUser(username, app)
	}

	return user, err
}

func adminHandler(w http.ResponseWriter, _ *http.Request) {
	// // for testing tokens, will go away soon
	// user, err := cmn.GetUserFromToken(r)
	// if err != nil || !user.Valid() {
	// 	w.WriteHeader(http.StatusUnauthorized)
	// 	fmt.Println(w, "Nope")
	// 	return
	// }

	// if !slices.Contains(user.Roles, "admin") {
	// 	log.Printf("User %s does not have admin role (has %s)", user.Username, user.Roles)
	// 	w.WriteHeader(http.StatusUnauthorized)
	// 	fmt.Println(w, "Not admin")
	// 	return
	// }

	w.WriteHeader(http.StatusOK)
}
