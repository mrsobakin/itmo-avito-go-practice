package repository

import (
	"context"
	"fmt"
	"net/http"
)

type HTTPBillingProvider struct {
	url    string
	client *http.Client
}

func NewHTTPBillingProvider(url string, client *http.Client) *HTTPBillingProvider {
	return &HTTPBillingProvider{url: url, client: client}
}

func (p *HTTPBillingProvider) Charge(ctx context.Context) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, p.url, nil)
	if err != nil {
		return err
	}

	resp, err := p.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("billing declined: status %d", resp.StatusCode)
	}
	return nil
}
