package main

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/bmizerany/assert"
	cmn "github.com/timkins666/distributed-playground/backend/pkg/common"
	tu "github.com/timkins666/distributed-playground/backend/pkg/testutils"
)

// MockDB is a mock implementation of the database interface for testing
type MockDB struct {
	tu.BaseTestDB
	users    map[int]cmn.User
	accounts map[int]cmn.Account
}

func NewMockDB() MockDB {
	return MockDB{
		users:    make(map[int]cmn.User),
		accounts: make(map[int]cmn.Account),
	}
}

func (m *MockDB) LoadUserByID(userID int) (cmn.User, error) {
	user, ok := m.users[userID]
	if !ok {
		return cmn.User{}, nil
	}
	return user, nil
}

func (m *MockDB) GetUserAccounts(userID int) ([]cmn.Account, error) {
	var accounts []cmn.Account
	for _, acc := range m.accounts {
		if acc.UserID == userID {
			accounts = append(accounts, acc)
		}
	}
	return accounts, nil
}

func (m *MockDB) GetAccountByID(accountID int) (*cmn.Account, error) {
	acc, ok := m.accounts[accountID]
	if !ok {
		return nil, nil
	}
	return &acc, nil
}

func (m *MockDB) CreateAccount(a cmn.Account) (int, error) {
	id := len(m.accounts) + 1
	a.AccountID = id
	m.accounts[id] = a
	return id, nil
}

// setupTestApp creates a test app with mock dependencies
func setupTestApp() appEnv {
	mockDB := NewMockDB()
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
		BankName:  "Bonzo",
	}

	return appEnv{
		BaseEnv:      cmn.BaseEnv{}.WithCancelCtx(context.Background()).WithDB(&mockDB),
		payReqReader: &mockReader,
		writer:       &mockWriter,
	}
}

// setupTestRequest creates a test HTTP request with the app and userID in the context
func setupTestRequest(method, url string, body []byte, app appEnv, userID int) *http.Request {
	req := httptest.NewRequest(method, url, bytes.NewBuffer(body))
	ctx := context.WithValue(req.Context(), cmn.AppKey, app)
	ctx = context.WithValue(ctx, cmn.UserIDKey, userID)
	return req.WithContext(ctx)
}

func TestGetAllBanksHandler(t *testing.T) {
	testApp := setupTestApp()

	tests := []struct {
		name           string
		userID         int
		expectedStatus int
		expectedBanks  []*cmn.Bank
	}{
		{
			name:           "valid user",
			userID:         1,
			expectedStatus: http.StatusOK,
			expectedBanks:  []*cmn.Bank{{Name: "Bonzo", ID: 1}},
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
			req := setupTestRequest("GET", "/banks", nil, testApp, tt.userID)
			w := httptest.NewRecorder()

			getAllBanksHandler(w, req)

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
	testApp := setupTestApp()

	tests := []struct {
		name           string
		userID         int
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
					BankName:  "Bonzo",
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
			req := setupTestRequest("GET", "/myaccounts", nil, testApp, tt.userID)
			w := httptest.NewRecorder()

			getUserAccountsHandler(w, req)

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
	testApp := setupTestApp()

	tests := []struct {
		name           string
		userID         int
		request        newAccountRequest
		expectedStatus int
		checkResponse  func(t *testing.T, resp []byte, mockWriter *tu.MockKafkaWriter)
	}{
		{
			name:   "new user first account",
			userID: 2,
			request: newAccountRequest{
				Name: "First Account",
			},
			expectedStatus: http.StatusOK,
			checkResponse: func(t *testing.T, resp []byte, mockWriter *tu.MockKafkaWriter) {
				var accounts []cmn.Account
				err := json.Unmarshal(resp, &accounts)
				assert.Equal(t, nil, err)
				assert.Equal(t, 1, len(accounts))
				assert.Equal(t, "First Account", accounts[0].Name)
				assert.Equal(t, 2, accounts[0].UserID)
				assert.Equal(t, 1, len(mockWriter.Messages))
			},
		},
		{
			name:   "existing user new account with source funds",
			userID: 1,
			request: newAccountRequest{
				Name:                 "Second Account",
				SourceFundsAccountID: 1,
				InitialBalance:       500,
			},
			expectedStatus: http.StatusOK,
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
				assert.Equal(t, 1, newAcc.UserID)

				// Check source account balance was reduced
				assert.Equal(t, int64(500), sourceAcc.Balance) // 1000 - 500

				// Check Kafka messages
				assert.Equal(t, 2, len(mockWriter.Messages))
			},
		},
		{
			name:   "insufficient funds in source account",
			userID: 1,
			request: newAccountRequest{
				Name:                 "Third Account",
				SourceFundsAccountID: 1,
				InitialBalance:       2000, // More than the 1000 balance
			},
			expectedStatus: http.StatusBadRequest,
			checkResponse: func(t *testing.T, resp []byte, mockWriter *tu.MockKafkaWriter) {
				var errorResp map[string]string
				err := json.Unmarshal(resp, &errorResp)
				assert.Equal(t, nil, err)
				assert.Equal(t, "Source account doesn't have enough funds", errorResp["errorReason"])
			},
		},
		{
			name:   "invalid source account",
			userID: 1,
			request: newAccountRequest{
				Name:                 "Fourth Account",
				SourceFundsAccountID: 999, // Non-existent account
				InitialBalance:       100,
			},
			expectedStatus: http.StatusBadRequest,
			checkResponse: func(t *testing.T, resp []byte, mockWriter *tu.MockKafkaWriter) {
				var errorResp map[string]string
				err := json.Unmarshal(resp, &errorResp)
				assert.Equal(t, nil, err)
				assert.Equal(t, "Invalid source account", errorResp["errorReason"])
			},
		},
		{
			name:   "existing user new account with no initial balance",
			userID: 1,
			request: newAccountRequest{
				Name:                 "Fifth Account",
				SourceFundsAccountID: 1,
				InitialBalance:       0, // No initial balance
			},
			expectedStatus: http.StatusBadRequest,
			checkResponse: func(t *testing.T, resp []byte, mockWriter *tu.MockKafkaWriter) {
				var errorResp map[string]string
				err := json.Unmarshal(resp, &errorResp)
				assert.Equal(t, nil, err)
				assert.Equal(t, "Must transfer with an initial balance", errorResp["errorReason"])
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Reset the mock writer messages
			writer, ok := testApp.writer.(*tu.MockKafkaWriter)
			if !ok {
				t.Fatal("writer is not mock writer")
			}

			writer.Messages = nil

			// Create request body
			reqBody, _ := json.Marshal(tt.request)
			req := setupTestRequest("POST", "/new", reqBody, testApp, tt.userID)
			w := httptest.NewRecorder()

			createUserAccountHandler(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
			tt.checkResponse(t, w.Body.Bytes(), writer)
		})
	}
}

// TestPaymentValidator tests the payment validation logic
func TestPaymentValidator(t *testing.T) {
	// This is a more complex test that would involve mocking Kafka messages
	// and testing the asynchronous processing. For simplicity, we'll focus on
	// the handler tests first.
}
