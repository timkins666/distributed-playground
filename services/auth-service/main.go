package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

var (
	secretKey = []byte("super-secret")
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

func createToken(username string) (string, error) {
	var role string
	if username == "admin" {
		role = "admin"
	} else {
		role = "customer"
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256,
		jwt.MapClaims{
			"username": username,
			"role":     role,
			"exp":      time.Now().Add(time.Hour * 24).Unix(),
		})

	tokenString, err := token.SignedString(secretKey)
	if err != nil {
		return "", err
	}

	return tokenString, nil
}

func verifyToken(tokenString string) (*jwt.Token, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		return secretKey, nil
	})

	if err != nil {
		return nil, err
	}

	if !token.Valid {
		return nil, fmt.Errorf("invalid token")
	}

	return token, nil
}

func loginHandler(w http.ResponseWriter, r *http.Request) {
	var req LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request", http.StatusBadRequest)
		return
	}

	token, err := createToken(req.Username)
	if err != nil {
		log.Println("Error creating token: ", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"token": token})
}

func adminHandler(w http.ResponseWriter, r *http.Request) {
	tokenStr := r.Header.Get("Authorization")

	if tokenStr == "" {
		w.WriteHeader(http.StatusUnauthorized)
		fmt.Println(w, "Missing auth header")
		return
	}

	tokenStr = tokenStr[len("Bearer "):]

	token, err := verifyToken(tokenStr)
	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		fmt.Println(w, "Nope")
		return
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if ok {
		log.Println("token claims: ", claims)
		log.Println("claims role: ", claims["role"])
	}

	role, ok := claims["role"]
	if !ok {
		w.WriteHeader(http.StatusUnauthorized)
		fmt.Println(w, "Nope")
		return
	}

	if role != "admin" {
		log.Printf("Not allowed with role %s", role)
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
