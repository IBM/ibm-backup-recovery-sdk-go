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
	"time"

	"github.com/IBM/go-sdk-core/v5/core"
	"github.com/IBM/ibm-backup-recovery-sdk-go/migrationv2"
	"github.com/IBM/ibm-backup-recovery-sdk-go/migrationv2/common/config"
	"github.com/IBM/ibm-backup-recovery-sdk-go/migrationv2/common/errors"
	"github.com/IBM/ibm-backup-recovery-sdk-go/migrationv2/common/types"
	"github.com/IBM/ibm-backup-recovery-sdk-go/migrationv2/middlelevel/datasources"
	"github.com/IBM/ibm-backup-recovery-sdk-go/migrationv2/middlelevel/datasources/kubernetes"
)

// Policy Task API Example
// This example demonstrates:
// - Initializing the migrationv2 client for Task API usage
// - Using an existing BRS instance via BRS_INSTANCE_CRN
// - Creating a policy using TaskAPI
// - Listing policies using TaskAPI
// - Optionally controlling the policy name through environment variables

func main() {
	ctx := context.Background()

	client, err := initializeClient(ctx)
	if err != nil {
		log.Fatalf("Failed to initialize client: %v", err)
	}

	policyName := getEnv("POLICY_NAME", fmt.Sprintf("migration-policy-%d", time.Now().Unix()))

	fmt.Println("=== Policy Task API Example ===")
	fmt.Printf("BRS instance CRN: %s\n", getEnv("BRS_INSTANCE_CRN", ""))
	fmt.Printf("Policy name: %s\n\n", policyName)

	createdPolicy, err := createPolicy(ctx, client, policyName)
	if err != nil {
		log.Fatalf("Failed to create policy: %v", err)
	}

	fmt.Println("✅ Policy created successfully")
	printPolicy(createdPolicy)

	fmt.Println("\n=== Listing Policies ===")
	policies, sdkErr := client.TaskAPI.ListPolicies(ctx)
	if sdkErr != nil && sdkErr.Message != "" {
		log.Fatalf("Failed to list policies: %v", sdkErr)
	}

	fmt.Printf("Found %d policies\n", len(policies))
	for i, policy := range policies {
		fmt.Printf("%d. %s (ID: %s)\n", i+1, policy.Name, policy.ID)
	}

	dataSource, err := client.CreateDataSource(&kubernetes.KubernetesDataSourceConfig{
		Authenticator: &core.IamAuthenticator{
			ApiKey: "apiKey",
			URL:    "https://iam.cloud.ibm.com",
		},
		ClusterName:           "SourceCluster.Name",
		ClusterType:           "SourceCluster.Type",
		ClusterEndpoint:       "SourceCluster.EndpointURL", // Use actual endpoint URL
		ContainerEndpoint:     "https://containers.cloud.ibm.com/global",
		ContainerEndpointType: "public",
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
								Key:   "app",
								Value: "label",
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
					Source: "zone1",
					Target: "zone2",
				},
			},
			RegionMapping: &types.MigrationMapParams{
				Source: "region1",
				Target: "region2",
			},
			RecoverToNewTarget: true,
		},
	})
	if err != nil {
		log.Fatalf("Failed to create datasource: %v", err)
	}

	// create protection group
	var policyId string = createdPolicy.ID
	var protectionGroupName = "migration-k8s-group-3"
	var registrationID int64 = 51249
	group, err := createProtectionGroup(ctx, client, dataSource, registrationID, policyId, protectionGroupName)
	if err != nil {
		log.Fatalf("Failed to create group: %v", err)
	}

	// create protection group run
	protectionGroupId := group.ProtectionGroupID
	grouprun, err := CreateBackupRun(ctx, client, protectionGroupId)
	if err != nil {
		log.Fatalf("Failed to create group: %v", err)
	}
	// recover to new target
	var backupRunId string = grouprun
	var targetRegistrationId int64 = 51249
	restoreesult, err := CreateRecovery(ctx, client, dataSource, targetRegistrationId, protectionGroupId, backupRunId)
	if err != nil {
		log.Fatalf("Failed to create recovery: %v", err)
	}
	fmt.Println(restoreesult.RestoreID)
}

func initializeClient(ctx context.Context) (*migrationv2.Client, error) {
	cfg := config.DefaultConfig()
	cfg.Region = getEnv("IBM_REGION", "us-south")
	cfg.APIKey = getEnv("IBM_API_KEY", "")
	cfg.BRSInstanceCRN = getEnv("BRS_INSTANCE_CRN", "")
	cfg.ResourceGroupID = getEnv("RESOURCE_GROUP_ID", "not-required-when-using-existing-instance")
	cfg.TenantId = getEnv("BRS_TENANT_ID", "")
	cfg.EnableTaskAPI = true
	cfg.EnableWorkflowAPI = false

	if cfg.APIKey == "" {
		return nil, fmt.Errorf("IBM_API_KEY environment variable is required")
	}

	if cfg.BRSInstanceCRN == "" {
		return nil, fmt.Errorf("BRS_INSTANCE_CRN environment variable is required")
	}

	return migrationv2.NewClient(ctx, cfg)
}

