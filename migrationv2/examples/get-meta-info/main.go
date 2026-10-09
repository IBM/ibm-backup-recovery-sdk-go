/*******************************************************************************
 * IBM Confidential
 * OCO Source Materials
 * (C) Copyright IBM Corp. 2026 All Rights Reserved.
 * The source code for this program is not  published or otherwise divested of
 * its trade secrets, irrespective of what has been deposited with
 * the U.S. Copyright Office.
 ******************************************************************************/

// get-meta-info: calls TaskAPI.GetMetaInfo for Kubernetes snapshot metadata.
// Supports retrieving meta-info by:
// 1. GroupID only (automatically queries the latest backup run)
// 2. GroupID + BackupID (queries a specific backup run)
// 3. GroupID + direct SnapshotIDs
//
// Required env vars:
//
//	IBM_API_KEY            – IBM Cloud API key
//	BRS_INSTANCE_CRN       – CRN of the BRS instance
//	BRS_RESOURCE_GROUP_ID  – resource group ID
//
// Optional env vars:
//
//	BRS_REGION             – region (default: us-east)
//
// Usage:
//
//	go run ./examples/get-meta-info/ <groupID> [backupID] [snapshotID]
package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/IBM/go-sdk-core/v5/core"
	"github.com/IBM/ibm-backup-recovery-sdk-go/backuprecoveryv1"
	"github.com/IBM/ibm-backup-recovery-sdk-go/migrationv2"
	"github.com/IBM/ibm-backup-recovery-sdk-go/migrationv2/common/config"
	"github.com/IBM/ibm-backup-recovery-sdk-go/migrationv2/common/types"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Fprintln(os.Stderr, "Usage: get-meta-info <groupID> [backupID] [snapshotID]")
		os.Exit(1)
	}

	groupID := os.Args[1]
	backupID := ""
	if len(os.Args) >= 3 {
		backupID = os.Args[2]
	}
	snapshotID := ""
	if len(os.Args) >= 4 {
		snapshotID = os.Args[3]
	}

	ctx := context.Background()

	// Call test or prod depending on your target environment
	if os.Getenv("IAM_ENDPOINT") != "" {
		runTest(ctx, groupID, backupID, snapshotID)
	} else {
		runProd(ctx, groupID, backupID, snapshotID)
	}
}

func runTest(ctx context.Context, groupID, backupID, snapshotID string) {
	cfg := config.Config{
		Region:                     mustEnv("BRS_REGION", "us-south"),
		APIKey:                     mustEnv("IBM_API_KEY", ""),
		BRSInstanceCRN:             mustEnv("BRS_INSTANCE_CRN", ""),
		ResourceGroupID:            mustEnv("BRS_RESOURCE_GROUP_ID", ""),
		IAMEndpoint:                "https://iam.test.cloud.ibm.com",
		ResourceControllerEndpoint: "https://resource-controller.test.cloud.ibm.com",
		EnableTaskAPI:              true,
		EnableWorkflowAPI:          false,
		UseDefaultLogger:           true,
		Timeout:                    30 * time.Minute,
	}
	run(ctx, &cfg, groupID, backupID, snapshotID)
}

func runProd(ctx context.Context, groupID, backupID, snapshotID string) {
	cfg := config.Config{
		Region:            mustEnv("BRS_REGION", "us-east"),
		APIKey:            mustEnv("IBM_API_KEY", ""),
		BRSInstanceCRN:    mustEnv("BRS_INSTANCE_CRN", ""),
		ResourceGroupID:   mustEnv("BRS_RESOURCE_GROUP_ID", ""),
		EnableTaskAPI:     true,
		EnableWorkflowAPI: false,
		UseDefaultLogger:  true,
		Timeout:           30 * time.Minute,
	}
	run(ctx, &cfg, groupID, backupID, snapshotID)
}

