/*******************************************************************************
 * IBM Confidential
 * OCO Source Materials
 * (C) Copyright IBM Corp. 2026 All Rights Reserved.
 * The source code for this program is not  published or otherwise divested of
 * its trade secrets, irrespective of what has been deposited with
 * the U.S. Copyright Office.
 ******************************************************************************/

package activity_tracker

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/IBM/go-sdk-core/v5/core"
	commoncontext "github.com/IBM/ibm-backup-recovery-sdk-go/migrationv2/common/context"
)

const (
	defaultIngestionPath = "/logs/v1/singles"
	defaultAuthHeader    = "Authorization"
	defaultHTTPTimeout   = 15 * time.Second
)

// HTTPSink emits activity events to an IBM Cloud Logs ingestion endpoint.
type HTTPSink struct {
	client         *http.Client
	url            string
	authHeaderName string
	iamAuth        *core.IamAuthenticator
}

// HTTPSinkConfig configures an [`activity_tracker.HTTPSink`](ibm-brs-sdk-go/ibm-backup-recovery-sdk-go/migrationv2/common/activity-tracker/http_sink.go:23).
type HTTPSinkConfig struct {
	// IngestionEndpoint is the IBM Cloud Logs ingestion endpoint.
	// The caller may provide either the base instance endpoint or the full singles ingestion path.
	IngestionEndpoint string `json:"ingestion_endpoint"`

	// Timeout is the timeout for HTTP calls to the ingestion endpoint.
	Timeout time.Duration `json:"timeout,omitempty"`

	// AuthHeaderName is typically "Authorization" and is only used when IAM authentication is enabled.
	AuthHeaderName string `json:"auth_header_name,omitempty"`

	// IAMAuthenticator allows advanced callers to provide a preconfigured IAM authenticator.
	// If provided, it takes precedence over IBMCloudAPIKey.
	IAMAuthenticator *core.IamAuthenticator
}

// NewHTTPSink validates the configuration and returns an [`activity_tracker.HTTPSink`](ibm-brs-sdk-go/ibm-backup-recovery-sdk-go/migrationv2/common/activity-tracker/http_sink.go:23).
func NewHTTPSink(cfg HTTPSinkConfig) (*HTTPSink, error) {
	if cfg.Timeout <= 0 {
		cfg.Timeout = defaultHTTPTimeout
	}
	if strings.TrimSpace(cfg.AuthHeaderName) == "" {
		cfg.AuthHeaderName = defaultAuthHeader
	}

	auth := cfg.IAMAuthenticator

	ingestionURL, err := normalizeIngestionEndpoint(cfg.IngestionEndpoint)
	if err != nil {
		return nil, err
	}

	return &HTTPSink{
		client:         &http.Client{Timeout: cfg.Timeout},
		url:            ingestionURL,
		authHeaderName: cfg.AuthHeaderName,
		iamAuth:        auth,
	}, nil
}

// Emit sends one event to the IBM Cloud Logs ingestion endpoint.
func (s *HTTPSink) Emit(ctx context.Context, event Event) error {
	// Populate TransactionID from context if not already set
	if event.TransactionID == "" {
		event.TransactionID = commoncontext.TransactionIDFromContext(ctx)
	}

	// Populate AccountID from context if not already set
	if event.AccountID == "" {
		if accountID := commoncontext.AccountIDFromContext(ctx); accountID != "" {
			event.AccountID = accountID
		}
	}

	payload, err := Marshal(event)
	if err != nil {
		return err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, s.url, bytes.NewReader(payload))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")

	if s.iamAuth != nil {
		tok, err := s.iamAuth.GetToken()
		if err != nil {
			return fmt.Errorf("failed to get IAM token for logs ingestion: %w", err)
		}
		req.Header.Set(s.authHeaderName, "Bearer "+tok)
	}

	resp, err := s.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 2048))
		if len(body) > 0 {
			return fmt.Errorf("http sink: status %d: %s", resp.StatusCode, strings.TrimSpace(string(body)))
		}
		return fmt.Errorf("http sink: status %d", resp.StatusCode)
	}

	return nil
}

// Close releases sink resources. [`activity_tracker.HTTPSink`](ibm-brs-sdk-go/ibm-backup-recovery-sdk-go/migrationv2/common/activity-tracker/http_sink.go:23) does not hold closable resources.
func (s *HTTPSink) Close(ctx context.Context) error {
	return nil
}

// Marshal converts an [`activity_tracker.Event`](ibm-brs-sdk-go/ibm-backup-recovery-sdk-go/migrationv2/common/activity-tracker/event.go:22) into IBM Cloud Logs ingestion JSON.
func Marshal(e Event) ([]byte, error) {
	ts := e.Timestamp
	if ts.IsZero() {
		ts = time.Now()
	}

	requestPayload := map[string]any{
		"accountId":  e.AccountID,
		"requestId":  e.RequestID,
		"requestURI": e.RequestURI,
	}

	eventPayload := map[string]any{
		"requestPayload": requestPayload,
		"API":            e.API,
		"step":           e.Substep,
		"status":         string(e.State),
		"transactionId":  e.TransactionID,
		"role":           e.Role,
		"attempt":        e.Attempt,
		"durationMs":     e.Duration.Milliseconds(),
		"message":        e.Message,
		"details":        e.Details,
		"error":          errString(e.Err),
		"serviceName":    "backup-recovery",
	}

	jsonBytes, err := json.Marshal(eventPayload)
	if err != nil {
		return nil, err
	}

	payload := []map[string]any{
		{
			"applicationName": "migration-sdk",
			"subsystemName":   "activity-tracking",
			"timestamp":       ts.UTC().UnixMilli(),
			"severity":        severityFromState(e.State),
			"text":            string(jsonBytes),
		},
	}

	return json.Marshal(payload)
}

func normalizeIngestionEndpoint(raw string) (string, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return "", fmt.Errorf("ingestion endpoint is required")
	}

	parsedURL, err := url.Parse(raw)
	if err != nil {
		return "", fmt.Errorf("invalid ingestion endpoint: %w", err)
	}
	if parsedURL.Scheme == "" || parsedURL.Host == "" {
		return "", fmt.Errorf("ingestion endpoint must include scheme and host")
	}

	parsedURL.Path = strings.TrimRight(parsedURL.Path, "/")
	if parsedURL.Path == "" {
		parsedURL.Path = defaultIngestionPath
	} else if !strings.Contains(parsedURL.Path, "/logs/v1/") {
		parsedURL.Path += defaultIngestionPath
	}

	return parsedURL.String(), nil
}

func severityFromState(state EventState) int {
	switch state {
	case EventStateStarted, EventStateSucceeded, EventStateInProgress:
		return 4
	case EventStateRetrying:
		return 3
	case EventStateFailed:
		return 2
	default:
		return 4
	}
}

func errString(err error) any {
	if err == nil {
		return nil
	}
	return err.Error()
}
