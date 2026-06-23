/*******************************************************************************
 * IBM Confidential
 * OCO Source Materials
 * (C) Copyright IBM Corp. 2026 All Rights Reserved.
 * The source code for this program is not  published or otherwise divested of
 * its trade secrets, irrespective of what has been deposited with
 * the U.S. Copyright Office.
 ******************************************************************************/
package workflow

import (
	"context"

	"github.com/IBM/ibm-backup-recovery-sdk-go/migrationv2/common/errors"
	"github.com/IBM/ibm-backup-recovery-sdk-go/migrationv2/common/types"
	"github.com/IBM/ibm-backup-recovery-sdk-go/migrationv2/middlelevel/datasources"
)

// WorkflowAPI provides high-level migration workflows
type WorkflowAPI interface {
	// ExecuteSetup creates connection, deploys connector, and registers source
	ExecuteSetup(ctx context.Context, dataSource datasources.DataSource, connectionParams *types.ConnectionParams) (*types.SetupResult, *errors.SDKError)

	// ExecuteProtect creates protection group with policy
	ExecuteProtect(ctx context.Context, registrationID int64, groupParams *types.ProtectionGroupParams, datasource datasources.DataSource) (*types.ProtectResult, *errors.SDKError)

	// ExecuteBackup starts backup job
	ExecuteBackup(ctx context.Context, protectionGroup *types.ProtectionGroupResult, backupParams *types.BackupParams) (*types.BackupResult, *errors.SDKError)

	// ExecuteRestore starts restore job
	ExecuteRestore(ctx context.Context, groupId, backupID string, targetRegistrationID int64, restoreParams *types.RestoreParams, datasource datasources.DataSource) (*types.RestoreResult, *errors.SDKError)

	// ExecuteMigration executes complete migration workflow
	ExecuteMigration(ctx context.Context, sourceDataSource datasources.DataSource, targetDataSource datasources.DataSource, migrationParams *types.MigrationParams) (*types.MigrationResult, *errors.SDKError)

	// ExecuteCleanup removes migration resources
	ExecuteCleanup(ctx context.Context, cleanupParams *types.CleanupParams) (*types.CleanupResult, *errors.SDKError)

	// ExecuteControl controls running workflows (pause, abort, resume, cancel)
	ExecuteControl(ctx context.Context, controlParams *types.ControlParams) (*types.ControlResult, *errors.SDKError)
}
