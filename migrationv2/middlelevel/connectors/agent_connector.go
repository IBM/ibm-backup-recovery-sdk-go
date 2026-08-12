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
	"fmt"
	"time"

	"github.com/IBM/ibm-backup-recovery-sdk-go/migrationv2/common/errors"
)

// KubernetesConnectorConfig holds Kubernetes-specific connector deployment configuration
type VsiConnectorConfig struct {
	// Namespace where the connector will be deployed
	VSIImage string `json:"namespace"`

	// Replicas specifies the number of connector pods to run
	// Default: 1 for Helm/Operator, ignored for DaemonSet
	Replicas int32 `json:"replicas"`

	// ChartVersion specifies the Helm chart version (for Helm connector)
	ResourceGrouID string `json:"chartVersion,omitempty"`

	// OperatorVersion specifies the operator version (for Operator connector)
	VPCID string `json:"operatorVersion,omitempty"`

	// CustomValues allows passing additional Helm values or operator config
	CustomValues map[string]interface{} `json:"customValues,omitempty"`
	Region       string
	// AuthConfig contains authentication configuration for accessing the datasource
	// For Kubernetes: KubernetesAuthConfig (kubeconfig, token, or certificate)
	// For VPC VSI: VSIAuthConfig (SSH key, password)
	AuthConfig AuthConfig `json:"authConfig"`
	// RegistrationToken is the token from BRS connection creation
	// RegistrationToken string `json:"registrationToken"`
}

// GetType returns the connector type
func (config *VsiConnectorConfig) CreateConnectorDeployer() (ConnectorDeployer, error) {
	return NewVSIConnector(config), nil
}

// GetType returns the connector type
func (config *VsiConnectorConfig) GetType() ConnectorType {
	return ConnectorTypeHelm
}

// Validate validates the deployment configuration
// func (config *VsiConnectorConfig) GetRegistrationToken() string {
// 	return config.RegistrationToken
// }

// Validate validates the deployment configuration
func (config *VsiConnectorConfig) Validate() error {
	if config == nil {
		return errors.NewInvalidConfigError("connector deploy config is required", nil)
	}

	// if config.GetRegistrationToken() == "" {
	// 	return errors.NewInvalidConfigError("registration token is required", nil)
	// }

	// Validate authentication config
	if config.AuthConfig == nil {
		return errors.NewInvalidConfigError("authentication config is required", nil)
	}

	return nil
}

// GetRequirements returns deployment requirements for Agent connector
func (a *VsiConnectorConfig) GetRequirements() *ConnectorRequirements {
	return &ConnectorRequirements{
		MinimumVersion: "1.0.0",
		RequiredTools:  []string{"ssh", "scp"},
		RequiredPermissions: []string{
			"ssh access to VSI",
			"sudo privileges on VSI",
			"network access to BRS endpoint",
		},
	}
}

// VSIConnector implements ConnectorDeployer for Agent-based VPC VSI connector
type VSIConnector struct {
	// Configuration
	installPath    string
	serviceAccount string
	vpcID          string
	vsiIDs         []string
	region         string
	VSIImage       string `json:"namespace"`

	// Replicas specifies the number of connector pods to run
	// Default: 1 for Helm/Operator, ignored for DaemonSet
	Replicas int32 `json:"replicas"`

	// ChartVersion specifies the Helm chart version (for Helm connector)
	ResourceGrouID string `json:"chartVersion,omitempty"`

	// OperatorVersion specifies the operator version (for Operator connector)
	VPCID string `json:"operatorVersion,omitempty"`

	// CustomValues allows passing additional Helm values or operator config
	CustomValues map[string]interface{} `json:"customValues,omitempty"`

	// State
	deployed bool
}

// NewVSIConnector creates a new Agent connector deployer
func NewVSIConnector(vsiConnectorConfig *VsiConnectorConfig) *VSIConnector {
	return &VSIConnector{
		installPath:    "/opt/brs-agent",
		serviceAccount: "brs-agent",
	}
}

// GetType returns the connector type
func (a *VSIConnector) GetType() ConnectorType {
	return ConnectorTypeAgent
}

