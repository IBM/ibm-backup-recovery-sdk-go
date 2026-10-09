/*******************************************************************************
 * IBM Confidential
 * OCO Source Materials
 * (C) Copyright IBM Corp. 2026 All Rights Reserved.
 * The source code for this program is not  published or otherwise divested of
 * its trade secrets, irrespective of what has been deposited with
 * the U.S. Copyright Office.
 ******************************************************************************/
package connectors

import (
	"context"
	"time"

	"github.com/IBM/ibm-backup-recovery-sdk-go/migrationv2/common/types"
)

// ConnectorDeployContext carries runtime context injected by the task layer into
// a ConnectorDeployer before Deploy is called. Adding fields here in the future
// will not change the interface signature.
type ConnectorDeployContext struct {
	// BRSClient is used to call BRS platform APIs (e.g. GetConnectorMetadata).
	BRSClient types.BRSClientWrapperInterface

	// PlatformType is the raw BRS k8sPlatformType string taken from
	// ConnectionResult.Type (e.g. "kRoksVpc", "kIksClassic").
	// Used to look up the correct entry in the metadata response.
	PlatformType string
}

// ConnectorType defines the type of connector
type ConnectorType string

const (
	// Kubernetes connector types
	ConnectorTypeHelm     ConnectorType = "helm"
	ConnectorTypeOperator ConnectorType = "operator"

	// VPC VSI connector types
	ConnectorTypeAgent     ConnectorType = "vsi"
	ConnectorTypeAgentless ConnectorType = "agentless"
)

// ConnectorDeployer handles connector deployment for a specific type
// Each connector type (Helm, Operator, Agent, etc.) implements this interface
type ConnectorDeployer interface {
	// GetType returns the connector type
	GetType() ConnectorType

	// SetDeployContext injects runtime context (BRS client + platform type) so
	// the deployer can resolve live chart metadata during Deploy.
	// Must be called before Deploy.
	SetDeployContext(ctx ConnectorDeployContext)

	// Deploy deploys the connector to the target DataSource
	Deploy(ctx context.Context, registrationToken string) (*ConnectorResult, error)

	// GetStatus retrieves the current status of the connector
	GetStatus(ctx context.Context, connectorID string) (string, error)

	// Delete removes the connector from the target DataSource
	Delete(ctx context.Context, connectorID string) error
}

type ConnectorDeployConfig interface {
	// GetType returns the connector type
	GetType() ConnectorType
	// Validate validates the deployment configuration before deployment
	Validate() error
	// GetRequirements returns deployment requirements (e.g., Helm version, permissions)
	GetRequirements() *ConnectorRequirements

	// GetRegistrationToken() string

	CreateConnectorDeployer() (ConnectorDeployer, error)
}

// ConnectorDeployConfig holds connector deployment configuration
// type ConnectorDeployConfig struct {
// 	// ConnectorType specifies which type of connector to deploy
// 	ConnectorType ConnectorType `json:"connectorType"`

// 	// RegistrationToken is the token from BRS connection creation
// 	RegistrationToken string `json:"registrationToken"`

// 	// AuthConfig contains authentication configuration for accessing the datasource
// 	// For Kubernetes: KubernetesAuthConfig (kubeconfig, token, or certificate)
// 	// For VPC VSI: VSIAuthConfig (SSH key, password)
// 	AuthConfig AuthConfig `json:"authConfig"`

// 	// DataSourceConfig contains datasource-specific configuration
// 	// For Kubernetes: cluster_id, cluster_endpoint, etc.
// 	// For VPC VSI: vpc_id, vsi_ids, region, etc.
// 	// DataSourceConfig map[string]interface{} `json:"dataSourceConfig"`

// 	// DeploymentParams contains connector-specific deployment parameters
// 	// For Helm: namespace, chart_version, values, etc.
// 	// For Agent: install_path, service_account, etc.
// 	// DeploymentParams map[string]interface{} `json:"deploymentParams"`
// }

// ConnectorRequirements defines what's needed to deploy a connector
type ConnectorRequirements struct {
	// MinimumVersion specifies minimum version of deployment tool (e.g., Helm 3.0+)
	MinimumVersion string `json:"minimumVersion,omitempty"`

	// RequiredTools lists tools that must be available (e.g., ["helm", "kubectl"])
	RequiredTools []string `json:"requiredTools,omitempty"`

	// RequiredPermissions lists permissions needed (e.g., ["cluster-admin"])
	RequiredPermissions []string `json:"requiredPermissions,omitempty"`
}

// ConnectorFactory creates connector deployers based on type
type ConnectorFactory interface {
	// CreateConnectorDeployer creates a connector deployer for the specified type
	CreateConnectorDeployer(connectorConfig ConnectorDeployConfig) (ConnectorDeployer, error)

	// SupportedTypes returns list of supported connector types
	SupportedTypes() []ConnectorType
}

// ConnectorParams holds parameters for deploying a connector
type ConnectorParams struct {
	ConnectorType    string                 `json:"connectorType"`
	DeploymentConfig map[string]interface{} `json:"deploymentConfig"`
	Metadata         map[string]interface{} `json:"metadata"`
}

// ConnectorResult represents the result of a connector operation
type ConnectorResult struct {
	ConnectorID  string    `json:"connectorId"`
	ConnectionID string    `json:"connectionId"`
	Endpoint     string    `json:"endpoint"` // Connector endpoint, I don't think so this is required, see if we can remove
	Status       string    `json:"status"`
	Message      string    `json:"message"`
	CreatedAt    time.Time `json:"createdAt"`
}
