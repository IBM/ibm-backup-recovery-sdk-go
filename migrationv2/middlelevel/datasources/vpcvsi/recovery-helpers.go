/*******************************************************************************
 * IBM Confidential
 * OCO Source Materials
 * (C) Copyright IBM Corp. 2026 All Rights Reserved.
 * The source code for this program is not  published or otherwise divested of
 * its trade secrets, irrespective of what has been deposited with
 * the U.S. Copyright Office.
 ******************************************************************************/
package vpcvsi

import (
	"fmt"

	"github.com/IBM/go-sdk-core/v5/core"
	"github.com/IBM/ibm-backup-recovery-sdk-go/backuprecoveryv1"
	"github.com/IBM/ibm-backup-recovery-sdk-go/migrationv2/common/errors"
	"github.com/IBM/ibm-backup-recovery-sdk-go/migrationv2/common/types"
)

// GetBackupRunSnapshotID fetches snapshot ID by backupId and groupId
// This is similar to the Kubernetes implementation
func (v *VPCVSIDataSource) GetBackupRunSnapshotID(groupId, backupId string) (*string, error) {
	getProtectionGroupRunOptions := &backuprecoveryv1.GetProtectionGroupRunOptions{
		XIBMTenantID:         core.StringPtr(v.brsClient.GetTenantId()),
		RunID:                &backupId,
		ID:                   &groupId,
		IncludeObjectDetails: core.BoolPtr(true),
	}

	backupRun, _, err := v.brsClient.GetBRSClient().GetProtectionGroupRun(getProtectionGroupRunOptions)
	if err != nil {
		return nil, fmt.Errorf("failed to get protection group run: %w", err)
	}

	if backupRun.Objects == nil ||
		len(backupRun.Objects) == 0 ||
		backupRun.Objects[0].ArchivalInfo == nil ||
		backupRun.Objects[0].ArchivalInfo.ArchivalTargetResults == nil ||
		len(backupRun.Objects[0].ArchivalInfo.ArchivalTargetResults) == 0 ||
		backupRun.Objects[0].ArchivalInfo.ArchivalTargetResults[0].SnapshotID == nil {
		return nil, fmt.Errorf("no snapshotId found for runId: %s under groupId: %s", backupId, groupId)
	}

	return backupRun.Objects[0].ArchivalInfo.ArchivalTargetResults[0].SnapshotID, nil
}

