/*******************************************************************************
 * IBM Confidential
 * OCO Source Materials
 * (C) Copyright IBM Corp. 2026 All Rights Reserved.
 * The source code for this program is not  published or otherwise divested of
 * its trade secrets, irrespective of what has been deposited with
 * the U.S. Copyright Office.
 ******************************************************************************/

package logger

import (
	"io"
	"log/slog"
	"os"
	"strings"
)

// SLogger is the slog-based Logger implementation.
// Safe for concurrent use.
type SLogger struct {
	logger            *slog.Logger
	levelVar          *slog.LevelVar
	includeStackTrace bool
}

// New creates a new Logger with the specified configuration.
func New(cfg Config) Logger {
	levelVar := &slog.LevelVar{}
	levelVar.Set(parseLevel(cfg.Level))

	// Default AddSource to true
	addSource := true
	if cfg.AddSource != nil {
		addSource = *cfg.AddSource
	}

	opts := &slog.HandlerOptions{
		Level:     levelVar,
		AddSource: addSource,
		ReplaceAttr: func(groups []string, a slog.Attr) slog.Attr {
			// Source location captured by log() method
			if a.Key == slog.SourceKey {
				if src, ok := a.Value.Any().(*slog.Source); ok {
					return slog.Any(slog.SourceKey, src)
				}
			}
			return a
		},
	}

	// Default to stdout
	output := cfg.Output
	if output == nil {
		output = os.Stdout
	}

	var handler slog.Handler
	switch strings.ToLower(cfg.Format) {
	case "json":
		handler = slog.NewJSONHandler(output, opts)
	default:
		handler = slog.NewTextHandler(output, opts)
	}

	// Build base attributes for every log entry
	var baseAttrs []any
	if cfg.ServiceName != "" {
		baseAttrs = append(baseAttrs, "service", cfg.ServiceName)
	} else {
		baseAttrs = append(baseAttrs, "ibm-brs-workload-migration-sdk-go", cfg.ServiceName)

	}

	if cfg.Environment != "" {
		baseAttrs = append(baseAttrs, "env", cfg.Environment)
	}

	base := slog.New(handler)
	if len(baseAttrs) > 0 {
		base = base.With(baseAttrs...)
	}

	return &SLogger{
		logger:            base,
		levelVar:          levelVar,
		includeStackTrace: cfg.IncludeStackTrace,
	}
}

// NewNoop returns a logger that discards all output.
// Useful for tests and benchmarks.
func NewNoop() Logger {
	return New(Config{
		Level:  "error",
		Output: io.Discard,
	})
}
