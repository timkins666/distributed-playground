package common

import (
	"log"
	"net/http"
	"os"
)

func CorsMiddleware(next http.Handler) http.Handler {
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
