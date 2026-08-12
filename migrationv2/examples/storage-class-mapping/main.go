/*******************************************************************************
 * IBM Confidential
 * OCO Source Materials
 * (C) Copyright IBM Corp. 2026 All Rights Reserved.
 * The source code for this program is not  published or otherwise divested of
 * its trade secrets, irrespective of what has been deposited with
 * the U.S. Copyright Office.
 ******************************************************************************/

// Storage-Class Mapping Example
//
// Demonstrates all storage-class mapping (SCM) scenarios available in the
// IBM Backup & Recovery SDK when restoring Kubernetes workloads across
// clusters (e.g. Classic IKS → VPC IKS):
//
//  1. Inspect a backup run – list namespaces and their snapshot IDs.
//  2. Single global SC mapping – one old→new mapping applied to every namespace.
//  3. Multiple global SC mappings – several old→new pairs in one spec.
//  4. Namespace filter + SC mapping – restore only selected namespaces.
//  5. Per-namespace SC override – different target SC for one specific namespace.
//  6. Combined: IncludeNamespaces + per-namespace SC override.
//  7. No SC mapping – plain restore without any storage-class remapping.
//
// Required environment variables (set before running):
//
//	IBM_API_KEY            – IBM Cloud API key
//	IBM_REGION             – e.g. us-south
//	BRS_INSTANCE_NAME      – BRS service instance name
//	RESOURCE_GROUP_ID      – IBM Cloud resource group ID
//	GROUP_ID               – Protection group ID of an existing backup
//	BACKUP_ID              – Backup run ID to restore from
//	VPC_CONN_NAME          – Name of the registered VPC connection
//	VPC_CLUSTER_NAME       – Name of the target VPC cluster
package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/IBM/go-sdk-core/v5/core"
	"github.com/IBM/ibm-backup-recovery-sdk-go/migrationv2"
	"github.com/IBM/ibm-backup-recovery-sdk-go/migrationv2/common/config"
	"github.com/IBM/ibm-backup-recovery-sdk-go/migrationv2/common/types"
	brsK8s "github.com/IBM/ibm-backup-recovery-sdk-go/migrationv2/middlelevel/datasources/kubernetes"
)

// ─── configuration ──────────────────────────────────────────────────────────

const (
	// Source cluster (Classic IKS) — used to call GetBackupRunNamespaceSnapshots.
	classicClusterName = "sdk-classic-testing-iks"
	classicClusterType = "IKS"

	// Target cluster (VPC IKS) — used for all restore scenarios.
	vpcClusterType = "IKS"

	containerEndpoint     = "https://containers.cloud.ibm.com/global"
	containerEndpointType = "public"

	// Default storage-class pair (Classic block → VPC block).
	scOld = "ibmc-block-silver"
	scNew = "ibmc-vpc-block-5iops-tier"
)

func main() {
	ctx := context.Background()
	client := initClient(ctx)

	groupID := mustEnv("GROUP_ID")
	backupID := mustEnv("BACKUP_ID")
	vpcConnName := getEnv("VPC_CONN_NAME", "migration-vpc-conn")
	vpcCluster := mustEnv("VPC_CLUSTER_NAME")
	apiKey := mustEnv("IBM_API_KEY")

	// ── 1. Inspect the backup run ────────────────────────────────────────────
	fmt.Println("\n=== 1. Inspect backup run: namespace→snapshot mappings ===")
	inspectBackupRun(ctx, client, apiKey, groupID, backupID)

	// Resolve the VPC connection and its registration — used by all restores.
	vpcConn, sdkErr := client.TaskAPI.GetConnectionByName(ctx, vpcConnName)
	if sdkErr != nil {
		log.Fatalf("GetConnectionByName(%s) failed: %s", vpcConnName, sdkErr.Message)
	}
	vpcReg, sdkErr := client.TaskAPI.GetRegistrationByConnection(ctx, vpcConn.ConnectionID)
	if sdkErr != nil {
		log.Fatalf("GetRegistrationByConnection failed: %s", sdkErr.Message)
	}
	targetRegID := vpcReg.RegistrationID

	// ── 2. Single global SC mapping ──────────────────────────────────────────
	fmt.Println("\n=== 2. Restore: single global storage-class mapping ===")
	restoreSingleSCMapping(ctx, client, apiKey, vpcCluster, groupID, backupID, targetRegID)

	// ── 3. Multiple global SC mappings ───────────────────────────────────────
	fmt.Println("\n=== 3. Restore: multiple global storage-class mappings ===")
	restoreMultipleSCMappings(ctx, client, apiKey, vpcCluster, groupID, backupID, targetRegID)

	// ── 4. Namespace filter + SC mapping ─────────────────────────────────────
	fmt.Println("\n=== 4. Restore: namespace filter with SC mapping ===")
	restoreNamespaceFilter(ctx, client, apiKey, vpcCluster, groupID, backupID, targetRegID)

	// ── 5. Per-namespace SC override ─────────────────────────────────────────
	fmt.Println("\n=== 5. Restore: per-namespace SC override ===")
	restorePerNamespaceOverride(ctx, client, apiKey, vpcCluster, groupID, backupID, targetRegID)

	// ── 6. Combined: IncludeNamespaces + per-namespace override ──────────────
	fmt.Println("\n=== 6. Restore: IncludeNamespaces + per-namespace SC override combined ===")
	restoreCombined(ctx, client, apiKey, vpcCluster, groupID, backupID, targetRegID)

	// ── 7. No SC mapping ─────────────────────────────────────────────────────
	fmt.Println("\n=== 7. Restore: no storage-class mapping ===")
	restoreNoSCMapping(ctx, client, apiKey, vpcCluster, groupID, backupID, targetRegID)

	fmt.Println("\n✅ All storage-class mapping examples completed.")
}

