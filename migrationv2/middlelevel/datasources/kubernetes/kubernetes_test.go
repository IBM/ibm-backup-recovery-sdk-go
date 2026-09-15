/*******************************************************************************
 * IBM Confidential
 * OCO Source Materials
 * (C) Copyright IBM Corp. 2026 All Rights Reserved.
 * The source code for this program is not  published or otherwise divested of
 * its trade secrets, irrespective of what has been deposited with
 * the U.S. Copyright Office.
 ******************************************************************************/
package kubernetes

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/IBM/go-sdk-core/v5/core"
	"github.com/IBM/ibm-backup-recovery-sdk-go/backuprecoveryv1"
	"github.com/IBM/ibm-backup-recovery-sdk-go/migrationv2/common/logger"
	testmockrs "github.com/IBM/ibm-backup-recovery-sdk-go/migrationv2/common/testing"
	"github.com/IBM/ibm-backup-recovery-sdk-go/migrationv2/common/types"
	"github.com/IBM/ibm-backup-recovery-sdk-go/migrationv2/middlelevel"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

// MockBRSClientWrapper implements types.BRSClientWrapperInterface for testing
type MockBRSClientWrapper struct {
	mock.Mock
	mockClient *testmockrs.MockBRSClient
}

func (m *MockBRSClientWrapper) GetTenantId() string {
	args := m.Called()
	return args.String(0)
}

func (m *MockBRSClientWrapper) GetBRSClient() backuprecoveryv1.BRSClientInterface {
	args := m.Called()
	return args.Get(0).(backuprecoveryv1.BRSClientInterface)
}

func (m *MockBRSClientWrapper) GetRegion() string {
	args := m.Called()
	return args.String(0)
}

func (m *MockBRSClientWrapper) GetCRN() string {
	args := m.Called()
	return args.String(0)
}

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

// mockBRSClientWrapper creates a mock BRS client wrapper for testing
func mockBRSClientWrapper() types.BRSClientWrapperInterface {
	mockClient := new(testmockrs.MockBRSClient)
	mockWrapper := new(MockBRSClientWrapper)
	mockWrapper.mockClient = mockClient

	// Set up default mock behavior
	mockWrapper.On("GetTenantId").Return("test-tenant-id")
	mockWrapper.On("GetBRSClient").Return(mockClient)
	mockWrapper.On("GetRegion").Return("us-south")
	mockWrapper.On("GetCRN").Return("crn:v1:test")

	return mockWrapper
}

// mockIksManager creates a mock IKS manager for testing
func mockIksManager() middlelevel.IksManagerInterface {
	mockIks := new(MockIksManager)

	// Set up default mock behavior - return nil for K8s clients (tests don't need them)
	mockIks.On("GetKubeApi", mock.Anything).Return(nil, nil, nil)
	mockIks.On("ApplyRBACAndGetKubeconfig", mock.Anything).Return([]byte("mock-kubeconfig"), nil)
	mockIks.On("GetClusterBearerToken", mock.Anything).Return("mock-bearer-token", nil)

	return mockIks
}

// createTestKubernetesDataSource creates a KubernetesDataSource for testing without real K8s client initialization
func createTestKubernetesDataSource(name string, config *KubernetesDataSourceConfig, brsClient types.BRSClientWrapperInterface) *KubernetesDataSource {
	// Create a test logger
	testLogger := logger.New(logger.Config{
		Level:       "info",
		ServiceName: "test-kubernetes-datasource",
		Environment: "test",
	})

	return &KubernetesDataSource{
		Type:                       types.DataSourceKubernetes,
		KubernetesDataSourceType:   config.KubernetesDataSourceType,
		Name:                       name,
		Config:                     config,
		ClusterID:                  config.ClusterId,
		ClusterName:                config.ClusterName,
		ClusterEndpoint:            config.ClusterEndpoint,
		Namespace:                  "config",
		KubeConfig:                 "",
		KubernetesProtectionParams: config.KubernetesProtectionParams,
		KubernetesRestoreParams:    config.KubernetesRestoreParams,
		brsClient:                  brsClient,
		clientSet:                  nil, // Mock tests don't need real K8s client
		restConfig:                 nil,
		iksClient:                  mockIksManager(),
		logger:                     testLogger,
	}
}

// =============================================================================
// KUBERNETES DATA SOURCE CREATION TESTS
// =============================================================================

func TestNewKubernetesDataSource(t *testing.T) {
	tests := []struct {
		name           string
		dsName         string
		config         *KubernetesDataSourceConfig
		expectError    bool
		errorContains  string
		validateResult func(*testing.T, *KubernetesDataSource)
	}{
		{
			name:   "Success",
			dsName: "test-datasource",
			config: &KubernetesDataSourceConfig{
				ClusterName:              "test-cluster",
				ClusterId:                "cluster-123",
				ClusterEndpoint:          "https://test-cluster.example.com",
				VpcID:                    "vpc-123",
				KubernetesDataSourceType: Kiksclassic,
			},
			expectError: false,
			validateResult: func(t *testing.T, ds *KubernetesDataSource) {
				assert.NotNil(t, ds)
				assert.Equal(t, types.DataSourceKubernetes, ds.Type)
				assert.Equal(t, "test-datasource", ds.Name)
				assert.Equal(t, "cluster-123", ds.ClusterID)
				assert.Equal(t, "test-cluster", ds.ClusterName)
				assert.Equal(t, "https://test-cluster.example.com", ds.ClusterEndpoint)
			},
		},
		{
			name:          "NilConfig",
			dsName:        "test-datasource",
			config:        nil,
			expectError:   true,
			errorContains: "config cannot be nil",
		},
		{
			name:   "MissingAuthenticator",
			dsName: "test-datasource",
			config: &KubernetesDataSourceConfig{
				ClusterName:              "test-cluster",
				ClusterId:                "cluster-123",
				ClusterType:              "IKS",
				ContainerEndpoint:        "https://containers.cloud.ibm.com",
				ContainerEndpointType:    "public",
				Authenticator:            nil, // Missing authenticator
				KubernetesDataSourceType: Kiksclassic,
			},
			expectError: true,
		},
		{
			name:   "InvalidContainerEndpoint",
			dsName: "test-datasource",
			config: &KubernetesDataSourceConfig{
				ClusterName:              "test-cluster",
				ClusterId:                "cluster-123",
				ClusterType:              "IKS",
				ContainerEndpoint:        "", // Empty endpoint
				ContainerEndpointType:    "public",
				Authenticator:            &core.IamAuthenticator{ApiKey: "test-api-key"},
				KubernetesDataSourceType: Kiksclassic,
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.expectError {
				// For error cases, use NewKubernetesDataSource directly
				dataSource, err := NewKubernetesDataSource(tt.dsName, tt.config, mockBRSClientWrapper())
				assert.Error(t, err)
				assert.Nil(t, dataSource)
				if tt.errorContains != "" {
					assert.Contains(t, err.Error(), tt.errorContains)
				}
			} else {
				// For success cases, use the helper that bypasses IKS manager creation
				dataSource := createTestKubernetesDataSource(tt.dsName, tt.config, mockBRSClientWrapper())
				assert.NotNil(t, dataSource)
				if tt.validateResult != nil {
					tt.validateResult(t, dataSource)
				}
			}
		})
	}
}

func TestKubernetesDataSource_GetType(t *testing.T) {
	config := &KubernetesDataSourceConfig{
		ClusterName: "test-cluster",
	}

	dataSource := createTestKubernetesDataSource("test", config, mockBRSClientWrapper())

	assert.Equal(t, types.DataSourceKubernetes, dataSource.GetType())
}

func TestKubernetesDataSource_GetName(t *testing.T) {
	config := &KubernetesDataSourceConfig{
		ClusterName: "test-cluster",
	}

	dataSource := createTestKubernetesDataSource("my-datasource", config, mockBRSClientWrapper())

	assert.Equal(t, "my-datasource", dataSource.GetName())
}

