/*******************************************************************************
 * IBM Confidential
 * OCO Source Materials
 * (C) Copyright IBM Corp. 2026 All Rights Reserved.
 * The source code for this program is not  published or otherwise divested of
 * its trade secrets, irrespective of what has been deposited with
 * the U.S. Copyright Office.
 ******************************************************************************/

// Complete Task-Based API Flow Example
// Demonstrates: Connection → Connector → Registration → Policy → Protection Group → Backup → Restore
// With logging, metrics, and activity tracking enabled
package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/IBM/go-sdk-core/v5/core"
	"github.com/IBM/ibm-backup-recovery-sdk-go/migrationv2"
	activity_tracker "github.com/IBM/ibm-backup-recovery-sdk-go/migrationv2/common/activity-tracker"
	"github.com/IBM/ibm-backup-recovery-sdk-go/migrationv2/common/config"
	"github.com/IBM/ibm-backup-recovery-sdk-go/migrationv2/common/logger"
	"github.com/IBM/ibm-backup-recovery-sdk-go/migrationv2/common/metrics"
	"github.com/IBM/ibm-backup-recovery-sdk-go/migrationv2/common/types"
	"github.com/IBM/ibm-backup-recovery-sdk-go/migrationv2/middlelevel/connectors"
	"github.com/IBM/ibm-backup-recovery-sdk-go/migrationv2/middlelevel/datasources/kubernetes"
)

// Configuration constants - Replace these with your actual values
const (
	IBM_API_KEY       = "your-ibm-cloud-api-key"
	IBM_ACCOUNT_ID    = "your-ibm-account-id" // Required when using metrics or activity tracker
	IBM_REGION        = "us-south"
	BRS_INSTANCE_CRN  = "crn:v1:bluemix:public:backup-recovery:us-south:a/your-account-id:instance-id::"
	RESOURCE_GROUP_ID = "your-resource-group-id"
	BRS_TENANT_ID     = "your-tenant-id"

	CLUSTER_NAME     = "my-iks-cluster"
	CLUSTER_ENDPOINT = "https://c100.us-south.containers.cloud.ibm.com:12345"
	CLUSTER_ID       = "your-cluster-id"

	POLICY_NAME         = "demo-daily-backup-policy"
	PROTECTION_GROUP    = "demo-k8s-protection-group"
	CONNECTION_NAME     = "demo-k8s-connection"
	NAMESPACE_TO_BACKUP = "default"
	METRICS_PORT        = 9090
)

func main() {
	ctx := context.Background()

	client, metricsServer, log := initializeClient(ctx)
	defer metricsServer.Close()

	connection, err := createConnection(ctx, client)
	if err != nil {
		os.Exit(1)
	}

	_, err = deployConnector(ctx, client, connection)
	if err != nil {
		os.Exit(1)
	}

	registration, dataSource, err := registerDataSource(ctx, client, connection.ConnectionID)
	if err != nil {
		os.Exit(1)
	}

	policy, err := createProtectionPolicy(ctx, client)
	if err != nil {
		os.Exit(1)
	}

	protectionGroup, err := createProtectionGroup(ctx, client, dataSource, registration.RegistrationID, policy.ID)
	if err != nil {
		os.Exit(1)
	}

	backupID, err := runBackup(ctx, client, protectionGroup.ProtectionGroupID)
	if err != nil {
		os.Exit(1)
	}

	err = waitForBackupCompletion(ctx, client, protectionGroup.ProtectionGroupID, backupID)
	if err != nil {
		os.Exit(1)
	}

	restoreID, err := runRestore(ctx, client, dataSource, registration.RegistrationID, protectionGroup.ProtectionGroupID, backupID)
	if err != nil {
		os.Exit(1)
	}

	err = monitorRestore(ctx, client, restoreID)
	if err != nil {
		os.Exit(1)
	}

	log.Info(ctx, "✅ Workflow completed successfully")
}

