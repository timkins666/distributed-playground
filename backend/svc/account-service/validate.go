package main

import (
	"context"
	"fmt"
	"math/rand/v2"
	"strings"
	"time"

	"github.com/segmentio/kafka-go"
	cmn "github.com/timkins666/distributed-playground/backend/pkg/common"
)

type CheckName string

const (
	BalanceCheck       CheckName = "balanceCheck"
	TargetAccountCheck CheckName = "targetAccountCheck"
)

type PaymentMsgType string

const (
	PaymentFailed PaymentMsgType = "paymentFailed"
)

// result of a validation check
type CheckResult struct {
	CheckName CheckName `json:"checkName"`
	Result    bool      `json:"result"`
	Error     string    `json:"error,omitempty"`
}

// aggregates all validation results
type PaymentValidationResult struct {
	PaymentRequest *cmn.PaymentRequest `json:"paymentRequest"`
	Results        []CheckResult       `json:"results"`
	TimedOut       bool                `json:"timedOut"`
	StartTime      time.Time           `json:"startTime"`
	EndTime        time.Time           `json:"endTime"`
}

// a payment failure message
type PaymentMsg struct {
	Type      PaymentMsgType `json:"type"`
	Reason    string         `json:"reason"`
	AppID     string         `json:"appId"`
	SystemID  string         `json:"systemId"`
	AccountID int32          `json:"accountId"`
	Timestamp time.Time      `json:"timestamp"`
}

// populates PaymentMsg from PaymentRequest
func (pm *PaymentMsg) FromReq(req *cmn.PaymentRequest) *PaymentMsg {
	pm.AccountID = req.SourceAccountID
	pm.AppID = req.AppID
	pm.SystemID = req.SystemID
	pm.Timestamp = time.Now()
	return pm
}

// checks if a result indicates success
func (pvr *PaymentValidationResult) IsValid() bool {
	if pvr.TimedOut {
		return false
	}
	for _, result := range pvr.Results {
		if !result.Result {
			return false
		}
	}
	return true
}

// returns a list of failure reasons
func (pvr *PaymentValidationResult) GetFailureReasons() []string {
	var reasons []string
	if pvr.TimedOut {
		reasons = append(reasons, "validation timeout")
	}
	for _, result := range pvr.Results {
		if !result.Result {
			reason := fmt.Sprintf("%s failed", result.CheckName)
			if result.Error != "" {
				reason += ": " + result.Error
			}
			reasons = append(reasons, reason)
		}
	}
	return reasons
}

// handles validation of requested payments
func paymentValidator(appCtx *accountsCtx) {
	max_errors := 10
	for {
		select {
		case <-appCtx.cancelCtx.Done():
			appCtx.logger.Println("Context cancelled")
			return
		default:
			msg, err := appCtx.payReqReader.ReadMessage(context.Background())
			if err != nil {
				appCtx.logger.Println("READ MSG ERROR", err)
				max_errors--
				if max_errors == 0 {
					appCtx.logger.Println("Max errors reached, something seems wrong...")
					break
				}
				continue
			}
			go handlePaymentRequestedMessage(msg, appCtx)
		}
	}
}

// handlePaymentRequestedMessage processes a payment request message
func handlePaymentRequestedMessage(message kafka.Message, appCtx *accountsCtx) {
	req, err := cmn.FromBytes[cmn.PaymentRequest](message.Value)
	if err != nil {
		appCtx.logger.Printf("Failed to parse payment request message: %v", err)
		return
	}

	appCtx.logger.Printf("Processing payment request: %s (amount: %d, from: %d, to: %d)",
		req.SystemID, req.Amount, req.SourceAccountID, req.TargetAccountID)

	if !req.Valid() {
		appCtx.logger.Printf("Invalid payment request: %s", req.SystemID)
		return
	}

	validationResult := &PaymentValidationResult{
		PaymentRequest: req,
		StartTime:      time.Now(),
	}

	const numChecks = 2
	results := make(chan CheckResult, numChecks)

	// Start validation checks concurrently
	go checkBalance(req, results, appCtx)
	go checkTargetAccount(req, results, appCtx)

	// Collect results with timeout
	timeout := 4500 * time.Millisecond
	for {
		select {
		case res := <-results:
			appCtx.logger.Printf("Check %s completed: %t", res.CheckName, res.Result)
			validationResult.Results = append(validationResult.Results, res)

			if len(validationResult.Results) == numChecks {
				validationResult.EndTime = time.Now()
				appCtx.logger.Printf("All checks completed for request %s in %v",
					req.SystemID, validationResult.EndTime.Sub(validationResult.StartTime))
				go handleValidationResults(validationResult, appCtx)
				return
			}
			appCtx.logger.Printf("Waiting for %d more checks for request %s",
				numChecks-len(validationResult.Results), req.SystemID)

		case <-time.After(timeout):
			validationResult.TimedOut = true
			validationResult.EndTime = time.Now()
			appCtx.logger.Printf("Validation timed out for request %s after %v (completed %d/%d checks)",
				req.SystemID, timeout, len(validationResult.Results), numChecks)
			go handleValidationResults(validationResult, appCtx)
			return

		case <-appCtx.cancelCtx.Done():
			appCtx.logger.Printf("Context cancelled while processing request %s", req.SystemID)
			return
		}
	}
}

