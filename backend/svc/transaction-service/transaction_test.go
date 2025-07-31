package main

import (
	"context"
	"testing"

	"github.com/bmizerany/assert"
	"github.com/segmentio/kafka-go"
	cmn "github.com/timkins666/distributed-playground/backend/pkg/common"
	tu "github.com/timkins666/distributed-playground/backend/pkg/testutils"
)

type MockDB struct {
	tu.BaseTestDB
	transactions []*cmn.Transaction
}

func (m *MockDB) CommitTransaction(transaction *cmn.Transaction) error {
	m.transactions = append(m.transactions, transaction)
	return nil
}

func TestProcessMessage(t *testing.T) {
	mockDB := MockDB{}

	env := appEnv{
		BaseEnv: cmn.BaseEnv{}.WithCancelCtx(context.Background()).WithDB(&mockDB),
	}

	tests := []struct {
		name    string
		tx      *cmn.Transaction
		wantErr error
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
			name: "success",
			tx: &cmn.Transaction{
				PaymentSysID: "123",
				Amount:       123456,
				AccountID:    42,
			},
			wantErr: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var msgValue []byte

			if tt.tx != nil {
				var err error
				msgValue, err = cmn.ToBytes(*tt.tx)
				if err != nil {
					t.Fatalf("Setup error")
				}
			}

			msg := kafka.Message{Value: msgValue}

			gotErr := processMessage(msg, &env)

			assert.Equal(t, tt.wantErr, gotErr)

			if tt.tx != nil && tt.tx.Valid() {
				assert.Equal(t, 1, len(mockDB.transactions))
				assert.NotEqual(t, "", mockDB.transactions[0].TxID)

				mockDB.transactions[0].TxID = ""
				assert.Equal(t, tt.tx, mockDB.transactions[0])
			} else {
				assert.Equal(t, 0, len(mockDB.transactions))
			}
		})
	}
}