// ─── 1. Inspect backup run ───────────────────────────────────────────────────

// inspectBackupRun lists each namespace in the backup run together with its
// snapshot ID.  It also shows how to obtain a plain []string of snapshot IDs
// via GetBackupRunSnapShotID (useful when you don't need namespace metadata).
func inspectBackupRun(ctx context.Context, client *migrationv2.Client, apiKey, groupID, backupID string) {
	raw, err := client.CreateDataSource(&brsK8s.KubernetesDataSourceConfig{
		ClusterName:           classicClusterName,
		ClusterType:           classicClusterType,
		ContainerEndpoint:     containerEndpoint,
		ContainerEndpointType: containerEndpointType,
		Authenticator:         iamAuth(apiKey),
	})
	if err != nil {
		log.Printf("  CreateDataSource (source) error: %v", err)
		return
	}
	ds, ok := raw.(*brsK8s.KubernetesDataSource)
	if !ok {
		log.Printf("  unexpected data source type")
		return
	}

	// Detailed namespace→snapshot mappings.
	mappings, err := ds.GetBackupRunNamespaceSnapshots(groupID, backupID)
	if err != nil {
		log.Printf("  GetBackupRunNamespaceSnapshots error: %v", err)
		return
	}
	fmt.Printf("  Namespaces in backup run %s:\n", backupID)
	for i, m := range mappings {
		fmt.Printf("    [%d] namespace=%-30s  snapshotID=%s  namespaceID=%d\n",
			i+1, m.NamespaceName, m.SnapshotID, m.NamespaceID)
	}

	// Convenience: just snapshot IDs.
	ids, err := ds.GetBackupRunSnapShotID(groupID, backupID)
	if err != nil {
		log.Printf("  GetBackupRunSnapShotID error: %v", err)
		return
	}
	fmt.Printf("  Snapshot IDs only: %v\n", ids)
}

// ─── 2. Single global SC mapping ─────────────────────────────────────────────

// restoreSingleSCMapping restores all namespaces from the backup using one
// global storage-class mapping: scOld → scNew.
//
//	KubernetesRestoreParams.RecoverObjectSpec.StorageClasses = [{Old: scOld, New: scNew}]
func restoreSingleSCMapping(
	ctx context.Context,
	client *migrationv2.Client,
	apiKey, vpcCluster, groupID, backupID string,
	targetRegID int64,
) {
	ds, err := client.CreateDataSource(&brsK8s.KubernetesDataSourceConfig{
		ClusterName:           vpcCluster,
		ClusterType:           vpcClusterType,
		ContainerEndpoint:     containerEndpoint,
		ContainerEndpointType: containerEndpointType,
		Authenticator:         iamAuth(apiKey),
		KubernetesRestoreParams: &types.KubernetesRestoreParams{
			RecoverToNewTarget: true,
			// Single global mapping: every namespace uses this SC pair.
			RecoverObjectSpec: &types.RecoverObjectSpec{
				UseStorageClassMapping: ptr(true),
				StorageClasses: []types.StorageClassMapping{
					{Old: scOld, New: scNew},
				},
			},
		},
	})
	if err != nil {
		log.Printf("  CreateDataSource error: %v", err)
		return
	}

	result, sdkErr := client.TaskAPI.RunRestore(
		ctx,
		groupID, backupID,
		targetRegID,
		&types.RestoreParams{Name: "scm-single-sc-restore"},
		ds,
	)
	if sdkErr != nil {
		log.Printf("  RunRestore error: %s", sdkErr.Message)
		return
	}
	fmt.Printf("  restore ID: %s  (mapping: %s → %s)\n", result.RestoreID, scOld, scNew)
}

