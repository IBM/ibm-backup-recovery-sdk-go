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
	"strings"

	"github.com/IBM/go-sdk-core/v5/core"
	"github.com/IBM/ibm-backup-recovery-sdk-go/backuprecoveryv1"
)

// GetPhysicalProtectionGroupParams builds the PhysicalProtectionGroupParams for VPC VSI
// This is analogous to GetK8Object for Kubernetes
func (v *VPCVSIDataSource) GetPhysicalProtectionGroupParams(registrationID int64) (*backuprecoveryv1.PhysicalProtectionGroupParams, error) {
	if v == nil || v.VPCVSIProtectionParams == nil {
		return nil, nil
	}

	// Build the file protection object params
	fileProtectionObjectParams, err := v.buildFileProtectionObjectParams(registrationID)
	if err != nil {
		return nil, err
	}

	// Build file protection type params
	fileProtectionTypeParams := &backuprecoveryv1.PhysicalFileProtectionGroupParams{
		Objects: fileProtectionObjectParams,
	}

	// Add optional parameters from VPCVSIProtectionParams
	if v.VPCVSIProtectionParams.QuiesceFilesystem {
		fileProtectionTypeParams.Quiesce = core.BoolPtr(true)
		fileProtectionTypeParams.ContinueOnQuiesceFailure = core.BoolPtr(v.VPCVSIProtectionParams.ContinueOnQuiesceFailure)
	}

	// Add pre/post scripts if provided
	if v.VPCVSIProtectionParams.PreBackupScript != "" || v.VPCVSIProtectionParams.PostBackupScript != "" {
		fileProtectionTypeParams.PrePostScript = v.buildPrePostScriptParams()
	}

	// Add global exclude paths if provided
	if len(v.VPCVSIProtectionParams.GlobalExcludePaths) > 0 {
		fileProtectionTypeParams.GlobalExcludePaths = v.VPCVSIProtectionParams.GlobalExcludePaths
	}

	// Add global exclude filesystems if provided
	if len(v.VPCVSIProtectionParams.GlobalExcludeFS) > 0 {
		fileProtectionTypeParams.GlobalExcludeFS = v.VPCVSIProtectionParams.GlobalExcludeFS
	}

	// Add VSS writers exclusion if provided
	if len(v.VPCVSIProtectionParams.ExcludedVssWriters) > 0 {
		fileProtectionTypeParams.ExcludedVssWriters = v.VPCVSIProtectionParams.ExcludedVssWriters
	}

	// Add deduplication settings
	if v.VPCVSIProtectionParams.PerformSourceSideDeduplication {
		fileProtectionTypeParams.PerformSourceSideDeduplication = core.BoolPtr(true)
	}
	if v.VPCVSIProtectionParams.PerformBrickBasedDeduplication {
		fileProtectionTypeParams.PerformBrickBasedDeduplication = core.BoolPtr(true)
	}

	// Add dedup exclusion source IDs if provided
	if len(v.VPCVSIProtectionParams.DedupExclusionSourceIds) > 0 {
		fileProtectionTypeParams.DedupExclusionSourceIds = v.VPCVSIProtectionParams.DedupExclusionSourceIds
	}

	// Add CoBMR backup setting
	if v.VPCVSIProtectionParams.CobmrBackup {
		fileProtectionTypeParams.CobmrBackup = core.BoolPtr(true)
	}

	// Add ignorable errors if provided
	if len(v.VPCVSIProtectionParams.IgnorableErrors) > 0 {
		fileProtectionTypeParams.IgnorableErrors = v.VPCVSIProtectionParams.IgnorableErrors
	}

	// Add parallel runs setting
	if v.VPCVSIProtectionParams.AllowParallelRuns {
		fileProtectionTypeParams.AllowParallelRuns = core.BoolPtr(true)
	}

	// Build the physical protection group params
	protectionType := backuprecoveryv1.PhysicalProtectionGroupParams_ProtectionType_Kfile
	physicalParams := &backuprecoveryv1.PhysicalProtectionGroupParams{
		ProtectionType:           &protectionType,
		FileProtectionTypeParams: fileProtectionTypeParams,
	}

	return physicalParams, nil
}

