package service

import (
	"context"

	"github.com/course-go-autumn-2026/practice6-variant4-subscription/internal/domain"
)

type SubscriptionRepository interface {
	Create(ctx context.Context, s *domain.Subscription) (int64, error)
	Save(ctx context.Context, s *domain.Subscription) error
	Get(ctx context.Context, id int64) (*domain.Subscription, error)
}

type AccountRepository interface {
	Get(ctx context.Context, id int64) (domain.Account, error)
}

type CardFraudChecker interface {
	IsBlocked(ctx context.Context) (bool, error)
}

type BillingProvider interface {
	Charge(ctx context.Context) error
}
