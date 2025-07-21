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
	Username string   `json:"username"`
	Roles    []string `json:"roles"`
}

func (u *User) Valid() bool {
	return len(u.Username) > 0 && len(u.Roles) > 0
}

type PaymentRequest struct {
	AppId           string    `json:"appId"`
	SystemId        string    `json:"systemId,omitempty"`
	Amount          int64     `json:"amount"`
	SourceAccountId int32     `json:"sourceAccountId"`
	TargetAccountId int32     `json:"targetAccountId"`
	Timestamp       time.Time `json:"timestamp"`
}

func (pr *PaymentRequest) Valid() bool {
	// checks all fields populated and imposes arbitrary timeout
	return pr.AppId != "" &&
		pr.SystemId != "" &&
		pr.Amount > 0 &&
		pr.SourceAccountId > 0 &&
		pr.TargetAccountId > 0 &&
		time.Now().UTC().After(pr.Timestamp) &&
		pr.Timestamp.After(time.Now().Add(-10*time.Second))
}

type Transaction struct {
	TxID      string
	Amount    int64
	AccountID int32
}
