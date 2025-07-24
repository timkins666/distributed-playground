package main

import (
	"context"
	"testing"

	"github.com/bmizerany/assert"
	cmn "github.com/timkins666/distributed-playground/backend/pkg/common"
	tu "github.com/timkins666/distributed-playground/backend/pkg/testutils"
)

func TestHandleResults(t *testing.T) {
	tests := []struct {
		name      string
		res       *paymentValidationResult
		wantTopic string
	}{
		{
			name: "timed out",
			res: &paymentValidationResult{
				timedOut:       true,
				paymentRequest: &cmn.PaymentRequest{},
			},
			wantTopic: cmn.Topics.PaymentFailed(),
		},
	}

	writer := tu.MockKafkaWriter{}

	env := appEnv{
		BaseEnv: cmn.BaseEnv{}.WithCancelCtx(context.Background()),
		writer:  &writer,
	}

	for _, tt := range tests {
		handleResults(tt.res, env)
		assert.Equal(t, writer.Messages[0].Topic, tt.wantTopic)
	}

}