// processes the validation results
func handleValidationResults(result *PaymentValidationResult, appCtx *accountsCtx) {
	if !result.IsValid() {
		reasons := result.GetFailureReasons()
		reason := strings.Join(reasons, ", ")
		sendPaymentFailed(result.PaymentRequest, reason, appCtx)
		return
	}

	appCtx.logger.Printf("Payment validation successful for request %s", result.PaymentRequest.SystemID)
	// TODO: Implement fund locking to prevent race conditions before transaction
	initiateTransaction(result.PaymentRequest, appCtx)
}

// publishes a payment failure message to Kafka
func sendPaymentFailed(req *cmn.PaymentRequest, reason string, appCtx *accountsCtx) {
	appCtx.logger.Printf("Payment failed - Amount: %d, From: %d, To: %d, Reason: %s",
		req.Amount, req.SourceAccountID, req.TargetAccountID, reason)

	msg := (&PaymentMsg{Type: PaymentFailed, Reason: reason}).FromReq(req)

	key, err := cmn.ToBytes(msg.AccountID)
	if err != nil {
		appCtx.logger.Printf("Failed to serialize account ID for payment failure: %v", err)
		return
	}

	val, err := cmn.ToBytes(msg)
	if err != nil {
		appCtx.logger.Printf("Failed to serialize payment failure message: %v", err)
		return
	}

	if err := appCtx.writer.WriteMessages(appCtx.cancelCtx, kafka.Message{
		Topic: cmn.Topics.PaymentFailed().S(),
		Key:   key,
		Value: val,
	}); err != nil {
		appCtx.logger.Printf("Failed to publish payment failure message: %v", err)
	}
}

// send message(s) for transaction service
func initiateTransaction(req *cmn.PaymentRequest, appCtx *accountsCtx) {
	appCtx.logger.Printf("Initiate transaction of £%d from account %d to account %d", req.Amount, req.SourceAccountID, req.TargetAccountID)

	txOut := cmn.Transaction{
		PaymentSysID: req.SystemID,
		AccountID:    req.SourceAccountID,
		Amount:       -req.Amount,
	}
	kOut, errkOut := cmn.ToBytes(txOut.AccountID)
	vOut, errvOut := cmn.ToBytes(txOut)

	txIn := cmn.Transaction{
		PaymentSysID: req.SystemID,
		AccountID:    req.TargetAccountID,
		Amount:       req.Amount,
	}
	kIn, errkIn := cmn.ToBytes(txIn.AccountID)
	vIn, errvIn := cmn.ToBytes(txIn)

	if errvOut != nil || errkOut != nil || errvIn != nil || errkIn != nil {
		sendPaymentFailed(req, "processing error", appCtx)
		return
	}

	err := appCtx.writer.WriteMessages(appCtx.cancelCtx,
		kafka.Message{
			Topic: cmn.Topics.TransactionRequested().S(),
			Key:   kOut,
			Value: vOut,
		},
		kafka.Message{
			Topic: cmn.Topics.TransactionRequested().S(),
			Key:   kIn,
			Value: vIn,
		})

	if err != nil {
		// TODO: what if one message sent
		sendPaymentFailed(req, "failed to initiate transaction", appCtx)
	}

}

// verifies that the source account has sufficient funds
func checkBalance(req *cmn.PaymentRequest, chn chan<- CheckResult, appCtx *accountsCtx) {
	// artificial delay for simulation
	sleep := rand.N(5000)
	appCtx.logger.Printf("%s sleeping for %dms", BalanceCheck, sleep)
	time.Sleep(time.Duration(sleep) * time.Millisecond)

	res := CheckResult{CheckName: BalanceCheck}

	srcAcc, err := appCtx.db.getAccountByID(req.SourceAccountID)
	if err != nil {
		appCtx.logger.Printf("Source account %d not found: %v", req.SourceAccountID, err)
		res.Result = false
		res.Error = fmt.Sprintf("account not found: %v", err)
		chn <- res
		return
	}

	appCtx.logger.Printf("Balance check - Account: %d, Balance: %d, Required: %d",
		srcAcc.AccountID, srcAcc.Balance, req.Amount)

	if srcAcc.Balance >= req.Amount {
		res.Result = true
		appCtx.logger.Printf("Balance check passed for account %d", srcAcc.AccountID)
	} else {
		res.Result = false
		res.Error = fmt.Sprintf("insufficient funds: has %d, needs %d", srcAcc.Balance, req.Amount)
		appCtx.logger.Printf("Balance check failed for account %d: %s", srcAcc.AccountID, res.Error)
	}

	chn <- res
}

// verifies that the target account exists
func checkTargetAccount(req *cmn.PaymentRequest, chn chan<- CheckResult, appCtx *accountsCtx) {
	// artificial delay for simulation
	sleep := rand.N(5000)
	appCtx.logger.Printf("%s sleeping for %dms", TargetAccountCheck, sleep)
	time.Sleep(time.Duration(sleep) * time.Millisecond)

	res := CheckResult{CheckName: TargetAccountCheck}

	_, err := appCtx.db.getAccountByID(req.TargetAccountID)
	if err != nil {
		appCtx.logger.Printf("Target account %d not found: %v", req.TargetAccountID, err)
		res.Result = false
	} else {
		res.Result = true
		appCtx.logger.Printf("Target account %d validated successfully", req.TargetAccountID)
	}

	chn <- res
}
