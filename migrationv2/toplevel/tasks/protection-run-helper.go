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

	"github.com/IBM/ibm-backup-recovery-sdk-go/backuprecoveryv1"
	"github.com/IBM/ibm-backup-recovery-sdk-go/migrationv2/common/types"
)

func (t *DefaultTaskAPI) buildBackupRunResult(ctx context.Context, backuprun *backuprecoveryv1.ProtectionGroupRun) (*types.BackupResult, error) {
	t.logger.Debug(ctx, "Building backup run result",
		"operation", "buildBackupRunResult",
		"backupRunID", *backuprun.ID,
		"protectionGroupID", *backuprun.ProtectionGroupID)

	// get backupRun progress percentage using `GetProtectionRunProgress` API. Returns nil in case progressPercentage param has not been populated yet.
	progressPercentage, err := t.GetProtectionRunProgress(ctx, *backuprun.ID)
	if err != nil {
		t.logger.Error(ctx, "Failed to fetch backup run progress",
			"operation", "buildBackupRunResult",
			"backupRunID", *backuprun.ID,
			"error", err.Error())
		return nil, fmt.Errorf("Failed to fetch percentage for backuprun %v: %w", *backuprun.ID, err)
	}

	t.logger.Debug(ctx, "Backup run progress fetched",
		"backupRunID", *backuprun.ID,
		"progress", progressPercentage)

	backupResult := &types.BackupResult{
		ProtectionGroupID: *backuprun.ProtectionGroupID,
		BackupID:          *backuprun.ID,
		Progress:          progressPercentage,
	}
	if backuprun.ArchivalInfo != nil && len(backuprun.ArchivalInfo.ArchivalTargetResults) > 0 {
		if backuprun.ArchivalInfo.ArchivalTargetResults[0].StartTimeUsecs != nil {
			backupResult.StartedAt = time.UnixMicro(*backuprun.ArchivalInfo.ArchivalTargetResults[0].StartTimeUsecs)
		}
		if backuprun.ArchivalInfo.ArchivalTargetResults[0].Status != nil {
			backupResult.Status = *backuprun.ArchivalInfo.ArchivalTargetResults[0].Status
		}
		if backuprun.ArchivalInfo.ArchivalTargetResults[0].EndTimeUsecs != nil {
			backupResult.CompletedAt = time.UnixMicro(*backuprun.ArchivalInfo.ArchivalTargetResults[0].EndTimeUsecs)
		}
	}

	t.logger.Info(ctx, "Backup run result built successfully",
		"operation", "buildBackupRunResult",
		"backupRunID", *backuprun.ID,
		"status", backupResult.Status,
		"progress", backupResult.Progress)

	return backupResult, nil
}
