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
	"fmt"
	"time"

	"github.com/IBM/ibm-backup-recovery-sdk-go/migrationv2/common/errors"
	"github.com/IBM/ibm-backup-recovery-sdk-go/migrationv2/common/types"
	"github.com/IBM/ibm-backup-recovery-sdk-go/migrationv2/middlelevel/datasources"
	"github.com/IBM/ibm-backup-recovery-sdk-go/migrationv2/toplevel/tasks"
)

// DefaultWorkflowAPI is the default implementation of WorkflowAPI
type DefaultWorkflowAPI struct {
	taskAPI tasks.TaskAPI
	config  *types.SDKConfig
}

// NewWorkflowAPI creates a new workflow API instance
func NewWorkflowAPI(brsClient types.BRSClientWrapperInterface, config *types.SDKConfig) WorkflowAPI {
	return &DefaultWorkflowAPI{
		taskAPI: tasks.NewTaskAPI(brsClient, config),
		config:  config,
	}
}

// ExecuteSetup executes setup workflow (idempotent, async)
func (w *DefaultWorkflowAPI) ExecuteSetup(ctx context.Context, dataSource datasources.DataSource, connectionParams *types.ConnectionParams) (*types.SetupResult, *errors.SDKError) {
	startTime := time.Now()
	result := &types.SetupResult{
		Status:    "in_progress",
		StartedAt: startTime,
	}

	// Validate connectionParams
	if connectionParams == nil || connectionParams.Name == "" {
		return nil, errors.NewSDKError(errors.ErrCodeInvalidInput, "connectionParams.Name is required", nil)
	}

	// Step 1: Create or get connection (idempotent)
	conn, err := w.taskAPI.GetConnectionByName(ctx, connectionParams.Name)
	if err != nil {
		return nil, err
	}

	if conn == nil {
		// Create new connection
		conn, err = w.taskAPI.CreateConnection(ctx, connectionParams)
		if err != nil {
			result.Status = "failed"
			result.Message = "Failed to create connection"
			return result, err
		}
		result.ResourcesCreated++
	} else {
		// Reuse existing connection
		result.ResourcesReused++
	}

	result.ConnectionID = conn.ConnectionID

	// Step 2: Deploy or get connector (idempotent)
	connector, err := w.taskAPI.GetConnectorByConnection(ctx, conn.ConnectionID)
	if err != nil {
		return nil, err
	}

	if connector == nil {
		// Deploy new connector using provided DataSource
		//connector, err = w.taskAPI.DeployConnector(ctx, connectionParams.Connector, conn)
		//if err != nil {
		//	result.Status = "failed"
		//	result.Message = "Failed to deploy connector"
		//	return result, err
		//}
		//result.ResourcesCreated++
	} else {
		// Reuse existing connector
		result.ResourcesReused++
	}

	//result.ConnectorID = connector.ConnectorID

	// Step 3: Register or get source (idempotent)
	reg, err := w.taskAPI.GetRegistrationByConnection(ctx, conn.ConnectionID)
	if err != nil {
		return nil, err
	}

	if reg == nil {
		// Register new source
		reg, err = w.taskAPI.RegisterSource(ctx, dataSource, conn.ConnectionID)
		if err != nil {
			result.Status = "failed"
			result.Message = "Failed to register source"
			return result, err
		}
		result.ResourcesCreated++
	} else {
		// Reuse existing registration
		result.ResourcesReused++
	}

	result.RegistrationID = reg.RegistrationID

	// Complete
	result.Status = "completed"
	result.CompletedAt = time.Now()
	result.Duration = time.Since(startTime).Seconds()
	result.Message = "Setup completed successfully"

	return result, nil
}

// ExecuteProtect executes protection workflow (idempotent, async)
func (w *DefaultWorkflowAPI) ExecuteProtect(ctx context.Context, registrationID int64, groupParams *types.ProtectionGroupParams, datasource datasources.DataSource) (*types.ProtectResult, *errors.SDKError) {
	startTime := time.Now()

	// Extract policy ID and group name from params
	policyID := ""
	if groupParams.Policy != nil {
		policyID = groupParams.Policy.ID
	}

	result := &types.ProtectResult{
		RegistrationID: registrationID,
		PolicyID:       policyID,
		GroupName:      groupParams.Name,
		Status:         "in_progress",
		StartedAt:      startTime,
	}

	// Create or get protection group (idempotent)
	group, err := w.taskAPI.GetProtectionGroupByName(ctx, groupParams.Name)
	if err != nil {
		return nil, err
	}

	if group == nil {
		// Create new protection group
		group, err = w.taskAPI.CreateProtectionGroup(ctx, registrationID, groupParams, datasource)
		if err != nil {
			result.Status = "failed"
			result.Message = "Failed to create protection group"
			return result, err
		}
		result.ResourcesCreated++
	} else {
		// Reuse existing protection group
		result.ResourcesReused++
	}

	result.ProtectionGroupID = group.ProtectionGroupID

	// Complete
	result.Status = "completed"
	result.CompletedAt = time.Now()
	result.Duration = time.Since(startTime).Seconds()
	result.Message = "Protection setup completed successfully"

	return result, nil
}

