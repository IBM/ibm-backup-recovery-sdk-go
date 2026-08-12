package kubernetes

import (
	"context"
	"fmt"
	"time"

	"github.com/IBM/go-sdk-core/v5/core"
	"github.com/IBM/ibm-backup-recovery-sdk-go/backuprecoveryv1"
	"github.com/IBM/ibm-backup-recovery-sdk-go/migrationv2/common/errors"
	"github.com/IBM/ibm-backup-recovery-sdk-go/migrationv2/common/types"
)

// NamespaceSnapshotMapping represents the relationship between a Kubernetes namespace and its backup snapshot.
// This struct is used to map each namespace in a backup run to its corresponding snapshot ID,
// enabling namespace-specific restore operations with different storage class mappings per namespace.
//
// Example: In a backup run with 2 namespaces:
//   - Namespace "app-1" -> Snapshot "snap-123"
//   - Namespace "app-2" -> Snapshot "snap-456"
type NamespaceSnapshotMapping struct {
	NamespaceName string // The name of the Kubernetes namespace (e.g., "busybox-app")
	NamespaceID   int64  // The unique ID of the namespace object in BRS
	SnapshotID    string // The snapshot ID for this namespace's backup
	ObjectID      int64  // The object ID (same as NamespaceID, kept for compatibility)
}

// GetBackupRunSnapShotID fetches all snapshot IDs from a backup run.
// This is a convenience function that returns just the snapshot IDs without namespace details.
// For namespace-aware operations, use GetBackupRunNamespaceSnapshots() instead.
//
// Parameters:
//   - groupId: The protection group ID
//   - backupId: The backup run ID
//
// Returns: Array of snapshot IDs (one per namespace in the backup)
func (r *KubernetesDataSource) GetBackupRunSnapShotID(groupId, backupId string) ([]string, error) {
	mappings, err := r.GetBackupRunNamespaceSnapshots(groupId, backupId)
	if err != nil {
		return nil, err
	}

	snapshotIDs := make([]string, 0, len(mappings))
	for _, mapping := range mappings {
		snapshotIDs = append(snapshotIDs, mapping.SnapshotID)
	}
	return snapshotIDs, nil
}

// GetBackupRunNamespaceSnapshots fetches detailed namespace-to-snapshot mappings for a backup run.
// This function retrieves all namespaces that were backed up in a specific backup run,
// along with their corresponding snapshot IDs. This enables namespace-specific restore operations.
//
// Use case: When you need to apply different storage class mappings to different namespaces,
// or when you want to filter which namespaces to restore.
//
// Parameters:
//   - groupId: The protection group ID
//   - backupId: The backup run ID
//
// Returns: Array of NamespaceSnapshotMapping, one entry per namespace in the backup
//
// Example return value for a backup with 2 namespaces:
//
//	[
//	  {NamespaceName: "app-1", SnapshotID: "snap-123", NamespaceID: 72667},
//	  {NamespaceName: "app-2", SnapshotID: "snap-456", NamespaceID: 78395}
//	]
func (r *KubernetesDataSource) GetBackupRunNamespaceSnapshots(groupId, backupId string) ([]NamespaceSnapshotMapping, error) {
	getProtectionGroupRunOptions := &backuprecoveryv1.GetProtectionGroupRunOptions{
		XIBMTenantID:         core.StringPtr(r.brsClient.GetTenantId()),
		RunID:                &backupId,
		ID:                   &groupId,
		IncludeObjectDetails: core.BoolPtr(true),
	}
	backupRun, _, err := r.brsClient.GetBRSClient().GetProtectionGroupRun(getProtectionGroupRunOptions)
	if err != nil {
		return nil, err
	}
	if backupRun.Objects == nil || len(backupRun.Objects) == 0 {
		return nil, fmt.Errorf("No objects found for runId: %s under groupId: %s", backupId, groupId)
	}

	// Collect namespace-to-snapshot mappings from all objects
	mappings := make([]NamespaceSnapshotMapping, 0)
	for _, obj := range backupRun.Objects {
		if obj.ArchivalInfo != nil &&
			obj.ArchivalInfo.ArchivalTargetResults != nil &&
			len(obj.ArchivalInfo.ArchivalTargetResults) > 0 &&
			obj.ArchivalInfo.ArchivalTargetResults[0].SnapshotID != nil &&
			obj.Object != nil {

			mapping := NamespaceSnapshotMapping{
				SnapshotID: *obj.ArchivalInfo.ArchivalTargetResults[0].SnapshotID,
			}

			// Extract namespace name and ID from object
			if obj.Object.Name != nil {
				mapping.NamespaceName = *obj.Object.Name
			}
			if obj.Object.ID != nil {
				mapping.NamespaceID = *obj.Object.ID
				mapping.ObjectID = *obj.Object.ID
			}

			mappings = append(mappings, mapping)
		}
	}

	if len(mappings) == 0 {
		return nil, fmt.Errorf("No snapshotId found for runId: %s under groupId: %s", backupId, groupId)
	}

	return mappings, nil
}

