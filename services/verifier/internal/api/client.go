package api

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

// Submission represents a pending submission from the list endpoint.
type Submission struct {
	SubmissionID string `json:"submission_id"`
	PostURL      string `json:"post_url"`
	Platform     string `json:"platform"`
}

// Snapshot represents a metric snapshot to record.
type Snapshot struct {
	SubmissionID string `json:"submission_id"`
	Platform     string `json:"platform"`
	Views        int64  `json:"views"`
	Likes        int64  `json:"likes"`
	Comments     int64  `json:"comments"`
	Shares       int64  `json:"shares"`
	CapturedAt   string `json:"captured_at"`
}

// Client is an HTTP client for the main API's internal verification endpoints.
type Client struct {
	BaseURL    string
	APIKey     string
	HTTPClient *http.Client
}

// NewClient creates a client with the given base URL, API key, and timeout.
func NewClient(baseURL, apiKey string, timeout time.Duration) *Client {
	return &Client{
		BaseURL: strings.TrimRight(baseURL, "/"),
		APIKey:  apiKey,
		HTTPClient: &http.Client{
			Timeout: timeout,
		},
	}
}

type listResponse struct {
	Submissions []Submission `json:"submissions"`
}

// ListPending fetches submissions that need verification.
func (c *Client) ListPending(ctx context.Context, limit int) ([]Submission, error) {
	body, err := json.Marshal(map[string]int{"limit": limit})
	if err != nil {
		return nil, fmt.Errorf("marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.BaseURL+"/internal/snapshots/list", bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("build request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	if c.APIKey != "" {
		req.Header.Set("X-Verifier-Key", c.APIKey)
	}

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(io.LimitReader(resp.Body, 512))
		return nil, fmt.Errorf("API returned %d: %s", resp.StatusCode, string(respBody))
	}

	var result listResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("decode response: %w", err)
	}
	return result.Submissions, nil
}

// RecordSnapshot sends a metric snapshot to the main API.
func (c *Client) RecordSnapshot(ctx context.Context, snap Snapshot) error {
	body, err := json.Marshal(snap)
	if err != nil {
		return fmt.Errorf("marshal snapshot: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.BaseURL+"/internal/snapshots", bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("build request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	if c.APIKey != "" {
		req.Header.Set("X-Verifier-Key", c.APIKey)
	}

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(io.LimitReader(resp.Body, 512))
		return fmt.Errorf("API returned %d: %s", resp.StatusCode, string(respBody))
	}
	return nil
}