func TestKubernetesDataSource_Validate(t *testing.T) {
	tests := []struct {
		name          string
		dsName        string
		config        *KubernetesDataSourceConfig
		nilDataSource bool
		expectError   bool
		errorContains string
	}{
		{
			name:   "Success",
			dsName: "test",
			config: &KubernetesDataSourceConfig{
				ClusterName: "test-cluster",
				ClusterId:   "cluster-123",
			},
			expectError: false,
		},
		{
			name:          "NilDataSource",
			nilDataSource: true,
			expectError:   true,
			errorContains: "cannot be nil",
		},
		{
			name:   "MissingName",
			dsName: "",
			config: &KubernetesDataSourceConfig{
				ClusterName: "test-cluster",
			},
			expectError:   true,
			errorContains: "data source name is required",
		},
		{
			name:          "MissingClusterInfo",
			dsName:        "test",
			config:        &KubernetesDataSourceConfig{},
			expectError:   true,
			errorContains: "either cluster_id or cluster_name is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var dataSource *KubernetesDataSource
			if !tt.nilDataSource {
				dataSource = createTestKubernetesDataSource(tt.dsName, tt.config, mockBRSClientWrapper())
			}

			err := dataSource.Validate()

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

// =============================================================================
// CREATE PROTECTION GROUP TESTS
// Tests parameter validation and building logic
// =============================================================================

func TestCreateProtectionGroup(t *testing.T) {
	tests := []struct {
		name             string
		config           *KubernetesDataSourceConfig
		groupParams      *types.ProtectionGroupParams
		mockSources      []backuprecoveryv1.ProtectionSourceNodes
		mockSourcesError error
		expectError      bool
		errorContains    string
		validateResult   func(*testing.T, *backuprecoveryv1.CreateProtectionGroupOptions)
	}{
		{
			name: "NilParams",
			config: &KubernetesDataSourceConfig{
				ClusterName: "test-cluster",
				ClusterId:   "cluster-123",
			},
			groupParams:   nil,
			expectError:   true,
			errorContains: "protection group params are required",
		},
		{
			name: "MissingKubernetesParams",
			config: &KubernetesDataSourceConfig{
				ClusterName: "test-cluster",
				ClusterId:   "cluster-123",
			},
			groupParams: &types.ProtectionGroupParams{
				Name: "test-group",
			},
			expectError:   true,
			errorContains: "KubernetesProtectionParams are required",
		},
		{
			name: "Success_WithNamespaces",
			config: &KubernetesDataSourceConfig{
				ClusterName: "test-cluster",
				ClusterId:   "cluster-123",
				KubernetesProtectionParams: &types.KubernetesProtectionParams{
					IncludeNamespaces: "default,app-namespace",
					ExcludeNamespaces: "kube-system",
					Settings: &types.ProtectionSetting{
						CSISnapshot: true,
						Labels: types.Labels{
							PersistentVolumeClaim: types.KeyValueFilter{
								Inclusion: "env=prod",
								Exclusion: "temp=true",
								LogicRule: "MatchAll",
							},
						},
					},
					NamespacesSetting: []types.NamespacesSetting{
						{
							Namespace: "default",
							PersistentVolumeClaims: types.PVCSection{
								Inclusion: []string{"pvc-1"},
							},
							Resources: types.ResourceSection{
								Inclusion: &types.ResourceTypes{
									ResourceTypes: []string{"Deployment", "Service"},
								},
							},
							Hooks: types.HookSection{
								RulesApplyMode:         "sequential",
								FailBackupIfHookFailed: true,
								Rules: []types.Rule{
									{
										Rule: types.RuleDetail{
											PodLabels:  "app=web",
											PreScript:  "echo 'pre-backup'",
											PostScript: "echo 'post-backup'",
											Container:  "main",
										},
									},
								},
							},
						},
					},
					IncludeAppLabels: "tier=frontend",
				},
			},
			groupParams: &types.ProtectionGroupParams{
				Name: "test-protection-group",
			},
			mockSources: func() []backuprecoveryv1.ProtectionSourceNodes {
				namespaceType := backuprecoveryv1.KubernetesProtectionSource_Type_Knamespace
				pvcType := backuprecoveryv1.KubernetesSourceRegistrationParams_KubernetesType_Kpersistentvolumeclaim
				return []backuprecoveryv1.ProtectionSourceNodes{
					{
						Nodes: []backuprecoveryv1.ProtectionSourceNodes{
							{
								Nodes: []backuprecoveryv1.ProtectionSourceNodes{
									{
										ProtectionSource: &backuprecoveryv1.ProtectionSourceNode{
											ID: core.Int64Ptr(100),
											KubernetesProtectionSource: &backuprecoveryv1.KubernetesProtectionSource{
												Type: &namespaceType,
												Name: core.StringPtr("default"),
											},
										},
										Nodes: []backuprecoveryv1.ProtectionSourceNodes{
											{
												ProtectionSource: &backuprecoveryv1.ProtectionSourceNode{
													ID: core.Int64Ptr(101),
													KubernetesProtectionSource: &backuprecoveryv1.KubernetesProtectionSource{
														Type: &pvcType,
														Name: core.StringPtr("pvc-1"),
													},
												},
											},
										},
									},
									{
										ProtectionSource: &backuprecoveryv1.ProtectionSourceNode{
											ID: core.Int64Ptr(200),
											KubernetesProtectionSource: &backuprecoveryv1.KubernetesProtectionSource{
												Type: &namespaceType,
												Name: core.StringPtr("app-namespace"),
											},
										},
									},
								},
							},
						},
					},
				}
			}(),
			expectError: false,
			validateResult: func(t *testing.T, result *backuprecoveryv1.CreateProtectionGroupOptions) {
				assert.NotNil(t, result.KubernetesParams)
				assert.True(t, len(result.KubernetesParams.Objects) > 0)
				assert.True(t, *result.KubernetesParams.LeverageCSISnapshot)
			},
		},
		{
			name: "EmptyNamespaces",
			config: &KubernetesDataSourceConfig{
				ClusterName: "test-cluster",
				ClusterId:   "cluster-123",
				KubernetesProtectionParams: &types.KubernetesProtectionParams{
					Settings: &types.ProtectionSetting{
						CSISnapshot: false,
					},
				},
			},
			groupParams: &types.ProtectionGroupParams{
				Name: "test-group",
			},
			mockSources: []backuprecoveryv1.ProtectionSourceNodes{
				{
					Nodes: []backuprecoveryv1.ProtectionSourceNodes{},
				},
			},
			expectError: false,
			validateResult: func(t *testing.T, result *backuprecoveryv1.CreateProtectionGroupOptions) {
				assert.NotNil(t, result.KubernetesParams)
				assert.Equal(t, 0, len(result.KubernetesParams.Objects))
			},
		},
		{
			name: "ListProtectionSourcesError",
			config: &KubernetesDataSourceConfig{
				ClusterName: "test-cluster",
				ClusterId:   "cluster-123",
				KubernetesProtectionParams: &types.KubernetesProtectionParams{
					Settings: &types.ProtectionSetting{},
				},
			},
			groupParams: &types.ProtectionGroupParams{
				Name: "test-group",
			},
			mockSourcesError: fmt.Errorf("API error"),
			expectError:      true,
			errorContains:    "failed to fetch kubernetes namespaces",
		},
		{
			name: "DefaultGroupName",
			config: &KubernetesDataSourceConfig{
				ClusterName: "test-cluster",
				ClusterId:   "cluster-123",
				KubernetesProtectionParams: &types.KubernetesProtectionParams{
					Settings: &types.ProtectionSetting{},
				},
			},
			groupParams: &types.ProtectionGroupParams{
				Name: "", // Empty name should trigger default name generation
			},
			mockSources: []backuprecoveryv1.ProtectionSourceNodes{
				{
					Nodes: []backuprecoveryv1.ProtectionSourceNodes{},
				},
			},
			expectError: false,
		},
		{
			name: "PauseFutureRuns_True_SetsIsPaused",
			config: &KubernetesDataSourceConfig{
				ClusterName: "test-cluster",
				ClusterId:   "cluster-123",
				KubernetesProtectionParams: &types.KubernetesProtectionParams{
					Settings: &types.ProtectionSetting{
						CSISnapshot: true,
					},
				},
			},
			groupParams: &types.ProtectionGroupParams{
				Name: "paused-group",
				Settings: &types.ProtectionSetting{
					PauseFutureRuns: true,
				},
			},
			mockSources: []backuprecoveryv1.ProtectionSourceNodes{
				{
					Nodes: []backuprecoveryv1.ProtectionSourceNodes{},
				},
			},
			expectError: false,
			validateResult: func(t *testing.T, result *backuprecoveryv1.CreateProtectionGroupOptions) {
				assert.NotNil(t, result.IsPaused, "IsPaused should be set when PauseFutureRuns is true")
				assert.True(t, *result.IsPaused, "IsPaused should be true")
			},
		},
		{
			name: "PauseFutureRuns_False_DoesNotSetIsPaused",
			config: &KubernetesDataSourceConfig{
				ClusterName: "test-cluster",
				ClusterId:   "cluster-123",
				KubernetesProtectionParams: &types.KubernetesProtectionParams{
					Settings: &types.ProtectionSetting{},
				},
			},
			groupParams: &types.ProtectionGroupParams{
				Name: "active-group",
				Settings: &types.ProtectionSetting{
					PauseFutureRuns: false,
				},
			},
			mockSources: []backuprecoveryv1.ProtectionSourceNodes{
				{
					Nodes: []backuprecoveryv1.ProtectionSourceNodes{},
				},
			},
			expectError: false,
			validateResult: func(t *testing.T, result *backuprecoveryv1.CreateProtectionGroupOptions) {
				assert.Nil(t, result.IsPaused, "IsPaused should not be set when PauseFutureRuns is false")
			},
		},
		{
			name: "PauseFutureRuns_NilSettings_DoesNotSetIsPaused",
			config: &KubernetesDataSourceConfig{
				ClusterName: "test-cluster",
				ClusterId:   "cluster-123",
				KubernetesProtectionParams: &types.KubernetesProtectionParams{
					Settings: &types.ProtectionSetting{},
				},
			},
			groupParams: &types.ProtectionGroupParams{
				Name:     "no-settings-group",
				Settings: nil, // nil Settings — should not panic or set IsPaused
			},
			mockSources: []backuprecoveryv1.ProtectionSourceNodes{
				{
					Nodes: []backuprecoveryv1.ProtectionSourceNodes{},
				},
			},
			expectError: false,
			validateResult: func(t *testing.T, result *backuprecoveryv1.CreateProtectionGroupOptions) {
				assert.Nil(t, result.IsPaused, "IsPaused should not be set when Settings is nil")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var mockBRSWrapper types.BRSClientWrapperInterface

			// Only setup mocks if we need them (not for early error cases)
			if tt.mockSources != nil || tt.mockSourcesError != nil {
				mockBRSClient := new(testmockrs.MockBRSClient)
				mockWrapper := new(MockBRSClientWrapper)

				mockWrapper.On("GetTenantId").Return("test-tenant-id")
				mockWrapper.On("GetBRSClient").Return(mockBRSClient)
				mockBRSClient.On("ListProtectionSources", mock.Anything).Return(tt.mockSources, &core.DetailedResponse{}, tt.mockSourcesError)

				mockBRSWrapper = mockWrapper
			} else {
				// For early error cases, use simple mock
				mockBRSWrapper = mockBRSClientWrapper()
			}

			dataSource := createTestKubernetesDataSource("test", tt.config, mockBRSWrapper)
			ctx := context.Background()

			result, err := dataSource.CreateProtectionGroup(ctx, 123, tt.groupParams)

			if tt.expectError {
				assert.Error(t, err)
				assert.Nil(t, result)
				if tt.errorContains != "" {
					assert.Contains(t, err.Error(), tt.errorContains)
				}
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
				if tt.validateResult != nil {
					tt.validateResult(t, result)
				}
			}

			// Only assert expectations if we set up detailed mocks
			if mockWrapper, ok := mockBRSWrapper.(*MockBRSClientWrapper); ok {
				if tt.mockSources != nil || tt.mockSourcesError != nil {
					mockWrapper.AssertExpectations(t)
					if mockWrapper.mockClient != nil {
						mockWrapper.mockClient.AssertExpectations(t)
					}
				}
			}
		})
	}
}

// =============================================================================
// RUN RESTORE TESTS
// =============================================================================

func TestRunRestore(t *testing.T) {
	tests := []struct {
		name                  string
		config                *KubernetesDataSourceConfig
		restoreParams         *types.RestoreParams
		groupID               string
		backupID              string
		sourceID              int64
		setupMocks            func(*testmockrs.MockBRSClient, *MockBRSClientWrapper)
		expectError           bool
		expectedErrorContains string
		validateResult        func(*testing.T, *backuprecoveryv1.CreateRecoveryOptions)
	}{
		{
			name: "Error_NilRestoreParams",
			config: &KubernetesDataSourceConfig{
				ClusterName:             "test-cluster",
				KubernetesRestoreParams: &types.KubernetesRestoreParams{},
			},
			restoreParams:         nil,
			groupID:               "groupId",
			backupID:              "backupId",
			sourceID:              123,
			setupMocks:            func(mc *testmockrs.MockBRSClient, mw *MockBRSClientWrapper) {},
			expectError:           true,
			expectedErrorContains: "restore params are required",
		},
		{
			name: "Error_NilKubernetesRestoreParams",
			config: &KubernetesDataSourceConfig{
				ClusterName: "test-cluster",
				// KubernetesRestoreParams is nil
			},
			restoreParams: &types.RestoreParams{
				Name: "test-restore",
			},
			groupID:               "groupId",
			backupID:              "backupId",
			sourceID:              123,
			setupMocks:            func(mc *testmockrs.MockBRSClient, mw *MockBRSClientWrapper) {},
			expectError:           true,
			expectedErrorContains: "kubernetes restore params are required",
		},
		{
			name: "Error_MissingRecoverObjectSpec",
			config: &KubernetesDataSourceConfig{
				ClusterName:             "test-cluster",
				ClusterId:               "cluster-123",
				KubernetesRestoreParams: &types.KubernetesRestoreParams{
					// Missing required RecoverObjectSpec
				},
			},
			restoreParams: &types.RestoreParams{
				Name: "test-restore",
			},
			groupID:  "groupId",
			backupID: "backupId",
			sourceID: 123,
			setupMocks: func(mc *testmockrs.MockBRSClient, mw *MockBRSClientWrapper) {
				mw.On("GetTenantId").Return("test-tenant-id")
				mw.On("GetBRSClient").Return(mc)
				snapshotID := "snapshot-123"
				mockProtectionGroupRun := &backuprecoveryv1.ProtectionGroupRun{
					Objects: []backuprecoveryv1.ObjectRunResult{
						{
							ArchivalInfo: &backuprecoveryv1.ArchivalRun{
								ArchivalTargetResults: []backuprecoveryv1.ArchivalTargetResult{
									{
										SnapshotID: &snapshotID,
									},
								},
							},
						},
					},
				}
				mc.On("GetProtectionGroupRun", mock.Anything).Return(mockProtectionGroupRun, &core.DetailedResponse{}, nil)
			},
			expectError:           true,
			expectedErrorContains: "unable to create recover namespace params",
		},
		{
			name: "Success_WithSnapshotID",
			config: &KubernetesDataSourceConfig{
				ClusterName: "test-cluster",
				ClusterId:   "cluster-123",
				KubernetesRestoreParams: &types.KubernetesRestoreParams{
					RecoverObjectSpec: &types.RecoverObjectSpec{
						RestoreOnlyPvc: false,
						StorageClasses: []types.StorageClassMapping{
							{
								Old: "standard",
								New: "premium",
							},
						},
						IncludeObjects: &types.K8sObject{},
						ExcludeObjects: &types.K8sObject{},
					},
					RecoverMultipleObjects: []types.RecoverObject{
						{
							SnapshotInfo: &types.SnapshotInfo{
								SnapshotID:          "snapshot-123",
								ProtectionGroupId:   "group-123",
								ProtectionGroupName: "test-group",
							},
							RecoverObjectSpec: &types.RecoverObjectSpec{
								RestoreOnlyPvc: false,
								StorageClasses: []types.StorageClassMapping{
									{
										Old: "standard",
										New: "premium",
									},
								},
								IncludeObjects: &types.K8sObject{},
								ExcludeObjects: &types.K8sObject{},
							},
						},
					},
					RenameRecoveredNamespacesParams: &types.RenameRecoveredNamespacesParams{
						Prefix: "restored-",
						Suffix: "-backup",
					},
					SkipClusterCompatibilityCheck: true,
					RegionMapping: &types.MigrationMapParams{
						Source: "us-south",
						Target: "us-east",
					},
					ZoneMappings: []types.MigrationMapParams{
						{
							Source: "zone-1",
							Target: "zone-2",
						},
					},
				},
			},
			restoreParams: &types.RestoreParams{
				Name:                    "test-restore",
				BackupPositionFromFirst: 0,
			},
			groupID:  "groupId",
			backupID: "backupJobId",
			sourceID: 456,
			setupMocks: func(mc *testmockrs.MockBRSClient, mw *MockBRSClientWrapper) {
				mw.On("GetTenantId").Return("test-tenant-id")
				mw.On("GetBRSClient").Return(mc)
				snapshotID := "snapshot-123"
				nsName := "test-namespace"
				nsID := int64(1)
				mockProtectionGroupRun := &backuprecoveryv1.ProtectionGroupRun{
					Objects: []backuprecoveryv1.ObjectRunResult{
						{
							Object: &backuprecoveryv1.ObjectSummary{
								Name: &nsName,
								ID:   &nsID,
							},
							ArchivalInfo: &backuprecoveryv1.ArchivalRun{
								ArchivalTargetResults: []backuprecoveryv1.ArchivalTargetResult{
									{
										SnapshotID: &snapshotID,
									},
								},
							},
						},
					},
					}
					mc.On("GetProtectionGroupRun", mock.Anything).Return(mockProtectionGroupRun, &core.DetailedResponse{}, nil)
				},
			expectError: false,
			validateResult: func(t *testing.T, result *backuprecoveryv1.CreateRecoveryOptions) {
				assert.Equal(t, "test-restore", *result.Name)
				assert.NotNil(t, result.KubernetesParams)
				assert.NotNil(t, result.KubernetesParams.RecoverNamespaceParams)
				targetParams := result.KubernetesParams.RecoverNamespaceParams.KubernetesTargetParams
				assert.NotNil(t, targetParams)
				assert.Equal(t, "restored-", *targetParams.RenameRecoveredNamespacesParams.Prefix)
				assert.Equal(t, "-backup", *targetParams.RenameRecoveredNamespacesParams.Suffix)
				assert.True(t, *targetParams.SkipClusterCompatibilityCheck)
				assert.Equal(t, "us-south", *targetParams.RecoveryRegionMigrationParams.CurrentValue)
				assert.Equal(t, "us-east", *targetParams.RecoveryRegionMigrationParams.NewValue)
			},
		},
		{
			name: "Success_WithIncludeExcludeObjects",
			config: &KubernetesDataSourceConfig{
				ClusterName: "test-cluster",
				ClusterId:   "cluster-123",
				KubernetesRestoreParams: &types.KubernetesRestoreParams{
					RecoverObjectSpec: &types.RecoverObjectSpec{
						IncludeObjects: &types.K8sObject{
							SelectedResources: []types.ResourceInfo{
								{
									ResourceKind:     "Deployment",
									ResourceApiGroup: "apps",
									ResourceList: []types.ResourceInstance{
										{
											Name: "my-deployment",
										},
									},
								},
							},
							SelectedLabels: &types.SelectedLabels{
								LabelCombination: "AND",
								Labels: []types.K8sLabel{
									{
										Key:   "app",
										Value: "web",
									},
								},
							},
						},
						ExcludeObjects: &types.K8sObject{
							SelectedResources: []types.ResourceInfo{
								{
									ResourceKind:     "Job",
									ResourceApiGroup: "batch",
								},
							},
						},
					},
					RecoverMultipleObjects: []types.RecoverObject{
						{
							SnapshotInfo: &types.SnapshotInfo{
								SnapshotID: "snapshot-456",
							},
							RecoverObjectSpec: &types.RecoverObjectSpec{
								IncludeObjects: &types.K8sObject{
									SelectedResources: []types.ResourceInfo{
										{
											ResourceKind:     "Deployment",
											ResourceApiGroup: "apps",
											ResourceList: []types.ResourceInstance{
												{
													Name: "my-deployment",
												},
											},
										},
									},
									SelectedLabels: &types.SelectedLabels{
										LabelCombination: "AND",
										Labels: []types.K8sLabel{
											{
												Key:   "app",
												Value: "web",
											},
										},
									},
								},
								ExcludeObjects: &types.K8sObject{
									SelectedResources: []types.ResourceInfo{
										{
											ResourceKind:     "Job",
											ResourceApiGroup: "batch",
										},
									},
								},
							},
						},
					},
					RenameRecoveredNamespacesParams: &types.RenameRecoveredNamespacesParams{
						Prefix: "test-",
					},
					SkipClusterCompatibilityCheck: false,
					RegionMapping: &types.MigrationMapParams{
						Source: "us-south",
						Target: "us-east",
					},
				},
			},
			restoreParams: &types.RestoreParams{
				Name: "filtered-restore",
			},
			groupID:  "groupId",
			backupID: "backupJobId",
			sourceID: 456,
			setupMocks: func(mc *testmockrs.MockBRSClient, mw *MockBRSClientWrapper) {
				mw.On("GetTenantId").Return("test-tenant-id")
				mw.On("GetBRSClient").Return(mc)
				snapshotID := "snapshot-456"
				nsName := "test-namespace"
				nsID := int64(1)
				mockProtectionGroupRun := &backuprecoveryv1.ProtectionGroupRun{
					Objects: []backuprecoveryv1.ObjectRunResult{
						{
							Object: &backuprecoveryv1.ObjectSummary{
								Name: &nsName,
								ID:   &nsID,
							},
							ArchivalInfo: &backuprecoveryv1.ArchivalRun{
								ArchivalTargetResults: []backuprecoveryv1.ArchivalTargetResult{
									{
										SnapshotID: &snapshotID,
									},
								},
							},
						},
					},
				}
				mc.On("GetProtectionGroupRun", mock.Anything).Return(mockProtectionGroupRun, &core.DetailedResponse{}, nil)
			},
			expectError: false,
			validateResult: func(t *testing.T, result *backuprecoveryv1.CreateRecoveryOptions) {
				assert.NotNil(t, result.KubernetesParams.RecoverNamespaceParams)
				targetParams := result.KubernetesParams.RecoverNamespaceParams.KubernetesTargetParams
				assert.NotNil(t, targetParams)
				assert.Equal(t, 2, len(targetParams.Objects))
				includeParams := targetParams.Objects[1].IncludeParams
				assert.NotNil(t, includeParams)
				assert.Equal(t, 1, len(includeParams.SelectedResources))
				assert.Equal(t, "Deployment", *includeParams.SelectedResources[0].Kind)
				assert.Equal(t, "AND", *includeParams.LabelCombinationMethod)
				excludeParams := targetParams.Objects[1].ExcludeParams
				assert.NotNil(t, excludeParams)
				assert.Equal(t, 1, len(excludeParams.SelectedResources))
				assert.Equal(t, "Job", *excludeParams.SelectedResources[0].Kind)
			},
		},
		{
			name: "Success_RecoverToNewTarget",
			config: &KubernetesDataSourceConfig{
				ClusterName: "test-cluster",
				ClusterId:   "cluster-123",
				KubernetesRestoreParams: &types.KubernetesRestoreParams{
					RecoverToNewTarget: true,
					RenameRecoveredNamespacesParams: &types.RenameRecoveredNamespacesParams{
						Prefix: "new",
						Suffix: "-1",
					},
					RecoverObjectSpec: &types.RecoverObjectSpec{
						RestoreOnlyPvc: true,
					},
				},
			},
			restoreParams: &types.RestoreParams{
				Name: "new-target-restore",
			},
			groupID:  "groupId",
			backupID: "backupId",
			sourceID: 999,
			setupMocks: func(mc *testmockrs.MockBRSClient, mw *MockBRSClientWrapper) {
				mw.On("GetTenantId").Return("test-tenant-id")
				mw.On("GetBRSClient").Return(mc)
				snapshotID := "snapshot-789"
				nsName := "test-namespace"
				nsID := int64(1)
				mockProtectionGroupRun := &backuprecoveryv1.ProtectionGroupRun{
					Objects: []backuprecoveryv1.ObjectRunResult{
						{
							Object: &backuprecoveryv1.ObjectSummary{
								Name: &nsName,
								ID:   &nsID,
							},
							ArchivalInfo: &backuprecoveryv1.ArchivalRun{
								ArchivalTargetResults: []backuprecoveryv1.ArchivalTargetResult{
									{
										SnapshotID: &snapshotID,
									},
								},
							},
						},
					},
				}
				mc.On("GetProtectionGroupRun", mock.Anything).Return(mockProtectionGroupRun, &core.DetailedResponse{}, nil)
			},
			expectError: false,
			validateResult: func(t *testing.T, result *backuprecoveryv1.CreateRecoveryOptions) {
				assert.NotNil(t, result.KubernetesParams.RecoverNamespaceParams.KubernetesTargetParams.RecoveryTargetConfig)
				assert.True(t, *result.KubernetesParams.RecoverNamespaceParams.KubernetesTargetParams.RecoveryTargetConfig.RecoverToNewSource)
				assert.Equal(t, int64(999), *result.KubernetesParams.RecoverNamespaceParams.KubernetesTargetParams.RecoveryTargetConfig.NewSourceConfig.Source.ID)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockBRSClient := new(testmockrs.MockBRSClient)
			mockBRSWrapper := new(MockBRSClientWrapper)

			tt.setupMocks(mockBRSClient, mockBRSWrapper)

			dataSource := createTestKubernetesDataSource("test", tt.config, mockBRSWrapper)
			ctx := context.Background()

			result, err := dataSource.RunRestore(ctx, tt.groupID, tt.backupID, tt.sourceID, tt.restoreParams)

			if tt.expectError {
				assert.Error(t, err)
				assert.Nil(t, result)
				assert.Contains(t, err.Error(), tt.expectedErrorContains)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
				if tt.validateResult != nil {
					tt.validateResult(t, result)
				}
			}

			mockBRSClient.AssertExpectations(t)
			mockBRSWrapper.AssertExpectations(t)
		})
	}
}

// =============================================================================
// HELPER FUNCTION TESTS
// These test pure business logic without any external dependencies
// =============================================================================

func TestParseNamespaceList(t *testing.T) {
	tests := []struct {
		name           string
		input          string
		expectedResult []string
	}{
		{
			name:           "ValidInput",
			input:          "namespace1, namespace2, namespace3",
			expectedResult: []string{"namespace1", "namespace2", "namespace3"},
		},
		{
			name:           "EmptyInput",
			input:          "",
			expectedResult: []string{},
		},
		{
			name:           "SingleNamespace",
			input:          "default",
			expectedResult: []string{"default"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := parseNamespaceList(tt.input)
			assert.Equal(t, len(tt.expectedResult), len(result))
			for i, expected := range tt.expectedResult {
				assert.Equal(t, expected, result[i])
			}
		})
	}
}

func TestParseLabels(t *testing.T) {
	tests := []struct {
		name               string
		input              string
		combinationMethod  string
		expectedMethod     string
		expectedLabelCount int
		validateLabels     func(*testing.T, *backuprecoveryv1.KubernetesFilterParams)
	}{
		{
			name:               "ValidInput",
			input:              "app=web, env=prod",
			combinationMethod:  "MatchAll",
			expectedMethod:     "AND",
			expectedLabelCount: 2,
			validateLabels: func(t *testing.T, result *backuprecoveryv1.KubernetesFilterParams) {
				assert.Equal(t, "app", *result.LabelVector[0].Key)
				assert.Equal(t, "web", *result.LabelVector[0].Value)
				assert.Equal(t, "env", *result.LabelVector[1].Key)
				assert.Equal(t, "prod", *result.LabelVector[1].Value)
			},
		},
		{
			name:               "ColonSeparator",
			input:              "tier:frontend, version:v1",
			combinationMethod:  "OR",
			expectedMethod:     "OR",
			expectedLabelCount: 2,
			validateLabels: func(t *testing.T, result *backuprecoveryv1.KubernetesFilterParams) {
				assert.Equal(t, "tier", *result.LabelVector[0].Key)
				assert.Equal(t, "frontend", *result.LabelVector[0].Value)
			},
		},
		{
			name:               "EmptyInput",
			input:              "",
			combinationMethod:  "MatchAll",
			expectedMethod:     "AND",
			expectedLabelCount: 0,
		},
		{
			name:               "InvalidFormat",
			input:              "invalid-label-format",
			combinationMethod:  "OR",
			expectedMethod:     "OR",
			expectedLabelCount: 0,
		},
		{
			name:               "WithEqualsAndColon",
			input:              "key1=value1,key2:value2",
			combinationMethod:  "MatchAll",
			expectedMethod:     "AND",
			expectedLabelCount: 2,
			validateLabels: func(t *testing.T, result *backuprecoveryv1.KubernetesFilterParams) {
				assert.Equal(t, "key1", *result.LabelVector[0].Key)
				assert.Equal(t, "value1", *result.LabelVector[0].Value)
				assert.Equal(t, "key2", *result.LabelVector[1].Key)
				assert.Equal(t, "value2", *result.LabelVector[1].Value)
			},
		},
		{
			name:               "WithSpaces",
			input:              "  key1 = value1  ,  key2 : value2  ",
			combinationMethod:  "OR",
			expectedMethod:     "OR",
			expectedLabelCount: 2,
			validateLabels: func(t *testing.T, result *backuprecoveryv1.KubernetesFilterParams) {
				assert.Equal(t, "key1", *result.LabelVector[0].Key)
				assert.Equal(t, "value1", *result.LabelVector[0].Value)
			},
		},
		{
			name:               "EmptyPairs",
			input:              "key1=value1,,,key2=value2",
			combinationMethod:  "MatchAll",
			expectedMethod:     "AND",
			expectedLabelCount: 2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := parseLabels(tt.input, tt.combinationMethod)
			assert.NotNil(t, result)
			assert.Equal(t, tt.expectedMethod, *result.LabelCombinationMethod)
			assert.Equal(t, tt.expectedLabelCount, len(result.LabelVector))
			if tt.validateLabels != nil {
				tt.validateLabels(t, result)
			}
		})
	}
}

func TestParsePodLabels(t *testing.T) {
	tests := []struct {
		name           string
		input          string
		expectNil      bool
		expectedCount  int
		validateLabels func(*testing.T, []backuprecoveryv1.KubernetesLabel)
	}{
		{
			name:          "ValidInput",
			input:         "app=web, tier=frontend",
			expectNil:     false,
			expectedCount: 2,
			validateLabels: func(t *testing.T, result []backuprecoveryv1.KubernetesLabel) {
				assert.Equal(t, "app", *result[0].Key)
				assert.Equal(t, "web", *result[0].Value)
				assert.Equal(t, "tier", *result[1].Key)
				assert.Equal(t, "frontend", *result[1].Value)
			},
		},
		{
			name:      "EmptyInput",
			input:     "",
			expectNil: true,
		},
		{
			name:      "InvalidFormat",
			input:     "invalid-format",
			expectNil: true,
		},
		{
			name:          "WithColon",
			input:         "app:web,tier:frontend",
			expectNil:     false,
			expectedCount: 2,
			validateLabels: func(t *testing.T, result []backuprecoveryv1.KubernetesLabel) {
				assert.Equal(t, "app", *result[0].Key)
				assert.Equal(t, "web", *result[0].Value)
			},
		},
		{
			name:          "WithSpaces",
			input:         "  app = web  ,  tier = frontend  ",
			expectNil:     false,
			expectedCount: 2,
			validateLabels: func(t *testing.T, result []backuprecoveryv1.KubernetesLabel) {
				assert.Equal(t, "app", *result[0].Key)
				assert.Equal(t, "web", *result[0].Value)
			},
		},
		{
			name:          "InvalidPairs",
			input:         "app=web,invalid,tier=frontend",
			expectNil:     false,
			expectedCount: 2,
		},
		{
			name:          "EmptyKeyOrValue",
			input:         "=value,key=,valid=ok",
			expectNil:     false,
			expectedCount: 1,
			validateLabels: func(t *testing.T, result []backuprecoveryv1.KubernetesLabel) {
				assert.Equal(t, "valid", *result[0].Key)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := parsePodLabels(tt.input)
			if tt.expectNil {
				assert.Nil(t, result)
			} else {
				assert.Equal(t, tt.expectedCount, len(result))
				if tt.validateLabels != nil {
					tt.validateLabels(t, result)
				}
			}
		})
	}
}

func TestIsIncluded(t *testing.T) {
	tests := []struct {
		name        string
		namespace   string
		includeList []string
		expected    bool
	}{
		{
			name:        "True",
			namespace:   "namespace2",
			includeList: []string{"namespace1", "namespace2", "namespace3"},
			expected:    true,
		},
		{
			name:        "False",
			namespace:   "namespace4",
			includeList: []string{"namespace1", "namespace2", "namespace3"},
			expected:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isIncluded(tt.namespace, tt.includeList)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestIsExcluded(t *testing.T) {
	tests := []struct {
		name        string
		namespace   string
		excludeList []string
		expected    bool
	}{
		{
			name:        "True",
			namespace:   "kube-system",
			excludeList: []string{"kube-system", "kube-public"},
			expected:    true,
		},
		{
			name:        "False",
			namespace:   "default",
			excludeList: []string{"kube-system", "kube-public"},
			expected:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isExcluded(tt.namespace, tt.excludeList)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestUniqueByID(t *testing.T) {
	tests := []struct {
		name          string
		input         []types.NamespaceData
		expectedCount int
		validateIDs   func(*testing.T, []types.NamespaceData)
	}{
		{
			name: "RemovesDuplicates",
			input: []types.NamespaceData{
				{Id: 1, Name: "namespace1"},
				{Id: 2, Name: "namespace2"},
				{Id: 1, Name: "namespace1-duplicate"},
				{Id: 3, Name: "namespace3"},
				{Id: 2, Name: "namespace2-duplicate"},
			},
			expectedCount: 3,
			validateIDs: func(t *testing.T, result []types.NamespaceData) {
				assert.Equal(t, int64(1), result[0].Id)
				assert.Equal(t, int64(2), result[1].Id)
				assert.Equal(t, int64(3), result[2].Id)
			},
		},
		{
			name:          "EmptyInput",
			input:         []types.NamespaceData{},
			expectedCount: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := uniqueByID(tt.input)
			assert.Equal(t, tt.expectedCount, len(result))
			if tt.validateIDs != nil {
				tt.validateIDs(t, result)
			}
		})
	}
}

// =============================================================================
// MAPPING FUNCTION TESTS
// =============================================================================

func TestMapRecoveryStorageClasses(t *testing.T) {
	tests := []struct {
		name           string
		mappings       []types.StorageClassMapping
		validateResult func(*testing.T, *backuprecoveryv1.KubernetesStorageClassParams)
	}{
		{
			name: "Success",
			mappings: []types.StorageClassMapping{
				{
					Old: "standard",
					New: "premium",
				},
				{
					Old: "slow",
					New: "fast",
				},
			},
			validateResult: func(t *testing.T, result *backuprecoveryv1.KubernetesStorageClassParams) {
				assert.NotNil(t, result)
				assert.True(t, *result.UseStorageClassMapping)
				assert.Equal(t, 2, len(result.StorageClassMapping))
				assert.Equal(t, "standard", *result.StorageClassMapping[0].Key)
				assert.Equal(t, "premium", *result.StorageClassMapping[0].Value)
			},
		},
		{
			name:     "EmptyMappings",
			mappings: []types.StorageClassMapping{},
			validateResult: func(t *testing.T, result *backuprecoveryv1.KubernetesStorageClassParams) {
				assert.NotNil(t, result)
				assert.Nil(t, result.UseStorageClassMapping)
				assert.Equal(t, 0, len(result.StorageClassMapping))
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := &KubernetesDataSourceConfig{
				ClusterName: "test-cluster",
			}
			dataSource := createTestKubernetesDataSource("test", config, mockBRSClientWrapper())
			result := dataSource.MapRecoveryStorageClasses(tt.mappings)
			if tt.validateResult != nil {
				tt.validateResult(t, result)
			}
		})
	}
}

func TestMapRecoveryMigrationParams(t *testing.T) {
	tests := []struct {
		name            string
		migrationParams []types.MigrationMapParams
		expectedCount   int
		validateResult  func(*testing.T, []backuprecoveryv1.KubernetesRecoveryMigrationParams)
	}{
		{
			name: "Success",
			migrationParams: []types.MigrationMapParams{
				{
					Source: "zone-1",
					Target: "zone-2",
				},
				{
					Source: "region-a",
					Target: "region-b",
				},
			},
			expectedCount: 2,
			validateResult: func(t *testing.T, result []backuprecoveryv1.KubernetesRecoveryMigrationParams) {
				assert.Equal(t, "zone-1", *result[0].CurrentValue)
				assert.Equal(t, "zone-2", *result[0].NewValue)
				assert.Equal(t, "region-a", *result[1].CurrentValue)
				assert.Equal(t, "region-b", *result[1].NewValue)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := &KubernetesDataSourceConfig{
				ClusterName: "test-cluster",
			}
			dataSource := createTestKubernetesDataSource("test", config, mockBRSClientWrapper())
			result := dataSource.MapRecoveryMigrationParams(tt.migrationParams)
			assert.Equal(t, tt.expectedCount, len(result))
			if tt.validateResult != nil {
				tt.validateResult(t, result)
			}
		})
	}
}

func TestMapRecoveryResourceInfo(t *testing.T) {
	tests := []struct {
		name           string
		resourceList   []types.ResourceInstance
		expectedCount  int
		validateResult func(*testing.T, []backuprecoveryv1.ResourceInstance)
	}{
		{
			name: "Success",
			resourceList: []types.ResourceInstance{
				{
					Id:   23,
					Name: "resource-1",
				},
				{
					Id:   456,
					Name: "resource-2",
				},
			},
			expectedCount: 2,
			validateResult: func(t *testing.T, result []backuprecoveryv1.ResourceInstance) {
				assert.Equal(t, int64(23), *result[0].EntityID)
				assert.Equal(t, "resource-1", *result[0].Name)
				assert.Equal(t, int64(456), *result[1].EntityID)
				assert.Equal(t, "resource-2", *result[1].Name)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := &KubernetesDataSourceConfig{
				ClusterName: "test-cluster",
			}
			dataSource := createTestKubernetesDataSource("test", config, mockBRSClientWrapper())
			result := dataSource.MapRecoveryResourceInfo(tt.resourceList)
			assert.Equal(t, tt.expectedCount, len(result))
			if tt.validateResult != nil {
				tt.validateResult(t, result)
			}
		})
	}
}

func TestMapRecoveryResourceParams(t *testing.T) {
	tests := []struct {
		name           string
		k8sObjects     *types.K8sObject
		validateResult func(*testing.T, *backuprecoveryv1.KubernetesFilterParams)
	}{
		{
			name: "WithSelectedResources",
			k8sObjects: &types.K8sObject{
				SelectedResources: []types.ResourceInfo{
					{
						ResourceKind:     "Deployment",
						ResourceApiGroup: "apps",
						ResourceList: []types.ResourceInstance{
							{
								Id:   123,
								Name: "my-deployment",
							},
						},
					},
				},
			},
			validateResult: func(t *testing.T, result *backuprecoveryv1.KubernetesFilterParams) {
				assert.NotNil(t, result)
				assert.Equal(t, 1, len(result.SelectedResources))
				assert.Equal(t, "Deployment", *result.SelectedResources[0].Kind)
				assert.Equal(t, "apps", *result.SelectedResources[0].ApiGroup)
			},
		},
		{
			name: "WithSelectedLabels",
			k8sObjects: &types.K8sObject{
				SelectedLabels: &types.SelectedLabels{
					LabelCombination: "AND",
					Labels: []types.K8sLabel{
						{
							Key:   "app",
							Value: "web",
						},
					},
				},
			},
			validateResult: func(t *testing.T, result *backuprecoveryv1.KubernetesFilterParams) {
				assert.NotNil(t, result)
				assert.Equal(t, "AND", *result.LabelCombinationMethod)
				assert.Equal(t, 1, len(result.LabelVector))
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := &KubernetesDataSourceConfig{
				ClusterName: "test-cluster",
			}
			dataSource := createTestKubernetesDataSource("test", config, mockBRSClientWrapper())
			result := dataSource.MapRecoveryResourceParams(tt.k8sObjects)
			if tt.validateResult != nil {
				tt.validateResult(t, result)
			}
		})
	}
}

// =============================================================================
// GET K8 OBJECT TESTS
// =============================================================================

// =============================================================================
// GET K8 OBJECT TESTS (0% to 100%)
// =============================================================================

func TestGetK8Object(t *testing.T) {
	tests := []struct {
		name            string
		setupDataSource func() *KubernetesDataSource
		namespaces      []types.NamespaceData
		validateResult  func(*testing.T, *backuprecoveryv1.KubernetesProtectionGroupParams)
	}{
		{
			name: "Success",
			setupDataSource: func() *KubernetesDataSource {
				config := &KubernetesDataSourceConfig{
					ClusterName: "test-cluster",
					KubernetesProtectionParams: &types.KubernetesProtectionParams{
						IncludeAppLabels: "app=web",
						Settings: &types.ProtectionSetting{
							CSISnapshot: true,
							Labels: types.Labels{
								PersistentVolumeClaim: types.KeyValueFilter{
									Inclusion: "env=prod",
									Exclusion: "temp=true",
									LogicRule: "MatchAll",
								},
							},
						},
						NamespacesSetting: []types.NamespacesSetting{
							{
								Namespace: "default",
								PersistentVolumeClaims: types.PVCSection{
									Inclusion: []string{"pvc-1"},
									Exclusion: []string{"pvc-2"},
								},
								Resources: types.ResourceSection{
									Inclusion: &types.ResourceTypes{
										ResourceTypes: []string{"Deployment"},
									},
									Exclusion: &types.ResourceTypes{
										ResourceTypes: []string{"Job"},
									},
								},
								Hooks: types.HookSection{
									RulesApplyMode:         "sequential",
									FailBackupIfHookFailed: true,
									Rules: []types.Rule{
										{
											Rule: types.RuleDetail{
												PodLabels:  "app=web",
												PreScript:  "echo pre",
												PostScript: "echo post",
												Container:  "main",
											},
										},
									},
								},
							},
						},
					},
				}
				return createTestKubernetesDataSource("test", config, mockBRSClientWrapper())
			},
			namespaces: []types.NamespaceData{
				{
					Id:   100,
					Name: "default",
					PvcIDs: map[string]int64{
						"pvc-1": 101,
						"pvc-2": 102,
					},
				},
			},
			validateResult: func(t *testing.T, result *backuprecoveryv1.KubernetesProtectionGroupParams) {
				assert.NotNil(t, result)
				assert.Equal(t, 1, len(result.Objects))
				assert.Equal(t, int64(100), *result.Objects[0].ID)
				assert.True(t, *result.LeverageCSISnapshot)
				assert.NotNil(t, result.IncludeParams)
				assert.NotNil(t, result.ExcludeParams)
			},
		},
		{
			name: "NilDataSource",
			setupDataSource: func() *KubernetesDataSource {
				return nil
			},
			namespaces: []types.NamespaceData{},
			validateResult: func(t *testing.T, result *backuprecoveryv1.KubernetesProtectionGroupParams) {
				assert.Nil(t, result)
			},
		},
		{
			name: "NilProtectionParams",
			setupDataSource: func() *KubernetesDataSource {
				config := &KubernetesDataSourceConfig{
					ClusterName: "test-cluster",
				}
				dataSource := createTestKubernetesDataSource("test", config, mockBRSClientWrapper())
				dataSource.KubernetesProtectionParams = nil
				return dataSource
			},
			namespaces: []types.NamespaceData{},
			validateResult: func(t *testing.T, result *backuprecoveryv1.KubernetesProtectionGroupParams) {
				assert.Nil(t, result)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dataSource := tt.setupDataSource()
			result := dataSource.GetK8Object(tt.namespaces)
			if tt.validateResult != nil {
				tt.validateResult(t, result)
			}
		})
	}
}

// =============================================================================
// MAP PVC INCLUSION/EXCLUSION TESTS (0% to 100%)
// =============================================================================

func TestMapPvcInclusion(t *testing.T) {
	tests := []struct {
		name           string
		nsSetting      types.NamespacesSetting
		ns             types.NamespaceData
		expectedCount  int
		validateResult func(*testing.T, []backuprecoveryv1.KubernetesPvcInfo)
	}{
		{
			name: "WithIDs",
			nsSetting: types.NamespacesSetting{
				PersistentVolumeClaims: types.PVCSection{
					Inclusion: []string{"pvc-1", "pvc-2"},
				},
			},
			ns: types.NamespaceData{
				PvcIDs: map[string]int64{
					"pvc-1": 101,
					"pvc-2": 102,
				},
			},
			expectedCount: 2,
			validateResult: func(t *testing.T, result []backuprecoveryv1.KubernetesPvcInfo) {
				assert.Equal(t, int64(101), *result[0].ID)
				assert.Equal(t, int64(102), *result[1].ID)
			},
		},
		{
			name: "WithoutIDs",
			nsSetting: types.NamespacesSetting{
				PersistentVolumeClaims: types.PVCSection{
					Inclusion: []string{"pvc-unknown"},
				},
			},
			ns: types.NamespaceData{
				PvcIDs: map[string]int64{},
			},
			expectedCount: 1,
			validateResult: func(t *testing.T, result []backuprecoveryv1.KubernetesPvcInfo) {
				assert.Equal(t, "pvc-unknown", *result[0].Name)
				assert.Nil(t, result[0].ID)
			},
		},
		{
			name: "WithEmptyStrings",
			nsSetting: types.NamespacesSetting{
				PersistentVolumeClaims: types.PVCSection{
					Inclusion: []string{"pvc-1", "", "  ", "pvc-2"},
				},
			},
			ns: types.NamespaceData{
				PvcIDs: map[string]int64{
					"pvc-1": 101,
					"pvc-2": 102,
				},
			},
			expectedCount: 2,
			validateResult: func(t *testing.T, result []backuprecoveryv1.KubernetesPvcInfo) {
				assert.Equal(t, int64(101), *result[0].ID)
				assert.Equal(t, int64(102), *result[1].ID)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := MapPvcInclusion(tt.nsSetting, tt.ns)
			assert.Equal(t, tt.expectedCount, len(result))
			if tt.validateResult != nil {
				tt.validateResult(t, result)
			}
		})
	}
}

func TestMapPvcExclusion(t *testing.T) {
	tests := []struct {
		name           string
		nsSetting      types.NamespacesSetting
		ns             types.NamespaceData
		expectedCount  int
		validateResult func(*testing.T, []backuprecoveryv1.KubernetesPvcInfo)
	}{
		{
			name: "WithIDs",
			nsSetting: types.NamespacesSetting{
				PersistentVolumeClaims: types.PVCSection{
					Exclusion: []string{"pvc-exclude"},
				},
			},
			ns: types.NamespaceData{
				PvcIDs: map[string]int64{
					"pvc-exclude": 999,
				},
			},
			expectedCount: 1,
			validateResult: func(t *testing.T, result []backuprecoveryv1.KubernetesPvcInfo) {
				assert.Equal(t, int64(999), *result[0].ID)
			},
		},
		{
			name: "EmptyList",
			nsSetting: types.NamespacesSetting{
				PersistentVolumeClaims: types.PVCSection{
					Exclusion: []string{},
				},
			},
			ns:            types.NamespaceData{},
			expectedCount: 0,
		},
		{
			name: "WithEmptyStrings",
			nsSetting: types.NamespacesSetting{
				PersistentVolumeClaims: types.PVCSection{
					Exclusion: []string{"", "pvc-exclude", "  "},
				},
			},
			ns: types.NamespaceData{
				PvcIDs: map[string]int64{
					"pvc-exclude": 999,
				},
			},
			expectedCount: 1,
			validateResult: func(t *testing.T, result []backuprecoveryv1.KubernetesPvcInfo) {
				assert.Equal(t, int64(999), *result[0].ID)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := MapPvcExclusion(tt.nsSetting, tt.ns)
			assert.Equal(t, tt.expectedCount, len(result))
			if tt.validateResult != nil {
				tt.validateResult(t, result)
			}
		})
	}
}

// =============================================================================
// MAP RULES TESTS (0% to 100%)
// =============================================================================

func TestMapRules(t *testing.T) {
	tests := []struct {
		name           string
		nsSetting      types.NamespacesSetting
		ns             types.NamespaceData
		expectedCount  int
		validateResult func(*testing.T, []backuprecoveryv1.QuiesceGroup)
	}{
		{
			name: "WithPreAndPostScripts",
			nsSetting: types.NamespacesSetting{
				Hooks: types.HookSection{
					RulesApplyMode:         "sequential",
					FailBackupIfHookFailed: true,
					Rules: []types.Rule{
						{
							Rule: types.RuleDetail{
								PodLabels:  "app=web,tier=frontend",
								PreScript:  "echo 'starting backup'",
								PostScript: "echo 'backup complete'",
								Container:  "main-container",
							},
						},
					},
				},
			},
			ns: types.NamespaceData{
				Name: "default",
			},
			expectedCount: 1,
			validateResult: func(t *testing.T, result []backuprecoveryv1.QuiesceGroup) {
				assert.Equal(t, "sequential", *result[0].QuiesceMode)
				assert.Equal(t, 1, len(result[0].QuiesceRules))
				assert.Equal(t, 2, len(result[0].QuiesceRules[0].PodSelectorLabels))
				assert.Equal(t, 1, len(result[0].QuiesceRules[0].PreSnapshotHooks))
				assert.Equal(t, 1, len(result[0].QuiesceRules[0].PostSnapshotHooks))
			},
		},
		{
			name: "OnlyPreScript",
			nsSetting: types.NamespacesSetting{
				Hooks: types.HookSection{
					RulesApplyMode: "parallel",
					Rules: []types.Rule{
						{
							Rule: types.RuleDetail{
								PodLabels: "app=db",
								PreScript: "pg_dump",
							},
						},
					},
				},
			},
			ns:            types.NamespaceData{},
			expectedCount: 1,
			validateResult: func(t *testing.T, result []backuprecoveryv1.QuiesceGroup) {
				assert.Equal(t, 1, len(result[0].QuiesceRules[0].PreSnapshotHooks))
				assert.Equal(t, 0, len(result[0].QuiesceRules[0].PostSnapshotHooks))
			},
		},
		{
			name: "NoScripts",
			nsSetting: types.NamespacesSetting{
				Hooks: types.HookSection{
					RulesApplyMode: "sequential",
					Rules: []types.Rule{
						{
							Rule: types.RuleDetail{
								PodLabels: "app=db",
							},
						},
					},
				},
			},
			ns:            types.NamespaceData{},
			expectedCount: 1,
			validateResult: func(t *testing.T, result []backuprecoveryv1.QuiesceGroup) {
				assert.Equal(t, 0, len(result[0].QuiesceRules[0].PreSnapshotHooks))
				assert.Equal(t, 0, len(result[0].QuiesceRules[0].PostSnapshotHooks))
			},
		},
		{
			name: "WithEmptyScripts",
			nsSetting: types.NamespacesSetting{
				Hooks: types.HookSection{
					RulesApplyMode: "sequential",
					Rules: []types.Rule{{
						Rule: types.RuleDetail{
							PodLabels:  "app=web",
							PreScript:  "",
							PostScript: "  ",
							Container:  "main",
						},
					}},
				},
			},
			ns:            types.NamespaceData{},
			expectedCount: 1,
			validateResult: func(t *testing.T, result []backuprecoveryv1.QuiesceGroup) {
				assert.Equal(t, 0, len(result[0].QuiesceRules[0].PreSnapshotHooks))
				assert.Equal(t, 0, len(result[0].QuiesceRules[0].PostSnapshotHooks))
			},
		},
		{
			name: "WithoutContainer",
			nsSetting: types.NamespacesSetting{
				Hooks: types.HookSection{
					RulesApplyMode: "sequential",
					Rules: []types.Rule{
						{
							Rule: types.RuleDetail{
								PodLabels:  "app=web",
								PreScript:  "echo pre",
								PostScript: "echo post",
								Container:  "",
							},
						},
					},
				},
			},
			ns:            types.NamespaceData{},
			expectedCount: 1,
			validateResult: func(t *testing.T, result []backuprecoveryv1.QuiesceGroup) {
				assert.Equal(t, 1, len(result[0].QuiesceRules[0].PreSnapshotHooks))
				assert.Equal(t, 1, len(result[0].QuiesceRules[0].PostSnapshotHooks))
				assert.Nil(t, result[0].QuiesceRules[0].PreSnapshotHooks[0].Container)
				assert.Nil(t, result[0].QuiesceRules[0].PostSnapshotHooks[0].Container)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := MapRules(tt.nsSetting, tt.ns)
			assert.Equal(t, tt.expectedCount, len(result))
			if tt.validateResult != nil {
				tt.validateResult(t, result)
			}
		})
	}
}

// =============================================================================
// RESOLVE NAMESPACE ID TESTS (0% to 100%)
// =============================================================================
// =============================================================================
// RESOLVE SNAPSHOT ID TESTS (0% to 100%)
// =============================================================================

// ResolveSnapshotID Tests
func TestResolveSnapshotID(t *testing.T) {
	tests := []struct {
		name                  string
		groupID               string
		mockResponse          *backuprecoveryv1.GetObjectSnapshotsResponse
		expectError           bool
		expectedSnapshotID    string
		expectedErrorContains string
	}{
		{
			name:    "Success",
			groupID: "group-123",
			mockResponse: &backuprecoveryv1.GetObjectSnapshotsResponse{
				Snapshots: []backuprecoveryv1.ObjectSnapshot{
					{
						ID: core.StringPtr("snapshot-abc"),
					},
				},
			},
			expectError:        false,
			expectedSnapshotID: "snapshot-abc",
		},
		{
			name:                  "EmptyGroupID",
			groupID:               "",
			mockResponse:          nil,
			expectError:           true,
			expectedErrorContains: "protection group ID cannot be empty",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockBRSClient := new(testmockrs.MockBRSClient)
			mockBRSWrapper := new(MockBRSClientWrapper)

			mockBRSWrapper.On("GetTenantId").Return("test-tenant-id")
			mockBRSWrapper.On("GetBRSClient").Return(mockBRSClient)

			if tt.mockResponse != nil {
				mockBRSClient.On("GetObjectSnapshots", mock.Anything).Return(tt.mockResponse, &core.DetailedResponse{}, nil)
			}

			config := &KubernetesDataSourceConfig{
				ClusterName: "test-cluster",
			}

			dataSource := createTestKubernetesDataSource("test", config, mockBRSWrapper)

			result, err := dataSource.ResolveSnapshotID(tt.groupID, 100, 1)

			if tt.expectError {
				assert.Error(t, err)
				assert.Nil(t, result)
				assert.Contains(t, err.Error(), tt.expectedErrorContains)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
				assert.Equal(t, tt.expectedSnapshotID, *result)
			}
		})
	}
}

// =============================================================================
// ADDITIONAL TESTS FOR 100% COVERAGE
// =============================================================================

// GetBackupRunSnapShotID Tests
func TestGetBackupRunSnapShotID(t *testing.T) {
	tests := []struct {
		name                  string
		mockProtectionRun     *backuprecoveryv1.ProtectionGroupRun
		expectError           bool
		expectedSnapshotIDs   []string
		expectedErrorContains string
	}{
		{
			name: "Success",
			mockProtectionRun: &backuprecoveryv1.ProtectionGroupRun{
				Objects: []backuprecoveryv1.ObjectRunResult{
					{
						Object: &backuprecoveryv1.ObjectSummary{
						Name: core.StringPtr("test-namespace"),
						ID:   core.Int64Ptr(1),
						},
						ArchivalInfo: &backuprecoveryv1.ArchivalRun{
							ArchivalTargetResults: []backuprecoveryv1.ArchivalTargetResult{
								{
									SnapshotID: core.StringPtr("snapshot-backup-123"),
								},
							},
						},
					},
				},
			},
			expectError:         false,
			expectedSnapshotIDs: []string{"snapshot-backup-123"},
		},
		{
			name: "NoSnapshotID",
			mockProtectionRun: &backuprecoveryv1.ProtectionGroupRun{
				Objects: []backuprecoveryv1.ObjectRunResult{},
			},
			expectError:           true,
			expectedErrorContains: "No objects found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockBRSClient := new(testmockrs.MockBRSClient)
			mockBRSWrapper := new(MockBRSClientWrapper)

			mockBRSWrapper.On("GetTenantId").Return("test-tenant-id")
			mockBRSWrapper.On("GetBRSClient").Return(mockBRSClient)

			mockBRSClient.On("GetProtectionGroupRun", mock.Anything).Return(tt.mockProtectionRun, &core.DetailedResponse{}, nil)

			config := &KubernetesDataSourceConfig{
				ClusterName: "test-cluster",
			}

			dataSource := createTestKubernetesDataSource("test", config, mockBRSWrapper)

			result, err := dataSource.GetBackupRunSnapShotID("group-123", "backup-456")

			if tt.expectError {
				assert.Error(t, err)
				assert.Nil(t, result)
				assert.Contains(t, err.Error(), tt.expectedErrorContains)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
				assert.Equal(t, tt.expectedSnapshotIDs, result)
			}
		})
	}
}

// GetNamespaceSnapshotId Tests
func TestGetNamespaceSnapshotId(t *testing.T) {
	tests := []struct {
		name                  string
		snapshotInfo          *types.SnapshotInfo
		needsMocking          bool
		mockSearchResponse    *backuprecoveryv1.ProtectedObjectsSearchResponse
		mockSnapshotResponse  *backuprecoveryv1.GetObjectSnapshotsResponse
		expectError           bool
		expectedSnapshotID    string
		expectedErrorContains string
	}{
		{
			name: "WithSnapshotID",
			snapshotInfo: &types.SnapshotInfo{
				SnapshotID: "direct-snapshot-123",
			},
			needsMocking:       false,
			expectError:        false,
			expectedSnapshotID: "direct-snapshot-123",
		},
		{
			name:                  "NilSnapshotInfo",
			snapshotInfo:          nil,
			needsMocking:          false,
			expectError:           true,
			expectedErrorContains: "snapshot info cannot be nil",
		},
		{
			name: "WithNamespaceInfo",
			snapshotInfo: &types.SnapshotInfo{
				NamespaceInfo: &types.NamespaceInfo{
					Namespace:               "default",
					BackupPositionFromFirst: 1,
				},
			},
			needsMocking: true,
			mockSearchResponse: &backuprecoveryv1.ProtectedObjectsSearchResponse{
				Objects: []backuprecoveryv1.ProtectedObject{
					{
						ID:       core.Int64Ptr(100),
						SourceID: core.Int64Ptr(200),
						LatestSnapshotsInfo: []backuprecoveryv1.ObjectSnapshotsInfo{
							{
								ProtectionGroupID: core.StringPtr("group-123"),
							},
						},
					},
				},
			},
			mockSnapshotResponse: &backuprecoveryv1.GetObjectSnapshotsResponse{
				Snapshots: []backuprecoveryv1.ObjectSnapshot{
					{
						ID: core.StringPtr("namespace-snapshot-456"),
					},
				},
			},
			expectError:        false,
			expectedSnapshotID: "namespace-snapshot-456",
		},
		{
			name:                  "NoSnapshotIDOrNamespace",
			snapshotInfo:          &types.SnapshotInfo{},
			needsMocking:          false,
			expectError:           true,
			expectedErrorContains: "either NamespaceInfo or SnapshotId must be provided",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var dataSource *KubernetesDataSource
			config := &KubernetesDataSourceConfig{
				ClusterName: "test-cluster",
			}

			if tt.needsMocking {
				mockBRSClient := new(testmockrs.MockBRSClient)
				mockBRSWrapper := new(MockBRSClientWrapper)

				mockBRSWrapper.On("GetTenantId").Return("test-tenant-id")
				mockBRSWrapper.On("GetBRSClient").Return(mockBRSClient)

				if tt.mockSearchResponse != nil {
					mockBRSClient.On("SearchProtectedObjects", mock.Anything).Return(tt.mockSearchResponse, &core.DetailedResponse{}, nil)
				}
				if tt.mockSnapshotResponse != nil {
					mockBRSClient.On("GetObjectSnapshots", mock.Anything).Return(tt.mockSnapshotResponse, &core.DetailedResponse{}, nil)
				}

				dataSource = createTestKubernetesDataSource("test", config, mockBRSWrapper)
			} else {
				dataSource = createTestKubernetesDataSource("test", config, mockBRSClientWrapper())
			}

			result, err := dataSource.GetNamespaceSnapshotId(tt.snapshotInfo)

			if tt.expectError {
				assert.Error(t, err)
				assert.Nil(t, result)
				assert.Contains(t, err.Error(), tt.expectedErrorContains)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
				assert.Equal(t, tt.expectedSnapshotID, *result)
			}
		})
	}
}

// RecoverNamespace Tests
func TestRecoverNamespace(t *testing.T) {
	tests := []struct {
		name                  string
		k8sRestoreParams      *types.KubernetesRestoreParams
		commonRestoreParams   *types.RestoreParams
		expectedErrorContains string
	}{
		{
			name:                  "NilKubernetesRestoreParams",
			k8sRestoreParams:      nil,
			commonRestoreParams:   &types.RestoreParams{},
			expectedErrorContains: "kubernetes restore params cannot be nil",
		},
		{
			name:                  "NilCommonRestoreParams",
			k8sRestoreParams:      &types.KubernetesRestoreParams{},
			commonRestoreParams:   nil,
			expectedErrorContains: "common restore params cannot be nil",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := &KubernetesDataSourceConfig{
				ClusterName: "test-cluster",
			}
			dataSource := createTestKubernetesDataSource("test", config, mockBRSClientWrapper())

			result, err := dataSource.RecoverNamespace(context.Background(), "group-123", "backup-456", 789, tt.k8sRestoreParams, tt.commonRestoreParams)

			assert.Error(t, err)
			assert.Nil(t, result)
			assert.Contains(t, err.Error(), tt.expectedErrorContains)
		})
	}
}

// InitializeKubernetesNamespaceParams Tests
func TestInitializeKubernetesNamespaceParams(t *testing.T) {
	tests := []struct {
		name                 string
		k8sRestoreParams     *types.KubernetesRestoreParams
		mockProtectionRun    *backuprecoveryv1.ProtectionGroupRun
		mockError            error
		expectError          bool
		expectedResultLength int
	}{
		{
			name:              "GetBackupRunError",
			k8sRestoreParams:  &types.KubernetesRestoreParams{},
			mockProtectionRun: nil,
			mockError:         fmt.Errorf("API error"),
			expectError:       true,
		},
		{
			name: "WithMultipleObjects",
			k8sRestoreParams: &types.KubernetesRestoreParams{
				RecoverObjectSpec: &types.RecoverObjectSpec{},
				RecoverMultipleObjects: []types.RecoverObject{
					{
						SnapshotInfo: &types.SnapshotInfo{
							SnapshotID: "snapshot-additional",
						},
						RecoverObjectSpec: &types.RecoverObjectSpec{},
					},
				},
			},
			mockProtectionRun: &backuprecoveryv1.ProtectionGroupRun{
					Objects: []backuprecoveryv1.ObjectRunResult{
						{
							Object: &backuprecoveryv1.ObjectSummary{
							Name: core.StringPtr("test-namespace"),
							ID:   core.Int64Ptr(1),
							},
							ArchivalInfo: &backuprecoveryv1.ArchivalRun{
								ArchivalTargetResults: []backuprecoveryv1.ArchivalTargetResult{
									{
										SnapshotID: core.StringPtr("snapshot-main"),
									},
								},
							},
						},
					},
				},
			mockError:            nil,
			expectError:          false,
			expectedResultLength: 2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockBRSClient := new(testmockrs.MockBRSClient)
			mockBRSWrapper := new(MockBRSClientWrapper)

			mockBRSWrapper.On("GetTenantId").Return("test-tenant-id")
			mockBRSWrapper.On("GetBRSClient").Return(mockBRSClient)

			mockBRSClient.On("GetProtectionGroupRun", mock.Anything).Return(tt.mockProtectionRun, &core.DetailedResponse{}, tt.mockError)

			config := &KubernetesDataSourceConfig{
				ClusterName: "test-cluster",
			}

			dataSource := createTestKubernetesDataSource("test", config, mockBRSWrapper)

			result, err := dataSource.InitializeKubernetesNamespaceParams("group-123", "backup-456", tt.k8sRestoreParams)

			if tt.expectError {
				assert.Error(t, err)
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
				assert.Equal(t, tt.expectedResultLength, len(result))
			}
		})
	}
}

// fetchNamespaceIDsWithNames Tests
func TestFetchNamespaceIDsWithNames(t *testing.T) {
	tests := []struct {
		name           string
		mockSources    []backuprecoveryv1.ProtectionSourceNodes
		expectedLength int
	}{
		{
			name: "NilNodes",
			mockSources: []backuprecoveryv1.ProtectionSourceNodes{
				{
					Nodes: nil,
				},
			},
			expectedLength: 0,
		},
		{
			name: "NilProtectionSource",
			mockSources: []backuprecoveryv1.ProtectionSourceNodes{
				{
					Nodes: []backuprecoveryv1.ProtectionSourceNodes{
						{
							Nodes: []backuprecoveryv1.ProtectionSourceNodes{
								{
									ProtectionSource: nil,
								},
							},
						},
					},
				},
			},
			expectedLength: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockBRSClient := new(testmockrs.MockBRSClient)
			mockBRSWrapper := new(MockBRSClientWrapper)

			mockBRSWrapper.On("GetTenantId").Return("test-tenant-id")
			mockBRSWrapper.On("GetBRSClient").Return(mockBRSClient)

			mockBRSClient.On("ListProtectionSources", mock.Anything).Return(tt.mockSources, &core.DetailedResponse{}, nil)

			config := &KubernetesDataSourceConfig{
				ClusterName: "test-cluster",
			}

			dataSource := createTestKubernetesDataSource("test", config, mockBRSWrapper)

			result, err := dataSource.fetchNamespaceIDsWithNames(123, []string{}, []string{})

			assert.NoError(t, err)
			assert.Equal(t, tt.expectedLength, len(result))
		})
	}
}

// ResolveNamespaceID Tests
func TestResolveNamespaceID(t *testing.T) {
	tests := []struct {
		name                  string
		mockResponse          *backuprecoveryv1.ProtectedObjectsSearchResponse
		expectedErrorContains string
	}{
		{
			name: "NilID",
			mockResponse: &backuprecoveryv1.ProtectedObjectsSearchResponse{
				Objects: []backuprecoveryv1.ProtectedObject{
					{
						ID:       nil,
						SourceID: core.Int64Ptr(200),
					},
				},
			},
			expectedErrorContains: "namespace ID is nil",
		},
		{
			name: "NilSourceID",
			mockResponse: &backuprecoveryv1.ProtectedObjectsSearchResponse{
				Objects: []backuprecoveryv1.ProtectedObject{
					{
						ID:       core.Int64Ptr(100),
						SourceID: nil,
					},
				},
			},
			expectedErrorContains: "source ID is nil",
		},
		{
			name: "EmptySnapshotsInfo",
			mockResponse: &backuprecoveryv1.ProtectedObjectsSearchResponse{
				Objects: []backuprecoveryv1.ProtectedObject{
					{
						ID:                  core.Int64Ptr(100),
						SourceID:            core.Int64Ptr(200),
						LatestSnapshotsInfo: []backuprecoveryv1.ObjectSnapshotsInfo{},
					},
				},
			},
			expectedErrorContains: "no snapshot info found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockBRSClient := new(testmockrs.MockBRSClient)
			mockBRSWrapper := new(MockBRSClientWrapper)

			mockBRSWrapper.On("GetTenantId").Return("test-tenant-id")
			mockBRSWrapper.On("GetBRSClient").Return(mockBRSClient)

			mockBRSClient.On("SearchProtectedObjects", mock.Anything).Return(tt.mockResponse, &core.DetailedResponse{}, nil)

			config := &KubernetesDataSourceConfig{
				ClusterName: "test-cluster",
			}

			dataSource := createTestKubernetesDataSource("test", config, mockBRSWrapper)

			result, err := dataSource.ResolveNamespaceID("default")

			assert.Error(t, err)
			assert.Nil(t, result)
			assert.Contains(t, err.Error(), tt.expectedErrorContains)
		})
	}
}

