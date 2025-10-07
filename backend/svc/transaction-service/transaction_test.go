package main

import (
	"context"
	"errors"
	"testing"

	"github.com/segmentio/kafka-go"
	cmn "github.com/timkins666/distributed-playground/backend/pkg/common"
)

type mockTransactionDB struct {
	transactions []*cmn.Transaction
	accounts     map[int32]*cmn.Account
	commitErr    error
	accountErr   error
}

func (m *mockTransactionDB) commitTransaction(transaction *cmn.Transaction) error {
	if m.commitErr != nil {
		return m.commitErr
	}
	m.transactions = append(m.transactions, transaction)
	return nil
}

func (m *mockTransactionDB) getAccountByID(accountID int32) (*cmn.Account, error) {
	if m.accountErr != nil {
		return nil, m.accountErr
	}
	if acc, ok := m.accounts[accountID]; ok {
		return acc, nil
	}
	return nil, errors.New("account not found")
}

func TestProcessMessage(t *testing.T) {
	tests := []struct {
		name      string
		tx        *cmn.Transaction
		commitErr error
		wantErr   error
	}{
		{
			name:    "invalid message",
			tx:      nil,
			wantErr: errorParsingTransaction,
		},
		{
			name: "invalid transaction",
			tx: &cmn.Transaction{
				PaymentSysID: "",
			},
			wantErr: errorInvalidTransaction,
		},
		{
			name: "commit error",
			tx: &cmn.Transaction{
				PaymentSysID: "123",
				Amount:       100,
				AccountID:    42,
			},
			commitErr: errors.New("db error"),
			wantErr:   errorCommittingTransaction,
		},
		{
			name: "success",
			tx: &cmn.Transaction{
				PaymentSysID: "123",
				Amount:       100,
				AccountID:    42,
			},
			wantErr: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockDB := &mockTransactionDB{
				commitErr: tt.commitErr,
				accounts:  make(map[int32]*cmn.Account),
			}

			appCtx := &transactionCtx{
				cancelCtx: context.Background(),
				logger:    cmn.AppLogger(),
				db:        mockDB,
			}

			var msgValue []byte
			if tt.tx != nil {
				var err error
				msgValue, err = cmn.ToBytes(*tt.tx)
				if err != nil {
					t.Fatalf("setup error: %v", err)
				}
			}

			msg := kafka.Message{
				Value:     msgValue,
				Topic:     "test-topic",
				Partition: 0,
				Offset:    1,
			}

			err := processMessage(msg, appCtx)

			if tt.wantErr != nil {
				if err != tt.wantErr {
					t.Errorf("expected error %v, got %v", tt.wantErr, err)
				}
			} else if err != nil {
				t.Errorf("unexpected error: %v", err)
			}

			if tt.wantErr == nil && tt.tx != nil {
				if len(mockDB.transactions) != 1 {
					t.Errorf("expected 1 transaction, got %d", len(mockDB.transactions))
				}
			}
		})
	}
}

func TestInvalidateCache(t *testing.T) {
	mockDB := &mockTransactionDB{
		accounts: map[int32]*cmn.Account{
			42: {AccountID: 42, UserID: 1, Balance: 1000},
		},
	}

	appCtx := &transactionCtx{
		cancelCtx:   context.Background(),
		db:          mockDB,
		redisClient: nil,
	}

	tx := &cmn.Transaction{AccountID: 42}

	// Should not panic with nil redis client
	invalidateCache(tx, appCtx)
}