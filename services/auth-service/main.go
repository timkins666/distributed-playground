package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"slices"
	"strings"
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

func getUser(username string) User {
	roles := make([]string, 0)
	if username == "admin" {
		roles = append(roles, "admin")
	} else {
		roles = append(roles, "customer")
	}

	if strings.HasPrefix(username, "s") {
		roles = append(roles, "admin")
	}

	return User{
		Username: username,
		Roles:    roles,
	}
}

func createUserToken(user User) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256,
		jwt.MapClaims{
			"username": user.Username,
			"roles":    user.Roles,
			"exp":      time.Now().Add(time.Hour * 24).Unix(),
		})

	tokenString, err := token.SignedString(secretKey)
	if err != nil {
		return "", err
	}

	return tokenString, nil
}

func verifyToken(tokenString string) (*jwt.Token, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (any, error) {
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

	user := getUser(req.Username)
	token, err := createUserToken(user)
	if err != nil {
		log.Println("Error creating token: ", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]any{"username": user.Username, "roles": user.Roles, "token": token})
}

func getAuthHeaderToken(r *http.Request) *jwt.Token {
	tokenStr := r.Header.Get("Authorization")[len("Bearer "):]
	token, _ := verifyToken(tokenStr)
	return token
}

func adminHandler(w http.ResponseWriter, r *http.Request) {
	token := getAuthHeaderToken(r)
	if token == nil {
		w.WriteHeader(http.StatusUnauthorized)
		fmt.Println(w, "Nope")
		return
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		w.WriteHeader(http.StatusUnauthorized)
		fmt.Println(w, "Nope")
		return
	}

	var username string
	claimUser, ok := claims["username"]
	if ok {
		username = claimUser.(string)
	}

	var roles []string
	claimRoles, ok := claims["roles"]
	if ok {
		claimRoles, ok := claimRoles.([]any)
		if ok {
			for _, r := range claimRoles {
				r, ok := r.(string)
				if ok {
					roles = append(roles, r)
				}
			}
		}
	}

	user := User{
		Username: username,
		Roles:    roles,
	}

	if !user.valid() {
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