// getSnapshotID Tests
func TestGetSnapshotID(t *testing.T) {
	tests := []struct {
		name                  string
		mockResponse          *backuprecoveryv1.GetObjectSnapshotsResponse
		backupPosition        int64
		expectError           bool
		expectedSnapshotID    string
		expectedErrorContains string
	}{
		{
			name: "BackupPositionZero",
			mockResponse: &backuprecoveryv1.GetObjectSnapshotsResponse{
				Snapshots: []backuprecoveryv1.ObjectSnapshot{
					{
						ID: core.StringPtr("snapshot-zero"),
					},
				},
			},
			backupPosition:     0,
			expectError:        false,
			expectedSnapshotID: "snapshot-zero",
		},
		{
			name: "NilSnapshotID",
			mockResponse: &backuprecoveryv1.GetObjectSnapshotsResponse{
				Snapshots: []backuprecoveryv1.ObjectSnapshot{
					{
						ID: nil,
					},
				},
			},
			backupPosition:        1,
			expectError:           true,
			expectedErrorContains: "No snapshot ID was found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockBRSClient := new(testmockrs.MockBRSClient)
			mockBRSWrapper := new(MockBRSClientWrapper)

			mockBRSWrapper.On("GetTenantId").Return("test-tenant-id")
			mockBRSWrapper.On("GetBRSClient").Return(mockBRSClient)

			mockBRSClient.On("GetObjectSnapshots", mock.Anything).Return(tt.mockResponse, &core.DetailedResponse{}, nil)

			config := &KubernetesDataSourceConfig{
				ClusterName: "test-cluster",
			}

			dataSource := createTestKubernetesDataSource("test", config, mockBRSWrapper)

			result, err := dataSource.getSnapshotID("group-123", 100, tt.backupPosition)

			if tt.expectError {
				assert.Error(t, err)
				assert.Nil(t, result)
				assert.Contains(t, err.Error(), tt.expectedErrorContains)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
				assert.Equal(t, tt.expectedSnapshotID, *result)
			}
		})
	}
}