func createPolicy(ctx context.Context, client *migrationv2.Client, policyName string) (*types.PolicyResult, error) {
	policyParams := &types.PolicyParams{
		Name:        policyName,
		Description: "Sample CAD policy created by the policy-task-api example",
		PrimaryBackupTarget: &types.PrimaryBackupTarget{
			UseDefaultBackupTarget: true,
		},
		IncrementalBackup: &types.IncrementalBackup{
			Unit: convertToPointer(types.IncrementalBackup_Unit_Days),
			DaySchedule: &types.UnitDaySchedule{
				Every: 1,
			},
		},
		FullBackups: []types.FullBackup{
			{
				Unit: convertToPointer(types.FullSchedule_Unit_Days),
				DaySchedule: &types.UnitDaySchedule{
					Every: 1,
				},
				Retention: &types.DataRetention{
					RetainFor: 7,
					Unit:      convertToPointer(types.Retention_Unit_Days),
				},
			},
		},
		BackupType: "full",
		DataRetention: &types.DataRetention{
			RetainFor: 7,
			Unit:      convertToPointer(types.Retention_Unit_Days),
		},
		RetryOption: &types.RetryOption{
			NumberOfRetry:       1,
			WaitInCaseOfFailure: 5,
		},
	}

	policy, sdkErr := client.TaskAPI.CreatePolicy(ctx, policyParams)
	if sdkErr != nil {
		return nil, sdkErr
	}

	return policy, nil
}

func printPolicy(policy *types.PolicyResult) {
	fmt.Printf("ID: %s\n", policy.ID)
	fmt.Printf("Name: %s\n", policy.Name)
	fmt.Printf("Description: %s\n", policy.Description)
	fmt.Printf("Backup Type: %s\n", policy.BackupType)
}

func convertToPointer[T any](v T) *T {
	return &v
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func createProtectionGroup(ctx context.Context, client *migrationv2.Client, dataSource datasources.DataSource, registrationId int64, policyId, protectionGroupName string) (*types.ProtectionGroupResult, error) {

	groupParams := &types.ProtectionGroupParams{
		Name:            protectionGroupName,
		NumberOfBackups: 1,
		Policy: &types.Policy{
			ID: policyId,
		},
	}
	var perr *errors.SDKError

	newGroup, perr := client.TaskAPI.CreateProtectionGroup(ctx, registrationId, groupParams, dataSource)
	if perr != nil {

		return nil, fmt.Errorf("failed to create protection group: %v", perr.Error())
	}
	fmt.Println(" CreateProtectionGroup:", newGroup.ProtectionGroupID)

	return newGroup, nil

}

func CreateBackupRun(ctx context.Context, client *migrationv2.Client, protectionGroupId string) (string, error) {

	var perr *errors.SDKError

	//create backupjob and list backup jobs
	backupParams := &types.BackupParams{
		BackupType: types.BackupType_Incremental,
	}

	//
	result, perr := client.TaskAPI.RunBackup(ctx, protectionGroupId, backupParams)
	if perr != nil {
		return "", fmt.Errorf("failed to create protection group run: %v", perr.Error())
	}
	fmt.Println("ProtectionGroupID:", result.ProtectionGroupID)

	// wait for active backup run
	var backupRun *types.BackupResult
	var totalTimeout time.Duration = types.GetBackup_totalTimeout
	var pollingInterval time.Duration = types.GetBackup_pollingInterval

	deadline := time.Now().Add(totalTimeout)
	for time.Now().Before(deadline) {
		backupRuns, sdkErr := client.TaskAPI.ListBackups(ctx, protectionGroupId)
		if sdkErr != nil {
			return "", fmt.Errorf("Failed to fetch backup runs for group Id: %v", sdkErr.Error())
		}

		if len(backupRuns) > 0 && backupRuns[0].Status == string(types.BackupRun_Status_Running) {
			backupRun = backupRuns[0]
			break
		}
		// fmt.Printf("Protection group run not found yet. Waiting %v before next attempt...\n", pollingInterval)
		time.Sleep(pollingInterval)
	}

	backup, perr := client.TaskAPI.GetBackup(ctx, backupRun.BackupID, protectionGroupId)
	if perr != nil {
		return "", fmt.Errorf("failed to create protection group: %v", perr.Error())
	}
	fmt.Println(backup.BackupID)
	fmt.Println(backup.Status)
	fmt.Println(backup.Progress)
	fmt.Println(backup.ProtectionGroupID)
	fmt.Println(backup.StartedAt)
	fmt.Println(backup.CompletedAt)

	return backupRun.BackupID, nil
}

func CreateRecovery(ctx context.Context, client *migrationv2.Client, dataSource datasources.DataSource, targetRegistrationId int64, protectionGroupId, backupRunID string) (*types.RestoreResult, error) {
	var perr *errors.SDKError

	// Common (datasource agnostic) recovery parameters
	recoveryParams := &types.RestoreParams{
		Name: "migration-recover-103",
	}

	result, perr := client.TaskAPI.RunRestore(ctx, protectionGroupId, backupRunID, targetRegistrationId, recoveryParams, dataSource)
	if perr != nil {
		return nil, fmt.Errorf("failed to create protection group: %v", perr.Error())
	}
	return result, nil
}
