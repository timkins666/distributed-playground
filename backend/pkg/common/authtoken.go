package common

import (
	"context"
	"errors"
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
const AppCtx ContextKey = "appCtx"

const authHeader string = "Authorization"
const authHeaderPrefix string = "Bearer "

var (
	errInvalidAuthHeader = errors.New("invalid auth header")
	errInvalidToken      = errors.New("invalid auth token")
	errTokenParse        = errors.New("could not parse token")
)

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

	*r = *r.WithContext(context.WithValue(r.Context(), UserIDKey, int32(id)))
	return true
}

func getToken(r *http.Request) (*jwt.Token, error) {
	// Extract the Bearer token from Authorization header & check expiry
	headerStr := r.Header.Get(authHeader)

	if !strings.HasPrefix(headerStr, authHeaderPrefix) {
		return nil, errInvalidAuthHeader
	}
	headerStr = strings.TrimPrefix(headerStr, authHeaderPrefix)
	token, err := parseToken(headerStr)
	if err != nil {
		return nil, errTokenParse
	}
	if !token.Valid {
		return nil, errInvalidToken
	}

	//TODO: check expiry
	iat, _ := token.Claims.GetIssuedAt()
	log.Println("iat:", iat)

	return token, nil
}

func parseToken(tokenStr string) (*jwt.Token, error) {
	// todo: see jwt.Parse docstring for alg checking
	token, err := jwt.Parse(tokenStr, func(token *jwt.Token) (any, error) {
		return secretKey, nil
	})

	if err != nil {
		return nil, err
	}

	sub, err := token.Claims.GetSubject()
	if err != nil {
		return nil, err
	}
	userId, err := strconv.Atoi(sub)
	if err != nil {
		return nil, err
	}
	if userId == 0 {
		return nil, errors.New("failed to get user id from token")
	}

	return token, nil
}

func CreateUserToken(user *User) (string, error) {
	log.Printf("Creating token for user: %+v", user)
	token := jwt.NewWithClaims(jwt.SigningMethodHS256,
		jwt.MapClaims{
			"sub": strconv.Itoa(int(user.ID)),
			"iat": time.Now().UTC().Unix(),
			"exp": time.Now().UTC().Add(15 * time.Minute).Unix(),
		})

	tokenString, err := token.SignedString(secretKey)
	if err != nil {
		return "", err
	}

	return tokenString, nil
}
