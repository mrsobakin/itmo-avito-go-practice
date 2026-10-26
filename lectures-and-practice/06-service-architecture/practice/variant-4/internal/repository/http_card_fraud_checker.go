package repository

import (
	"context"
	"net/http"
)

type HTTPCardFraudChecker struct {
	url    string
	client *http.Client
}

func NewHTTPCardFraudChecker(url string, client *http.Client) *HTTPCardFraudChecker {
	return &HTTPCardFraudChecker{url: url, client: client}
}

func (c *HTTPCardFraudChecker) IsBlocked(ctx context.Context) (bool, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.url, nil)
	if err != nil {
		return false, err
	}

	resp, err := c.client.Do(req)
	if err != nil {
		return false, err
	}
	defer resp.Body.Close()

	return resp.StatusCode == http.StatusForbidden, nil
}
