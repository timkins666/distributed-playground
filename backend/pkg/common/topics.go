package common

// defines grouped immutable topic constants, possibly not very idiomatic

type Topic string

func (t Topic) S() string {
	return string(t)
}

type topics struct{}

func (t *topics) PaymentRequested() Topic {
	return "payment-requested"
}
func (t *topics) PaymentVerified() Topic {
	return "payment-verified"
}
func (t *topics) PaymentFailed() Topic {
	return "payment-failed"
}
func (t *topics) TransactionRequested() Topic {
	return "transaction-requested"
}
func (t *topics) TransactionComplete() Topic {
	return "transaction-completed"
}
func (t *topics) TransactionFailed() Topic {
	return "transaction-failed"
}

var Topics = topics{}
