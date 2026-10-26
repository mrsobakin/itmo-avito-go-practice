package handler

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/course-go-autumn-2026/practice6-variant4-subscription/internal/domain"
	"github.com/course-go-autumn-2026/practice6-variant4-subscription/internal/service"
)

type SubscriptionHandler struct {
	service *service.SubscriptionService
}

func NewSubscriptionHandler(s *service.SubscriptionService) *SubscriptionHandler {
	return &SubscriptionHandler{service: s}
}

func (h *SubscriptionHandler) Register(mux *http.ServeMux) {
	mux.HandleFunc("POST /subscriptions", h.Create)
	mux.HandleFunc("GET /subscriptions/{id}", h.Get)
	mux.HandleFunc("PATCH /subscriptions/{id}/plan", h.UpdatePlan)
}

type createSubscriptionRequest struct {
	AccountID int64  `json:"account_id"`
	Plan      string `json:"plan"`
}

type createSubscriptionResponse struct {
	ID     int64  `json:"id"`
	Status string `json:"status"`
}

type updatePlanRequest struct {
	Plan string `json:"plan"`
}

type subscriptionResponse struct {
	ID        int64
	AccountID int64
	Plan      string
	Status    string
	CreatedAt time.Time
}

func (h *SubscriptionHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req createSubscriptionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	sub, err := h.service.Subscribe(r.Context(), req.AccountID, req.Plan)
	if err != nil {
		writeError(w, err)
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(createSubscriptionResponse{ID: sub.ID(), Status: string(sub.Status())})
}

func (h *SubscriptionHandler) Get(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	sub, err := h.service.Get(r.Context(), id)
	if err != nil {
		writeError(w, err)
		return
	}
	writeSubscription(w, sub)
}

func (h *SubscriptionHandler) UpdatePlan(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	var req updatePlanRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	sub, err := h.service.ChangePlan(r.Context(), id, req.Plan)
	if err != nil {
		writeError(w, err)
		return
	}
	writeSubscription(w, sub)
}

func writeSubscription(w http.ResponseWriter, s *domain.Subscription) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(subscriptionResponse{
		ID:        s.ID(),
		AccountID: s.AccountID(),
		Plan:      string(s.Plan()),
		Status:    string(s.Status()),
		CreatedAt: s.CreatedAt(),
	})
}

func writeError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, domain.ErrUnknownPlan):
		http.Error(w, err.Error(), http.StatusBadRequest)
	case errors.Is(err, domain.ErrCardFraudulent):
		http.Error(w, err.Error(), http.StatusForbidden)
	case errors.Is(err, domain.ErrNotFound):
		w.WriteHeader(http.StatusNotFound)
	case errors.Is(err, domain.ErrSuspended), errors.Is(err, domain.ErrDowngradeWithDebt):
		http.Error(w, err.Error(), http.StatusConflict)
	default:
		log.Printf("internal error: %v", err)
		http.Error(w, "internal error", http.StatusInternalServerError)
	}
}
