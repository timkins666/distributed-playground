package main

import (
	"context"
	"testing"

	"github.com/bmizerany/assert"
	"github.com/segmentio/kafka-go"
	cmn "github.com/timkins666/distributed-playground/backend/pkg/common"
	tu "github.com/timkins666/distributed-playground/backend/pkg/testutils"
)

func TestHandleResultsFailedChecks(t *testing.T) {
	sysID := "sys123"
	appID := "app456"
	pr := &cmn.PaymentRequest{
		SystemID:        sysID,
		AppID:           appID,
		SourceAccountID: 99,
	}

	tests := []struct {
		name       string
		res        *PaymentValidationResult
		wantTopic  cmn.Topic
		wantReason string
	}{
		{
			name: "timed out",
			res: &PaymentValidationResult{
				TimedOut:       true,
				PaymentRequest: pr,
			},
			wantTopic:  cmn.Topics.PaymentFailed(),
			wantReason: "validation timeout",
		},
		{
			name: "single check failed",
			res: &PaymentValidationResult{
				PaymentRequest: pr,
				Results: []CheckResult{
					{CheckName: "chk1", Result: false},
				},
			},
			wantTopic:  cmn.Topics.PaymentFailed(),
			wantReason: "chk1 failed",
		},
		{
			name: "single check of many failed",
			res: &PaymentValidationResult{
				PaymentRequest: pr,
				Results: []CheckResult{
					{CheckName: "chk1", Result: true},
					{CheckName: "chk2", Result: false},
					{CheckName: "chk3", Result: true},
				},
			},
			wantTopic:  cmn.Topics.PaymentFailed(),
			wantReason: "chk2 failed",
		},
		{
			name: "all checks failed",
			res: &PaymentValidationResult{
				PaymentRequest: pr,
				Results: []CheckResult{
					{CheckName: "chk1", Result: false},
					{CheckName: "chk2", Result: false},
					{CheckName: "chk3", Result: false},
				},
			},
			wantTopic:  cmn.Topics.PaymentFailed(),
			wantReason: "chk1 failed, chk2 failed, chk3 failed",
		},
	}

	writer := tu.MockKafkaWriter{}

	appCtx := accountsCtx{
		cancelCtx: context.Background(),
		writer:    &writer,
		logger:    cmn.AppLogger(),
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// reset messages
			writer.Messages = []kafka.Message{}

			handleValidationResults(tt.res, &appCtx)

			assert.Equal(t, len(writer.Messages), 1)
			assert.Equal(t, writer.Messages[0].Topic, tt.wantTopic.S())

			pm, err := cmn.FromBytes[PaymentMsg](writer.Messages[0].Value)
			if err != nil {
				t.Fatal("error decoding message value")
			}
			assert.Equal(t, int32(pr.SourceAccountID), pm.AccountID)
			assert.Equal(t, pr.SystemID, pm.SystemID)
			assert.Equal(t, tt.wantReason, pm.Reason)

			key, err := cmn.FromBytes[int32](writer.Messages[0].Key)
			if err != nil {
				t.Fatal("error decoding key", err)
			}
			assert.Equal(t, pr.SourceAccountID, *key)
		})
	}
}

func TestHandleResultsChecksPassed(t *testing.T) {
	sysID := "sys123"
	appID := "app456"
	pr := &cmn.PaymentRequest{
		SystemID:        sysID,
		AppID:           appID,
		SourceAccountID: 99,
		TargetAccountID: 666,
	}

	tests := []struct {
		name string
		res  *PaymentValidationResult
	}{{
		name: "all checks passed",
		res: &PaymentValidationResult{
			PaymentRequest: pr,
			Results: []CheckResult{
				{CheckName: "chk1", Result: true},
				{CheckName: "chk2", Result: true},
			},
		},
	},
	}

	writer := tu.MockKafkaWriter{}

	appCtx := accountsCtx{
		cancelCtx: context.Background(),
		writer:    &writer,
		logger:    cmn.AppLogger(),
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// reset messages
			writer.Messages = []kafka.Message{}

			handleValidationResults(tt.res, &appCtx)

			assert.Equal(t, len(writer.Messages), 2)
			m1 := writer.Messages[0]
			m2 := writer.Messages[1]

			assert.Equal(t, m1.Topic, cmn.Topics.TransactionRequested().S())
			assert.Equal(t, m2.Topic, cmn.Topics.TransactionRequested().S())

			tx, err := cmn.FromBytes[cmn.Transaction](m1.Value)
			if err != nil {
				t.Fatal("error decoding message value")
			}
			assert.Equal(t, tx.AccountID, int32(pr.SourceAccountID))
			assert.Equal(t, tx.PaymentSysID, pr.SystemID)

			key, err := cmn.FromBytes[int32](writer.Messages[0].Key)
			if err != nil {
				t.Fatal("error decoding key", err)
			}
			assert.Equal(t, pr.SourceAccountID, *key)
		})
	}
}
