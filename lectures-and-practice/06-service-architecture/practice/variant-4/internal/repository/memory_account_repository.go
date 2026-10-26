package repository

import (
	"context"

	"github.com/course-go-autumn-2026/practice6-variant4-subscription/internal/domain"
)

type accountRecord struct {
	ID      int64
	HasDebt bool
}

type MemoryAccountRepository struct {
	rows map[int64]accountRecord
}

func NewMemoryAccountRepository() *MemoryAccountRepository {
	return &MemoryAccountRepository{rows: map[int64]accountRecord{
		1: {ID: 1, HasDebt: false},
		2: {ID: 2, HasDebt: true},
	}}
}

func (r *MemoryAccountRepository) Get(_ context.Context, id int64) (domain.Account, error) {
	rec, ok := r.rows[id]
	if !ok {
		return domain.Account{ID: id}, nil
	}
	return domain.Account{ID: rec.ID, HasDebt: rec.HasDebt}, nil
}
