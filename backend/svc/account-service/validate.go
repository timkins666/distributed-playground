package main

import (
	"bytes"
	"context"
	"encoding/binary"
	"encoding/gob"
	"encoding/json"
	"fmt"
	"math/rand/v2"
	"strings"
	"time"

	"github.com/segmentio/kafka-go"
	cmn "github.com/timkins666/distributed-playground/backend/pkg/common"
)

type checkName string

var (
	balanceCheck  checkName = "balanceCheck"
	targetAccount checkName = "targetAccount"
)

type paymentMsgType string

var (
	paymentFailed paymentMsgType = "paymentFailed"
)

type checkResult struct {
	checkName checkName
	result    bool
}

type paymentValidationResult struct {
	paymentRequest *cmn.PaymentRequest
	results        []checkResult
	timedOut       bool
}

type paymentMsg struct {
	Type      paymentMsgType
	Reason    string
	AppID     string
	SystemID  string
	AccountID int32
}

func (pm *paymentMsg) FromReq(req *cmn.PaymentRequest) *paymentMsg {
	pm.AccountID = int32(req.SourceAccountID)
	pm.AppID = req.AppID
	pm.SystemID = req.SystemID
	return pm
}
func (pm paymentMsg) FromBytes(b []byte) (*paymentMsg, error) {
	buf := new(bytes.Buffer)
	buf.Write(b)
	err := gob.NewDecoder(buf).Decode(&pm)
	return &pm, err
}
func (pm *paymentMsg) Key() ([]byte, error) {
	buf := new(bytes.Buffer)
	err := binary.Write(buf, binary.LittleEndian, pm.AccountID)
	return buf.Bytes(), err
}
func (pm *paymentMsg) Value() ([]byte, error) {
	buf := new(bytes.Buffer)
	err := gob.NewEncoder(buf).Encode(pm)
	return buf.Bytes(), err
}

// handles validation of requested payments
func paymentValidator(env appEnv) {
	max_errors := 10
	for {
		select {
		case <-env.CancelCtx().Done():
			env.Logger().Println("Context cancelled")
			return
		default:
			msg, err := env.payReqReader.ReadMessage(context.Background())
			if err != nil {
				env.Logger().Println("READ MSG ERROR", err)
				max_errors--
				if max_errors == 0 {
					env.Logger().Println("Max errors reached, something seems wrong...")
					break
				}
				continue
			}
			go handlePaymentRequestedMessage(msg, env)
		}
	}
}

func handlePaymentRequestedMessage(message kafka.Message, env appEnv) {
	var req cmn.PaymentRequest
	err := json.Unmarshal(message.Value, &req)
	if err != nil {
		env.Logger().Println("ERROR failed to parse message:", err)
		return
	}
	env.Logger().Println("Read message: ", req)

	validateResult := &paymentValidationResult{
		paymentRequest: &req,
	}

	numChecks := 2 // being lazy
	results := make(chan checkResult, numChecks)
	go checkBalance(&req, results, env)
	go checkTargetAccount(&req, results, env)

	for {
		select {
		case res := <-results:
			// update central results here for single update point
			env.Logger().Println(res.checkName, ":", res.result)
			validateResult.results = append(validateResult.results, res)

			if len(validateResult.results) == numChecks {
				env.Logger().Println("all checks complete for request", req.SystemID)
				go handleResults(validateResult, env)
				return
			}
			env.Logger().Printf("waiting for %d remaining check for request %s", numChecks-len(validateResult.results), req.SystemID)
		case <-time.After(4500 * time.Millisecond): // 4.5s timeout to return so some will fail
			env.Logger().Println("Checks timed out for request", req.SystemID)
			validateResult.timedOut = true
			go handleResults(validateResult, env)
			return
		case <-env.CancelCtx().Done():
			return
		}
	}
}

