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

// Fetch snapshot id by backupId and  groupId
func (r *KubernetesDataSource) GetBackupRunSnapShotID(groupId, backupId string) (*string, error) {
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
	if backupRun.Objects == nil ||
		len(backupRun.Objects) == 0 ||
		backupRun.Objects[0].ArchivalInfo == nil ||
		backupRun.Objects[0].ArchivalInfo.ArchivalTargetResults == nil ||
		len(backupRun.Objects[0].ArchivalInfo.ArchivalTargetResults) == 0 ||
		backupRun.Objects[0].ArchivalInfo.ArchivalTargetResults[0].SnapshotID == nil {
		return nil, fmt.Errorf("No snapshotId found for runId: %s under groupId: %s", backupId, groupId)
	}
	return backupRun.Objects[0].ArchivalInfo.ArchivalTargetResults[0].SnapshotID, nil
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

func (k *KubernetesDataSource) InitializeKubernetesNamespaceParams(groupId, backupId string, k8sRestoreParms *types.KubernetesRestoreParams) ([]backuprecoveryv1.KubernetesRecoveryObjectParams, error) {
	// Initialize kubernetes Recovery Object Params
	kubernetesRecoveryObjectParams := make([]backuprecoveryv1.KubernetesRecoveryObjectParams, 0)

	// fetch snapshot ID corresponding to user provided `backupId` and `groupId`
	snapshotID, err := k.GetBackupRunSnapShotID(groupId, backupId)
	if err != nil {
		return nil, err
	}

	if k8sRestoreParms.RecoverObjectSpec == nil {
		return nil, fmt.Errorf("RecoverObjectSpec is a required field")
	}

	// return namespace recovery params when snapshotId corresponding to user provided `backupId` and `groupId` is found
	if snapshotID != nil {
		recoveryObject := k.buildKubernetesRecoveryObject(*snapshotID, k8sRestoreParms.RecoverObjectSpec)
		kubernetesRecoveryObjectParams = append(kubernetesRecoveryObjectParams, *recoveryObject)

	}

	// if multiple additional objects need to be recovered, fetch snapshot Id for each using namespace details and return recovery params for all these snapshot IDs
	for _, restoreObject := range k8sRestoreParms.RecoverMultipleObjects {
		snapshotId, err := k.GetNamespaceSnapshotId(restoreObject.SnapshotInfo)
		if err != nil {
			return nil, fmt.Errorf("failed to get recovery snapshot ID: %w", err)
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
