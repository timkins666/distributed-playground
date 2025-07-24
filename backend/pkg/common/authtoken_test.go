package common

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/bmizerany/assert"
	"github.com/golang-jwt/jwt/v5"
)

func TestGetToken(t *testing.T) {
	tests := []func() (string, *http.Request, *jwt.Token, error){
		func() (string, *http.Request, *jwt.Token, error) {
			name := "error when no auth header present"
			r := httptest.NewRequest("post", "/", nil)
			return name, r, nil, errInvalidAuthHeader
		},
		func() (string, *http.Request, *jwt.Token, error) {
			name := "error when header empty"
			r := httptest.NewRequest("post", "/", nil)
			r.Header.Add(authHeader, "")
			return name, r, nil, errInvalidAuthHeader
		},
		func() (string, *http.Request, *jwt.Token, error) {
			name := "error when header has non-matching prefix"
			r := httptest.NewRequest("post", "/", nil)
			r.Header.Add(authHeader, "Bearerx 123")
			return name, r, nil, errInvalidAuthHeader
		},
		func() (string, *http.Request, *jwt.Token, error) {
			name := "token parse error"
			r := httptest.NewRequest("post", "/", nil)
			r.Header.Add(authHeader, authHeaderPrefix+"monkey")
			return name, r, nil, errTokenParse
		},
		func() (string, *http.Request, *jwt.Token, error) {
			name := "success"
			token := jwt.NewWithClaims(jwt.SigningMethodHS256,
				jwt.MapClaims{
					"sub": "3",
					"iat": time.Now().Unix(),
				})

			tokenString, _ := token.SignedString(secretKey)
			r := httptest.NewRequest("post", "/", nil)
			r.Header.Add(authHeader, authHeaderPrefix+tokenString)
			return name, r, token, nil
		}}

	for _, tt := range tests {
		name, r, want, wantErr := tt()

		t.Run(name, func(t *testing.T) {
			result, err := getToken(r)

			assert.Equal(t, err, wantErr)
			if want != nil {
				k := []byte("x")
				t1, _ := want.SignedString(k)
				t2, _ := result.SignedString(k)
				if t1 != t2 {
					t.Error("returned unexpected token")
				}
			}
		})
	}
}

func TestCreateUserToken(t *testing.T) {
	tests := []struct {
		claim string
		check func(*jwt.Token) bool
	}{
		{claim: "sub", check: func(tk *jwt.Token) bool {
			sub, _ := tk.Claims.GetSubject()
			return sub == "12345"
		}},
		{claim: "iat", check: func(tk *jwt.Token) bool {
			iat, _ := tk.Claims.GetIssuedAt()
			now := time.Now().Unix()
			then := time.Now().Add(-1 * time.Second).Unix()
			return then < iat.Unix() && iat.Unix() <= now
		}},
	}

	user := User{
		ID: 12345,
	}

	for _, tt := range tests {
		t.Run(tt.claim, func(t *testing.T) {
			tokenStr, err := CreateUserToken(user)
			if err != nil {
				t.Fatal(err.Error())
			}
			tokenParsed, _ := parseToken(tokenStr)

			if !tt.check(tokenParsed) {
				t.Fatalf("bad %s", tt.claim)
			}
		})
	}
}

func TestSetUserID(t *testing.T) {
	r := httptest.NewRequest("get", "/", nil)
	token, _ := CreateUserToken(User{ID: 123})
	r.Header.Add(authHeader, authHeaderPrefix+token)

	setUserID(r)

	result, ok := r.Context().Value(UserIDKey).(int)
	assert.Equal(t, ok, true)
	assert.Equal(t, result, 123)
}
