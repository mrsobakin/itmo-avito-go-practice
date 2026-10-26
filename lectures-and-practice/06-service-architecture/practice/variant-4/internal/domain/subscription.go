package domain

import "time"

type Status string

const (
	StatusTrial          Status = "trial"
	StatusActive         Status = "active"
	StatusSuspended      Status = "suspended"
	StatusPaymentPending Status = "payment_pending"
)

type Subscription struct {
	id        int64
	accountID int64
	plan      Plan
	status    Status
	createdAt time.Time
}

func NewSubscription(accountID int64, plan Plan) *Subscription {
	return &Subscription{
		accountID: accountID,
		plan:      plan,
		status:    StatusTrial,
		createdAt: time.Now(),
	}
}

func RestoreSubscription(id, accountID int64, plan Plan, status Status, createdAt time.Time) *Subscription {
	return &Subscription{
		id:        id,
		accountID: accountID,
		plan:      plan,
		status:    status,
		createdAt: createdAt,
	}
}

func (s *Subscription) ID() int64            { return s.id }
func (s *Subscription) AccountID() int64     { return s.accountID }
func (s *Subscription) Plan() Plan           { return s.plan }
func (s *Subscription) Status() Status       { return s.status }
func (s *Subscription) CreatedAt() time.Time { return s.createdAt }

func (s *Subscription) AssignID(id int64) { s.id = id }

func (s *Subscription) ChangePlan(plan Plan, account Account) (needsBilling bool, err error) {
	if account.ID != s.accountID {
		return false, ErrAccountMismatch
	}
	if s.status == StatusSuspended {
		return false, ErrSuspended
	}
	if plan.IsDowngradeFrom(s.plan) && account.HasDebt {
		return false, ErrDowngradeWithDebt
	}

	needsBilling = plan.IsUpgradeFrom(s.plan) && s.status == StatusActive
	s.plan = plan
	return needsBilling, nil
}

func (s *Subscription) MarkPaymentFailed() error {
	if s.status != StatusActive {
		return ErrInvalidTransition
	}
	s.status = StatusPaymentPending
	return nil
}
