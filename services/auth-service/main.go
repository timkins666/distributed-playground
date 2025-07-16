package main

import (
    "encoding/json"
    "log"
    "net/http"
    "os"

    "github.com/gorilla/mux"
    _ "github.com/lib/pq"
)

type LoginRequest struct {
    Username string `json:"username"`
    Password string `json:"password"`
}

func loginHandler(w http.ResponseWriter, r *http.Request) {
    var req LoginRequest
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        http.Error(w, "invalid request", http.StatusBadRequest)
        return
    }

    // Stub: In real app, validate against DB and return JWT
    token := "mock-jwt-token-for-" + req.Username
    json.NewEncoder(w).Encode(map[string]string{"token": token})
}

func main() {
    r := mux.NewRouter()
    r.HandleFunc("/login", loginHandler).Methods("POST")

    log.Println("Auth service running on :8081")
    log.Fatal(http.ListenAndServe(":8081", r))
}
