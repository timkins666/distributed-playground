package main

import (
	"database/sql"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"os"

	cmn "github.com/timkins666/distributed-playground/backend/pkg/common"
)

func main() {
	cancelCtx, stop := cmn.GetCancelContext()
	defer stop()

	app := newAppCtx(cancelCtx)

	mux := http.NewServeMux()
	mux.HandleFunc("/login", loginHandler)

	port := ":" + os.Getenv("SERVE_PORT")
	log.Printf("Auth service running on %s", port)
	log.Fatal(http.ListenAndServe(port,
		cmn.SetContextValuesMiddleware(
			map[cmn.ContextKey]any{cmn.AppCtx: app})(mux)))
}

func loginHandler(w http.ResponseWriter, r *http.Request) {
	app, ok := r.Context().Value(cmn.AppCtx).(*authCtx)
	if !ok {
		log.Println("Failed to get app context")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	var req cmn.LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request", http.StatusBadRequest)
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
	_ = json.NewEncoder(w).Encode(map[string]any{
		"username": user.Username,
		"roles":    user.Roles,
		"token":    token,
	})
}

// don't care about passwords, that's not why we're here.
func getOrCreateUser(username string, app *authCtx) (*cmn.User, error) {
	user, err := app.db.getUserByName(username)
	if err == nil {
		return user, nil
	}

	if err == sql.ErrNoRows {
		log.Println("User not found, creating")
		return createUser(username, app)
	}

	return user, err
}

// create new user. all users are customers and admins for the time being.
func createUser(username string, app *authCtx) (*cmn.User, error) {
	roles := []string{"admin", "customer"}

	user := cmn.User{
		Username: username,
		Roles:    roles,
	}

	log.Printf("Creating user %+v", user)
	newId, err := app.db.createUser(&user)

	if err != nil {
		return nil, err
	}
	if newId <= 0 {
		return nil, errors.New("user not saved correctly")
	}

	user.ID = newId
	log.Printf("Created user with id %d", user.ID)
	return &user, nil
}
