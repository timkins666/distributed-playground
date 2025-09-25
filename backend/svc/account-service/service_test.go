package main

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	cmn "github.com/timkins666/distributed-playground/backend/pkg/common"
	tu "github.com/timkins666/distributed-playground/backend/pkg/testutils"
)

// mockDB implements database interface for testing
type mockDB struct {
	users     map[int32]*cmn.User
	accounts  map[int32]*cmn.Account
	nextAccID int32
}

func NewMockDB() *mockDB {
	return &mockDB{
		users:     make(map[int32]*cmn.User),
		accounts:  make(map[int32]*cmn.Account),
		nextAccID: 1,
	}
}

func (m *mockDB) getUserByID(id int32) (*cmn.User, error) {
	if user, exists := m.users[id]; exists {
		return user, nil
	}
	return &cmn.User{}, cmn.ErrUserNotFound
}

func (m *mockDB) getUserAccounts(userID int32) ([]cmn.Account, error) {
	var accounts []cmn.Account
	for _, acc := range m.accounts {
		if acc.UserID == userID {
			accounts = append(accounts, *acc)
		}
	}
	return accounts, nil
}

func (m *mockDB) getAccountByID(id int32) (*cmn.Account, error) {
	if acc, exists := m.accounts[id]; exists {
		return acc, nil
	}
	return nil, cmn.ErrAccountNotFound
}

func (m *mockDB) createAccount(acc cmn.Account) (int32, error) {
	id := m.nextAccID
	m.nextAccID++
	acc.AccountID = id
	m.accounts[id] = &acc
	return id, nil
}

// Test helper to create a test service
func createTestService(t *testing.T) *Service {
	mockDB := NewMockDB()
	// Add a test user
	mockDB.users[1] = &cmn.User{
		ID:       1,
		Username: "testuser",
		Roles:    []string{"foo"},
	}

	appCtx := AccountsCtx{
		cancelCtx:    context.Background(),
		db:           mockDB,
		payReqReader: &tu.MockKafkaReader{},
		writer:       &tu.MockKafkaWriter{},
		logger:       cmn.AppLogger(),
	}

	srv, err := initializeService(&appCtx)

	if err != nil {
		t.Fatalf("Failed to initialize service for tests: %v", err)
	}

	return srv
}

// Test helper to create HTTP request with context
func createRequestWithContext(method, url string, body []byte, userID int32, appCtx *AccountsCtx) *http.Request {
	req := httptest.NewRequest(method, url, bytes.NewReader(body))

	// Add user context
	ctx := context.WithValue(req.Context(), cmn.UserIDKey, userID)

	// Create mock appCtx and add to context
	mockDB := NewMockDB()
	mockDB.users[userID] = &cmn.User{
		ID:       userID,
		Username: "testuser",
	}

	ctx = context.WithValue(ctx, cmn.AppCtx, appCtx)

	return req.WithContext(ctx)
}

func TestService_GetAllBanksHandler(t *testing.T) {
	service := createTestService(t)

	tests := []struct {
		name           string
		userID         int32
		expectedStatus int
		expectedBanks  int
	}{
		{
			name:           "valid user gets banks",
			userID:         1,
			expectedStatus: http.StatusOK,
			expectedBanks:  1,
		},
		{
			name:           "invalid user gets unauthorized",
			userID:         0,
			expectedStatus: http.StatusUnauthorized,
			expectedBanks:  0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := createRequestWithContext("GET", "/banks", nil, tt.userID, service.appCtx)
			rr := httptest.NewRecorder()

			service.getAllBanksHandler(rr, req)

			if rr.Code != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d", tt.expectedStatus, rr.Code)
			}

			if tt.expectedStatus == http.StatusOK {
				var banks []*cmn.Bank
				if err := json.Unmarshal(rr.Body.Bytes(), &banks); err != nil {
					t.Errorf("Failed to unmarshal response: %v", err)
				}

				if len(banks) != tt.expectedBanks {
					t.Errorf("Expected %d banks, got %d", tt.expectedBanks, len(banks))
				}
			}
		})
	}
}

func TestService_GetUserAccountsHandler(t *testing.T) {
	service := createTestService(t)

	tests := []struct {
		name           string
		userID         int32
		expectedStatus int
	}{
		{
			name:           "valid user gets empty accounts",
			userID:         1,
			expectedStatus: http.StatusOK,
		},
		{
			name:           "invalid user gets unauthorized",
			userID:         0,
			expectedStatus: http.StatusUnauthorized,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := createRequestWithContext("GET", "/myaccounts", nil, tt.userID, service.appCtx)
			rr := httptest.NewRecorder()

			service.getUserAccountsHandler(rr, req)

			if rr.Code != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d", tt.expectedStatus, rr.Code)
			}

			if tt.expectedStatus == http.StatusOK {
				var accounts []cmn.Account
				if err := json.Unmarshal(rr.Body.Bytes(), &accounts); err != nil {
					t.Errorf("Failed to unmarshal response: %v", err)
				}

				// Should return empty array, not null
				if accounts == nil {
					t.Error("Expected empty array, got nil")
				}
			}
		})
	}
}