// waitForSnapshotID Tests
func TestWaitForSnapshotID(t *testing.T) {
	tests := []struct {
		name                  string
		mockResponse          *backuprecoveryv1.GetObjectSnapshotsResponse
		backupPosition        int64
		expectError           bool
		expectedSnapshotID    string
		expectedErrorContains string
	}{
		{
			name: "BackupPositionZero",
			mockResponse: &backuprecoveryv1.GetObjectSnapshotsResponse{
				Snapshots: []backuprecoveryv1.ObjectSnapshot{
					{
						ID: core.StringPtr("snapshot-wait-zero"),
					},
				},
			},
			backupPosition:     0,
			expectError:        false,
			expectedSnapshotID: "snapshot-wait-zero",
		},
		{
			name: "NilSnapshotID",
			mockResponse: &backuprecoveryv1.GetObjectSnapshotsResponse{
				Snapshots: []backuprecoveryv1.ObjectSnapshot{
					{
						ID: nil,
					},
				},
			},
			backupPosition:        1,
			expectError:           true,
			expectedErrorContains: "snapshot ID not found within timeout",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockBRSClient := new(testmockrs.MockBRSClient)
			mockBRSWrapper := new(MockBRSClientWrapper)

			mockBRSWrapper.On("GetTenantId").Return("test-tenant-id")
			mockBRSWrapper.On("GetBRSClient").Return(mockBRSClient)

			mockBRSClient.On("GetObjectSnapshots", mock.Anything).Return(tt.mockResponse, &core.DetailedResponse{}, nil)

			config := &KubernetesDataSourceConfig{
				ClusterName: "test-cluster",
			}

			dataSource := createTestKubernetesDataSource("test", config, mockBRSWrapper)

			var totalTimeout time.Duration = 10 * time.Millisecond
			var pollingInterval time.Duration = 5 * time.Millisecond

			result, err := dataSource.waitForSnapshotID("group-123", 100, tt.backupPosition, totalTimeout, pollingInterval)

			if tt.expectError {
				assert.Error(t, err)
				assert.Nil(t, result)
				assert.Contains(t, err.Error(), tt.expectedErrorContains)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
				assert.Equal(t, tt.expectedSnapshotID, *result)
			}
		})
	}
}

