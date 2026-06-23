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
	"testing"

	"github.com/IBM/ibm-backup-recovery-sdk-go/migrationv2/common/logger"
	"github.com/IBM/ibm-backup-recovery-sdk-go/migrationv2/common/metrics"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"helm.sh/helm/v3/pkg/action"
	"helm.sh/helm/v3/pkg/registry"
	"helm.sh/helm/v3/pkg/release"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/fake"
	"k8s.io/client-go/rest"
)

// MockIksManager implements middlelevel.IksManagerInterface for testing
type MockIksManager struct {
	mock.Mock
}

func (m *MockIksManager) ApplyRBACAndGetKubeconfig(clusterName string) ([]byte, error) {
	args := m.Called(clusterName)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]byte), args.Error(1)
}

func (m *MockIksManager) GetKubeApi(clusterName string) (*kubernetes.Clientset, *rest.Config, error) {
	args := m.Called(clusterName)
	var clientSet *kubernetes.Clientset
	var restConfig *rest.Config

	if args.Get(0) != nil {
		clientSet = args.Get(0).(*kubernetes.Clientset)
	}
	if args.Get(1) != nil {
		restConfig = args.Get(1).(*rest.Config)
	}

	return clientSet, restConfig, args.Error(2)
}

func (m *MockIksManager) GetClusterBearerToken(clientSet *kubernetes.Clientset) (string, error) {
	args := m.Called(clientSet)
	return args.String(0), args.Error(1)
}

func TestConfig_Methods(t *testing.T) {
	tests := []struct {
		name     string
		testFunc func(t *testing.T)
	}{
		{
			name: "GetType",
			testFunc: func(t *testing.T) {
				config := &HelmKubeConnectorConfig{
					ConnectorType: ConnectorTypeHelm,
				}
				assert.Equal(t, ConnectorTypeHelm, config.GetType())
			},
		},
		{
			name: "GetRequirements",
			testFunc: func(t *testing.T) {
				config := &HelmKubeConnectorConfig{}
				requirements := config.GetRequirements()
				assert.NotNil(t, requirements)
				assert.Equal(t, "3.0.0", requirements.MinimumVersion)
				assert.Contains(t, requirements.RequiredTools, "helm")
				assert.Contains(t, requirements.RequiredTools, "kubectl")
				assert.Contains(t, requirements.RequiredPermissions, "create/update/delete deployments")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.testFunc(t)
		})
	}
}

func TestConfigValidation(t *testing.T) {
	validAuthConfig := &KubernetesAuthConfig{
		AuthMethod: AuthMethodAPIKey,
		IamURL:     "https://iam.cloud.ibm.com",
		ApiKey:     "test-api-key",
		Host:       "icr.io",
	}

	tests := []struct {
		name          string
		config        *HelmKubeConnectorConfig
		expectError   bool
		errorContains string
		validateFunc  func(t *testing.T, config *HelmKubeConnectorConfig)
	}{
		{
			name: "Valid_Config",
			config: &HelmKubeConnectorConfig{
				Namespace:     "test-namespace",
				ClusterName:   "test-cluster",
				AuthConfig:    validAuthConfig,
				ConnectorType: ConnectorTypeHelm,
			},
			expectError: false,
		},
		{
			name:          "Nil_Config",
			config:        nil,
			expectError:   true,
			errorContains: "connector deploy config is required",
		},
		{
			name: "Nil_AuthConfig",
			config: &HelmKubeConnectorConfig{
				Namespace:     "test-namespace",
				ClusterName:   "test-cluster",
				AuthConfig:    nil,
				ConnectorType: ConnectorTypeHelm,
			},
			expectError:   true,
			errorContains: "authentication config is required",
		},
		{
			name: "Missing_Namespace",
			config: &HelmKubeConnectorConfig{
				ClusterName: "test-cluster",
				AuthConfig:  validAuthConfig,
			},
			expectError:   true,
			errorContains: "namespace is required",
		},
		{
			name: "Missing_ClusterName",
			config: &HelmKubeConnectorConfig{
				Namespace:  "test-namespace",
				AuthConfig: validAuthConfig,
			},
			expectError:   true,
			errorContains: "clusterName is required",
		},
		{
			name: "Invalid_AuthConfig",
			config: &HelmKubeConnectorConfig{
				Namespace:   "test-namespace",
				ClusterName: "test-cluster",
				AuthConfig: &KubernetesAuthConfig{
					AuthMethod: AuthMethodAPIKey,
				},
			},
			expectError: true,
		},
		{
			name: "WithStorageClass",
			config: &HelmKubeConnectorConfig{
				Namespace:    "test-namespace",
				ClusterName:  "test-cluster",
				StorageClass: "ibmc-block-silver",
				AuthConfig:   validAuthConfig,
			},
			expectError: false,
			validateFunc: func(t *testing.T, config *HelmKubeConnectorConfig) {
				assert.Equal(t, "ibmc-block-silver", config.StorageClass)
			},
		},
		{
			name: "WithoutStorageClass",
			config: &HelmKubeConnectorConfig{
				Namespace:   "test-namespace",
				ClusterName: "test-cluster",
				AuthConfig:  validAuthConfig,
			},
			expectError: false,
			validateFunc: func(t *testing.T, config *HelmKubeConnectorConfig) {
				assert.Equal(t, "", config.StorageClass)
			},
		},
		{
			name: "With_CustomValues",
			config: &HelmKubeConnectorConfig{
				Namespace:   "test-namespace",
				ClusterName: "test-cluster",
				AuthConfig:  validAuthConfig,
				CustomValues: map[string]interface{}{
					"custom_key": "custom_value",
					"replicas":   3,
				},
			},
			expectError: false,
			validateFunc: func(t *testing.T, config *HelmKubeConnectorConfig) {
				assert.NotNil(t, config.CustomValues)
				assert.Equal(t, "custom_value", config.CustomValues["custom_key"])
				assert.Equal(t, 3, config.CustomValues["replicas"])
			},
		},
		{
			name: "With_DeploymentParams",
			config: &HelmKubeConnectorConfig{
				Namespace:   "test-namespace",
				ClusterName: "test-cluster",
				AuthConfig:  validAuthConfig,
				DeploymentParams: map[string]interface{}{
					"timeout":     "10m",
					"waitForPods": true,
				},
			},
			expectError: false,
			validateFunc: func(t *testing.T, config *HelmKubeConnectorConfig) {
				assert.NotNil(t, config.DeploymentParams)
				assert.Equal(t, "10m", config.DeploymentParams["timeout"])
				assert.Equal(t, true, config.DeploymentParams["waitForPods"])
			},
		},
		{
			name: "With_DataSourceConfig",
			config: &HelmKubeConnectorConfig{
				Namespace:   "test-namespace",
				ClusterName: "test-cluster",
				AuthConfig:  validAuthConfig,
				DataSourceConfig: map[string]interface{}{
					"cluster_id":       "test-cluster-id",
					"cluster_endpoint": "https://test.endpoint.com",
				},
			},
			expectError: false,
			validateFunc: func(t *testing.T, config *HelmKubeConnectorConfig) {
				assert.NotNil(t, config.DataSourceConfig)
				assert.Equal(t, "test-cluster-id", config.DataSourceConfig["cluster_id"])
			},
		},
		{
			name: "With_NodeSelector",
			config: &HelmKubeConnectorConfig{
				Namespace:   "test-namespace",
				ClusterName: "test-cluster",
				AuthConfig:  validAuthConfig,
				NodeSelector: map[string]string{
					"disktype": "ssd",
					"region":   "us-south",
				},
			},
			expectError: false,
			validateFunc: func(t *testing.T, config *HelmKubeConnectorConfig) {
				assert.NotNil(t, config.NodeSelector)
				assert.Equal(t, "ssd", config.NodeSelector["disktype"])
				assert.Equal(t, "us-south", config.NodeSelector["region"])
			},
		},
		{
			name: "With_Tolerations",
			config: &HelmKubeConnectorConfig{
				Namespace:   "test-namespace",
				ClusterName: "test-cluster",
				AuthConfig:  validAuthConfig,
				Tolerations: []map[string]interface{}{
					{
						"key":      "key1",
						"operator": "Equal",
						"value":    "value1",
						"effect":   "NoSchedule",
					},
				},
			},
			expectError: false,
			validateFunc: func(t *testing.T, config *HelmKubeConnectorConfig) {
				assert.NotNil(t, config.Tolerations)
				assert.Len(t, config.Tolerations, 1)
				assert.Equal(t, "key1", config.Tolerations[0]["key"])
			},
		},
		{
			name: "With_ServiceAccount",
			config: &HelmKubeConnectorConfig{
				Namespace:      "test-namespace",
				ClusterName:    "test-cluster",
				AuthConfig:     validAuthConfig,
				ServiceAccount: "custom-service-account",
			},
			expectError: false,
			validateFunc: func(t *testing.T, config *HelmKubeConnectorConfig) {
				assert.Equal(t, "custom-service-account", config.ServiceAccount)
			},
		},
		{
			name: "With_ImagePullPolicy",
			config: &HelmKubeConnectorConfig{
				Namespace:       "test-namespace",
				ClusterName:     "test-cluster",
				AuthConfig:      validAuthConfig,
				ImagePullPolicy: "Always",
			},
			expectError: false,
			validateFunc: func(t *testing.T, config *HelmKubeConnectorConfig) {
				assert.Equal(t, "Always", config.ImagePullPolicy)
			},
		},
		{
			name: "With_Resources",
			config: &HelmKubeConnectorConfig{
				Namespace:   "test-namespace",
				ClusterName: "test-cluster",
				AuthConfig:  validAuthConfig,
				Resources: &ResourceRequirements{
					Requests: &ResourceList{CPU: "250m", Memory: "256Mi"},
					Limits:   &ResourceList{CPU: "500m", Memory: "512Mi"},
				},
			},
			expectError: false,
			validateFunc: func(t *testing.T, config *HelmKubeConnectorConfig) {
				assert.NotNil(t, config.Resources)
				assert.Equal(t, "250m", config.Resources.Requests.CPU)
				assert.Equal(t, "512Mi", config.Resources.Limits.Memory)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.Validate()

			if tt.expectError {
				assert.Error(t, err)
				if tt.errorContains != "" {
					assert.Contains(t, err.Error(), tt.errorContains)
				}
			} else {
				assert.NoError(t, err)
				if tt.validateFunc != nil {
					tt.validateFunc(t, tt.config)
				}
			}
		})
	}
}

