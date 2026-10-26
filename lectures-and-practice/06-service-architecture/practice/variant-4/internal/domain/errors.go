package domain

import "errors"

var (
	ErrUnknownPlan       = errors.New("unknown plan")
	ErrCardFraudulent    = errors.New("card looks fraudulent")
	ErrNotFound          = errors.New("subscription not found")
	ErrSuspended         = errors.New("cannot change plan of a suspended subscription")
	ErrDowngradeWithDebt = errors.New("cannot downgrade plan with an active debt")
	ErrAccountMismatch   = errors.New("account does not own subscription")
	ErrInvalidTransition = errors.New("invalid subscription status transition")
)