// ─── 3. Multiple global SC mappings ──────────────────────────────────────────

// restoreMultipleSCMappings demonstrates providing two storage-class pairs in
// a single RecoverObjectSpec.  Both mappings are forwarded to BRS and applied
// to every namespace in the restore.
func restoreMultipleSCMappings(
	ctx context.Context,
	client *migrationv2.Client,
	apiKey, vpcCluster, groupID, backupID string,
	targetRegID int64,
) {
	// Second mapping pair — override via env vars if needed.
	scOld2 := getEnv("SC_OLD_2", "ibmc-file-gold")
	scNew2 := getEnv("SC_NEW_2", "ibmc-vpc-block-10iops-tier")

	ds, err := client.CreateDataSource(&brsK8s.KubernetesDataSourceConfig{
		ClusterName:           vpcCluster,
		ClusterType:           vpcClusterType,
		ContainerEndpoint:     containerEndpoint,
		ContainerEndpointType: containerEndpointType,
		Authenticator:         iamAuth(apiKey),
		KubernetesRestoreParams: &types.KubernetesRestoreParams{
			RecoverToNewTarget: true,
			// Two storage-class mappings applied globally.
			RecoverObjectSpec: &types.RecoverObjectSpec{
				UseStorageClassMapping: ptr(true),
				StorageClasses: []types.StorageClassMapping{
					{Old: scOld, New: scNew},
					{Old: scOld2, New: scNew2},
				},
			},
		},
	})
	if err != nil {
		log.Printf("  CreateDataSource error: %v", err)
		return
	}

	result, sdkErr := client.TaskAPI.RunRestore(
		ctx,
		groupID, backupID,
		targetRegID,
		&types.RestoreParams{Name: "scm-multi-sc-restore"},
		ds,
	)
	if sdkErr != nil {
		log.Printf("  RunRestore error: %s", sdkErr.Message)
		return
	}
	fmt.Printf("  restore ID: %s  (mappings: %s→%s, %s→%s)\n",
		result.RestoreID, scOld, scNew, scOld2, scNew2)
}

// ─── 4. Namespace filter + SC mapping ────────────────────────────────────────

// restoreNamespaceFilter restores only the namespaces listed in
// IncludeNamespaces while still applying a global storage-class mapping.
// Namespaces not in the list are skipped entirely.
func restoreNamespaceFilter(
	ctx context.Context,
	client *migrationv2.Client,
	apiKey, vpcCluster, groupID, backupID string,
	targetRegID int64,
) {
	includeNS := getEnv("INCLUDE_NAMESPACE", "migration-test-app")
	fmt.Printf("  restoring only namespace: %s\n", includeNS)

	ds, err := client.CreateDataSource(&brsK8s.KubernetesDataSourceConfig{
		ClusterName:           vpcCluster,
		ClusterType:           vpcClusterType,
		ContainerEndpoint:     containerEndpoint,
		ContainerEndpointType: containerEndpointType,
		Authenticator:         iamAuth(apiKey),
		KubernetesRestoreParams: &types.KubernetesRestoreParams{
			RecoverToNewTarget: true,
			// Only restore this subset of namespaces.
			IncludeNamespaces: []string{includeNS},
			RecoverObjectSpec: &types.RecoverObjectSpec{
				UseStorageClassMapping: ptr(true),
				StorageClasses: []types.StorageClassMapping{
					{Old: scOld, New: scNew},
				},
			},
		},
	})
	if err != nil {
		log.Printf("  CreateDataSource error: %v", err)
		return
	}

	result, sdkErr := client.TaskAPI.RunRestore(
		ctx,
		groupID, backupID,
		targetRegID,
		&types.RestoreParams{Name: "scm-ns-filter-restore"},
		ds,
	)
	if sdkErr != nil {
		log.Printf("  RunRestore error: %s", sdkErr.Message)
		return
	}
	fmt.Printf("  restore ID: %s  (namespace filter: [%s])\n", result.RestoreID, includeNS)
}

