/*******************************************************************************
 * IBM Confidential
 * OCO Source Materials
 * (C) Copyright IBM Corp. 2026 All Rights Reserved.
 * The source code for this program is not  published or otherwise divested of
 * its trade secrets, irrespective of what has been deposited with
 * the U.S. Copyright Office.
 ******************************************************************************/
package helpers

import (
	"context"
	"fmt"
	"time"

	"github.com/IBM/go-sdk-core/v5/core"
	"github.com/IBM/ibm-backup-recovery-sdk-go/backuprecoveryv1"
	"github.com/IBM/ibm-backup-recovery-sdk-go/migrationv2/common/errors"
	"github.com/IBM/ibm-backup-recovery-sdk-go/migrationv2/common/types"
)

// BRSQueryHelper provides helper functions to query BRS for existing resources
type BRSQueryHelper struct {
	brsClient types.BRSClientWrapperInterface
}

// NewBRSQueryHelper creates a new BRS query helper
func NewBRSQueryHelper(brsClient types.BRSClientWrapperInterface) *BRSQueryHelper {
	return &BRSQueryHelper{
		brsClient: brsClient,
	}
}

// GetConnectionByName checks if a connection with the given name exists in BRS
// Returns the connection if found, nil if not found, error on failure
func (h *BRSQueryHelper) GetConnectionByName(ctx context.Context, connectionName string) (*backuprecoveryv1.DataSourceConnection, *errors.SDKError) {
	listOptions := &backuprecoveryv1.GetDataSourceConnectionsOptions{
		ConnectionNames: []string{connectionName},
		XIBMTenantID:    core.StringPtr(h.brsClient.GetTenantId()),
	}

	connections, _, err := h.brsClient.GetBRSClient().GetDataSourceConnections(listOptions)
	if err != nil {
		return nil, errors.NewConnectionFailedError("failed to list connections", err)
	}

	if connections == nil || len(connections.Connections) == 0 {
		return nil, errors.NewSDKError(errors.ErrCodeConnectionNotFound, fmt.Sprintf("Connection details not found with name: %v", connectionName), nil)
	}

	return &connections.Connections[0], nil
}

// GetConnectorByConnection checks if a connector exists for the given connection
// Returns the connector info if found, nil if not found, error on failure
func (h *BRSQueryHelper) GetConnectorByConnection(ctx context.Context, connectionID string) ([]backuprecoveryv1.DataSourceConnector, *errors.SDKError) {
	result, _, err := h.brsClient.GetBRSClient().GetDataSourceConnectors(&backuprecoveryv1.GetDataSourceConnectorsOptions{
		XIBMTenantID: core.StringPtr(h.brsClient.GetTenantId()),
		ConnectionID: core.StringPtr(connectionID),
	})
	if err != nil {
		return nil, errors.NewConnectorFailedError("failed to get connector by connectionID", err)
	}

	if result == nil || len(result.Connectors) == 0 {
		return nil, errors.NewConnectorNotFound(fmt.Sprintf("No Connectors found for connectionID: %v", connectionID), nil)
	}

	return result.Connectors, nil
}

// GetRegistrationByConnection checks if a registration exists for the given connection
// Returns the registration if found, nil if not found, error on failure
func (h *BRSQueryHelper) GetRegistrationByConnection(ctx context.Context, connectionID string) (*backuprecoveryv1.SourceRegistrationResponseParams, *errors.SDKError) {
	// List all registrations and filter by connection ID
	listOptions := &backuprecoveryv1.GetSourceRegistrationsOptions{
		XIBMTenantID: core.StringPtr(h.brsClient.GetTenantId()),
	}

	registrations, _, err := h.brsClient.GetBRSClient().GetSourceRegistrationsWithContext(ctx, listOptions)
	if err != nil {
		return nil, errors.NewRegistrationFailedError("failed to list registrations", err)
	}

	// Find registration by connection ID
	if registrations != nil && registrations.Registrations != nil {
		for _, reg := range registrations.Registrations {
			// Check if this registration is associated with the connection
			// This logic depends on how BRS stores the connection-registration relationship
			if reg.ConnectionID != nil && fmt.Sprintf("%v", *reg.ConnectionID) == connectionID {
				return &reg, nil
			}
		}
	}

	return nil, nil // Not found
}

