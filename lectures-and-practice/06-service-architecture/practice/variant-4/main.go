package main

import (
	"log"
	"net/http"

	"github.com/course-go-autumn-2026/practice6-variant4-subscription/internal/handler"
	"github.com/course-go-autumn-2026/practice6-variant4-subscription/internal/repository"
	"github.com/course-go-autumn-2026/practice6-variant4-subscription/internal/service"
)

const (
	billingURL   = "http://billing-provider.internal/charge"
	cardFraudURL = "http://card-fraud.internal/check"
)

func main() {
	app := newApp(repository.NewMemorySubscriptionRepository(), cardFraudURL, billingURL)

	log.Println("listening on :8080")
	log.Fatal(http.ListenAndServe(":8080", app))
}

func newApp(subscriptions service.SubscriptionRepository, fraudURL, billingURL string) http.Handler {
	svc := service.NewSubscriptionService(
		subscriptions,
		repository.NewMemoryAccountRepository(),
		repository.NewHTTPCardFraudChecker(fraudURL, http.DefaultClient),
		repository.NewHTTPBillingProvider(billingURL, http.DefaultClient),
	)

	mux := http.NewServeMux()
	handler.NewSubscriptionHandler(svc).Register(mux)
	return mux
}