func TestCreateAccountRequest_Validate(t *testing.T) {
	tests := []struct {
		name    string
		req     CreateAccountRequest
		wantErr bool
	}{
		{
			name: "valid request",
			req: CreateAccountRequest{
				Name:           "Test Account",
				InitialBalance: 1000,
			},
			wantErr: false,
		},
		{
			name: "empty name",
			req: CreateAccountRequest{
				Name:           "",
				InitialBalance: 1000,
			},
			wantErr: true,
		},
		{
			name: "name too long",
			req: CreateAccountRequest{
				Name:           string(make([]byte, 101)), // 101 characters
				InitialBalance: 1000,
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.req.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestService_CreateUserAccountHandler(t *testing.T) {
	service := createTestService(t)

	tests := []struct {
		name           string
		method         string
		userID         int32
		requestBody    CreateAccountRequest
		expectedStatus int
	}{
		{
			name:   "valid first account creation",
			method: "POST",
			userID: 1,
			requestBody: CreateAccountRequest{
				Name: "My First Account",
			},
			expectedStatus: http.StatusCreated,
		},
		{
			name:           "invalid method",
			method:         "GET",
			userID:         1,
			expectedStatus: http.StatusMethodNotAllowed,
		},
		{
			name:   "invalid request - empty name",
			method: "POST",
			userID: 1,
			requestBody: CreateAccountRequest{
				Name: "",
			},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "unauthorized user",
			method:         "POST",
			userID:         0,
			expectedStatus: http.StatusUnauthorized,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var body []byte
			var err error

			if tt.method == "POST" {
				body, err = json.Marshal(tt.requestBody)
				if err != nil {
					t.Fatalf("Failed to marshal request: %v", err)
				}
			}

			req := createRequestWithContext(tt.method, "/new", body, tt.userID, service.appCtx)
			rr := httptest.NewRecorder()

			service.createUserAccountHandler(rr, req)

			if rr.Code != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d. Response: %s",
					tt.expectedStatus, rr.Code, rr.Body.String())
			}
		})
	}
}

func TestPaymentValidationResult_IsValid(t *testing.T) {
	tests := []struct {
		name   string
		result PaymentValidationResult
		want   bool
	}{
		{
			name: "all checks pass",
			result: PaymentValidationResult{
				Results: []CheckResult{
					{CheckName: BalanceCheck, Result: true},
					{CheckName: TargetAccountCheck, Result: true},
				},
				TimedOut: false,
			},
			want: true,
		},
		{
			name: "one check fails",
			result: PaymentValidationResult{
				Results: []CheckResult{
					{CheckName: BalanceCheck, Result: false},
					{CheckName: TargetAccountCheck, Result: true},
				},
				TimedOut: false,
			},
			want: false,
		},
		{
			name: "timed out",
			result: PaymentValidationResult{
				Results: []CheckResult{
					{CheckName: BalanceCheck, Result: true},
				},
				TimedOut: true,
			},
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.result.IsValid(); got != tt.want {
				t.Errorf("IsValid() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestPaymentValidationResult_GetFailureReasons(t *testing.T) {
	result := PaymentValidationResult{
		Results: []CheckResult{
			{CheckName: BalanceCheck, Result: false, Error: "insufficient funds"},
			{CheckName: TargetAccountCheck, Result: true},
		},
		TimedOut: false,
	}

	reasons := result.GetFailureReasons()

	if len(reasons) != 1 {
		t.Errorf("Expected 1 failure reason, got %d", len(reasons))
	}

	expectedReason := "balanceCheck failed: insufficient funds"
	if reasons[0] != expectedReason {
		t.Errorf("Expected reason '%s', got '%s'", expectedReason, reasons[0])
	}
}

func TestConfig_Validate(t *testing.T) {
	tests := []struct {
		name    string
		config  Config
		wantErr bool
	}{
		{
			name: "valid config",
			config: Config{
				Server: ServerConfig{Port: "8080"},
				Kafka:  KafkaConfig{Broker: "localhost:9092"},
			},
			wantErr: false,
		},
		{
			name: "missing kafka broker",
			config: Config{
				Server: ServerConfig{Port: "8080"},
				Kafka:  KafkaConfig{Broker: ""},
			},
			wantErr: true,
		},
		{
			name: "invalid port",
			config: Config{
				Server: ServerConfig{Port: "invalid"},
				Kafka:  KafkaConfig{Broker: "localhost:9092"},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

// Benchmark tests
func BenchmarkCreateAccountRequest_Validate(b *testing.B) {
	req := CreateAccountRequest{
		Name:           "Test Account",
		InitialBalance: 1000,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = req.Validate()
	}
}

func BenchmarkPaymentValidationResult_IsValid(b *testing.B) {
	result := PaymentValidationResult{
		Results: []CheckResult{
			{CheckName: BalanceCheck, Result: true},
			{CheckName: TargetAccountCheck, Result: true},
		},
		TimedOut: false,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = result.IsValid()
	}
}
