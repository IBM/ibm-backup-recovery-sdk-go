/*******************************************************************************
 * IBM Confidential
 * OCO Source Materials
 * (C) Copyright IBM Corp. 2026 All Rights Reserved.
 * The source code for this program is not  published or otherwise divested of
 * its trade secrets, irrespective of what has been deposited with
 * the U.S. Copyright Office.
 ******************************************************************************/

package context

import (
	"context"

	"github.com/google/uuid"
)

// contextKey is a private type for context keys.
type contextKey string

// Context keys for operational fields.
const (
	requestIDKey     contextKey = "request_id"
	transactionIDKey contextKey = "transaction_id"
	accountIDKey     contextKey = "account_id"
)

// Fields holds contextual fields for attaching to context.
// Use NewFields() to create and chain methods.
type Fields struct {
	requestID     string
	transactionID string
	accountID     string
}

// NewFields creates a new Fields builder.
func NewFields() *Fields {
	return &Fields{}
}

// WithRequestID sets a request ID for tracking a single request.
func (f *Fields) WithRequestID(id string) *Fields {
	f.requestID = id
	return f
}

// WithTransactionID sets a transaction ID for tracking a business transaction.
func (f *Fields) WithTransactionID(id string) *Fields {
	f.transactionID = id
	return f
}

// WithAccountID sets the account ID.
func (f *Fields) WithAccountID(id string) *Fields {
	f.accountID = id
	return f
}

// Context attaches all fields to the given context.
func (f *Fields) Context(ctx context.Context) context.Context {
	if f.requestID != "" {
		ctx = context.WithValue(ctx, requestIDKey, f.requestID)
	}

	if f.transactionID != "" {
		ctx = context.WithValue(ctx, transactionIDKey, f.transactionID)
	}

	if f.accountID != "" {
		ctx = context.WithValue(ctx, accountIDKey, f.accountID)
	}
	return ctx
}

// WithRequestID returns a context with request ID.
func WithRequestID(ctx context.Context, id string) context.Context {
	return context.WithValue(ctx, requestIDKey, id)
}

// WithTransactionID returns a context with transaction ID.
func WithTransactionID(ctx context.Context, id string) context.Context {
	return context.WithValue(ctx, transactionIDKey, id)
}

// WithAccountID returns a context with account ID.
func WithAccountID(ctx context.Context, id string) context.Context {
	return context.WithValue(ctx, accountIDKey, id)
}

// GetRequestID retrieves the request ID from context.
func GetRequestID(ctx context.Context) string {
	return getStringValue(ctx, requestIDKey)
}

// GetTransactionID retrieves the transaction ID from context.
func GetTransactionID(ctx context.Context) string {
	return getStringValue(ctx, transactionIDKey)
}

// GetAccountID retrieves the account ID from context.
func GetAccountID(ctx context.Context) string {
	return getStringValue(ctx, accountIDKey)
}

// TransactionIDFromContext extracts a transaction identifier from the context, if present.
// If no transaction ID is found in the context, it generates and returns a new UUID.
func TransactionIDFromContext(ctx context.Context) string {
	if v := GetTransactionID(ctx); v != "" {
		return v
	}
	// Generate a new UUID if no transaction ID is present
	return uuid.New().String()
}

// AccountIDFromContext extracts an account identifier from the context, if present.
// Returns empty string if no account ID is found.
func AccountIDFromContext(ctx context.Context) string {
	return GetAccountID(ctx)
}

// getStringValue extracts string values from context.
func getStringValue(ctx context.Context, key contextKey) string {
	if ctx == nil {
		return ""
	}
	if v := ctx.Value(key); v != nil {
		if s, ok := v.(string); ok {
			return s
		}
	}
	return ""
}

// ExtractContextFields extracts all known fields from context as key-value pairs.
func ExtractContextFields(ctx context.Context) []any {
	if ctx == nil {
		return nil
	}

	var args []any

	if v := GetRequestID(ctx); v != "" {
		args = append(args, "request_id", v)
	}

	if v := GetTransactionID(ctx); v != "" {
		args = append(args, "transaction_id", v)
	}

	if v := GetAccountID(ctx); v != "" {
		args = append(args, "account_id", v)
	}

	return args
}

// Made with Bob
