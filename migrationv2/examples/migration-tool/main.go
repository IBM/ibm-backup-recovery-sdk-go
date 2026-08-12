/*******************************************************************************
 * IBM Confidential
 * OCO Source Materials
 * (C) Copyright IBM Corp. 2026 All Rights Reserved.
 * The source code for this program is not  published or otherwise divested of
 * its trade secrets, irrespective of what has been deposited with
 * the U.S. Copyright Office.
 ******************************************************************************/
package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"strconv"
	"time"

	"github.com/IBM/go-sdk-core/v5/core"
	"github.com/IBM/ibm-backup-recovery-sdk-go/migrationv2"
	activitytracker "github.com/IBM/ibm-backup-recovery-sdk-go/migrationv2/common/activity-tracker"
	"github.com/IBM/ibm-backup-recovery-sdk-go/migrationv2/common/config"
	commoncontext "github.com/IBM/ibm-backup-recovery-sdk-go/migrationv2/common/context"
	"github.com/IBM/ibm-backup-recovery-sdk-go/migrationv2/common/errors"
	"github.com/IBM/ibm-backup-recovery-sdk-go/migrationv2/common/types"
	"github.com/IBM/ibm-backup-recovery-sdk-go/migrationv2/middlelevel/datasources"
	"github.com/IBM/ibm-backup-recovery-sdk-go/migrationv2/middlelevel/datasources/kubernetes"
	"github.com/IBM/ibm-backup-recovery-sdk-go/migrationv2/middlelevel/datasources/vpcvsi"
)

// Migration Tool Example - Uses both WorkflowAPI and TaskAPI, TODO: Not tested yet as APIs implementation not yet done
// This example demonstrates:
// - WorkflowAPI for initiating migration
// - TaskAPI for monitoring progress and status
// - Polling for completion
// - Detailed status reporting

// func main() {
// 	if len(os.Args) < 2 {
// 		printUsage()
// 		os.Exit(1)
// 	}

// 	command := os.Args[1]
// 	ctx := context.Background()

// 	// Initialize SDK client
// 	client, err := initializeClient(ctx)
// 	if err != nil {
// 		log.Fatalf("Failed to initialize client: %v", err)
// 	}

// 	// Execute command
// 	switch command {
// 	case "start":
// 		startMigration(ctx, client)
// 	case "status":
// 		checkStatus(ctx, client)
// 	case "monitor":
// 		monitorMigration(ctx, client)
// 	case "list":
// 		listResources(ctx, client)
// 	default:
// 		fmt.Printf("Unknown command: %s\n", command)
// 		printUsage()
// 		os.Exit(1)
// 	}
// }

func printUsage() {
	fmt.Println("Migration Tool - Complete migration management")
	fmt.Println("\nUsage: migration-tool <command> [options]")
	fmt.Println("\nCommands:")
	fmt.Println("  start    - Start a new migration")
	fmt.Println("  status   - Check status of backup/restore")
	fmt.Println("  monitor  - Monitor migration progress (polls until complete)")
	fmt.Println("  list     - List resources (connections, backups, etc.)")
	fmt.Println("\nExamples:")
	fmt.Println("  migration-tool start")
	fmt.Println("  migration-tool status --backup-id=<id>")
	fmt.Println("  migration-tool status --restore-id=<id>")
	fmt.Println("  migration-tool monitor --backup-id=<id>")
	fmt.Println("  migration-tool list --type=connections")
}

func initializeClient(ctx context.Context) (*migrationv2.Client, error) {
	cfg := &config.Config{
		Region:          getEnv("IBM_REGION", "us-south"),
		APIKey:          getEnv("IBM_API_KEY", ""),
		BRSInstanceName: getEnv("BRS_INSTANCE_NAME", "my-brs-instance"),
		ResourceGroupID: getEnv("RESOURCE_GROUP_ID", ""),
	}

	if cfg.APIKey == "" {
		return nil, fmt.Errorf("IBM_API_KEY environment variable is required")
	}

	return migrationv2.NewClient(ctx, cfg)
}

