package main

import (
	"log"
	"net/http"
	"os"

	cmn "github.com/timkins666/distributed-playground/backend/pkg/common"
)

func corsMiddleware(next http.Handler) http.Handler {
	frontend_host := os.Getenv("FRONTEND_HOST")
	log.Println("allowing cross origin for", frontend_host)

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "http://"+frontend_host)
		w.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusOK)
			return
		}
		next.ServeHTTP(w, r)
	})
}

// Middleware for basic JWT auth
func withAuth(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		_, err := cmn.VerifyToken(r)
		if err != nil {
			log.Println("ERROR:", err)
			http.Error(w, "Unauthorised", http.StatusUnauthorized)
			return
		}
		next(w, r)
	}
}
