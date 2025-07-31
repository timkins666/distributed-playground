package main

import (
	"database/sql"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"os"
	"strings"

	cmn "github.com/timkins666/distributed-playground/backend/pkg/common"
)

type env struct {
	cmn.BaseEnv
}

func main() {
	cancelCtx, stop := cmn.GetCancelContext()
	defer stop()

	db, err := cmn.InitDB(cmn.DefaultConfig)

	if err != nil || db == nil {
		log.Panicln("Failed to initialise postgres")
	}

	app := env{
		BaseEnv: cmn.BaseEnv{}.WithCancelCtx(cancelCtx).WithDB(db),
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/login", loginHandler)

	port := ":" + os.Getenv("SERVE_PORT")
	log.Printf("Auth service running on %s", port)
	log.Fatal(http.ListenAndServe(port,
		cmn.SetContextValuesMiddleware(
			map[cmn.ContextKey]any{cmn.EnvKey: app})(mux)))
}

func loginHandler(w http.ResponseWriter, r *http.Request) {
	app, _ := r.Context().Value(cmn.EnvKey).(env)

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

func createUser(username string, app env) (cmn.User, error) {
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
	newId, err := app.DB().CreateUser(&user)

	if err != nil {
		return cmn.User{}, err
	}
	if newId <= 0 {
		return cmn.User{}, errors.New("user not saved correctly")
	}

	user.ID = newId
	log.Printf("Created user with id %d", user.ID)
	return user, nil
}

func getOrCreateUser(username string, app env) (cmn.User, error) {
	// fakes getting existing user info. don't care about passwords, that's not why we're here.
	user, err := app.DB().LoadUserByName(username)
	if err == nil {
		return user, nil
	}

	if err == sql.ErrNoRows {
		log.Println("User not found, creating")
		return createUser(username, app)
	}

	return user, err
}
