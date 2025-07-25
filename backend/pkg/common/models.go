package common

import (
	"bytes"
	"encoding/binary"
	"encoding/gob"
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
	return u != nil && u.Username != "" && len(u.Roles) > 0
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
	TxID         string
	PaymentSysID string
	Amount       int64
	AccountID    int32
}

func (t Transaction) FromBytes(b []byte) (*Transaction, error) {
	buf := new(bytes.Buffer)
	buf.Write(b)
	err := gob.NewDecoder(buf).Decode(&t)
	return &t, err
}
func (t Transaction) MsgKey() ([]byte, error) {
	buf := new(bytes.Buffer)
	err := binary.Write(buf, binary.LittleEndian, t.AccountID)
	return buf.Bytes(), err
}
func (t Transaction) MsgValue() ([]byte, error) {
	buf := new(bytes.Buffer)
	err := gob.NewEncoder(buf).Encode(t)
	return buf.Bytes(), err
}

type Account struct {
	AccountID int32  `json:"accountId"`
	Name      string `json:"name"`
	UserID    int32  `json:"userId"`
	Balance   int64  `json:"balance"`
	BankID    int32  `json:"bankId"`
	BankName  string `json:"bankName"`
}

type Bank struct {
	Name string `json:"name"`
	ID   int32  `json:"id"`
}