// Fetch snapshot id by namespace name or namespaceId
func (k *KubernetesDataSource) GetNamespaceSnapshotId(snapShotInfo *types.SnapshotInfo) (*string, error) {
	if snapShotInfo == nil {
		return nil, errors.NewInvalidConfigError("snapshot info cannot be nil", nil)
	}

	var snapshotID *string

	if snapShotInfo.NamespaceInfo != nil {
		//get namespaceId by namesapceName
		snapshotNameSpaceInfo, err := k.ResolveNamespaceID(snapShotInfo.NamespaceInfo.Namespace)
		if err != nil {
			return nil, fmt.Errorf("failed to resolve namespace ID: %w", err)
		}
		// get snapshotId by namespaceId, GroupId and user defined BackupPositionFromFirst
		snapshotID, err = k.ResolveSnapshotID(snapshotNameSpaceInfo.GroupId, snapshotNameSpaceInfo.NamespaceId, snapShotInfo.NamespaceInfo.BackupPositionFromFirst)
		if err != nil {
			return nil, fmt.Errorf("failed to resolve snapshot ID: %w", err)
		}
	}

	// if snapshot Id already provided by user , then overwrite
	if snapShotInfo.SnapshotID != "" {
		snapshotID = core.StringPtr(snapShotInfo.SnapshotID)
	}

	if snapshotID == nil {
		return nil, fmt.Errorf("either NamespaceInfo or SnapshotId must be provided")
	}
	return snapshotID, nil
}

// create RecoverNamespace params
func (k *KubernetesDataSource) RecoverNamespace(ctx context.Context, groupId, backupId string, targetRegistrationID int64, k8sRestoreParms *types.KubernetesRestoreParams, commonRestoreParams *types.RestoreParams) (*backuprecoveryv1.RecoverKubernetesParamsRecoverNamespaceParams, error) {
	if k8sRestoreParms == nil {
		return nil, errors.NewInvalidConfigError("kubernetes restore params cannot be nil", nil)
	}
	if commonRestoreParams == nil {
		return nil, errors.NewInvalidConfigError("common restore params cannot be nil", nil)
	}

	// Initialize Kubernetes Namespace Target Params
	recoverTargetParams := &backuprecoveryv1.RecoverKubernetesNamespaceParamsKubernetesTargetParams{
		SkipClusterCompatibilityCheck: core.BoolPtr(k8sRestoreParms.SkipClusterCompatibilityCheck),
	}

	// check rename namespace params
	if k8sRestoreParms.RenameRecoveredNamespacesParams != nil {
		recoverTargetParams.RenameRecoveredNamespacesParams = &backuprecoveryv1.KubernetesTargetParamsForRecoverKubernetesNamespaceRenameRecoveredNamespacesParams{
			Prefix: core.StringPtr(k8sRestoreParms.RenameRecoveredNamespacesParams.Prefix),
			Suffix: core.StringPtr(k8sRestoreParms.RenameRecoveredNamespacesParams.Suffix),
		}
	}
	// check RegionMapping params
	if k8sRestoreParms.RegionMapping != nil {
		recoverTargetParams.RecoveryRegionMigrationParams = &backuprecoveryv1.KubernetesRecoveryMigrationParams{}
		if k8sRestoreParms.RegionMapping.Source != k8sRestoreParms.RegionMapping.Target {
			recoverTargetParams.RecoveryRegionMigrationParams.CurrentValue = core.StringPtr(k8sRestoreParms.RegionMapping.Source)
			recoverTargetParams.RecoveryRegionMigrationParams.NewValue = core.StringPtr(k8sRestoreParms.RegionMapping.Target)
		} else {
			k.logger.Info(ctx, "Skipping region migration - source and target regions are identical",
				"source", k8sRestoreParms.RegionMapping.Source,
				"target", k8sRestoreParms.RegionMapping.Target)
		}
	}

	// check ZoneMappings params
	if k8sRestoreParms.ZoneMappings != nil {
		recoverTargetParams.RecoveryZoneMigrationParams = k.MapRecoveryMigrationParams(k8sRestoreParms.ZoneMappings)
	}

	// check recover to new target
	targetConfig := &backuprecoveryv1.KubernetesTargetParamsForRecoverKubernetesNamespaceRecoveryTargetConfig{
		RecoverToNewSource: core.BoolPtr(k8sRestoreParms.RecoverToNewTarget),
	}

	// Only provide NewSourceConfig for cross-cluster restores.
	if k8sRestoreParms.RecoverToNewTarget {
		targetConfig.NewSourceConfig = &backuprecoveryv1.KubernetesNamespaceRecoveryTargetConfigNewSourceConfig{
			Source: &backuprecoveryv1.KubernetesNamespaceRecoveryNewSourceConfigSource{
				ID: core.Int64Ptr(targetRegistrationID),
			},
		}
	}
	recoverTargetParams.RecoveryTargetConfig = targetConfig

	// Initialize Kubernetes Recover Namespace Params
	kubernetesRecoveryObjectParams, err := k.InitializeKubernetesNamespaceParams(groupId, backupId, k8sRestoreParms)
	if err != nil {
		return nil, fmt.Errorf("Failed to initialize recovery namespace object params: %w", err)
	}
	recoverTargetParams.Objects = kubernetesRecoveryObjectParams

	return &backuprecoveryv1.RecoverKubernetesParamsRecoverNamespaceParams{
		TargetEnvironment:      core.StringPtr(backuprecoveryv1.RecoverKubernetesParamsRecoverNamespaceParams_TargetEnvironment_Kkubernetes),
		KubernetesTargetParams: recoverTargetParams,
	}, nil

}