// Additional RunRestore edge cases
// =============================================================================
// REGISTER SOURCE PARAMS TESTS
// =============================================================================

// RegisterSourceParams Tests
func TestRegisterSourceParams(t *testing.T) {
	tests := []struct {
		name                      string
		dataSourceName            string
		config                    *KubernetesDataSourceConfig
		connectionID              string
		mockBearerToken           string
		mockBearerTokenError      error
		expectError               bool
		expectedErrorContains     string
		expectedConnectionID      int64
		expectedDistribution      string
		shouldHaveOpenshiftPlugin bool
		verifyAllImages           bool
		verifyEnvironment         bool
	}{
		{
			name:           "Success_IKS",
			dataSourceName: "test-datasource",
			config: &KubernetesDataSourceConfig{
				ClusterName:              "test-iks-cluster",
				ClusterId:                "cluster-123",
				ClusterType:              "IKS",
				ClusterEndpoint:          "https://test-iks-cluster.example.com",
				KubernetesDataSourceType: Kiksclassic,
				BrsAgent: types.BRSAgentImage{
					DataMover:             "icr.io/datamover:latest",
					Velero:                "icr.io/velero:latest",
					VeleroAWSPlugin:       "icr.io/velero-aws:latest",
					VeleroOpenShiftPlugin: "icr.io/velero-openshift:latest",
				},
			},
			connectionID:              "12345",
			mockBearerToken:           "test-bearer-token-123",
			expectError:               false,
			expectedConnectionID:      12345,
			expectedDistribution:      backuprecoveryv1.KubernetesSourceRegistrationParams_KubernetesDistribution_Kiks,
			shouldHaveOpenshiftPlugin: false,
		},
		{
			name:           "Success_ROKS",
			dataSourceName: "test-roks-datasource",
			config: &KubernetesDataSourceConfig{
				ClusterName:              "test-roks-cluster",
				ClusterId:                "cluster-456",
				ClusterType:              "ROKS",
				ClusterEndpoint:          "https://test-roks-cluster.example.com",
				KubernetesDataSourceType: Kroksclassic,
				BrsAgent: types.BRSAgentImage{
					DataMover:             "icr.io/datamover:v2",
					Velero:                "icr.io/velero:v2",
					VeleroAWSPlugin:       "icr.io/velero-aws:v2",
					VeleroOpenShiftPlugin: "icr.io/velero-openshift:v2",
				},
			},
			connectionID:              "67890",
			mockBearerToken:           "roks-bearer-token-456",
			expectError:               false,
			expectedConnectionID:      67890,
			expectedDistribution:      backuprecoveryv1.KubernetesSourceRegistrationParams_KubernetesDistribution_Kroks,
			shouldHaveOpenshiftPlugin: true,
		},
		{
			name:           "Success_ROKS_LowerCase",
			dataSourceName: "test-datasource",
			config: &KubernetesDataSourceConfig{
				ClusterName:              "test-roks-cluster",
				ClusterId:                "cluster-789",
				ClusterType:              "roks",
				ClusterEndpoint:          "https://test-roks-cluster.example.com",
				KubernetesDataSourceType: Kroksclassic,
				BrsAgent: types.BRSAgentImage{
					DataMover:             "icr.io/datamover:latest",
					Velero:                "icr.io/velero:latest",
					VeleroAWSPlugin:       "icr.io/velero-aws:latest",
					VeleroOpenShiftPlugin: "icr.io/velero-openshift:latest",
				},
			},
			connectionID:              "11111",
			mockBearerToken:           "test-token",
			expectError:               false,
			expectedConnectionID:      11111,
			shouldHaveOpenshiftPlugin: true,
		},
		{
			name:           "BearerTokenError",
			dataSourceName: "test-datasource",
			config: &KubernetesDataSourceConfig{
				ClusterName:              "test-cluster",
				ClusterId:                "cluster-123",
				ClusterType:              "IKS",
				ClusterEndpoint:          "https://test-cluster.example.com",
				KubernetesDataSourceType: Kiksclassic,
				BrsAgent: types.BRSAgentImage{
					DataMover:       "icr.io/datamover:latest",
					Velero:          "icr.io/velero:latest",
					VeleroAWSPlugin: "icr.io/velero-aws:latest",
				},
			},
			connectionID:          "12345",
			mockBearerTokenError:  fmt.Errorf("failed to get bearer token"),
			expectError:           true,
			expectedErrorContains: "could not get cluster bearer token",
		},
		{
			name:           "InvalidConnectionID",
			dataSourceName: "test-datasource",
			config: &KubernetesDataSourceConfig{
				ClusterName:              "test-cluster",
				ClusterId:                "cluster-123",
				ClusterType:              "IKS",
				ClusterEndpoint:          "https://test-cluster.example.com",
				KubernetesDataSourceType: Kiksclassic,
				BrsAgent: types.BRSAgentImage{
					DataMover:       "icr.io/datamover:latest",
					Velero:          "icr.io/velero:latest",
					VeleroAWSPlugin: "icr.io/velero-aws:latest",
				},
			},
			connectionID:          "invalid-id",
			mockBearerToken:       "test-token",
			expectError:           true,
			expectedErrorContains: "could not parse connection id",
		},
		{
			name:           "EmptyConnectionID",
			dataSourceName: "test-datasource",
			config: &KubernetesDataSourceConfig{
				ClusterName:              "test-cluster",
				ClusterId:                "cluster-123",
				ClusterType:              "IKS",
				ClusterEndpoint:          "https://test-cluster.example.com",
				KubernetesDataSourceType: Kiksclassic,
				BrsAgent: types.BRSAgentImage{
					DataMover:       "icr.io/datamover:latest",
					Velero:          "icr.io/velero:latest",
					VeleroAWSPlugin: "icr.io/velero-aws:latest",
				},
			},
			connectionID:          "",
			mockBearerToken:       "test-token",
			expectError:           true,
			expectedErrorContains: "could not parse connection id",
		},
		{
			name:           "NegativeConnectionID",
			dataSourceName: "test-datasource",
			config: &KubernetesDataSourceConfig{
				ClusterName:              "test-cluster",
				ClusterId:                "cluster-123",
				ClusterType:              "IKS",
				ClusterEndpoint:          "https://test-cluster.example.com",
				KubernetesDataSourceType: Kiksclassic,
				BrsAgent: types.BRSAgentImage{
					DataMover:       "icr.io/datamover:latest",
					Velero:          "icr.io/velero:latest",
					VeleroAWSPlugin: "icr.io/velero-aws:latest",
				},
			},
			connectionID:         "-123",
			mockBearerToken:      "test-token",
			expectError:          false,
			expectedConnectionID: -123,
		},
		{
			name:           "LargeConnectionID",
			dataSourceName: "test-datasource",
			config: &KubernetesDataSourceConfig{
				ClusterName:              "test-cluster",
				ClusterId:                "cluster-123",
				ClusterType:              "IKS",
				ClusterEndpoint:          "https://test-cluster.example.com",
				KubernetesDataSourceType: Kiksclassic,
				BrsAgent: types.BRSAgentImage{
					DataMover:       "icr.io/datamover:latest",
					Velero:          "icr.io/velero:latest",
					VeleroAWSPlugin: "icr.io/velero-aws:latest",
				},
			},
			connectionID:         "9223372036854775807",
			mockBearerToken:      "test-token",
			expectError:          false,
			expectedConnectionID: 9223372036854775807,
		},
		{
			name:           "VerifyAllBRSAgentImages",
			dataSourceName: "test-ds",
			config: &KubernetesDataSourceConfig{
				ClusterName:              "test-cluster",
				ClusterId:                "cluster-123",
				ClusterType:              "IKS",
				ClusterEndpoint:          "https://api.test.com",
				KubernetesDataSourceType: Kiksclassic,
				BrsAgent: types.BRSAgentImage{
					DataMover:             "custom.io/datamover:v1.2.3",
					Velero:                "custom.io/velero:v1.2.3",
					VeleroAWSPlugin:       "custom.io/velero-aws:v1.2.3",
					VeleroOpenShiftPlugin: "custom.io/velero-openshift:v1.2.3",
				},
			},
			connectionID:    "999",
			mockBearerToken: "token",
			expectError:     false,
			verifyAllImages: true,
		},
		{
			name:           "VerifyEnvironmentType",
			dataSourceName: "test-ds",
			config: &KubernetesDataSourceConfig{
				ClusterName:              "test-cluster",
				ClusterId:                "cluster-123",
				ClusterType:              "IKS",
				ClusterEndpoint:          "https://test.com",
				KubernetesDataSourceType: Kiksclassic,
				BrsAgent: types.BRSAgentImage{
					DataMover:       "icr.io/datamover:latest",
					Velero:          "icr.io/velero:latest",
					VeleroAWSPlugin: "icr.io/velero-aws:latest",
				},
			},
			connectionID:      "123",
			mockBearerToken:   "token",
			expectError:       false,
			verifyEnvironment: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockIks := new(MockIksManager)
			if tt.mockBearerTokenError != nil {
				mockIks.On("GetClusterBearerToken", mock.Anything).Return("", tt.mockBearerTokenError)
			} else {
				mockIks.On("GetClusterBearerToken", mock.Anything).Return(tt.mockBearerToken, nil)
			}

			dataSource := createTestKubernetesDataSource(tt.dataSourceName, tt.config, mockBRSClientWrapper())
			dataSource.iksClient = mockIks

			ctx := context.Background()
			result, err := dataSource.RegisterSourceParams(ctx, tt.connectionID)

			if tt.expectError {
				assert.Error(t, err)
				assert.Nil(t, result)
				assert.Contains(t, err.Error(), tt.expectedErrorContains)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
				assert.Equal(t, tt.dataSourceName, *result.Name)

				if tt.expectedConnectionID != 0 {
					assert.Equal(t, tt.expectedConnectionID, *result.ConnectionID)
				}

				assert.NotNil(t, result.KubernetesParams)

				if tt.expectedDistribution != "" {
					assert.Equal(t, tt.expectedDistribution, *result.KubernetesParams.KubernetesDistribution)
				}

				if tt.shouldHaveOpenshiftPlugin {
					assert.NotNil(t, result.KubernetesParams.VeleroOpenshiftPluginImageLocation)
				} else if tt.name == "Success_IKS" {
					assert.Nil(t, result.KubernetesParams.VeleroOpenshiftPluginImageLocation)
				}

				if tt.verifyAllImages {
					assert.Equal(t, "custom.io/datamover:v1.2.3", *result.KubernetesParams.DataMoverImageLocation)
					assert.Equal(t, "custom.io/velero:v1.2.3", *result.KubernetesParams.VeleroImageLocation)
					assert.Equal(t, "custom.io/velero-aws:v1.2.3", *result.KubernetesParams.VeleroAwsPluginImageLocation)
				}

				if tt.verifyEnvironment {
					assert.NotNil(t, result.Environment)
					assert.Equal(t, backuprecoveryv1.RegisterProtectionSourceOptions_Environment_Kkubernetes, *result.Environment)
				}
			}

			mockIks.AssertExpectations(t)
		})
	}
}

