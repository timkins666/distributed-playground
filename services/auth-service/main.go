package main

import (
	"encoding/json"
	"log"
	"net/http"
	"os"
	"time"

	_ "github.com/lib/pq"
)

var (
	userTokens map[string]UserToken = make(map[string]UserToken, 0)
)

func corsMiddleware(next http.Handler) http.Handler {
	frontend_port := os.Getenv("FRONTEND_PORT")
	log.Println("frontend port", frontend_port)

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "http://localhost:"+frontend_port)
		w.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusOK)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func newToken(user string) UserToken {
	return UserToken{
		Username:  user,
		Value:     user + "-token",
		CreatedAt: time.Now(),
	}
}

// func checkToken(user string, rqToken string) bool {
// 	userToken, ok := userTokens[user]
// 	if !ok {
// 		log.Printf("User %s has no token!", user)
// 		return false
// 	}

// 	if userToken.Value != rqToken {
// 		log.Printf("User %s wrong token! (Submitted %s, expected %s)", user, rqToken, userToken.Value)
// 		return false
// 	}

// 	log.Printf("User %s token ok :)", user)
// 	return true
// }

func loginHandler(w http.ResponseWriter, r *http.Request) {
	var req LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request", http.StatusBadRequest)
		return
	}

	// TODO: db not local map
	token := newToken(req.Username)
	userTokens[req.Username] = token

	json.NewEncoder(w).Encode(map[string]string{"token": token.Value})
}

func main() {
	mux := http.NewServeMux()
	mux.HandleFunc("/login", loginHandler)

	log.Println("Auth service running on :8081")
	log.Fatal(http.ListenAndServe(":8081", corsMiddleware(mux)))
}
