package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/bmizerany/assert"
	cmn "github.com/timkins666/distributed-playground/backend/pkg/common"
	tu "github.com/timkins666/distributed-playground/backend/pkg/testutils"
)

type MockAccDB struct {
	users    map[int32]cmn.User
	accounts map[int32]cmn.Account
	payments []cmn.PaymentRequest
}

func NewMockAccDB() *MockAccDB {
	return &MockAccDB{
		users:    make(map[int32]cmn.User),
		accounts: make(map[int32]cmn.Account),
	}
}

func (m *MockAccDB) getUserByID(userID int32) (*cmn.User, error) {
	user, ok := m.users[userID]
	if !ok {
		return nil, fmt.Errorf("user %d not set up", userID)
	}
	return &user, nil
}

func (m *MockAccDB) getUserAccounts(userID int32) ([]cmn.Account, error) {
	var accounts []cmn.Account
	for _, acc := range m.accounts {
		if acc.UserID == userID {
			accounts = append(accounts, acc)
		}
	}
	return accounts, nil
}

func (m *MockAccDB) getAccountByID(accountID int32) (*cmn.Account, error) {
	acc, ok := m.accounts[accountID]
	if !ok {
		return nil, cmn.ErrAccountNotFound
	}
	return &acc, nil
}

func (m *MockAccDB) createAccount(a cmn.Account) (int32, error) {
	id := int32(len(m.accounts) + 1)
	a.AccountID = id
	m.accounts[id] = a
	return id, nil
}

// getTestService creates a test app with mock dependencies
func getTestService() Service {
	mockDB := NewMockAccDB()
	mockWriter := tu.MockKafkaWriter{}
	mockReader := tu.MockKafkaReader{}

	// Add test user
	mockDB.users[1] = cmn.User{
		ID:       1,
		Username: "testuser",
		Roles:    []string{"user"},
	}

	// Add test account
	mockDB.accounts[1] = cmn.Account{
		AccountID: 1,
		Name:      "Test Account",
		UserID:    1,
		Balance:   1000,
		BankID:    1,
		BankName:  "BankOfTim",
	}

	return Service{
		appCtx: &AccountsCtx{
			cancelCtx:    context.Background(),
			db:           mockDB,
			logger:       cmn.AppLogger(),
			payReqReader: &mockReader,
			writer:       &mockWriter,
		},
		banks: []*cmn.Bank{{Name: "BankOfTim", ID: 1}},
	}
}

// setupTestRequest creates a test HTTP request with the app and userID in the context
func setupTestRequest(method, url string, body []byte, appCtx AccountsCtx, userID int32) *http.Request {
	req := httptest.NewRequest(method, url, bytes.NewBuffer(body))
	ctx := context.WithValue(req.Context(), cmn.AppCtx, &appCtx)
	ctx = context.WithValue(ctx, cmn.UserIDKey, userID)
	return req.WithContext(ctx)
}

func TestGetAllBanksHandler(t *testing.T) {
	service := getTestService()

	tests := []struct {
		name           string
		userID         int32
		expectedStatus int
		expectedBanks  []*cmn.Bank
	}{
		{
			name:           "valid user",
			userID:         1,
			expectedStatus: http.StatusOK,
			expectedBanks:  []*cmn.Bank{{Name: "BankOfTim", ID: 1}},
		},
		{
			name:           "invalid user",
			userID:         0,
			expectedStatus: http.StatusUnauthorized,
			expectedBanks:  nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := setupTestRequest("GET", "/banks", nil, *service.appCtx, tt.userID)
			w := httptest.NewRecorder()

			service.getAllBanksHandler(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)

			if tt.expectedBanks != nil {
				var banks []*cmn.Bank
				err := json.Unmarshal(w.Body.Bytes(), &banks)
				assert.Equal(t, nil, err)
				assert.Equal(t, len(tt.expectedBanks), len(banks))
				assert.Equal(t, tt.expectedBanks[0].Name, banks[0].Name)
				assert.Equal(t, tt.expectedBanks[0].ID, banks[0].ID)
			}
		})
	}
}

func TestGetUserAccountsHandler(t *testing.T) {
	service := getTestService()

	tests := []struct {
		name           string
		userID         int32
		expectedStatus int
		expectedAccs   []cmn.Account
	}{
		{
			name:           "valid user with accounts",
			userID:         1,
			expectedStatus: http.StatusOK,
			expectedAccs: []cmn.Account{
				{
					AccountID: 1,
					Name:      "Test Account",
					UserID:    1,
					Balance:   1000,
					BankID:    1,
					BankName:  "BankOfTim",
				},
			},
		},
		{
			name:           "valid user with no accounts",
			userID:         2,
			expectedStatus: http.StatusOK,
			expectedAccs:   []cmn.Account{},
		},
		{
			name:           "invalid user",
			userID:         0,
			expectedStatus: http.StatusUnauthorized,
			expectedAccs:   nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := setupTestRequest("GET", "/myaccounts", nil, *service.appCtx, tt.userID)
			w := httptest.NewRecorder()

			service.getUserAccountsHandler(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)

			if tt.expectedAccs != nil {
				var accounts []cmn.Account
				err := json.Unmarshal(w.Body.Bytes(), &accounts)
				assert.Equal(t, nil, err)
				assert.Equal(t, len(tt.expectedAccs), len(accounts))
				if len(tt.expectedAccs) > 0 {
					assert.Equal(t, tt.expectedAccs[0].AccountID, accounts[0].AccountID)
					assert.Equal(t, tt.expectedAccs[0].Name, accounts[0].Name)
					assert.Equal(t, tt.expectedAccs[0].UserID, accounts[0].UserID)
					assert.Equal(t, tt.expectedAccs[0].Balance, accounts[0].Balance)
				}
			}
		})
	}
}