func handleResults(result *paymentValidationResult, env appEnv) {
	if result.timedOut {
		sendPaymentFailed(result.paymentRequest, "timeout", env)
		return
	}

	errs := []string{}
	for _, check := range result.results {
		if !check.result {
			errs = append(errs, fmt.Sprintf("%s failed", check.checkName))
		}
	}

	if len(errs) > 0 {
		sendPaymentFailed(result.paymentRequest, strings.Join(errs, ", "), env)
		return
	}

	// TODO: lock funds to prevent races before submitting transaction
	initiateTransaction(result.paymentRequest, env)
}

// sends a message to the payment failed topic
func sendPaymentFailed(req *cmn.PaymentRequest, reason string, env appEnv) {
	env.Logger().Printf("Payment of £%d failed for account %d: %s", req.Amount, req.TargetAccountID, reason)
	msg := (&paymentMsg{Type: paymentFailed, Reason: reason}).FromReq(req)

	key, err := msg.Key()
	if err != nil {
		env.Logger().Println(err)
		return
	}
	val, err := msg.Value()
	if err != nil {
		env.Logger().Println(err)
		return
	}

	//TODO: write error handling
	_ = env.writer.WriteMessages(env.CancelCtx(), kafka.Message{
		Topic: cmn.Topics.PaymentFailed().S(),
		Key:   key,
		Value: val,
	})

}

// send message(s) for transaction service
func initiateTransaction(req *cmn.PaymentRequest, env appEnv) {
	env.Logger().Printf("Initiate transaction of £%d from account %d to account %d", req.Amount, req.SourceAccountID, req.TargetAccountID)

	txOut, err1 := json.Marshal(cmn.Transaction{
		PaymentSysID: req.SystemID,
		AccountID:    req.SourceAccountID,
		Amount:       -req.Amount,
	})
	txIn, err2 := json.Marshal(cmn.Transaction{
		PaymentSysID: req.SystemID,
		AccountID:    req.TargetAccountID,
		Amount:       req.Amount,
	})

	if err1 != nil || err2 != nil {
		sendPaymentFailed(req, "processing error", env)
	}

	err := env.writer.WriteMessages(env.CancelCtx(),
		kafka.Message{
			Topic: cmn.Topics.TransactionRequested().S(),
			Value: txOut,
		},
		kafka.Message{
			Topic: cmn.Topics.TransactionRequested().S(),
			Value: txIn,
		})

	if err != nil {
		sendPaymentFailed(req, "failed to initiate transaction", env)
	}

}

// check source account has required funds
func checkBalance(req *cmn.PaymentRequest, chn chan<- checkResult, env appEnv) {
	// artificial delay
	sleep := rand.N(5000)
	env.Logger().Printf("%s sleeping for %d", balanceCheck, sleep)
	time.Sleep(time.Duration(sleep) * time.Millisecond)

	result := checkResult{checkName: balanceCheck}

	srcAcc, err := env.DB().GetAccountByID(req.SourceAccountID)

	if err != nil {
		env.Logger().Printf("ERROR: Account id %d not found. %s", req.SourceAccountID, err)
		result.result = false // TODO: unneccessary cos default but wait for tests
		chn <- result
		return
	}

	env.Logger().Printf("Account %d current balance £%d, requested payment of £%d", srcAcc.AccountID, srcAcc.Balance, req.Amount)
	result.result = srcAcc.Balance >= req.Amount
	chn <- result
}

// check target account exists
func checkTargetAccount(req *cmn.PaymentRequest, chn chan<- checkResult, env appEnv) {
	// artificial delay
	sleep := rand.N(5000)
	env.Logger().Printf("%s sleeping for %d", targetAccount, sleep)
	time.Sleep(time.Duration(sleep) * time.Millisecond)

	result := checkResult{checkName: targetAccount}

	_, err := env.DB().GetAccountByID(req.SourceAccountID)

	if err != nil {
		env.Logger().Printf("ERROR: Target account id %d not found. %s", req.TargetAccountID, err)
		result.result = false // TODO: unneccessary cos default but wait for tests
		chn <- result
		return
	}

	result.result = true
	chn <- result
}