func TestNewConnector(t *testing.T) {
	tests := []struct {
		name          string
		config        *HelmKubeConnectorConfig
		expectError   bool
		errorContains string
		description   string
	}{
		{
			name:          "NilConfig",
			config:        nil,
			expectError:   true,
			errorContains: "connector config is required",
			description:   "Should fail with nil config",
		},
		{
			name: "InvalidAuthConfig_MissingFields",
			config: &HelmKubeConnectorConfig{
				Namespace:   "test-namespace",
				ClusterName: "test-cluster",
				AuthConfig: &KubernetesAuthConfig{
					AuthMethod: AuthMethodAPIKey,
				},
			},
			expectError: true,
			description: "Should fail with invalid auth config",
		},
		{
			name: "MissingNamespace",
			config: &HelmKubeConnectorConfig{
				ClusterName: "test-cluster",
				AuthConfig: &KubernetesAuthConfig{
					AuthMethod: AuthMethodAPIKey,
					IamURL:     "https://iam.cloud.ibm.com",
					ApiKey:     "test-api-key",
					Host:       "https://test.containers.cloud.ibm.com",
				},
			},
			expectError: true,
			description: "Should fail with missing namespace",
		},
		{
			name: "MissingClusterName",
			config: &HelmKubeConnectorConfig{
				Namespace: "test-namespace",
				AuthConfig: &KubernetesAuthConfig{
					AuthMethod: AuthMethodAPIKey,
					IamURL:     "https://iam.cloud.ibm.com",
					ApiKey:     "test-api-key",
					Host:       "https://test.containers.cloud.ibm.com",
				},
			},
			expectError: true,
			description: "Should fail with missing cluster name",
		},
		{
			name: "ValidConfig_Classic",
			config: &HelmKubeConnectorConfig{
				Namespace:             "test-namespace",
				ClusterName:           "test-cluster",
				ContainerEndpoint:     "https://containers.cloud.ibm.com",
				ContainerEndpointType: "public",
				StorageClass:          "ibmc-block-silver",
				Replicas:              1,
				ChartVersion:          "1.0.0",
				ReleaseName:           "test-release",
				ChartName:             "test-chart",
				ChartReference:        "oci://icr.io/test/chart",
				ImagePullPolicy:       "IfNotPresent",
				AuthConfig: &KubernetesAuthConfig{
					AuthMethod: AuthMethodAPIKey,
					IamURL:     "https://iam.cloud.ibm.com",
					ApiKey:     "test-api-key",
					Host:       "icr.io",
				},
				ConnectorType: ConnectorTypeHelm,
				Logger:        logger.NewNoop(),
				Metrics:       metrics.NewNoop(),
				AccountID:     "test-account-id",
			},
			expectError: true, // Expected to fail without real credentials
			description: "Valid config but fails without real IKS credentials",
		},
		{
			name: "ValidConfig_DefaultValues",
			config: &HelmKubeConnectorConfig{
				Namespace:             "test-namespace",
				ClusterName:           "test-cluster",
				ContainerEndpoint:     "https://containers.cloud.ibm.com",
				ContainerEndpointType: "public",
				Replicas:              2,
				AuthConfig: &KubernetesAuthConfig{
					AuthMethod: AuthMethodAPIKey,
					IamURL:     "https://iam.cloud.ibm.com",
					ApiKey:     "test-api-key",
					Host:       "icr.io",
				},
				ConnectorType: ConnectorTypeHelm,
			},
			expectError: true,
			description: "Valid config but fails without real credentials",
		},
		{
			name: "WithCustomValues",
			config: &HelmKubeConnectorConfig{
				Namespace:             "test-namespace",
				ClusterName:           "test-cluster",
				ContainerEndpoint:     "https://containers.cloud.ibm.com",
				ContainerEndpointType: "public",
				Replicas:              2,
				ChartVersion:          "1.0.0",
				ReleaseName:           "my-connector",
				ChartName:             "my-chart",
				ChartReference:        "oci://my-registry/my-chart",
				WaitTillDeploy:        true,
				ImagePullPolicy:       "Always",
				AuthConfig: &KubernetesAuthConfig{
					AuthMethod: AuthMethodAPIKey,
					IamURL:     "https://iam.cloud.ibm.com",
					ApiKey:     "test-api-key",
					Host:       "icr.io",
				},
				Logger:    logger.NewNoop(),
				Metrics:   metrics.NewNoop(),
				AccountID: "test-account-id",
			},
			expectError: true,
			description: "Config with custom values but fails without real credentials",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			connector, err := NewHelmConnector(tt.config)

			if tt.expectError {
				assert.Error(t, err)
				assert.Nil(t, connector)
				if tt.errorContains != "" {
					assert.Contains(t, err.Error(), tt.errorContains)
				}
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, connector)
			}
		})
	}
}

