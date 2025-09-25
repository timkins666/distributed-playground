package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"math/rand/v2"
	"net/http"

	"github.com/google/uuid"
	"github.com/segmentio/kafka-go"
	cmn "github.com/timkins666/distributed-playground/backend/pkg/common"
)

type Service struct {
	appCtx *AccountsCtx
	banks  []*cmn.Bank
}

// sets up the service with all dependencies
func initializeService(appCtx *AccountsCtx) (*Service, error) {
	return &Service{
		appCtx: appCtx,
		banks:  []*cmn.Bank{{Name: "BankOfTim", ID: 1}}, // TODO: load from database
	}, nil
}

// getAllBanksHandler returns all available banks
func (s *Service) getAllBanksHandler(w http.ResponseWriter, r *http.Request) {
	appCtx, ok := r.Context().Value(cmn.AppCtx).(*AccountsCtx)
	if !ok {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	userID, ok := r.Context().Value(cmn.UserIDKey).(int32)
	if !ok || userID == 0 {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	user, err := appCtx.db.getUserByID(userID)
	if err != nil || !user.Valid() {
		var errStr string
		if err != nil {
			errStr = err.Error()
		} else {
			errStr = "user not valid"
		}

		appCtx.logger.Printf("Failed to load user %d: %v", userID, errStr)
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	appCtx.logger.Printf("User %s requested banks list", user.Username)
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(s.banks); err != nil {
		appCtx.logger.Printf("Failed to encode banks response: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
	}
}

// getUserAccountsHandler returns accounts for the authenticated user
func (s *Service) getUserAccountsHandler(w http.ResponseWriter, r *http.Request) {
	appCtx, ok := r.Context().Value(cmn.AppCtx).(*AccountsCtx)
	if !ok {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	userID, ok := r.Context().Value(cmn.UserIDKey).(int32)
	if !ok || userID == 0 {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	accs, err := appCtx.db.getUserAccounts(userID)
	if err != nil && err != sql.ErrNoRows {
		appCtx.logger.Printf("Failed to get accounts for user %d: %v", userID, err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// Ensure we return an empty array instead of null
	if accs == nil {
		accs = []cmn.Account{}
	}

	appCtx.logger.Printf("Retrieved %d accounts for user %d", len(accs), userID)
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(accs); err != nil {
		appCtx.logger.Printf("Failed to encode accounts response: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
	}
}

// createUserAccountHandler creates a new account for the authenticated user
func (s *Service) createUserAccountHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	appCtx, ok := r.Context().Value(cmn.AppCtx).(*AccountsCtx)
	if !ok {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	userID, ok := r.Context().Value(cmn.UserIDKey).(int32)
	if !ok || userID == 0 {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	var req CreateAccountRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		appCtx.logger.Printf("Failed to decode request: %v", err)
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if err := req.Validate(); err != nil {
		appCtx.logger.Printf("Invalid request: %v", err)
		s.writeErrorResponse(w, err.Error(), http.StatusBadRequest)
		return
	}

	accounts, err := s.createAccount(appCtx, userID, &req)
	if err != nil {
		appCtx.logger.Printf("Failed to create account: %v", err)
		s.writeErrorResponse(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	if err := json.NewEncoder(w).Encode(accounts); err != nil {
		appCtx.logger.Printf("Failed to encode response: %v", err)
	}
}

// createAccount handles the business logic for account creation
func (s *Service) createAccount(appCtx *AccountsCtx, userID int32, req *CreateAccountRequest) ([]cmn.Account, error) {
	userAccounts, err := appCtx.db.getUserAccounts(userID)
	if err != nil && err != sql.ErrNoRows {
		return nil, fmt.Errorf("failed to get user accounts: %w", err)
	}

	var sourceAcc *cmn.Account
	isFirstAccount := len(userAccounts) == 0

	if !isFirstAccount {
		if req.InitialBalance <= 0 {
			return nil, cmn.ErrNoNewAccountBalance
		}

		sourceAcc, err = s.validateSourceAccount(appCtx, req.SourceFundsAccountID, userID, req.InitialBalance)
		if err != nil {
			return nil, err
		}
	} else {
		// First account gets random balance
		req.InitialBalance = rand.Int64N(10e5) + 1000 // Ensure minimum balance
	}

	newAccount := cmn.Account{
		Name:     req.Name,
		UserID:   userID,
		Balance:  req.InitialBalance,
		BankID:   s.banks[0].ID,
		BankName: s.banks[0].Name,
	}

	accID, err := appCtx.db.createAccount(newAccount)
	if err != nil || accID <= 0 {
		return nil, fmt.Errorf("failed to create account: %w", err)
	}
	newAccount.AccountID = accID

	appCtx.logger.Printf("Created new account %d for user %d with balance %d", accID, userID, req.InitialBalance)

	// Create transactions for the account creation
	if err := s.createAccountTransactions(appCtx, &newAccount, sourceAcc, req.SourceFundsAccountID); err != nil {
		appCtx.logger.Printf("Failed to create transactions: %v", err)
		// TODO: Account is created, but transactions failed - rollback
	}

	respAccounts := []cmn.Account{newAccount}
	if sourceAcc != nil {
		sourceAcc.Balance -= req.InitialBalance
		respAccounts = append(respAccounts, *sourceAcc)
	}

	return respAccounts, nil
}

// validateSourceAccount validates the source account for fund transfer
func (s *Service) validateSourceAccount(appCtx *AccountsCtx, sourceAccountID, userID int32, amount int64) (*cmn.Account, error) {
	if sourceAccountID == 0 {
		return nil, fmt.Errorf("source account ID is required for additional accounts")
	}

	sourceAcc, err := appCtx.db.getAccountByID(sourceAccountID)
	if err != nil {
		return nil, cmn.ErrAccountNotFound
	}

	if sourceAcc.UserID != userID {
		return nil, fmt.Errorf("source account does not belong to user")
	}

	if sourceAcc.Balance < amount {
		return nil, fmt.Errorf("source account doesn't have enough funds")
	}

	return sourceAcc, nil
}

// creates the necessary Kafka messages for initial account balance transfer
func (s *Service) createAccountTransactions(appCtx *AccountsCtx, newAccount *cmn.Account, sourceAcc *cmn.Account, sourceAccountID int32) error {
	paymentID := uuid.NewString()

	// Transaction for the new account (credit)
	txCredit := cmn.Transaction{
		TxID:         uuid.NewString(),
		Amount:       newAccount.Balance,
		AccountID:    newAccount.AccountID,
		PaymentSysID: paymentID,
	}

	txKey, err := cmn.ToBytes(newAccount.AccountID)
	if err != nil {
		return fmt.Errorf("failed to serialize account ID: %w", err)
	}

	txMsg, err := cmn.ToBytes(txCredit)
	if err != nil {
		return fmt.Errorf("failed to serialize transaction: %w", err)
	}

	messages := []kafka.Message{{
		Topic: cmn.Topics.TransactionRequested().S(),
		Key:   txKey,
		Value: txMsg,
	}}

	// If there's a source account, create debit transaction
	if sourceAcc != nil {
		txDebit := cmn.Transaction{
			TxID:         uuid.NewString(),
			Amount:       -newAccount.Balance,
			AccountID:    sourceAccountID,
			PaymentSysID: paymentID,
		}

		debitKey, err := cmn.ToBytes(sourceAccountID)
		if err != nil {
			return fmt.Errorf("failed to serialize source account ID: %w", err)
		}

		debitMsg, err := cmn.ToBytes(txDebit)
		if err != nil {
			return fmt.Errorf("failed to serialize debit transaction: %w", err)
		}

		messages = append(messages, kafka.Message{
			Topic: cmn.Topics.TransactionRequested().S(),
			Key:   debitKey,
			Value: debitMsg,
		})
	}

	appCtx.logger.Printf("Created %d transactions for new account %d", len(messages), newAccount.AccountID)
	return appCtx.writer.WriteMessages(appCtx.cancelCtx, messages...)
}

// writeErrorResponse writes a JSON error response
func (s *Service) writeErrorResponse(w http.ResponseWriter, message string, statusCode int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	resp := map[string]string{"error": message}
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		// Fallback to plain text if JSON encoding fails
		w.Header().Set("Content-Type", "text/plain")
		fmt.Fprintf(w, "Error: %s", message)
	}
}