// GetProtectionGroupByName checks if a protection group with the given name exists
// Returns the protection group if found, nil if not found, error on failure
func (h *BRSQueryHelper) GetProtectionGroupByName(ctx context.Context, groupName string) (*backuprecoveryv1.ProtectionGroupResponse, *errors.SDKError) {
	// List all protection groups and filter by name
	listOptions := &backuprecoveryv1.GetProtectionGroupsOptions{}

	groups, _, err := h.brsClient.GetBRSClient().GetProtectionGroupsWithContext(ctx, listOptions)
	if err != nil {
		return nil, errors.NewProtectionFailedError("failed to list protection groups", err)
	}

	// Find protection group by name
	if groups != nil && groups.ProtectionGroups != nil {
		for _, group := range groups.ProtectionGroups {
			if group.Name != nil && *group.Name == groupName {
				return &group, nil
			}
		}
	}

	return nil, nil // Not found
}

// GetLatestBackupForGroup gets the latest successful backup for a protection group
// Returns the backup info if found, nil if not found, error on failure
func (h *BRSQueryHelper) GetLatestBackupForGroup(ctx context.Context, groupID string) (map[string]interface{}, *errors.SDKError) {
	// Query BRS for latest backup run of this protection group
	// This is a placeholder - actual implementation depends on BRS API
	return nil, nil
}

// GetOperationStatus queries BRS for the status of an async operation
// Returns the operation status, error on failure
func (h *BRSQueryHelper) GetOperationStatus(ctx context.Context, operationID string) (*types.AsyncOperation, *errors.SDKError) {
	// Query BRS for operation status
	// This is a placeholder - actual implementation depends on BRS API
	// BRS may track operations differently (e.g., protection runs, restore tasks)

	// For now, return a placeholder
	return &types.AsyncOperation{
		OperationID:   operationID,
		OperationType: "unknown",
		Status:        "completed",
		Message:       "Operation completed",
	}, nil
}

// WaitForOperation polls BRS until an operation completes or times out
func (h *BRSQueryHelper) WaitForOperation(ctx context.Context, operationID string, config *types.PollingConfig) (*types.AsyncOperation, *errors.SDKError) {
	if config == nil {
		config = types.DefaultPollingConfig()
	}

	// Initial delay
	if config.InitialDelay > 0 {
		select {
		case <-ctx.Done():
			return nil, errors.NewSDKError(errors.ErrCodeUnknown, "context cancelled", ctx.Err())
		case <-time.After(config.InitialDelay):
		}
	}

	attempts := 0
	deadline := time.Now().Add(config.Timeout)

	for {
		// Check timeout
		if time.Now().After(deadline) {
			return nil, errors.NewTimeoutError(fmt.Sprintf("operation %s timed out after %v", operationID, config.Timeout))
		}

		// Check max attempts
		if config.MaxAttempts > 0 && attempts >= config.MaxAttempts {
			return nil, errors.NewTimeoutError(fmt.Sprintf("operation %s exceeded max attempts %d", operationID, config.MaxAttempts))
		}

		// Query operation status
		op, err := h.GetOperationStatus(ctx, operationID)
		if err != nil {
			return nil, err
		}

		// Check if operation is complete
		if op.Status == "completed" {
			return op, nil
		}
		if op.Status == "failed" {
			var opErr error
			if op.Error != "" {
				opErr = fmt.Errorf("%s", op.Error)
			}
			return op, errors.NewSDKError(errors.ErrCodeUnknown, fmt.Sprintf("operation %s failed: %s", operationID, op.Message), opErr)
		}

		// Wait before next poll
		attempts++
		select {
		case <-ctx.Done():
			return nil, errors.NewSDKError(errors.ErrCodeUnknown, "context cancelled", ctx.Err())
		case <-time.After(config.Interval):
		}
	}
}