// SetDeployContext satisfies ConnectorDeployer. VSI connector does not use BRS
// metadata, so this is a no-op.
func (a *VSIConnector) SetDeployContext(ctx ConnectorDeployContext) {}

// Deploy deploys the Agent-based connector to VPC VSI
func (a *VSIConnector) Deploy(ctx context.Context, registraionToken string) (*ConnectorResult, error) {

	// Extract deployment parameters
	// if err := a.extractDeploymentParams(config); err != nil {
	// 	return nil, err
	// }

	// TODO: Implement actual agent deployment
	// This would typically:
	// 1. SSH/connect to VSI instances
	// 2. Download agent installer
	// 3. Install agent
	// 4. Configure agent with connection token
	// 5. Start agent service, if required
	// 6. Verify agent registration with BRS

	// For now, return a mock result
	connectorID := fmt.Sprintf("agent-%s-%s", a.vpcID, a.vsiIDs[0])

	result := &ConnectorResult{
		ConnectorID: connectorID,
		// ConnectionID: extractConnectionID(config),
		// Endpoint:     extractVSIEndpoint(config),
		Status:    "deployed",
		Message:   fmt.Sprintf("Agent connector deployed successfully on %d VSI(s)", len(a.vsiIDs)),
		CreatedAt: time.Now(),
	}

	a.deployed = true
	return result, nil
}

// GetStatus retrieves the status of the Agent connector
func (a *VSIConnector) GetStatus(ctx context.Context, connectorID string) (string, error) {
	// TODO: Implement actual status check
	// This would typically:
	// 1. Check agent service status on VSI
	// 2. Verify connectivity to BRS
	// 3. Check agent health endpoint
	// 4. Verify backup/restore capabilities

	if !a.deployed {
		return "not_deployed", nil
	}

	return "running", nil
}

// Delete removes the Agent connector from the VSI
func (a *VSIConnector) Delete(ctx context.Context, connectorID string) error {
	// TODO: Implement actual agent uninstall
	// This would typically:
	// 1. Stop agent service
	// 2. Uninstall agent software
	// 3. Clean up configuration files
	// 4. Remove service account (if created)
	// 5. Verify cleanup completion

	a.deployed = false
	return nil
}

// extractDeploymentParams extracts and sets deployment parameters from config
func (a *VSIConnector) extractDeploymentParams(config *ConnectorDeployConfig) error {
	// Extract VPC ID
	// if vpcID, ok := config.DataSourceConfig["vpc_id"].(string); ok {
	// 	a.vpcID = vpcID
	// }

	// // Extract VSI IDs
	// if vsiIDs, ok := config.DataSourceConfig["vsi_ids"].([]string); ok {
	// 	a.vsiIDs = vsiIDs
	// } else if vsiIDs, ok := config.DataSourceConfig["vsi_ids"].([]interface{}); ok {
	// 	a.vsiIDs = make([]string, len(vsiIDs))
	// 	for i, id := range vsiIDs {
	// 		if strID, ok := id.(string); ok {
	// 			a.vsiIDs[i] = strID
	// 		}
	// 	}
	// }

	// // Extract region
	// if region, ok := config.DataSourceConfig["region"].(string); ok {
	// 	a.region = region
	// }

	// if config.DeploymentParams == nil {
	// 	return nil // Use defaults
	// }

	// // Extract install path
	// if path, ok := config.DeploymentParams["install_path"].(string); ok && path != "" {
	// 	a.installPath = path
	// }

	// // Extract service account
	// if sa, ok := config.DeploymentParams["service_account"].(string); ok && sa != "" {
	// 	a.serviceAccount = sa
	// }

	return nil
}

// extractVSIEndpoint extracts VSI endpoint from config
func extractVSIEndpoint(config *ConnectorDeployConfig) string {
	// if endpoint, ok := config.DataSourceConfig["vsi_endpoint"].(string); ok {
	// 	return endpoint
	// }
	// // If no explicit endpoint, construct from first VSI ID
	// if vsiIDs, ok := config.DataSourceConfig["vsi_ids"].([]string); ok && len(vsiIDs) > 0 {
	// 	return fmt.Sprintf("vsi://%s", vsiIDs[0])
	// }
	return ""
}
