package main

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"


	cmn "github.com/timkins666/distributed-playground/backend/pkg/common"
	tu "github.com/timkins666/distributed-playground/backend/pkg/testutils"
)

type mockDB struct {
	createPaymentErr error
}

func (m *mockDB) createPayment(*cmn.PaymentRequest) error {
	return m.createPaymentErr
}

func TestHandlePaymentRequest(t *testing.T) {
	tests := []struct {
		name           string
		request        cmn.PaymentRequest
		dbErr          error
		writerErr      error
		expectedStatus int
	}{
		{
			name: "valid payment request",
			request: cmn.PaymentRequest{
				SourceAccountID: 123,
				TargetAccountID: 789,
				Amount:          10050,
				AppID:           "aID",
				SystemID:        "sID",
			},
			expectedStatus: http.StatusAccepted,
		},
		{
			name: "invalid request - missing source account",
			request: cmn.PaymentRequest{
				TargetAccountID: 789,
				Amount:          10050,
				AppID:           "aID",
				SystemID:        "sID",
			},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "database error",
			request: cmn.PaymentRequest{
				SourceAccountID: 123,
				TargetAccountID: 789,
				Amount:          10050,
				AppID:           "aID",
				SystemID:        "sID",
			},
			dbErr:          errors.New("oh no"),
			expectedStatus: http.StatusInternalServerError,
		},
		{
			name: "kafka write error",
			request: cmn.PaymentRequest{
				SourceAccountID: 123,
				TargetAccountID: 789,
				Amount:          10050,
				AppID:           "aID",
				SystemID:        "sID",
			},
			writerErr:      errors.New("oh shucks"),
			expectedStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockDB := &mockDB{createPaymentErr: tt.dbErr}
			mockWriter := &tu.MockKafkaWriter{WriteErr: tt.writerErr}

			appCtx := &paymentCtx{
				cancelCtx: context.Background(),
				db:        mockDB,
				writer:    mockWriter,
				logger:    cmn.AppLogger(),
			}

			body, _ := json.Marshal(tt.request)
			req := httptest.NewRequest("POST", "/transfer", bytes.NewReader(body))
			req = req.WithContext(context.WithValue(req.Context(), cmn.AppCtx, appCtx))

			w := httptest.NewRecorder()
			handlePaymentRequest(w, req)

			if w.Code != tt.expectedStatus {
				t.Errorf("expected status %d, got %d", tt.expectedStatus, w.Code)
			}

			if tt.expectedStatus == http.StatusAccepted && len(mockWriter.Messages) != 1 {
				t.Errorf("expected 1 kafka message, got %d", len(mockWriter.Messages))
			}
		})
	}
}

func TestHandlePaymentRequestInvalidJSON(t *testing.T) {
	appCtx := &paymentCtx{
		cancelCtx: context.Background(),
		db:        &mockDB{},
		writer:    &tu.MockKafkaWriter{},
		logger:    cmn.AppLogger(),
	}

	req := httptest.NewRequest("POST", "/transfer", bytes.NewReader([]byte("invalid json")))
	req = req.WithContext(context.WithValue(req.Context(), cmn.AppCtx, appCtx))

	w := httptest.NewRecorder()
	handlePaymentRequest(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected status %d, got %d", http.StatusBadRequest, w.Code)
	}
}

func TestHandlePaymentRequestMissingContext(t *testing.T) {
	req := httptest.NewRequest("POST", "/transfer", nil)
	w := httptest.NewRecorder()
	handlePaymentRequest(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Errorf("expected status %d, got %d", http.StatusInternalServerError, w.Code)
	}
}