// ─── 5. Per-namespace SC override ────────────────────────────────────────────

// restorePerNamespaceOverride sets a default global SC mapping
// (RecoverObjectSpec) and then overrides the target storage class for one
// specific namespace via RecoverMultipleObjects.
//
// Scenario:
//   - Default for all NSes : scOld → scNew
//   - Override for OVERRIDE_NAMESPACE: scOld → ibmc-vpc-block-10iops-tier
func restorePerNamespaceOverride(
	ctx context.Context,
	client *migrationv2.Client,
	apiKey, vpcCluster, groupID, backupID string,
	targetRegID int64,
) {
	overrideNS := getEnv("OVERRIDE_NAMESPACE", "migration-test-app")
	overrideSCNew := getEnv("OVERRIDE_SC_NEW", "ibmc-vpc-block-10iops-tier")
	fmt.Printf("  per-namespace override: namespace=%s  SC=%s\n", overrideNS, overrideSCNew)

	ds, err := client.CreateDataSource(&brsK8s.KubernetesDataSourceConfig{
		ClusterName:           vpcCluster,
		ClusterType:           vpcClusterType,
		ContainerEndpoint:     containerEndpoint,
		ContainerEndpointType: containerEndpointType,
		Authenticator:         iamAuth(apiKey),
		KubernetesRestoreParams: &types.KubernetesRestoreParams{
			RecoverToNewTarget: true,
			// Default mapping applied to ALL namespaces.
			RecoverObjectSpec: &types.RecoverObjectSpec{
				UseStorageClassMapping: ptr(true),
				StorageClasses: []types.StorageClassMapping{
					{Old: scOld, New: scNew},
				},
			},
			// Override for one specific namespace — different target SC.
			RecoverMultipleObjects: []types.RecoverObject{
				{
					SnapshotInfo: &types.SnapshotInfo{
						NamespaceInfo: &types.NamespaceInfo{
							Namespace: overrideNS,
						},
					},
					RecoverObjectSpec: &types.RecoverObjectSpec{
						UseStorageClassMapping: ptr(true),
						StorageClasses: []types.StorageClassMapping{
							{Old: scOld, New: overrideSCNew},
						},
					},
				},
			},
		},
	})
	if err != nil {
		log.Printf("  CreateDataSource error: %v", err)
		return
	}

	result, sdkErr := client.TaskAPI.RunRestore(
		ctx,
		groupID, backupID,
		targetRegID,
		&types.RestoreParams{Name: "scm-per-ns-override-restore"},
		ds,
	)
	if sdkErr != nil {
		log.Printf("  RunRestore error: %s", sdkErr.Message)
		return
	}
	fmt.Printf("  restore ID: %s  (override %s → %s for namespace %s)\n",
		result.RestoreID, scOld, overrideSCNew, overrideNS)
}

// ─── 6. Combined: IncludeNamespaces + per-namespace override ─────────────────

// restoreCombined shows IncludeNamespaces (filter) and RecoverMultipleObjects
// (per-NS SC override) used together in one restore call.
//
// Scenario (set via env vars):
//   - INCLUDE_NAMESPACE=app-1   → only app-1 is restored
//   - OVERRIDE_NAMESPACE=app-1  → app-1 uses overrideSCNew instead of scNew
func restoreCombined(
	ctx context.Context,
	client *migrationv2.Client,
	apiKey, vpcCluster, groupID, backupID string,
	targetRegID int64,
) {
	includeNS := getEnv("INCLUDE_NAMESPACE", "migration-test-app")
	overrideNS := getEnv("OVERRIDE_NAMESPACE", "migration-test-app")
	overrideSCNew := getEnv("OVERRIDE_SC_NEW", "ibmc-vpc-block-10iops-tier")
	fmt.Printf("  include=[%s]  override NS=%s SC=%s\n", includeNS, overrideNS, overrideSCNew)

	ds, err := client.CreateDataSource(&brsK8s.KubernetesDataSourceConfig{
		ClusterName:           vpcCluster,
		ClusterType:           vpcClusterType,
		ContainerEndpoint:     containerEndpoint,
		ContainerEndpointType: containerEndpointType,
		Authenticator:         iamAuth(apiKey),
		KubernetesRestoreParams: &types.KubernetesRestoreParams{
			RecoverToNewTarget: true,
			// Namespace filter: only restore selected namespaces.
			IncludeNamespaces: []string{includeNS},
			// Default SC mapping for all included namespaces.
			RecoverObjectSpec: &types.RecoverObjectSpec{
				UseStorageClassMapping: ptr(true),
				StorageClasses: []types.StorageClassMapping{
					{Old: scOld, New: scNew},
				},
			},
			// Override the SC for one specific namespace.
			RecoverMultipleObjects: []types.RecoverObject{
				{
					SnapshotInfo: &types.SnapshotInfo{
						NamespaceInfo: &types.NamespaceInfo{Namespace: overrideNS},
					},
					RecoverObjectSpec: &types.RecoverObjectSpec{
						UseStorageClassMapping: ptr(true),
						StorageClasses: []types.StorageClassMapping{
							{Old: scOld, New: overrideSCNew},
						},
					},
				},
			},
		},
	})
	if err != nil {
		log.Printf("  CreateDataSource error: %v", err)
		return
	}

	result, sdkErr := client.TaskAPI.RunRestore(
		ctx,
		groupID, backupID,
		targetRegID,
		&types.RestoreParams{Name: "scm-combined-restore"},
		ds,
	)
	if sdkErr != nil {
		log.Printf("  RunRestore error: %s", sdkErr.Message)
		return
	}
	fmt.Printf("  restore ID: %s  (include=[%s], override NS=%s SC=%s)\n",
		result.RestoreID, includeNS, overrideNS, overrideSCNew)
}

