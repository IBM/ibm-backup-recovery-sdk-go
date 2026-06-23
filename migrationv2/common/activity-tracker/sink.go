/*******************************************************************************
 * IBM Confidential
 * OCO Source Materials
 * (C) Copyright IBM Corp. 2026 All Rights Reserved.
 * The source code for this program is not  published or otherwise divested of
 * its trade secrets, irrespective of what has been deposited with
 * the U.S. Copyright Office.
 ******************************************************************************/

package activity_tracker

import "context"

// Sink represents a destination for SDK activity events.
//
// Implementations should keep [`Sink.Emit()`](ibm-brs-sdk-go/ibm-backup-recovery-sdk-go/migrationv2/common/activity-tracker/sink.go:11)
// lightweight so core SDK operations are not slowed down by telemetry delivery.
type Sink interface {
	// Emit sends one activity event.
	Emit(ctx context.Context, event Event) error

	// Close releases any resources held by the sink.
	Close(ctx context.Context) error
}

// NoOpSink is a sink implementation that discards all events.
type NoOpSink struct{}

// Emit discards the supplied event and always succeeds.
func (NoOpSink) Emit(ctx context.Context, event Event) error {
	return nil
}

// Close performs no action and always succeeds.
func (NoOpSink) Close(ctx context.Context) error {
	return nil
}