func run(ctx context.Context, cfg *config.Config, groupID, backupID, snapshotID string) {
	client, err := migrationv2.NewClient(ctx, cfg)
	if err != nil {
		fmt.Fprintf(os.Stderr, "NewClient: %v\n", err)
		os.Exit(1)
	}

	// Prepare MetaInfoParams
	metaParams := &types.MetaInfoParams{
		GroupID:  groupID,
		BackupID: backupID,
	}

	if snapshotID != "" {
		metaParams.SnapshotIDs = []string{snapshotID}
	}

	fmt.Printf("\n── Inspecting Protection Group Runs via GetProtectionGroupRuns ────\n")
	brsClient := client.GetBRSClient().(types.BRSClientWrapperInterface)
	opts := &backuprecoveryv1.GetProtectionGroupRunsOptions{
		XIBMTenantID:         core.StringPtr(brsClient.GetTenantId()),
		ID:                   core.StringPtr(groupID),
		IncludeObjectDetails: core.BoolPtr(true),
	}
	runsResp, _, err := brsClient.GetBRSClient().GetProtectionGroupRunsWithContext(ctx, opts)
	if err != nil {
		fmt.Fprintf(os.Stderr, "GetProtectionGroupRuns error: %v\n", err)
	} else if runsResp != nil {
		fmt.Printf("Total runs returned: %d\n", len(runsResp.Runs))
		for idx, r := range runsResp.Runs {
			rID := "<nil>"
			if r.ID != nil {
				rID = *r.ID
			}
			fmt.Printf("  [%d] RunID: %s, Objects count: %d\n", idx, rID, len(r.Objects))
			for oIdx, obj := range r.Objects {
				oName := "<nil>"
				if obj.Object != nil && obj.Object.Name != nil {
					oName = *obj.Object.Name
				}
				snapID := "<none>"
				if obj.LocalSnapshotInfo != nil && obj.LocalSnapshotInfo.SnapshotInfo != nil && obj.LocalSnapshotInfo.SnapshotInfo.SnapshotID != nil {
					snapID = "Local:" + *obj.LocalSnapshotInfo.SnapshotInfo.SnapshotID
				}
				if obj.ArchivalInfo != nil && len(obj.ArchivalInfo.ArchivalTargetResults) > 0 && obj.ArchivalInfo.ArchivalTargetResults[0].SnapshotID != nil {
					snapID = "Archival:" + *obj.ArchivalInfo.ArchivalTargetResults[0].SnapshotID
				}
				fmt.Printf("      - obj[%d]: %s, SnapshotID: %s\n", oIdx, oName, snapID)
			}
		}
	}
	fmt.Println("─────────────────────────────────────────────────────")

	results, sdkErr := client.TaskAPI.GetMetaInfo(ctx, metaParams)
	if sdkErr != nil {
		fmt.Fprintf(os.Stderr, "GetMetaInfo error: %s (cause: %v)\n", sdkErr.Message, sdkErr.Cause)
		os.Exit(1)
	}

	fmt.Printf("\n── MetaInfo Results: %d snapshot(s) ──────────────────\n", len(results))
	for i, res := range results {
		fmt.Printf("\n[%d] SnapshotID  : %s\n", i, res.SnapshotID)
		fmt.Printf("    Namespace   : %s\n", res.Namespace)
		fmt.Printf("    Environment : %s\n", res.Environment)
		if res.KubernetesParams != nil {
			k8s := res.KubernetesParams
			fmt.Printf("    BackedUpResourceCount          : %d\n", k8s.BackedUpResourceCount)
			fmt.Printf("    IncludesClusterScopedResources : %v\n", k8s.IncludesClusterScopedResources)
			fmt.Printf("    BackedUpPvcs Count             : %d\n", len(k8s.BackedUpPvcs))
			for _, pvc := range k8s.BackedUpPvcs {
				pvcName := "<unknown>"
				if pvc.Name != nil {
					pvcName = *pvc.Name
				}
				metadataOnly := false
				if pvc.MetadataOnly != nil {
					metadataOnly = *pvc.MetadataOnly
				}
				fmt.Printf("      - PVC: %-25s MetadataOnly: %v\n", pvcName, metadataOnly)
			}
			fmt.Printf("    BackedUpResources Count        : %d\n", len(k8s.BackedUpResources))
			for _, r := range k8s.BackedUpResources {
				rName := "<unknown>"
				if r.Name != nil {
					rName = *r.Name
				}
				rKind := "<unknown>"
				if r.Kind != nil {
					rKind = *r.Kind
				}
				fmt.Printf("      - Resource: %-25s Kind: %s\n", rName, rKind)
			}
			if len(k8s.IncludedResources) > 0 {
				fmt.Printf("    IncludedResources              : %v\n", k8s.IncludedResources)
			}
			if len(k8s.ExcludedResources) > 0 {
				fmt.Printf("    ExcludedResources              : %v\n", k8s.ExcludedResources)
			}
			if len(k8s.QuiesceRuleStatus) > 0 {
				fmt.Printf("    QuiesceRuleStatus Count        : %d\n", len(k8s.QuiesceRuleStatus))
			}
		}

		// Also pretty print as JSON for inspection
		jsonBytes, err := json.MarshalIndent(res, "    ", "  ")
		if err == nil {
			fmt.Printf("\n    JSON Representation:\n    %s\n", string(jsonBytes))
		}
	}
	fmt.Println("─────────────────────────────────────────────────────")
}

func mustEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	if fallback != "" {
		return fallback
	}
	fmt.Fprintf(os.Stderr, "env var %s is required\n", key)
	os.Exit(1)
	return ""
}