// InitializeKubernetesNamespaceParams creates recovery parameters for Kubernetes namespaces.
// This function implements the core logic for namespace-level storage class mapping with precedence.
//
// How it works:
// 1. Fetches all namespaces from the backup run (with their snapshot IDs)
// 2. Applies namespace filtering if IncludeNamespaces is specified
// 3. For each namespace, determines which storage class mapping to use:
//   - Namespace-level mapping (from RecoverMultipleObjects) takes HIGHEST PRIORITY
//   - Top-level mapping (from RecoverObjectSpec) is used as DEFAULT
//
// Example scenario:
//
//	Top-level spec says: "class-A" -> "class-B" (applies to all namespaces)
//	Namespace "app-1" override says: "class-A" -> "class-C" (only for app-1)
//	Result: app-1 uses class-C, all other namespaces use class-B
//
// Parameters:
//   - groupId: Protection group ID
//   - backupId: Backup run ID
//   - k8sRestoreParms: Restore parameters including namespace-specific overrides
//
// Returns: Array of recovery object parameters, one per namespace to be restored
func (k *KubernetesDataSource) InitializeKubernetesNamespaceParams(groupId, backupId string, k8sRestoreParms *types.KubernetesRestoreParams) ([]backuprecoveryv1.KubernetesRecoveryObjectParams, error) {
	// Initialize kubernetes Recovery Object Params
	kubernetesRecoveryObjectParams := make([]backuprecoveryv1.KubernetesRecoveryObjectParams, 0)

	// Step 1: Fetch namespace-to-snapshot mappings for the backup run
	// This gives us all namespaces that were backed up, with their snapshot IDs
	namespaceMappings, err := k.GetBackupRunNamespaceSnapshots(groupId, backupId)
	if err != nil {
		return nil, err
	}

	if k8sRestoreParms.RecoverObjectSpec == nil {
		return nil, fmt.Errorf("RecoverObjectSpec is a required field")
	}

	// Step 2: Build a map of namespace-specific recovery specs
	// Users provide namespace NAMES (not snapshot IDs) in RecoverMultipleObjects array
	// Example: {"namespace": "app-1", "storageClasses": [...]}
	namespaceSpecificSpecs := make(map[string]*types.RecoverObjectSpec)
	for _, restoreObject := range k8sRestoreParms.RecoverMultipleObjects {
		if restoreObject.SnapshotInfo != nil && restoreObject.SnapshotInfo.NamespaceInfo != nil {
			namespaceName := restoreObject.SnapshotInfo.NamespaceInfo.Namespace
			namespaceSpecificSpecs[namespaceName] = restoreObject.RecoverObjectSpec
		}
	}

	// Step 3: Create a filter set for namespace inclusion (if specified)
	// If IncludeNamespaces is empty, all namespaces are included
	// If IncludeNamespaces has values, only those namespaces are restored
	includeNamespaceSet := make(map[string]bool)
	if len(k8sRestoreParms.IncludeNamespaces) > 0 {
		for _, ns := range k8sRestoreParms.IncludeNamespaces {
			includeNamespaceSet[ns] = true
		}
	}

	// Track which namespaces we've already processed
	processedNamespaces := make(map[string]bool)

	// Step 4: Process each namespace from the backup run
	for _, mapping := range namespaceMappings {
		// Apply namespace filter if specified
		if len(includeNamespaceSet) > 0 && !includeNamespaceSet[mapping.NamespaceName] {
			continue // Skip - this namespace is not in the include list
		}

		// Step 5: Determine which RecoverObjectSpec to use (PRECEDENCE LOGIC)
		var recoverSpec *types.RecoverObjectSpec

		// Check if there's a namespace-specific spec (HIGHEST PRIORITY)
		if spec, exists := namespaceSpecificSpecs[mapping.NamespaceName]; exists && spec != nil {
			recoverSpec = spec // Use namespace-level override
		} else {
			recoverSpec = k8sRestoreParms.RecoverObjectSpec // Use top-level default
		}

		// Build recovery object with the determined spec
		recoveryObject := k.buildKubernetesRecoveryObject(mapping.SnapshotID, recoverSpec)
		kubernetesRecoveryObjectParams = append(kubernetesRecoveryObjectParams, *recoveryObject)

		// Mark this namespace as processed
		processedNamespaces[mapping.NamespaceName] = true
	}

	// Step 6: Process any additional namespaces from RecoverMultipleObjects
	// that weren't part of the current backup run (e.g., from different backup runs)
	for _, restoreObject := range k8sRestoreParms.RecoverMultipleObjects {
		if restoreObject.SnapshotInfo == nil {
			continue
		}

		// Check if this namespace was already processed from the backup run
		if restoreObject.SnapshotInfo.NamespaceInfo != nil {
			namespaceName := restoreObject.SnapshotInfo.NamespaceInfo.Namespace
			if processedNamespaces[namespaceName] {
				continue // Already processed from backup run
			}
		}

		// Resolve snapshot ID for this namespace
		snapshotId, err := k.GetNamespaceSnapshotId(restoreObject.SnapshotInfo)
		if err != nil {
			return nil, fmt.Errorf("failed to get recovery snapshot ID for namespace: %w", err)
		}

		recoveryObject := k.buildKubernetesRecoveryObject(*snapshotId, restoreObject.RecoverObjectSpec)
		kubernetesRecoveryObjectParams = append(kubernetesRecoveryObjectParams, *recoveryObject)
	}

	return kubernetesRecoveryObjectParams, nil
}

