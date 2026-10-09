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
	"io"
	"net/http"
	"testing"
	"time"

	"github.com/IBM/go-sdk-core/v5/core"
	backuprecoveryv1 "github.com/IBM/ibm-backup-recovery-sdk-go/backuprecoveryv1"
	"github.com/IBM/ibm-backup-recovery-sdk-go/migrationv2/common/logger"
	"github.com/IBM/ibm-backup-recovery-sdk-go/migrationv2/common/metrics"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"helm.sh/helm/v4/pkg/action"
	releasev1 "helm.sh/helm/v4/pkg/release/v1"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/fake"
	"k8s.io/client-go/rest"
	k8stesting "k8s.io/client-go/testing"
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
			name: "Delete_NilClientSet_ReturnsError",
			testFunc: func(t *testing.T) {
				connector := &HelmConnector{
					deployed: true,
					logger:   logger.NewNoOpLogger(),
					// clientSet intentionally nil — guard must return an error, not panic
				}
				ctx := context.Background()
				assert.NotPanics(t, func() {
					err := connector.Delete(ctx, "test-connector-id")
					assert.EqualError(t, err, "kubernetes client not initialized")
					// deployed must remain true — deletion did not complete
					assert.True(t, connector.deployed)
				})
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
			},
			expectError: false,
		},
		{
			name: "APIKey_MissingIamURL",
			config: &KubernetesAuthConfig{
				AuthMethod: AuthMethodAPIKey,
				ApiKey:     "test-api-key",
			},
			expectError:   true,
			errorContains: "Iam URL is required",
		},
		{
			name: "APIKey_MissingAPIKey",
			config: &KubernetesAuthConfig{
				AuthMethod: AuthMethodAPIKey,
				IamURL:     "https://iam.cloud.ibm.com",
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

func (m *MockHelmClient) Install(actionConfig *action.Configuration, config *HelmInstallConfig) (*releasev1.Release, error) {
	args := m.Called(actionConfig, config)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*releasev1.Release), args.Error(1)
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
	})).Return(&releasev1.Release{
		Name:      "brs-connector",
		Namespace: "test-namespace",
	}, nil)

	connector := &HelmConnector{
		namespace:       "test-namespace",
		releaseName:     "brs-connector",
		chartRepo:       "oci://icr.io/ext/brs/brs-ds-connector-chart",
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
			},
		},
		helmClient: mockHelmClient,
		logger:     logger.NewNoop(),
	}

	result, err := connector.Deploy(context.Background(), "test-token")

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, "deployed", result.Status)
	assert.Contains(t, result.Message, "test-namespace")
	assert.True(t, connector.deployed)
	mockHelmClient.AssertExpectations(t)
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
	})).Return(&releasev1.Release{
		Name:      "brs-connector",
		Namespace: "test-namespace",
	}, nil)

	connector := &HelmConnector{
		namespace:       "test-namespace",
		releaseName:     "brs-connector",
		chartRepo:       "oci://icr.io/ext/brs/brs-ds-connector-chart",
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
			},
		},
		helmClient: mockHelmClient,
		logger:     logger.NewNoop(),
	}

	result, err := connector.Deploy(context.Background(), "test-token")

	assert.NoError(t, err)
	assert.NotNil(t, result)
	mockHelmClient.AssertExpectations(t)
}

func TestDeploy_Errors(t *testing.T) {
	tests := []struct {
		name          string
		setupMocks    func() (*HelmConnector, *MockHelmClient)
		errorContains string
	}{
		{
			name: "AuthValidationError",
			setupMocks: func() (*HelmConnector, *MockHelmClient) {
				connector := &HelmConnector{
					config: &HelmKubeConnectorConfig{
						AuthConfig: &KubernetesAuthConfig{
							AuthMethod: AuthMethodAPIKey,
						},
					},
					logger: logger.NewNoOpLogger(),
				}
				return connector, nil
			},
			errorContains: "",
		},
		{
			name: "HelmInitError",
			setupMocks: func() (*HelmConnector, *MockHelmClient) {
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
						},
					},
					helmClient: mockHelmClient,
					logger:     logger.NewNoop(),
				}

				return connector, mockHelmClient
			},
			errorContains: "helm init failed",
		},
		{
			name: "LocateChartError",
			setupMocks: func() (*HelmConnector, *MockHelmClient) {
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
						},
					},
					helmClient: mockHelmClient,
					logger:     logger.NewNoop(),
				}

				return connector, mockHelmClient
			},
			errorContains: "chart not found",
		},
		{
			name: "InstallError",
			setupMocks: func() (*HelmConnector, *MockHelmClient) {
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
						},
					},
					helmClient: mockHelmClient,
					logger:     logger.NewNoOpLogger(),
				}
				return connector, mockHelmClient
			},
			errorContains: "install failed",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			connector, mockHelmClient := tt.setupMocks()

			result, err := connector.Deploy(context.Background(), "test-token")

			assert.Error(t, err)
			assert.Nil(t, result)
			if tt.errorContains != "" {
				assert.Contains(t, err.Error(), tt.errorContains)
			}

			if mockHelmClient != nil {
				mockHelmClient.AssertExpectations(t)
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
			})).Return(&releasev1.Release{
				Name:      "brs-connector",
				Namespace: "test-namespace",
			}, nil)

			connector := &HelmConnector{
				namespace:       "test-namespace",
				releaseName:     "brs-connector",
				chartRepo:       "oci://icr.io/ext/brs/brs-ds-connector-chart",
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
					},
				},
				helmClient: mockHelmClient,
				logger:     logger.NewNoOpLogger(),
			}

			result, err := connector.Deploy(context.Background(), "test-token")
			assert.NoError(t, err)
			assert.NotNil(t, result)
			assert.Equal(t, "deployed", result.Status)
			assert.Contains(t, result.Message, "test-namespace")
			assert.True(t, connector.deployed)
			mockHelmClient.AssertExpectations(t)
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
			name: "APIKey_Success",
			connector: &HelmConnector{
				config: &HelmKubeConnectorConfig{
					AuthConfig: &KubernetesAuthConfig{
						AuthMethod: AuthMethodAPIKey,
						IamURL:     "https://iam.cloud.ibm.com",
						ApiKey:     "test-api-key",
					},
				},
			},
			expectError: false,
		},
		{
			name: "TokenAuth_Success",
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
			name: "KubeconfigAuth_Success",
			connector: &HelmConnector{
				config: &HelmKubeConnectorConfig{
					AuthConfig: &KubernetesAuthConfig{
						AuthMethod:     AuthMethodKubeconfig,
						KubeconfigPath: "/path/to/kubeconfig",
					},
				},
			},
			expectError: false,
		},
		{
			name: "UnsupportedAuthMethod",
			connector: &HelmConnector{
				config: &HelmKubeConnectorConfig{
					AuthConfig: &VSIAuthConfig{
						AuthMethod: AuthMethodSSHKey,
					},
				},
			},
			expectError:   true,
			errorContains: "unsupported auth method",
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
	t.Run("NewHelmRegistryClient_Success", func(t *testing.T) {
		client, err := NewHelmRegistryClient()
		assert.NoError(t, err)
		assert.NotNil(t, client)
		assert.NotNil(t, client.client)
	})

	t.Run("GetClient", func(t *testing.T) {
		client, err := NewHelmRegistryClient()
		assert.NoError(t, err)
		result := client.GetClient()
		assert.NotNil(t, result)
		assert.Equal(t, client.client, result)
	})
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

// ---------------------------------------------------------------------------
// Deploy — namespace auto-creation
// ---------------------------------------------------------------------------

func TestDeploy_AutoCreatesNamespace(t *testing.T) {
	// Namespace does NOT exist — Deploy must create it before installing.
	fakeClientSet := fake.NewSimpleClientset()

	mockHelmClient := new(MockHelmClient)
	mockHelmClient.On("Init", mock.Anything, "new-namespace").Return(nil)
	mockHelmClient.On("Install", mock.Anything, mock.Anything).
		Return(&releasev1.Release{Name: "brs-connector", Namespace: "new-namespace"}, nil)

	connector := &HelmConnector{
		namespace:       "new-namespace",
		releaseName:     "brs-connector",
		chartRepo:       "oci://icr.io/ext/brs/brs-ds-connector-chart",
		ImagePullPolicy: "IfNotPresent",
		replicas:        1,
		clientSet:       fakeClientSet,
		restConfig:      &rest.Config{},
		config: &HelmKubeConnectorConfig{
			Namespace:   "new-namespace",
			ClusterName: "test-cluster",
			AuthConfig:  &KubernetesAuthConfig{AuthMethod: AuthMethodAPIKey, IamURL: "https://iam.cloud.ibm.com", ApiKey: "test-api-key"},
		},
		helmClient: mockHelmClient,
		logger:     logger.NewNoop(),
	}

	result, err := connector.Deploy(context.Background(), "test-token")
	assert.NoError(t, err)
	assert.NotNil(t, result)

	// Verify the namespace was actually created in the fake client.
	ctx := context.Background()
	ns, err := fakeClientSet.CoreV1().Namespaces().Get(ctx, "new-namespace", metav1.GetOptions{})
	assert.NoError(t, err)
	assert.Equal(t, "new-namespace", ns.Name)

	mockHelmClient.AssertExpectations(t)
}

func TestDeploy_NamespaceGetError_Non404_FailsDeploy(t *testing.T) {
	// When Namespace Get returns a non-404 error (e.g. 403 Forbidden), Deploy must fail immediately.
	fakeClientSet := fake.NewSimpleClientset()
	fakeClientSet.PrependReactor("get", "namespaces", func(action k8stesting.Action) (handled bool, ret runtime.Object, err error) {
		return true, nil, apierrors.NewForbidden(corev1.Resource("namespaces"), "test-namespace", fmt.Errorf("forbidden"))
	})

	mockHelmClient := new(MockHelmClient)

	connector := &HelmConnector{
		namespace:       "test-namespace",
		releaseName:     "brs-connector",
		chartRepo:       "oci://icr.io/ext/brs/brs-ds-connector-chart",
		ImagePullPolicy: "IfNotPresent",
		replicas:        1,
		clientSet:       fakeClientSet,
		restConfig:      &rest.Config{},
		config: &HelmKubeConnectorConfig{
			Namespace:   "test-namespace",
			ClusterName: "test-cluster",
			AuthConfig:  &KubernetesAuthConfig{AuthMethod: AuthMethodAPIKey, IamURL: "https://iam.cloud.ibm.com", ApiKey: "test-api-key"},
		},
		helmClient: mockHelmClient,
		logger:     logger.NewNoop(),
	}

	result, err := connector.Deploy(context.Background(), "test-token")
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "failed to check namespace \"test-namespace\"")
	assert.Contains(t, err.Error(), "forbidden")

	// Verify Helm Init / Install was never called
	mockHelmClient.AssertNotCalled(t, "Init", mock.Anything, mock.Anything)
	mockHelmClient.AssertNotCalled(t, "Install", mock.Anything, mock.Anything)
}