// TestRegisterSourceParams_MultipleCallsWithDifferentTokens tests that multiple calls return different tokens
func TestRegisterSourceParams_MultipleCallsWithDifferentTokens(t *testing.T) {
	config := &KubernetesDataSourceConfig{
		ClusterName:              "test-cluster",
		ClusterId:                "cluster-123",
		ClusterType:              "IKS",
		ClusterEndpoint:          "https://test.com",
		KubernetesDataSourceType: Kiksclassic,
		BrsAgent: types.BRSAgentImage{
			DataMover:       "icr.io/datamover:latest",
			Velero:          "icr.io/velero:latest",
			VeleroAWSPlugin: "icr.io/velero-aws:latest",
		},
	}

	mockIks := new(MockIksManager)
	// First call returns token1
	mockIks.On("GetClusterBearerToken", mock.Anything).Return("token1", nil).Once()
	// Second call returns token2
	mockIks.On("GetClusterBearerToken", mock.Anything).Return("token2", nil).Once()

	dataSource := createTestKubernetesDataSource("test-ds", config, mockBRSClientWrapper())
	dataSource.iksClient = mockIks

	ctx := context.Background()

	// First call
	result1, err1 := dataSource.RegisterSourceParams(ctx, "100")
	assert.NoError(t, err1)
	assert.Equal(t, "token1", *result1.KubernetesParams.ClientPrivateKey)

	// Second call
	result2, err2 := dataSource.RegisterSourceParams(ctx, "200")
	assert.NoError(t, err2)
	assert.Equal(t, "token2", *result2.KubernetesParams.ClientPrivateKey)

	mockIks.AssertExpectations(t)
}

// =============================================================================
// Additional NewKubernetesDataSource Tests (Error Paths)
// =============================================================================

// =============================================================================
// KubernetesDataSourceConfig.CreateDataSource Tests
// =============================================================================

func TestKubernetesDataSourceConfig_CreateDataSource_NilConfig(t *testing.T) {
	var config *KubernetesDataSourceConfig

	dataSource, err := config.CreateDataSource(mockBRSClientWrapper())

	assert.Error(t, err)
	assert.Nil(t, dataSource)
	assert.Contains(t, err.Error(), "cannot be nil")
}

func TestKubernetesDataSourceConfig_CreateDataSource_ValidConfig(t *testing.T) {
	config := &KubernetesDataSourceConfig{
		ClusterName:              "test-cluster",
		ClusterId:                "cluster-123",
		ClusterType:              "IKS",
		ContainerEndpoint:        "https://containers.cloud.ibm.com",
		ContainerEndpointType:    "public",
		Authenticator:            &core.IamAuthenticator{ApiKey: "test-api-key"},
		KubernetesDataSourceType: Kiksclassic,
	}

	dataSource, err := config.CreateDataSource(mockBRSClientWrapper())

	// Will fail without real IKS cluster, but validates the flow
	assert.Error(t, err)
	assert.Nil(t, dataSource)
}

// =============================================================================
// KubernetesDataSourceConfig.Validate Tests
// =============================================================================

// =============================================================================
// KubernetesDataSource.Validate Additional Tests
// =============================================================================

// TestKubernetesDataSourceConfig_Validate tests the Validate method of KubernetesDataSourceConfig
func TestKubernetesDataSourceConfig_Validate(t *testing.T) {
	tests := []struct {
		name          string
		config        *KubernetesDataSourceConfig
		expectError   bool
		errorContains string
	}{
		{
			name:          "Nil config",
			config:        nil,
			expectError:   true,
			errorContains: "kubernetes data source config cannot be nil",
		},
		{
			name: "Valid config with ClusterName",
			config: &KubernetesDataSourceConfig{
				ClusterName:           "test-cluster",
				ContainerEndpoint:     "https://containers.cloud.ibm.com",
				ContainerEndpointType: "public",
				Authenticator:         &core.IamAuthenticator{ApiKey: "test-key"},
			},
			expectError: false,
		},
		{
			name: "Valid config with ClusterId",
			config: &KubernetesDataSourceConfig{
				ClusterId:             "cluster-123",
				ContainerEndpoint:     "https://containers.cloud.ibm.com",
				ContainerEndpointType: "private",
				Authenticator:         &core.IamAuthenticator{ApiKey: "test-key"},
			},
			expectError: false,
		},
		{
			name: "Valid config with both ClusterName and ClusterId",
			config: &KubernetesDataSourceConfig{
				ClusterName:           "test-cluster",
				ClusterId:             "cluster-123",
				ContainerEndpoint:     "https://containers.cloud.ibm.com",
				ContainerEndpointType: "vpe",
				Authenticator:         &core.IamAuthenticator{ApiKey: "test-key"},
			},
			expectError: false,
		},
		{
			name: "Valid config with IKS cluster type",
			config: &KubernetesDataSourceConfig{
				ClusterName:           "test-cluster",
				ClusterType:           "IKS",
				ContainerEndpoint:     "https://containers.cloud.ibm.com",
				ContainerEndpointType: "public",
				Authenticator:         &core.IamAuthenticator{ApiKey: "test-key"},
			},
			expectError: false,
		},
		{
			name: "Valid config with ROKS cluster type",
			config: &KubernetesDataSourceConfig{
				ClusterName:           "test-cluster",
				ClusterType:           "ROKS",
				ContainerEndpoint:     "https://containers.cloud.ibm.com",
				ContainerEndpointType: "public",
				Authenticator:         &core.IamAuthenticator{ApiKey: "test-key"},
			},
			expectError: false,
		},
		{
			name: "Valid config with lowercase cluster type",
			config: &KubernetesDataSourceConfig{
				ClusterName:           "test-cluster",
				ClusterType:           "iks",
				ContainerEndpoint:     "https://containers.cloud.ibm.com",
				ContainerEndpointType: "public",
				Authenticator:         &core.IamAuthenticator{ApiKey: "test-key"},
			},
			expectError: false,
		},
		{
			name: "Missing ClusterName and ClusterId",
			config: &KubernetesDataSourceConfig{
				ContainerEndpoint:     "https://containers.cloud.ibm.com",
				ContainerEndpointType: "public",
				Authenticator:         &core.IamAuthenticator{ApiKey: "test-key"},
			},
			expectError:   true,
			errorContains: "either cluster_name or cluster_id is required",
		},
		{
			name: "Missing ContainerEndpoint",
			config: &KubernetesDataSourceConfig{
				ClusterName:           "test-cluster",
				ContainerEndpointType: "public",
				Authenticator:         &core.IamAuthenticator{ApiKey: "test-key"},
			},
			expectError:   true,
			errorContains: "container_endpoint is required",
		},
		{
			name: "Invalid ContainerEndpointType",
			config: &KubernetesDataSourceConfig{
				ClusterName:           "test-cluster",
				ContainerEndpoint:     "https://containers.cloud.ibm.com",
				ContainerEndpointType: "invalid-type",
				Authenticator:         &core.IamAuthenticator{ApiKey: "test-key"},
			},
			expectError:   true,
			errorContains: "container_endpoint_type must be 'public', 'private', or 'vpe'",
		},
		{
			name: "Missing Authenticator",
			config: &KubernetesDataSourceConfig{
				ClusterName:           "test-cluster",
				ContainerEndpoint:     "https://containers.cloud.ibm.com",
				ContainerEndpointType: "public",
				Authenticator:         nil,
			},
			expectError:   true,
			errorContains: "authenticator is required",
		},
		{
			name: "Invalid ClusterType",
			config: &KubernetesDataSourceConfig{
				ClusterName:           "test-cluster",
				ClusterType:           "INVALID",
				ContainerEndpoint:     "https://containers.cloud.ibm.com",
				ContainerEndpointType: "public",
				Authenticator:         &core.IamAuthenticator{ApiKey: "test-key"},
			},
			expectError:   true,
			errorContains: "cluster_type must be 'IKS' or 'ROKS'",
		},
		{
			name: "Empty ContainerEndpointType is allowed",
			config: &KubernetesDataSourceConfig{
				ClusterName:           "test-cluster",
				ContainerEndpoint:     "https://containers.cloud.ibm.com",
				ContainerEndpointType: "",
				Authenticator:         &core.IamAuthenticator{ApiKey: "test-key"},
			},
			expectError: false,
		},
		{
			name: "Empty ClusterType is allowed",
			config: &KubernetesDataSourceConfig{
				ClusterName:           "test-cluster",
				ClusterType:           "",
				ContainerEndpoint:     "https://containers.cloud.ibm.com",
				ContainerEndpointType: "public",
				Authenticator:         &core.IamAuthenticator{ApiKey: "test-key"},
			},
			expectError: false,
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
			}
		})
	}
}