func (k *KubernetesDataSource) buildKubernetesRecoveryObject(snapshotId string, recoverObjectSpec *types.RecoverObjectSpec) *backuprecoveryv1.KubernetesRecoveryObjectParams {
	recoveryObject := &backuprecoveryv1.KubernetesRecoveryObjectParams{
		SnapshotID: &snapshotId,
	}
	if recoverObjectSpec != nil {
		recoveryObject.RecoverPvcsOnly = core.BoolPtr(recoverObjectSpec.RestoreOnlyPvc)
		recoveryObject.StorageClass = k.MapRecoveryStorageClasses(recoverObjectSpec.StorageClasses)
		recoveryObject.IncludeParams = k.MapRecoveryResourceParams(recoverObjectSpec.IncludeObjects)
		recoveryObject.ExcludeParams = k.MapRecoveryResourceParams(recoverObjectSpec.ExcludeObjects)
	}
	return recoveryObject
}

func (r *KubernetesDataSource) MapRecoveryMigrationParams(migrationMapParams []types.MigrationMapParams) []backuprecoveryv1.KubernetesRecoveryMigrationParams {

	recoveryMigrationParams := make([]backuprecoveryv1.KubernetesRecoveryMigrationParams, 0)
	for _, param := range migrationMapParams {
		recoveryMigrationParams = append(recoveryMigrationParams, backuprecoveryv1.KubernetesRecoveryMigrationParams{
			CurrentValue: core.StringPtr(param.Source),
			NewValue:     core.StringPtr(param.Target),
		})
	}
	return recoveryMigrationParams
}