// ---------------------------------------------------------------------------
// Deploy — deployed flag stays false on Install error
// ---------------------------------------------------------------------------

func TestDeploy_DeployedFlag_FalseOnError(t *testing.T) {
	fakeClientSet := fake.NewSimpleClientset()
	_, _ = fakeClientSet.CoreV1().Namespaces().Create(
		context.Background(),
		&corev1.Namespace{ObjectMeta: metav1.ObjectMeta{Name: "test-namespace"}},
		metav1.CreateOptions{},
	)

	mockHelmClient := new(MockHelmClient)
	mockHelmClient.On("Init", mock.Anything, "test-namespace").Return(nil)
	mockHelmClient.On("Install", mock.Anything, mock.Anything).Return(nil, fmt.Errorf("install boom"))

	connector := &HelmConnector{
		namespace:       "test-namespace",
		releaseName:     "brs-connector",
		chartRepo:       "oci://icr.io/ext/brs/brs-ds-connector-chart",
		ImagePullPolicy: "IfNotPresent",
		replicas:        1,
		clientSet:       fakeClientSet,
		restConfig:      &rest.Config{},
		config: &HelmKubeConnectorConfig{
			Namespace:   "test-namespace",
			ClusterName: "test-cluster",
			AuthConfig:  &KubernetesAuthConfig{AuthMethod: AuthMethodAPIKey, IamURL: "https://iam.cloud.ibm.com", ApiKey: "test-api-key"},
		},
		helmClient: mockHelmClient,
		logger:     logger.NewNoop(),
	}

	result, err := connector.Deploy(context.Background(), "test-token")
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.False(t, connector.deployed, "deployed must remain false when Install fails")
}

// ---------------------------------------------------------------------------
// Deploy — WaitTillDeploy is forwarded to HelmInstallConfig
// ---------------------------------------------------------------------------

func TestDeploy_WaitTillDeploy_ForwardedToInstallConfig(t *testing.T) {
	fakeClientSet := fake.NewSimpleClientset()
	_, _ = fakeClientSet.CoreV1().Namespaces().Create(
		context.Background(),
		&corev1.Namespace{ObjectMeta: metav1.ObjectMeta{Name: "test-namespace"}},
		metav1.CreateOptions{},
	)

	mockHelmClient := new(MockHelmClient)
	mockHelmClient.On("Init", mock.Anything, "test-namespace").Return(nil)
	mockHelmClient.On("Install", mock.Anything, mock.MatchedBy(func(cfg *HelmInstallConfig) bool {
		return cfg.Wait == true
	})).Return(&releasev1.Release{Name: "brs-connector", Namespace: "test-namespace"}, nil)

	connector := &HelmConnector{
		namespace:       "test-namespace",
		releaseName:     "brs-connector",
		chartRepo:       "oci://icr.io/ext/brs/brs-ds-connector-chart",
		ImagePullPolicy: "IfNotPresent",
		replicas:        1,
		clientSet:       fakeClientSet,
		restConfig:      &rest.Config{},
		config: &HelmKubeConnectorConfig{
			Namespace:      "test-namespace",
			ClusterName:    "test-cluster",
			WaitTillDeploy: true,
			AuthConfig:     &KubernetesAuthConfig{AuthMethod: AuthMethodAPIKey, IamURL: "https://iam.cloud.ibm.com", ApiKey: "test-api-key"},
		},
		helmClient: mockHelmClient,
		logger:     logger.NewNoop(),
	}

	result, err := connector.Deploy(context.Background(), "test-token")
	assert.NoError(t, err)
	assert.NotNil(t, result)
	mockHelmClient.AssertExpectations(t)
}

// ---------------------------------------------------------------------------
// Deploy — CustomValues override base values
// ---------------------------------------------------------------------------

func TestDeploy_CustomValues_OverrideBaseValues(t *testing.T) {
	fakeClientSet := fake.NewSimpleClientset()
	_, _ = fakeClientSet.CoreV1().Namespaces().Create(
		context.Background(),
		&corev1.Namespace{ObjectMeta: metav1.ObjectMeta{Name: "test-namespace"}},
		metav1.CreateOptions{},
	)

	mockHelmClient := new(MockHelmClient)
	mockHelmClient.On("Init", mock.Anything, "test-namespace").Return(nil)
	mockHelmClient.On("Install", mock.Anything, mock.MatchedBy(func(cfg *HelmInstallConfig) bool {
		// CustomValues["replicaCount"] = 5 must override the struct replicas field (1).
		return cfg.Values["replicaCount"] == 5 &&
			cfg.Values["extraKey"] == "extraVal"
	})).Return(&releasev1.Release{Name: "brs-connector", Namespace: "test-namespace"}, nil)

	connector := &HelmConnector{
		namespace:       "test-namespace",
		releaseName:     "brs-connector",
		chartRepo:       "oci://icr.io/ext/brs/brs-ds-connector-chart",
		ImagePullPolicy: "IfNotPresent",
		replicas:        1,
		clientSet:       fakeClientSet,
		restConfig:      &rest.Config{},
		config: &HelmKubeConnectorConfig{
			Namespace:   "test-namespace",
			ClusterName: "test-cluster",
			AuthConfig:  &KubernetesAuthConfig{AuthMethod: AuthMethodAPIKey, IamURL: "https://iam.cloud.ibm.com", ApiKey: "test-api-key"},
			CustomValues: map[string]interface{}{
				"replicaCount": 5,
				"extraKey":     "extraVal",
			},
		},
		helmClient: mockHelmClient,
		logger:     logger.NewNoop(),
	}

	result, err := connector.Deploy(context.Background(), "test-token")
	assert.NoError(t, err)
	assert.NotNil(t, result)
	mockHelmClient.AssertExpectations(t)
}

// ---------------------------------------------------------------------------
// Deploy — ImagePullSecrets and SCCEnabled configuration
// ---------------------------------------------------------------------------

func TestDeploy_ImagePullSecrets_And_SCCEnabled(t *testing.T) {
	sccTrue := true
	fakeClientSet := fake.NewSimpleClientset()
	_, _ = fakeClientSet.CoreV1().Namespaces().Create(
		context.Background(),
		&corev1.Namespace{ObjectMeta: metav1.ObjectMeta{Name: "test-namespace"}},
		metav1.CreateOptions{},
	)

	mockHelmClient := new(MockHelmClient)
	mockHelmClient.On("Init", mock.Anything, "test-namespace").Return(nil)
	mockHelmClient.On("Install", mock.Anything, mock.MatchedBy(func(cfg *HelmInstallConfig) bool {
		secrets, ok := cfg.Values["imagePullSecrets"].([]string)
		if !ok || len(secrets) != 2 || secrets[0] != "custom-secret-1" || secrets[1] != "custom-secret-2" {
			return false
		}
		deployPlatform, ok := cfg.Values["deploymentPlatform"].(map[string]interface{})
		if !ok {
			return false
		}
		rocp, ok := deployPlatform["rocp"].(map[string]interface{})
		if !ok {
			return false
		}
		scc, ok := rocp["sccEnabled"].(bool)
		return ok && scc == true
	})).Return(&releasev1.Release{Name: "brs-connector", Namespace: "test-namespace"}, nil)

	connector := &HelmConnector{
		namespace:       "test-namespace",
		releaseName:     "brs-connector",
		chartRepo:       "oci://icr.io/ext/brs/brs-ds-connector-chart",
		ImagePullPolicy: "IfNotPresent",
		replicas:        1,
		clientSet:       fakeClientSet,
		restConfig:      &rest.Config{},
		config: &HelmKubeConnectorConfig{
			Namespace:        "test-namespace",
			ClusterName:      "test-cluster",
			ImagePullSecrets: []string{"custom-secret-1", "custom-secret-2"},
			SCCEnabled:       &sccTrue,
			AuthConfig:       &KubernetesAuthConfig{AuthMethod: AuthMethodAPIKey, IamURL: "https://iam.cloud.ibm.com", ApiKey: "test-api-key"},
		},
		helmClient: mockHelmClient,
		logger:     logger.NewNoop(),
	}

	result, err := connector.Deploy(context.Background(), "test-token")
	assert.NoError(t, err)
	assert.NotNil(t, result)
	mockHelmClient.AssertExpectations(t)
}