// RecoverPhysicalVolumes creates parameters for physical volume recovery
func (v *VPCVSIDataSource) RecoverPhysicalVolumes(
	groupId, backupId string,
	targetRegistrationID int64,
	vsiRestoreParams *types.VsiVpcRestoreParams,
	commonRestoreParams *types.RestoreParams,

) (*backuprecoveryv1.RecoverPhysicalParams, error) {
	if vsiRestoreParams == nil {
		return nil, errors.NewInvalidConfigError("VSI restore params cannot be nil", nil)
	}
	if commonRestoreParams == nil {
		return nil, errors.NewInvalidConfigError("common restore params cannot be nil", nil)
	}
	if vsiRestoreParams.RecoverVolumeParams == nil {
		return nil, errors.NewInvalidConfigError("recover volume params are required for volume recovery", nil)
	}

	// Get snapshot ID from backup run
	snapshotID, err := v.GetBackupRunSnapshotID(groupId, backupId)
	if err != nil {
		return nil, fmt.Errorf("failed to get snapshot ID: %w", err)
	}

	// Build recovery objects
	recoveryObjects := []backuprecoveryv1.CommonRecoverObjectSnapshotParams{
		{
			SnapshotID: snapshotID,
		},
	}

	// Build volume mapping
	volumeMapping := make([]backuprecoveryv1.RecoverVolumeMapping, 0)
	for _, mapping := range vsiRestoreParams.RecoverVolumeParams.VolumeMapping {
		volumeMapping = append(volumeMapping, backuprecoveryv1.RecoverVolumeMapping{
			SourceVolumeGuid:      core.StringPtr(mapping.SourceVolumeGUID),
			DestinationVolumeGuid: core.StringPtr(mapping.DestinationVolumeGUID),
		})
	}

	// Build physical target params for volume recovery
	mountTarget := &backuprecoveryv1.PhysicalTargetParamsForRecoverVolumeMountTarget{
		ID: core.Int64Ptr(vsiRestoreParams.RecoverVolumeParams.MountTargetID),
	}

	physicalTargetParams := &backuprecoveryv1.RecoverPhysicalVolumeParamsPhysicalTargetParams{
		MountTarget:        mountTarget,
		VolumeMapping:      volumeMapping,
		ForceUnmountVolume: core.BoolPtr(vsiRestoreParams.RecoverVolumeParams.ForceUnmountVolume),
	}

	// Add VLAN config if provided
	if vsiRestoreParams.RecoverVolumeParams.VlanConfigID != nil {
		physicalTargetParams.VlanConfig = &backuprecoveryv1.PhysicalTargetParamsForRecoverVolumeVlanConfig{
			ID: vsiRestoreParams.RecoverVolumeParams.VlanConfigID,
		}
	}

	// Build recover volume params
	recoverVolumeParams := &backuprecoveryv1.RecoverPhysicalParamsRecoverVolumeParams{
		TargetEnvironment:    core.StringPtr(backuprecoveryv1.RecoverPhysicalParamsRecoverVolumeParams_TargetEnvironment_Kphysical),
		PhysicalTargetParams: physicalTargetParams,
	}

	// Build and return RecoverPhysicalParams
	recoveryAction := backuprecoveryv1.RecoverPhysicalParams_RecoveryAction_Recoverphysicalvolumes
	recoverPhysicalParams := &backuprecoveryv1.RecoverPhysicalParams{
		Objects:             recoveryObjects,
		RecoveryAction:      core.StringPtr(recoveryAction),
		RecoverVolumeParams: recoverVolumeParams,
	}

	return recoverPhysicalParams, nil
}