func TestConnector_Methods(t *testing.T) {
	tests := []struct {
		name     string
		testFunc func(t *testing.T)
	}{
		{
			name: "GetType",
			testFunc: func(t *testing.T) {
				connector := &HelmConnector{}
				assert.Equal(t, ConnectorTypeHelm, connector.GetType())
			},
		},
		{
			name: "Delete",
			testFunc: func(t *testing.T) {
				connector := &HelmConnector{
					deployed: true,
					logger:   logger.NewNoOpLogger(),
				}
				ctx := context.Background()
				err := connector.Delete(ctx, "test-connector-id")
				assert.NoError(t, err)
				assert.False(t, connector.deployed)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.testFunc(t)
		})
	}
}

func TestConnector_GetStatus(t *testing.T) {
	tests := []struct {
		name           string
		deployed       bool
		expectedStatus string
	}{
		{
			name:           "NotDeployed",
			deployed:       false,
			expectedStatus: "not_deployed",
		},
		{
			name:           "Deployed",
			deployed:       true,
			expectedStatus: "running",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			connector := &HelmConnector{
				deployed: tt.deployed,
				logger:   logger.NewNoop(),
			}

			ctx := context.Background()
			status, err := connector.GetStatus(ctx, "test-connector-id")

			assert.NoError(t, err)
			assert.Equal(t, tt.expectedStatus, status)
		})
	}
}

func TestKubeAuthConfig(t *testing.T) {
	tests := []struct {
		name          string
		config        *KubernetesAuthConfig
		expectError   bool
		errorContains string
		skipValidate  bool
		validateFunc  func(t *testing.T, config *KubernetesAuthConfig)
	}{
		{
			name: "GetAuthMethod",
			config: &KubernetesAuthConfig{
				AuthMethod: AuthMethodAPIKey,
			},
			expectError:  false,
			skipValidate: true,
			validateFunc: func(t *testing.T, config *KubernetesAuthConfig) {
				assert.Equal(t, AuthMethodAPIKey, config.GetAuthMethod())
			},
		},
		{
			name: "GetIamURL",
			config: &KubernetesAuthConfig{
				IamURL: "https://iam.cloud.ibm.com",
			},
			expectError:  false,
			skipValidate: true,
			validateFunc: func(t *testing.T, config *KubernetesAuthConfig) {
				assert.Equal(t, "https://iam.cloud.ibm.com", config.GetIamURL())
			},
		},
		{
			name: "GetAPIKey",
			config: &KubernetesAuthConfig{
				ApiKey: "test-api-key-123",
			},
			expectError:  false,
			skipValidate: true,
			validateFunc: func(t *testing.T, config *KubernetesAuthConfig) {
				assert.Equal(t, "test-api-key-123", config.GetAPIKey())
			},
		},
		{
			name: "APIKey_Success",
			config: &KubernetesAuthConfig{
				AuthMethod: AuthMethodAPIKey,
				IamURL:     "https://iam.cloud.ibm.com",
				ApiKey:     "test-api-key",
				Host:       "icr.io",
			},
			expectError: false,
		},
		{
			name: "APIKey_MissingIamURL",
			config: &KubernetesAuthConfig{
				AuthMethod: AuthMethodAPIKey,
				ApiKey:     "test-api-key",
				Host:       "icr.io",
			},
			expectError:   true,
			errorContains: "Iam URL is required",
		},
		{
			name: "APIKey_MissingHost",
			config: &KubernetesAuthConfig{
				AuthMethod: AuthMethodAPIKey,
				IamURL:     "https://iam.cloud.ibm.com",
				ApiKey:     "test-api-key",
			},
			expectError:   true,
			errorContains: "host is required",
		},
		{
			name: "APIKey_MissingAPIKey",
			config: &KubernetesAuthConfig{
				AuthMethod: AuthMethodAPIKey,
				IamURL:     "https://iam.cloud.ibm.com",
				Host:       "icr.io",
			},
			expectError:   true,
			errorContains: "apiKey is required",
		},
		{
			name: "Token_Success",
			config: &KubernetesAuthConfig{
				AuthMethod:  AuthMethodToken,
				IamURL:      "https://iam.cloud.ibm.com",
				APIServer:   "https://api.cluster.example.com",
				BearerToken: "test-bearer-token",
			},
			expectError: false,
		},
		{
			name: "Token_MissingAPIServer",
			config: &KubernetesAuthConfig{
				AuthMethod:  AuthMethodToken,
				IamURL:      "https://iam.cloud.ibm.com",
				BearerToken: "test-bearer-token",
			},
			expectError:   true,
			errorContains: "apiServer is required",
		},
		{
			name: "Token_MissingBearerToken",
			config: &KubernetesAuthConfig{
				AuthMethod: AuthMethodToken,
				IamURL:     "https://iam.cloud.ibm.com",
				APIServer:  "https://api.cluster.example.com",
			},
			expectError:   true,
			errorContains: "bearerToken is required",
		},
		{
			name: "Kubeconfig_Success",
			config: &KubernetesAuthConfig{
				AuthMethod:        AuthMethodKubeconfig,
				IamURL:            "https://iam.cloud.ibm.com",
				KubeconfigContent: "apiVersion: v1\nkind: Config",
			},
			expectError: false,
		},
		{
			name: "Kubeconfig_MissingContent",
			config: &KubernetesAuthConfig{
				AuthMethod: AuthMethodKubeconfig,
				IamURL:     "https://iam.cloud.ibm.com",
			},
			expectError:   true,
			errorContains: "either kubeconfigContent or kubeconfigPath is required",
		},
		{
			name: "InvalidAuthMethod",
			config: &KubernetesAuthConfig{
				AuthMethod: AuthMethod("invalid"),
				IamURL:     "https://iam.cloud.ibm.com",
			},
			expectError:   true,
			errorContains: "invalid authentication method",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if !tt.skipValidate {
				err := tt.config.Validate()

				if tt.expectError {
					assert.Error(t, err)
					if tt.errorContains != "" {
						assert.Contains(t, err.Error(), tt.errorContains)
					}
				} else {
					assert.NoError(t, err)
				}
			}

			if tt.validateFunc != nil {
				tt.validateFunc(t, tt.config)
			}
		})
	}
}

