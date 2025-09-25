package main

import (
	"context"
	"fmt"
	"log"
)

func main() {
	config, err := LoadConfig()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	cancelCtx, cancel := context.WithCancel(context.Background())
	defer cancel()

	appCtx := newAppCtx(cancelCtx, config)
	defer appCtx.Close()

	service, err := initializeService(appCtx)
	if err != nil {
		log.Fatalf("Failed to initialize service: %v", err)
	}

	server := NewHTTPServer(service, config)
	if err := server.Start(cancelCtx); err != nil {
		log.Fatalf("Server error: %v", err)
	}
}

// represents the request to create a new account
type CreateAccountRequest struct {
	Name                 string `json:"name" validate:"required,min=1,max=30"` // TODO: this vs method checks
	SourceFundsAccountID int32  `json:"sourceFundsAccountId,omitempty"`
	InitialBalance       int64  `json:"initialBalance,omitempty"`
}

// checks if the request is valid. Account/balance checks deferred until later.
func (r *CreateAccountRequest) Validate() error {
	if r.Name == "" {
		return fmt.Errorf("account name is required")
	}
	if len(r.Name) > 30 {
		return fmt.Errorf("account name too long (max 30 characters)")
	}
	return nil
}