// RecoverPhysicalFiles creates parameters for file and folder recovery
func (v *VPCVSIDataSource) RecoverPhysicalFiles(
	groupId, backupId string,
	targetRegistrationID int64,
	vsiRestoreParams *types.VsiVpcRestoreParams,
	commonRestoreParams *types.RestoreParams,

) (*backuprecoveryv1.RecoverPhysicalParams, error) {
	if vsiRestoreParams == nil {
		return nil, errors.NewInvalidConfigError("VSI restore params cannot be nil", nil)
	}
	if commonRestoreParams == nil {
		return nil, errors.NewInvalidConfigError("common restore params cannot be nil", nil)
	}
	if vsiRestoreParams.RecoverFileAndFolderParams == nil {
		return nil, errors.NewInvalidConfigError("recover file and folder params are required for file recovery", nil)
	}

	// Get snapshot ID from backup run
	snapshotID, err := v.GetBackupRunSnapshotID(groupId, backupId)
	if err != nil {
		return nil, fmt.Errorf("failed to get snapshot ID: %w", err)
	}

	// Build files and folders list
	filesAndFolders := make([]backuprecoveryv1.CommonRecoverFileAndFolderInfo, 0)
	for _, file := range vsiRestoreParams.RecoverFileAndFolderParams.FilesAndFolders {
		filesAndFolders = append(filesAndFolders, backuprecoveryv1.CommonRecoverFileAndFolderInfo{
			AbsolutePath:       core.StringPtr(file.AbsolutePath),
			IsDirectory:        core.BoolPtr(file.IsDirectory),
			IsViewFileRecovery: core.BoolPtr(file.IsViewFileRecovery),
		})
	}

	// Build physical target params for file recovery
	recoverTarget := &backuprecoveryv1.PhysicalTargetParamsForRecoverFileAndFolderRecoverTarget{
		ID: core.Int64Ptr(targetRegistrationID),
	}

	physicalTargetParams := &backuprecoveryv1.RecoverPhysicalFileAndFolderParamsPhysicalTargetParams{
		RecoverTarget:             recoverTarget,
		RestoreToOriginalPaths:    core.BoolPtr(vsiRestoreParams.RecoverFileAndFolderParams.RestoreToOriginalPaths),
		OverwriteExisting:         core.BoolPtr(vsiRestoreParams.RecoverFileAndFolderParams.OverwriteExisting),
		AlternateRestoreDirectory: core.StringPtr(vsiRestoreParams.RecoverFileAndFolderParams.AlternateRestoreDirectory),
		PreserveAttributes:        core.BoolPtr(vsiRestoreParams.RecoverFileAndFolderParams.PreserveAttributes),
		PreserveTimestamps:        core.BoolPtr(vsiRestoreParams.RecoverFileAndFolderParams.PreserveTimestamps),
		PreserveAcls:              core.BoolPtr(vsiRestoreParams.RecoverFileAndFolderParams.PreserveACLs),
		SaveSuccessFiles:          core.BoolPtr(vsiRestoreParams.RecoverFileAndFolderParams.SaveSuccessFiles),
		RestoreEntityType:         core.StringPtr(vsiRestoreParams.RecoverFileAndFolderParams.RestoreEntityType),
		ContinueOnError:           core.BoolPtr(vsiRestoreParams.RecoverFileAndFolderParams.ContinueOnError),
	}

	// Add VLAN config if provided
	if vsiRestoreParams.RecoverFileAndFolderParams.VlanConfigID != nil {
		physicalTargetParams.VlanConfig = &backuprecoveryv1.PhysicalTargetParamsForRecoverFileAndFolderVlanConfig{
			ID: vsiRestoreParams.RecoverFileAndFolderParams.VlanConfigID,
		}
	}

	// Build recover file and folder params
	recoverFileAndFolderParams := &backuprecoveryv1.RecoverPhysicalParamsRecoverFileAndFolderParams{
		FilesAndFolders:      filesAndFolders,
		TargetEnvironment:    core.StringPtr(backuprecoveryv1.RecoverPhysicalParamsRecoverFileAndFolderParams_TargetEnvironment_Kphysical),
		PhysicalTargetParams: physicalTargetParams,
	}

	// Build recovery objects
	recoveryObjects := []backuprecoveryv1.CommonRecoverObjectSnapshotParams{
		{
			SnapshotID: snapshotID,
		},
	}

	// Build and return RecoverPhysicalParams
	recoveryAction := backuprecoveryv1.RecoverPhysicalParams_RecoveryAction_Recoverfiles
	recoverPhysicalParams := &backuprecoveryv1.RecoverPhysicalParams{
		Objects:                    recoveryObjects,
		RecoveryAction:             core.StringPtr(recoveryAction),
		RecoverFileAndFolderParams: recoverFileAndFolderParams,
	}

	return recoverPhysicalParams, nil
}