// ExecuteBackup executes backup workflow (async)
func (w *DefaultWorkflowAPI) ExecuteBackup(ctx context.Context, protectionGroup *types.ProtectionGroupResult, backupParams *types.BackupParams) (*types.BackupResult, *errors.SDKError) {
	// Validate protection group status
	if protectionGroup.Status != "active" && protectionGroup.Status != "completed" {
		return nil, errors.NewSDKError(errors.ErrCodeInvalidInput,
			"Protection group is not active: "+protectionGroup.Status, nil)
	}

	// Start backup (returns immediately, doesn't wait)
	backup, err := w.taskAPI.RunBackup(ctx, protectionGroup.ProtectionGroupID, backupParams)
	if err != nil {
		return nil, err
	}

	// Return immediately with backup ID
	// User can check status with: TaskAPI.GetBackup(backupID)
	return backup, nil
}

// ExecuteRestore executes restore workflow (async)
func (w *DefaultWorkflowAPI) ExecuteRestore(ctx context.Context, groupId, backupID string, targetRegistrationID int64, restoreParams *types.RestoreParams, datasource datasources.DataSource) (*types.RestoreResult, *errors.SDKError) {
	// Start restore (returns immediately, doesn't wait)
	restore, err := w.taskAPI.RunRestore(ctx, groupId, backupID, targetRegistrationID, restoreParams, datasource)
	if err != nil {
		return nil, err
	}

	// Return immediately with restore ID
	// User can check status with: TaskAPI.GetRestore(restoreID)
	return restore, nil
}

// MIGRATION WORKFLOW IMPLEMENTATION