func TestVSIAuthConfig(t *testing.T) {
	tests := []struct {
		name          string
		config        *VSIAuthConfig
		expectError   bool
		errorContains string
		skipValidate  bool
		validateFunc  func(t *testing.T, config *VSIAuthConfig)
	}{
		{
			name: "GetAuthMethod",
			config: &VSIAuthConfig{
				AuthMethod: AuthMethodSSHKey,
			},
			expectError:  false,
			skipValidate: true,
			validateFunc: func(t *testing.T, config *VSIAuthConfig) {
				assert.Equal(t, AuthMethodSSHKey, config.GetAuthMethod())
			},
		},
		{
			name:         "GetIamURL_Empty",
			config:       &VSIAuthConfig{},
			expectError:  false,
			skipValidate: true,
			validateFunc: func(t *testing.T, config *VSIAuthConfig) {
				assert.Equal(t, "", config.GetIamURL())
			},
		},
		{
			name:         "GetAPIKey_Empty",
			config:       &VSIAuthConfig{},
			expectError:  false,
			skipValidate: true,
			validateFunc: func(t *testing.T, config *VSIAuthConfig) {
				assert.Equal(t, "", config.GetAPIKey())
			},
		},
		{
			name: "SSHKey_Success",
			config: &VSIAuthConfig{
				AuthMethod: AuthMethodSSHKey,
				Host:       "192.168.1.100",
				Port:       22,
				User:       "root",
				PrivateKey: "-----BEGIN RSA PRIVATE KEY-----\n...",
			},
			expectError: false,
		},
		{
			name: "SSHKey_MissingHost",
			config: &VSIAuthConfig{
				AuthMethod: AuthMethodSSHKey,
				User:       "root",
				PrivateKey: "-----BEGIN RSA PRIVATE KEY-----\n...",
			},
			expectError:   true,
			errorContains: "host (IP address or hostname) is required",
		},
		{
			name: "SSHKey_MissingUser",
			config: &VSIAuthConfig{
				AuthMethod: AuthMethodSSHKey,
				Host:       "192.168.1.100",
				PrivateKey: "-----BEGIN RSA PRIVATE KEY-----\n...",
			},
			expectError:   true,
			errorContains: "user is required",
		},
		{
			name: "SSHKey_MissingPrivateKey",
			config: &VSIAuthConfig{
				AuthMethod: AuthMethodSSHKey,
				Host:       "192.168.1.100",
				User:       "root",
			},
			expectError:   true,
			errorContains: "either privateKey or privateKeyPath is required",
		},
		{
			name: "SSHKey_DefaultPort",
			config: &VSIAuthConfig{
				AuthMethod: AuthMethodSSHKey,
				Host:       "192.168.1.100",
				User:       "root",
				PrivateKey: "-----BEGIN RSA PRIVATE KEY-----\n...",
			},
			expectError: false,
			validateFunc: func(t *testing.T, config *VSIAuthConfig) {
				assert.Equal(t, 22, config.Port)
			},
		},
		{
			name: "SSHPassword_Success",
			config: &VSIAuthConfig{
				AuthMethod: AuthMethodSSHPassword,
				Host:       "192.168.1.100",
				Port:       22,
				User:       "root",
				Password:   "secure-password",
			},
			expectError: false,
		},
		{
			name: "SSHPassword_MissingPassword",
			config: &VSIAuthConfig{
				AuthMethod: AuthMethodSSHPassword,
				Host:       "192.168.1.100",
				User:       "root",
			},
			expectError:   true,
			errorContains: "password is required",
		},
		{
			name: "InvalidAuthMethod",
			config: &VSIAuthConfig{
				AuthMethod: AuthMethod("invalid"),
				Host:       "192.168.1.100",
				User:       "root",
			},
			expectError:   true,
			errorContains: "invalid authentication method for VSI",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if !tt.skipValidate {
				err := tt.config.Validate()

				if tt.expectError {
					assert.Error(t, err)
					if tt.errorContains != "" {
						assert.Contains(t, err.Error(), tt.errorContains)
					}
				} else {
					assert.NoError(t, err)
				}
			}

			if tt.validateFunc != nil {
				tt.validateFunc(t, tt.config)
			}
		})
	}
}

func TestConnectorError(t *testing.T) {
	tests := []struct {
		name     string
		testFunc func(t *testing.T)
	}{
		{
			name: "Error_Method",
			testFunc: func(t *testing.T) {
				err := &ConnectorError{
					Code:    "TEST_ERROR",
					Message: "This is a test error",
				}
				assert.Equal(t, "This is a test error", err.Error())
			},
		},
		{
			name: "NewConnectorError",
			testFunc: func(t *testing.T) {
				err := NewConnectorError("TEST_CODE", "Test message")
				assert.NotNil(t, err)
				assert.Equal(t, "TEST_CODE", err.Code)
				assert.Equal(t, "Test message", err.Message)
				assert.NotNil(t, err.Details)
				assert.Empty(t, err.Details)
			},
		},
		{
			name: "WithDetails",
			testFunc: func(t *testing.T) {
				err := NewConnectorError("TEST_CODE", "Test message")
				err.WithDetails("key1", "value1")
				err.WithDetails("key2", 123)
				assert.Equal(t, "value1", err.Details["key1"])
				assert.Equal(t, 123, err.Details["key2"])
			},
		},
		{
			name: "WithDetails_Chaining",
			testFunc: func(t *testing.T) {
				err := NewConnectorError("TEST_CODE", "Test message").
					WithDetails("key1", "value1").
					WithDetails("key2", "value2")
				assert.Equal(t, "value1", err.Details["key1"])
				assert.Equal(t, "value2", err.Details["key2"])
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.testFunc(t)
		})
	}
}