func TestCreateUserAccountHandler(t *testing.T) {
	service := getTestService()

	tests := []struct {
		name           string
		userID         int32
		request        CreateAccountRequest
		expectedStatus int
		checkResponse  func(t *testing.T, resp []byte, mockWriter *tu.MockKafkaWriter)
	}{
		{
			name:   "new user first account",
			userID: 2,
			request: CreateAccountRequest{
				Name: "First Account",
			},
			expectedStatus: http.StatusCreated,
			checkResponse: func(t *testing.T, resp []byte, mockWriter *tu.MockKafkaWriter) {
				var accounts []cmn.Account
				err := json.Unmarshal(resp, &accounts)
				assert.Equal(t, nil, err)
				assert.Equal(t, 1, len(accounts))
				assert.Equal(t, "First Account", accounts[0].Name)
				assert.Equal(t, int32(2), accounts[0].UserID)
				assert.Equal(t, 1, len(mockWriter.Messages))
			},
		},
		{
			name:   "existing user new account with source funds",
			userID: 1,
			request: CreateAccountRequest{
				Name:                 "Second Account",
				SourceFundsAccountID: 1,
				InitialBalance:       500,
			},
			expectedStatus: http.StatusCreated,
			checkResponse: func(t *testing.T, resp []byte, mockWriter *tu.MockKafkaWriter) {
				var accounts []cmn.Account
				err := json.Unmarshal(resp, &accounts)
				assert.Equal(t, nil, err)
				assert.Equal(t, 2, len(accounts))

				// Find the new account
				var newAcc cmn.Account
				var sourceAcc cmn.Account
				for _, acc := range accounts {
					if acc.AccountID != 1 {
						newAcc = acc
					} else {
						sourceAcc = acc
					}
				}

				assert.Equal(t, "Second Account", newAcc.Name)
				assert.Equal(t, int64(500), newAcc.Balance)
				assert.Equal(t, int32(1), newAcc.UserID)

				// Check source account balance was reduced
				assert.Equal(t, int64(500), sourceAcc.Balance) // 1000 - 500

				// Check Kafka messages
				assert.Equal(t, 2, len(mockWriter.Messages))
			},
		},
		{
			name:   "insufficient funds in source account",
			userID: 1,
			request: CreateAccountRequest{
				Name:                 "Third Account",
				SourceFundsAccountID: 1,
				InitialBalance:       2000, // More than the 1000 balance
			},
			expectedStatus: http.StatusBadRequest,
			checkResponse: func(t *testing.T, resp []byte, mockWriter *tu.MockKafkaWriter) {
				var errorResp map[string]string
				err := json.Unmarshal(resp, &errorResp)
				assert.Equal(t, nil, err)
				assert.Equal(t, "source account doesn't have enough funds", errorResp["error"])
			},
		},
		{
			name:   "invalid source account",
			userID: 1,
			request: CreateAccountRequest{
				Name:                 "Fourth Account",
				SourceFundsAccountID: 999, // Non-existent account
				InitialBalance:       100,
			},
			expectedStatus: http.StatusBadRequest,
			checkResponse: func(t *testing.T, resp []byte, mockWriter *tu.MockKafkaWriter) {
				var errorResp map[string]string
				err := json.Unmarshal(resp, &errorResp)
				assert.Equal(t, nil, err)
				assert.Equal(t, cmn.ErrAccountNotFound.Error(), errorResp["error"])
			},
		},
		{
			name:   "existing user new account with no initial balance",
			userID: 1,
			request: CreateAccountRequest{
				Name:                 "Fifth Account",
				SourceFundsAccountID: 1,
				InitialBalance:       0, // No initial balance
			},
			expectedStatus: http.StatusBadRequest,
			checkResponse: func(t *testing.T, resp []byte, mockWriter *tu.MockKafkaWriter) {
				var errorResp map[string]string
				err := json.Unmarshal(resp, &errorResp)
				assert.Equal(t, nil, err)
				assert.Equal(t, cmn.ErrNoNewAccountBalance.Error(), errorResp["error"])
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Reset the mock writer messages
			writer, ok := service.appCtx.writer.(*tu.MockKafkaWriter)
			if !ok {
				t.Fatal("writer is not mock writer")
			}

			writer.Messages = nil

			// Create request body
			reqBody, _ := json.Marshal(tt.request)
			req := setupTestRequest("POST", "/new", reqBody, *service.appCtx, tt.userID)
			w := httptest.NewRecorder()

			service.createUserAccountHandler(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
			tt.checkResponse(t, w.Body.Bytes(), writer)
		})
	}
}
