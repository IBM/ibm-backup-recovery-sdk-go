/*******************************************************************************
 * IBM Confidential
 * OCO Source Materials
 * (C) Copyright IBM Corp. 2026 All Rights Reserved.
 * The source code for this program is not  published or otherwise divested of
 * its trade secrets, irrespective of what has been deposited with
 * the U.S. Copyright Office.
 ******************************************************************************/

// Package logger provides structured logging for the IBM BRS Workload Migration SDK.
//
// Basic usage:
//
//	log := logger.New(logger.Config{Level: "info", Format: "json"})
//	log.Info(ctx, "message", "key", "value")
//
// Context fields (transaction_id, account_id, etc.) are automatically extracted.
package logger

import (
	"context"
)

// Logger defines the structured logging interface.
// All methods require context for request tracking.
type Logger interface {
	// Debug logs a message at DEBUG level.
	Debug(ctx context.Context, msg string, args ...any)

	// Info logs a message at INFO level.
	Info(ctx context.Context, msg string, args ...any)

	// Warn logs a message at WARN level.
	Warn(ctx context.Context, msg string, args ...any)

	// Error logs a message at ERROR level.
	// Use "error", err or logger.Err(err) to include an error.
	Error(ctx context.Context, msg string, args ...any)

	// With returns a new Logger with permanent key-value pairs.
	With(args ...any) Logger
}

// Err returns key-value pairs for logging an error.
func Err(err error) []any {
	if err == nil {
		return nil
	}
	return []any{"error", err.Error()}
}

// NoOpLogger is a no-op logger implementation.
type NoOpLogger struct{}

// NewNoOpLogger creates a no-op logger.
func NewNoOpLogger() Logger {
	return &NoOpLogger{}
}

func (n *NoOpLogger) Debug(ctx context.Context, msg string, args ...any) {}
func (n *NoOpLogger) Info(ctx context.Context, msg string, args ...any)  {}
func (n *NoOpLogger) Warn(ctx context.Context, msg string, args ...any)  {}
func (n *NoOpLogger) Error(ctx context.Context, msg string, args ...any) {}
func (n *NoOpLogger) With(args ...any) Logger                            { return n }