func TestDeploy_AlreadyDeployed_ReturnsError(t *testing.T) {
	connector := &HelmConnector{
		namespace:   "test-namespace",
		releaseName: "brs-connector",
		deployed:    true,
		logger:      logger.NewNoop(),
	}

	result, err := connector.Deploy(context.Background(), "test-token")
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "connector is already deployed")
}

// ---------------------------------------------------------------------------
// Deploy — Resources partial: only Requests set, no Limits
// ---------------------------------------------------------------------------

func TestDeploy_Resources_RequestsOnly(t *testing.T) {
	fakeClientSet := fake.NewSimpleClientset()
	_, _ = fakeClientSet.CoreV1().Namespaces().Create(
		context.Background(),
		&corev1.Namespace{ObjectMeta: metav1.ObjectMeta{Name: "test-namespace"}},
		metav1.CreateOptions{},
	)

	mockHelmClient := new(MockHelmClient)
	mockHelmClient.On("Init", mock.Anything, "test-namespace").Return(nil)
	mockHelmClient.On("Install", mock.Anything, mock.MatchedBy(func(cfg *HelmInstallConfig) bool {
		res, ok := cfg.Values["resources"].(map[string]interface{})
		if !ok {
			return false
		}
		_, hasRequests := res["requests"]
		_, hasLimits := res["limits"]
		return hasRequests && !hasLimits
	})).Return(&releasev1.Release{Name: "brs-connector", Namespace: "test-namespace"}, nil)

	connector := &HelmConnector{
		namespace:       "test-namespace",
		releaseName:     "brs-connector",
		chartRepo:       "oci://icr.io/ext/brs/brs-ds-connector-chart",
		ImagePullPolicy: "IfNotPresent",
		replicas:        1,
		clientSet:       fakeClientSet,
		restConfig:      &rest.Config{},
		config: &HelmKubeConnectorConfig{
			Namespace:   "test-namespace",
			ClusterName: "test-cluster",
			AuthConfig:  &KubernetesAuthConfig{AuthMethod: AuthMethodAPIKey, IamURL: "https://iam.cloud.ibm.com", ApiKey: "test-api-key"},
			Resources: &ResourceRequirements{
				Requests: &ResourceList{CPU: "250m", Memory: "256Mi"},
				Limits:   nil, // no limits
			},
		},
		helmClient: mockHelmClient,
		logger:     logger.NewNoop(),
	}

	result, err := connector.Deploy(context.Background(), "test-token")
	assert.NoError(t, err)
	assert.NotNil(t, result)
	mockHelmClient.AssertExpectations(t)
}

// Deploy — Resources partial: only Limits set, no Requests
func TestDeploy_Resources_LimitsOnly(t *testing.T) {
	fakeClientSet := fake.NewSimpleClientset()
	_, _ = fakeClientSet.CoreV1().Namespaces().Create(
		context.Background(),
		&corev1.Namespace{ObjectMeta: metav1.ObjectMeta{Name: "test-namespace"}},
		metav1.CreateOptions{},
	)

	mockHelmClient := new(MockHelmClient)
	mockHelmClient.On("Init", mock.Anything, "test-namespace").Return(nil)
	mockHelmClient.On("Install", mock.Anything, mock.MatchedBy(func(cfg *HelmInstallConfig) bool {
		res, ok := cfg.Values["resources"].(map[string]interface{})
		if !ok {
			return false
		}
		_, hasRequests := res["requests"]
		_, hasLimits := res["limits"]
		return !hasRequests && hasLimits
	})).Return(&releasev1.Release{Name: "brs-connector", Namespace: "test-namespace"}, nil)

	connector := &HelmConnector{
		namespace:       "test-namespace",
		releaseName:     "brs-connector",
		chartRepo:       "oci://icr.io/ext/brs/brs-ds-connector-chart",
		ImagePullPolicy: "IfNotPresent",
		replicas:        1,
		clientSet:       fakeClientSet,
		restConfig:      &rest.Config{},
		config: &HelmKubeConnectorConfig{
			Namespace:   "test-namespace",
			ClusterName: "test-cluster",
			AuthConfig:  &KubernetesAuthConfig{AuthMethod: AuthMethodAPIKey, IamURL: "https://iam.cloud.ibm.com", ApiKey: "test-api-key"},
			Resources: &ResourceRequirements{
				Requests: nil, // no requests
				Limits:   &ResourceList{CPU: "1000m", Memory: "1Gi"},
			},
		},
		helmClient: mockHelmClient,
		logger:     logger.NewNoop(),
	}

	result, err := connector.Deploy(context.Background(), "test-token")
	assert.NoError(t, err)
	assert.NotNil(t, result)
	mockHelmClient.AssertExpectations(t)
}

// Deploy — Resources with empty CPU/Memory strings produces no sub-map
func TestDeploy_Resources_EmptySubfields_Omitted(t *testing.T) {
	fakeClientSet := fake.NewSimpleClientset()
	_, _ = fakeClientSet.CoreV1().Namespaces().Create(
		context.Background(),
		&corev1.Namespace{ObjectMeta: metav1.ObjectMeta{Name: "test-namespace"}},
		metav1.CreateOptions{},
	)

	mockHelmClient := new(MockHelmClient)
	mockHelmClient.On("Init", mock.Anything, "test-namespace").Return(nil)
	// Both Requests and Limits have empty CPU and Memory → resources map must be absent entirely.
	mockHelmClient.On("Install", mock.Anything, mock.MatchedBy(func(cfg *HelmInstallConfig) bool {
		_, hasResources := cfg.Values["resources"]
		return !hasResources
	})).Return(&releasev1.Release{Name: "brs-connector", Namespace: "test-namespace"}, nil)

	connector := &HelmConnector{
		namespace:       "test-namespace",
		releaseName:     "brs-connector",
		chartRepo:       "oci://icr.io/ext/brs/brs-ds-connector-chart",
		ImagePullPolicy: "IfNotPresent",
		replicas:        1,
		clientSet:       fakeClientSet,
		restConfig:      &rest.Config{},
		config: &HelmKubeConnectorConfig{
			Namespace:   "test-namespace",
			ClusterName: "test-cluster",
			AuthConfig:  &KubernetesAuthConfig{AuthMethod: AuthMethodAPIKey, IamURL: "https://iam.cloud.ibm.com", ApiKey: "test-api-key"},
			Resources: &ResourceRequirements{
				Requests: &ResourceList{CPU: "", Memory: ""},
				Limits:   &ResourceList{CPU: "", Memory: ""},
			},
		},
		helmClient: mockHelmClient,
		logger:     logger.NewNoop(),
	}

	result, err := connector.Deploy(context.Background(), "test-token")
	assert.NoError(t, err)
	assert.NotNil(t, result)
	mockHelmClient.AssertExpectations(t)
}

// ---------------------------------------------------------------------------
// Delete — success path
// ---------------------------------------------------------------------------

func TestDelete_Success(t *testing.T) {
	fakeClientSet := fake.NewSimpleClientset()
	_, _ = fakeClientSet.CoreV1().Namespaces().Create(
		context.Background(),
		&corev1.Namespace{ObjectMeta: metav1.ObjectMeta{Name: "test-namespace"}},
		metav1.CreateOptions{},
	)

	connector := &HelmConnector{
		namespace: "test-namespace",
		clientSet: fakeClientSet,
		deployed:  true,
		logger:    logger.NewNoop(),
	}

	err := connector.Delete(context.Background(), "test-connector-id")
	assert.NoError(t, err)
	assert.False(t, connector.deployed, "deployed must be false after successful Delete")

	// Namespace must be gone from the fake store.
	_, getErr := fakeClientSet.CoreV1().Namespaces().Get(
		context.Background(), "test-namespace", metav1.GetOptions{},
	)
	assert.Error(t, getErr, "namespace should no longer exist after Delete")
}

// ---------------------------------------------------------------------------
// MockBRSClientWrapper — minimal implementation of BRSClientWrapperInterface
// used by resolveChartRef unit tests.
// ---------------------------------------------------------------------------

type mockBRSWrapper struct {
	client   *mockBRSClientForMetadata
	tenantID string
}

func (m *mockBRSWrapper) GetTenantId() string                               { return m.tenantID }
func (m *mockBRSWrapper) GetBRSClient() backuprecoveryv1.BRSClientInterface { return m.client }
func (m *mockBRSWrapper) GetRegion() string                                 { return "us-south" }
func (m *mockBRSWrapper) GetCRN() string                                    { return "crn:test" }

// mockBRSClientForMetadata stubs only GetConnectorMetadata; all other methods
// are provided by embeddedNoOpBRSClient so the interface is satisfied.
type mockBRSClientForMetadata struct {
	noOpBRSClient
	result *backuprecoveryv1.ConnectorMetadata
	err    error
}