func TestConfig_CreateDeployer(t *testing.T) {
	tests := []struct {
		name          string
		config        *HelmKubeConnectorConfig
		expectError   bool
		errorContains string
	}{
		{
			name: "Success_ValidConfig",
			config: &HelmKubeConnectorConfig{
				Namespace:             "test-namespace",
				ClusterName:           "test-cluster",
				ContainerEndpoint:     "https://test.containers.cloud.ibm.com",
				ContainerEndpointType: "public",
				StorageClass:          "ibmc-block-silver",
				Replicas:              1,
				ConnectorType:         ConnectorTypeHelm,
				AuthConfig: &KubernetesAuthConfig{
					AuthMethod: AuthMethodAPIKey,
					IamURL:     "https://iam.cloud.ibm.com",
					ApiKey:     "test-api-key",
					Host:       "https://test-cluster.containers.cloud.ibm.com",
				},
				Logger:    logger.NewNoop(),
				Metrics:   metrics.NewNoop(),
				AccountID: "test-account-id",
			},
			expectError: true, // Expected to fail without real credentials
		},
		{
			name:          "NilConfig",
			config:        nil,
			expectError:   true,
			errorContains: "connector config is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			deployer, err := tt.config.CreateConnectorDeployer()

			if tt.expectError {
				assert.Error(t, err)
				assert.Nil(t, deployer)
				if tt.errorContains != "" {
					assert.Contains(t, err.Error(), tt.errorContains)
				}
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, deployer)
			}
		})
	}
}

func TestResources(t *testing.T) {
	tests := []struct {
		name     string
		testFunc func(t *testing.T)
	}{
		{
			name: "ResourceRequirements_Structure",
			testFunc: func(t *testing.T) {
				resources := &ResourceRequirements{
					Requests: &ResourceList{CPU: "500m", Memory: "512Mi"},
					Limits:   &ResourceList{CPU: "1000m", Memory: "1Gi"},
				}
				assert.NotNil(t, resources.Requests)
				assert.NotNil(t, resources.Limits)
				assert.Equal(t, "500m", resources.Requests.CPU)
				assert.Equal(t, "512Mi", resources.Requests.Memory)
				assert.Equal(t, "1000m", resources.Limits.CPU)
				assert.Equal(t, "1Gi", resources.Limits.Memory)
			},
		},
		{
			name: "ResourceList_EmptyValues",
			testFunc: func(t *testing.T) {
				resources := &ResourceList{}
				assert.Empty(t, resources.CPU)
				assert.Empty(t, resources.Memory)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.testFunc(t)
		})
	}
}

func TestSetupRegistry_Auth(t *testing.T) {
	tests := []struct {
		name           string
		config         *HelmKubeConnectorConfig
		expectError    bool
		expectedMethod AuthMethod
	}{
		{
			name: "APIKeyAuth",
			config: &HelmKubeConnectorConfig{
				Namespace:   "test-namespace",
				ClusterName: "test-cluster",
				AuthConfig: &KubernetesAuthConfig{
					AuthMethod: AuthMethodAPIKey,
					IamURL:     "https://iam.cloud.ibm.com",
					ApiKey:     "test-api-key",
					Host:       "https://test.containers.cloud.ibm.com",
				},
			},
			expectError:    false,
			expectedMethod: AuthMethodAPIKey,
		},
		{
			name: "TokenAuth",
			config: &HelmKubeConnectorConfig{
				Namespace:   "test-namespace",
				ClusterName: "test-cluster",
				AuthConfig: &KubernetesAuthConfig{
					AuthMethod:  AuthMethodToken,
					BearerToken: "test-bearer-token",
					IamURL:      "https://iam.cloud.ibm.com",
					APIServer:   "https://api.test-cluster.cloud.ibm.com",
					Host:        "https://test.containers.cloud.ibm.com",
				},
				Logger:    logger.NewNoop(),
				Metrics:   metrics.NewNoop(),
				AccountID: "test-account-id",
			},
			expectError:    false,
			expectedMethod: AuthMethodToken,
		},
		{
			name: "InvalidAuthMethod",
			config: &HelmKubeConnectorConfig{
				Namespace:   "test-namespace",
				ClusterName: "test-cluster",
				AuthConfig: &KubernetesAuthConfig{
					AuthMethod: "UNSUPPORTED",
					Host:       "https://test.containers.cloud.ibm.com",
				},
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.AuthConfig.Validate()

			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedMethod, tt.config.AuthConfig.GetAuthMethod())
			}
		})
	}
}

type MockHelmClient struct {
	mock.Mock
}

func (m *MockHelmClient) Init(actionConfig *action.Configuration, namespace string) error {
	args := m.Called(actionConfig, namespace)
	return args.Error(0)
}

func (m *MockHelmClient) Install(actionConfig *action.Configuration, config *HelmInstallConfig) (*release.Release, error) {
	args := m.Called(actionConfig, config)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*release.Release), args.Error(1)
}

type MockRegistryClient struct {
	mock.Mock
	client *registry.Client
}

func (m *MockRegistryClient) Login(host string, username, password string) error {
	args := m.Called(host, username, password)
	return args.Error(0)
}

func (m *MockRegistryClient) Logout(host string) error {
	args := m.Called(host)
	return args.Error(0)
}

func (m *MockRegistryClient) GetClient() *registry.Client {
	return m.client
}

func TestDeploy_Success(t *testing.T) {
	fakeClientSet := fake.NewSimpleClientset()
	ns := &corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: "test-namespace",
		},
	}
	_, err := fakeClientSet.CoreV1().Namespaces().Create(context.Background(), ns, metav1.CreateOptions{})
	assert.NoError(t, err)

	mockHelmClient := new(MockHelmClient)
	mockHelmClient.On("Init", mock.Anything, "test-namespace").Return(nil)
	mockHelmClient.On("Install", mock.Anything, mock.MatchedBy(func(config *HelmInstallConfig) bool {
		return config.ReleaseName == "brs-connector" &&
			config.Namespace == "test-namespace" &&
			config.ChartRef == "oci://icr.io/ext/brs/brs-ds-connector-chart" &&
			config.Version == "" &&
			config.Wait == false
	})).Return(&release.Release{
		Name:      "brs-connector",
		Namespace: "test-namespace",
	}, nil)

	mockRegistryClient := new(MockRegistryClient)
	mockRegistryClient.On("Login", "icr.io", "iamapikey", "test-api-key").Return(nil)
	mockRegistryClient.client = &registry.Client{}

	connector := &HelmConnector{
		namespace:       "test-namespace",
		releaseName:     "brs-connector",
		chartRepo:       "oci://icr.io/ext/brs/brs-ds-connector-chart",
		chartName:       "brs-connector",
		ImagePullPolicy: "IfNotPresent",
		replicas:        1,
		clientSet:       fakeClientSet,
		restConfig:      &rest.Config{},
		config: &HelmKubeConnectorConfig{
			Namespace:      "test-namespace",
			ClusterName:    "test-cluster",
			WaitTillDeploy: false,
			AuthConfig: &KubernetesAuthConfig{
				AuthMethod: AuthMethodAPIKey,
				IamURL:     "https://iam.cloud.ibm.com",
				ApiKey:     "test-api-key",
				Host:       "icr.io",
			},
		},
		helmClient:     mockHelmClient,
		registryClient: mockRegistryClient,
		logger:         logger.NewNoop(),
	}

	result, err := connector.Deploy(context.Background(), "test-token")

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, "deployed", result.Status)
	assert.Contains(t, result.Message, "test-namespace")
	assert.True(t, connector.deployed)
	mockHelmClient.AssertExpectations(t)
	mockRegistryClient.AssertExpectations(t)
}