func startMigration(ctx context.Context, client *migrationv2.Client) {
	fmt.Println("=== Starting Migration ===")
	fmt.Println("This will:")
	fmt.Println("  1. Setup source and target environments")
	fmt.Println("  2. Create protection policy")
	fmt.Println("  3. Configure protection group")
	fmt.Println("  4. Run backup from source")
	fmt.Println("  5. Restore to target")
	fmt.Println()

	// Create data sources
	var sourceDS datasources.DataSource = nil
	var targetDS datasources.DataSource = nil

	// Configure migration
	migrationParams := &types.MigrationParams{
		SourceConnectionParams: &types.ConnectionParams{
			Name: getEnv("SOURCE_CONNECTION", "source-cluster"),
			Type: types.ConnectionType_IKS_CLASSIC,
			Metadata: map[string]interface{}{
				"environment": "production",
			},
		},
		TargetConnectionParams: &types.ConnectionParams{
			Name: getEnv("TARGET_CONNECTION", "target-cluster"),
			Type: types.ConnectionType_IKS_VPC,
			Metadata: map[string]interface{}{
				"environment": "disaster-recovery",
			},
		},
		ProtectionGroupName: "migration-tool-protection",
		CreateDefaultPolicy: true,
		PolicyName:          "migration-tool-policy",
		BackupParams: &types.BackupParams{
			BackupType: types.BackupType_Full,
		},
		RestoreParams: &types.RestoreParams{
			Name:                    "migration-tool-restore",
			BackupPositionFromFirst: 0,
		},
		WaitForBackup:  false, // We'll monitor separately
		WaitForRestore: false, // We'll monitor separately
	}

	// Start migration
	fmt.Println("Initiating migration workflow...")
	result, err := client.WorkflowAPI.ExecuteMigration(ctx, sourceDS, targetDS, migrationParams)
	if err != nil {
		log.Fatalf("Failed to start migration: %v", err)
	}

	// Print results
	fmt.Println("\n✅ Migration workflow started successfully!")
	printMigrationResult(result)

	// Save IDs for monitoring
	fmt.Println("\n📝 Save these IDs for monitoring:")
	fmt.Printf("export BACKUP_ID=%s\n", result.BackupID)
	fmt.Printf("export RESTORE_ID=%s\n", result.RestoreID)
	fmt.Printf("export PROTECTION_GROUP_ID=%s\n", result.ProtectionGroupID)

	fmt.Println("\n💡 Monitor progress with:")
	fmt.Printf("  migration-tool monitor --backup-id=%s\n", result.BackupID)
	fmt.Printf("  migration-tool monitor --restore-id=%s\n", result.RestoreID)
}

func checkStatus(ctx context.Context, client *migrationv2.Client) {
	fmt.Println("=== Checking Status ===")

	backupID := getEnv("BACKUP_ID", "")
	restoreID := getEnv("RESTORE_ID", "")
	// groupID := getEnv("GROUP_ID", "")

	if backupID == "" && restoreID == "" {
		log.Fatal("Either BACKUP_ID or RESTORE_ID environment variable is required")
	}

	// Check backup status
	if backupID != "" {
		fmt.Printf("\nBackup Status (ID: %s):\n", backupID)
		groupID := getEnv("PROTECTION_GROUP_ID", "")
		if groupID == "" {
			log.Fatal("PROTECTION_GROUP_ID environment variable is required when BACKUP_ID is set")
		}
		backup, err := client.TaskAPI.GetBackup(ctx, backupID, groupID)
		if err != nil {
			log.Printf("  Error: %v", err)
		} else {
			printBackupStatus(backup)
		}
	}

	// Check restore status
	if restoreID != "" {
		fmt.Printf("\nRestore Status (ID: %s):\n", restoreID)
		restore, err := client.TaskAPI.GetRestore(ctx, restoreID)
		if err != nil {
			log.Printf("  Error: %v", err)
		} else {
			printRestoreStatus(restore)
		}
	}
}