func (k *KubernetesDataSource) MapRecoveryResourceParams(k8sObjects *types.K8sObject) *backuprecoveryv1.KubernetesFilterParams {
	filterParams := &backuprecoveryv1.KubernetesFilterParams{}

	// Nil check before accessing k8sObjects
	if k8sObjects == nil {
		return filterParams
	}

	selectedResources := make([]backuprecoveryv1.ResourceInfo, 0)
	selectedLabels := make([]backuprecoveryv1.KubernetesLabel, 0)

	if k8sObjects.SelectedResources != nil {
		for _, resource := range k8sObjects.SelectedResources {
			selectedResources = append(selectedResources, backuprecoveryv1.ResourceInfo{
				Kind:         core.StringPtr(resource.ResourceKind),
				ApiGroup:     core.StringPtr(resource.ResourceApiGroup),
				ResourceList: k.MapRecoveryResourceInfo(resource.ResourceList),
			})
		}
		filterParams.SelectedResources = selectedResources
	}

	if k8sObjects.SelectedLabels != nil {
		for _, label := range k8sObjects.SelectedLabels.Labels {
			selectedLabels = append(selectedLabels, backuprecoveryv1.KubernetesLabel{
				Key:   core.StringPtr(label.Key),
				Value: core.StringPtr(label.Value),
			})
		}
		filterParams.LabelVector = selectedLabels
		filterParams.LabelCombinationMethod = &k8sObjects.SelectedLabels.LabelCombination
	}

	return filterParams
}

func (k *KubernetesDataSource) MapRecoveryResourceInfo(resourceList []types.ResourceInstance) []backuprecoveryv1.ResourceInstance {
	resourceInstanceList := make([]backuprecoveryv1.ResourceInstance, 0)
	for _, res := range resourceList {
		resourceInstanceList = append(resourceInstanceList, backuprecoveryv1.ResourceInstance{
			EntityID: core.Int64Ptr(res.Id),
			Name:     core.StringPtr(res.Name),
		})
	}
	return resourceInstanceList
}

func (r *KubernetesDataSource) MapRecoveryStorageClasses(storageClassMappings []types.StorageClassMapping) *backuprecoveryv1.KubernetesStorageClassParams {

	storageClassParams := &backuprecoveryv1.KubernetesStorageClassParams{}
	mappings := make([]backuprecoveryv1.KubernetesLabel, 0)

	for _, v := range storageClassMappings {
		mappings = append(mappings, backuprecoveryv1.KubernetesLabel{
			Key:   core.StringPtr(v.Old),
			Value: core.StringPtr(v.New),
		})
	}

	if len(mappings) > 0 {
		storageClassParams.UseStorageClassMapping = core.BoolPtr(true)
		storageClassParams.StorageClassMapping = mappings
	}

	return storageClassParams
}

func (k *KubernetesDataSource) ResolveNamespaceID(namespace string) (*types.ResolveNamespaceResult, error) {
	if namespace == "" {
		return nil, errors.NewInvalidConfigError("namespace cannot be empty", nil)
	}

	protectedObjectOptions := &backuprecoveryv1.SearchProtectedObjectsOptions{
		XIBMTenantID:    core.StringPtr(k.brsClient.GetTenantId()),
		SearchString:    &namespace,
		SnapshotActions: []string{backuprecoveryv1.GetObjectSnapshotsOptions_SnapshotActions_Recovernamespaces},
	}

	objects, _, err := k.brsClient.GetBRSClient().SearchProtectedObjects(protectedObjectOptions)
	if err != nil {
		return nil, fmt.Errorf("failed to find the protected namespace: %w", err)
	}
	// Check len() before slice index access
	if objects == nil || objects.Objects == nil || len(objects.Objects) == 0 {
		return nil, fmt.Errorf("no protected namespace found for: %s", namespace)
	}

	// Nil checks before pointer dereferences
	if objects.Objects[0].ID == nil {
		return nil, fmt.Errorf("namespace ID is nil for namespace: %s", namespace)
	}
	if objects.Objects[0].SourceID == nil {
		return nil, fmt.Errorf("source ID is nil for namespace: %s", namespace)
	}
	if len(objects.Objects[0].LatestSnapshotsInfo) == 0 {
		return nil, fmt.Errorf("no snapshot info found for namespace: %s", namespace)
	}
	if objects.Objects[0].LatestSnapshotsInfo[0].ProtectionGroupID == nil {
		return nil, fmt.Errorf("protection group ID is nil for namespace: %s", namespace)
	}

	return &types.ResolveNamespaceResult{
		NamespaceId: *objects.Objects[0].ID,
		SourceId:    *objects.Objects[0].SourceID,
		GroupId:     *objects.Objects[0].LatestSnapshotsInfo[0].ProtectionGroupID,
	}, nil
}

