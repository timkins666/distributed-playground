package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"math/rand/v2"
	"strings"
	"time"

	"github.com/segmentio/kafka-go"
	cmn "github.com/timkins666/distributed-playground/backend/pkg/common"
)

type checkResult struct {
	checkName string
	result    bool
}

type paymentValidationResult struct {
	paymentRequest *cmn.PaymentRequest
	results        []checkResult
	timedOut       bool
}

func paymentValidator(kafkaBroker string, cancelCtx context.Context) {
	// handles validation of requested payments
	reader := kafka.NewReader(kafka.ReaderConfig{
		Brokers: []string{kafkaBroker},
		Topic:   cmn.Topics.PaymentRequested(),
		GroupID: "payment-validator",
	})
	defer reader.Close()

	max_errors := 10
	for {
		select {
		case <-cancelCtx.Done():
			log.Println("Context cancelled")
			return
		default:
			msg, err := reader.ReadMessage(context.Background())
			if err != nil {
				log.Println("READ MSG ERROR", err)
				max_errors -= 1
				if max_errors == 0 {
					log.Println("Max errors reached, something seems wrong...")
					break
				}
				continue
			}
			go handlePaymentRequestedMessage(msg, cancelCtx)
		}
	}
}

func handlePaymentRequestedMessage(message kafka.Message, cancelCtx context.Context) {
	var req cmn.PaymentRequest
	err := json.Unmarshal(message.Value, &req)
	if err != nil {
		log.Println("ERROR failed to parse message:", err)
		return
	}
	log.Println("Read message: ", req)

	validateResult := &paymentValidationResult{
		paymentRequest: &req,
	}

	numChecks := 2 // being lazy
	results := make(chan checkResult, numChecks)
	go checkBalance(&req, results)
	go checkTargetAccount(&req, results)

	for {
		select {
		case res := <-results:
			// update central results here for single update point
			log.Println(res.checkName, ":", res.result)
			validateResult.results = append(validateResult.results, res)

			if len(validateResult.results) == numChecks {
				log.Println("all checks complete for request", req.SystemId)
				go handleResults(validateResult, cancelCtx)
				return
			}
			log.Printf("waiting for %d remaining check for request %s", numChecks-len(validateResult.results), req.SystemId)
		case <-time.After(4500 * time.Millisecond): // 4.5s timeout to return so some will fail
			log.Println("Checks timed out for request", req.SystemId)
			validateResult.timedOut = true
			go handleResults(validateResult, cancelCtx)
			return
		case <-cancelCtx.Done():
			return
		}
	}
}

func handleResults(result *paymentValidationResult, cancelCtx context.Context) {
	if result.timedOut {
		sendPaymentFailed(result.paymentRequest, "timeout", cancelCtx)
		return
	}

	errs := []string{}
	for _, check := range result.results {
		if !check.result {
			errs = append(errs, fmt.Sprintf("%s failed", check.checkName))
		}
	}

	if len(errs) > 0 {
		go sendPaymentFailed(result.paymentRequest, strings.Join(errs, ", "), cancelCtx)
		return
	}

	go initiateTransaction(result.paymentRequest, cancelCtx)
}

func sendPaymentFailed(req *cmn.PaymentRequest, reason string, _ context.Context) {
	// send payment failed message for gateway or future notification service
	log.Printf("Payment of £%f failed for account %d: %s", req.Amount, req.TargetAccountId, reason)
}

func initiateTransaction(req *cmn.PaymentRequest, _ context.Context) {
	// send message for transaction service
	log.Printf("Initiate transaction of £%f from account %d to account %d", req.Amount, req.SourceAccountId, req.TargetAccountId)
}

func checkBalance(req *cmn.PaymentRequest, chn chan<- checkResult) {
	// check source account has funds
	checkName := "balance"

	// artificial delay
	sleep := rand.N(5000)
	log.Printf("%s sleeping for %d", checkName, sleep)
	time.Sleep(time.Duration(sleep) * time.Millisecond)

	result := checkResult{checkName: checkName}

	var srcAcc *Account
	for _, acc := range openAccounts {
		if acc.AccountId == req.SourceAccountId {
			srcAcc = &acc
			break
		}
	}

	if srcAcc == nil {
		log.Printf("ERROR: Account id %d not found", req.SourceAccountId)
		result.result = false // TODO: unneccessary cos default but wait for tests
		chn <- result
		return
	}

	log.Printf("Account %d current balance £%f, requested payment of £%f", srcAcc.AccountId, srcAcc.Balance, req.Amount)
	result.result = srcAcc.Balance >= float64(req.Amount)
	chn <- result
}

func checkTargetAccount(req *cmn.PaymentRequest, chn chan<- checkResult) {
	// check target account is valid
	checkName := "targetAccount"

	// artificial delay
	sleep := rand.N(5000)
	log.Printf("%s sleeping for %d", checkName, sleep)
	time.Sleep(time.Duration(sleep) * time.Millisecond)

	result := checkResult{checkName: checkName}

	var tgtAcc *Account
	for _, acc := range openAccounts {
		if acc.AccountId == req.SourceAccountId {
			tgtAcc = &acc
			break
		}
	}

	if tgtAcc == nil {
		log.Printf("ERROR: Target account id %d not found", req.TargetAccountId)
		result.result = false
		chn <- result
		return
	}

	result.result = true
	chn <- result
}
