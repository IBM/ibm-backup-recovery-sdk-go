/*******************************************************************************
 * IBM Confidential
 * OCO Source Materials
 * (C) Copyright IBM Corp. 2026 All Rights Reserved.
 * The source code for this program is not  published or otherwise divested of
 * its trade secrets, irrespective of what has been deposited with
 * the U.S. Copyright Office.
 ******************************************************************************/
package errors

import (
	"fmt"

	"github.com/IBM/go-sdk-core/v5/core"
)

// ErrorCode represents specific error codes
type ErrorCode string

const (
	ErrCodeInvalidConfig        ErrorCode = "INVALID_CONFIG"
	ErrCodeInvalidInput         ErrorCode = "INVALID_INPUT"
	ErrCodeConnectionFailed     ErrorCode = "CONNECTION_FAILED"
	ErrCodeConnectorNotFound    ErrorCode = "CONNECTOR_NOT_FOUND"
	ErrCodeConnectorFailed      ErrorCode = "CONNECTOR_FAILED"
	ErrCodeRegistrationFailed   ErrorCode = "REGISTRATION_FAILED"
	ErrCodeProtectionFailed     ErrorCode = "PROTECTION_FAILED"
	ErrCodeBackupFailed         ErrorCode = "BACKUP_FAILED"
	ErrCodeRestoreFailed        ErrorCode = "RESTORE_FAILED"
	ErrCodeDataSourceNotFound   ErrorCode = "DATASOURCE_NOT_FOUND"
	ErrCodeRegistrationNotFound ErrorCode = "REGISTRATION_NOT_FOUND"
	ErrCodeInvalidDataSource    ErrorCode = "INVALID_DATASOURCE"
	ErrCodeWorkflowFailed       ErrorCode = "WORKFLOW_FAILED"
	ErrCodeTaskFailed           ErrorCode = "TASK_FAILED"
	ErrCodeTimeout              ErrorCode = "TIMEOUT"
	ErrCodeUnknown              ErrorCode = "UNKNOWN"
	ErrCodeConnectionNotFound   ErrorCode = "CONNECTION_NOT_FOUND"
)

// SDKError represents a structured error from the SDK
type SDKError struct {
	Code    ErrorCode
	Message string
	Cause   error
	Details map[string]interface{}
}

// Error implements the error interface
func (e *SDKError) Error() string {
	if e.Cause != nil {
		return fmt.Sprintf("[%s] %s: %v", e.Code, e.Message, e.Cause)
	}

	return fmt.Sprintf("[%s] %s", e.Code, e.Message)
}

// Unwrap returns the underlying cause
func (e *SDKError) Unwrap() error {
	return e.Cause
}

// NewSDKError creates a new SDK error
func NewSDKError(code ErrorCode, message string, cause error) *SDKError {
	return &SDKError{
		Code:    code,
		Message: message,
		Cause:   cause,
		Details: make(map[string]interface{}),
	}
}

// WithDetails adds details to the error
func (e *SDKError) WithDetails(key string, value interface{}) *SDKError {
	e.Details[key] = value
	return e
}

// Helper functions for common errors
func NewInvalidConfigError(message string, cause error) *SDKError {
	return NewSDKError(ErrCodeInvalidConfig, message, cause)
}

func NewConnectionFailedError(message string, cause error) *SDKError {
	return NewSDKError(ErrCodeConnectionFailed, message, cause)
}

func NewConnectorFailedError(message string, cause error) *SDKError {
	return NewSDKError(ErrCodeConnectorFailed, message, cause)
}

func NewConnectorNotFound(message string, cause error) *SDKError {
	return NewSDKError(ErrCodeConnectorNotFound, message, cause)
}

func NewRegistrationFailedError(message string, cause error) *SDKError {
	return NewSDKError(ErrCodeRegistrationFailed, message, cause)
}

func NewProtectionFailedError(message string, cause error) *SDKError {
	return NewSDKError(ErrCodeProtectionFailed, message, cause)
}

func NewBackupFailedError(message string, cause error) *SDKError {
	return NewSDKError(ErrCodeBackupFailed, message, cause)
}

func NewRestoreFailedError(message string, cause error) *SDKError {
	return NewSDKError(ErrCodeRestoreFailed, message, cause)
}

func NewDataSourceNotFoundError(message string) *SDKError {
	return NewSDKError(ErrCodeDataSourceNotFound, message, nil)
}

func NewInvalidDataSourceError(message string) *SDKError {
	return NewSDKError(ErrCodeInvalidDataSource, message, nil)
}

func NewWorkflowFailedError(message string, cause error) *SDKError {
	return NewSDKError(ErrCodeWorkflowFailed, message, cause)
}

func NewTaskFailedError(message string, cause error) *SDKError {
	return NewSDKError(ErrCodeTaskFailed, message, cause)
}

func NewTimeoutError(message string) *SDKError {
	return NewSDKError(ErrCodeTimeout, message, nil)
}

func IsStatusCode(err error, statusCode int) bool {
	sdkProb, ok := err.(core.SDKProblem)
	if !ok {
		return false
	}

	cause := sdkProb.GetCausedBy()
	httpProb, ok := cause.(*core.HTTPProblem)
	if !ok || httpProb == nil || httpProb.Response == nil {
		return false
	}

	return httpProb.Response.StatusCode == statusCode
}

func IsNotFoundError(err error) bool {
	sdkErr, ok := err.(*SDKError)
	if !ok {
		return false
	}

	return sdkErr.Code == ErrCodeConnectionNotFound
}

func ErrorCodeMatch(err error, code ErrorCode) bool {
	sdkErr, ok := err.(*SDKError)
	if !ok {
		return false
	}

	return sdkErr.Code == code
}