func (m *mockBRSClientForMetadata) GetConnectorMetadata(
	opts *backuprecoveryv1.GetConnectorMetadataOptions,
) (*backuprecoveryv1.ConnectorMetadata, *core.DetailedResponse, error) {
	return m.result, nil, m.err
}

func (m *mockBRSClientForMetadata) GetConnectorMetadataWithContext(
	_ context.Context,
	opts *backuprecoveryv1.GetConnectorMetadataOptions,
) (*backuprecoveryv1.ConnectorMetadata, *core.DetailedResponse, error) {
	return m.result, nil, m.err
}

// noOpBRSClient satisfies backuprecoveryv1.BRSClientInterface with no-op stubs
// for every method we don't need in these tests.
type noOpBRSClient struct{}

func (n *noOpBRSClient) Clone() *backuprecoveryv1.BackupRecoveryV1 { return nil }
func (n *noOpBRSClient) SetServiceURL(url string) error            { return nil }
func (n *noOpBRSClient) GetServiceURL() string                     { return "" }
func (n *noOpBRSClient) SetDefaultHeaders(h http.Header)           {}
func (n *noOpBRSClient) SetEnableGzipCompression(bool)             {}
func (n *noOpBRSClient) GetEnableGzipCompression() bool            { return false }
func (n *noOpBRSClient) EnableRetries(int, time.Duration)          {}
func (n *noOpBRSClient) DisableRetries()                           {}

