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
	"fmt"
	"time"

	"github.com/IBM/go-sdk-core/v5/core"
	"github.com/IBM/ibm-backup-recovery-sdk-go/backuprecoveryv1"
	"github.com/IBM/ibm-backup-recovery-sdk-go/migrationv2/common/types"
)

func (t *DefaultTaskAPI) buildBackupRunResult(ctx context.Context, backuprun *backuprecoveryv1.ProtectionGroupRun) (*types.BackupResult, error) {
	t.logger.Debug(ctx, "Building backup run result",
		"operation", "buildBackupRunResult",
		"backupRunID", *backuprun.ID,
		"protectionGroupID", *backuprun.ProtectionGroupID)

	// Call GetProtectionRunProgress once — extracts both overall % and per-namespace objects.
	progressBody, progressErr := t.fetchRunProgress(ctx, *backuprun.ID)
	if progressErr != nil {
		t.logger.Error(ctx, "Failed to fetch backup run progress",
			"operation", "buildBackupRunResult",
			"backupRunID", *backuprun.ID,
			"error", progressErr.Error())
		return nil, fmt.Errorf("Failed to fetch percentage for backuprun %v: %w", *backuprun.ID, progressErr)
	}

	// Extract overall percentage from ArchivalRun[0].PercentageCompleted
	var progressPercentage *float32
	if progressBody != nil && len(progressBody.ArchivalRun) > 0 && progressBody.ArchivalRun[0].PercentageCompleted != nil {
		progressPercentage = progressBody.ArchivalRun[0].PercentageCompleted
	}

	t.logger.Debug(ctx, "Backup run progress fetched",
		"backupRunID", *backuprun.ID,
		"progress", progressPercentage)

	backupResult := &types.BackupResult{
		RunResultBase:     types.RunResultBase{BackupID: *backuprun.ID},
		ProtectionGroupID: *backuprun.ProtectionGroupID,
		Progress:          progressPercentage,
	}

	// Populate run-level status, timing, and messages from ArchivalInfo
	if backuprun.ArchivalInfo != nil && len(backuprun.ArchivalInfo.ArchivalTargetResults) > 0 {
		archival := backuprun.ArchivalInfo.ArchivalTargetResults[0]
		if archival.StartTimeUsecs != nil {
			backupResult.StartedAt = time.UnixMicro(*archival.StartTimeUsecs)
		}
		if archival.Status != nil {
			backupResult.Status = *archival.Status
		}
		if archival.EndTimeUsecs != nil {
			t := time.UnixMicro(*archival.EndTimeUsecs)
			backupResult.CompletedAt = &t
		}
	}

	// Populate per-namespace progress.
	//
	// For cloud-archival-direct runs (isCloudArchivalDirect=true), GetProtectionGroupRun
	// always returns objects=null. Per-namespace data lives in the progress API response:
	// GetProtectionRunProgressBody.ArchivalRun[0].Objects[].
	//
	// For non-cloud-archival-direct runs, objects[] IS populated in GetProtectionGroupRun
	// and we read status/progress from LocalSnapshotInfo as before.
	if progressBody != nil && len(progressBody.ArchivalRun) > 0 && len(progressBody.ArchivalRun[0].Objects) > 0 {
		// Cloud-archival-direct path: per-namespace data from progress API
		archObjs := progressBody.ArchivalRun[0].Objects
		nsProgressList := make([]types.NamespaceBackupProgress, 0, len(archObjs))
		for i := range archObjs {
			obj := &archObjs[i]
			nsProgress := types.NamespaceBackupProgress{}
			if obj.Name != nil {
				nsProgress.NamespaceName = *obj.Name
			}
			if obj.Status != nil {
				nsProgress.Status = *obj.Status
			}
			if obj.PercentageCompleted != nil {
				nsProgress.Progress = int(*obj.PercentageCompleted)
			}
			// When the namespace is in a terminal-success state the server stops
			// updating PercentageCompleted (leaves it at 0). Force it to 100.
			if nsProgress.Status == "Finished" || nsProgress.Status == "Succeeded" {
				nsProgress.Progress = 100
			}
			nsProgressList = append(nsProgressList, nsProgress)
		}
		backupResult.NamespaceProgress = nsProgressList
		t.logger.Debug(ctx, "NamespaceProgress populated from progress API (cloud-archival-direct)",
			"backupRunID", *backuprun.ID,
			"count", len(nsProgressList))
	} else if len(backuprun.Objects) > 0 {
		// Non-cloud-archival-direct path: per-namespace data from GetProtectionGroupRun objects[]
		nsProgressList := make([]types.NamespaceBackupProgress, 0, len(backuprun.Objects))
		for i := range backuprun.Objects {
			obj := &backuprun.Objects[i]
			nsProgress := types.NamespaceBackupProgress{}
			if obj.Object != nil && obj.Object.Name != nil {
				nsProgress.NamespaceName = *obj.Object.Name
			}
			if obj.LocalSnapshotInfo != nil && obj.LocalSnapshotInfo.SnapshotInfo != nil {
				snap := obj.LocalSnapshotInfo.SnapshotInfo
				if snap.Status != nil {
					nsProgress.Status = *snap.Status
				}
				if snap.ProgressTaskID != nil && *snap.ProgressTaskID != "" {
					progress, pErr := t.getRestoreProgress(ctx, *snap.ProgressTaskID)
					if pErr != nil {
						t.logger.Warn(ctx, "Failed to fetch namespace backup progress",
							"backupRunID", *backuprun.ID,
							"namespace", nsProgress.NamespaceName,
							"error", pErr.Error())
					} else if progress != nil {
						nsProgress.Progress = *progress
					}
				}
			}
			nsProgressList = append(nsProgressList, nsProgress)
		}
		if len(nsProgressList) > 0 {
			backupResult.NamespaceProgress = nsProgressList
		}
	}

	t.logger.Info(ctx, "Backup run result built successfully",
		"operation", "buildBackupRunResult",
		"backupRunID", *backuprun.ID,
		"status", backupResult.Status,
		"progress", backupResult.Progress)

	return backupResult, nil
}

// fetchRunProgress calls GetProtectionRunProgress and returns the raw body.
// Returns (nil, nil) when:
//   - no progress data is available yet (normal during early run phase), or
//   - the run record has left the active-progress window after completion (404).
func (t *DefaultTaskAPI) fetchRunProgress(ctx context.Context, backupID string) (*backuprecoveryv1.GetProtectionRunProgressBody, error) {
	result, httpResp, err := t.brsClient.GetBRSClient().GetProtectionRunProgressWithContext(ctx,
		&backuprecoveryv1.GetProtectionRunProgressOptions{
			XIBMTenantID: core.StringPtr(t.brsClient.GetTenantId()),
			RunID:        core.StringPtr(backupID),
		})
	if err != nil {
		// 404 means the run has left the active-progress window (completed runs).
		// Treat as no progress data — not a fatal error.
		if httpResp != nil && httpResp.StatusCode == 404 {
			t.logger.Debug(ctx, "Progress API 404 — run no longer in active window",
				"backupRunID", backupID)
			return nil, nil
		}
		return nil, fmt.Errorf("GetProtectionRunProgress failed: %w", err)
	}
	return result, nil
}
