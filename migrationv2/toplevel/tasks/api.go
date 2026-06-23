/*******************************************************************************
 * IBM Confidential
 * OCO Source Materials
 * (C) Copyright IBM Corp. 2026 All Rights Reserved.
 * The source code for this program is not  published or otherwise divested of
 * its trade secrets, irrespective of what has been deposited with
 * the U.S. Copyright Office.
 ******************************************************************************/
package tasks

import (
	"context"
	"time"

	"github.com/IBM/ibm-backup-recovery-sdk-go/migrationv2/common/errors"
	"github.com/IBM/ibm-backup-recovery-sdk-go/migrationv2/common/types"
	"github.com/IBM/ibm-backup-recovery-sdk-go/migrationv2/middlelevel/connectors"
	"github.com/IBM/ibm-backup-recovery-sdk-go/migrationv2/middlelevel/datasources"
)

// TaskAPI provides CRUD operations for BRS resources
type TaskAPI interface {
	//===================================================================
	//
	// CONNECTION OPERATIONS
	//
	//===================================================================
	CreateConnection(ctx context.Context, connectionParams *types.ConnectionParams) (*types.ConnectionResult, *errors.SDKError)
	GetConnection(ctx context.Context, connectionID string) (*types.ConnectionResult, *errors.SDKError)
	GetConnectionByName(ctx context.Context, connectionName string) (*types.ConnectionResult, *errors.SDKError)
	ListConnections(ctx context.Context) ([]*types.ConnectionResult, *errors.SDKError)
	DeleteConnection(ctx context.Context, connectionID string) *errors.SDKError

	//===================================================================
	//
	// CONNECTOR OPERATIONS
	//
	//===================================================================
	DeployConnector(ctx context.Context, connector connectors.ConnectorDeployer, connectionResult *types.ConnectionResult) (*connectors.ConnectorResult, *errors.SDKError)
	GetConnector(ctx context.Context, connectorID string) (*connectors.ConnectorResult, *errors.SDKError)
	GetConnectorByConnection(ctx context.Context, connectionID string) ([]*connectors.ConnectorResult, *errors.SDKError)
	DeleteConnector(ctx context.Context, connectorID string, connector ...connectors.ConnectorDeployer) *errors.SDKError

	//===================================================================
	//
	// REGISTRATION OPERATIONS
	//
	//===================================================================
	RegisterSource(ctx context.Context, datasource datasources.DataSource, connectionID string) (*types.RegistrationResult, *errors.SDKError)
	GetRegistration(ctx context.Context, registrationID int64) (*types.RegistrationResult, *errors.SDKError)
	GetRegistrationByConnection(ctx context.Context, connectionID string) (*types.RegistrationResult, *errors.SDKError)
	ListRegistrations(ctx context.Context) ([]*types.RegistrationResult, *errors.SDKError)
	UnregisterSource(ctx context.Context, registrationID int64) *errors.SDKError
	RefreshRegistration(ctx context.Context, registrationID int64) *errors.SDKError

	//===================================================================
	//
	// PROTECTION OPERATIONS
	//
	//===================================================================
	CreateProtectionGroup(ctx context.Context, registrationID int64, groupParams *types.ProtectionGroupParams, dataSource datasources.DataSource) (*types.ProtectionGroupResult, *errors.SDKError)
	GetProtectionGroup(ctx context.Context, groupID string) (*types.ProtectionGroupResult, *errors.SDKError)
	GetProtectionGroupByName(ctx context.Context, groupName string) (*types.ProtectionGroupResult, *errors.SDKError)
	ListProtectionGroups(ctx context.Context, registrationID int64) ([]*types.ProtectionGroupResult, *errors.SDKError)
	UpdateProtectionGroup(ctx context.Context, groupID string, updateParams *types.ProtectionGroupParams) (*types.ProtectionGroupResult, *errors.SDKError)
	DeleteProtectionGroup(ctx context.Context, groupID string, deleteSnapshots bool) *errors.SDKError

	//===================================================================
	//
	// POLICY OPERATIONS
	//
	//===================================================================
	CreatePolicy(ctx context.Context, policyParams *types.PolicyParams) (*types.PolicyResult, *errors.SDKError)
	GetPolicy(ctx context.Context, policyID string) (*types.PolicyResult, *errors.SDKError)
	GetPolicyByName(ctx context.Context, policyName string) (*types.PolicyResult, *errors.SDKError)
	ListPolicies(ctx context.Context) ([]*types.PolicyResult, *errors.SDKError)
	UpdatePolicy(ctx context.Context, policyID string, policyParams *types.PolicyParams) (*types.PolicyResult, *errors.SDKError)
	DeletePolicy(ctx context.Context, policyID string) *errors.SDKError

	//===================================================================
	//
	// BACKUP OPERATIONS
	//
	//===================================================================
	RunBackup(ctx context.Context, groupID string, backupParams *types.BackupParams) (*types.BackupResult, *errors.SDKError)
	GetBackup(ctx context.Context, backupID, groupID string) (*types.BackupResult, *errors.SDKError)
	ListBackups(ctx context.Context, groupID string) ([]*types.BackupResult, *errors.SDKError)
	WaitForBackup(ctx context.Context, backupID string, timeout time.Duration) (*types.BackupResult, *errors.SDKError)

	//===================================================================
	//
	// RESTORE OPERATIONS
	//
	//===================================================================
	RunRestore(ctx context.Context, groupId, backupId string, targetRegistrationID int64, restoreParams *types.RestoreParams, datasource datasources.DataSource) (*types.RestoreResult, *errors.SDKError)
	GetRestore(ctx context.Context, restoreID string) (*types.RestoreResult, *errors.SDKError)
	ListRestores(ctx context.Context, registrationID int64) ([]*types.RestoreResult, *errors.SDKError)
	WaitForRestore(ctx context.Context, restoreID string, timeout time.Duration) (*types.RestoreResult, *errors.SDKError)

	//===================================================================
	//
	// CONTROL OPERATIONS
	//
	//===================================================================
	GetProtectionRunProgress(ctx context.Context, backupID string) (*float32, *errors.SDKError)
	PauseBackup(ctx context.Context, backupID, groupID string, force bool) *errors.SDKError
	ResumeBackup(ctx context.Context, backupID, groupID string) *errors.SDKError
	AbortBackup(ctx context.Context, backupID, groupID string, force bool) *errors.SDKError
	AbortRestore(ctx context.Context, restoreID string, force bool) *errors.SDKError
}