func (n *noOpBRSClient) DownloadAgent(*backuprecoveryv1.DownloadAgentOptions) (io.ReadCloser, *core.DetailedResponse, error) {
	return nil, nil, nil
}
func (n *noOpBRSClient) DownloadAgentWithContext(context.Context, *backuprecoveryv1.DownloadAgentOptions) (io.ReadCloser, *core.DetailedResponse, error) {
	return nil, nil, nil
}
func (n *noOpBRSClient) GetUpgradeTasks(*backuprecoveryv1.GetUpgradeTasksOptions) (*backuprecoveryv1.AgentUpgradeTaskStates, *core.DetailedResponse, error) {
	return nil, nil, nil
}
func (n *noOpBRSClient) GetUpgradeTasksWithContext(context.Context, *backuprecoveryv1.GetUpgradeTasksOptions) (*backuprecoveryv1.AgentUpgradeTaskStates, *core.DetailedResponse, error) {
	return nil, nil, nil
}
func (n *noOpBRSClient) CreateUpgradeTask(*backuprecoveryv1.CreateUpgradeTaskOptions) (*backuprecoveryv1.AgentUpgradeTaskState, *core.DetailedResponse, error) {
	return nil, nil, nil
}
func (n *noOpBRSClient) CreateUpgradeTaskWithContext(context.Context, *backuprecoveryv1.CreateUpgradeTaskOptions) (*backuprecoveryv1.AgentUpgradeTaskState, *core.DetailedResponse, error) {
	return nil, nil, nil
}
func (n *noOpBRSClient) ListProtectionSources(*backuprecoveryv1.ListProtectionSourcesOptions) ([]backuprecoveryv1.ProtectionSourceNodes, *core.DetailedResponse, error) {
	return nil, nil, nil
}
func (n *noOpBRSClient) ListProtectionSourcesWithContext(context.Context, *backuprecoveryv1.ListProtectionSourcesOptions) ([]backuprecoveryv1.ProtectionSourceNodes, *core.DetailedResponse, error) {
	return nil, nil, nil
}
func (n *noOpBRSClient) ListProtectionSourcesRegistrationInfo(*backuprecoveryv1.ListProtectionSourcesRegistrationInfoOptions) (*backuprecoveryv1.GetRegistrationInfoResponse, *core.DetailedResponse, error) {
	return nil, nil, nil
}
func (n *noOpBRSClient) ListProtectionSourcesRegistrationInfoWithContext(context.Context, *backuprecoveryv1.ListProtectionSourcesRegistrationInfoOptions) (*backuprecoveryv1.GetRegistrationInfoResponse, *core.DetailedResponse, error) {
	return nil, nil, nil
}
func (n *noOpBRSClient) GetDataSourceConnections(*backuprecoveryv1.GetDataSourceConnectionsOptions) (*backuprecoveryv1.DataSourceConnectionList, *core.DetailedResponse, error) {
	return nil, nil, nil
}
func (n *noOpBRSClient) GetDataSourceConnectionsWithContext(context.Context, *backuprecoveryv1.GetDataSourceConnectionsOptions) (*backuprecoveryv1.DataSourceConnectionList, *core.DetailedResponse, error) {
	return nil, nil, nil
}
func (n *noOpBRSClient) CreateDataSourceConnection(*backuprecoveryv1.CreateDataSourceConnectionOptions) (*backuprecoveryv1.DataSourceConnection, *core.DetailedResponse, error) {
	return nil, nil, nil
}
func (n *noOpBRSClient) CreateDataSourceConnectionWithContext(context.Context, *backuprecoveryv1.CreateDataSourceConnectionOptions) (*backuprecoveryv1.DataSourceConnection, *core.DetailedResponse, error) {
	return nil, nil, nil
}
func (n *noOpBRSClient) DeleteDataSourceConnection(*backuprecoveryv1.DeleteDataSourceConnectionOptions) (*core.DetailedResponse, error) {
	return nil, nil
}
func (n *noOpBRSClient) DeleteDataSourceConnectionWithContext(context.Context, *backuprecoveryv1.DeleteDataSourceConnectionOptions) (*core.DetailedResponse, error) {
	return nil, nil
}
func (n *noOpBRSClient) PatchDataSourceConnection(*backuprecoveryv1.PatchDataSourceConnectionOptions) (*backuprecoveryv1.DataSourceConnection, *core.DetailedResponse, error) {
	return nil, nil, nil
}
func (n *noOpBRSClient) PatchDataSourceConnectionWithContext(context.Context, *backuprecoveryv1.PatchDataSourceConnectionOptions) (*backuprecoveryv1.DataSourceConnection, *core.DetailedResponse, error) {
	return nil, nil, nil
}
func (n *noOpBRSClient) GenerateDataSourceConnectionRegistrationToken(*backuprecoveryv1.GenerateDataSourceConnectionRegistrationTokenOptions) (*string, *core.DetailedResponse, error) {
	return nil, nil, nil
}
func (n *noOpBRSClient) GenerateDataSourceConnectionRegistrationTokenWithContext(context.Context, *backuprecoveryv1.GenerateDataSourceConnectionRegistrationTokenOptions) (*string, *core.DetailedResponse, error) {
	return nil, nil, nil
}
func (n *noOpBRSClient) GetDataSourceConnectors(*backuprecoveryv1.GetDataSourceConnectorsOptions) (*backuprecoveryv1.DataSourceConnectorList, *core.DetailedResponse, error) {
	return nil, nil, nil
}
func (n *noOpBRSClient) GetDataSourceConnectorsWithContext(context.Context, *backuprecoveryv1.GetDataSourceConnectorsOptions) (*backuprecoveryv1.DataSourceConnectorList, *core.DetailedResponse, error) {
	return nil, nil, nil
}
func (n *noOpBRSClient) GetConnectorMetadata(*backuprecoveryv1.GetConnectorMetadataOptions) (*backuprecoveryv1.ConnectorMetadata, *core.DetailedResponse, error) {
	return nil, nil, nil
}
func (n *noOpBRSClient) GetConnectorMetadataWithContext(context.Context, *backuprecoveryv1.GetConnectorMetadataOptions) (*backuprecoveryv1.ConnectorMetadata, *core.DetailedResponse, error) {
	return nil, nil, nil
}
func (n *noOpBRSClient) DeleteDataSourceConnector(*backuprecoveryv1.DeleteDataSourceConnectorOptions) (*core.DetailedResponse, error) {
	return nil, nil
}
func (n *noOpBRSClient) DeleteDataSourceConnectorWithContext(context.Context, *backuprecoveryv1.DeleteDataSourceConnectorOptions) (*core.DetailedResponse, error) {
	return nil, nil
}
func (n *noOpBRSClient) PatchDataSourceConnector(*backuprecoveryv1.PatchDataSourceConnectorOptions) (*backuprecoveryv1.DataSourceConnector, *core.DetailedResponse, error) {
	return nil, nil, nil
}
func (n *noOpBRSClient) PatchDataSourceConnectorWithContext(context.Context, *backuprecoveryv1.PatchDataSourceConnectorOptions) (*backuprecoveryv1.DataSourceConnector, *core.DetailedResponse, error) {
	return nil, nil, nil
}
func (n *noOpBRSClient) GetObjectSnapshots(*backuprecoveryv1.GetObjectSnapshotsOptions) (*backuprecoveryv1.GetObjectSnapshotsResponse, *core.DetailedResponse, error) {
	return nil, nil, nil
}
func (n *noOpBRSClient) GetObjectSnapshotsWithContext(context.Context, *backuprecoveryv1.GetObjectSnapshotsOptions) (*backuprecoveryv1.GetObjectSnapshotsResponse, *core.DetailedResponse, error) {
	return nil, nil, nil
}
func (n *noOpBRSClient) GetProtectionPolicies(*backuprecoveryv1.GetProtectionPoliciesOptions) (*backuprecoveryv1.ProtectionPoliciesResponse, *core.DetailedResponse, error) {
	return nil, nil, nil
}
func (n *noOpBRSClient) GetProtectionPoliciesWithContext(context.Context, *backuprecoveryv1.GetProtectionPoliciesOptions) (*backuprecoveryv1.ProtectionPoliciesResponse, *core.DetailedResponse, error) {
	return nil, nil, nil
}
func (n *noOpBRSClient) CreateProtectionPolicy(*backuprecoveryv1.CreateProtectionPolicyOptions) (*backuprecoveryv1.ProtectionPolicyResponse, *core.DetailedResponse, error) {
	return nil, nil, nil
}
func (n *noOpBRSClient) CreateProtectionPolicyWithContext(context.Context, *backuprecoveryv1.CreateProtectionPolicyOptions) (*backuprecoveryv1.ProtectionPolicyResponse, *core.DetailedResponse, error) {
	return nil, nil, nil
}
func (n *noOpBRSClient) GetProtectionPolicyByID(*backuprecoveryv1.GetProtectionPolicyByIdOptions) (*backuprecoveryv1.ProtectionPolicyResponse, *core.DetailedResponse, error) {
	return nil, nil, nil
}
func (n *noOpBRSClient) GetProtectionPolicyByIDWithContext(context.Context, *backuprecoveryv1.GetProtectionPolicyByIdOptions) (*backuprecoveryv1.ProtectionPolicyResponse, *core.DetailedResponse, error) {
	return nil, nil, nil
}
func (n *noOpBRSClient) UpdateProtectionPolicy(*backuprecoveryv1.UpdateProtectionPolicyOptions) (*backuprecoveryv1.ProtectionPolicyResponse, *core.DetailedResponse, error) {
	return nil, nil, nil
}
func (n *noOpBRSClient) UpdateProtectionPolicyWithContext(context.Context, *backuprecoveryv1.UpdateProtectionPolicyOptions) (*backuprecoveryv1.ProtectionPolicyResponse, *core.DetailedResponse, error) {
	return nil, nil, nil
}
func (n *noOpBRSClient) DeleteProtectionPolicy(*backuprecoveryv1.DeleteProtectionPolicyOptions) (*core.DetailedResponse, error) {
	return nil, nil
}
func (n *noOpBRSClient) DeleteProtectionPolicyWithContext(context.Context, *backuprecoveryv1.DeleteProtectionPolicyOptions) (*core.DetailedResponse, error) {
	return nil, nil
}
func (n *noOpBRSClient) GetProtectionGroups(*backuprecoveryv1.GetProtectionGroupsOptions) (*backuprecoveryv1.ProtectionGroupsResponse, *core.DetailedResponse, error) {
	return nil, nil, nil
}
func (n *noOpBRSClient) GetProtectionGroupsWithContext(context.Context, *backuprecoveryv1.GetProtectionGroupsOptions) (*backuprecoveryv1.ProtectionGroupsResponse, *core.DetailedResponse, error) {
	return nil, nil, nil
}
func (n *noOpBRSClient) CreateProtectionGroup(*backuprecoveryv1.CreateProtectionGroupOptions) (*backuprecoveryv1.ProtectionGroupResponse, *core.DetailedResponse, error) {
	return nil, nil, nil
}
func (n *noOpBRSClient) CreateProtectionGroupWithContext(context.Context, *backuprecoveryv1.CreateProtectionGroupOptions) (*backuprecoveryv1.ProtectionGroupResponse, *core.DetailedResponse, error) {
	return nil, nil, nil
}
func (n *noOpBRSClient) GetProtectionGroupByID(*backuprecoveryv1.GetProtectionGroupByIdOptions) (*backuprecoveryv1.ProtectionGroupResponse, *core.DetailedResponse, error) {
	return nil, nil, nil
}
func (n *noOpBRSClient) GetProtectionGroupByIDWithContext(context.Context, *backuprecoveryv1.GetProtectionGroupByIdOptions) (*backuprecoveryv1.ProtectionGroupResponse, *core.DetailedResponse, error) {
	return nil, nil, nil
}
func (n *noOpBRSClient) UpdateProtectionGroup(*backuprecoveryv1.UpdateProtectionGroupOptions) (*backuprecoveryv1.ProtectionGroupResponse, *core.DetailedResponse, error) {
	return nil, nil, nil
}
func (n *noOpBRSClient) UpdateProtectionGroupWithContext(context.Context, *backuprecoveryv1.UpdateProtectionGroupOptions) (*backuprecoveryv1.ProtectionGroupResponse, *core.DetailedResponse, error) {
	return nil, nil, nil
}
func (n *noOpBRSClient) DeleteProtectionGroup(*backuprecoveryv1.DeleteProtectionGroupOptions) (*core.DetailedResponse, error) {
	return nil, nil
}
func (n *noOpBRSClient) DeleteProtectionGroupWithContext(context.Context, *backuprecoveryv1.DeleteProtectionGroupOptions) (*core.DetailedResponse, error) {
	return nil, nil
}
func (n *noOpBRSClient) GetProtectionGroupRuns(*backuprecoveryv1.GetProtectionGroupRunsOptions) (*backuprecoveryv1.ProtectionGroupRunsResponse, *core.DetailedResponse, error) {
	return nil, nil, nil
}
func (n *noOpBRSClient) GetProtectionGroupRunsWithContext(context.Context, *backuprecoveryv1.GetProtectionGroupRunsOptions) (*backuprecoveryv1.ProtectionGroupRunsResponse, *core.DetailedResponse, error) {
	return nil, nil, nil
}
func (n *noOpBRSClient) UpdateProtectionGroupRun(*backuprecoveryv1.UpdateProtectionGroupRunOptions) (*backuprecoveryv1.UpdateProtectionGroupRunResponse, *core.DetailedResponse, error) {
	return nil, nil, nil
}
func (n *noOpBRSClient) UpdateProtectionGroupRunWithContext(context.Context, *backuprecoveryv1.UpdateProtectionGroupRunOptions) (*backuprecoveryv1.UpdateProtectionGroupRunResponse, *core.DetailedResponse, error) {
	return nil, nil, nil
}
func (n *noOpBRSClient) CreateProtectionGroupRun(*backuprecoveryv1.CreateProtectionGroupRunOptions) (*backuprecoveryv1.CreateProtectionGroupRunResponse, *core.DetailedResponse, error) {
	return nil, nil, nil
}
func (n *noOpBRSClient) CreateProtectionGroupRunWithContext(context.Context, *backuprecoveryv1.CreateProtectionGroupRunOptions) (*backuprecoveryv1.CreateProtectionGroupRunResponse, *core.DetailedResponse, error) {
	return nil, nil, nil
}
func (n *noOpBRSClient) PerformActionOnProtectionGroupRun(*backuprecoveryv1.PerformActionOnProtectionGroupRunOptions) (*backuprecoveryv1.PerformRunActionResponse, *core.DetailedResponse, error) {
	return nil, nil, nil
}
func (n *noOpBRSClient) PerformActionOnProtectionGroupRunWithContext(context.Context, *backuprecoveryv1.PerformActionOnProtectionGroupRunOptions) (*backuprecoveryv1.PerformRunActionResponse, *core.DetailedResponse, error) {
	return nil, nil, nil
}
func (n *noOpBRSClient) GetProtectionGroupRun(*backuprecoveryv1.GetProtectionGroupRunOptions) (*backuprecoveryv1.ProtectionGroupRun, *core.DetailedResponse, error) {
	return nil, nil, nil
}
func (n *noOpBRSClient) GetProtectionGroupRunWithContext(context.Context, *backuprecoveryv1.GetProtectionGroupRunOptions) (*backuprecoveryv1.ProtectionGroupRun, *core.DetailedResponse, error) {
	return nil, nil, nil
}
func (n *noOpBRSClient) GetRecoveries(*backuprecoveryv1.GetRecoveriesOptions) (*backuprecoveryv1.RecoveriesResponse, *core.DetailedResponse, error) {
	return nil, nil, nil
}
func (n *noOpBRSClient) GetRecoveriesWithContext(context.Context, *backuprecoveryv1.GetRecoveriesOptions) (*backuprecoveryv1.RecoveriesResponse, *core.DetailedResponse, error) {
	return nil, nil, nil
}
func (n *noOpBRSClient) CreateRecovery(*backuprecoveryv1.CreateRecoveryOptions) (*backuprecoveryv1.Recovery, *core.DetailedResponse, error) {
	return nil, nil, nil
}
func (n *noOpBRSClient) CreateRecoveryWithContext(context.Context, *backuprecoveryv1.CreateRecoveryOptions) (*backuprecoveryv1.Recovery, *core.DetailedResponse, error) {
	return nil, nil, nil
}
func (n *noOpBRSClient) CreateDownloadFilesAndFoldersRecovery(*backuprecoveryv1.CreateDownloadFilesAndFoldersRecoveryOptions) (*backuprecoveryv1.Recovery, *core.DetailedResponse, error) {
	return nil, nil, nil
}
func (n *noOpBRSClient) CreateDownloadFilesAndFoldersRecoveryWithContext(context.Context, *backuprecoveryv1.CreateDownloadFilesAndFoldersRecoveryOptions) (*backuprecoveryv1.Recovery, *core.DetailedResponse, error) {
	return nil, nil, nil
}
func (n *noOpBRSClient) GetRecoveryByID(*backuprecoveryv1.GetRecoveryByIdOptions) (*backuprecoveryv1.Recovery, *core.DetailedResponse, error) {
	return nil, nil, nil
}
func (n *noOpBRSClient) GetRecoveryByIDWithContext(context.Context, *backuprecoveryv1.GetRecoveryByIdOptions) (*backuprecoveryv1.Recovery, *core.DetailedResponse, error) {
	return nil, nil, nil
}
func (n *noOpBRSClient) DownloadFilesFromRecovery(*backuprecoveryv1.DownloadFilesFromRecoveryOptions) (*core.DetailedResponse, error) {
	return nil, nil
}
func (n *noOpBRSClient) DownloadFilesFromRecoveryWithContext(context.Context, *backuprecoveryv1.DownloadFilesFromRecoveryOptions) (*core.DetailedResponse, error) {
	return nil, nil
}
func (n *noOpBRSClient) CancelRecoveryByID(*backuprecoveryv1.CancelRecoveryByIdOptions) (*core.DetailedResponse, error) {
	return nil, nil
}
func (n *noOpBRSClient) CancelRecoveryByIDWithContext(context.Context, *backuprecoveryv1.CancelRecoveryByIdOptions) (*core.DetailedResponse, error) {
	return nil, nil
}
func (n *noOpBRSClient) GetRestorePointsInTimeRange(*backuprecoveryv1.GetRestorePointsInTimeRangeOptions) (*backuprecoveryv1.GetRestorePointsInTimeRangeResponse, *core.DetailedResponse, error) {
	return nil, nil, nil
}
func (n *noOpBRSClient) GetRestorePointsInTimeRangeWithContext(context.Context, *backuprecoveryv1.GetRestorePointsInTimeRangeOptions) (*backuprecoveryv1.GetRestorePointsInTimeRangeResponse, *core.DetailedResponse, error) {
	return nil, nil, nil
}
func (n *noOpBRSClient) DownloadIndexedFile(*backuprecoveryv1.DownloadIndexedFileOptions) (*core.DetailedResponse, error) {
	return nil, nil
}
func (n *noOpBRSClient) DownloadIndexedFileWithContext(context.Context, *backuprecoveryv1.DownloadIndexedFileOptions) (*core.DetailedResponse, error) {
	return nil, nil
}
func (n *noOpBRSClient) SearchIndexedObjects(*backuprecoveryv1.SearchIndexedObjectsOptions) (*backuprecoveryv1.SearchIndexedObjectsResponse, *core.DetailedResponse, error) {
	return nil, nil, nil
}
func (n *noOpBRSClient) SearchIndexedObjectsWithContext(context.Context, *backuprecoveryv1.SearchIndexedObjectsOptions) (*backuprecoveryv1.SearchIndexedObjectsResponse, *core.DetailedResponse, error) {
	return nil, nil, nil
}
func (n *noOpBRSClient) SearchObjects(*backuprecoveryv1.SearchObjectsOptions) (*backuprecoveryv1.ObjectsSearchResponseBody, *core.DetailedResponse, error) {
	return nil, nil, nil
}
func (n *noOpBRSClient) SearchObjectsWithContext(context.Context, *backuprecoveryv1.SearchObjectsOptions) (*backuprecoveryv1.ObjectsSearchResponseBody, *core.DetailedResponse, error) {
	return nil, nil, nil
}
func (n *noOpBRSClient) SearchProtectedObjects(*backuprecoveryv1.SearchProtectedObjectsOptions) (*backuprecoveryv1.ProtectedObjectsSearchResponse, *core.DetailedResponse, error) {
	return nil, nil, nil
}
func (n *noOpBRSClient) SearchProtectedObjectsWithContext(context.Context, *backuprecoveryv1.SearchProtectedObjectsOptions) (*backuprecoveryv1.ProtectedObjectsSearchResponse, *core.DetailedResponse, error) {
	return nil, nil, nil
}
func (n *noOpBRSClient) GetSourceRegistrations(*backuprecoveryv1.GetSourceRegistrationsOptions) (*backuprecoveryv1.SourceRegistrations, *core.DetailedResponse, error) {
	return nil, nil, nil
}
func (n *noOpBRSClient) GetSourceRegistrationsWithContext(context.Context, *backuprecoveryv1.GetSourceRegistrationsOptions) (*backuprecoveryv1.SourceRegistrations, *core.DetailedResponse, error) {
	return nil, nil, nil
}
func (n *noOpBRSClient) RegisterProtectionSource(*backuprecoveryv1.RegisterProtectionSourceOptions) (*backuprecoveryv1.SourceRegistrationResponseParams, *core.DetailedResponse, error) {
	return nil, nil, nil
}
func (n *noOpBRSClient) RegisterProtectionSourceWithContext(context.Context, *backuprecoveryv1.RegisterProtectionSourceOptions) (*backuprecoveryv1.SourceRegistrationResponseParams, *core.DetailedResponse, error) {
	return nil, nil, nil
}
func (n *noOpBRSClient) GetProtectionSourceRegistration(*backuprecoveryv1.GetProtectionSourceRegistrationOptions) (*backuprecoveryv1.SourceRegistrationResponseParams, *core.DetailedResponse, error) {
	return nil, nil, nil
}
func (n *noOpBRSClient) GetProtectionSourceRegistrationWithContext(context.Context, *backuprecoveryv1.GetProtectionSourceRegistrationOptions) (*backuprecoveryv1.SourceRegistrationResponseParams, *core.DetailedResponse, error) {
	return nil, nil, nil
}
func (n *noOpBRSClient) UpdateProtectionSourceRegistration(*backuprecoveryv1.UpdateProtectionSourceRegistrationOptions) (*backuprecoveryv1.SourceRegistrationResponseParams, *core.DetailedResponse, error) {
	return nil, nil, nil
}
func (n *noOpBRSClient) UpdateProtectionSourceRegistrationWithContext(context.Context, *backuprecoveryv1.UpdateProtectionSourceRegistrationOptions) (*backuprecoveryv1.SourceRegistrationResponseParams, *core.DetailedResponse, error) {
	return nil, nil, nil
}
func (n *noOpBRSClient) PatchProtectionSourceRegistration(*backuprecoveryv1.PatchProtectionSourceRegistrationOptions) (*backuprecoveryv1.SourceRegistrationResponseParams, *core.DetailedResponse, error) {
	return nil, nil, nil
}
func (n *noOpBRSClient) PatchProtectionSourceRegistrationWithContext(context.Context, *backuprecoveryv1.PatchProtectionSourceRegistrationOptions) (*backuprecoveryv1.SourceRegistrationResponseParams, *core.DetailedResponse, error) {
	return nil, nil, nil
}
func (n *noOpBRSClient) DeleteProtectionSourceRegistration(*backuprecoveryv1.DeleteProtectionSourceRegistrationOptions) (*core.DetailedResponse, error) {
	return nil, nil
}
func (n *noOpBRSClient) DeleteProtectionSourceRegistrationWithContext(context.Context, *backuprecoveryv1.DeleteProtectionSourceRegistrationOptions) (*core.DetailedResponse, error) {
	return nil, nil
}
func (n *noOpBRSClient) RefreshProtectionSourceByID(*backuprecoveryv1.RefreshProtectionSourceByIdOptions) (*core.DetailedResponse, error) {
	return nil, nil
}
func (n *noOpBRSClient) RefreshProtectionSourceByIDWithContext(context.Context, *backuprecoveryv1.RefreshProtectionSourceByIdOptions) (*core.DetailedResponse, error) {
	return nil, nil
}
func (n *noOpBRSClient) GetProgressMonitors(*backuprecoveryv1.GetProgressMonitorsOptions) (*backuprecoveryv1.GetTasksResult, *core.DetailedResponse, error) {
	return nil, nil, nil
}
func (n *noOpBRSClient) GetProgressMonitorsWithContext(context.Context, *backuprecoveryv1.GetProgressMonitorsOptions) (*backuprecoveryv1.GetTasksResult, *core.DetailedResponse, error) {
	return nil, nil, nil
}
func (n *noOpBRSClient) GetProtectionRunProgress(*backuprecoveryv1.GetProtectionRunProgressOptions) (*backuprecoveryv1.GetProtectionRunProgressBody, *core.DetailedResponse, error) {
	return nil, nil, nil
}
func (n *noOpBRSClient) GetProtectionRunProgressWithContext(context.Context, *backuprecoveryv1.GetProtectionRunProgressOptions) (*backuprecoveryv1.GetProtectionRunProgressBody, *core.DetailedResponse, error) {
	return nil, nil, nil
}
func (n *noOpBRSClient) ConstructMetaInfo(*backuprecoveryv1.ConstructMetaInfoOptions) (*backuprecoveryv1.ConstructMetaInfoResult, *core.DetailedResponse, error) {
	return nil, nil, nil
}
func (n *noOpBRSClient) ConstructMetaInfoWithContext(context.Context, *backuprecoveryv1.ConstructMetaInfoOptions) (*backuprecoveryv1.ConstructMetaInfoResult, *core.DetailedResponse, error) {
	return nil, nil, nil
}

