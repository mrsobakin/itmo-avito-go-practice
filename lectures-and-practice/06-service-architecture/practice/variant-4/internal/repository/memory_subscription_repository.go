package repository

import (
	"context"
	"sync"
	"time"

	"github.com/course-go-autumn-2026/practice6-variant4-subscription/internal/domain"
)

type subscriptionRecord struct {
	ID        int64
	AccountID int64
	Plan      string
	Status    string
	CreatedAt time.Time
}

type MemorySubscriptionRepository struct {
	mu     sync.Mutex
	rows   map[int64]subscriptionRecord
	nextID int64
}

func NewMemorySubscriptionRepository() *MemorySubscriptionRepository {
	return &MemorySubscriptionRepository{rows: map[int64]subscriptionRecord{}}
}

func (r *MemorySubscriptionRepository) Create(_ context.Context, s *domain.Subscription) (int64, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.nextID++
	rec := toRecord(s)
	rec.ID = r.nextID
	r.rows[rec.ID] = rec
	return rec.ID, nil
}

func (r *MemorySubscriptionRepository) Save(_ context.Context, s *domain.Subscription) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, ok := r.rows[s.ID()]; !ok {
		return domain.ErrNotFound
	}
	r.rows[s.ID()] = toRecord(s)
	return nil
}

func (r *MemorySubscriptionRepository) Get(_ context.Context, id int64) (*domain.Subscription, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	rec, ok := r.rows[id]
	if !ok {
		return nil, domain.ErrNotFound
	}
	return toDomain(rec), nil
}

func toRecord(s *domain.Subscription) subscriptionRecord {
	return subscriptionRecord{
		ID:        s.ID(),
		AccountID: s.AccountID(),
		Plan:      string(s.Plan()),
		Status:    string(s.Status()),
		CreatedAt: s.CreatedAt(),
	}
}

func toDomain(rec subscriptionRecord) *domain.Subscription {
	return domain.RestoreSubscription(rec.ID, rec.AccountID, domain.Plan(rec.Plan),
		domain.Status(rec.Status), rec.CreatedAt)
}
