package main

import (
	"bytes"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"os"

	cmn "github.com/timkins666/distributed-playground/backend/pkg/common"
)

func loginHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Only POST allowed", http.StatusMethodNotAllowed)
		return
	}
	var creds cmn.LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&creds); err != nil {
		log.Println("ERROR:", err)
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	jsonBody, _ := json.Marshal(creds)
	resp, err := http.Post(
		os.Getenv("AUTH_SERVICE_HOST")+"/login",
		"application/json",
		bytes.NewBuffer(jsonBody),
	)
	if err != nil || resp.StatusCode != http.StatusOK {
		log.Println("ERROR:", err)
		http.Error(w, "Login failed", http.StatusUnauthorized)
		return
	}
	defer resp.Body.Close()

	// Forward the JWT response to the client
	w.Header().Set("Content-Type", "application/json")
	io.Copy(w, resp.Body)
}