// ---------------------------------------------------------------------------
// helpers
// ---------------------------------------------------------------------------

func strPtr(s string) *string { return &s }

func makeMetadata(platformType, registryHost, namespace, repository, tag string) *backuprecoveryv1.ConnectorMetadata {
	return &backuprecoveryv1.ConnectorMetadata{
		K8sConnectorInfoList: []backuprecoveryv1.K8sConnectorInfo{
			{
				K8sPlatformType: strPtr(platformType),
				HelmChartOciRef: &backuprecoveryv1.OciArtifactReference{
					RegistryHost: strPtr(registryHost),
					Namespace:    strPtr(namespace),
					Repository:   strPtr(repository),
					Tag:          strPtr(tag),
					Digest:       strPtr("sha256:abc123"),
				},
			},
		},
	}
}

func baseConnector() *HelmConnector {
	return &HelmConnector{
		namespace:       "test-namespace",
		releaseName:     "brs-connector",
		chartRepo:       "oci://icr.io/ext/brs/brs-ds-connector-chart",
		ImagePullPolicy: "IfNotPresent",
		replicas:        1,
		config:          &HelmKubeConnectorConfig{Namespace: "test-namespace"},
		logger:          logger.NewNoop(),
	}
}

// ---------------------------------------------------------------------------
// TestSetDeployContext
// ---------------------------------------------------------------------------