func TestDeploy_WithStorageClass(t *testing.T) {
	fakeClientSet := fake.NewSimpleClientset()
	ns := &corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: "test-namespace",
		},
	}
	_, err := fakeClientSet.CoreV1().Namespaces().Create(context.Background(), ns, metav1.CreateOptions{})
	assert.NoError(t, err)

	mockHelmClient := new(MockHelmClient)
	mockHelmClient.On("Init", mock.Anything, "test-namespace").Return(nil)
	mockHelmClient.On("Install", mock.Anything, mock.MatchedBy(func(config *HelmInstallConfig) bool {
		vct, ok := config.Values["volumeClaimTemplate"].(map[string]interface{})
		if !ok {
			return false
		}
		storageClass, ok := vct["storageClass"].(string)
		return ok && storageClass == "ibmc-block-silver"
	})).Return(&release.Release{
		Name:      "brs-connector",
		Namespace: "test-namespace",
	}, nil)

	mockRegistryClient := new(MockRegistryClient)
	mockRegistryClient.On("Login", "icr.io", "iamapikey", "test-api-key").Return(nil)
	mockRegistryClient.client = &registry.Client{}

	connector := &HelmConnector{
		namespace:       "test-namespace",
		releaseName:     "brs-connector",
		chartRepo:       "oci://icr.io/ext/brs/brs-ds-connector-chart",
		chartName:       "brs-connector",
		ImagePullPolicy: "IfNotPresent",
		replicas:        1,
		clientSet:       fakeClientSet,
		restConfig:      &rest.Config{},
		config: &HelmKubeConnectorConfig{
			Namespace:      "test-namespace",
			ClusterName:    "test-cluster",
			StorageClass:   "ibmc-block-silver",
			WaitTillDeploy: false,
			AuthConfig: &KubernetesAuthConfig{
				AuthMethod: AuthMethodAPIKey,
				IamURL:     "https://iam.cloud.ibm.com",
				ApiKey:     "test-api-key",
				Host:       "icr.io",
			},
		},
		helmClient:     mockHelmClient,
		registryClient: mockRegistryClient,
		logger:         logger.NewNoop(),
	}

	result, err := connector.Deploy(context.Background(), "test-token")

	assert.NoError(t, err)
	assert.NotNil(t, result)
	mockHelmClient.AssertExpectations(t)
	mockRegistryClient.AssertExpectations(t)
}

