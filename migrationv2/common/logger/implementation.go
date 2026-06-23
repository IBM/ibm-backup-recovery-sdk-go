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
	"context"
	"log/slog"
	"runtime"
	"runtime/debug"
	"time"

	commoncontext "github.com/IBM/ibm-backup-recovery-sdk-go/migrationv2/common/context"
)

// callerPC returns the program counter of the caller.
func callerPC(skip int) uintptr {
	var pcs [1]uintptr
	runtime.Callers(skip+2, pcs[:])
	return pcs[0]
}

// log handles the actual logging with context field extraction.
func (l *SLogger) log(ctx context.Context, level slog.Level, msg string, args []any) {
	if !l.logger.Enabled(ctx, level) {
		return
	}

	// Extract and append context fields
	contextArgs := commoncontext.ExtractContextFields(ctx)
	if len(contextArgs) > 0 {
		args = append(args, contextArgs...)
	}

	pc := callerPC(3)

	r := slog.NewRecord(time.Now(), level, msg, pc)
	r.Add(args...)
	_ = l.logger.Handler().Handle(ctx, r)
}

func (l *SLogger) Debug(ctx context.Context, msg string, args ...any) {
	l.log(ctx, slog.LevelDebug, msg, args)
}

func (l *SLogger) Info(ctx context.Context, msg string, args ...any) {
	l.log(ctx, slog.LevelInfo, msg, args)
}

func (l *SLogger) Warn(ctx context.Context, msg string, args ...any) {
	l.log(ctx, slog.LevelWarn, msg, args)
}

func (l *SLogger) Error(ctx context.Context, msg string, args ...any) {
	if l.includeStackTrace {
		args = append(args, "stack", string(debug.Stack()))
	}
	l.log(ctx, slog.LevelError, msg, args)
}

// With returns a new Logger with permanent key-value pairs.
func (l *SLogger) With(args ...any) Logger {
	return &SLogger{
		logger:            l.logger.With(args...),
		levelVar:          l.levelVar,
		includeStackTrace: l.includeStackTrace,
	}
}

// SetLevel dynamically changes the log level.
// Valid levels: "debug", "info", "warn", "error".
// Requires type assertion: log.(*logger.SLogger).SetLevel("debug")
func (l *SLogger) SetLevel(level string) {
	l.levelVar.Set(parseLevel(level))
}
