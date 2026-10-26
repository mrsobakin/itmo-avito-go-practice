package service

import (
	"context"

	"github.com/course-go-autumn-2026/practice6-variant4-subscription/internal/domain"
)

type SubscriptionService struct {
	subscriptions SubscriptionRepository
	accounts      AccountRepository
	fraud         CardFraudChecker
	billing       BillingProvider
}

func NewSubscriptionService(
	subscriptions SubscriptionRepository,
	accounts AccountRepository,
	fraud CardFraudChecker,
	billing BillingProvider,
) *SubscriptionService {
	return &SubscriptionService{
		subscriptions: subscriptions,
		accounts:      accounts,
		fraud:         fraud,
		billing:       billing,
	}
}

func (s *SubscriptionService) Subscribe(ctx context.Context, accountID int64, rawPlan string) (*domain.Subscription, error) {
	plan, err := domain.ParsePlan(rawPlan)
	if err != nil {
		return nil, err
	}

	if s.isCardBlocked(ctx) {
		return nil, domain.ErrCardFraudulent
	}

	sub := domain.NewSubscription(accountID, plan)
	id, err := s.subscriptions.Create(ctx, sub)
	if err != nil {
		return nil, err
	}
	sub.AssignID(id)
	return sub, nil
}

func (s *SubscriptionService) Get(ctx context.Context, id int64) (*domain.Subscription, error) {
	return s.subscriptions.Get(ctx, id)
}

func (s *SubscriptionService) ChangePlan(ctx context.Context, id int64, rawPlan string) (*domain.Subscription, error) {
	plan, err := domain.ParsePlan(rawPlan)
	if err != nil {
		return nil, err
	}

	sub, err := s.subscriptions.Get(ctx, id)
	if err != nil {
		return nil, err
	}

	account, err := s.accounts.Get(ctx, sub.AccountID())
	if err != nil {
		return nil, err
	}

	needsBilling, err := sub.ChangePlan(plan, account)
	if err != nil {
		return nil, err
	}

	if needsBilling && !s.isCharged(ctx) {
		if err := sub.MarkPaymentFailed(); err != nil {
			return nil, err
		}
	}

	if err := s.subscriptions.Save(ctx, sub); err != nil {
		return nil, err
	}
	return sub, nil
}

func (s *SubscriptionService) isCardBlocked(ctx context.Context) bool {
	blocked, err := s.fraud.IsBlocked(ctx)
	if err != nil {
		return false
	}
	return blocked
}

func (s *SubscriptionService) isCharged(ctx context.Context) bool {
	return s.billing.Charge(ctx) == nil
}