func TestDeploy_Errors(t *testing.T) {
	tests := []struct {
		name          string
		setupMocks    func() (*HelmConnector, *MockHelmClient, *MockRegistryClient)
		errorContains string
	}{
		{
			name: "AuthValidationError",
			setupMocks: func() (*HelmConnector, *MockHelmClient, *MockRegistryClient) {
				connector := &HelmConnector{
					config: &HelmKubeConnectorConfig{
						AuthConfig: &KubernetesAuthConfig{
							AuthMethod: AuthMethodAPIKey,
						},
					},
					logger: logger.NewNoOpLogger(),
				}
				return connector, nil, nil
			},
			errorContains: "",
		},
		{
			name: "HelmInitError",
			setupMocks: func() (*HelmConnector, *MockHelmClient, *MockRegistryClient) {
				fakeClientSet := fake.NewSimpleClientset()
				ns := &corev1.Namespace{
					ObjectMeta: metav1.ObjectMeta{
						Name: "test-namespace",
					},
				}
				_, _ = fakeClientSet.CoreV1().Namespaces().Create(context.Background(), ns, metav1.CreateOptions{})

				mockHelmClient := new(MockHelmClient)
				mockHelmClient.On("Init", mock.Anything, "test-namespace").Return(fmt.Errorf("helm init failed"))

				connector := &HelmConnector{
					namespace:  "test-namespace",
					clientSet:  fakeClientSet,
					restConfig: &rest.Config{},
					config: &HelmKubeConnectorConfig{
						AuthConfig: &KubernetesAuthConfig{
							AuthMethod: AuthMethodAPIKey,
							IamURL:     "https://iam.cloud.ibm.com",
							ApiKey:     "test-api-key",
							Host:       "icr.io",
						},
					},
					helmClient: mockHelmClient,
					logger:     logger.NewNoop(),
				}

				return connector, mockHelmClient, nil
			},
			errorContains: "helm init failed",
		},
		{
			name: "RegistryLoginError",
			setupMocks: func() (*HelmConnector, *MockHelmClient, *MockRegistryClient) {
				fakeClientSet := fake.NewSimpleClientset()
				ns := &corev1.Namespace{
					ObjectMeta: metav1.ObjectMeta{
						Name: "test-namespace",
					},
				}
				_, _ = fakeClientSet.CoreV1().Namespaces().Create(context.Background(), ns, metav1.CreateOptions{})

				mockHelmClient := new(MockHelmClient)
				mockHelmClient.On("Init", mock.Anything, "test-namespace").Return(nil)

				mockRegistryClient := new(MockRegistryClient)
				mockRegistryClient.On("Login", "icr.io", "iamapikey", "test-api-key").Return(fmt.Errorf("login failed"))

				connector := &HelmConnector{
					namespace:  "test-namespace",
					clientSet:  fakeClientSet,
					restConfig: &rest.Config{},
					config: &HelmKubeConnectorConfig{
						AuthConfig: &KubernetesAuthConfig{
							AuthMethod: AuthMethodAPIKey,
							IamURL:     "https://iam.cloud.ibm.com",
							ApiKey:     "test-api-key",
							Host:       "icr.io",
						},
					},
					helmClient:     mockHelmClient,
					registryClient: mockRegistryClient,
					logger:         logger.NewNoop(),
				}

				return connector, mockHelmClient, mockRegistryClient
			},
			errorContains: "login failed",
		},
		{
			name: "LocateChartError",
			setupMocks: func() (*HelmConnector, *MockHelmClient, *MockRegistryClient) {
				fakeClientSet := fake.NewSimpleClientset()
				ns := &corev1.Namespace{
					ObjectMeta: metav1.ObjectMeta{
						Name: "test-namespace",
					},
				}
				_, _ = fakeClientSet.CoreV1().Namespaces().Create(context.Background(), ns, metav1.CreateOptions{})

				mockHelmClient := new(MockHelmClient)
				mockHelmClient.On("Init", mock.Anything, "test-namespace").Return(nil)
				mockHelmClient.On("Install", mock.Anything, mock.Anything).Return(nil, fmt.Errorf("chart not found"))

				mockRegistryClient := new(MockRegistryClient)
				mockRegistryClient.On("Login", "icr.io", "iamapikey", "test-api-key").Return(nil)
				mockRegistryClient.client = &registry.Client{}

				connector := &HelmConnector{
					namespace:  "test-namespace",
					chartRepo:  "oci://icr.io/ext/brs/brs-ds-connector-chart",
					clientSet:  fakeClientSet,
					restConfig: &rest.Config{},
					config: &HelmKubeConnectorConfig{
						AuthConfig: &KubernetesAuthConfig{
							AuthMethod: AuthMethodAPIKey,
							IamURL:     "https://iam.cloud.ibm.com",
							ApiKey:     "test-api-key",
							Host:       "icr.io",
						},
					},
					helmClient:     mockHelmClient,
					registryClient: mockRegistryClient,
					logger:         logger.NewNoop(),
				}

				return connector, mockHelmClient, mockRegistryClient
			},
			errorContains: "chart not found",
		},
		{
			name: "InstallError",
			setupMocks: func() (*HelmConnector, *MockHelmClient, *MockRegistryClient) {
				fakeClientSet := fake.NewSimpleClientset()
				ns := &corev1.Namespace{
					ObjectMeta: metav1.ObjectMeta{
						Name: "test-namespace",
					},
				}
				_, _ = fakeClientSet.CoreV1().Namespaces().Create(context.Background(), ns, metav1.CreateOptions{})

				mockHelmClient := new(MockHelmClient)
				mockHelmClient.On("Init", mock.Anything, "test-namespace").Return(nil)
				mockHelmClient.On("Install", mock.Anything, mock.Anything).Return(nil, fmt.Errorf("install failed"))

				mockRegistryClient := new(MockRegistryClient)
				mockRegistryClient.On("Login", "icr.io", "iamapikey", "test-api-key").Return(nil)
				mockRegistryClient.client = &registry.Client{}

				connector := &HelmConnector{
					namespace:   "test-namespace",
					releaseName: "brs-connector",
					chartRepo:   "oci://icr.io/ext/brs/brs-ds-connector-chart",
					clientSet:   fakeClientSet,
					restConfig:  &rest.Config{},
					config: &HelmKubeConnectorConfig{
						AuthConfig: &KubernetesAuthConfig{
							AuthMethod: AuthMethodAPIKey,
							IamURL:     "https://iam.cloud.ibm.com",
							ApiKey:     "test-api-key",
							Host:       "icr.io",
						},
					},
					helmClient:     mockHelmClient,
					registryClient: mockRegistryClient,
					logger:         logger.NewNoOpLogger(),
				}
				return connector, mockHelmClient, mockRegistryClient
			},
			errorContains: "install failed",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			connector, mockHelmClient, mockRegistryClient := tt.setupMocks()

			result, err := connector.Deploy(context.Background(), "test-token")

			assert.Error(t, err)
			assert.Nil(t, result)
			if tt.errorContains != "" {
				assert.Contains(t, err.Error(), tt.errorContains)
			}

			if mockHelmClient != nil {
				mockHelmClient.AssertExpectations(t)
			}
			if mockRegistryClient != nil {
				mockRegistryClient.AssertExpectations(t)
			}
		})
	}
}
func TestDeploy_Scheduling(t *testing.T) {
	tests := []struct {
		name           string
		tolerations    []map[string]interface{}
		nodeSelector   map[string]string
		resources      *ResourceRequirements
		serviceAccount string
		expectedValues []string // Keys expected in Helm values
		validateValues func(t *testing.T, values map[string]interface{})
	}{
		{
			name: "Tolerations_And_NodeSelector",
			tolerations: []map[string]interface{}{
				{
					"key":      "sdk-taint",
					"operator": "Equal",
					"value":    "connector",
					"effect":   "NoSchedule",
				},
			},
			nodeSelector: map[string]string{
				"workload": "connector",
			},
			expectedValues: []string{"tolerations", "nodeSelector"},
			validateValues: func(t *testing.T, values map[string]interface{}) {
				assert.Contains(t, values, "tolerations")
				assert.Contains(t, values, "nodeSelector")

				tolerationsVal := values["tolerations"].([]map[string]interface{})
				assert.Len(t, tolerationsVal, 1)
				assert.Equal(t, "sdk-taint", tolerationsVal[0]["key"])

				nodeSelectorVal := values["nodeSelector"].(map[string]string)
				assert.Equal(t, "connector", nodeSelectorVal["workload"])
			},
		},
		{
			name: "Resources_And_ServiceAccount",
			resources: &ResourceRequirements{
				Requests: &ResourceList{CPU: "500m", Memory: "512Mi"},
				Limits:   &ResourceList{CPU: "1000m", Memory: "1Gi"},
			},
			serviceAccount: "brs-connector-sa",
			expectedValues: []string{"resources", "serviceAccount"},
			validateValues: func(t *testing.T, values map[string]interface{}) {
				resourcesVal, ok := values["resources"]
				assert.True(t, ok, "resources should be present")
				resourcesMap := resourcesVal.(map[string]interface{})
				assert.Contains(t, resourcesMap, "requests")
				assert.Contains(t, resourcesMap, "limits")

				serviceAccountVal, ok := values["serviceAccount"]
				assert.True(t, ok, "serviceAccount should be present")
				serviceAccountMap := serviceAccountVal.(map[string]interface{})
				assert.Equal(t, "brs-connector-sa", serviceAccountMap["name"])
			},
		},
		{
			name:           "NodeSelector_Only",
			nodeSelector:   map[string]string{"workload": "connector"},
			expectedValues: []string{"nodeSelector"},
			validateValues: func(t *testing.T, values map[string]interface{}) {
				assert.Contains(t, values, "nodeSelector")
				assert.NotContains(t, values, "tolerations")
			},
		},
		{
			name: "Tolerations_Only",
			tolerations: []map[string]interface{}{
				{"key": "test", "operator": "Equal", "value": "true", "effect": "NoSchedule"},
			},
			expectedValues: []string{"tolerations"},
			validateValues: func(t *testing.T, values map[string]interface{}) {
				assert.Contains(t, values, "tolerations")
				assert.NotContains(t, values, "nodeSelector")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fakeClientSet := fake.NewSimpleClientset()
			ns := &corev1.Namespace{
				ObjectMeta: metav1.ObjectMeta{Name: "test-namespace"},
			}
			_, err := fakeClientSet.CoreV1().Namespaces().Create(context.Background(), ns, metav1.CreateOptions{})
			assert.NoError(t, err)

			mockHelmClient := new(MockHelmClient)
			mockHelmClient.On("Init", mock.Anything, "test-namespace").Return(nil)
			mockHelmClient.On("Install", mock.Anything, mock.MatchedBy(func(config *HelmInstallConfig) bool {
				for _, key := range tt.expectedValues {
					if _, ok := config.Values[key]; !ok {
						t.Errorf("Expected %s in Helm values, but not found", key)
						return false
					}
				}

				if tt.validateValues != nil {
					tt.validateValues(t, config.Values)
				}

				return true
			})).Return(&release.Release{
				Name:      "brs-connector",
				Namespace: "test-namespace",
			}, nil)

			mockRegistryClient := new(MockRegistryClient)
			mockRegistryClient.On("Login", "icr.io", "iamapikey", "test-api-key").Return(nil)
			mockRegistryClient.client = &registry.Client{}

			connector := &HelmConnector{
				namespace:       "test-namespace",
				releaseName:     "brs-connector",
				chartRepo:       "oci://icr.io/ext/brs/brs-ds-connector-chart",
				chartName:       "brs-connector",
				ImagePullPolicy: "IfNotPresent",
				replicas:        1,
				clientSet:       fakeClientSet,
				restConfig:      &rest.Config{},
				config: &HelmKubeConnectorConfig{
					Namespace:      "test-namespace",
					ClusterName:    "test-cluster",
					WaitTillDeploy: false,
					Tolerations:    tt.tolerations,
					NodeSelector:   tt.nodeSelector,
					Resources:      tt.resources,
					ServiceAccount: tt.serviceAccount,
					AuthConfig: &KubernetesAuthConfig{
						AuthMethod: AuthMethodAPIKey,
						IamURL:     "https://iam.cloud.ibm.com",
						ApiKey:     "test-api-key",
						Host:       "icr.io",
					},
				},
				helmClient:     mockHelmClient,
				registryClient: mockRegistryClient,
				logger:         logger.NewNoOpLogger(),
			}

			result, err := connector.Deploy(context.Background(), "test-token")
			assert.NoError(t, err)
			assert.NotNil(t, result)
			assert.Equal(t, "deployed", result.Status)
			assert.Contains(t, result.Message, "test-namespace")
			assert.True(t, connector.deployed)
			mockHelmClient.AssertExpectations(t)
			mockRegistryClient.AssertExpectations(t)
		})
	}
}