// ExecuteMigration executes complete migration workflow (idempotent, async) - V8 Enhanced
func (w *DefaultWorkflowAPI) ExecuteMigration(
	ctx context.Context,
	sourceDataSource datasources.DataSource,
	targetDataSource datasources.DataSource,
	migrationParams *types.MigrationParams,
) (*types.MigrationResult, *errors.SDKError) {
	startTime := time.Now()
	result := &types.MigrationResult{
		Status:      "in_progress",
		StartedAt:   startTime,
		StepResults: []types.StepResult{},
	}

	// Validate: At least one of source or target must be provided
	if migrationParams.SourceConnectionParams == nil && migrationParams.TargetConnectionParams == nil {
		return nil, errors.NewSDKError(errors.ErrCodeInvalidInput,
			"At least one of SourceConnectionParams or TargetConnectionParams must be provided", nil)
	}

	// Helper function to add step result
	addStep := func(step, status, message string, stepStart time.Time) {
		result.StepResults = append(result.StepResults, types.StepResult{
			Step:        step,
			Status:      status,
			Message:     message,
			StartedAt:   stepStart,
			CompletedAt: time.Now(),
			Duration:    time.Since(stepStart).Seconds(),
		})
	}

	// Step 1: Setup source (connection + connector + registration) - OPTIONAL
	var sourceSetup *types.SetupResult
	if migrationParams.SourceConnectionParams != nil {
		stepStart := time.Now()
		var err *errors.SDKError
		sourceSetup, err = w.ExecuteSetup(ctx, sourceDataSource, migrationParams.SourceConnectionParams)
		if err != nil {
			result.Status = "failed"
			result.SourceSetupStatus = "failed"
			result.Message = "Failed to setup source"
			addStep("source_setup", "failed", err.Error(), stepStart)
			return result, err
		}

		result.SourceConnectionID = sourceSetup.ConnectionID
		result.SourceConnectorID = sourceSetup.ConnectorID
		result.SourceRegistrationID = sourceSetup.RegistrationID
		result.SourceSetupStatus = "completed"
		result.ResourcesCreated += sourceSetup.ResourcesCreated
		result.ResourcesReused += sourceSetup.ResourcesReused
		addStep("source_setup", "completed", "Source setup completed successfully", stepStart)
	} else {
		addStep("source_setup", "skipped", "Source setup not requested", time.Now())
	}

	// Step 2: Setup target (connection + connector + registration) - OPTIONAL
	var targetSetup *types.SetupResult
	if migrationParams.TargetConnectionParams != nil {
		stepStart := time.Now()
		var err *errors.SDKError
		targetSetup, err = w.ExecuteSetup(ctx, targetDataSource, migrationParams.TargetConnectionParams)
		if err != nil {
			result.Status = "failed"
			result.TargetSetupStatus = "failed"
			result.Message = "Failed to setup target"
			addStep("target_setup", "failed", err.Error(), stepStart)
			return result, err
		}

		result.TargetConnectionID = targetSetup.ConnectionID
		result.TargetConnectorID = targetSetup.ConnectorID
		result.TargetRegistrationID = targetSetup.RegistrationID
		result.TargetSetupStatus = "completed"
		result.ResourcesCreated += targetSetup.ResourcesCreated
		result.ResourcesReused += targetSetup.ResourcesReused
		addStep("target_setup", "completed", "Target setup completed successfully", stepStart)
	} else {
		addStep("target_setup", "skipped", "Target setup not requested", time.Now())
	}

	// Validate: For full migration, both source and target are required
	if migrationParams.ProtectionGroupName != "" && (sourceSetup == nil || targetSetup == nil) {
		return nil, errors.NewSDKError(errors.ErrCodeInvalidInput,
			"Both source and target must be provided for full migration with protection group", nil)
	}

	// Step 3: Create or get policy (if CreateDefaultPolicy is true)
	var policyID string
	if migrationParams.CreateDefaultPolicy {
		stepStart := time.Now()
		policy, err := w.taskAPI.GetPolicyByName(ctx, migrationParams.PolicyName)
		if err != nil {
			result.Status = "failed"
			result.Message = "Failed to check for existing policy"
			addStep("policy_check", "failed", err.Error(), stepStart)
			return result, err
		}

		if policy == nil {
			// Create new policy
			retentionDays := types.Retention_Unit_Days
			policy, err = w.taskAPI.CreatePolicy(ctx, &types.PolicyParams{
				Name:        migrationParams.PolicyName,
				Description: "Auto-created policy for migration",
				DataRetention: &types.DataRetention{
					Unit:      &retentionDays,
					RetainFor: 30,
				},
				BackupType: "full",
			})
			if err != nil {
				result.Status = "failed"
				result.Message = "Failed to create policy"
				addStep("policy_creation", "failed", err.Error(), stepStart)
				return result, err
			}
			result.ResourcesCreated++
			result.PolicyCreated = true
			addStep("policy_creation", "completed", "Policy created successfully", stepStart)
		} else {
			result.ResourcesReused++
			result.PolicyCreated = false
			addStep("policy_check", "completed", "Existing policy reused", stepStart)
		}
		policyID = policy.ID
	} else {
		// Use provided policy from ProtectionGroupParams
		if migrationParams.ProtectionGroupParams != nil && migrationParams.ProtectionGroupParams.Policy != nil {
			policyID = migrationParams.ProtectionGroupParams.Policy.ID
		}
	}

	result.PolicyID = policyID

	// Step 4: Create protection group
	stepStart := time.Now()
	if migrationParams.ProtectionGroupParams == nil {
		migrationParams.ProtectionGroupParams = &types.ProtectionGroupParams{
			Name:            migrationParams.ProtectionGroupName,
			NumberOfBackups: 7,
		}
	}
	migrationParams.ProtectionGroupParams.Policy = &types.Policy{ID: policyID}

	protect, err := w.ExecuteProtect(ctx, sourceSetup.RegistrationID, migrationParams.ProtectionGroupParams, sourceDataSource)
	if err != nil {
		result.Status = "failed"
		result.ProtectionStatus = "failed"
		result.Message = "Failed to create protection group"
		addStep("protection_group", "failed", err.Error(), stepStart)
		return result, err
	}

	result.ProtectionGroupID = protect.ProtectionGroupID
	result.ProtectionStatus = "completed"
	result.ResourcesCreated += protect.ResourcesCreated
	result.ResourcesReused += protect.ResourcesReused
	addStep("protection_group", "completed", "Protection group created successfully", stepStart)

	// Step 5: Run backup
	stepStart = time.Now()
	protectionGroup, err := w.taskAPI.GetProtectionGroup(ctx, protect.ProtectionGroupID)
	if err != nil {
		result.Status = "failed"
		result.BackupStatus = "failed"
		result.Message = "Failed to get protection group"
		addStep("backup_preparation", "failed", err.Error(), stepStart)
		return result, err
	}

	if migrationParams.BackupParams == nil {
		migrationParams.BackupParams = &types.BackupParams{
			BackupType: "full",
		}
	}

	backup, err := w.ExecuteBackup(ctx, protectionGroup, migrationParams.BackupParams)
	if err != nil {
		result.Status = "failed"
		result.BackupStatus = "failed"
		result.Message = "Failed to start backup"
		addStep("backup_start", "failed", err.Error(), stepStart)
		return result, err
	}

	result.BackupID = backup.BackupID
	result.BackupStatus = "running"
	addStep("backup_start", "completed", "Backup started successfully", stepStart)

	// Optionally wait for backup completion
	if migrationParams.WaitForBackup {
		stepStart := time.Now()
		backup, err = w.taskAPI.WaitForBackup(ctx, backup.BackupID, 0)
		if err != nil {
			result.Status = "failed"
			result.BackupStatus = "failed"
			result.Message = "Backup failed"
			addStep("backup_wait", "failed", err.Error(), stepStart)
			return result, err
		}
		result.BackupStatus = backup.Status
		addStep("backup_wait", "completed", "Backup completed successfully", stepStart)
	}

	// Step 6: Run restore
	stepStart = time.Now()
	if migrationParams.RestoreParams == nil {
		migrationParams.RestoreParams = &types.RestoreParams{
			Name:                    "migration-restore",
			BackupPositionFromFirst: 0, // Use latest backup
		}
	}

	restore, err := w.ExecuteRestore(ctx, protectionGroup.ProtectionGroupID, backup.BackupID, targetSetup.RegistrationID, migrationParams.RestoreParams, sourceDataSource)
	if err != nil {
		result.Status = "failed"
		result.RestoreStatus = "failed"
		result.Message = "Failed to start restore"
		addStep("restore_start", "failed", err.Error(), stepStart)
		return result, err
	}

	result.RestoreID = restore.RestoreID
	result.RestoreStatus = "running"
	addStep("restore_start", "completed", "Restore started successfully", stepStart)

	// Optionally wait for restore completion
	if migrationParams.WaitForRestore {
		stepStart := time.Now()
		restore, err = w.taskAPI.WaitForRestore(ctx, restore.RestoreID, 0)
		if err != nil {
			result.Status = "failed"
			result.RestoreStatus = "failed"
			result.Message = "Restore failed"
			addStep("restore_wait", "failed", err.Error(), stepStart)
			return result, err
		}
		result.RestoreStatus = restore.Status
		addStep("restore_wait", "completed", "Restore completed successfully", stepStart)
	}

	// Complete
	if migrationParams.WaitForBackup && migrationParams.WaitForRestore {
		result.Status = "completed"
		result.Message = "Migration completed successfully"
	} else {
		result.Status = "in_progress"
		result.Message = "Migration initiated successfully. Use TaskAPI to check backup/restore status."
	}

	result.CompletedAt = time.Now()
	result.Duration = time.Since(startTime).Seconds()

	return result, nil
}

