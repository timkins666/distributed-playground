package common

import "errors"

var (
	ErrNotImplemented      error = errors.New("not implemented")
	ErrUserNotFound        error = errors.New("user not found")
	ErrAccountNotFound     error = errors.New("account not found")
	ErrNoNewAccountBalance error = errors.New("new account balance must be greater than zero")
)
