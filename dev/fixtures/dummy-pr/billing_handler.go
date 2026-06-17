package dummypr

import "errors"

var ErrForbidden = errors.New("forbidden")

type BillingRequest struct {
	AccountID   string
	AmountCents int
}

func HandleBillingRequest(account Account, request BillingRequest) error {
	if request.AmountCents <= 0 {
		return errors.New("amount must be positive")
	}
	if !CanAccessBilling(account, request.AccountID) {
		return ErrForbidden
	}
	return nil
}
