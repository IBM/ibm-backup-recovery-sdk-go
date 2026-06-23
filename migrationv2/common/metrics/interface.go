/*******************************************************************************
 * IBM Confidential
 * OCO Source Materials
 * (C) Copyright IBM Corp. 2026 All Rights Reserved.
 * The source code for this program is not  published or otherwise divested of
 * its trade secrets, irrespective of what has been deposited with
 * the U.S. Copyright Office.
 ******************************************************************************/

// Package metrics provides Prometheus metrics for the IBM BRS Workload Migration SDK.
package metrics

import (
	"context"
	"net/http"
	"time"
)

// Metrics defines a generic interface for recording metrics.
// This interface is stable and allows for custom implementations.
// All methods accept context.Context for future tracing/correlation support.
type Metrics interface {
	// Counter operations
	IncCounter(ctx context.Context, name string, labels ...Label)
	AddCounter(ctx context.Context, name string, value float64, labels ...Label)

	// Gauge operations
	SetGauge(ctx context.Context, name string, value float64, labels ...Label)
	IncGauge(ctx context.Context, name string, labels ...Label)
	DecGauge(ctx context.Context, name string, labels ...Label)

	// Histogram operations
	RecordDuration(ctx context.Context, name string, duration time.Duration, labels ...Label)

	// HTTP handler for exposing metrics
	Handler() http.Handler
}

// Label represents a key-value pair for metric labeling.
type Label struct {
	Key   string
	Value string
}