// MountPhysicalVolumes creates parameters for mounting physical volumes
func (v *VPCVSIDataSource) MountPhysicalVolumes(
	groupId, backupId string,
	targetRegistrationID int64,
	vsiRestoreParams *types.VsiVpcRestoreParams,
	commonRestoreParams *types.RestoreParams,

) (*backuprecoveryv1.RecoverPhysicalParams, error) {
	if vsiRestoreParams == nil {
		return nil, errors.NewInvalidConfigError("VSI restore params cannot be nil", nil)
	}
	if commonRestoreParams == nil {
		return nil, errors.NewInvalidConfigError("common restore params cannot be nil", nil)
	}
	if vsiRestoreParams.MountVolumeParams == nil {
		return nil, errors.NewInvalidConfigError("mount volume params are required for volume mounting", nil)
	}

	// Get snapshot ID from backup run
	snapshotID, err := v.GetBackupRunSnapshotID(groupId, backupId)
	if err != nil {
		return nil, fmt.Errorf("failed to get snapshot ID: %w", err)
	}

	// Build physical target params for mount
	physicalTargetParams := &backuprecoveryv1.MountPhysicalVolumeParamsPhysicalTargetParams{
		MountToOriginalTarget: core.BoolPtr(vsiRestoreParams.MountVolumeParams.MountToOriginalTarget),
		ReadOnlyMount:         core.BoolPtr(vsiRestoreParams.MountVolumeParams.ReadOnlyMount),
		VolumeNames:           vsiRestoreParams.MountVolumeParams.VolumeNames,
	}

	// Add VLAN config if provided
	if vsiRestoreParams.MountVolumeParams.VlanConfigID != nil {
		physicalTargetParams.VlanConfig = &backuprecoveryv1.PhysicalTargetParamsForMountVolumeVlanConfig{
			ID: vsiRestoreParams.MountVolumeParams.VlanConfigID,
		}
	}

	// Build mount volume params
	mountVolumeParams := &backuprecoveryv1.RecoverPhysicalParamsMountVolumeParams{
		TargetEnvironment:    core.StringPtr(backuprecoveryv1.RecoverPhysicalParamsMountVolumeParams_TargetEnvironment_Kphysical),
		PhysicalTargetParams: physicalTargetParams,
	}

	// Build recovery objects
	recoveryObjects := []backuprecoveryv1.CommonRecoverObjectSnapshotParams{
		{
			SnapshotID: snapshotID,
		},
	}

	// Build and return RecoverPhysicalParams
	recoveryAction := backuprecoveryv1.RecoverPhysicalParams_RecoveryAction_Instantvolumemount
	recoverPhysicalParams := &backuprecoveryv1.RecoverPhysicalParams{
		Objects:           recoveryObjects,
		RecoveryAction:    core.StringPtr(recoveryAction),
		MountVolumeParams: mountVolumeParams,
	}

	return recoverPhysicalParams, nil
}

// RecoverPhysicalSystem creates parameters for system recovery
func (v *VPCVSIDataSource) RecoverPhysicalSystem(
	groupId, backupId string,
	targetRegistrationID int64,
	vsiRestoreParams *types.VsiVpcRestoreParams,
	commonRestoreParams *types.RestoreParams,

) (*backuprecoveryv1.RecoverPhysicalParams, error) {
	if vsiRestoreParams == nil {
		return nil, errors.NewInvalidConfigError("VSI restore params cannot be nil", nil)
	}
	if commonRestoreParams == nil {
		return nil, errors.NewInvalidConfigError("common restore params cannot be nil", nil)
	}
	if vsiRestoreParams.SystemRecoveryParams == nil {
		return nil, errors.NewInvalidConfigError("system recovery params are required for system recovery", nil)
	}

	// Get snapshot ID from backup run
	snapshotID, err := v.GetBackupRunSnapshotID(groupId, backupId)
	if err != nil {
		return nil, fmt.Errorf("failed to get snapshot ID: %w", err)
	}

	// Build system recovery params
	systemRecoveryParams := &backuprecoveryv1.RecoverPhysicalParamsSystemRecoveryParams{
		FullNasPath: core.StringPtr(vsiRestoreParams.SystemRecoveryParams.FullNasPath),
	}

	// Build recovery objects
	recoveryObjects := []backuprecoveryv1.CommonRecoverObjectSnapshotParams{
		{
			SnapshotID: snapshotID,
		},
	}

	// Build and return RecoverPhysicalParams
	recoveryAction := backuprecoveryv1.RecoverPhysicalParams_RecoveryAction_Recoversystem
	recoverPhysicalParams := &backuprecoveryv1.RecoverPhysicalParams{
		Objects:              recoveryObjects,
		RecoveryAction:       core.StringPtr(recoveryAction),
		SystemRecoveryParams: systemRecoveryParams,
	}

	return recoverPhysicalParams, nil
}