func TestSetDeployContext(t *testing.T) {
	h := baseConnector()
	assert.Empty(t, h.deployCtx.PlatformType)
	assert.Nil(t, h.deployCtx.BRSClient)

	wrapper := &mockBRSWrapper{tenantID: "t1", client: &mockBRSClientForMetadata{}}
	h.SetDeployContext(ConnectorDeployContext{BRSClient: wrapper, PlatformType: "kRoksVpc"})

	assert.Equal(t, "kRoksVpc", h.deployCtx.PlatformType)
	assert.Equal(t, wrapper, h.deployCtx.BRSClient)
}

// ---------------------------------------------------------------------------
// TestResolveChartRef
// ---------------------------------------------------------------------------

func TestResolveChartRef(t *testing.T) {
	tests := []struct {
		name             string
		configChartRef   string // config.ChartReference — user-provided or ""
		configChartVer   string // config.ChartVersion   — user-provided or ""
		platformType     string
		metadataResult   *backuprecoveryv1.ConnectorMetadata
		metadataErr      error
		noBRSClient      bool
		wantChartRepo    string
		wantChartVersion string
	}{
		{
			name:           "BothUserProvided_SkipsAPICall",
			configChartRef: "oci://my-reg/my-chart",
			configChartVer: "1.2.3",
			platformType:   "kRoksVpc",
			// metadataResult left nil — if the API were called it would be a nil deref.
			// When user provides chartRef, h.chartRepo is preserved as the user-provided value.
			// Note: wantChartVersion is "" because resolveChartRef modifies h.chartVersion in-place
			// only when resolving from metadata; in actual deployment, h.config.ChartVersion is used
			// by NewHelmConnector/Deploy, while baseConnector() initialized h.chartVersion to "".
			wantChartRepo:    "oci://my-reg/my-chart",
			wantChartVersion: "",
		},
		{
			name:             "NoBRSClient_SkipsAPICall",
			configChartRef:   "",
			configChartVer:   "",
			platformType:     "kRoksVpc",
			noBRSClient:      true,
			wantChartRepo:    "oci://icr.io/ext/brs/brs-ds-connector-chart", // fallback unchanged
			wantChartVersion: "",
		},
		{
			name:             "EmptyPlatformType_SkipsAPICall",
			configChartRef:   "",
			configChartVer:   "",
			platformType:     "",
			metadataResult:   makeMetadata("kRoksVpc", "icr.io/", "ext", "brs/brs-ds-connector-chart", "7.3.0"),
			wantChartRepo:    "oci://icr.io/ext/brs/brs-ds-connector-chart",
			wantChartVersion: "",
		},
		{
			name:             "APIError_FallsBackToStatic",
			configChartRef:   "",
			configChartVer:   "",
			platformType:     "kRoksVpc",
			metadataErr:      fmt.Errorf("BRS unavailable"),
			wantChartRepo:    "oci://icr.io/ext/brs/brs-ds-connector-chart",
			wantChartVersion: "",
		},
		{
			name:             "NoMatchingPlatformType_FallsBackToStatic",
			configChartRef:   "",
			configChartVer:   "",
			platformType:     "kRoksVpc",
			metadataResult:   makeMetadata("kIksClassic", "icr.io/", "ext", "brs/brs-ds-connector-chart", "7.3.0"),
			wantChartRepo:    "oci://icr.io/ext/brs/brs-ds-connector-chart",
			wantChartVersion: "",
		},
		{
			name:             "BothEmpty_ResolvesFromMetadata",
			configChartRef:   "",
			configChartVer:   "",
			platformType:     "kRoksVpc",
			metadataResult:   makeMetadata("kRoksVpc", "icr.io/", "ext", "brs/brs-ds-connector-chart", "7.3.12"),
			wantChartRepo:    "oci://icr.io/ext/brs/brs-ds-connector-chart",
			wantChartVersion: "7.3.12",
		},
		{
			name:             "UserProvidedRef_OnlyVersionResolved",
			configChartRef:   "oci://my-reg/my-chart",
			configChartVer:   "",
			platformType:     "kRoksVpc",
			metadataResult:   makeMetadata("kRoksVpc", "icr.io/", "ext", "brs/brs-ds-connector-chart", "7.3.12"),
			wantChartRepo:    "oci://my-reg/my-chart",
			wantChartVersion: "7.3.12",
		},
		{
			name:             "UserProvidedVersion_OnlyRepoResolved",
			configChartRef:   "",
			configChartVer:   "1.0.0",
			platformType:     "kRoksVpc",
			metadataResult:   makeMetadata("kRoksVpc", "icr.io/", "ext", "brs/brs-ds-connector-chart", "7.3.12"),
			wantChartRepo:    "oci://icr.io/ext/brs/brs-ds-connector-chart",
			wantChartVersion: "", // h.chartVersion stays as "" (config.ChartVersion check blocks overwrite)
		},
		{
			name:             "AllFourPlatformTypes_kIksClassic",
			configChartRef:   "",
			configChartVer:   "",
			platformType:     "kIksClassic",
			metadataResult:   makeMetadata("kIksClassic", "icr.io/", "ext", "brs/brs-ds-connector-chart", "7.3.12"),
			wantChartRepo:    "oci://icr.io/ext/brs/brs-ds-connector-chart",
			wantChartVersion: "7.3.12",
		},
		{
			name:             "AllFourPlatformTypes_kIksVpc",
			configChartRef:   "",
			configChartVer:   "",
			platformType:     "kIksVpc",
			metadataResult:   makeMetadata("kIksVpc", "icr.io/", "ext", "brs/brs-ds-connector-chart", "7.3.12"),
			wantChartRepo:    "oci://icr.io/ext/brs/brs-ds-connector-chart",
			wantChartVersion: "7.3.12",
		},
		{
			name:             "AllFourPlatformTypes_kRoksClassic",
			configChartRef:   "",
			configChartVer:   "",
			platformType:     "kRoksClassic",
			metadataResult:   makeMetadata("kRoksClassic", "icr.io/", "ext", "brs/brs-ds-connector-chart", "7.3.12"),
			wantChartRepo:    "oci://icr.io/ext/brs/brs-ds-connector-chart",
			wantChartVersion: "7.3.12",
		},
		{
			name:           "MetadataEmptyList_FallsBackToStatic",
			configChartRef: "",
			configChartVer: "",
			platformType:   "kRoksVpc",
			metadataResult: &backuprecoveryv1.ConnectorMetadata{
				K8sConnectorInfoList: []backuprecoveryv1.K8sConnectorInfo{},
			},
			wantChartRepo:    "oci://icr.io/ext/brs/brs-ds-connector-chart",
			wantChartVersion: "",
		},
		{
			name:           "NilHelmChartOciRef_FallsBackToStatic",
			configChartRef: "",
			configChartVer: "",
			platformType:   "kRoksVpc",
			metadataResult: &backuprecoveryv1.ConnectorMetadata{
				K8sConnectorInfoList: []backuprecoveryv1.K8sConnectorInfo{
					{K8sPlatformType: strPtr("kRoksVpc"), HelmChartOciRef: nil},
				},
			},
			wantChartRepo:    "oci://icr.io/ext/brs/brs-ds-connector-chart",
			wantChartVersion: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := baseConnector()
			if tt.configChartRef != "" {
				h.chartRepo = tt.configChartRef
			}
			h.config.ChartReference = tt.configChartRef
			h.config.ChartVersion = tt.configChartVer

			if !tt.noBRSClient {
				h.deployCtx = ConnectorDeployContext{
					BRSClient: &mockBRSWrapper{
						tenantID: "test-tenant",
						client:   &mockBRSClientForMetadata{result: tt.metadataResult, err: tt.metadataErr},
					},
					PlatformType: tt.platformType,
				}
			}

			h.resolveChartRef(context.Background())

			assert.Equal(t, tt.wantChartRepo, h.chartRepo, "chartRepo")
			assert.Equal(t, tt.wantChartVersion, h.chartVersion, "chartVersion")
		})
	}
}

// ---------------------------------------------------------------------------
// TestDeploy_ResolvesChartFromMetadata
// Full Deploy pipeline: metadata resolves both repo + version before Install.
// ---------------------------------------------------------------------------

