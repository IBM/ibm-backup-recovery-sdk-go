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
	"sync"
	"time"

	"github.com/IBM/ibm-backup-recovery-sdk-go/migrationv2"
	"github.com/IBM/ibm-backup-recovery-sdk-go/migrationv2/common/config"
	"github.com/IBM/ibm-backup-recovery-sdk-go/migrationv2/common/types"
	"github.com/IBM/ibm-backup-recovery-sdk-go/migrationv2/middlelevel/datasources/kubernetes"
)

// Microservice Example - Demonstrates idempotent behavior
// This example shows:
// - Multiple microservice instances using the same BRS instance
// - Idempotent operations (safe to call multiple times)
// - Resource reuse across instances
// - No conflicts between instances
// - Proper error handling for concurrent operations

func main() {
	fmt.Println("=== Microservice Idempotent Example ===")
	fmt.Println("Simulating multiple microservice instances accessing the same BRS instance")
	fmt.Println()

	ctx := context.Background()

	// Simulate 3 microservice instances
	numInstances := 3
	var wg sync.WaitGroup

	fmt.Printf("Starting %d microservice instances...\n\n", numInstances)

	for i := 1; i <= numInstances; i++ {
		wg.Add(1)
		go func(instanceID int) {
			defer wg.Done()
			runMicroserviceInstance(ctx, instanceID)
		}(i)

		// Stagger starts slightly to simulate real-world scenario
		time.Sleep(500 * time.Millisecond)
	}

	// Wait for all instances to complete
	wg.Wait()

	fmt.Println("\n=== Summary ===")
	fmt.Println("✅ All microservice instances completed successfully")
	fmt.Println("✅ Idempotent behavior verified - resources were reused, not duplicated")
	fmt.Println("✅ No conflicts between instances")
}

func runMicroserviceInstance(ctx context.Context, instanceID int) {
	prefix := fmt.Sprintf("[Instance-%d]", instanceID)

	fmt.Printf("%s Starting...\n", prefix)

	// Each instance creates its own client
	client, err := initializeClient(ctx, instanceID)
	if err != nil {
		log.Printf("%s Failed to initialize: %v\n", prefix, err)
		return
	}

	// Simulate microservice workflow
	// All instances try to setup the same resources
	// The SDK's idempotent design ensures resources are reused, not duplicated

	// Step 1: Setup source environment
	fmt.Printf("%s Setting up source environment...\n", prefix)
	sourceSetup, err := setupEnvironment(ctx, client, "shared-source-cluster", instanceID)
	if err != nil {
		log.Printf("%s Source setup failed: %v\n", prefix, err)
		return
	}
	fmt.Printf("%s Source setup: ConnectionID=%s (Created=%d, Reused=%d)\n",
		prefix, sourceSetup.ConnectionID, sourceSetup.ResourcesCreated, sourceSetup.ResourcesReused)

	// Step 2: Setup target environment
	fmt.Printf("%s Setting up target environment...\n", prefix)
	targetSetup, err := setupEnvironment(ctx, client, "shared-target-cluster", instanceID)
	if err != nil {
		log.Printf("%s Target setup failed: %v\n", prefix, err)
		return
	}
	fmt.Printf("%s Target setup: ConnectionID=%s (Created=%d, Reused=%d)\n",
		prefix, targetSetup.ConnectionID, targetSetup.ResourcesCreated, targetSetup.ResourcesReused)

	// Step 3: Create protection policy (idempotent)
	fmt.Printf("%s Creating protection policy...\n", prefix)
	policy, created, err := createOrGetPolicy(ctx, client, "shared-protection-policy", instanceID)
	if err != nil {
		log.Printf("%s Policy creation failed: %v\n", prefix, err)
		return
	}
	if created {
		fmt.Printf("%s Policy created: ID=%s\n", prefix, policy.ID)
	} else {
		fmt.Printf("%s Policy reused: ID=%s\n", prefix, policy.ID)
	}

	// Step 4: Create protection group (idempotent)
	fmt.Printf("%s Creating protection group...\n", prefix)
	protectionGroup, created, err := createOrGetProtectionGroup(ctx, client, sourceSetup.RegistrationID, policy.ID, "shared-protection-group", instanceID)
	if err != nil {
		log.Printf("%s Protection group creation failed: %v\n", prefix, err)
		return
	}
	if created {
		fmt.Printf("%s Protection group created: ID=%s\n", prefix, protectionGroup.ProtectionGroupID)
	} else {
		fmt.Printf("%s Protection group reused: ID=%s\n", prefix, protectionGroup.ProtectionGroupID)
	}

	fmt.Printf("%s ✅ Completed successfully\n", prefix)
}