// ─── 7. No SC mapping ────────────────────────────────────────────────────────

// restoreNoSCMapping demonstrates a plain restore with UseStorageClassMapping
// set to false — all SCM fields are optional and can be omitted entirely.
func restoreNoSCMapping(
	ctx context.Context,
	client *migrationv2.Client,
	apiKey, vpcCluster, groupID, backupID string,
	targetRegID int64,
) {
	ds, err := client.CreateDataSource(&brsK8s.KubernetesDataSourceConfig{
		ClusterName:           vpcCluster,
		ClusterType:           vpcClusterType,
		ContainerEndpoint:     containerEndpoint,
		ContainerEndpointType: containerEndpointType,
		Authenticator:         iamAuth(apiKey),
		KubernetesRestoreParams: &types.KubernetesRestoreParams{
			RecoverToNewTarget: true,
			// Explicitly opt-out of storage-class mapping.
			RecoverObjectSpec: &types.RecoverObjectSpec{
				UseStorageClassMapping: ptr(false),
			},
		},
	})
	if err != nil {
		log.Printf("  CreateDataSource error: %v", err)
		return
	}

	result, sdkErr := client.TaskAPI.RunRestore(
		ctx,
		groupID, backupID,
		targetRegID,
		&types.RestoreParams{Name: "scm-no-sc-mapping-restore"},
		ds,
	)
	if sdkErr != nil {
		log.Printf("  RunRestore error: %s", sdkErr.Message)
		return
	}
	fmt.Printf("  restore ID: %s  (no SC mapping)\n", result.RestoreID)
}

// ─── helpers ─────────────────────────────────────────────────────────────────

// initClient creates a migrationv2.Client from environment variables.
func initClient(ctx context.Context) *migrationv2.Client {
	apiKey := mustEnv("IBM_API_KEY")
	region := mustEnv("IBM_REGION")
	brsInstance := mustEnv("BRS_INSTANCE_NAME")
	rgID := mustEnv("RESOURCE_GROUP_ID")

	cfg := config.DefaultConfig().
		WithAPIKey(apiKey).
		WithRegion(region).
		WithBRSInstanceName(brsInstance).
		WithResourceGroupID(rgID).
		WithTaskAPI(true)

	client, err := migrationv2.NewClient(ctx, cfg)
	if err != nil {
		log.Fatalf("Failed to create SDK client: %v", err)
	}
	return client
}

// iamAuth returns an IAM authenticator for the given API key.
func iamAuth(apiKey string) *core.IamAuthenticator {
	return &core.IamAuthenticator{
		ApiKey: apiKey,
		URL:    "https://iam.cloud.ibm.com",
	}
}

// ptr returns a pointer to any value (generic helper).
func ptr[T any](v T) *T { return &v }

// mustEnv returns the value of the environment variable or exits.
func mustEnv(key string) string {
	v := os.Getenv(key)
	if v == "" {
		log.Fatalf("Required environment variable %s is not set", key)
	}
	return v
}

// getEnv returns the environment variable value or the provided default.
func getEnv(key, defaultValue string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return defaultValue
}