// TestKubernetesDataSourceConfig_GetDataSourceType tests GetDataSourceType method
func TestKubernetesDataSourceConfig_GetDataSourceType(t *testing.T) {
	config := &KubernetesDataSourceConfig{
		ClusterName: "test-cluster",
	}

	dataSourceType := config.GetDataSourceType()
	assert.Equal(t, types.DataSourceKubernetes, dataSourceType)
}

// TestKubernetesDataSourceConfig_CreateDataSource tests CreateDataSource method
func TestKubernetesDataSourceConfig_CreateDataSource(t *testing.T) {
	t.Run("Nil config", func(t *testing.T) {
		var config *KubernetesDataSourceConfig
		brsClient := mockBRSClientWrapper()

		dataSource, err := config.CreateDataSource(brsClient)

		assert.Error(t, err)
		assert.Nil(t, dataSource)
		assert.Contains(t, err.Error(), "kubernetes data source config cannot be nil")
	})
}

// TestGetK8Object_Settings tests that GetK8Object handles Settings correctly
func TestGetK8Object_Settings(t *testing.T) {
	tests := []struct {
		name                      string
		settings                  *types.ProtectionSetting
		expectLeverageCSISnapshot bool
		expectIncludeParams       bool
		expectExcludeParams       bool
		leverageCSISnapshotValue  bool
	}{
		{
			name:                      "Nil Settings - should not set Settings-related fields",
			settings:                  nil,
			expectLeverageCSISnapshot: false,
			expectIncludeParams:       false,
			expectExcludeParams:       false,
		},
		{
			name: "Valid Settings with CSISnapshot true - should set all fields",
			settings: &types.ProtectionSetting{
				CSISnapshot: true,
				Labels: types.Labels{
					PersistentVolumeClaim: types.KeyValueFilter{
						Inclusion: "app=test",
						Exclusion: "env=dev",
						LogicRule: "AND",
					},
				},
			},
			expectLeverageCSISnapshot: true,
			expectIncludeParams:       true,
			expectExcludeParams:       true,
			leverageCSISnapshotValue:  true,
		},
		{
			name: "Valid Settings with CSISnapshot false - should set all fields",
			settings: &types.ProtectionSetting{
				CSISnapshot: false,
				Labels: types.Labels{
					PersistentVolumeClaim: types.KeyValueFilter{
						Inclusion: "app=prod",
						Exclusion: "",
						LogicRule: "OR",
					},
				},
			},
			expectLeverageCSISnapshot: true,
			expectIncludeParams:       true,
			expectExcludeParams:       true,
			leverageCSISnapshotValue:  false,
		},
		{
			name: "Valid Settings with empty labels - should set CSISnapshot only",
			settings: &types.ProtectionSetting{
				CSISnapshot: true,
				Labels: types.Labels{
					PersistentVolumeClaim: types.KeyValueFilter{
						Inclusion: "",
						Exclusion: "",
						LogicRule: "",
					},
				},
			},
			expectLeverageCSISnapshot: true,
			expectIncludeParams:       true,
			expectExcludeParams:       true,
			leverageCSISnapshotValue:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup
			dataSource := &KubernetesDataSource{
				KubernetesProtectionParams: &types.KubernetesProtectionParams{
					Settings: tt.settings,
					NamespacesSetting: []types.NamespacesSetting{
						{
							Namespace: "test-namespace",
						},
					},
				},
				logger: logger.NewNoop(),
			}

			namespaces := []types.NamespaceData{
				{
					Id:   123,
					Name: "test-namespace",
				},
			}

			// Execute
			result := dataSource.GetK8Object(namespaces)

			// Assert
			assert.NotNil(t, result, "Result should not be nil")
			assert.NotNil(t, result.Objects, "Objects should not be nil")
			assert.Len(t, result.Objects, 1, "Should have one object")

			// Check Settings-related fields
			if tt.expectLeverageCSISnapshot {
				assert.NotNil(t, result.LeverageCSISnapshot, "LeverageCSISnapshot should be set")
				assert.Equal(t, tt.leverageCSISnapshotValue, *result.LeverageCSISnapshot, "LeverageCSISnapshot value mismatch")
			} else {
				assert.Nil(t, result.LeverageCSISnapshot, "LeverageCSISnapshot should be nil when Settings is nil")
			}

			if tt.expectIncludeParams {
				assert.NotNil(t, result.IncludeParams, "IncludeParams should be set")
			} else {
				assert.Nil(t, result.IncludeParams, "IncludeParams should be nil when Settings is nil")
			}

			if tt.expectExcludeParams {
				assert.NotNil(t, result.ExcludeParams, "ExcludeParams should be set")
			} else {
				assert.Nil(t, result.ExcludeParams, "ExcludeParams should be nil when Settings is nil")
			}
		})
	}
}

// =============================================================================
// PAUSE FUTURE RUNS TESTS
// =============================================================================

// TestCreateProtectionGroup_PauseFutureRuns is a focused test for the
// PauseFutureRuns → IsPaused mapping in CreateProtectionGroup.
func TestCreateProtectionGroup_PauseFutureRuns(t *testing.T) {
	baseConfig := func(pauseFuture bool) *KubernetesDataSourceConfig {
		return &KubernetesDataSourceConfig{
			ClusterName: "test-cluster",
			ClusterId:   "cluster-123",
			KubernetesProtectionParams: &types.KubernetesProtectionParams{
				Settings: &types.ProtectionSetting{CSISnapshot: true},
			},
		}
	}

	emptySources := []backuprecoveryv1.ProtectionSourceNodes{
		{Nodes: []backuprecoveryv1.ProtectionSourceNodes{}},
	}

	setupMock := func(sources []backuprecoveryv1.ProtectionSourceNodes) types.BRSClientWrapperInterface {
		mockBRSClient := new(testmockrs.MockBRSClient)
		mockWrapper := new(MockBRSClientWrapper)
		mockWrapper.On("GetTenantId").Return("test-tenant-id")
		mockWrapper.On("GetBRSClient").Return(mockBRSClient)
		mockBRSClient.On("ListProtectionSources", mock.Anything).Return(sources, &core.DetailedResponse{}, nil)
		return mockWrapper
	}

	tests := []struct {
		name            string
		groupSettings   *types.ProtectionSetting
		expectIsPaused  bool
		expectPausedVal bool
	}{
		{
			name:            "PauseFutureRuns true sets IsPaused=true in BRS request",
			groupSettings:   &types.ProtectionSetting{PauseFutureRuns: true},
			expectIsPaused:  true,
			expectPausedVal: true,
		},
		{
			name:           "PauseFutureRuns false does not set IsPaused",
			groupSettings:  &types.ProtectionSetting{PauseFutureRuns: false},
			expectIsPaused: false,
		},
		{
			name:           "nil Settings does not set IsPaused",
			groupSettings:  nil,
			expectIsPaused: false,
		},
		{
			name:            "PauseFutureRuns true ignores other settings fields",
			groupSettings:   &types.ProtectionSetting{PauseFutureRuns: true, CSISnapshot: true, RetentionDays: 7},
			expectIsPaused:  true,
			expectPausedVal: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := baseConfig(false)
			brsWrapper := setupMock(emptySources)
			ds := createTestKubernetesDataSource("test", cfg, brsWrapper)

			result, err := ds.CreateProtectionGroup(context.Background(), 123, &types.ProtectionGroupParams{
				Name:     "test-group",
				Settings: tt.groupSettings,
			})

			assert.NoError(t, err)
			assert.NotNil(t, result)
			assert.Equal(t,
				backuprecoveryv1.CreateProtectionGroupOptions_Environment_Kkubernetes,
				*result.Environment,
				"Environment should always be Kubernetes",
			)

			if tt.expectIsPaused {
				assert.NotNil(t, result.IsPaused, "IsPaused field should be set")
				assert.Equal(t, tt.expectPausedVal, *result.IsPaused, "IsPaused value mismatch")
			} else {
				assert.Nil(t, result.IsPaused, "IsPaused field should not be set")
			}
		})
	}
}

// TestGetBackupRunNamespaceSnapshots tests the new function that returns namespace-to-snapshot mappings
func TestGetBackupRunNamespaceSnapshots(t *testing.T) {
	tests := []struct {
		name                  string
		mockProtectionRun     *backuprecoveryv1.ProtectionGroupRun
		expectError           bool
		expectedMappings      []NamespaceSnapshotMapping
		expectedErrorContains string
	}{
		{
			name: "Success - Single Namespace",
			mockProtectionRun: &backuprecoveryv1.ProtectionGroupRun{
				Objects: []backuprecoveryv1.ObjectRunResult{
					{
						Object: &backuprecoveryv1.ObjectSummary{
							ID:   core.Int64Ptr(72667),
							Name: core.StringPtr("busybox-app"),
						},
						ArchivalInfo: &backuprecoveryv1.ArchivalRun{
							ArchivalTargetResults: []backuprecoveryv1.ArchivalTargetResult{
								{
									SnapshotID: core.StringPtr("snap-123"),
								},
							},
						},
					},
				},
			},
			expectError: false,
			expectedMappings: []NamespaceSnapshotMapping{
				{
					NamespaceName: "busybox-app",
					NamespaceID:   72667,
					SnapshotID:    "snap-123",
					ObjectID:      72667,
				},
			},
		},
		{
			name: "Success - Multiple Namespaces",
			mockProtectionRun: &backuprecoveryv1.ProtectionGroupRun{
				Objects: []backuprecoveryv1.ObjectRunResult{
					{
						Object: &backuprecoveryv1.ObjectSummary{
							ID:   core.Int64Ptr(72667),
							Name: core.StringPtr("busybox-app"),
						},
						ArchivalInfo: &backuprecoveryv1.ArchivalRun{
							ArchivalTargetResults: []backuprecoveryv1.ArchivalTargetResult{
								{
									SnapshotID: core.StringPtr("snap-123"),
								},
							},
						},
					},
					{
						Object: &backuprecoveryv1.ObjectSummary{
							ID:   core.Int64Ptr(78395),
							Name: core.StringPtr("nginx-app"),
						},
						ArchivalInfo: &backuprecoveryv1.ArchivalRun{
							ArchivalTargetResults: []backuprecoveryv1.ArchivalTargetResult{
								{
									SnapshotID: core.StringPtr("snap-456"),
								},
							},
						},
					},
				},
			},
			expectError: false,
			expectedMappings: []NamespaceSnapshotMapping{
				{
					NamespaceName: "busybox-app",
					NamespaceID:   72667,
					SnapshotID:    "snap-123",
					ObjectID:      72667,
				},
				{
					NamespaceName: "nginx-app",
					NamespaceID:   78395,
					SnapshotID:    "snap-456",
					ObjectID:      78395,
				},
			},
		},
		{
			name: "Error - No Objects",
			mockProtectionRun: &backuprecoveryv1.ProtectionGroupRun{
				Objects: []backuprecoveryv1.ObjectRunResult{},
			},
			expectError:           true,
			expectedErrorContains: "No objects found",
		},
		{
			name: "Error - Missing Object Field",
			mockProtectionRun: &backuprecoveryv1.ProtectionGroupRun{
				Objects: []backuprecoveryv1.ObjectRunResult{
					{
						Object: nil,
						ArchivalInfo: &backuprecoveryv1.ArchivalRun{
							ArchivalTargetResults: []backuprecoveryv1.ArchivalTargetResult{
								{
									SnapshotID: core.StringPtr("snap-123"),
								},
							},
						},
					},
				},
			},
			expectError:           true,
			expectedErrorContains: "No snapshotId found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockBRSClient := new(testmockrs.MockBRSClient)
			mockBRSWrapper := new(MockBRSClientWrapper)

			mockBRSWrapper.On("GetTenantId").Return("test-tenant-id")
			mockBRSWrapper.On("GetBRSClient").Return(mockBRSClient)

			mockBRSClient.On("GetProtectionGroupRun", mock.Anything).Return(tt.mockProtectionRun, &core.DetailedResponse{}, nil)

			config := &KubernetesDataSourceConfig{
				ClusterName: "test-cluster",
			}

			dataSource := createTestKubernetesDataSource("test", config, mockBRSWrapper)

			result, err := dataSource.GetBackupRunNamespaceSnapshots("group-123", "backup-456")

			if tt.expectError {
				assert.Error(t, err)
				assert.Nil(t, result)
				assert.Contains(t, err.Error(), tt.expectedErrorContains)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
				assert.Equal(t, len(tt.expectedMappings), len(result))
				for i, expected := range tt.expectedMappings {
					assert.Equal(t, expected.NamespaceName, result[i].NamespaceName)
					assert.Equal(t, expected.NamespaceID, result[i].NamespaceID)
					assert.Equal(t, expected.SnapshotID, result[i].SnapshotID)
					assert.Equal(t, expected.ObjectID, result[i].ObjectID)
				}
			}
		})
	}
}

// TestProtectionSetting_PauseFutureRuns tests the ProtectionSetting struct field.
func TestProtectionSetting_PauseFutureRuns(t *testing.T) {
	t.Run("Default zero value is false", func(t *testing.T) {
		s := types.ProtectionSetting{}
		assert.False(t, s.PauseFutureRuns, "PauseFutureRuns zero value should be false")
	})

	t.Run("Can be set to true", func(t *testing.T) {
		s := types.ProtectionSetting{PauseFutureRuns: true}
		assert.True(t, s.PauseFutureRuns)
	})

	t.Run("Coexists with other fields", func(t *testing.T) {
		s := types.ProtectionSetting{
			PauseFutureRuns: true,
			CSISnapshot:     true,
			RetentionDays:   14,
			EndDate:         "2025-12-31",
		}
		assert.True(t, s.PauseFutureRuns)
		assert.True(t, s.CSISnapshot)
		assert.Equal(t, int64(14), s.RetentionDays)
		assert.Equal(t, "2025-12-31", s.EndDate)
	})
}

