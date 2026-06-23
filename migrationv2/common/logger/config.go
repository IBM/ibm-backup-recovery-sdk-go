/*******************************************************************************
 * IBM Confidential
 * OCO Source Materials
 * (C) Copyright IBM Corp. 2026 All Rights Reserved.
 * The source code for this program is not  published or otherwise divested of
 * its trade secrets, irrespective of what has been deposited with
 * the U.S. Copyright Office.
 ******************************************************************************/

package logger

import "io"

// Config holds logger configuration options.
// All fields are optional with sensible defaults.
type Config struct {
	// Level sets minimum log level: "debug", "info", "warn", "error".
	// Defaults to "info".
	Level string

	// Format specifies output format: "json" or "text".
	// Defaults to "text".
	Format string

	// ServiceName is included in every log entry as "service" field.
	ServiceName string

	// Environment is included in every log entry as "env" field.
	Environment string

	// Output specifies log destination (any io.Writer).
	// Defaults to os.Stdout.
	Output io.Writer

	// AddSource controls whether source location is included in logs.
	// Defaults to true.
	AddSource *bool

	// IncludeStackTrace enables stack traces for Error logs.
	// Warning: Expensive operation. Defaults to false.
	IncludeStackTrace bool
}

// DefaultConfig returns production-ready defaults.
// Level: "info", Format: "json", AddSource: true, IncludeStackTrace: false.
func DefaultConfig() Config {
	addSource := true
	return Config{
		Level:             "info",
		Format:            "json",
		AddSource:         &addSource,
		IncludeStackTrace: false,
	}
}
