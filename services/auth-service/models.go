package main

import (
	"time"

	_ "github.com/lib/pq"
)

type LoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type UserToken struct {
	Username  string
	Value     string
	CreatedAt time.Time
}
