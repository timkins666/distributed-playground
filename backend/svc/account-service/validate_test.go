package main

import (
	"bytes"
	"context"
	"encoding/binary"
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
		res        *paymentValidationResult
		wantTopic  cmn.Topic
		wantReason string
	}{
		{
			name: "timed out",
			res: &paymentValidationResult{
				timedOut:       true,
				paymentRequest: pr,
			},
			wantTopic:  cmn.Topics.PaymentFailed(),
			wantReason: "timeout",
		},
		{
			name: "single check failed",
			res: &paymentValidationResult{
				paymentRequest: pr,
				results: []checkResult{
					{checkName: "chk1", result: false},
				},
			},
			wantTopic:  cmn.Topics.PaymentFailed(),
			wantReason: "chk1 failed",
		},
		{
			name: "single check of many failed",
			res: &paymentValidationResult{
				paymentRequest: pr,
				results: []checkResult{
					{checkName: "chk1", result: true},
					{checkName: "chk2", result: false},
					{checkName: "chk3", result: true},
				},
			},
			wantTopic:  cmn.Topics.PaymentFailed(),
			wantReason: "chk2 failed",
		},
		{
			name: "all checks failed",
			res: &paymentValidationResult{
				paymentRequest: pr,
				results: []checkResult{
					{checkName: "chk1", result: false},
					{checkName: "chk2", result: false},
					{checkName: "chk3", result: false},
				},
			},
			wantTopic:  cmn.Topics.PaymentFailed(),
			wantReason: "chk1 failed, chk2 failed, chk3 failed",
		},
	}

	writer := tu.MockKafkaWriter{}

	env := appEnv{
		BaseEnv: cmn.BaseEnv{}.WithCancelCtx(context.Background()),
		writer:  &writer,
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// reset messages
			writer.Messages = []kafka.Message{}

			handleResults(tt.res, env)

			assert.Equal(t, len(writer.Messages), 1)
			assert.Equal(t, writer.Messages[0].Topic, tt.wantTopic.S())

			pm, err := paymentMsg{}.FromBytes(writer.Messages[0].Value)
			if err != nil {
				t.Fatal("error decoding message value")
			}
			assert.Equal(t, pm.AccountID, int32(pr.SourceAccountID))
			assert.Equal(t, pm.SystemID, pr.SystemID)
			assert.Equal(t, pm.Reason, tt.wantReason)

			buf := new(bytes.Buffer)
			buf.Write(writer.Messages[0].Key)

			var key int32
			if err := binary.Read(buf, binary.LittleEndian, &key); err != nil {
				t.Fatal("error decoding key", err)
			}
			assert.Equal(t, int32(key), pr.SourceAccountID)
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
		res  *paymentValidationResult
	}{{
		name: "all checks passed",
		res: &paymentValidationResult{
			paymentRequest: pr,
			results: []checkResult{
				{checkName: "chk1", result: true},
				{checkName: "chk2", result: true},
			},
		},
	},
	}

	writer := tu.MockKafkaWriter{}

	env := appEnv{
		BaseEnv: cmn.BaseEnv{}.WithCancelCtx(context.Background()),
		writer:  &writer,
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// reset messages
			writer.Messages = []kafka.Message{}

			handleResults(tt.res, env)

			assert.Equal(t, len(writer.Messages), 2)
			m1 := writer.Messages[0]
			m2 := writer.Messages[1]

			assert.Equal(t, m1.Topic, cmn.Topics.TransactionRequested().S())
			assert.Equal(t, m2.Topic, cmn.Topics.TransactionRequested().S())

			tx, err := cmn.Transaction{}.FromBytes(m1.Value)
			if err != nil {
				t.Fatal("error decoding message value")
			}
			assert.Equal(t, tx.AccountID, int32(pr.SourceAccountID))
			assert.Equal(t, tx.PaymentSysID, pr.SystemID)

			buf := new(bytes.Buffer)
			buf.Write(writer.Messages[0].Key)

			var key int32
			if err := binary.Read(buf, binary.LittleEndian, &key); err != nil {
				t.Fatal("error decoding key", err)
			}
			assert.Equal(t, int32(key), pr.SourceAccountID)
		})
	}
}
