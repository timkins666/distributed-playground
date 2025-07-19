package common

import (
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

var secretKey = []byte("super-secret-shh")

func getAuthHeaderToken(r *http.Request) *jwt.Token {
	tokenStr := r.Header.Get("Authorization")[len("Bearer "):]
	token, _ := VerifyToken(tokenStr)
	return token
}

func CreateUserToken(user User) (string, error) {
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

func VerifyToken(tokenString string) (*jwt.Token, error) {
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

func GetUserFromClaims(r *http.Request) (User, error) {
	token := getAuthHeaderToken(r)
	if token == nil {
		return User{}, errors.New("Nope")
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return User{}, errors.New("Nope")
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
	return user, nil
}
