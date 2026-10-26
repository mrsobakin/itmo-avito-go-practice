package service_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/course-go-autumn-2026/practice6-variant4-subscription/internal/domain"
	"github.com/course-go-autumn-2026/practice6-variant4-subscription/internal/repository"
	"github.com/course-go-autumn-2026/practice6-variant4-subscription/internal/service"
)

type fakeRepo struct {
	sub *domain.Subscription
}

func (r *fakeRepo) Create(context.Context, *domain.Subscription) (int64, error) {
	return 0, errors.New("not implemented")
}

func (r *fakeRepo) Save(_ context.Context, s *domain.Subscription) error {
	r.sub = s
	return nil
}

func (r *fakeRepo) Get(_ context.Context, id int64) (*domain.Subscription, error) {
	if r.sub == nil || r.sub.ID() != id {
		return nil, domain.ErrNotFound
	}
	return r.sub, nil
}

type fakeFraud struct{}

func (fakeFraud) IsBlocked(context.Context) (bool, error) { return false, nil }

type fakeBilling struct {
	err   error
	calls int
}

func (f *fakeBilling) Charge(context.Context) error {
	f.calls++
	return f.err
}

func TestChangePlan_ActiveSubscription(t *testing.T) {
	tests := []struct {
		name        string
		from, to    domain.Plan
		accountID   int64
		billingErr  error
		wantErr     error
		wantStatus  domain.Status
		wantCharges int
	}{
		{"upgrade billed", domain.PlanBasic, domain.PlanMax, 1, nil, nil, domain.StatusActive, 1},
		{"upgrade declined", domain.PlanBasic, domain.PlanPro, 1, errors.New("declined"), nil, domain.StatusPaymentPending, 1},
		{"downgrade without debt", domain.PlanMax, domain.PlanBasic, 1, nil, nil, domain.StatusActive, 0},
		{"downgrade with debt", domain.PlanMax, domain.PlanPro, 2, nil, domain.ErrDowngradeWithDebt, domain.StatusActive, 0},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &fakeRepo{sub: domain.RestoreSubscription(1, tt.accountID, tt.from, domain.StatusActive, time.Now())}
			billing := &fakeBilling{err: tt.billingErr}
			svc := service.NewSubscriptionService(repo, repository.NewMemoryAccountRepository(), fakeFraud{}, billing)

			_, err := svc.ChangePlan(context.Background(), 1, string(tt.to))
			if !errors.Is(err, tt.wantErr) {
				t.Fatalf("want %v, got %v", tt.wantErr, err)
			}
			if repo.sub.Status() != tt.wantStatus {
				t.Errorf("want %s, got %s", tt.wantStatus, repo.sub.Status())
			}
			if billing.calls != tt.wantCharges {
				t.Errorf("want %d billing calls, got %d", tt.wantCharges, billing.calls)
			}
		})
	}
}
