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

	"github.com/IBM/ibm-backup-recovery-sdk-go/migrationv2"
	"github.com/IBM/ibm-backup-recovery-sdk-go/migrationv2/common/config"
	"github.com/IBM/ibm-backup-recovery-sdk-go/migrationv2/common/types"
	"github.com/IBM/ibm-backup-recovery-sdk-go/migrationv2/middlelevel/connectors"
)

func initializeClient(ctx context.Context) (*migrationv2.Client, error) {
	cfg := &config.Config{
		Region:          getEnv("IBM_REGION", "us-south"),
		APIKey:          getEnv("IBM_API_KEY", ""),
		BRSInstanceName: getEnv("BRS_INSTANCE_NAME", "brs-instance"),
		ResourceGroupID: getEnv("RESOURCE_GROUP_ID", "default"),
		// AccountID is required only when using metrics or activity tracker
		// AccountID:       getEnv("IBM_ACCOUNT_ID", ""),
	}

	if cfg.APIKey == "" {
		return nil, fmt.Errorf("IBM_API_KEY environment variable is required")
	}

	return migrationv2.NewClient(ctx, cfg)
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// Example 1: Basic Helm Connector Deployment
func deployBasicHelmConnector(ctx context.Context) {
	fmt.Println("\n=== Example 1: Basic Helm Connector Deployment ===")

	client, err := initializeClient(ctx)
	if err != nil {
		log.Fatalf("Failed to initialize client: %v", err)
	}

	connectorConfig := &connectors.HelmKubeConnectorConfig{
		ClusterName:           getEnv("CLUSTER_NAME", "my-cluster"),
		ContainerEndpoint:     "https://containers.cloud.ibm.com/global",
		ContainerEndpointType: "public",
		Namespace:             "brs-connector",
		ChartVersion:          getEnv("CHART_VERSION", "7.2.18-release-20260226-49768040"),
		Replicas:              1,
		ImagePullPolicy:       "Always",
		AuthConfig: &connectors.KubernetesAuthConfig{
			ApiKey:     getEnv("IBM_API_KEY", ""),
			AuthMethod: connectors.AuthMethodAPIKey,
			Host:       "icr.io",
			IamURL:     "https://iam.cloud.ibm.com",
		},
	}

	connector, err := client.ConnectorFactory.CreateConnectorDeployer(connectorConfig)
	if err != nil {
		log.Fatalf("Failed to create connector deployer: %v", err)
	}

	connParam := &types.ConnectionParams{
		Name: "basic-helm-connection",
		Type: types.ConnectionType_ROKS_CLASSIC, // User-friendly constant
	}

	connectionResult, err := client.TaskAPI.CreateConnection(ctx, connParam)
	if err != nil {
		log.Fatalf("Failed to create connection: %v", err)
	}
	fmt.Printf("✓ Connection created: %s\n", connectionResult.ConnectionID)

	connectorResult, err := client.TaskAPI.DeployConnector(ctx, connector, connectionResult)
	if err != nil {
		log.Fatalf("Failed to deploy connector: %v", err)
	}

	fmt.Printf("✓ Helm connector deployed: %s\n", connectorResult.ConnectorID)
	fmt.Printf("  Status: %s\n", connectorResult.Status)
}

// Example 2: Helm Connector with NodeSelector
func deployWithNodeSelector(ctx context.Context) {
	fmt.Println("\n=== Example 2: Helm Connector with NodeSelector ===")

	client, err := initializeClient(ctx)
	if err != nil {
		log.Fatalf("Failed to initialize client: %v", err)
	}

	connectorConfig := &connectors.HelmKubeConnectorConfig{
		ClusterName:           getEnv("CLUSTER_NAME", "my-cluster"),
		ContainerEndpoint:     "https://containers.cloud.ibm.com/global",
		ContainerEndpointType: "public",
		Namespace:             "brs-connector-nodeselector",
		ChartVersion:          getEnv("CHART_VERSION", "7.2.18-release-20260226-49768040"),
		Replicas:              1,
		ImagePullPolicy:       "Always",

		// NodeSelector: Only run on nodes with this label
		NodeSelector: map[string]string{
			"dedicated": "data-source-connector",
		},

		AuthConfig: &connectors.KubernetesAuthConfig{
			ApiKey:     getEnv("IBM_API_KEY", ""),
			AuthMethod: connectors.AuthMethodAPIKey,
			Host:       "icr.io",
			IamURL:     "https://iam.cloud.ibm.com",
		},
	}

	fmt.Printf("  NodeSelector: %v\n", connectorConfig.NodeSelector)

	connector, err := client.ConnectorFactory.CreateConnectorDeployer(connectorConfig)
	if err != nil {
		log.Fatalf("Failed to create connector deployer: %v", err)
	}

	connParam := &types.ConnectionParams{
		Name: "nodeselector-connection",
		Type: types.ConnectionType_ROKS_CLASSIC, // User-friendly constant
	}

	connectionResult, err := client.TaskAPI.CreateConnection(ctx, connParam)
	if err != nil {
		log.Fatalf("Failed to create connection: %v", err)
	}

	connectorResult, err := client.TaskAPI.DeployConnector(ctx, connector, connectionResult)
	if err != nil {
		log.Fatalf("Failed to deploy connector: %v", err)
	}

	fmt.Printf("✓ Connector deployed with NodeSelector: %s\n", connectorResult.ConnectorID)
}

// Example 3: Helm Connector with Tolerations
func deployWithTolerations(ctx context.Context) {
	fmt.Println("\n=== Example 3: Helm Connector with Tolerations + NodeSelector ===")

	client, err := initializeClient(ctx)
	if err != nil {
		log.Fatalf("Failed to initialize client: %v", err)
	}

	connectorConfig := &connectors.HelmKubeConnectorConfig{
		ClusterName:           getEnv("CLUSTER_NAME", "my-cluster"),
		ContainerEndpoint:     "https://containers.cloud.ibm.com/global",
		ContainerEndpointType: "public",
		Namespace:             "brs-connector-tolerations",
		ChartVersion:          getEnv("CHART_VERSION", "7.2.18-release-20260226-49768040"),
		Replicas:              1,
		ImagePullPolicy:       "Always",

		NodeSelector: map[string]string{
			"dedicated": "data-source-connector",
		},

		// Tolerations: Allow scheduling on tainted nodes
		Tolerations: []map[string]interface{}{
			{
				"key":      "dedicated",
				"operator": "Equal",
				"value":    "data-source-connector",
				"effect":   "NoSchedule",
			},
		},

		AuthConfig: &connectors.KubernetesAuthConfig{
			ApiKey:     getEnv("IBM_API_KEY", ""),
			AuthMethod: connectors.AuthMethodAPIKey,
			Host:       "icr.io",
			IamURL:     "https://iam.cloud.ibm.com",
		},
	}

	fmt.Printf("  NodeSelector: %v\n", connectorConfig.NodeSelector)
	fmt.Printf("  Tolerations: %d configured\n", len(connectorConfig.Tolerations))

	connector, err := client.ConnectorFactory.CreateConnectorDeployer(connectorConfig)
	if err != nil {
		log.Fatalf("Failed to create connector deployer: %v", err)
	}

	connParam := &types.ConnectionParams{
		Name: "tolerations-connection",
		Type: types.ConnectionType_ROKS_CLASSIC, // User-friendly constant
	}

	connectionResult, err := client.TaskAPI.CreateConnection(ctx, connParam)
	if err != nil {
		log.Fatalf("Failed to create connection: %v", err)
	}

	connectorResult, err := client.TaskAPI.DeployConnector(ctx, connector, connectionResult)
	if err != nil {
		log.Fatalf("Failed to deploy connector: %v", err)
	}

	fmt.Printf("✓ Connector deployed with Tolerations: %s\n", connectorResult.ConnectorID)
}

// Example 4: Helm Connector with Resources
func deployWithResources(ctx context.Context) {
	fmt.Println("\n=== Example 4: Helm Connector with Resources ===")

	client, err := initializeClient(ctx)
	if err != nil {
		log.Fatalf("Failed to initialize client: %v", err)
	}

	connectorConfig := &connectors.HelmKubeConnectorConfig{
		ClusterName:           getEnv("CLUSTER_NAME", "my-cluster"),
		ContainerEndpoint:     "https://containers.cloud.ibm.com/global",
		ContainerEndpointType: "public",
		Namespace:             "brs-connector-resources",
		ChartVersion:          getEnv("CHART_VERSION", "7.2.18-release-20260226-49768040"),
		Replicas:              1,
		ImagePullPolicy:       "Always",

		NodeSelector: map[string]string{
			"dedicated": "data-source-connector",
		},

		Tolerations: []map[string]interface{}{
			{
				"key":      "dedicated",
				"operator": "Equal",
				"value":    "data-source-connector",
				"effect":   "NoSchedule",
			},
		},

		// Resources: CPU and Memory limits
		Resources: &connectors.ResourceRequirements{
			Requests: &connectors.ResourceList{
				CPU:    "1",
				Memory: "2Gi",
			},
			Limits: &connectors.ResourceList{
				CPU:    "2",
				Memory: "4Gi",
			},
		},

		AuthConfig: &connectors.KubernetesAuthConfig{
			ApiKey:     getEnv("IBM_API_KEY", ""),
			AuthMethod: connectors.AuthMethodAPIKey,
			Host:       "icr.io",
			IamURL:     "https://iam.cloud.ibm.com",
		},
	}

	fmt.Printf("  NodeSelector: %v\n", connectorConfig.NodeSelector)
	fmt.Printf("  Tolerations: %d configured\n", len(connectorConfig.Tolerations))
	fmt.Printf("  Resources: Requests(CPU=%s, Memory=%s), Limits(CPU=%s, Memory=%s)\n",
		connectorConfig.Resources.Requests.CPU,
		connectorConfig.Resources.Requests.Memory,
		connectorConfig.Resources.Limits.CPU,
		connectorConfig.Resources.Limits.Memory)

	connector, err := client.ConnectorFactory.CreateConnectorDeployer(connectorConfig)
	if err != nil {
		log.Fatalf("Failed to create connector deployer: %v", err)
	}

	connParam := &types.ConnectionParams{
		Name: "resources-connection",
		Type: types.ConnectionType_ROKS_CLASSIC, // User-friendly constant
	}

	connectionResult, err := client.TaskAPI.CreateConnection(ctx, connParam)
	if err != nil {
		log.Fatalf("Failed to create connection: %v", err)
	}

	connectorResult, err := client.TaskAPI.DeployConnector(ctx, connector, connectionResult)
	if err != nil {
		log.Fatalf("Failed to deploy connector: %v", err)
	}

	fmt.Printf("✓ Connector deployed with Resources: %s\n", connectorResult.ConnectorID)
}

// Example 5: VPC VSI with Agent Connector
func deployVSIWithSSHKey(ctx context.Context) {
	fmt.Println("\n=== Example 5: VPC VSI with Agent Connector (SSH Key Auth) ===")

	client, err := initializeClient(ctx)
	if err != nil {
		log.Fatalf("Failed to initialize client: %v", err)
	}

	connectorConfig := &connectors.VsiConnectorConfig{
		AuthConfig: &connectors.VSIAuthConfig{
			AuthMethod:     connectors.AuthMethodSSHKey,
			Host:           "169.48.123.45",
			Port:           22,
			User:           "ubuntu",
			PrivateKeyPath: "/home/user/.ssh/id_rsa",
			ConnectTimeout: 30,
			KeepAlive:      60,
			StrictHostKey:  true,
			RequireSudo:    true,
		},
	}

	connector, err := client.ConnectorFactory.CreateConnectorDeployer(connectorConfig)
	if err != nil {
		log.Fatalf("Failed to create connector deployer: %v", err)
	}

	connParam := &types.ConnectionParams{
		Name: "vsi-agent-connection",
		Type: types.ConnectionType_VSI, // User-friendly constant
	}

	connectionResult, err := client.TaskAPI.CreateConnection(ctx, connParam)
	if err != nil {
		log.Fatalf("Failed to create connection: %v", err)
	}

	connectorResult, err := client.TaskAPI.DeployConnector(ctx, connector, connectionResult)
	if err != nil {
		log.Fatalf("Failed to deploy connector: %v", err)
	}

	fmt.Printf("✓ Agent connector deployed: %s\n", connectorResult.ConnectorID)
	fmt.Printf("  Status: %s\n", connectorResult.Status)
}

// Example 6: Check Supported Connector Types
func checkSupportedConnectorTypes() {
	fmt.Println("\n=== Example 6: Check Supported Connector Types ===")

	k8sTypes := connectors.GetSupportedTypesForDataSource("kubernetes")
	fmt.Printf("Kubernetes supported connector types: %v\n", k8sTypes)

	defaultK8sType := connectors.GetDefaultConnectorType("kubernetes")
	fmt.Printf("Kubernetes default connector type: %s\n", defaultK8sType)

	vsiTypes := connectors.GetSupportedTypesForDataSource("vpcvsi")
	fmt.Printf("VPC VSI supported connector types: %v\n", vsiTypes)

	defaultVSIType := connectors.GetDefaultConnectorType("vpcvsi")
	fmt.Printf("VPC VSI default connector type: %s\n", defaultVSIType)
}

// Example 7: Validate Authentication Configuration
func validateAuthConfigs() {
	fmt.Println("\n=== Example 7: Validate Authentication Configuration ===")

	k8sAuth1 := &connectors.KubernetesAuthConfig{
		AuthMethod:     connectors.AuthMethodKubeconfig,
		KubeconfigPath: "/path/to/kubeconfig",
	}
	if err := k8sAuth1.Validate(); err != nil {
		fmt.Printf("✗ Kubeconfig auth validation failed: %v\n", err)
	} else {
		fmt.Printf("✓ Kubeconfig auth is valid\n")
	}

	k8sAuth2 := &connectors.KubernetesAuthConfig{
		AuthMethod: connectors.AuthMethodToken,
		APIServer:  "https://kubernetes.example.com:6443",
	}
	if err := k8sAuth2.Validate(); err != nil {
		fmt.Printf("✗ Token auth validation failed (expected): %v\n", err)
	}

	vsiAuth1 := &connectors.VSIAuthConfig{
		AuthMethod:     connectors.AuthMethodSSHKey,
		Host:           "169.48.123.45",
		User:           "ubuntu",
		PrivateKeyPath: "/path/to/key",
	}
	if err := vsiAuth1.Validate(); err != nil {
		fmt.Printf("✗ SSH key auth validation failed: %v\n", err)
	} else {
		fmt.Printf("✓ SSH key auth is valid\n")
	}

	vsiAuth2 := &connectors.VSIAuthConfig{
		AuthMethod: connectors.AuthMethodSSHPassword,
		Host:       "169.48.123.45",
		User:       "ubuntu",
	}
	if err := vsiAuth2.Validate(); err != nil {
		fmt.Printf("✗ SSH password auth validation failed (expected): %v\n", err)
	}
}

func main() {
	fmt.Println("==============================================")
	fmt.Println("Connector Deployment Examples")
	fmt.Println("==============================================")

	// Run validation examples (don't require API key)
	checkSupportedConnectorTypes()
	validateAuthConfigs()

	// Check if API key is set for deployment examples
	apiKey := os.Getenv("IBM_API_KEY")
	if apiKey == "" {
		fmt.Println("\n⚠️  IBM_API_KEY not set. Skipping deployment examples.")
		fmt.Println("To run deployment examples, set:")
		fmt.Println("  export IBM_API_KEY='your-api-key'")
		fmt.Println("  export CLUSTER_NAME='your-cluster-name'")
		fmt.Println("  export CHART_VERSION='7.2.18-release-20260226-49768040'")
		fmt.Println("\n==============================================")
		fmt.Println("Examples completed!")
		fmt.Println("==============================================")
		return
	}

	// Uncomment the examples you want to run:
	// ctx := context.Background()
	// deployBasicHelmConnector(ctx)
	// deployWithNodeSelector(ctx)
	// deployWithTolerations(ctx)
	// deployWithResources(ctx)
	// deployVSIWithSSHKey(ctx)

	fmt.Println("\n==============================================")
	fmt.Println("Examples completed!")
	fmt.Println("==============================================")
}