func monitorMigration(ctx context.Context, client *migrationv2.Client) {
	fmt.Println("=== Monitoring Migration ===")

	backupID := getEnv("BACKUP_ID", "")
	// groupID := getEnv("GROUP_ID", "")
	restoreID := getEnv("RESTORE_ID", "")

	if backupID == "" && restoreID == "" {
		log.Fatal("Either BACKUP_ID or RESTORE_ID environment variable is required")
	}

	pollInterval := 10 * time.Second
	maxAttempts := 180 // 30 minutes

	// Monitor backup
	if backupID != "" {
		fmt.Printf("\n📊 Monitoring backup: %s\n", backupID)
		fmt.Println("Press Ctrl+C to stop monitoring")

		for attempt := 0; attempt < maxAttempts; attempt++ {
			groupID := getEnv("PROTECTION_GROUP_ID", "")
			if groupID == "" {
				log.Fatal("PROTECTION_GROUP_ID environment variable is required when BACKUP_ID is set")
			}
			backup, err := client.TaskAPI.GetBackup(ctx, backupID, groupID)
			if err != nil {
				log.Printf("Error checking backup: %v", err)
				time.Sleep(pollInterval)
				continue
			}

			// Print progress
			fmt.Printf("\r[%s] Status: %-12s Progress: %3d%% | %s",
				time.Now().Format("15:04:05"),
				backup.Status,
				backup.Progress,
				backup.Message)

			// Check if complete
			if backup.Status == "completed" {
				fmt.Println("\n\n✅ Backup completed successfully!")
				printBackupStatus(backup)
				break
			} else if backup.Status == "failed" {
				fmt.Println("\n\n❌ Backup failed!")
				printBackupStatus(backup)
				os.Exit(1)
			}

			time.Sleep(pollInterval)
		}
	}

	// Monitor restore
	if restoreID != "" {
		fmt.Printf("\n📊 Monitoring restore: %s\n", restoreID)
		fmt.Println("Press Ctrl+C to stop monitoring")

		for attempt := 0; attempt < maxAttempts; attempt++ {
			restore, err := client.TaskAPI.GetRestore(ctx, restoreID)
			if err != nil {
				log.Printf("Error checking restore: %v", err)
				time.Sleep(pollInterval)
				continue
			}

			// Print progress
			fmt.Printf("\r[%s] Status: %-12s Progress: %3d%% | %s",
				time.Now().Format("15:04:05"),
				restore.Status,
				restore.Progress,
				restore.Message)

			// Check if complete
			if restore.Status == "completed" {
				fmt.Println("\n\n✅ Restore completed successfully!")
				printRestoreStatus(restore)
				break
			} else if restore.Status == "failed" {
				fmt.Println("\n\n❌ Restore failed!")
				printRestoreStatus(restore)
				os.Exit(1)
			}

			time.Sleep(pollInterval)
		}
	}
}

func listResources(ctx context.Context, client *migrationv2.Client) {
	resourceType := getEnv("TYPE", "connections")

	fmt.Printf("=== Listing %s ===\n\n", resourceType)

	switch resourceType {
	case "connections":
		connections, err := client.TaskAPI.ListConnections(ctx)
		if err != nil {
			log.Fatalf("Failed to list connections: %v", err)
		}
		fmt.Printf("Found %d connections:\n", len(connections))
		for i, conn := range connections {
			fmt.Printf("%d. %s (ID: %s, Status: %s)\n", i+1, conn.ConnectionName, conn.ConnectionID, conn.Status)
		}

	case "registrations":
		registrations, err := client.TaskAPI.ListRegistrations(ctx)
		if err != nil {
			log.Fatalf("Failed to list registrations: %v", err)
		}
		fmt.Printf("Found %d registrations:\n", len(registrations))
		for i, reg := range registrations {
			fmt.Printf("%d. %s (ID: %d, Status: %s)\n", i+1, reg.SourceName, reg.RegistrationID, reg.Status)
		}

	case "policies":
		policies, err := client.TaskAPI.ListPolicies(ctx)
		if err != nil {
			log.Fatalf("Failed to list policies: %v", err)
		}
		fmt.Printf("Found %d policies:\n", len(policies))
		for i, policy := range policies {
			fmt.Printf("%d. %s (ID: %s, Retention: %d days)\n", i+1, policy.Name, policy.ID, policy.RetentionDays)
		}

	case "backups":
		groupID := getEnv("PROTECTION_GROUP_ID", "")
		if groupID == "" {
			log.Fatal("PROTECTION_GROUP_ID environment variable is required for listing backups")
		}
		backups, err := client.TaskAPI.ListBackups(ctx, groupID)
		if err != nil {
			log.Fatalf("Failed to list backups: %v", err)
		}
		fmt.Printf("Found %d backups for group %s:\n", len(backups), groupID)
		for i, backup := range backups {
			fmt.Printf("%d. %s (Status: %s, Progress: %d%%)\n", i+1, backup.BackupID, backup.Status, backup.Progress)
		}

	default:
		log.Fatalf("Unknown resource type: %s", resourceType)
	}
}