// buildFileProtectionObjectParams creates the PhysicalFileProtectionGroupObjectParams
// The ID corresponds to the registrationID (as noted in the task description)
func (v *VPCVSIDataSource) buildFileProtectionObjectParams(registrationID int64) ([]backuprecoveryv1.PhysicalFileProtectionGroupObjectParams, error) {
	var objects []backuprecoveryv1.PhysicalFileProtectionGroupObjectParams

	// Create the object with the registration ID
	obj := backuprecoveryv1.PhysicalFileProtectionGroupObjectParams{
		ID: core.Int64Ptr(registrationID),
	}

	// Add file paths if specified
	if len(v.VPCVSIProtectionParams.IncludePaths) > 0 {
		obj.FilePaths = v.buildFileBackupPathParams()
	}

	// Add follow symlink setting
	if v.VPCVSIProtectionParams.FollowSymlinks {
		obj.FollowNasSymlinkTarget = core.BoolPtr(true)
	}

	objects = append(objects, obj)

	return objects, nil
}

// buildFileBackupPathParams creates PhysicalFileBackupPathParams from include/exclude paths
func (v *VPCVSIDataSource) buildFileBackupPathParams() []backuprecoveryv1.PhysicalFileBackupPathParams {
	var filePaths []backuprecoveryv1.PhysicalFileBackupPathParams

	for _, includePath := range v.VPCVSIProtectionParams.IncludePaths {
		includePath = strings.TrimSpace(includePath)
		if includePath == "" {
			continue
		}

		filePathParam := backuprecoveryv1.PhysicalFileBackupPathParams{
			IncludedPath: core.StringPtr(includePath),
		}

		// Add excluded paths if any
		if len(v.VPCVSIProtectionParams.ExcludePaths) > 0 {
			filePathParam.ExcludedPaths = v.VPCVSIProtectionParams.ExcludePaths
		}

		// Set skip nested volumes if needed
		filePathParam.SkipNestedVolumes = core.BoolPtr(false)

		filePaths = append(filePaths, filePathParam)
	}

	return filePaths
}

// buildPrePostScriptParams creates PrePostScriptParams from VPCVSIProtectionParams
func (v *VPCVSIDataSource) buildPrePostScriptParams() *backuprecoveryv1.PrePostScriptParams {
	prePostScript := &backuprecoveryv1.PrePostScriptParams{}

	// Add pre-backup script
	if v.VPCVSIProtectionParams.PreBackupScript != "" {
		prePostScript.PreScript = &backuprecoveryv1.CommonPreBackupScriptParams{
			Path: core.StringPtr(v.VPCVSIProtectionParams.PreBackupScript),
		}

		if v.VPCVSIProtectionParams.ScriptTimeout > 0 {
			prePostScript.PreScript.Params = core.StringPtr("")
			prePostScript.PreScript.TimeoutSecs = core.Int64Ptr(v.VPCVSIProtectionParams.ScriptTimeout)
		}

		prePostScript.PreScript.IsActive = core.BoolPtr(true)
		prePostScript.PreScript.ContinueOnError = core.BoolPtr(!v.VPCVSIProtectionParams.FailOnScriptError)
	}

	// Add post-backup script
	if v.VPCVSIProtectionParams.PostBackupScript != "" {
		prePostScript.PostScript = &backuprecoveryv1.CommonPostBackupScriptParams{
			Path: core.StringPtr(v.VPCVSIProtectionParams.PostBackupScript),
		}

		if v.VPCVSIProtectionParams.ScriptTimeout > 0 {
			prePostScript.PostScript.Params = core.StringPtr("")
			prePostScript.PostScript.TimeoutSecs = core.Int64Ptr(v.VPCVSIProtectionParams.ScriptTimeout)
		}

		prePostScript.PostScript.IsActive = core.BoolPtr(true)
	}

	return prePostScript
}

// parsePathList parses a comma-separated list of paths
func parsePathList(paths string) []string {
	if paths == "" {
		return []string{}
	}

	parts := strings.Split(paths, ",")
	for i := range parts {
		parts[i] = strings.TrimSpace(parts[i])
	}

	return parts
}
