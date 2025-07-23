package common

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

var secretKey = []byte("super-secret-shh")

type ContextKey string

const UserIDKey ContextKey = "userIDKey"
const AppKey ContextKey = "app"

// handlerfunc template
func _(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// do stuff
		next(w, r)
	}
}

func SetUserIDMiddlewareHandler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if ok := setUserID(r); !ok {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func SetUserIDMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if ok := setUserID(r); !ok {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		next.ServeHTTP(w, r)
	}
}

func setUserID(r *http.Request) bool {
	token, err := getToken(r)
	if err != nil {
		log.Println(err)
		return false
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return false
	}

	idStr, err := claims.GetSubject()
	if err != nil {
		log.Println(err)
		return false
	}
	id, err := strconv.Atoi(idStr)
	if err != nil {
		log.Println(err)
		return false
	}

	*r = *r.WithContext(context.WithValue(r.Context(), UserIDKey, id))
	return true
}

func getToken(r *http.Request) (*jwt.Token, error) {
	// Extract the Bearer token from Authorization header & check expiry
	headerStr := r.Header.Get("Authorization")

	if !strings.HasPrefix(headerStr, "Bearer ") {
		return nil, errors.New("invalid header")
	}
	headerStr = strings.TrimPrefix(headerStr, "Bearer ")

	token, err := jwt.Parse(headerStr, func(token *jwt.Token) (any, error) {
		return secretKey, nil
	})

	if err != nil {
		return nil, err
	}
	if !token.Valid {
		return nil, fmt.Errorf("invalid token")
	}

	//TODO: check expiry
	iat, _ := token.Claims.GetIssuedAt()
	log.Println("iat:", iat)

	return token, nil
}

func CreateUserToken(user User) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256,
		jwt.MapClaims{
			"sub": strconv.Itoa(int(user.ID)),
			"iat": time.Now().Unix(),
		})

	tokenString, err := token.SignedString(secretKey)
	if err != nil {
		return "", err
	}

	return tokenString, nil
}
