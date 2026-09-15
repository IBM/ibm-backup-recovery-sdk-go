/*******************************************************************************
 * IBM Confidential
 * OCO Source Materials
 * (C) Copyright IBM Corp. 2026 All Rights Reserved.
 * The source code for this program is not  published or otherwise divested of
 * its trade secrets, irrespective of what has been deposited with
 * the U.S. Copyright Office.
 ******************************************************************************/

// get-backup-and-restore: calls GetBackup and GetRestore and prints both results
// including per-namespace progress.
//
// Required env vars:
//
//	IBM_API_KEY            – IBM Cloud API key
//	BRS_INSTANCE_CRN       – CRN of the BRS instance
//	BRS_RESOURCE_GROUP_ID  – resource group ID
//
// Optional env vars:
//
//	BRS_REGION  – region (default: us-east)
//
// Usage:
//
//	go run ./examples/get-backup-and-restore/ <backupID> <groupID> <restoreID>
package main

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/IBM/ibm-backup-recovery-sdk-go/migrationv2"
	"github.com/IBM/ibm-backup-recovery-sdk-go/migrationv2/common/config"
)

func main() {
	if len(os.Args) < 4 {
		fmt.Fprintln(os.Stderr, "Usage: get-backup-and-restore <backupID> <groupID> <restoreID>")
		os.Exit(1)
	}
	backupID := os.Args[1]
	groupID := os.Args[2]
	restoreID := os.Args[3]

	ctx := context.Background()

	// Call the test or prod function depending on your target environment.
	runTest(ctx, backupID, groupID, restoreID)
	// runProd(ctx, backupID, groupID, restoreID)
}

func runTest(ctx context.Context, backupID, groupID, restoreID string) {
	cfg := config.Config{
		Region:                     mustEnv("BRS_REGION", "us-east"),
		APIKey:                     mustEnv("IBM_API_KEY", ""),
		BRSInstanceCRN:             mustEnv("BRS_INSTANCE_CRN", ""),
		ResourceGroupID:            mustEnv("BRS_RESOURCE_GROUP_ID", ""),
		ResourceControllerEndpoint: "https://resource-controller.test.cloud.ibm.com",
		EnableTaskAPI:              true,
		EnableWorkflowAPI:          false,
		UseDefaultLogger:           true,
		Timeout:                    30 * time.Minute,
	}
	run(ctx, &cfg, backupID, groupID, restoreID)
}

func runProd(ctx context.Context, backupID, groupID, restoreID string) {
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
	run(ctx, &cfg, backupID, groupID, restoreID)
}

func run(ctx context.Context, cfg *config.Config, backupID, groupID, restoreID string) {
	client, err := migrationv2.NewClient(ctx, cfg)
	if err != nil {
		fmt.Fprintf(os.Stderr, "NewClient: %v\n", err)
		os.Exit(1)
	}

	// ── GetBackup ─────────────────────────────────────────────────────────────
	backup, sdkErr := client.TaskAPI.GetBackup(ctx, backupID, groupID)
	if sdkErr != nil {
		fmt.Fprintf(os.Stderr, "GetBackup error: %s\n", sdkErr.Message)
		os.Exit(1)
	}

	fmt.Println("\n── BackupResult ─────────────────────────────────────")
	fmt.Printf("  backupID          : %s\n", backup.BackupID)
	fmt.Printf("  protectionGroupID : %s\n", backup.ProtectionGroupID)
	fmt.Printf("  status            : %s\n", backup.Status)
	fmt.Printf("  startedAt         : %s\n", backup.StartedAt)
	fmt.Printf("  completedAt       : %s\n", backup.CompletedAt)
	if backup.Progress != nil {
		fmt.Printf("  progress          : %.1f%%\n", *backup.Progress)
	} else {
		fmt.Println("  progress          : <nil>")
	}
	fmt.Printf("  namespaceProgress : %d namespace(s)\n", len(backup.NamespaceProgress))
	for i, ns := range backup.NamespaceProgress {
		fmt.Printf("    [%d] name=%-25s  status=%-20s  progress=%d%%\n",
			i, ns.NamespaceName, ns.Status, ns.Progress)
	}
	fmt.Println("─────────────────────────────────────────────────────")

	// ── GetRestore ────────────────────────────────────────────────────────────
	restore, sdkErr := client.TaskAPI.GetRestore(ctx, restoreID)
	if sdkErr != nil {
		fmt.Fprintf(os.Stderr, "GetRestore error: %s\n", sdkErr.Message)
		os.Exit(1)
	}

	fmt.Println("\n── RestoreResult ────────────────────────────────────")
	fmt.Printf("  restoreID         : %s\n", restore.RestoreID)
	fmt.Printf("  backupID          : %s\n", restore.BackupID)
	fmt.Printf("  status            : %s\n", restore.Status)
	fmt.Printf("  startedAt         : %s\n", restore.StartedAt)
	fmt.Printf("  completedAt       : %s\n", restore.CompletedAt)
	fmt.Printf("  progress          : %d%%\n", restore.Progress)
	if len(restore.Messages) > 0 {
		fmt.Printf("  messages          : %v\n", restore.Messages)
	}
	fmt.Printf("  namespaceProgress : %d namespace(s)\n", len(restore.NamespaceProgress))
	for i, ns := range restore.NamespaceProgress {
		fmt.Printf("    [%d] name=%-25s  status=%-20s  progress=%d%%\n",
			i, ns.NamespaceName, ns.Status, ns.Progress)
		if len(ns.Messages) > 0 {
			fmt.Printf("         messages: %v\n", ns.Messages)
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