func initializeClient(ctx context.Context) (*migrationv2.Client, *http.Server, logger.Logger) {
	log := logger.New(logger.Config{
		Level:             "info",
		Format:            "json",
		ServiceName:       "brs-complete-demo",
		Environment:       "development",
		IncludeStackTrace: true,
	})

	cfg := config.DefaultConfig().
		WithRegion(IBM_REGION).
		WithAPIKey(IBM_API_KEY).
		WithBRSInstanceCRN(BRS_INSTANCE_CRN).
		WithResourceGroupID(RESOURCE_GROUP_ID).
		WithAccountID(IBM_ACCOUNT_ID).
		WithLogger(log).
		WithTaskAPI(true).
		WithWorkflowAPI(false).
		WithMetricsEnabled(true).
		WithPrometheusConfig(&metrics.PrometheusConfig{
			Namespace: "brs_demo",
			Logger:    log,
		}).
		WithActivityTrackerEnabled(true).
		WithActivityTrackerConfig(&activity_tracker.HTTPSinkConfig{
			IngestionEndpoint: "https://your-activity-tracker-endpoint.com",
			// IAMAuthenticator is populated automatically from cfg.APIKey.
		})
	// TenantId is optional – set only when you know the exact value.
	cfg.TenantId = BRS_TENANT_ID

	client, err := migrationv2.NewClient(ctx, cfg)
	if err != nil {
		log.Error(ctx, "Failed to create client", "error", err)
		os.Exit(1)
	}

	// Expose the metrics instance that the SDK already built internally.
	metricsCollector := cfg.GetMetrics().(*metrics.PrometheusMetrics)
	metricsServer := &http.Server{
		Addr:    fmt.Sprintf(":%d", METRICS_PORT),
		Handler: metricsCollector.Handler(),
	}
	go func() {
		log.Info(ctx, "Starting metrics server", "port", METRICS_PORT)
		if err := metricsServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Error(ctx, "Metrics server error", "error", err)
		}
	}()

	log.Info(ctx, "SDK client initialized", "region", IBM_REGION)
	return client, metricsServer, log
}

func createConnection(ctx context.Context, client *migrationv2.Client) (*types.ConnectionResult, error) {
	connectionParams := &types.ConnectionParams{
		Name: CONNECTION_NAME,
		Type: types.ConnectionType_IKS_VPC, // User-friendly constant
	}

	connection, sdkErr := client.TaskAPI.CreateConnection(ctx, connectionParams)
	if sdkErr != nil {
		return nil, fmt.Errorf("create connection failed: %s", sdkErr.Message)
	}

	return connection, nil
}

func deployConnector(ctx context.Context, client *migrationv2.Client, connection *types.ConnectionResult) (*connectors.ConnectorResult, error) {
	helmConfig := &connectors.HelmKubeConnectorConfig{
		Namespace:             "ibm-backup-recovery",
		ClusterName:           CLUSTER_NAME,
		ContainerEndpoint:     "https://containers.cloud.ibm.com/global",
		ContainerEndpointType: "public",
		Replicas:              1,
		ChartVersion:          "latest",
		ReleaseName:           "brs-connector",
		ChartName:             "ibm-backup-recovery-agent",
		ChartReference:        "oci://icr.io/ext/brs/brs-ds-connector-chart",
		WaitTillDeploy:        true,
		ConnectorType:         connectors.ConnectorTypeHelm,
		AuthConfig: &connectors.KubernetesAuthConfig{
			AuthMethod: connectors.AuthMethodAPIKey,
			ApiKey:     IBM_API_KEY,
			IamURL:     "https://iam.cloud.ibm.com",
		},
		//BRSconfig -- Private variable
	}
	// Create connector deployer from config
	helmConnector, err := helmConfig.CreateConnectorDeployer()
	if err != nil {
		return nil, fmt.Errorf("create connector deployer failed: %v", err)
	}

	connector, sdkErr := client.TaskAPI.DeployConnector(ctx, helmConnector, connection)
	if sdkErr != nil {
		return nil, fmt.Errorf("deploy connector failed: %s", sdkErr.Message)
	}

	return connector, nil
}

