package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"slices"
	"strings"

	cm "github.com/timkins666/distributed-playground/backend/pkg/common"
)

func corsMiddleware(next http.Handler) http.Handler {
	frontend_port := os.Getenv("FRONTEND_PORT")
	log.Println("frontend port", frontend_port)

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "http://localhost:"+frontend_port)
		w.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusOK)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func getUser(username string) cm.User {
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

	return cm.User{
		Username: username,
		Roles:    roles,
	}
}

func loginHandler(w http.ResponseWriter, r *http.Request) {
	var req cm.LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request", http.StatusBadRequest)
		return
	}

	if strings.ToLower(req.Username[0:1]) == "x" {
		log.Println("x users not allowed")
		http.Error(w, "no x users", http.StatusUnauthorized)
		return
	}

	user := getUser(req.Username)
	token, err := cm.CreateUserToken(user)
	if err != nil {
		log.Println("Error creating token: ", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]any{"username": user.Username, "roles": user.Roles, "token": token})
}

func adminHandler(w http.ResponseWriter, r *http.Request) {
	user, err := cm.GetUserFromClaims(r)
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

	log.Println("Auth service running on :8081")
	log.Fatal(http.ListenAndServe(":8081", corsMiddleware(mux)))
}