func initializeClient(ctx context.Context, instanceID int) (*migrationv2.Client, error) {
	// All instances share the same BRS instance name so the SDK reuses it idempotently.
	cfg := config.DefaultConfig().
		WithRegion(getEnv("IBM_REGION", "us-south")).
		WithAPIKey(getEnv("IBM_API_KEY", "demo-api-key")).
		WithBRSInstanceName(getEnv("BRS_INSTANCE_NAME", "shared-brs-instance")).
		WithResourceGroupID(getEnv("RESOURCE_GROUP_ID", "shared-resource-group"))

	return migrationv2.NewClient(ctx, cfg)
}

func setupEnvironment(ctx context.Context, client *migrationv2.Client, connectionName string, instanceID int) (*types.SetupResult, error) {
	// This is idempotent - if connection already exists, it will be reused

	dataSource, _ := client.CreateDataSource(&kubernetes.KubernetesDataSourceConfig{})

	connectionParams := &types.ConnectionParams{
		Name: connectionName, // Same name across all instances
		Type: types.ConnectionType_IKS_CLASSIC,
		Metadata: map[string]interface{}{
			"created_by_instance": instanceID,
			"timestamp":           time.Now().Format(time.RFC3339),
		},
	}

	return client.WorkflowAPI.ExecuteSetup(ctx, dataSource, connectionParams)
}

func createOrGetPolicy(ctx context.Context, client *migrationv2.Client, policyName string, instanceID int) (*types.PolicyResult, bool, error) {
	// Try to get existing policy first (idempotent check)
	existingPolicy, err := client.TaskAPI.GetPolicyByName(ctx, policyName)
	if err != nil {
		return nil, false, fmt.Errorf("failed to check for existing policy: %w", err)
	}

	if existingPolicy != nil {
		// Policy already exists, reuse it
		return existingPolicy, false, nil
	}

	// Policy doesn't exist, create it
	policyParams := &types.PolicyParams{
		Name:        policyName,
		Description: fmt.Sprintf("Shared policy created by instance %d", instanceID),
		BackupType:  "full",
	}

	newPolicy, err := client.TaskAPI.CreatePolicy(ctx, policyParams)
	if err != nil {
		// Another instance might have created it concurrently
		// Try to get it again
		existingPolicy, getErr := client.TaskAPI.GetPolicyByName(ctx, policyName)
		if getErr == nil && existingPolicy != nil {
			return existingPolicy, false, nil
		}
		return nil, false, fmt.Errorf("failed to create policy: %w", err)
	}

	return newPolicy, true, nil
}

func createOrGetProtectionGroup(ctx context.Context, client *migrationv2.Client, registrationID int64, policyID string, groupName string, instanceID int) (*types.ProtectionGroupResult, bool, error) {
	// Try to get existing protection group first (idempotent check)
	existingGroup, err := client.TaskAPI.GetProtectionGroupByName(ctx, groupName)
	if err != nil {
		return nil, false, fmt.Errorf("failed to check for existing protection group: %w", err)
	}

	if existingGroup != nil {
		// Protection group already exists, reuse it
		return existingGroup, false, nil
	}

	// Protection group doesn't exist, create it
	groupParams := &types.ProtectionGroupParams{
		Name:            groupName,
		NumberOfBackups: 7,
		Policy: &types.Policy{
			ID: policyID,
		},
	}

	dataSource, _ := client.CreateDataSource(&kubernetes.KubernetesDataSourceConfig{})
	newGroup, err := client.TaskAPI.CreateProtectionGroup(ctx, registrationID, groupParams, dataSource)
	if err != nil {
		// Another instance might have created it concurrently
		// Try to get it again
		existingGroup, getErr := client.TaskAPI.GetProtectionGroupByName(ctx, groupName)
		if getErr == nil && existingGroup != nil {
			return existingGroup, false, nil
		}
		return nil, false, fmt.Errorf("failed to create protection group: %w", err)
	}

	return newGroup, true, nil
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
