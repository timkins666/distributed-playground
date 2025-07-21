package common

import (
	"time"

	_ "github.com/lib/pq"
)

type LoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type User struct {
	ID       int32    `json:"id"`
	Username string   `json:"username"`
	Roles    []string `json:"roles"`
}

func (u *User) Valid() bool {
	return len(u.Username) > 0 && len(u.Roles) > 0
}

type PaymentRequest struct {
	AppID           string    `json:"appId"`
	SystemID        string    `json:"systemId,omitempty"`
	Amount          int64     `json:"amount"`
	SourceAccountID int32     `json:"sourceAccountId"`
	TargetAccountID int32     `json:"targetAccountId"`
	Timestamp       time.Time `json:"timestamp"`
}

func (pr *PaymentRequest) Valid() bool {
	// checks all fields populated and imposes arbitrary timeout
	return pr.AppID != "" &&
		pr.SystemID != "" &&
		pr.Amount > 0 &&
		pr.SourceAccountID > 0 &&
		pr.TargetAccountID > 0 &&
		time.Now().UTC().After(pr.Timestamp) &&
		pr.Timestamp.After(time.Now().Add(-10*time.Second))
}

type Transaction struct {
	TxID      string
	Amount    int64
	AccountID int32
}
