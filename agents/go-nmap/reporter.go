package main

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"
)

// Reporter sends scan results back to the ZEPSEC Rails server.
type Reporter struct {
	serverURL string
	apiToken  string
	client    *http.Client
	logger    *log.Logger
}

// NewReporter creates a Reporter configured for the ZEPSEC server.
func NewReporter(cfg ServerConfig, logger *log.Logger) *Reporter {
	transport := &http.Transport{
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: !cfg.VerifyTLS,
		},
		MaxIdleConns:        10,
		IdleConnTimeout:     90 * time.Second,
		DisableCompression:  false,
	}

	return &Reporter{
		serverURL: cfg.URL,
		apiToken:  cfg.APIToken,
		client: &http.Client{
			Transport: transport,
			Timeout:   60 * time.Second,
		},
		logger: logger,
	}
}

// SendResults posts scan results to ZEPSEC's /api/v1/ra_api endpoint.
// Retries up to 3 times on transient failures with exponential backoff.
func (r *Reporter) SendResults(payload *ScanResultPayload) error {
	body, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("marshaling payload: %w", err)
	}

	url := fmt.Sprintf("%s/api/v1/ra_api", r.serverURL)

	var lastErr error
	for attempt := 0; attempt < 4; attempt++ {
		if attempt > 0 {
			backoff := time.Duration(1<<uint(attempt)) * time.Second
			r.logger.Printf("[REPORT] Retry %d after %s (jid=%s)", attempt, backoff, payload.JID)
			time.Sleep(backoff)
		}

		req, err := http.NewRequest(http.MethodPost, url, bytes.NewReader(body))
		if err != nil {
			return fmt.Errorf("creating request: %w", err)
		}

		// ZEPSEC uses Rails' authenticate_with_http_token which expects:
		//   Authorization: Token token="<value>"
		req.Header.Set("Authorization", fmt.Sprintf("Token token=\"%s\"", r.apiToken))
		req.Header.Set("Content-Type", "application/json")

		resp, err := r.client.Do(req)
		if err != nil {
			lastErr = fmt.Errorf("HTTP request failed: %w", err)
			r.logger.Printf("[REPORT] Request error (jid=%s, attempt=%d): %v", payload.JID, attempt, err)
			continue
		}

		respBody, _ := io.ReadAll(resp.Body)
		resp.Body.Close()

		if resp.StatusCode == http.StatusOK {
			r.logger.Printf("[REPORT] Results accepted by server (jid=%s, hosts=%d)",
				payload.JID, len(payload.Hosts))
			return nil
		}

		lastErr = fmt.Errorf("server returned %d: %s", resp.StatusCode, string(respBody))
		r.logger.Printf("[REPORT] Server rejected results (jid=%s, status=%d): %s",
			payload.JID, resp.StatusCode, string(respBody))

		// Don't retry on client errors (4xx) — only on server errors (5xx)
		if resp.StatusCode >= 400 && resp.StatusCode < 500 {
			return lastErr
		}
	}

	return fmt.Errorf("all retries exhausted: %w", lastErr)
}
