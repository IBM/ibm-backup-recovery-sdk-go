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
	"fmt"
	"strconv"
	"time"
)

// EventState represents the lifecycle state of an activity tracker event.
type EventState string

const (
	// EventStateStarted indicates that an operation or step has started.
	EventStateStarted EventState = "started"
	// EventStateSucceeded indicates that an operation or step completed successfully.
	EventStateSucceeded EventState = "succeeded"
	// EventStateFailed indicates that an operation or step failed.
	EventStateFailed EventState = "failed"
	// EventStateRetrying indicates that an operation or step is being retried.
	EventStateRetrying EventState = "retrying"
	// EventStateInProgress indicates that an operation or step is still in progress.
	EventStateInProgress EventState = "in_progress"
)

// Event represents one structured activity emitted by the SDK.
type Event struct {
	// Timestamp is when this activity occurred.
	// If zero, the sink or marshaler may default it to the current time.
	Timestamp time.Time `json:"timestamp"`

	// TransactionID correlates multiple SDK operations belonging to the same workflow.
	TransactionID string `json:"transaction_id,omitempty"`

	// API is the high-level phase of the SDK operation API.
	API string `json:"api,omitempty"`

	// Substep is the fine-grained step within a workflow.
	Substep string `json:"substep,omitempty"`

	// Role describes the actor or component performing the action, if relevant.
	Role string `json:"role,omitempty"`

	// State is the outcome or status for this event.
	State EventState `json:"state,omitempty"`

	// Attempt is the retry attempt number for this step.
	Attempt int `json:"attempt,omitempty"`

	// Duration is how long the step took.
	Duration time.Duration `json:"duration,omitempty"`

	// Details contains structured metadata for debugging and querying.
	Details map[string]string `json:"details,omitempty"`

	// Message is a short human-readable description of the activity.
	Message string `json:"message,omitempty"`

	// Err is the underlying error for failed steps.
	Err error `json:"error,omitempty"`

	// AccountID is the IBM Cloud account identifier, if available.
	AccountID string `json:"account_id,omitempty"`

	// RequestID is the service request identifier returned by IBM service calls, if available.
	RequestID string `json:"request_id,omitempty"`

	// RequestURI is the request endpoint or URI used for the call, if available.
	RequestURI string `json:"request_uri,omitempty"`

	// SourceIP is the IP address of the source/client initiating the request, if available.
	SourceIP string `json:"source_ip,omitempty"`
}

type TaskActivityEventInput struct {
	Substep   string            `json:"substep,omitempty"`
	Message   string            `json:"message,omitempty"`
	State     EventState        `json:"state,omitempty"`
	Err       error             `json:"error,omitempty"`
	Attempt   int               `json:"attempt,omitempty"`
	StartedAt time.Time         `json:"started_at,omitempty"`
	Details   map[string]string `json:"details,omitempty"`
	SourceIP  string            `json:"source_ip,omitempty"`
}

// SetCommonFields sets the common fields that are repeated across all event builders
func (input *TaskActivityEventInput) SetCommonFields(state EventState, message string, err error, attempt int, startedAt time.Time, sourceIP string) {
	input.State = state
	input.Message = message
	input.Err = err
	input.Attempt = attempt
	input.StartedAt = startedAt
	input.SourceIP = sourceIP
}

func BuildTaskActivityEvent(input TaskActivityEventInput) Event {
	event := Event{
		API:      "task_api",
		Substep:  input.Substep,
		State:    input.State,
		Attempt:  input.Attempt,
		Message:  input.Message,
		Details:  cloneDetails(input.Details),
		Err:      input.Err,
		SourceIP: input.SourceIP,
	}

	if !input.StartedAt.IsZero() {
		event.Duration = time.Since(input.StartedAt)
	}

	return event
}

func BuildPolicyActivityEvent(policyName string, state EventState, message string, err error, attempt int, startedAt time.Time, sourceIP string) Event {
	input := TaskActivityEventInput{
		Substep: "create_policy",
		Details: map[string]string{
			"policyName": policyName,
		},
	}
	input.SetCommonFields(state, message, err, attempt, startedAt, sourceIP)
	return BuildTaskActivityEvent(input)
}

func BuildConnectionActivityEvent(connectionName string, connectionID string, state EventState, message string, err error, attempt int, startedAt time.Time, sourceIP string) Event {
	input := TaskActivityEventInput{
		Substep: "create_connection",
		Details: map[string]string{
			"connectionName": connectionName,
			"connectionId":   connectionID,
		},
	}
	input.SetCommonFields(state, message, err, attempt, startedAt, sourceIP)
	return BuildTaskActivityEvent(input)
}

func BuildDeleteConnectionActivityEvent(connectionID string, state EventState, message string, err error, attempt int, startedAt time.Time, sourceIP string) Event {
	input := TaskActivityEventInput{
		Substep: "delete_connection",
		Details: map[string]string{
			"connectionId": connectionID,
		},
	}
	input.SetCommonFields(state, message, err, attempt, startedAt, sourceIP)
	return BuildTaskActivityEvent(input)
}