// Helper functions

func printMigrationResult(result *types.MigrationResult) {
	fmt.Printf("\n📋 Migration Details:\n")
	fmt.Printf("  Status: %s\n", result.Status)
	fmt.Printf("  Duration: %.2f seconds\n", result.Duration)

	fmt.Printf("\n🔗 Resource IDs:\n")
	fmt.Printf("  Source Connection:   %s\n", result.SourceConnectionID)
	fmt.Printf("  Source Connector:    %s\n", result.SourceConnectorID)
	fmt.Printf("  Source Registration: %d\n", result.SourceRegistrationID)
	fmt.Printf("  Target Connection:   %s\n", result.TargetConnectionID)
	fmt.Printf("  Target Connector:    %s\n", result.TargetConnectorID)
	fmt.Printf("  Target Registration: %d\n", result.TargetRegistrationID)
	fmt.Printf("  Policy:              %s (Created: %v)\n", result.PolicyID, result.PolicyCreated)
	fmt.Printf("  Protection Group:    %s\n", result.ProtectionGroupID)
	fmt.Printf("  Backup:              %s\n", result.BackupID)
	fmt.Printf("  Restore:             %s\n", result.RestoreID)

	fmt.Printf("\n📊 Component Status:\n")
	fmt.Printf("  Source Setup:  %s\n", result.SourceSetupStatus)
	fmt.Printf("  Target Setup:  %s\n", result.TargetSetupStatus)
	fmt.Printf("  Protection:    %s\n", result.ProtectionStatus)
	fmt.Printf("  Backup:        %s\n", result.BackupStatus)
	fmt.Printf("  Restore:       %s\n", result.RestoreStatus)

	fmt.Printf("\n♻️  Resources:\n")
	fmt.Printf("  Created: %d\n", result.ResourcesCreated)
	fmt.Printf("  Reused:  %d\n", result.ResourcesReused)

	if len(result.StepResults) > 0 {
		fmt.Printf("\n📝 Step Results:\n")
		for _, step := range result.StepResults {
			icon := "✅"
			if step.Status == "failed" {
				icon = "❌"
			} else if step.Status == "skipped" {
				icon = "⏭️ "
			}
			fmt.Printf("  %s [%s] %s (%.2fs)\n", icon, step.Step, step.Message, step.Duration)
		}
	}
}

func printBackupStatus(backup *types.BackupResult) {
	fmt.Printf("  Status:     %s\n", backup.Status)
	if backup.Progress != nil {
		fmt.Printf("  Progress:   %.0f%%\n", *backup.Progress)
	} else {
		fmt.Printf("  Progress:   N/A\n")
	}
	fmt.Printf("  Message:    %s\n", backup.Message)
	fmt.Printf("  Started:    %s\n", backup.StartedAt.Format(time.RFC3339))
	if !backup.CompletedAt.IsZero() {
		fmt.Printf("  Completed:  %s\n", backup.CompletedAt.Format(time.RFC3339))
		duration := backup.CompletedAt.Sub(backup.StartedAt)
		fmt.Printf("  Duration:   %s\n", duration.Round(time.Second))
	}
}

func printRestoreStatus(restore *types.RestoreResult) {
	fmt.Printf("  Status:     %s\n", restore.Status)
	fmt.Printf("  Progress:   %d%%\n", restore.Progress)
	fmt.Printf("  Message:    %s\n", restore.Message)
	fmt.Printf("  Started:    %s\n", restore.StartedAt.Format(time.RFC3339))
	if restore.CompletedAt != nil {
		fmt.Printf("  Completed:  %s\n", restore.CompletedAt.Format(time.RFC3339))
		duration := restore.CompletedAt.Sub(restore.StartedAt)
		fmt.Printf("  Duration:   %s\n", duration.Round(time.Second))
	}
}