func registerDataSource(ctx context.Context, client *migrationv2.Client, connectionID string) (*types.RegistrationResult, *kubernetes.KubernetesDataSource, error) {
	dataSourceConfig := &kubernetes.KubernetesDataSourceConfig{
		Authenticator: &core.IamAuthenticator{
			ApiKey: IBM_API_KEY,
			URL:    "https://iam.cloud.ibm.com",
		},
		ClusterName:           CLUSTER_NAME,
		ClusterType:           "IKS",
		ClusterEndpoint:       CLUSTER_ENDPOINT,
		ContainerEndpoint:     "https://containers.cloud.ibm.com/global",
		ContainerEndpointType: "public",
		KubernetesProtectionParams: &types.KubernetesProtectionParams{
			IncludeNamespaces: NAMESPACE_TO_BACKUP,
			ExcludeNamespaces: "",
			FailOnHookError:   false,
			Settings:          &types.ProtectionSetting{CSISnapshot: true},
		},
		KubernetesRestoreParams: &types.KubernetesRestoreParams{
			RecoverObjectSpec: &types.RecoverObjectSpec{
				RestoreOnlyPvc:         false,
				UseStorageClassMapping: core.BoolPtr(true),
			},
			RenameRecoveredNamespacesParams: &types.RenameRecoveredNamespacesParams{
				Prefix: "restored-",
				Suffix: "-backup",
			},
			RecoverToNewTarget: false,
		},
	}

	dataSource, err := client.CreateDataSource(dataSourceConfig)
	if err != nil {
		return nil, nil, fmt.Errorf("create data source failed: %v", err)
	}

	registration, sdkErr := client.TaskAPI.RegisterSource(ctx, dataSource, connectionID)
	if sdkErr != nil {
		return nil, nil, fmt.Errorf("register source failed: %s", sdkErr.Message)
	}

	return registration, dataSource.(*kubernetes.KubernetesDataSource), nil
}

func createProtectionPolicy(ctx context.Context, client *migrationv2.Client) (*types.PolicyResult, error) {
	policyParams := &types.PolicyParams{
		Name:        POLICY_NAME,
		Description: "Daily backup policy for Kubernetes workloads - Demo",
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
			NumberOfRetry:       3,
			WaitInCaseOfFailure: 5,
		},
	}

	policy, sdkErr := client.TaskAPI.CreatePolicy(ctx, policyParams)
	if sdkErr != nil {
		return nil, fmt.Errorf("create policy failed: %s", sdkErr.Message)
	}

	return policy, nil
}

func createProtectionGroup(ctx context.Context, client *migrationv2.Client, dataSource *kubernetes.KubernetesDataSource, registrationID int64, policyID string) (*types.ProtectionGroupResult, error) {
	groupParams := &types.ProtectionGroupParams{
		Name:            PROTECTION_GROUP,
		NumberOfBackups: 1,
		Policy: &types.Policy{
			ID: policyID,
		},
	}

	group, sdkErr := client.TaskAPI.CreateProtectionGroup(ctx, registrationID, groupParams, dataSource)
	if sdkErr != nil {
		return nil, fmt.Errorf("create protection group failed: %s", sdkErr.Message)
	}

	return group, nil
}

func runBackup(ctx context.Context, client *migrationv2.Client, protectionGroupID string) (string, error) {
	backupParams := &types.BackupParams{
		BackupType: types.BackupType_Full,
	}

	_, sdkErr := client.TaskAPI.RunBackup(ctx, protectionGroupID, backupParams)
	if sdkErr != nil {
		return "", fmt.Errorf("run backup failed: %s", sdkErr.Message)
	}

	var backupID string
	maxWait := 2 * time.Minute
	pollInterval := 5 * time.Second
	deadline := time.Now().Add(maxWait)

	for time.Now().Before(deadline) {
		backups, err := client.TaskAPI.ListBackups(ctx, protectionGroupID)
		if err == nil && len(backups) > 0 {
			backupID = backups[0].BackupID
			break
		}
		time.Sleep(pollInterval)
	}

	if backupID == "" {
		return "", fmt.Errorf("backup did not start within %v", maxWait)
	}

	return backupID, nil
}

