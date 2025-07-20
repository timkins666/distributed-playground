package common

// defines immutable topic constants

type topics struct{}

func (t *topics) PaymentRequested() string {
	return "payment-requested"
}
func (t *topics) PaymentVerified() string {
	return "payment-verified"
}
func (t *topics) PaymentFailed() string {
	return "payment-failed"
}
func (t *topics) TransactionRequested() string {
	return "transaction-requested"
}
func (t *topics) TransactionComplete() string {
	return "transaction-completed"
}
func (t *topics) TransactionFailed() string {
	return "transaction-failed"
}

var Topics = topics{}