func parseStorageClassMappings() []types.StorageClassMapping {
	mappings := getEnv("STORAGE_CLASS_MAPPINGS", "")
	if mappings == "" {
		return nil
	}

	// Format: "old1:new1,old2:new2"
	// Example: "gp2:standard-rwo,io1:premium-rwo"
	var result []types.StorageClassMapping
	// Parse logic would go here
	return result
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvBool(key string, defaultValue bool) bool {
	if value := os.Getenv(key); value != "" {
		b, err := strconv.ParseBool(value)
		if err == nil {
			return b
		}
	}
	return defaultValue
}

func convertToPointer[T any](v T) *T {
	return &v
}

func CreateBackupRun(ctx context.Context, client *migrationv2.Client) {

	//create backupjob and list backup jobs

	backupParams := &types.BackupParams{
		BackupType: types.BackupType_Incremental,
	}

	//
	result, perr := client.TaskAPI.RunBackup(ctx, "8305184241232842:1757331781254:2106059", backupParams)
	if perr != nil {
		fmt.Println("failed to create protection group run:", perr.Error())
		return
	}
	fmt.Println("ProtectionGroupID:", result.ProtectionGroupID)

	backups, perr := client.TaskAPI.ListBackups(ctx, "8305184241232842:1757331781254:2106059")
	if perr != nil {
		fmt.Println("failed to list protection group runs:", perr.Error())
		return
	}
	for _, backup := range backups {
		fmt.Println(backup.BackupID)
		fmt.Println(backup.Status)
	}

	backup, perr := client.TaskAPI.GetBackup(ctx, "2106059:1774593683993191", "8305184241232842:1757331781254:2106059")
	if perr != nil {
		fmt.Println("failed to create protection group:", perr.Error())
		return
	}
	fmt.Println(backup.BackupID)
	fmt.Println(backup.Status)
	fmt.Println(backup.Progress)
	fmt.Println(backup.ProtectionGroupID)
	fmt.Println(backup.StartedAt)
	fmt.Println(backup.CompletedAt)
}

func main() {
	ctx := context.Background()
	ctx = commoncontext.WithTransactionID(ctx, "123abc#")

	cfg := config.DefaultConfig().
		WithRegion(getEnv("IBM_REGION", "us-east")).
		WithAPIKey(getEnv("IBM_API_KEY", "")).
		WithBRSInstanceName(getEnv("BRS_INSTANCE_NAME", "brs-iks-roks-7216p-instance")).
		// WithBRSInstanceCRN("crn:v1:bluemix:public:backup-recovery-tests:us-east:a/0f628e88c6594675bbefa097a63b9293:63cc7397-1455-4b92-b4f5-6d39f1d352e5::").
		WithResourceGroupID(getEnv("RESOURCE_GROUP_ID", "shared-resource-group")).
		WithTimeout(30 * time.Second).
		WithTaskAPI(true).
		WithWorkflowAPI(false).
		WithActivityTrackerEnabled(true).
		WithActivityTrackerConfig(&activitytracker.HTTPSinkConfig{
			IngestionEndpoint: "https://56b8a22f-9bc7-429b-9f23-3faf26e9e33d.ingress.us-south.logs.cloud.ibm.com/",
			Timeout:           30 * time.Second,
			// IAMAuthenticator is populated automatically from cfg.APIKey.
		})
	// TenantId is optional – set only when you know the exact value.
	cfg.TenantId = "u77h8fih5n/"

	client, err := migrationv2.NewClient(ctx, cfg)
	if err != nil {
		fmt.Println(err)
		return
	}

	// // ---- policy-----
	policyParams := types.PolicyParams{
		Name:       "migration-policy-104",
		BackupType: "full",
		PrimaryBackupTarget: &types.PrimaryBackupTarget{
			UseDefaultBackupTarget: true,
		},
		IncrementalBackup: &types.IncrementalBackup{
			Unit: convertToPointer(types.IncrementalBackup_Unit_Days),
			DaySchedule: &types.UnitDaySchedule{
				Every: 3,
			},
		},
		FullBackups: []types.FullBackup{{
			Unit: convertToPointer(types.FullSchedule_Unit_Days),
			DaySchedule: &types.UnitDaySchedule{
				Every: 1,
			},
			Retention: &types.DataRetention{
				RetainFor: 2,
				Unit:      convertToPointer(types.Retention_Unit_Days),
			},
		}},
		DataRetention: &types.DataRetention{
			RetainFor: 3,
			Unit:      convertToPointer(types.Retention_Unit_Days),
		},
		RetryOption: &types.RetryOption{
			NumberOfRetry:       3,
			WaitInCaseOfFailure: 4,
		},
	}

	newPolicy, perr := client.TaskAPI.CreatePolicy(ctx, &policyParams)
	if perr != nil {
		fmt.Println("failed to create protection policy:", perr.Error())
		return
	}
	fmt.Println(newPolicy.ID)
	getPolicy, perr := client.TaskAPI.GetPolicy(ctx, newPolicy.ID)
	if perr != nil {
		fmt.Println("failed to create protection group:", perr.Error())
		return
	}
	fmt.Println(getPolicy.Name)

}

func CreatePolicy(ctx context.Context, client *migrationv2.Client) {
	policyParams := types.PolicyParams{
		Name: "migration-policy-102",
		PrimaryBackupTarget: &types.PrimaryBackupTarget{
			UseDefaultBackupTarget: true,
		},
		IncrementalBackup: &types.IncrementalBackup{
			Unit: convertToPointer(types.IncrementalBackup_Unit_Days),
			DaySchedule: &types.UnitDaySchedule{
				Every: 3,
			},
		},
		FullBackups: []types.FullBackup{{
			Unit: convertToPointer(types.FullSchedule_Unit_Days),
			DaySchedule: &types.UnitDaySchedule{
				Every: 1,
			},
			Retention: &types.DataRetention{
				RetainFor: 2,
				Unit:      convertToPointer(types.Retention_Unit_Days),
				DataLockConfig: &types.DataLockConfig{
					Unit:     convertToPointer(types.DataLockConfig_Unit_Days),
					Mode:     convertToPointer(types.DataLockConfig_Mode_Administrative),
					Duration: 1,
				},
			},
		}},
		BackupType: "full",
		DataRetention: &types.DataRetention{
			RetainFor: 3,
			Unit:      convertToPointer(types.Retention_Unit_Days),
			DataLockConfig: &types.DataLockConfig{
				Unit:     convertToPointer(types.DataLockConfig_Unit_Days),
				Mode:     convertToPointer(types.DataLockConfig_Mode_Administrative),
				Duration: 2,
			},
		},
		RetryOption: &types.RetryOption{
			NumberOfRetry:       3,
			WaitInCaseOfFailure: 4,
		},
	}

	newPolicy, perr := client.TaskAPI.CreatePolicy(ctx, &policyParams)
	if perr != nil {
		fmt.Println("failed to create protection policy:", perr.Error())
		return
	}
	fmt.Println(newPolicy.ID)
	getPolicy, perr := client.TaskAPI.GetPolicy(ctx, newPolicy.ID)
	if perr != nil {
		fmt.Println("failed to create protection policy:", perr.Error())
		return
	}

	fmt.Println(getPolicy.Name)
}

func createProtectionGroup(ctx context.Context, client *migrationv2.Client) {

	dataSource, err := client.CreateDataSource(&kubernetes.KubernetesDataSourceConfig{
		KubernetesProtectionParams: &types.KubernetesProtectionParams{
			IncludeNamespaces: "my-app",
			ExcludeNamespaces: "temp-my-app-1",
			FailOnHookError:   false,
			Settings:          &types.ProtectionSetting{CSISnapshot: true},
			NamespacesSetting: []types.NamespacesSetting{
				{Namespace: "my-app",
					PersistentVolumeClaims: types.PVCSection{Inclusion: []string{"mysql-pv-claim"}},
					Resources:              types.ResourceSection{Inclusion: &types.ResourceTypes{ResourceTypes: []string{"Deployment"}}},
					Hooks: types.HookSection{
						RulesApplyMode:         "kQuiesceTogether",
						FailBackupIfHookFailed: false,
						Rules: []types.Rule{
							{Rule: types.RuleDetail{
								PodLabels:  "key:value, key1:value1",
								PreScript:  "scriptPath",
								PostScript: "postScriptPath",
							},
							},
						},
					},
				},
			},
		},
	})
	if err != nil {
		fmt.Println("failed to create datasource:", err)
		return
	}
	groupParams := &types.ProtectionGroupParams{
		Name:            "migration-k8s-group-3",
		NumberOfBackups: 1,
		Policy: &types.Policy{
			ID: "8305184241232842:1757331781254:1077874",
		},
	}
	var perr *errors.SDKError
	newGroup, perr := client.TaskAPI.CreateProtectionGroup(ctx, 51249, groupParams, dataSource)
	if perr != nil {
		fmt.Println("failed to create protection group:", perr.Error())
		return
	}
	fmt.Println(" CreateProtectionGroup:", newGroup.ProtectionGroupID)

	getGroup, perr := client.TaskAPI.GetProtectionGroup(ctx, newGroup.ProtectionGroupID)

	if perr != nil {
		fmt.Println("failed to create protection group:", perr.Error())
		return
	}
	fmt.Println("GetProtectionGroup:", getGroup.GroupName)

	time.Sleep(5 * time.Second)
	perr = client.TaskAPI.DeleteProtectionGroup(ctx, newGroup.ProtectionGroupID, false)
	if perr != nil {
		fmt.Println("failed to delete protection group:", perr.Error())
		return
	}

}

func CreateRecovery(ctx context.Context, client *migrationv2.Client) {

	// --recovery
	var perr *errors.SDKError
	dataSource, err := client.CreateDataSource(&kubernetes.KubernetesDataSourceConfig{
		KubernetesRestoreParams: &types.KubernetesRestoreParams{
			RecoverObjectSpec: &types.RecoverObjectSpec{
				RestoreOnlyPvc: false,
				IncludeObjects: &types.K8sObject{
					SelectedResources: []types.ResourceInfo{
						{
							ResourceKind:     "PersistentVolumeClaim",
							ResourceApiGroup: "",
							ResourceList: []types.ResourceInstance{{
								Name: "mysql-pv-claim",
							}},
						},
					},
					SelectedLabels: &types.SelectedLabels{
						LabelCombination: "OR",
						Labels: []types.K8sLabel{
							{
								Key:   "a",
								Value: "1",
							},
						},
					},
				},
				ExcludeObjects: &types.K8sObject{
					SelectedResources: []types.ResourceInfo{
						{
							ResourceKind:     "ReplicaSet",
							ResourceApiGroup: "apps",
						},
					},
				},
				UseStorageClassMapping: core.BoolPtr(true),
				StorageClasses: []types.StorageClassMapping{{
					Old: "ibmc-vpc-block-10iops-tier",
					New: "ibmc-vpc-block-10iops-tier",
				}},
			},
			RenameRecoveredNamespacesParams: &types.RenameRecoveredNamespacesParams{
				Prefix: "copy-1",
				Suffix: "-1",
			},
			ZoneMappings: []types.MigrationMapParams{
				{
					Source: "z1",
					Target: "z2",
				},
			},
			RegionMapping: &types.MigrationMapParams{
				Source: "r1",
				Target: "r2",
			},
			RecoverToNewTarget: true,
		},
	})
	// Protection group doesn't exist, create it
	recoveryParams := &types.RestoreParams{
		Name: "mig-recover-103",
	}
	if err != nil {
		fmt.Println("failed to create datasource:", err)
		return
	}

	result, perr := client.TaskAPI.RunRestore(ctx, "8305184241232842:1757331781254:2106059", "2106059:1774593683993191", 51249, recoveryParams, dataSource)
	if perr != nil {
		fmt.Println("failed to create protection group:", perr.Error())
		return
	}
	fmt.Println(result.RestoreID)

}

func CreateVSIProtectionGroup() {
	ctx := context.Background()
	// Initialize SDK client
	cfg := &config.Config{
		Region:          getEnv("IBM_REGION", "us-east"),
		APIKey:          getEnv("IBM_API_KEY", ""),
		BRSInstanceName: getEnv("BRS_INSTANCE_NAME", "baas-vsi-migration-sdk-testing"), // Same for all
		// BRSInstanceCRN:  "crn:v1:bluemix:public:backup-recovery-tests:us-east:a/0f628e88c6594675bbefa097a63b9293:63cc7397-1455-4b92-b4f5-6d39f1d352e5::",
		ResourceGroupID: getEnv("RESOURCE_GROUP_ID", "shared-resource-group"),
		TenantId:        "nlevs3dsrx/",
		Timeout:         time.Duration(30 * time.Second),
		EnableTaskAPI:   true,
	}

	client, err := migrationv2.NewClient(ctx, cfg)
	if err != nil {
		fmt.Println(err)
		return
	}
	// --- group----
	dataSource, err := client.CreateDataSource(&vpcvsi.PhysicalDataSourceConfig{
		VPCVSIProtectionParams: &types.VPCVSIProtectionParams{
			IncludePaths: []string{"/"},
		},
	})
	if err != nil {
		fmt.Println("failed to create datasource:", err)
		return
	}
	groupParams := &types.ProtectionGroupParams{
		Name:            "migration-physical-group-3",
		NumberOfBackups: 1,
		Policy: &types.Policy{
			ID: "183326634598472:1757392362704:33476",
		},
	}
	var perr *errors.SDKError
	newGroup, perr := client.TaskAPI.CreateProtectionGroup(ctx, 297, groupParams, dataSource)
	if perr != nil {
		fmt.Println("failed to create protection group:", perr.Error())
		return
	}
	fmt.Println(" CreateProtectionGroup:", newGroup.ProtectionGroupID)

	getGroup, perr := client.TaskAPI.GetProtectionGroup(ctx, newGroup.ProtectionGroupID)
	if perr != nil {
		fmt.Println("failed to create protection group:", perr.Error())
		return
	}
	fmt.Println("GetProtectionGroup:", getGroup.GroupName)

	time.Sleep(5 * time.Second)
	perr = client.TaskAPI.DeleteProtectionGroup(ctx, newGroup.ProtectionGroupID, false)
	if perr != nil {
		fmt.Println("failed to delete protection group:", perr.Error())
		return
	}
}

func CreateVSIRestoreParams() {
	ctx := context.Background()
	// Initialize SDK client
	cfg := &config.Config{
		Region:          getEnv("IBM_REGION", "us-east"),
		APIKey:          getEnv("IBM_API_KEY", ""),
		BRSInstanceName: getEnv("BRS_INSTANCE_NAME", "baas-vsi-migration-sdk-testing"), // Same for all
		// BRSInstanceCRN:  "crn:v1:bluemix:public:backup-recovery-tests:us-east:a/0f628e88c6594675bbefa097a63b9293:63cc7397-1455-4b92-b4f5-6d39f1d352e5::",
		ResourceGroupID: getEnv("RESOURCE_GROUP_ID", "shared-resource-group"),
		TenantId:        "nlevs3dsrx/",
		Timeout:         time.Duration(30 * time.Second),
		EnableTaskAPI:   true,
	}
	client, err := migrationv2.NewClient(ctx, cfg)
	if err != nil {
		fmt.Println(err)
		return
	}

	dataSource, err := client.CreateDataSource(&vpcvsi.PhysicalDataSourceConfig{
		VPCVSIRestoreParams: &types.VsiVpcRestoreParams{
			RecoveryAction: types.RecoveryAction_RecoverFiles,
			RecoverFileAndFolderParams: &types.RecoverFileAndFolderParams{
				FilesAndFolders: []types.FileAndFolderInfo{
					{AbsolutePath: "/home",
						IsDirectory: true,
					},
				},
				RestoreToOriginalPaths: true,
				OverwriteExisting:      false,
				PreserveAttributes:     true,
				ContinueOnError:        false,
				SaveSuccessFiles:       true,
				RestoreEntityType:      "kRegular",
			},
		},
	})

	restoreParams := &types.RestoreParams{
		Name:                    "restore-dee-1",
		BackupPositionFromFirst: 1,
	}

	var perr *errors.SDKError
	newRecovery, perr := client.TaskAPI.RunRestore(ctx, "1589079364046703:1757422346343:203404", "203404:1776737326830585", 14276, restoreParams, dataSource)
	if perr != nil {
		fmt.Println("failed to create protection group:", perr.Error())
		return
	}
	fmt.Println(" CreateRecovery:", newRecovery.RestoreID)
}