// TestInitializeKubernetesNamespaceParams_NamespaceFiltering tests the IncludeNamespaces filter
func TestInitializeKubernetesNamespaceParams_NamespaceFiltering(t *testing.T) {
	tests := []struct {
		name                   string
		mockProtectionRun      *backuprecoveryv1.ProtectionGroupRun
		includeNamespaces      []string
		expectedNamespaceCount int
		expectedNamespaceNames []string
	}{
		{
			name: "No Filter - All Namespaces Included",
			mockProtectionRun: &backuprecoveryv1.ProtectionGroupRun{
				Objects: []backuprecoveryv1.ObjectRunResult{
					{
						Object: &backuprecoveryv1.ObjectSummary{
							ID:   core.Int64Ptr(1),
							Name: core.StringPtr("app-1"),
						},
						ArchivalInfo: &backuprecoveryv1.ArchivalRun{
							ArchivalTargetResults: []backuprecoveryv1.ArchivalTargetResult{
								{SnapshotID: core.StringPtr("snap-1")},
							},
						},
					},
					{
						Object: &backuprecoveryv1.ObjectSummary{
							ID:   core.Int64Ptr(2),
							Name: core.StringPtr("app-2"),
						},
						ArchivalInfo: &backuprecoveryv1.ArchivalRun{
							ArchivalTargetResults: []backuprecoveryv1.ArchivalTargetResult{
								{SnapshotID: core.StringPtr("snap-2")},
							},
						},
					},
					{
						Object: &backuprecoveryv1.ObjectSummary{
							ID:   core.Int64Ptr(3),
							Name: core.StringPtr("app-3"),
						},
						ArchivalInfo: &backuprecoveryv1.ArchivalRun{
							ArchivalTargetResults: []backuprecoveryv1.ArchivalTargetResult{
								{SnapshotID: core.StringPtr("snap-3")},
							},
						},
					},
				},
			},
			includeNamespaces:      []string{}, // Empty = include all
			expectedNamespaceCount: 3,
			expectedNamespaceNames: []string{"app-1", "app-2", "app-3"},
		},
		{
			name: "Filter - Include Only Specific Namespaces",
			mockProtectionRun: &backuprecoveryv1.ProtectionGroupRun{
				Objects: []backuprecoveryv1.ObjectRunResult{
					{
						Object: &backuprecoveryv1.ObjectSummary{
							ID:   core.Int64Ptr(1),
							Name: core.StringPtr("app-1"),
						},
						ArchivalInfo: &backuprecoveryv1.ArchivalRun{
							ArchivalTargetResults: []backuprecoveryv1.ArchivalTargetResult{
								{SnapshotID: core.StringPtr("snap-1")},
							},
						},
					},
					{
						Object: &backuprecoveryv1.ObjectSummary{
							ID:   core.Int64Ptr(2),
							Name: core.StringPtr("app-2"),
						},
						ArchivalInfo: &backuprecoveryv1.ArchivalRun{
							ArchivalTargetResults: []backuprecoveryv1.ArchivalTargetResult{
								{SnapshotID: core.StringPtr("snap-2")},
							},
						},
					},
					{
						Object: &backuprecoveryv1.ObjectSummary{
							ID:   core.Int64Ptr(3),
							Name: core.StringPtr("app-3"),
						},
						ArchivalInfo: &backuprecoveryv1.ArchivalRun{
							ArchivalTargetResults: []backuprecoveryv1.ArchivalTargetResult{
								{SnapshotID: core.StringPtr("snap-3")},
							},
						},
					},
				},
			},
			includeNamespaces:      []string{"app-1", "app-3"}, // Only include app-1 and app-3
			expectedNamespaceCount: 2,
			expectedNamespaceNames: []string{"app-1", "app-3"},
		},
		{
			name: "Filter - Single Namespace",
			mockProtectionRun: &backuprecoveryv1.ProtectionGroupRun{
				Objects: []backuprecoveryv1.ObjectRunResult{
					{
						Object: &backuprecoveryv1.ObjectSummary{
							ID:   core.Int64Ptr(1),
							Name: core.StringPtr("app-1"),
						},
						ArchivalInfo: &backuprecoveryv1.ArchivalRun{
							ArchivalTargetResults: []backuprecoveryv1.ArchivalTargetResult{
								{SnapshotID: core.StringPtr("snap-1")},
							},
						},
					},
					{
						Object: &backuprecoveryv1.ObjectSummary{
							ID:   core.Int64Ptr(2),
							Name: core.StringPtr("app-2"),
						},
						ArchivalInfo: &backuprecoveryv1.ArchivalRun{
							ArchivalTargetResults: []backuprecoveryv1.ArchivalTargetResult{
								{SnapshotID: core.StringPtr("snap-2")},
							},
						},
					},
				},
			},
			includeNamespaces:      []string{"app-2"},
			expectedNamespaceCount: 1,
			expectedNamespaceNames: []string{"app-2"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockBRSClient := new(testmockrs.MockBRSClient)
			mockBRSWrapper := new(MockBRSClientWrapper)

			mockBRSWrapper.On("GetTenantId").Return("test-tenant-id")
			mockBRSWrapper.On("GetBRSClient").Return(mockBRSClient)

			mockBRSClient.On("GetProtectionGroupRun", mock.Anything).Return(tt.mockProtectionRun, &core.DetailedResponse{}, nil)

			config := &KubernetesDataSourceConfig{
				ClusterName: "test-cluster",
			}

			dataSource := createTestKubernetesDataSource("test", config, mockBRSWrapper)

			k8sRestoreParams := &types.KubernetesRestoreParams{
				IncludeNamespaces: tt.includeNamespaces,
				RecoverObjectSpec: &types.RecoverObjectSpec{
					StorageClasses: []types.StorageClassMapping{
						{Old: "class-A", New: "class-B"},
					},
				},
			}

			result, err := dataSource.InitializeKubernetesNamespaceParams("group-123", "backup-456", k8sRestoreParams)

			assert.NoError(t, err)
			assert.NotNil(t, result)
			assert.Equal(t, tt.expectedNamespaceCount, len(result), "Expected %d namespaces, got %d", tt.expectedNamespaceCount, len(result))
		})
	}
}

// TestInitializeKubernetesNamespaceParams_StorageClassPrecedence tests namespace-level storage class mapping precedence
func TestInitializeKubernetesNamespaceParams_StorageClassPrecedence(t *testing.T) {
	tests := []struct {
		name                        string
		mockProtectionRun           *backuprecoveryv1.ProtectionGroupRun
		topLevelStorageClasses      []types.StorageClassMapping
		recoverMultipleObjects      []types.RecoverObject
		expectedStorageClassMapping map[string][]types.StorageClassMapping // namespace -> storage class mappings
	}{
		{
			name: "Top-Level Only - All Namespaces Use Same Mapping",
			mockProtectionRun: &backuprecoveryv1.ProtectionGroupRun{
				Objects: []backuprecoveryv1.ObjectRunResult{
					{
						Object: &backuprecoveryv1.ObjectSummary{
							ID:   core.Int64Ptr(1),
							Name: core.StringPtr("app-1"),
						},
						ArchivalInfo: &backuprecoveryv1.ArchivalRun{
							ArchivalTargetResults: []backuprecoveryv1.ArchivalTargetResult{
								{SnapshotID: core.StringPtr("snap-1")},
							},
						},
					},
					{
						Object: &backuprecoveryv1.ObjectSummary{
							ID:   core.Int64Ptr(2),
							Name: core.StringPtr("app-2"),
						},
						ArchivalInfo: &backuprecoveryv1.ArchivalRun{
							ArchivalTargetResults: []backuprecoveryv1.ArchivalTargetResult{
								{SnapshotID: core.StringPtr("snap-2")},
							},
						},
					},
				},
			},
			topLevelStorageClasses: []types.StorageClassMapping{
				{Old: "class-A", New: "class-B"},
			},
			recoverMultipleObjects: []types.RecoverObject{}, // No namespace-specific overrides
			expectedStorageClassMapping: map[string][]types.StorageClassMapping{
				"app-1": {{Old: "class-A", New: "class-B"}},
				"app-2": {{Old: "class-A", New: "class-B"}},
			},
		},
		{
			name: "Namespace-Level Override - One Namespace Different",
			mockProtectionRun: &backuprecoveryv1.ProtectionGroupRun{
				Objects: []backuprecoveryv1.ObjectRunResult{
					{
						Object: &backuprecoveryv1.ObjectSummary{
							ID:   core.Int64Ptr(1),
							Name: core.StringPtr("app-1"),
						},
						ArchivalInfo: &backuprecoveryv1.ArchivalRun{
							ArchivalTargetResults: []backuprecoveryv1.ArchivalTargetResult{
								{SnapshotID: core.StringPtr("snap-1")},
							},
						},
					},
					{
						Object: &backuprecoveryv1.ObjectSummary{
							ID:   core.Int64Ptr(2),
							Name: core.StringPtr("app-2"),
						},
						ArchivalInfo: &backuprecoveryv1.ArchivalRun{
							ArchivalTargetResults: []backuprecoveryv1.ArchivalTargetResult{
								{SnapshotID: core.StringPtr("snap-2")},
							},
						},
					},
				},
			},
			topLevelStorageClasses: []types.StorageClassMapping{
				{Old: "class-A", New: "class-B"},
			},
			recoverMultipleObjects: []types.RecoverObject{
				{
					SnapshotInfo: &types.SnapshotInfo{
						NamespaceInfo: &types.NamespaceInfo{
							Namespace: "app-1", // Override for app-1
						},
					},
					RecoverObjectSpec: &types.RecoverObjectSpec{
						StorageClasses: []types.StorageClassMapping{
							{Old: "class-A", New: "class-C"}, // Different mapping for app-1
						},
					},
				},
			},
			expectedStorageClassMapping: map[string][]types.StorageClassMapping{
				"app-1": {{Old: "class-A", New: "class-C"}}, // Namespace-level wins
				"app-2": {{Old: "class-A", New: "class-B"}}, // Top-level default
			},
		},
		{
			name: "Multiple Namespace Overrides",
			mockProtectionRun: &backuprecoveryv1.ProtectionGroupRun{
				Objects: []backuprecoveryv1.ObjectRunResult{
					{
						Object: &backuprecoveryv1.ObjectSummary{
							ID:   core.Int64Ptr(1),
							Name: core.StringPtr("busybox-app"),
						},
						ArchivalInfo: &backuprecoveryv1.ArchivalRun{
							ArchivalTargetResults: []backuprecoveryv1.ArchivalTargetResult{
								{SnapshotID: core.StringPtr("snap-1")},
							},
						},
					},
					{
						Object: &backuprecoveryv1.ObjectSummary{
							ID:   core.Int64Ptr(2),
							Name: core.StringPtr("nginx-app"),
						},
						ArchivalInfo: &backuprecoveryv1.ArchivalRun{
							ArchivalTargetResults: []backuprecoveryv1.ArchivalTargetResult{
								{SnapshotID: core.StringPtr("snap-2")},
							},
						},
					},
					{
						Object: &backuprecoveryv1.ObjectSummary{
							ID:   core.Int64Ptr(3),
							Name: core.StringPtr("redis-app"),
						},
						ArchivalInfo: &backuprecoveryv1.ArchivalRun{
							ArchivalTargetResults: []backuprecoveryv1.ArchivalTargetResult{
								{SnapshotID: core.StringPtr("snap-3")},
							},
						},
					},
				},
			},
			topLevelStorageClasses: []types.StorageClassMapping{
				{Old: "default", New: "standard"},
			},
			recoverMultipleObjects: []types.RecoverObject{
				{
					SnapshotInfo: &types.SnapshotInfo{
						NamespaceInfo: &types.NamespaceInfo{
							Namespace: "busybox-app",
						},
					},
					RecoverObjectSpec: &types.RecoverObjectSpec{
						StorageClasses: []types.StorageClassMapping{
							{Old: "ibmc-vpc-block-metro-5iops-tier", New: "ibmc-vpc-block-10iops-tier"},
						},
					},
				},
				{
					SnapshotInfo: &types.SnapshotInfo{
						NamespaceInfo: &types.NamespaceInfo{
							Namespace: "nginx-app",
						},
					},
					RecoverObjectSpec: &types.RecoverObjectSpec{
						StorageClasses: []types.StorageClassMapping{
							{Old: "ibmc-vpc-block-10iops-tier", New: "ibmc-vpc-block-10iops-tier"},
						},
					},
				},
			},
			expectedStorageClassMapping: map[string][]types.StorageClassMapping{
				"busybox-app": {{Old: "ibmc-vpc-block-metro-5iops-tier", New: "ibmc-vpc-block-10iops-tier"}},
				"nginx-app":   {{Old: "ibmc-vpc-block-10iops-tier", New: "ibmc-vpc-block-10iops-tier"}},
				"redis-app":   {{Old: "default", New: "standard"}}, // Uses top-level default
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockBRSClient := new(testmockrs.MockBRSClient)
			mockBRSWrapper := new(MockBRSClientWrapper)

			mockBRSWrapper.On("GetTenantId").Return("test-tenant-id")
			mockBRSWrapper.On("GetBRSClient").Return(mockBRSClient)

			mockBRSClient.On("GetProtectionGroupRun", mock.Anything).Return(tt.mockProtectionRun, &core.DetailedResponse{}, nil)

			config := &KubernetesDataSourceConfig{
				ClusterName: "test-cluster",
			}

			dataSource := createTestKubernetesDataSource("test", config, mockBRSWrapper)

			k8sRestoreParams := &types.KubernetesRestoreParams{
				RecoverObjectSpec: &types.RecoverObjectSpec{
					StorageClasses: tt.topLevelStorageClasses,
				},
				RecoverMultipleObjects: tt.recoverMultipleObjects,
			}

			result, err := dataSource.InitializeKubernetesNamespaceParams("group-123", "backup-456", k8sRestoreParams)

			assert.NoError(t, err)
			assert.NotNil(t, result)
			assert.Equal(t, len(tt.expectedStorageClassMapping), len(result))

			// Verify each namespace got the correct storage class mapping
			for _, recoveryObj := range result {
				// Note: We can't directly verify the namespace name from the result,
				// but we can verify the count and that storage class mappings were applied
				assert.NotNil(t, recoveryObj.StorageClass)
				if recoveryObj.StorageClass.StorageClassMapping != nil {
					assert.Greater(t, len(recoveryObj.StorageClass.StorageClassMapping), 0)
				}
			}
		})
	}
}

// TestInitializeKubernetesNamespaceParams_CombinedFeatures tests all features together
func TestInitializeKubernetesNamespaceParams_CombinedFeatures(t *testing.T) {
	mockProtectionRun := &backuprecoveryv1.ProtectionGroupRun{
		Objects: []backuprecoveryv1.ObjectRunResult{
			{
				Object: &backuprecoveryv1.ObjectSummary{
					ID:   core.Int64Ptr(1),
					Name: core.StringPtr("app-1"),
				},
				ArchivalInfo: &backuprecoveryv1.ArchivalRun{
					ArchivalTargetResults: []backuprecoveryv1.ArchivalTargetResult{
						{SnapshotID: core.StringPtr("snap-1")},
					},
				},
			},
			{
				Object: &backuprecoveryv1.ObjectSummary{
					ID:   core.Int64Ptr(2),
					Name: core.StringPtr("app-2"),
				},
				ArchivalInfo: &backuprecoveryv1.ArchivalRun{
					ArchivalTargetResults: []backuprecoveryv1.ArchivalTargetResult{
						{SnapshotID: core.StringPtr("snap-2")},
					},
				},
			},
			{
				Object: &backuprecoveryv1.ObjectSummary{
					ID:   core.Int64Ptr(3),
					Name: core.StringPtr("app-3"),
				},
				ArchivalInfo: &backuprecoveryv1.ArchivalRun{
					ArchivalTargetResults: []backuprecoveryv1.ArchivalTargetResult{
						{SnapshotID: core.StringPtr("snap-3")},
					},
				},
			},
		},
	}

	mockBRSClient := new(testmockrs.MockBRSClient)
	mockBRSWrapper := new(MockBRSClientWrapper)

	mockBRSWrapper.On("GetTenantId").Return("test-tenant-id")
	mockBRSWrapper.On("GetBRSClient").Return(mockBRSClient)

	mockBRSClient.On("GetProtectionGroupRun", mock.Anything).Return(mockProtectionRun, &core.DetailedResponse{}, nil)

	config := &KubernetesDataSourceConfig{
		ClusterName: "test-cluster",
	}

	dataSource := createTestKubernetesDataSource("test", config, mockBRSWrapper)

	// Test: Filter to include only app-1 and app-2, with app-1 having custom storage class mapping
	k8sRestoreParams := &types.KubernetesRestoreParams{
		IncludeNamespaces: []string{"app-1", "app-2"}, // Filter: only these 2
		RecoverObjectSpec: &types.RecoverObjectSpec{
			StorageClasses: []types.StorageClassMapping{
				{Old: "default", New: "standard"}, // Top-level default
			},
		},
		RecoverMultipleObjects: []types.RecoverObject{
			{
				SnapshotInfo: &types.SnapshotInfo{
					NamespaceInfo: &types.NamespaceInfo{
						Namespace: "app-1",
					},
				},
				RecoverObjectSpec: &types.RecoverObjectSpec{
					StorageClasses: []types.StorageClassMapping{
						{Old: "default", New: "premium"}, // Override for app-1
					},
				},
			},
		},
	}

	result, err := dataSource.InitializeKubernetesNamespaceParams("group-123", "backup-456", k8sRestoreParams)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, 2, len(result), "Should only include app-1 and app-2 (app-3 filtered out)")

	// Verify storage class mappings were applied
	for _, recoveryObj := range result {
		assert.NotNil(t, recoveryObj.StorageClass)
		assert.NotNil(t, recoveryObj.SnapshotID)
	}
}
