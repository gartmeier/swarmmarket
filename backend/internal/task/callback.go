package task

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

// HTTPCallbackDeliverer delivers task callbacks via HTTP webhooks.
type HTTPCallbackDeliverer struct {
	client *http.Client
}

// NewHTTPCallbackDeliverer creates a new HTTP callback deliverer.
func NewHTTPCallbackDeliverer() *HTTPCallbackDeliverer {
	return &HTTPCallbackDeliverer{
		client: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

// DeliverCallback sends a callback to the specified URL with HMAC signature.
func (d *HTTPCallbackDeliverer) DeliverCallback(ctx context.Context, url, secret string, payload TaskCallback) error {
	if url == "" {
		return nil
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal callback payload: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("failed to create callback request: %w", err)
	}

	// Set headers
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", "SwarmMarket-Callback/1.0")
	req.Header.Set("X-SwarmMarket-Event", "task."+string(payload.Status))
	req.Header.Set("X-SwarmMarket-Task-ID", payload.TaskID.String())
	req.Header.Set("X-SwarmMarket-Timestamp", fmt.Sprintf("%d", payload.Timestamp.Unix()))

	// Add HMAC signature if secret is provided
	if secret != "" {
		mac := hmac.New(sha256.New, []byte(secret))
		mac.Write(body)
		signature := "sha256=" + hex.EncodeToString(mac.Sum(nil))
		req.Header.Set("X-SwarmMarket-Signature", signature)
	}

	resp, err := d.client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to deliver callback: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return fmt.Errorf("callback returned status %d", resp.StatusCode)
	}

	return nil
}

// AsyncCallbackDeliverer wraps HTTPCallbackDeliverer for async delivery with retries.
type AsyncCallbackDeliverer struct {
	deliverer  *HTTPCallbackDeliverer
	maxRetries int
	retryDelay time.Duration
}

// NewAsyncCallbackDeliverer creates a new async callback deliverer.
func NewAsyncCallbackDeliverer() *AsyncCallbackDeliverer {
	return &AsyncCallbackDeliverer{
		deliverer:  NewHTTPCallbackDeliverer(),
		maxRetries: 3,
		retryDelay: 5 * time.Second,
	}
}

// DeliverCallback delivers a callback asynchronously with retries.
func (d *AsyncCallbackDeliverer) DeliverCallback(ctx context.Context, url, secret string, payload TaskCallback) error {
	// Run delivery in background
	go func() {
		for attempt := 0; attempt <= d.maxRetries; attempt++ {
			if attempt > 0 {
				time.Sleep(d.retryDelay * time.Duration(attempt))
			}

			// Use background context since original may be cancelled
			deliveryCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
			err := d.deliverer.DeliverCallback(deliveryCtx, url, secret, payload)
			cancel()

			if err == nil {
				return
			}

			// Log error but continue retrying
			// In production, this would use a proper logger
			if attempt == d.maxRetries {
				// Final attempt failed - in production, would log or queue for later
				return
			}
		}
	}()

	return nil
}