func BuildConnectorActivityEvent(connectionID string, connectorID string, state EventState, message string, err error, attempt int, startedAt time.Time, sourceIP string) Event {
	input := TaskActivityEventInput{
		Substep: "create_connector",
		Details: map[string]string{
			"connectionId": connectionID,
			"connectorId":  connectorID,
		},
	}
	input.SetCommonFields(state, message, err, attempt, startedAt, sourceIP)
	return BuildTaskActivityEvent(input)
}

func BuildDeleteConnectorActivityEvent(connectorID string, state EventState, message string, err error, attempt int, startedAt time.Time, sourceIP string) Event {
	input := TaskActivityEventInput{
		Substep: "delete_connector",
		Details: map[string]string{
			"connectorId": connectorID,
		},
	}
	input.SetCommonFields(state, message, err, attempt, startedAt, sourceIP)
	return BuildTaskActivityEvent(input)
}

func BuildRegistrationActivityEvent(connectionID string, registrationID int64, sourceName string, state EventState, message string, err error, attempt int, startedAt time.Time, sourceIP string) Event {
	input := TaskActivityEventInput{
		Substep: "register_source",
		Details: map[string]string{
			"connectionId":   connectionID,
			"registrationId": strconv.Itoa(int(registrationID)),
			"sourceName":     sourceName,
		},
	}
	input.SetCommonFields(state, message, err, attempt, startedAt, sourceIP)
	return BuildTaskActivityEvent(input)
}

func BuildUnregisterSourceActivityEvent(registrationID int64, state EventState, message string, err error, attempt int, startedAt time.Time, sourceIP string) Event {
	input := TaskActivityEventInput{
		Substep: "unregister_source",
		Details: map[string]string{
			"registrationId": strconv.Itoa(int(registrationID)),
		},
	}
	input.SetCommonFields(state, message, err, attempt, startedAt, sourceIP)
	return BuildTaskActivityEvent(input)
}

func BuildProtectionGroupActivityEvent(groupName string, groupID string, registrationID int64, policyID string, state EventState, message string, err error, attempt int, startedAt time.Time, sourceIP string) Event {
	input := TaskActivityEventInput{
		Substep: "create_protection_group",
		Details: map[string]string{
			"groupName":      groupName,
			"groupId":        groupID,
			"registrationId": strconv.Itoa(int(registrationID)),
			"policyId":       policyID,
		},
	}
	input.SetCommonFields(state, message, err, attempt, startedAt, sourceIP)
	return BuildTaskActivityEvent(input)
}

func BuildDeleteProtectionGroupActivityEvent(groupID string, state EventState, message string, err error, attempt int, startedAt time.Time, sourceIP string) Event {
	input := TaskActivityEventInput{
		Substep: "delete_protection_group",
		Details: map[string]string{
			"groupId": groupID,
		},
	}
	input.SetCommonFields(state, message, err, attempt, startedAt, sourceIP)
	return BuildTaskActivityEvent(input)
}

func BuildDeletePolicyActivityEvent(policyID string, state EventState, message string, err error, attempt int, startedAt time.Time, sourceIP string) Event {
	input := TaskActivityEventInput{
		Substep: "delete_policy",
		Details: map[string]string{
			"policyId": policyID,
		},
	}
	input.SetCommonFields(state, message, err, attempt, startedAt, sourceIP)
	return BuildTaskActivityEvent(input)
}

func BuildBackupActivityEvent(groupID string, backupID string, backupType string, state EventState, message string, err error, attempt int, startedAt time.Time, sourceIP string) Event {
	input := TaskActivityEventInput{
		Substep: "create_backup",
		Details: map[string]string{
			"groupId":    groupID,
			"backupId":   backupID,
			"backupType": backupType,
		},
	}
	input.SetCommonFields(state, message, err, attempt, startedAt, sourceIP)
	return BuildTaskActivityEvent(input)
}

func BuildRestoreActivityEvent(groupID string, backupID string, restoreID string, registrationID int64, state EventState, message string, err error, attempt int, startedAt time.Time, sourceIP string) Event {
	input := TaskActivityEventInput{
		Substep: "create_restore",
		Details: map[string]string{
			"groupId":        groupID,
			"backupId":       backupID,
			"restoreId":      restoreID,
			"registrationId": strconv.Itoa(int(registrationID)),
		},
	}
	input.SetCommonFields(state, message, err, attempt, startedAt, sourceIP)
	return BuildTaskActivityEvent(input)
}

func BuildAbortRestoreActivityEvent(restoreID string, force bool, state EventState, message string, err error, attempt int, startedAt time.Time, sourceIP string) Event {
	input := TaskActivityEventInput{
		Substep: "abort_restore",
		Details: map[string]string{
			"restoreId": restoreID,
			"force":     fmt.Sprintf("%t", force),
		},
	}
	input.SetCommonFields(state, message, err, attempt, startedAt, sourceIP)
	return BuildTaskActivityEvent(input)
}

func cloneDetails(details map[string]string) map[string]string {
	if len(details) == 0 {
		return map[string]string{}
	}

	cloned := make(map[string]string, len(details))
	for key, value := range details {
		if value == "" {
			continue
		}
		cloned[key] = value
	}

	return cloned
}
