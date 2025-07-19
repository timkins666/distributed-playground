package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"slices"
	"strings"

	cmn "github.com/timkins666/distributed-playground/backend/pkg/common"
)

func getLoginUser(username string) cmn.User {
	// fakes getting existing user info.
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

	return cmn.User{
		Username: username,
		Roles:    roles,
	}
}

func loginHandler(w http.ResponseWriter, r *http.Request) {
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

	user := getLoginUser(req.Username)
	token, err := cmn.CreateUserToken(user)
	if err != nil {
		log.Println("Error creating token: ", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]any{"username": user.Username, "roles": user.Roles, "token": token})
}

func adminHandler(w http.ResponseWriter, r *http.Request) {
	// for testing tokens
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
	mux := http.NewServeMux()
	mux.HandleFunc("/login", loginHandler)
	mux.HandleFunc("/admin", adminHandler)

	port := ":" + os.Getenv("SERVE_PORT")
	log.Printf("Auth service running on %s", port)
	log.Fatal(http.ListenAndServe(port, mux))
}