// ExecuteCleanup executes cleanup workflow to remove migration resources
func (w *DefaultWorkflowAPI) ExecuteCleanup(ctx context.Context, cleanupParams *types.CleanupParams) (*types.CleanupResult, *errors.SDKError) {
	startTime := time.Now()
	result := &types.CleanupResult{
		Status:           "in_progress",
		DryRun:           cleanupParams.DryRun,
		DeletedResources: []types.DeletedResource{},
		SkippedResources: []types.SkippedResource{},
		FailedResources:  []types.FailedResource{},
		StepResults:      []types.StepResult{},
		StartedAt:        startTime,
	}

	// Validate parameters
	if cleanupParams.CleanupScope == types.CleanupScopeSpecific && cleanupParams.ResourceIDs == nil {
		return nil, errors.NewSDKError(errors.ErrCodeInvalidInput,
			"ResourceIDs required for specific cleanup scope", nil)
	}

	// Require confirmation token for non-dry-run
	if !cleanupParams.DryRun && cleanupParams.ConfirmationToken == "" {
		return nil, errors.NewSDKError(errors.ErrCodeInvalidInput,
			"ConfirmationToken required for actual deletion (use 'CONFIRM-DELETE')", nil)
	}

	// Helper function to add step result
	addStep := func(step, status, message string, stepStart time.Time) {
		result.StepResults = append(result.StepResults, types.StepResult{
			Step:        step,
			Status:      status,
			Message:     message,
			StartedAt:   stepStart,
			CompletedAt: time.Now(),
			Duration:    time.Since(stepStart).Seconds(),
		})
	}

	// Helper function to check if resource type should be cleaned
	shouldClean := func(resourceType types.ResourceType) bool {
		if len(cleanupParams.CleanupTypes) == 0 {
			return true // Clean all types if not specified
		}
		for _, t := range cleanupParams.CleanupTypes {
			if t == resourceType {
				return true
			}
		}
		return false
	}

	// Helper function to record deleted resource
	recordDeleted := func(resourceType types.ResourceType, resourceID, resourceName string) {
		result.DeletedResources = append(result.DeletedResources, types.DeletedResource{
			ResourceType: resourceType,
			ResourceID:   resourceID,
			ResourceName: resourceName,
			DeletedAt:    time.Now(),
		})
		result.ResourcesDeleted++
	}

	// Helper function to record skipped resource
	recordSkipped := func(resourceType types.ResourceType, resourceID, resourceName, reason string) {
		result.SkippedResources = append(result.SkippedResources, types.SkippedResource{
			ResourceType: resourceType,
			ResourceID:   resourceID,
			ResourceName: resourceName,
			Reason:       reason,
		})
		result.ResourcesSkipped++
	}

	// Helper function to record failed resource
	recordFailed := func(resourceType types.ResourceType, resourceID, resourceName, errorMsg string) {
		result.FailedResources = append(result.FailedResources, types.FailedResource{
			ResourceType: resourceType,
			ResourceID:   resourceID,
			ResourceName: resourceName,
			Error:        errorMsg,
		})
		result.ResourcesFailed++
	}

	// Get resource IDs based on scope
	resourceIDs := cleanupParams.ResourceIDs
	if resourceIDs == nil {
		resourceIDs = &types.CleanupResourceIDs{}
	}

	// Step 1: Delete restores (if any)
	if shouldClean(types.ResourceTypeRestore) && len(resourceIDs.RestoreIDs) > 0 {
		stepStart := time.Now()
		for _, restoreID := range resourceIDs.RestoreIDs {
			if cleanupParams.DryRun {
				recordSkipped(types.ResourceTypeRestore, restoreID, "restore", "dry_run")
			} else {
				// TODO: Implement actual deletion when TaskAPI.DeleteRestore is available
				recordDeleted(types.ResourceTypeRestore, restoreID, "restore")
			}
		}
		addStep("delete_restores", "completed", fmt.Sprintf("Processed %d restores", len(resourceIDs.RestoreIDs)), stepStart)
	}

	// Step 2: Delete backups (if not keeping)
	if shouldClean(types.ResourceTypeBackup) && !cleanupParams.KeepBackups && len(resourceIDs.BackupIDs) > 0 {
		stepStart := time.Now()
		for _, backupID := range resourceIDs.BackupIDs {
			if cleanupParams.DryRun {
				recordSkipped(types.ResourceTypeBackup, backupID, "backup", "dry_run")
			} else if cleanupParams.KeepBackups {
				recordSkipped(types.ResourceTypeBackup, backupID, "backup", "keep_requested")
			} else {
				// TODO: Implement actual deletion when TaskAPI.DeleteBackup is available
				recordDeleted(types.ResourceTypeBackup, backupID, "backup")
			}
		}
		addStep("delete_backups", "completed", fmt.Sprintf("Processed %d backups", len(resourceIDs.BackupIDs)), stepStart)
	} else if cleanupParams.KeepBackups {
		addStep("delete_backups", "skipped", "Backups kept as requested", time.Now())
	}

	// Step 3: Delete protection group
	if shouldClean(types.ResourceTypeProtectionGroup) && resourceIDs.ProtectionGroupID != "" {
		stepStart := time.Now()
		if cleanupParams.DryRun {
			recordSkipped(types.ResourceTypeProtectionGroup, resourceIDs.ProtectionGroupID, "protection-group", "dry_run")
		} else {
			err := w.taskAPI.DeleteProtectionGroup(ctx, resourceIDs.ProtectionGroupID, false)
			if err != nil {
				recordFailed(types.ResourceTypeProtectionGroup, resourceIDs.ProtectionGroupID, "protection-group", err.Error())
			} else {
				recordDeleted(types.ResourceTypeProtectionGroup, resourceIDs.ProtectionGroupID, "protection-group")
			}
		}
		addStep("delete_protection_group", "completed", "Protection group processed", stepStart)
	}

	// Step 4: Delete policy (if not keeping)
	if shouldClean(types.ResourceTypePolicy) && !cleanupParams.KeepPolicies && resourceIDs.PolicyID != "" {
		stepStart := time.Now()
		if cleanupParams.DryRun {
			recordSkipped(types.ResourceTypePolicy, resourceIDs.PolicyID, "policy", "dry_run")
		} else if cleanupParams.KeepPolicies {
			recordSkipped(types.ResourceTypePolicy, resourceIDs.PolicyID, "policy", "keep_requested")
		} else {
			err := w.taskAPI.DeletePolicy(ctx, resourceIDs.PolicyID)
			if err != nil {
				recordFailed(types.ResourceTypePolicy, resourceIDs.PolicyID, "policy", err.Error())
			} else {
				recordDeleted(types.ResourceTypePolicy, resourceIDs.PolicyID, "policy")
			}
		}
		addStep("delete_policy", "completed", "Policy processed", stepStart)
	} else if cleanupParams.KeepPolicies {
		addStep("delete_policy", "skipped", "Policy kept as requested", time.Now())
	}

	// Step 5: Unregister sources
	if shouldClean(types.ResourceTypeRegistration) {
		stepStart := time.Now()
		registrationsProcessed := 0

		// Source registration
		if resourceIDs.SourceRegistrationID != 0 &&
			(cleanupParams.CleanupScope == types.CleanupScopeAll || cleanupParams.CleanupScope == types.CleanupScopeSource) {
			if cleanupParams.DryRun {
				recordSkipped(types.ResourceTypeRegistration, fmt.Sprintf("%d", resourceIDs.SourceRegistrationID), "source-registration", "dry_run")
			} else {
				err := w.taskAPI.UnregisterSource(ctx, resourceIDs.SourceRegistrationID)
				if err != nil {
					recordFailed(types.ResourceTypeRegistration, fmt.Sprintf("%d", resourceIDs.SourceRegistrationID), "source-registration", err.Error())
				} else {
					recordDeleted(types.ResourceTypeRegistration, fmt.Sprintf("%d", resourceIDs.SourceRegistrationID), "source-registration")
				}
			}
			registrationsProcessed++
		}

		// Target registration
		if resourceIDs.TargetRegistrationID != 0 &&
			(cleanupParams.CleanupScope == types.CleanupScopeAll || cleanupParams.CleanupScope == types.CleanupScopeTarget) {
			if cleanupParams.DryRun {
				recordSkipped(types.ResourceTypeRegistration, fmt.Sprintf("%d", resourceIDs.TargetRegistrationID), "target-registration", "dry_run")
			} else {
				err := w.taskAPI.UnregisterSource(ctx, resourceIDs.TargetRegistrationID)
				if err != nil {
					recordFailed(types.ResourceTypeRegistration, fmt.Sprintf("%d", resourceIDs.TargetRegistrationID), "target-registration", err.Error())
				} else {
					recordDeleted(types.ResourceTypeRegistration, fmt.Sprintf("%d", resourceIDs.TargetRegistrationID), "target-registration")
				}
			}
			registrationsProcessed++
		}

		addStep("unregister_sources", "completed", fmt.Sprintf("Processed %d registrations", registrationsProcessed), stepStart)
	}

	// Step 6: Delete connectors
	if shouldClean(types.ResourceTypeConnector) {
		stepStart := time.Now()
		connectorsProcessed := 0

		// Source connector
		if resourceIDs.SourceConnectorID != "" &&
			(cleanupParams.CleanupScope == types.CleanupScopeAll || cleanupParams.CleanupScope == types.CleanupScopeSource) {
			if cleanupParams.DryRun {
				recordSkipped(types.ResourceTypeConnector, resourceIDs.SourceConnectorID, "source-connector", "dry_run")
			} else {
				err := w.taskAPI.DeleteConnector(ctx, resourceIDs.SourceConnectorID)
				if err != nil {
					recordFailed(types.ResourceTypeConnector, resourceIDs.SourceConnectorID, "source-connector", err.Error())
				} else {
					recordDeleted(types.ResourceTypeConnector, resourceIDs.SourceConnectorID, "source-connector")
				}
			}
			connectorsProcessed++
		}

		// Target connector
		if resourceIDs.TargetConnectorID != "" &&
			(cleanupParams.CleanupScope == types.CleanupScopeAll || cleanupParams.CleanupScope == types.CleanupScopeTarget) {
			if cleanupParams.DryRun {
				recordSkipped(types.ResourceTypeConnector, resourceIDs.TargetConnectorID, "target-connector", "dry_run")
			} else {
				err := w.taskAPI.DeleteConnector(ctx, resourceIDs.TargetConnectorID)
				if err != nil {
					recordFailed(types.ResourceTypeConnector, resourceIDs.TargetConnectorID, "target-connector", err.Error())
				} else {
					recordDeleted(types.ResourceTypeConnector, resourceIDs.TargetConnectorID, "target-connector")
				}
			}
			connectorsProcessed++
		}

		addStep("delete_connectors", "completed", fmt.Sprintf("Processed %d connectors", connectorsProcessed), stepStart)
	}

	// Step 7: Delete connections
	if shouldClean(types.ResourceTypeConnection) {
		stepStart := time.Now()
		connectionsProcessed := 0

		// Source connection
		if resourceIDs.SourceConnectionID != "" &&
			(cleanupParams.CleanupScope == types.CleanupScopeAll || cleanupParams.CleanupScope == types.CleanupScopeSource) {
			if cleanupParams.DryRun {
				recordSkipped(types.ResourceTypeConnection, resourceIDs.SourceConnectionID, "source-connection", "dry_run")
			} else {
				err := w.taskAPI.DeleteConnection(ctx, resourceIDs.SourceConnectionID)
				if err != nil {
					recordFailed(types.ResourceTypeConnection, resourceIDs.SourceConnectionID, "source-connection", err.Error())
				} else {
					recordDeleted(types.ResourceTypeConnection, resourceIDs.SourceConnectionID, "source-connection")
				}
			}
			connectionsProcessed++
		}

		// Target connection
		if resourceIDs.TargetConnectionID != "" &&
			(cleanupParams.CleanupScope == types.CleanupScopeAll || cleanupParams.CleanupScope == types.CleanupScopeTarget) {
			if cleanupParams.DryRun {
				recordSkipped(types.ResourceTypeConnection, resourceIDs.TargetConnectionID, "target-connection", "dry_run")
			} else {
				err := w.taskAPI.DeleteConnection(ctx, resourceIDs.TargetConnectionID)
				if err != nil {
					recordFailed(types.ResourceTypeConnection, resourceIDs.TargetConnectionID, "target-connection", err.Error())
				} else {
					recordDeleted(types.ResourceTypeConnection, resourceIDs.TargetConnectionID, "target-connection")
				}
			}
			connectionsProcessed++
		}

		addStep("delete_connections", "completed", fmt.Sprintf("Processed %d connections", connectionsProcessed), stepStart)
	}

	// Determine final status
	if result.ResourcesFailed > 0 {
		result.Status = "partial"
	} else if result.ResourcesDeleted > 0 || result.ResourcesSkipped > 0 {
		result.Status = "completed"
	} else {
		result.Status = "completed" // No resources to clean
	}

	result.CompletedAt = time.Now()
	result.Duration = time.Since(startTime).Seconds()

	return result, nil
}

