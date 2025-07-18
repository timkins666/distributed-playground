package main

import (
	_ "github.com/lib/pq"
)

type LoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type User struct {
	Username string   `json:"username"`
	Roles    []string `json:"roles"`
}

func (u *User) valid() bool {
	return len(u.Username) > 0 && len(u.Roles) > 0
}