func (k *KubernetesDataSource) ResolveSnapshotID(protectionGroupId string, nsId int64, backupPositionFromFirst int64) (*string, error) {
	if protectionGroupId == "" {
		return nil, errors.NewInvalidConfigError("protection group ID cannot be empty", nil)
	}

	snapshotID, err := k.getSnapshotID(protectionGroupId, nsId, backupPositionFromFirst)
	if err != nil {
		return nil, fmt.Errorf("unable to get snapshot ID: %w", err)
	}

	return snapshotID, nil
}

func (k *KubernetesDataSource) waitForSnapshotID(protectionGroupId string, namespaceObjectID, BackupPositionFromFirst int64, totalTimeout, pollingInterval time.Duration) (*string, error) {
	// var totalTimeout time.Duration = types.DefaultSnapahotWait_TotalTimeout       // 10 minutes total wait time
	// var pollingInterval time.Duration = types.DefaultSnapahotWait_PollingInterval // Poll every 30 seconds
	// Calculate deadline
	deadline := time.Now().Add(totalTimeout)

	// Polling loop
	for time.Now().Before(deadline) {
		opts := &backuprecoveryv1.GetObjectSnapshotsOptions{
			XIBMTenantID:       core.StringPtr(k.brsClient.GetTenantId()),
			ID:                 core.Int64Ptr(namespaceObjectID),
			ProtectionGroupIds: []string{protectionGroupId},
		}

		result, _, err := k.brsClient.GetBRSClient().GetObjectSnapshots(opts)
		if err != nil {
			// Log the error but continue polling
			fmt.Printf("Error fetching snapshots: %v. Retrying...\n", err)
			time.Sleep(pollingInterval)
			continue
		}

		// Check if we have snapshots
		if result != nil && result.Snapshots != nil && len(result.Snapshots) > 0 && len(result.Snapshots) >= int(BackupPositionFromFirst) {
			var snap backuprecoveryv1.ObjectSnapshot

			// Get the snapshot at the specified position
			if BackupPositionFromFirst > 0 {
				snap = result.Snapshots[BackupPositionFromFirst-1]
			} else {
				snap = result.Snapshots[BackupPositionFromFirst]
			}

			// Check if snapshot ID exists
			if snap.ID != nil {
				// Success! Return the snapshot ID
				return snap.ID, nil
			}
		}

		// Wait before next poll
		// fmt.Printf("Snapshot ID not found yet. Waiting %v before next attempt...\n", pollingInterval)
		time.Sleep(pollingInterval)
	}

	// Timeout reached without finding snapshot ID
	return nil, fmt.Errorf("snapshot ID not found within timeout of %v", totalTimeout)
}

func (k *KubernetesDataSource) getSnapshotID(protectionGroupId string, namespaceObjectID, BackupPositionFromFirst int64) (*string, error) {
	opts := &backuprecoveryv1.GetObjectSnapshotsOptions{
		XIBMTenantID:       core.StringPtr(k.brsClient.GetTenantId()),
		ID:                 core.Int64Ptr(namespaceObjectID),
		ProtectionGroupIds: []string{protectionGroupId},
	}
	result, _, err := k.brsClient.GetBRSClient().GetObjectSnapshots(opts)
	if err != nil {
		// Log the error but continue polling
		return nil, fmt.Errorf("Error fetching snapshot id: %v. ", err)
	}

	// Check if we have snapshots
	if result != nil && result.Snapshots != nil && len(result.Snapshots) > 0 && len(result.Snapshots) >= int(BackupPositionFromFirst) {
		var snap backuprecoveryv1.ObjectSnapshot

		// Get the snapshot at the specified position
		if BackupPositionFromFirst > 0 {
			snap = result.Snapshots[BackupPositionFromFirst-1]
		} else {
			snap = result.Snapshots[BackupPositionFromFirst]
		}

		// Check if snapshot ID exists
		if snap.ID != nil {
			// Success! Return the snapshot ID
			return snap.ID, nil
		}
	}
	// Timeout reached without finding snapshot ID
	return nil, fmt.Errorf("No snapshot ID was found")
}