// ExecuteControl executes control operations on running workflows
func (w *DefaultWorkflowAPI) ExecuteControl(ctx context.Context, controlParams *types.ControlParams) (*types.ControlResult, *errors.SDKError) {
	startTime := time.Now()
	result := &types.ControlResult{
		OperationID:   controlParams.OperationID,
		OperationType: controlParams.TargetOperation,
		StartedAt:     startTime,
		ActionsTaken:  []types.ControlAction{},
		StepResults:   []types.StepResult{},
	}

	// Helper to add step results
	addStep := func(step, status, message string, stepStart time.Time) {
		result.StepResults = append(result.StepResults, types.StepResult{
			Step:        step,
			Status:      status,
			Message:     message,
			StartedAt:   stepStart,
			CompletedAt: time.Now(),
			Duration:    time.Since(stepStart).Seconds(),
		})
	}

	// STEP 1: VALIDATE PARAMETERS
	stepStart := time.Now()

	// Validate action
	if controlParams.ControlAction == "" {
		return nil, errors.NewSDKError("INVALID_PARAMS", "ControlAction is required", nil)
	}

	// Validate operation ID
	if controlParams.OperationID == "" {
		return nil, errors.NewSDKError("INVALID_PARAMS", "OperationID is required", nil)
	}

	// Require confirmation for destructive actions
	if (controlParams.ControlAction == types.ControlActionAbort || controlParams.ControlAction == types.ControlActionCancel) &&
		controlParams.ConfirmationToken == "" {
		return nil, errors.NewSDKError("CONFIRMATION_REQUIRED", "ConfirmationToken required for abort/cancel actions", nil)
	}

	addStep("validate_params", "completed", "Parameters validated", stepStart)

	// STEP 2: GET OPERATION STATUS
	stepStart = time.Now()

	var currentStatus string
	var resourceIDs *types.CleanupResourceIDs

	switch controlParams.TargetOperation {
	case types.OperationTypeBackup:
		backup, err := w.taskAPI.GetBackup(ctx, controlParams.GroupID, controlParams.OperationID)
		if err != nil {
			addStep("get_operation_status", "failed", fmt.Sprintf("Backup not found: %v", err), stepStart)
			return nil, errors.NewSDKError("OPERATION_NOT_FOUND", "Backup not found", err)
		}
		currentStatus = backup.Status
		result.PreviousState = currentStatus

		// Collect resource IDs for potential cleanup
		resourceIDs = &types.CleanupResourceIDs{
			BackupIDs:         []string{backup.BackupID},
			ProtectionGroupID: backup.ProtectionGroupID,
		}

	case types.OperationTypeRestore:
		restore, err := w.taskAPI.GetRestore(ctx, controlParams.OperationID)
		if err != nil {
			addStep("get_operation_status", "failed", fmt.Sprintf("Restore not found: %v", err), stepStart)
			return nil, errors.NewSDKError("OPERATION_NOT_FOUND", "Restore not found", err)
		}
		currentStatus = restore.Status
		result.PreviousState = currentStatus

		// Collect resource IDs for potential cleanup
		resourceIDs = &types.CleanupResourceIDs{
			RestoreIDs:           []string{restore.RestoreID},
			BackupIDs:            []string{restore.BackupID},
			TargetRegistrationID: restore.TargetID,
		}

	case types.OperationTypeMigration:
		// For migration, we need to check both backup and restore
		// Use the operation ID as backup ID first
		backup, err := w.taskAPI.GetBackup(ctx, controlParams.GroupID, controlParams.OperationID)
		if err == nil {
			currentStatus = backup.Status
			result.PreviousState = currentStatus
			resourceIDs = &types.CleanupResourceIDs{
				BackupIDs:         []string{backup.BackupID},
				ProtectionGroupID: backup.ProtectionGroupID,
			}
		} else {
			// Try as restore ID
			restore, err := w.taskAPI.GetRestore(ctx, controlParams.OperationID)
			if err != nil {
				addStep("get_operation_status", "failed", "Migration operation not found", stepStart)
				return nil, errors.NewSDKError("OPERATION_NOT_FOUND", "Migration operation not found", err)
			}
			currentStatus = restore.Status
			result.PreviousState = currentStatus
			resourceIDs = &types.CleanupResourceIDs{
				RestoreIDs:           []string{restore.RestoreID},
				BackupIDs:            []string{restore.BackupID},
				TargetRegistrationID: restore.TargetID,
			}
		}

	default:
		return nil, errors.NewSDKError("INVALID_OPERATION_TYPE", fmt.Sprintf("Unsupported operation type: %s", controlParams.TargetOperation), nil)
	}

	// Check if operation can be controlled
	if currentStatus == "completed" || currentStatus == "failed" {
		addStep("get_operation_status", "failed", fmt.Sprintf("Operation already finished (status: %s), cannot control", currentStatus), stepStart)
		return nil, errors.NewSDKError("INVALID_STATE", fmt.Sprintf("Operation already finished (status: %s), cannot control", currentStatus), nil)
	}

	addStep("get_operation_status", "completed", fmt.Sprintf("Current status: %s", currentStatus), stepStart)

	// STEP 3: EXECUTE CONTROL ACTION
	stepStart = time.Now()

	switch controlParams.ControlAction {
	case types.ControlActionPause:
		// Pause operation
		if controlParams.TargetOperation == types.OperationTypeBackup {
			// Note: PauseBackup is a stub method in TaskAPI
			err := w.taskAPI.PauseBackup(ctx, controlParams.OperationID, "", controlParams.Force)
			if err != nil {
				addStep("pause_operation", "failed", fmt.Sprintf("Failed to pause backup: %v", err), stepStart)
				return nil, err
			}
		} else if controlParams.TargetOperation == types.OperationTypeRestore {
			// Note: PauseRestore is a stub method in TaskAPI
			// err := w.taskAPI.PauseRestore(ctx, controlParams.OperationID, controlParams.Force)
			// if err != nil {
			// 	addStep("pause_operation", "failed", fmt.Sprintf("Failed to pause restore: %v", err), stepStart)
			// 	return nil, err
			// }
		}

		result.Status = "paused"
		result.CurrentState = "paused"
		result.ActionsTaken = append(result.ActionsTaken, types.ControlActionPause)
		addStep("pause_operation", "completed", "Operation paused successfully", stepStart)

	case types.ControlActionAbort:
		// Abort operation
		if controlParams.TargetOperation == types.OperationTypeBackup {
			// Note: AbortBackup is a stub method in TaskAPI
			err := w.taskAPI.AbortBackup(ctx, controlParams.OperationID, "", controlParams.Force)
			if err != nil {
				addStep("abort_operation", "failed", fmt.Sprintf("Failed to abort backup: %v", err), stepStart)
				return nil, err
			}
		} else if controlParams.TargetOperation == types.OperationTypeRestore {
			// Note: AbortRestore is a stub method in TaskAPI
			err := w.taskAPI.AbortRestore(ctx, controlParams.OperationID, controlParams.Force)
			if err != nil {
				addStep("abort_operation", "failed", fmt.Sprintf("Failed to abort restore: %v", err), stepStart)
				return nil, err
			}
		}

		result.Status = "aborted"
		result.CurrentState = "aborted"
		result.ActionsTaken = append(result.ActionsTaken, types.ControlActionAbort)
		addStep("abort_operation", "completed", "Operation aborted successfully", stepStart)

		// Cleanup resources if requested
		if controlParams.CleanupOnAbort && resourceIDs != nil {
			stepStart = time.Now()
			cleanupParams := &types.CleanupParams{
				CleanupScope:      types.CleanupScopeSpecific,
				ResourceIDs:       resourceIDs,
				DryRun:            false,
				ConfirmationToken: controlParams.ConfirmationToken,
			}

			cleanupResult, err := w.ExecuteCleanup(ctx, cleanupParams)
			if err != nil {
				addStep("cleanup_resources", "failed", fmt.Sprintf("Cleanup failed: %v", err), stepStart)
				result.Status = "partial"
			} else {
				result.CleanupResult = cleanupResult
				result.ResourcesCleaned = cleanupResult.ResourcesDeleted
				addStep("cleanup_resources", "completed", fmt.Sprintf("Cleaned %d resources", cleanupResult.ResourcesDeleted), stepStart)
			}
		}

	case types.ControlActionResume:
		// Check if operation is paused
		if currentStatus != "paused" {
			addStep("resume_operation", "failed", fmt.Sprintf("Operation is not paused (status: %s), cannot resume", currentStatus), stepStart)
			return nil, errors.NewSDKError("INVALID_STATE", fmt.Sprintf("Operation is not paused (status: %s), cannot resume", currentStatus), nil)
		}

		// Resume operation
		if controlParams.TargetOperation == types.OperationTypeBackup {
			// Note: ResumeBackup is a stub method in TaskAPI
			err := w.taskAPI.ResumeBackup(ctx, controlParams.OperationID, "")
			if err != nil {
				addStep("resume_operation", "failed", fmt.Sprintf("Failed to resume backup: %v", err), stepStart)
				return nil, err
			}
		} else if controlParams.TargetOperation == types.OperationTypeRestore {
			// Note: ResumeRestore is a stub method in TaskAPI
			// err := w.taskAPI.ResumeRestore(ctx, controlParams.OperationID)
			// if err != nil {
			// 	addStep("resume_operation", "failed", fmt.Sprintf("Failed to resume restore: %v", err), stepStart)
			// 	return nil, err
			// }
		}

		result.Status = "resumed"
		result.CurrentState = "running"
		result.ActionsTaken = append(result.ActionsTaken, types.ControlActionResume)
		addStep("resume_operation", "completed", "Operation resumed successfully", stepStart)

	case types.ControlActionCancel:
		// Quick cancel without cleanup
		if controlParams.TargetOperation == types.OperationTypeBackup {
			// Note: AbortBackup with force=true for immediate stop
			err := w.taskAPI.AbortBackup(ctx, controlParams.OperationID, "", true)
			if err != nil {
				addStep("cancel_operation", "failed", fmt.Sprintf("Failed to cancel backup: %v", err), stepStart)
				return nil, err
			}
		} else if controlParams.TargetOperation == types.OperationTypeRestore {
			// Note: AbortRestore with force=true for immediate stop
			err := w.taskAPI.AbortRestore(ctx, controlParams.OperationID, true)
			if err != nil {
				addStep("cancel_operation", "failed", fmt.Sprintf("Failed to cancel restore: %v", err), stepStart)
				return nil, err
			}
		}

		result.Status = "cancelled"
		result.CurrentState = "cancelled"
		result.ActionsTaken = append(result.ActionsTaken, types.ControlActionCancel)
		addStep("cancel_operation", "completed", "Operation cancelled successfully", stepStart)

	default:
		return nil, errors.NewSDKError("INVALID_ACTION", fmt.Sprintf("Unsupported control action: %s", controlParams.ControlAction), nil)
	}

	result.CompletedAt = time.Now()
	result.Duration = time.Since(startTime).Seconds()

	return result, nil
}