func TestSetupRegistry(t *testing.T) {
	tests := []struct {
		name          string
		connector     *HelmConnector
		expectError   bool
		errorContains string
	}{
		{
			name: "TokenAuth_NoOp",
			connector: &HelmConnector{
				config: &HelmKubeConnectorConfig{
					AuthConfig: &KubernetesAuthConfig{
						AuthMethod: AuthMethodToken,
					},
				},
			},
			expectError: false,
		},
		{
			name: "InvalidAuthConfig",
			connector: &HelmConnector{
				config: &HelmKubeConnectorConfig{
					AuthConfig: &VSIAuthConfig{ // Wrong auth config type
						AuthMethod: AuthMethodAPIKey,
					},
				},
			},
			expectError:   true,
			errorContains: "invalid kubernetes auth configuration",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.connector.setupRegistryClient(&action.Configuration{})

			if tt.expectError {
				assert.Error(t, err)
				if tt.errorContains != "" {
					assert.Contains(t, err.Error(), tt.errorContains)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestRegistryClient(t *testing.T) {
	tests := []struct {
		name         string
		testFunc     func(t *testing.T)
		expectError  bool
		validateFunc func(t *testing.T, wrapper *HelmRegistryClientWrapper, err error)
	}{
		{
			name: "NewHelmRegistryClient_Success",
			testFunc: func(t *testing.T) {
				client, err := NewHelmRegistryClient()
				assert.NoError(t, err)
				assert.NotNil(t, client)
				assert.NotNil(t, client.client)
			},
		},
		{
			name: "GetClient",
			testFunc: func(t *testing.T) {
				regClient, _ := registry.NewClient()
				wrapper := &HelmRegistryClientWrapper{
					client: regClient,
				}
				result := wrapper.GetClient()
				assert.Equal(t, regClient, result)
			},
		},
		{
			name: "Login_WrapperStructure",
			testFunc: func(t *testing.T) {
				wrapper, err := NewHelmRegistryClient()
				assert.NoError(t, err)
				assert.NotNil(t, wrapper)
				assert.NotNil(t, wrapper.client)
			},
		},
		{
			name: "Logout_NoError",
			testFunc: func(t *testing.T) {
				wrapper, _ := NewHelmRegistryClient()
				err := wrapper.Logout("test-host")
				_ = err
			},
		},
		{
			name: "Login_Error",
			testFunc: func(t *testing.T) {
				wrapper, _ := NewHelmRegistryClient()
				err := wrapper.Login("invalid-host", "user", "pass")
				assert.Error(t, err)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.testFunc(t)
		})
	}
}

func TestHelmClient(t *testing.T) {
	tests := []struct {
		name     string
		testFunc func(t *testing.T)
	}{
		{
			name: "NewDefaultHelmClient",
			testFunc: func(t *testing.T) {
				restConfig := &rest.Config{
					Host: "https://test-cluster.example.com",
				}
				client := NewDefaultHelmClient(restConfig, "test-namespace")
				assert.NotNil(t, client)
				assert.Equal(t, restConfig, client.restConfig)
				assert.Equal(t, "test-namespace", client.namespace)
			},
		},
		{
			name: "Init",
			testFunc: func(t *testing.T) {
				restConfig := &rest.Config{
					Host: "https://test-cluster.example.com",
				}
				client := NewDefaultHelmClient(restConfig, "test-namespace")
				actionConfig := new(action.Configuration)
				err := client.Init(actionConfig, "test-namespace")
				_ = err
			},
		},
		{
			name: "Install",
			testFunc: func(t *testing.T) {
				restConfig := &rest.Config{
					Host: "https://test-cluster.example.com",
				}
				client := NewDefaultHelmClient(restConfig, "test-namespace")
				actionConfig := new(action.Configuration)
				installConfig := &HelmInstallConfig{
					ReleaseName: "test-release",
					Namespace:   "test-namespace",
					ChartRef:    "/non/existent/chart",
					Version:     "",
					Values: map[string]interface{}{
						"test": "value",
					},
					Wait: false,
				}
				_, err := client.Install(actionConfig, installConfig)
				assert.Error(t, err)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.testFunc(t)
		})
	}
}