// GetBackup response shape (types.BackupResult):
//
//	{
//	  "backupId":          "159354:1788883592073654",
//	  "protectionGroupId": "2712860000048009:1757348677013:159354",
//	  "status":            "Succeeded",          // Running | Succeeded | Failed | Canceled
//	  "progress":          100.0,                // *float32, nil when progress API has no data
//	  "startedAt":         "2026-09-07T13:06:32Z",
//	  "completedAt":       "2026-09-07T13:08:04Z", // *time.Time, nil while still running
//	  "namespaceProgress": [                     // populated from progress API while running
//	    {
//	      "namespaceName": "e2e-app-nginx",
//	      "status":        "Finished",           // Active | Finished | FinishedWithError
//	      "progress":      100                   // int 0-100; forced to 100 on Finished/Succeeded
//	    },
//	    {
//	      "namespaceName": "e2e-app-busybox",
//	      "status":        "Finished",
//	      "progress":      100
//	    }
//	  ]
//	}
//
// Notes:
//   - progress is *float32 — always check for nil before dereferencing
//   - completedAt is *time.Time — nil while the run is still active
//   - namespaceProgress is empty for completed runs once the progress API
//     window closes (404); poll GetBackup while running to capture it
//   - For cloud-archival-direct (CAD) Kubernetes runs the API always returns
//     objects=null; namespace data comes from the progress API instead
func waitForBackupCompletion(ctx context.Context, client *migrationv2.Client, protectionGroupID, backupID string) error {
	maxWait := 30 * time.Minute
	pollInterval := 10 * time.Second
	deadline := time.Now().Add(maxWait)

	for time.Now().Before(deadline) {
		backup, sdkErr := client.TaskAPI.GetBackup(ctx, backupID, protectionGroupID)
		if sdkErr != nil {
			time.Sleep(pollInterval)
			continue
		}

		if backup.Status == "Succeeded" || backup.Status == "succeeded" {
			return nil
		} else if backup.Status == "Failed" || backup.Status == "failed" {
			return fmt.Errorf("backup failed")
		}

		time.Sleep(pollInterval)
	}

	return fmt.Errorf("backup did not complete within %v", maxWait)
}

func runRestore(ctx context.Context, client *migrationv2.Client, dataSource *kubernetes.KubernetesDataSource, targetRegistrationID int64, protectionGroupID, backupID string) (string, error) {
	restoreParams := &types.RestoreParams{
		Name: fmt.Sprintf("restore-%s-%d", PROTECTION_GROUP, time.Now().Unix()),
	}

	result, sdkErr := client.TaskAPI.RunRestore(ctx, protectionGroupID, backupID, targetRegistrationID, restoreParams, dataSource)
	if sdkErr != nil {
		return "", fmt.Errorf("run restore failed: %s", sdkErr.Message)
	}

	return result.RestoreID, nil
}

// GetRestore response shape (types.RestoreResult):
//
//	{
//	  "restoreId": "2712860000048009:1757348677013:159360",
//	  "backupId":  "159354:1788883592073654",
//	  "targetId":  6047,
//	  "status":    "Succeeded",           // Accepted | Running | Succeeded | Failed | Canceled
//	  "progress":  100,                   // int 0-100, averaged across all namespace progress values
//	  "messages":  ["restored 2 PVCs"],   // top-level messages from the API (may be empty)
//	  "startedAt": "2026-09-07T13:09:00Z",
//	  "completedAt": "2026-09-07T13:18:30Z", // *time.Time, nil while still running
//	  "namespaceProgress": [
//	    {
//	      "namespaceName": "e2e-app-nginx",
//	      "status":        "Succeeded",    // Running | Succeeded | Failed
//	      "progress":      100,            // int 0-100 from ProgressMonitors API
//	      "messages":      ["restored 1 PVC"]  // per-namespace messages (may be empty)
//	    },
//	    {
//	      "namespaceName": "e2e-app-busybox",
//	      "status":        "Succeeded",
//	      "progress":      100,
//	      "messages":      []
//	    }
//	  ]
//	}
//
// Notes:
//   - progress is an int averaged across all namespace progress values
//   - completedAt is *time.Time — nil while the restore is still active
//   - messages (top-level) comes from the BRS Recovery API response
//   - namespaceProgress[].messages comes from KubernetesRecoveryObjectParams.Messages
func monitorRestore(ctx context.Context, client *migrationv2.Client, restoreID string) error {
	maxWait := 30 * time.Minute
	pollInterval := 10 * time.Second
	deadline := time.Now().Add(maxWait)

	for time.Now().Before(deadline) {
		restore, sdkErr := client.TaskAPI.GetRestore(ctx, restoreID)
		if sdkErr != nil {
			time.Sleep(pollInterval)
			continue
		}

		if restore.Status == "Succeeded" || restore.Status == "succeeded" {
			return nil
		} else if restore.Status == "Failed" || restore.Status == "failed" {
			return fmt.Errorf("restore failed")
		}

		time.Sleep(pollInterval)
	}

	return fmt.Errorf("restore did not complete within %v", maxWait)
}

func convertToPointer[T any](v T) *T {
	return &v
}
