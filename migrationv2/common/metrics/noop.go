/*******************************************************************************
 * IBM Confidential
 * OCO Source Materials
 * (C) Copyright IBM Corp. 2026 All Rights Reserved.
 * The source code for this program is not  published or otherwise divested of
 * its trade secrets, irrespective of what has been deposited with
 * the U.S. Copyright Office.
 ******************************************************************************/
package metrics

import (
	"context"
	"net/http"
	"time"
)

// NoopMetrics is a metrics implementation that discards all metrics.
// Useful for testing or when metrics collection is not needed.
type NoopMetrics struct{}

// NewNoop creates a new no-op metrics implementation.
func NewNoop() *NoopMetrics {
	return &NoopMetrics{}
}

func (n *NoopMetrics) IncCounter(_ context.Context, _ string, _ ...Label)                      {}
func (n *NoopMetrics) AddCounter(_ context.Context, _ string, _ float64, _ ...Label)           {}
func (n *NoopMetrics) SetGauge(_ context.Context, _ string, _ float64, _ ...Label)             {}
func (n *NoopMetrics) IncGauge(_ context.Context, _ string, _ ...Label)                        {}
func (n *NoopMetrics) DecGauge(_ context.Context, _ string, _ ...Label)                        {}
func (n *NoopMetrics) RecordDuration(_ context.Context, _ string, _ time.Duration, _ ...Label) {}
func (n *NoopMetrics) Handler() http.Handler                                                   { return nil }