func TestDeploy_ResolvesChartFromMetadata(t *testing.T) {
	fakeClientSet := fake.NewSimpleClientset()
	_, _ = fakeClientSet.CoreV1().Namespaces().Create(
		context.Background(),
		&corev1.Namespace{ObjectMeta: metav1.ObjectMeta{Name: "test-namespace"}},
		metav1.CreateOptions{},
	)

	resolvedRef := "oci://icr.io/ext/brs/brs-ds-connector-chart"
	resolvedTag := "7.3.12-release-20260609"

	mockHelmClient := new(MockHelmClient)
	mockHelmClient.On("Init", mock.Anything, "test-namespace").Return(nil)
	mockHelmClient.On("Install", mock.Anything, mock.MatchedBy(func(cfg *HelmInstallConfig) bool {
		return cfg.ChartRef == resolvedRef && cfg.Version == resolvedTag
	})).Return(&releasev1.Release{Name: "brs-connector", Namespace: "test-namespace"}, nil)

	connector := &HelmConnector{
		namespace:       "test-namespace",
		releaseName:     "brs-connector",
		chartRepo:       "oci://icr.io/ext/brs/brs-ds-connector-chart", // default
		ImagePullPolicy: "IfNotPresent",
		replicas:        1,
		clientSet:       fakeClientSet,
		restConfig:      &rest.Config{},
		config: &HelmKubeConnectorConfig{
			Namespace:      "test-namespace",
			ClusterName:    "test-cluster",
			ChartReference: "", // empty → resolve from metadata
			ChartVersion:   "", // empty → resolve from metadata
			AuthConfig:     &KubernetesAuthConfig{AuthMethod: AuthMethodAPIKey, IamURL: "https://iam.cloud.ibm.com", ApiKey: "test-api-key"},
		},
		helmClient: mockHelmClient,
		logger:     logger.NewNoop(),
		deployCtx: ConnectorDeployContext{
			BRSClient: &mockBRSWrapper{
				tenantID: "test-tenant",
				client: &mockBRSClientForMetadata{
					result: makeMetadata("kRoksVpc", "icr.io/", "ext", "brs/brs-ds-connector-chart", resolvedTag),
				},
			},
			PlatformType: "kRoksVpc",
		},
	}

	result, err := connector.Deploy(context.Background(), "test-token")
	assert.NoError(t, err)
	assert.NotNil(t, result)
	mockHelmClient.AssertExpectations(t)
}

// ---------------------------------------------------------------------------
// TestDeploy_UserValuesNotOverriddenByMetadata
// When user provides both ChartReference and ChartVersion, metadata is skipped.
// ---------------------------------------------------------------------------

func TestDeploy_UserValuesNotOverriddenByMetadata(t *testing.T) {
	fakeClientSet := fake.NewSimpleClientset()
	_, _ = fakeClientSet.CoreV1().Namespaces().Create(
		context.Background(),
		&corev1.Namespace{ObjectMeta: metav1.ObjectMeta{Name: "test-namespace"}},
		metav1.CreateOptions{},
	)

	userRef := "oci://my-registry/my-chart"
	userTag := "my-custom-version"

	mockHelmClient := new(MockHelmClient)
	mockHelmClient.On("Init", mock.Anything, "test-namespace").Return(nil)
	mockHelmClient.On("Install", mock.Anything, mock.MatchedBy(func(cfg *HelmInstallConfig) bool {
		return cfg.ChartRef == userRef && cfg.Version == userTag
	})).Return(&releasev1.Release{Name: "brs-connector", Namespace: "test-namespace"}, nil)

	connector := &HelmConnector{
		namespace:       "test-namespace",
		releaseName:     "brs-connector",
		chartRepo:       userRef,
		chartVersion:    userTag,
		ImagePullPolicy: "IfNotPresent",
		replicas:        1,
		clientSet:       fakeClientSet,
		restConfig:      &rest.Config{},
		config: &HelmKubeConnectorConfig{
			Namespace:      "test-namespace",
			ClusterName:    "test-cluster",
			ChartReference: userRef, // user-provided → metadata must not overwrite
			ChartVersion:   userTag, // user-provided → metadata must not overwrite
			AuthConfig:     &KubernetesAuthConfig{AuthMethod: AuthMethodAPIKey, IamURL: "https://iam.cloud.ibm.com", ApiKey: "test-api-key"},
		},
		helmClient: mockHelmClient,
		logger:     logger.NewNoop(),
		// BRSClient present but API result would give different values — must not win
		deployCtx: ConnectorDeployContext{
			BRSClient: &mockBRSWrapper{
				tenantID: "test-tenant",
				client: &mockBRSClientForMetadata{
					result: makeMetadata("kRoksVpc", "icr.io/", "ext", "brs/brs-ds-connector-chart", "999.0.0"),
				},
			},
			PlatformType: "kRoksVpc",
		},
	}

	result, err := connector.Deploy(context.Background(), "test-token")
	assert.NoError(t, err)
	assert.NotNil(t, result)
	mockHelmClient.AssertExpectations(t)
}

// ---------------------------------------------------------------------------
// TestDeploy_MetadataAPIError_ContinuesWithStatic
// API error is non-fatal — deploy continues with static chart ref.
// ---------------------------------------------------------------------------

func TestDeploy_MetadataAPIError_ContinuesWithStatic(t *testing.T) {
	fakeClientSet := fake.NewSimpleClientset()
	_, _ = fakeClientSet.CoreV1().Namespaces().Create(
		context.Background(),
		&corev1.Namespace{ObjectMeta: metav1.ObjectMeta{Name: "test-namespace"}},
		metav1.CreateOptions{},
	)

	staticRef := "oci://icr.io/ext/brs/brs-ds-connector-chart"

	mockHelmClient := new(MockHelmClient)
	mockHelmClient.On("Init", mock.Anything, "test-namespace").Return(nil)
	mockHelmClient.On("Install", mock.Anything, mock.MatchedBy(func(cfg *HelmInstallConfig) bool {
		return cfg.ChartRef == staticRef && cfg.Version == ""
	})).Return(&releasev1.Release{Name: "brs-connector", Namespace: "test-namespace"}, nil)

	connector := &HelmConnector{
		namespace:       "test-namespace",
		releaseName:     "brs-connector",
		chartRepo:       staticRef,
		chartVersion:    "",
		ImagePullPolicy: "IfNotPresent",
		replicas:        1,
		clientSet:       fakeClientSet,
		restConfig:      &rest.Config{},
		config: &HelmKubeConnectorConfig{
			Namespace:      "test-namespace",
			ClusterName:    "test-cluster",
			ChartReference: "",
			ChartVersion:   "",
			AuthConfig:     &KubernetesAuthConfig{AuthMethod: AuthMethodAPIKey, IamURL: "https://iam.cloud.ibm.com", ApiKey: "test-api-key"},
		},
		helmClient: mockHelmClient,
		logger:     logger.NewNoop(),
		deployCtx: ConnectorDeployContext{
			BRSClient: &mockBRSWrapper{
				tenantID: "test-tenant",
				client:   &mockBRSClientForMetadata{err: fmt.Errorf("BRS unavailable")},
			},
			PlatformType: "kRoksVpc",
		},
	}

	result, err := connector.Deploy(context.Background(), "test-token")
	assert.NoError(t, err, "metadata API error must not fail Deploy")
	assert.NotNil(t, result)
	mockHelmClient.AssertExpectations(t)
}

// ---------------------------------------------------------------------------
// TestDeploy_NoBRSClient_ContinuesWithStatic
// No BRS client set — deploy proceeds with statically configured values.
// ---------------------------------------------------------------------------

func TestDeploy_NoBRSClient_ContinuesWithStatic(t *testing.T) {
	fakeClientSet := fake.NewSimpleClientset()
	_, _ = fakeClientSet.CoreV1().Namespaces().Create(
		context.Background(),
		&corev1.Namespace{ObjectMeta: metav1.ObjectMeta{Name: "test-namespace"}},
		metav1.CreateOptions{},
	)

	staticRef := "oci://icr.io/ext/brs/brs-ds-connector-chart"

	mockHelmClient := new(MockHelmClient)
	mockHelmClient.On("Init", mock.Anything, "test-namespace").Return(nil)
	mockHelmClient.On("Install", mock.Anything, mock.MatchedBy(func(cfg *HelmInstallConfig) bool {
		return cfg.ChartRef == staticRef
	})).Return(&releasev1.Release{Name: "brs-connector", Namespace: "test-namespace"}, nil)

	connector := &HelmConnector{
		namespace:       "test-namespace",
		releaseName:     "brs-connector",
		chartRepo:       staticRef,
		ImagePullPolicy: "IfNotPresent",
		replicas:        1,
		clientSet:       fakeClientSet,
		restConfig:      &rest.Config{},
		config: &HelmKubeConnectorConfig{
			Namespace:   "test-namespace",
			ClusterName: "test-cluster",
			AuthConfig:  &KubernetesAuthConfig{AuthMethod: AuthMethodAPIKey, IamURL: "https://iam.cloud.ibm.com", ApiKey: "test-api-key"},
		},
		helmClient: mockHelmClient,
		logger:     logger.NewNoop(),
		// deployCtx intentionally zero — no BRS client injected
	}

	result, err := connector.Deploy(context.Background(), "test-token")
	assert.NoError(t, err)
	assert.NotNil(t, result)
	mockHelmClient.AssertExpectations(t)
}
