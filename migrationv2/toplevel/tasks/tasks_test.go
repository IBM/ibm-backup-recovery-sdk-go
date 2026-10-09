/*******************************************************************************
 * IBM Confidential
 * OCO Source Materials
 * (C) Copyright IBM Corp. 2026 All Rights Reserved.
 * The source code for this program is not  published or otherwise divested of
 * its trade secrets, irrespective of what has been deposited with
 * the U.S. Copyright Office.
 ******************************************************************************/
package tasks

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/IBM/go-sdk-core/v5/core"
	"github.com/IBM/ibm-backup-recovery-sdk-go/backuprecoveryv1"
	commoncontext "github.com/IBM/ibm-backup-recovery-sdk-go/migrationv2/common/context"
	testmockrs "github.com/IBM/ibm-backup-recovery-sdk-go/migrationv2/common/testing"
	"github.com/IBM/ibm-backup-recovery-sdk-go/migrationv2/common/types"
	"github.com/IBM/ibm-backup-recovery-sdk-go/migrationv2/middlelevel/connectors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// =============================================================================
// NOTE ON TESTING STRATEGY:
//
// The TaskAPI methods make direct calls to BRS client (concrete type), which
// makes it difficult to mock for unit testing. This test suite focuses on:
//
// 1. INPUT VALIDATION - Testing parameter validation logic
// 2. HELPER FUNCTIONS - Testing pure business logic functions
// 3. DATA TRANSFORMATION - Testing buildRetryOptions, buildFullSchedule, etc.
//
// Integration tests (not included here) would test actual BRS API interactions.
// =============================================================================

// Helper function to create a mock SDK config
func createMockConfig() *types.SDKConfig {
	return &types.SDKConfig{
		Region:                 "us-south",
		APIKey:                 "test-api-key",
		BRSInstanceName:        "test-instance",
		ResourceGroupID:        "test-rg",
		Timeout:                30 * time.Minute,
		DefaultPollingInterval: 10 * time.Second,
		DefaultPollingTimeout:  30 * time.Minute,
	}
}

// MockBRSClientWrapper wraps the mock BRS client
type MockBRSClientWrapper struct {
	Client   *testmockrs.MockBRSClient
	TenantID string
	Region   string
	CRN      string
}

func (m *MockBRSClientWrapper) GetTenantId() string {
	return m.TenantID
}

func (m *MockBRSClientWrapper) GetBRSClient() backuprecoveryv1.BRSClientInterface {
	return m.Client
}

func (m *MockBRSClientWrapper) GetRegion() string {
	return m.Region
}

func (m *MockBRSClientWrapper) GetCRN() string {
	return m.CRN
}

// Helper to create mock BRS client wrapper
func createMockBRSClientWrapper() *MockBRSClientWrapper {
	return &MockBRSClientWrapper{
		Client:   &testmockrs.MockBRSClient{},
		TenantID: "test-tenant-id",
		Region:   "us-south",
		CRN:      "crn:test",
	}
}

// Helper function to create a mock SDK config
func createTestMockConfig() *types.SDKConfig {
	return &types.SDKConfig{
		Region:                 "us-south",
		APIKey:                 "test-api-key",
		BRSInstanceName:        "test-instance",
		ResourceGroupID:        "test-rg",
		Timeout:                30 * time.Minute,
		DefaultPollingInterval: 10 * time.Second,
		DefaultPollingTimeout:  30 * time.Minute,
	}
}

type mockDataSource struct {
	createProtectionGroupFunc func(context.Context, int64, *types.ProtectionGroupParams) (*backuprecoveryv1.CreateProtectionGroupOptions, error)
	runRestoreFunc            func(context.Context, string, string, int64, *types.RestoreParams) (*backuprecoveryv1.CreateRecoveryOptions, error)
}

func (m *mockDataSource) GetType() types.DataSourceType {
	return types.DataSourceKubernetes
}

func (m *mockDataSource) GetName() string {
	return "mock-datasource"
}

func (m *mockDataSource) Validate() error {
	return nil
}

func (m *mockDataSource) RegisterSourceParams(ctx context.Context, connectionID string) (*backuprecoveryv1.RegisterProtectionSourceOptions, error) {
	// Return a valid RegisterProtectionSourceOptions object for testing
	return &backuprecoveryv1.RegisterProtectionSourceOptions{
		Environment: core.StringPtr("kKubernetes"),
	}, nil
}

func (m *mockDataSource) CreateProtectionGroup(ctx context.Context, registrationID int64, params *types.ProtectionGroupParams) (*backuprecoveryv1.CreateProtectionGroupOptions, error) {
	if m.createProtectionGroupFunc != nil {
		return m.createProtectionGroupFunc(ctx, registrationID, params)
	}
	return &backuprecoveryv1.CreateProtectionGroupOptions{}, nil
}

func (m *mockDataSource) RunBackup(ctx context.Context, groupID string) (*types.BackupResult, error) {
	return nil, nil
}

func (m *mockDataSource) GetBackupStatus(ctx context.Context, backupID string) (interface{}, error) {
	return nil, nil
}

func (m *mockDataSource) RunRestore(ctx context.Context, groupID, backupID string, targetRegistrationID int64, restoreParams *types.RestoreParams) (*backuprecoveryv1.CreateRecoveryOptions, error) {
	if m.runRestoreFunc != nil {
		return m.runRestoreFunc(ctx, groupID, backupID, targetRegistrationID, restoreParams)
	}
	return &backuprecoveryv1.CreateRecoveryOptions{}, nil
}


func minimalProtectionPolicyResponse(id, name string, retention int64) *backuprecoveryv1.ProtectionPolicyResponse {
	return &backuprecoveryv1.ProtectionPolicyResponse{
		ID:   core.StringPtr(id),
		Name: core.StringPtr(name),
		BackupPolicy: &backuprecoveryv1.BackupPolicy{
			Regular: &backuprecoveryv1.RegularBackupPolicy{
				Retention: &backuprecoveryv1.Retention{
					Duration: core.Int64Ptr(retention),
				},
			},
		},
	}
}

func minimalProtectionGroupResponse(id, name string) *backuprecoveryv1.ProtectionGroupResponse {
	return &backuprecoveryv1.ProtectionGroupResponse{
		ID:   core.StringPtr(id),
		Name: core.StringPtr(name),
	}
}

func minimalRun(id, groupID, status string) backuprecoveryv1.ProtectionGroupRun {
	return backuprecoveryv1.ProtectionGroupRun{
		ID:                core.StringPtr(id),
		ProtectionGroupID: core.StringPtr(groupID),
		ArchivalInfo: &backuprecoveryv1.ArchivalRunSummary{
			ArchivalTargetResults: []backuprecoveryv1.ArchivalTargetResult{
				{
					Status: core.StringPtr(status),
				},
			},
		},
	}
}

func TestCreateProtectionPolicy(t *testing.T) {
	incUnit := types.IncrementalBackup_Unit_Days
	retentionUnit := types.Retention_Unit_Days
	fullUnit := types.FullSchedule_Unit_Days

	tests := []struct {
		name           string
		policyParams   *types.PolicyParams
		mockSetup      func(*MockBRSClientWrapper)
		expectError    bool
		errorContains  string
		validateResult func(*testing.T, *types.PolicyResult)
	}{
		{
			name:          "NilParams",
			policyParams:  nil,
			mockSetup:     func(mockWrapper *MockBRSClientWrapper) {},
			expectError:   true,
			errorContains: "[INVALID_INPUT] policyParams.Name is required",
		},
		{
			name: "EmptyName",
			policyParams: &types.PolicyParams{
				Name: "",
			},
			mockSetup:     func(mockWrapper *MockBRSClientWrapper) {},
			expectError:   true,
			errorContains: "[INVALID_INPUT] policyParams.Name is required",
		},
		{
			name: "ExistingPolicy_Idempotent",
			policyParams: &types.PolicyParams{
				Name: "test-policy",
			},
			mockSetup: func(mockWrapper *MockBRSClientWrapper) {
				mockWrapper.Client.On("GetProtectionPoliciesWithContext", mock.Anything, mock.Anything).Return(
					&backuprecoveryv1.ProtectionPoliciesResponse{Policies: []backuprecoveryv1.ProtectionPolicyResponse{
						{
							ID:   core.StringPtr("existing-policy-id"),
							Name: core.StringPtr("test-policy"),
							BackupPolicy: &backuprecoveryv1.BackupPolicy{
								Regular: &backuprecoveryv1.RegularBackupPolicy{
									Retention: &backuprecoveryv1.Retention{
										Duration: core.Int64Ptr(30),
									},
								},
							},
						},
					}}, nil, nil)
			},
			expectError: false,
			validateResult: func(t *testing.T, result *types.PolicyResult) {
				assert.NotNil(t, result)
				assert.Equal(t, "existing-policy-id", result.ID)
				assert.Equal(t, "test-policy", result.Name)
			},
		},
		{
			name: "GetPolicyByNameFails",
			policyParams: &types.PolicyParams{
				Name: "test-policy",
			},
			mockSetup: func(mockWrapper *MockBRSClientWrapper) {
				mockWrapper.Client.On("GetProtectionPoliciesWithContext", mock.Anything, mock.Anything).Return(
					nil, nil, fmt.Errorf("API error"))
			},
			expectError:   true,
			errorContains: "[PROTECTION_FAILED]",
		},
		{
			name: "IncrementalBackup_MissingUnit",
			policyParams: &types.PolicyParams{
				Name: "test-policy",
				IncrementalBackup: &types.IncrementalBackup{
					Unit: nil,
				},
			},
			mockSetup: func(mockWrapper *MockBRSClientWrapper) {
				mockWrapper.Client.On("GetProtectionPoliciesWithContext", mock.Anything, mock.Anything).Return(
					&backuprecoveryv1.ProtectionPoliciesResponse{Policies: []backuprecoveryv1.ProtectionPolicyResponse{}}, nil, nil)
			},
			expectError:   true,
			errorContains: "[UNKNOWN]",
		},
		{
			name: "IncrementalBackup_MissingDaySchedule",
			policyParams: &types.PolicyParams{
				Name: "test-policy",
				IncrementalBackup: &types.IncrementalBackup{
					Unit:        &incUnit,
					DaySchedule: nil,
				},
			},
			mockSetup: func(mockWrapper *MockBRSClientWrapper) {
				mockWrapper.Client.On("GetProtectionPoliciesWithContext", mock.Anything, mock.Anything).Return(
					&backuprecoveryv1.ProtectionPoliciesResponse{Policies: []backuprecoveryv1.ProtectionPolicyResponse{}}, nil, nil)
			},
			expectError:   true,
			errorContains: "[UNKNOWN]",
		},
		{
			name: "DataRetention_MissingUnit",
			policyParams: &types.PolicyParams{
				Name: "test-policy",
				DataRetention: &types.DataRetention{
					Unit:      nil,
					RetainFor: 30,
				},
			},
			mockSetup: func(mockWrapper *MockBRSClientWrapper) {
				mockWrapper.Client.On("GetProtectionPoliciesWithContext", mock.Anything, mock.Anything).Return(
					&backuprecoveryv1.ProtectionPoliciesResponse{Policies: []backuprecoveryv1.ProtectionPolicyResponse{}}, nil, nil)
			},
			expectError:   true,
			errorContains: "[PROTECTION_FAILED]",
		},
		{
			name: "DataRetention_MissingRetainFor",
			policyParams: &types.PolicyParams{
				Name: "test-policy",
				DataRetention: &types.DataRetention{
					Unit:      &retentionUnit,
					RetainFor: 0,
				},
			},
			mockSetup: func(mockWrapper *MockBRSClientWrapper) {
				mockWrapper.Client.On("GetProtectionPoliciesWithContext", mock.Anything, mock.Anything).Return(
					&backuprecoveryv1.ProtectionPoliciesResponse{Policies: []backuprecoveryv1.ProtectionPolicyResponse{}}, nil, nil)
			},
			expectError:   true,
			errorContains: "[PROTECTION_FAILED]",
		},
		{
			name: "FullBackups_MissingUnit",
			policyParams: &types.PolicyParams{
				Name: "test-policy",
				FullBackups: []types.FullBackup{
					{Unit: nil},
				},
			},
			mockSetup: func(mockWrapper *MockBRSClientWrapper) {
				mockWrapper.Client.On("GetProtectionPoliciesWithContext", mock.Anything, mock.Anything).Return(
					&backuprecoveryv1.ProtectionPoliciesResponse{Policies: []backuprecoveryv1.ProtectionPolicyResponse{}}, nil, nil)
			},
			expectError:   true,
			errorContains: "[UNKNOWN]",
		},
		{
			name: "FullBackups_MissingRetention",
			policyParams: &types.PolicyParams{
				Name: "test-policy",
				FullBackups: []types.FullBackup{
					{
						Unit:      &fullUnit,
						Retention: nil,
					},
				},
			},
			mockSetup: func(mockWrapper *MockBRSClientWrapper) {
				mockWrapper.Client.On("GetProtectionPoliciesWithContext", mock.Anything, mock.Anything).Return(
					&backuprecoveryv1.ProtectionPoliciesResponse{Policies: []backuprecoveryv1.ProtectionPolicyResponse{}}, nil, nil)
			},
			expectError:   true,
			errorContains: "[UNKNOWN]",
		},
		{
			name: "CreatePolicyAPIFails",
			policyParams: &types.PolicyParams{
				Name: "test-policy",
			},
			mockSetup: func(mockWrapper *MockBRSClientWrapper) {
				mockWrapper.Client.On("GetProtectionPoliciesWithContext", mock.Anything, mock.Anything).Return(
					&backuprecoveryv1.ProtectionPoliciesResponse{Policies: []backuprecoveryv1.ProtectionPolicyResponse{}}, nil, nil)
				mockWrapper.Client.On("CreateProtectionPolicyWithContext", mock.Anything, mock.Anything).Return(
					nil, nil, fmt.Errorf("API error"))
			},
			expectError:   true,
			errorContains: "[UNKNOWN]",
		},
		{
			name: "Success_MinimalParams",
			policyParams: &types.PolicyParams{
				Name: "test-policy",
			},
			mockSetup: func(mockWrapper *MockBRSClientWrapper) {
				mockWrapper.Client.On("GetProtectionPoliciesWithContext", mock.Anything, mock.Anything).Return(
					&backuprecoveryv1.ProtectionPoliciesResponse{Policies: []backuprecoveryv1.ProtectionPolicyResponse{}}, nil, nil)
				mockWrapper.Client.On("CreateProtectionPolicyWithContext", mock.Anything, mock.MatchedBy(
					func(opts *backuprecoveryv1.CreateProtectionPolicyOptions) bool {
						return *opts.XIBMTenantID == "test-tenant-id" && *opts.Name == "test-policy"
					})).Return(&backuprecoveryv1.ProtectionPolicyResponse{
					ID:   core.StringPtr("new-policy-id"),
					Name: core.StringPtr("test-policy"),
				}, nil, nil)
			},
			expectError: false,
			validateResult: func(t *testing.T, result *types.PolicyResult) {
				assert.NotNil(t, result)
				assert.Equal(t, "new-policy-id", result.ID)
				assert.Equal(t, "test-policy", result.Name)
			},
		},
		{
			name: "Success_WithIncrementalBackup",
			policyParams: &types.PolicyParams{
				Name: "test-policy",
				IncrementalBackup: &types.IncrementalBackup{
					Unit: &incUnit,
					DaySchedule: &types.UnitDaySchedule{
						Every: 1,
					},
				},
			},
			mockSetup: func(mockWrapper *MockBRSClientWrapper) {
				mockWrapper.Client.On("GetProtectionPoliciesWithContext", mock.Anything, mock.Anything).Return(
					&backuprecoveryv1.ProtectionPoliciesResponse{Policies: []backuprecoveryv1.ProtectionPolicyResponse{}}, nil, nil)
				mockWrapper.Client.On("CreateProtectionPolicyWithContext", mock.Anything, mock.MatchedBy(
					func(opts *backuprecoveryv1.CreateProtectionPolicyOptions) bool {
						return opts.BackupPolicy != nil && opts.BackupPolicy.Regular != nil && opts.BackupPolicy.Regular.Incremental != nil
					})).Return(&backuprecoveryv1.ProtectionPolicyResponse{
					ID:   core.StringPtr("new-policy-id"),
					Name: core.StringPtr("test-policy"),
				}, nil, nil)
			},
			expectError: false,
			validateResult: func(t *testing.T, result *types.PolicyResult) {
				assert.NotNil(t, result)
				assert.Equal(t, "new-policy-id", result.ID)
			},
		},
		{
			name: "Success_WithAllParameters",
			policyParams: &types.PolicyParams{
				Name: "test-policy",
				IncrementalBackup: &types.IncrementalBackup{
					Unit: &incUnit,
					DaySchedule: &types.UnitDaySchedule{
						Every: 1,
					},
				},
				DataRetention: &types.DataRetention{
					Unit:      &retentionUnit,
					RetainFor: 30,
				},
				PrimaryBackupTarget: &types.PrimaryBackupTarget{
					UseDefaultBackupTarget: true,
				},
				FullBackups: []types.FullBackup{
					{
						Unit: &fullUnit,
						DaySchedule: &types.UnitDaySchedule{
							Every: 7,
						},
						Retention: &types.DataRetention{
							Unit:      &retentionUnit,
							RetainFor: 90,
						},
					},
				},
				RetryOption: &types.RetryOption{
					NumberOfRetry:       3,
					WaitInCaseOfFailure: 5,
				},
			},
			mockSetup: func(mockWrapper *MockBRSClientWrapper) {
				mockWrapper.Client.On("GetProtectionPoliciesWithContext", mock.Anything, mock.Anything).Return(
					&backuprecoveryv1.ProtectionPoliciesResponse{Policies: []backuprecoveryv1.ProtectionPolicyResponse{}}, nil, nil)
				mockWrapper.Client.On("CreateProtectionPolicyWithContext", mock.Anything, mock.MatchedBy(
					func(opts *backuprecoveryv1.CreateProtectionPolicyOptions) bool {
						return *opts.Name == "test-policy" &&
							opts.BackupPolicy.Regular.Incremental != nil &&
							opts.BackupPolicy.Regular.Retention != nil &&
							opts.BackupPolicy.Regular.PrimaryBackupTarget != nil &&
							opts.BackupPolicy.Regular.FullBackups != nil &&
							opts.RetryOptions != nil &&
							opts.IsCBSEnabled != nil
					})).Return(&backuprecoveryv1.ProtectionPolicyResponse{
					ID:   core.StringPtr("new-policy-id"),
					Name: core.StringPtr("test-policy"),
				}, nil, nil)
			},
			expectError: false,
			validateResult: func(t *testing.T, result *types.PolicyResult) {
				assert.NotNil(t, result)
				assert.Equal(t, "new-policy-id", result.ID)
				assert.Equal(t, "test-policy", result.Name)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockWrapper := createMockBRSClientWrapper()
			tt.mockSetup(mockWrapper)
			taskAPI := NewTaskAPI(mockWrapper, createTestMockConfig())
			ctx := context.Background()

			result, err := taskAPI.CreatePolicy(ctx, tt.policyParams)

			if tt.expectError {
				assert.NotNil(t, err)
				assert.Nil(t, result)
				if tt.errorContains != "" {
					assert.Contains(t, err.Error(), tt.errorContains)
				}
			} else {
				assert.Nil(t, err)
				if tt.validateResult != nil {
					tt.validateResult(t, result)
				}
			}
			mockWrapper.Client.AssertExpectations(t)
		})
	}
}

func TestCreateProtectionGroup(t *testing.T) {
	tests := []struct {
		name           string
		registrationID int64
		groupParams    *types.ProtectionGroupParams
		dataSource     *mockDataSource
		mockSetup      func(*MockBRSClientWrapper)
		expectError    bool
		errorContains  string
		validateResult func(*testing.T, *types.ProtectionGroupResult)
	}{
		{
			name:           "ValidationError_NilParams",
			registrationID: 99,
			groupParams:    nil,
			dataSource:     &mockDataSource{},
			mockSetup:      func(mockWrapper *MockBRSClientWrapper) {},
			expectError:    true,
			errorContains:  "groupParams.Name is required",
		},
		{
			name:           "ExistingGroup_Idempotent",
			registrationID: 99,
			groupParams: &types.ProtectionGroupParams{
				Name:   "pg-existing",
				Policy: &types.Policy{ID: "policy-1"},
			},
			dataSource: &mockDataSource{},
			mockSetup: func(mockWrapper *MockBRSClientWrapper) {
				mockWrapper.Client.On("GetProtectionGroupsWithContext", mock.Anything, mock.MatchedBy(
					func(opts *backuprecoveryv1.GetProtectionGroupsOptions) bool {
						return len(opts.Names) == 1 && opts.Names[0] == "pg-existing"
					})).Return(&backuprecoveryv1.ProtectionGroupsResponse{
					ProtectionGroups: []backuprecoveryv1.ProtectionGroupResponse{
						*minimalProtectionGroupResponse("pg-1", "pg-existing"),
					},
				}, nil, nil)
			},
			expectError: false,
			validateResult: func(t *testing.T, result *types.ProtectionGroupResult) {
				assert.NotNil(t, result)
				assert.Equal(t, "pg-1", result.ProtectionGroupID)
				assert.Equal(t, "pg-existing", result.GroupName)
			},
		},
		{
			name:           "LookupByNameFails",
			registrationID: 99,
			groupParams: &types.ProtectionGroupParams{
				Name:   "pg-new",
				Policy: &types.Policy{ID: "policy-1"},
			},
			dataSource: &mockDataSource{},
			mockSetup: func(mockWrapper *MockBRSClientWrapper) {
				mockWrapper.Client.On("GetProtectionGroupsWithContext", mock.Anything, mock.Anything).Return(
					nil, nil, fmt.Errorf("lookup failed"))
			},
			expectError:   true,
			errorContains: "failed to fetch protection group by name",
		},
		{
			name:           "DataSourceBuildFails",
			registrationID: 99,
			groupParams: &types.ProtectionGroupParams{
				Name:   "pg-new",
				Policy: &types.Policy{ID: "policy-1"},
			},
			dataSource: &mockDataSource{
				createProtectionGroupFunc: func(context.Context, int64, *types.ProtectionGroupParams) (*backuprecoveryv1.CreateProtectionGroupOptions, error) {
					return nil, fmt.Errorf("datasource build failed")
				},
			},
			mockSetup: func(mockWrapper *MockBRSClientWrapper) {
				mockWrapper.Client.On("GetProtectionGroupsWithContext", mock.Anything, mock.Anything).Return(
					&backuprecoveryv1.ProtectionGroupsResponse{}, nil, nil)
			},
			expectError:   true,
			errorContains: "CreateProtectionGroup BRS API call failed",
		},
		{
			name:           "CreateAPIFails",
			registrationID: 99,
			groupParams: &types.ProtectionGroupParams{
				Name:   "pg-new",
				Policy: &types.Policy{ID: "policy-1"},
			},
			dataSource: &mockDataSource{},
			mockSetup: func(mockWrapper *MockBRSClientWrapper) {
				mockWrapper.Client.On("GetProtectionGroupsWithContext", mock.Anything, mock.Anything).Return(
					&backuprecoveryv1.ProtectionGroupsResponse{}, nil, nil)
				mockWrapper.Client.On("CreateProtectionGroupWithContext", mock.Anything, mock.MatchedBy(
					func(opts *backuprecoveryv1.CreateProtectionGroupOptions) bool {
						return *opts.XIBMTenantID == "test-tenant-id" &&
							*opts.PolicyID == "policy-1" &&
							*opts.Name == "pg-new"
					})).Return(nil, nil, fmt.Errorf("create failed"))
			},
			expectError:   true,
			errorContains: "failed to create protection group",
		},
		{
			name:           "Success",
			registrationID: 1234,
			groupParams: &types.ProtectionGroupParams{
				Name:   "pg-new",
				Policy: &types.Policy{ID: "policy-1"},
			},
			dataSource: &mockDataSource{},
			mockSetup: func(mockWrapper *MockBRSClientWrapper) {
				mockWrapper.Client.On("GetProtectionGroupsWithContext", mock.Anything, mock.Anything).Return(
					&backuprecoveryv1.ProtectionGroupsResponse{}, nil, nil)
				mockWrapper.Client.On("CreateProtectionGroupWithContext", mock.Anything, mock.Anything).Return(
					minimalProtectionGroupResponse("pg-1", "pg-new"), nil, nil)
			},
			expectError: false,
			validateResult: func(t *testing.T, result *types.ProtectionGroupResult) {
				assert.NotNil(t, result)
				assert.Equal(t, "pg-1", result.ProtectionGroupID)
				assert.Equal(t, "pg-new", result.GroupName)
				assert.Equal(t, int64(1234), result.RegistrationID)
				assert.Equal(t, "policy-1", result.PolicyID)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockWrapper := createMockBRSClientWrapper()
			tt.mockSetup(mockWrapper)
			taskAPI := NewTaskAPI(mockWrapper, createTestMockConfig())
			ctx := context.Background()

			result, err := taskAPI.CreateProtectionGroup(ctx, tt.registrationID, tt.groupParams, tt.dataSource)

			if tt.expectError {
				assert.NotNil(t, err)
				assert.Nil(t, result)
				if tt.errorContains != "" {
					assert.Contains(t, err.Error(), tt.errorContains)
				}
			} else {
				assert.Nil(t, err)
				if tt.validateResult != nil {
					tt.validateResult(t, result)
				}
			}
			mockWrapper.Client.AssertExpectations(t)
		})
	}
}

func TestGetProtectionGroup(t *testing.T) {
	tests := []struct {
		name           string
		groupID        string
		mockSetup      func(*MockBRSClientWrapper)
		expectError    bool
		errorContains  string
		validateResult func(*testing.T, *types.ProtectionGroupResult)
	}{
		{
			name:    "APIError",
			groupID: "group-1",
			mockSetup: func(mockWrapper *MockBRSClientWrapper) {
				mockWrapper.Client.On("GetProtectionGroupByIDWithContext", mock.Anything, mock.MatchedBy(
					func(opts *backuprecoveryv1.GetProtectionGroupByIdOptions) bool {
						return *opts.ID == "group-1"
					})).Return(nil, nil, fmt.Errorf("get failed"))
			},
			expectError:   true,
			errorContains: "failed to fetch protection group by Id",
		},
		{
			name:    "Success",
			groupID: "group-1",
			mockSetup: func(mockWrapper *MockBRSClientWrapper) {
				mockWrapper.Client.On("GetProtectionGroupByIDWithContext", mock.Anything, mock.Anything).Return(
					minimalProtectionGroupResponse("group-1", "group-name"), nil, nil)
			},
			expectError: false,
			validateResult: func(t *testing.T, result *types.ProtectionGroupResult) {
				assert.NotNil(t, result)
				assert.Equal(t, "group-1", result.ProtectionGroupID)
				assert.Equal(t, "group-name", result.GroupName)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockWrapper := createMockBRSClientWrapper()
			tt.mockSetup(mockWrapper)
			taskAPI := NewTaskAPI(mockWrapper, createTestMockConfig())
			ctx := context.Background()

			result, err := taskAPI.GetProtectionGroup(ctx, tt.groupID)

			if tt.expectError {
				assert.NotNil(t, err)
				assert.Nil(t, result)
				if tt.errorContains != "" {
					assert.Contains(t, err.Error(), tt.errorContains)
				}
			} else {
				assert.Nil(t, err)
				if tt.validateResult != nil {
					tt.validateResult(t, result)
				}
			}
			mockWrapper.Client.AssertExpectations(t)
		})
	}
}

func TestGetProtectionGroupByName(t *testing.T) {
	tests := []struct {
		name           string
		groupName      string
		mockSetup      func(*MockBRSClientWrapper)
		expectError    bool
		errorContains  string
		validateResult func(*testing.T, *types.ProtectionGroupResult)
	}{
		{
			name:      "APIError",
			groupName: "group-name",
			mockSetup: func(mockWrapper *MockBRSClientWrapper) {
				mockWrapper.Client.On("GetProtectionGroupsWithContext", mock.Anything, mock.Anything).Return(
					nil, nil, fmt.Errorf("list failed"))
			},
			expectError:   true,
			errorContains: "failed to fetch protection group by name",
		},
		{
			name:      "NotFound",
			groupName: "group-name",
			mockSetup: func(mockWrapper *MockBRSClientWrapper) {
				mockWrapper.Client.On("GetProtectionGroupsWithContext", mock.Anything, mock.Anything).Return(
					&backuprecoveryv1.ProtectionGroupsResponse{
						ProtectionGroups: []backuprecoveryv1.ProtectionGroupResponse{},
					}, nil, nil)
			},
			expectError: false,
			validateResult: func(t *testing.T, result *types.ProtectionGroupResult) {
				assert.Nil(t, result)
			},
		},
		{
			name:      "Success",
			groupName: "group-name",
			mockSetup: func(mockWrapper *MockBRSClientWrapper) {
				mockWrapper.Client.On("GetProtectionGroupsWithContext", mock.Anything, mock.MatchedBy(
					func(opts *backuprecoveryv1.GetProtectionGroupsOptions) bool {
						return len(opts.Names) == 1 && opts.Names[0] == "group-name"
					})).Return(&backuprecoveryv1.ProtectionGroupsResponse{
					ProtectionGroups: []backuprecoveryv1.ProtectionGroupResponse{
						*minimalProtectionGroupResponse("group-1", "group-name"),
					},
				}, nil, nil)
			},
			expectError: false,
			validateResult: func(t *testing.T, result *types.ProtectionGroupResult) {
				assert.NotNil(t, result)
				assert.Equal(t, "group-1", result.ProtectionGroupID)
				assert.Equal(t, "group-name", result.GroupName)
				assert.Equal(t, "active", result.Status)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockWrapper := createMockBRSClientWrapper()
			tt.mockSetup(mockWrapper)
			taskAPI := NewTaskAPI(mockWrapper, createTestMockConfig())
			ctx := context.Background()

			result, err := taskAPI.GetProtectionGroupByName(ctx, tt.groupName)

			if tt.expectError {
				assert.NotNil(t, err)
				assert.Nil(t, result)
				if tt.errorContains != "" {
					assert.Contains(t, err.Error(), tt.errorContains)
				}
			} else {
				assert.Nil(t, err)
				if tt.validateResult != nil {
					tt.validateResult(t, result)
				}
			}
			mockWrapper.Client.AssertExpectations(t)
		})
	}
}

func TestListProtectionGroups(t *testing.T) {
	tests := []struct {
		name           string
		registrationID int64
		mockSetup      func(*MockBRSClientWrapper)
		expectError    bool
		errorContains  string
		validateResult func(*testing.T, []*types.ProtectionGroupResult)
	}{
		{
			name:           "APIError",
			registrationID: 55,
			mockSetup: func(mockWrapper *MockBRSClientWrapper) {
				mockWrapper.Client.On("GetProtectionGroupsWithContext", mock.Anything, mock.Anything).Return(
					nil, nil, fmt.Errorf("list failed"))
			},
			expectError:   true,
			errorContains: "failed to fetch list of protection groups",
		},
		{
			name:           "Success",
			registrationID: 55,
			mockSetup: func(mockWrapper *MockBRSClientWrapper) {
				mockWrapper.Client.On("GetProtectionGroupsWithContext", mock.Anything, mock.Anything).Return(
					&backuprecoveryv1.ProtectionGroupsResponse{
						ProtectionGroups: []backuprecoveryv1.ProtectionGroupResponse{
							*minimalProtectionGroupResponse("group-1", "group-one"),
							*minimalProtectionGroupResponse("group-2", "group-two"),
						},
					}, nil, nil)
			},
			expectError: false,
			validateResult: func(t *testing.T, results []*types.ProtectionGroupResult) {
				assert.NotNil(t, results)
				assert.Len(t, results, 2)
				assert.Equal(t, "group-1", results[0].ProtectionGroupID)
				assert.Equal(t, "group-one", results[0].GroupName)
				assert.Equal(t, "group-2", results[1].ProtectionGroupID)
				assert.Equal(t, "group-two", results[1].GroupName)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockWrapper := createMockBRSClientWrapper()
			tt.mockSetup(mockWrapper)
			taskAPI := NewTaskAPI(mockWrapper, createTestMockConfig())
			ctx := context.Background()

			result, err := taskAPI.ListProtectionGroups(ctx, tt.registrationID)

			if tt.expectError {
				assert.NotNil(t, err)
				assert.Nil(t, result)
				if tt.errorContains != "" {
					assert.Contains(t, err.Error(), tt.errorContains)
				}
			} else {
				assert.Nil(t, err)
				if tt.validateResult != nil {
					tt.validateResult(t, result)
				}
			}
			mockWrapper.Client.AssertExpectations(t)
		})
	}
}

func TestDeleteProtectionGroup(t *testing.T) {
	tests := []struct {
		name          string
		groupID       string
		mockSetup     func(*MockBRSClientWrapper)
		expectError   bool
		errorContains string
	}{
		{
			name:    "APIError",
			groupID: "group-1",
			mockSetup: func(mockWrapper *MockBRSClientWrapper) {
				mockWrapper.Client.On("DeleteProtectionGroupWithContext", mock.Anything, mock.MatchedBy(
					func(opts *backuprecoveryv1.DeleteProtectionGroupOptions) bool {
						return *opts.ID == "group-1"
					})).Return(nil, fmt.Errorf("delete failed"))
			},
			expectError:   true,
			errorContains: "failed to delete protection group by Id",
		},
		{
			name:    "Success",
			groupID: "group-1",
			mockSetup: func(mockWrapper *MockBRSClientWrapper) {
				mockWrapper.Client.On("DeleteProtectionGroupWithContext", mock.Anything, mock.Anything).Return(nil, nil)
			},
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockWrapper := createMockBRSClientWrapper()
			tt.mockSetup(mockWrapper)
			taskAPI := NewTaskAPI(mockWrapper, createTestMockConfig())
			ctx := context.Background()

			err := taskAPI.DeleteProtectionGroup(ctx, tt.groupID, false)

			if tt.expectError {
				assert.NotNil(t, err)
				if tt.errorContains != "" {
					assert.Contains(t, err.Error(), tt.errorContains)
				}
			} else {
				assert.Nil(t, err)
			}
			mockWrapper.Client.AssertExpectations(t)
		})
	}
}

func TestGetPolicy(t *testing.T) {
	tests := []struct {
		name           string
		policyID       string
		mockSetup      func(*MockBRSClientWrapper)
		expectError    bool
		errorContains  string
		validateResult func(*testing.T, *types.PolicyResult)
	}{
		{
			name:     "APIError",
			policyID: "policy-1",
			mockSetup: func(mockWrapper *MockBRSClientWrapper) {
				mockWrapper.Client.On("GetProtectionPolicyByIDWithContext", mock.Anything, mock.MatchedBy(
					func(opts *backuprecoveryv1.GetProtectionPolicyByIdOptions) bool {
						return *opts.ID == "policy-1"
					})).Return(nil, nil, fmt.Errorf("get failed"))
			},
			expectError:   true,
			errorContains: "failed to fetch protection policy by id",
		},
		{
			name:     "Success",
			policyID: "policy-1",
			mockSetup: func(mockWrapper *MockBRSClientWrapper) {
				mockWrapper.Client.On("GetProtectionPolicyByIDWithContext", mock.Anything, mock.Anything).Return(
					minimalProtectionPolicyResponse("policy-1", "policy-name", 30), nil, nil)
			},
			expectError: false,
			validateResult: func(t *testing.T, result *types.PolicyResult) {
				assert.NotNil(t, result)
				assert.Equal(t, "policy-1", result.ID)
				assert.Equal(t, "policy-name", result.Name)
				assert.Equal(t, int64(30), result.RetentionDays)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockWrapper := createMockBRSClientWrapper()
			tt.mockSetup(mockWrapper)
			taskAPI := NewTaskAPI(mockWrapper, createTestMockConfig())
			ctx := context.Background()

			result, err := taskAPI.GetPolicy(ctx, tt.policyID)

			if tt.expectError {
				assert.NotNil(t, err)
				assert.Nil(t, result)
				if tt.errorContains != "" {
					assert.Contains(t, err.Error(), tt.errorContains)
				}
			} else {
				assert.Nil(t, err)
				if tt.validateResult != nil {
					tt.validateResult(t, result)
				}
			}
			mockWrapper.Client.AssertExpectations(t)
		})
	}
}

func TestGetPolicyByName(t *testing.T) {
	tests := []struct {
		name           string
		policyName     string
		mockSetup      func(*MockBRSClientWrapper)
		expectError    bool
		errorContains  string
		validateResult func(*testing.T, *types.PolicyResult)
	}{
		{
			name:       "APIError",
			policyName: "policy-name",
			mockSetup: func(mockWrapper *MockBRSClientWrapper) {
				mockWrapper.Client.On("GetProtectionPoliciesWithContext", mock.Anything, mock.Anything).Return(
					nil, nil, fmt.Errorf("list failed"))
			},
			expectError:   true,
			errorContains: "failed to fetch protection policy by name",
		},
		{
			name:       "NotFound",
			policyName: "policy-name",
			mockSetup: func(mockWrapper *MockBRSClientWrapper) {
				mockWrapper.Client.On("GetProtectionPoliciesWithContext", mock.Anything, mock.Anything).Return(
					&backuprecoveryv1.ProtectionPoliciesResponse{
						Policies: []backuprecoveryv1.ProtectionPolicyResponse{},
					}, nil, nil)
			},
			expectError: false,
			validateResult: func(t *testing.T, result *types.PolicyResult) {
				assert.Nil(t, result)
			},
		},
		{
			name:       "Success",
			policyName: "policy-name",
			mockSetup: func(mockWrapper *MockBRSClientWrapper) {
				mockWrapper.Client.On("GetProtectionPoliciesWithContext", mock.Anything, mock.MatchedBy(
					func(opts *backuprecoveryv1.GetProtectionPoliciesOptions) bool {
						return len(opts.PolicyNames) == 1 && opts.PolicyNames[0] == "policy-name"
					})).Return(&backuprecoveryv1.ProtectionPoliciesResponse{
					Policies: []backuprecoveryv1.ProtectionPolicyResponse{
						*minimalProtectionPolicyResponse("policy-1", "policy-name", 45),
					},
				}, nil, nil)
			},
			expectError: false,
			validateResult: func(t *testing.T, result *types.PolicyResult) {
				assert.NotNil(t, result)
				assert.Equal(t, "policy-1", result.ID)
				assert.Equal(t, "policy-name", result.Name)
				assert.Equal(t, int64(45), result.RetentionDays)
				assert.Equal(t, "active", result.Status)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockWrapper := createMockBRSClientWrapper()
			tt.mockSetup(mockWrapper)
			taskAPI := NewTaskAPI(mockWrapper, createTestMockConfig())
			ctx := context.Background()

			result, err := taskAPI.GetPolicyByName(ctx, tt.policyName)

			if tt.expectError {
				assert.NotNil(t, err)
				assert.Nil(t, result)
				if tt.errorContains != "" {
					assert.Contains(t, err.Error(), tt.errorContains)
				}
			} else {
				assert.Nil(t, err)
				if tt.validateResult != nil {
					tt.validateResult(t, result)
				}
			}
			mockWrapper.Client.AssertExpectations(t)
		})
	}
}

func TestListPolicies(t *testing.T) {
	tests := []struct {
		name           string
		mockSetup      func(*MockBRSClientWrapper)
		expectError    bool
		errorContains  string
		validateResult func(*testing.T, []*types.PolicyResult)
	}{
		{
			name: "APIError",
			mockSetup: func(mockWrapper *MockBRSClientWrapper) {
				mockWrapper.Client.On("GetProtectionPoliciesWithContext", mock.Anything, mock.Anything).Return(
					nil, nil, fmt.Errorf("list failed"))
			},
			expectError:   true,
			errorContains: "failed to fetch protection policies",
		},
		{
			name: "Success_MultiplePolicies",
			mockSetup: func(mockWrapper *MockBRSClientWrapper) {
				mockWrapper.Client.On("GetProtectionPoliciesWithContext", mock.Anything, mock.Anything).Return(
					&backuprecoveryv1.ProtectionPoliciesResponse{
						Policies: []backuprecoveryv1.ProtectionPolicyResponse{
							*minimalProtectionPolicyResponse("policy-1", "policy-one", 7),
							*minimalProtectionPolicyResponse("policy-2", "policy-two", 14),
						},
					}, nil, nil)
			},
			expectError: false,
			validateResult: func(t *testing.T, results []*types.PolicyResult) {
				assert.NotNil(t, results)
				assert.Len(t, results, 2)
				assert.Equal(t, "policy-1", results[0].ID)
				assert.Equal(t, "policy-one", results[0].Name)
				assert.Equal(t, int64(7), results[0].RetentionDays)
				assert.Equal(t, "policy-2", results[1].ID)
				assert.Equal(t, "policy-two", results[1].Name)
				assert.Equal(t, int64(14), results[1].RetentionDays)
			},
		},
		{
			name: "EmptyList_NilResponse",
			mockSetup: func(mockWrapper *MockBRSClientWrapper) {
				mockWrapper.Client.On("GetProtectionPoliciesWithContext", mock.Anything, mock.Anything).Return(
					nil, nil, nil)
			},
			expectError: false,
			validateResult: func(t *testing.T, results []*types.PolicyResult) {
				assert.NotNil(t, results)
				assert.Len(t, results, 0)
			},
		},
		{
			name: "EmptyList_EmptyArray",
			mockSetup: func(mockWrapper *MockBRSClientWrapper) {
				mockWrapper.Client.On("GetProtectionPoliciesWithContext", mock.Anything, mock.Anything).Return(
					&backuprecoveryv1.ProtectionPoliciesResponse{
						Policies: []backuprecoveryv1.ProtectionPolicyResponse{},
					}, nil, nil)
			},
			expectError: false,
			validateResult: func(t *testing.T, results []*types.PolicyResult) {
				assert.NotNil(t, results)
				assert.Len(t, results, 0)
			},
		},
		{
			name: "IncrementalBackupType",
			mockSetup: func(mockWrapper *MockBRSClientWrapper) {
				mockWrapper.Client.On("GetProtectionPoliciesWithContext", mock.Anything, mock.Anything).Return(
					&backuprecoveryv1.ProtectionPoliciesResponse{
						Policies: []backuprecoveryv1.ProtectionPolicyResponse{
							{
								ID:   core.StringPtr("policy-incremental"),
								Name: core.StringPtr("Incremental Policy"),
								BackupPolicy: &backuprecoveryv1.BackupPolicy{
									Regular: &backuprecoveryv1.RegularBackupPolicy{
										Retention: &backuprecoveryv1.Retention{
											Duration: core.Int64Ptr(30),
										},
										Incremental: &backuprecoveryv1.IncrementalBackupPolicy{
											Schedule: &backuprecoveryv1.IncrementalSchedule{},
										},
									},
								},
							},
						},
					}, nil, nil)
			},
			expectError: false,
			validateResult: func(t *testing.T, results []*types.PolicyResult) {
				assert.Len(t, results, 1)
				assert.Equal(t, "incremental", results[0].BackupType)
			},
		},
		{
			name: "FullBackupType",
			mockSetup: func(mockWrapper *MockBRSClientWrapper) {
				mockWrapper.Client.On("GetProtectionPoliciesWithContext", mock.Anything, mock.Anything).Return(
					&backuprecoveryv1.ProtectionPoliciesResponse{
						Policies: []backuprecoveryv1.ProtectionPolicyResponse{
							{
								ID:   core.StringPtr("policy-full"),
								Name: core.StringPtr("Full Policy"),
								BackupPolicy: &backuprecoveryv1.BackupPolicy{
									Regular: &backuprecoveryv1.RegularBackupPolicy{
										Retention: &backuprecoveryv1.Retention{
											Duration: core.Int64Ptr(30),
										},
										FullBackups: []backuprecoveryv1.FullScheduleAndRetention{
											{},
										},
									},
								},
							},
						},
					}, nil, nil)
			},
			expectError: false,
			validateResult: func(t *testing.T, results []*types.PolicyResult) {
				assert.Len(t, results, 1)
				assert.Equal(t, "full", results[0].BackupType)
			},
		},
		{
			name: "IncrementalPlusFullBackupType",
			mockSetup: func(mockWrapper *MockBRSClientWrapper) {
				mockWrapper.Client.On("GetProtectionPoliciesWithContext", mock.Anything, mock.Anything).Return(
					&backuprecoveryv1.ProtectionPoliciesResponse{
						Policies: []backuprecoveryv1.ProtectionPolicyResponse{
							{
								ID:   core.StringPtr("policy-both"),
								Name: core.StringPtr("Incremental+Full Policy"),
								BackupPolicy: &backuprecoveryv1.BackupPolicy{
									Regular: &backuprecoveryv1.RegularBackupPolicy{
										Retention: &backuprecoveryv1.Retention{
											Duration: core.Int64Ptr(30),
										},
										Incremental: &backuprecoveryv1.IncrementalBackupPolicy{
											Schedule: &backuprecoveryv1.IncrementalSchedule{},
										},
										FullBackups: []backuprecoveryv1.FullScheduleAndRetention{
											{},
										},
									},
								},
							},
						},
					}, nil, nil)
			},
			expectError: false,
			validateResult: func(t *testing.T, results []*types.PolicyResult) {
				assert.Len(t, results, 1)
				assert.Equal(t, "incremental+full", results[0].BackupType)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockWrapper := createMockBRSClientWrapper()
			tt.mockSetup(mockWrapper)
			taskAPI := NewTaskAPI(mockWrapper, createTestMockConfig())
			ctx := context.Background()

			result, err := taskAPI.ListPolicies(ctx)

			if tt.expectError {
				assert.NotNil(t, err)
				assert.Nil(t, result)
				if tt.errorContains != "" {
					assert.Contains(t, err.Error(), tt.errorContains)
				}
			} else {
				assert.Nil(t, err)
				if tt.validateResult != nil {
					tt.validateResult(t, result)
				}
			}
			mockWrapper.Client.AssertExpectations(t)
		})
	}
}

func TestDeletePolicy(t *testing.T) {
	tests := []struct {
		name          string
		policyID      string
		mockSetup     func(*MockBRSClientWrapper)
		expectError   bool
		errorContains string
	}{
		{
			name:     "APIError",
			policyID: "policy-1",
			mockSetup: func(mockWrapper *MockBRSClientWrapper) {
				mockWrapper.Client.On("DeleteProtectionPolicyWithContext", mock.Anything, mock.MatchedBy(
					func(opts *backuprecoveryv1.DeleteProtectionPolicyOptions) bool {
						return *opts.ID == "policy-1"
					})).Return(nil, fmt.Errorf("delete failed"))
			},
			expectError:   true,
			errorContains: "failed to delete protection POlicy by Id",
		},
		{
			name:     "Success",
			policyID: "policy-1",
			mockSetup: func(mockWrapper *MockBRSClientWrapper) {
				mockWrapper.Client.On("DeleteProtectionPolicyWithContext", mock.Anything, mock.Anything).Return(nil, nil)
			},
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockWrapper := createMockBRSClientWrapper()
			tt.mockSetup(mockWrapper)
			taskAPI := NewTaskAPI(mockWrapper, createTestMockConfig())
			ctx := context.Background()

			err := taskAPI.DeletePolicy(ctx, tt.policyID)

			if tt.expectError {
				assert.NotNil(t, err)
				if tt.errorContains != "" {
					assert.Contains(t, err.Error(), tt.errorContains)
				}
			} else {
				assert.Nil(t, err)
			}
			mockWrapper.Client.AssertExpectations(t)
		})
	}
}
func TestUpdatePolicy(t *testing.T) {
	incUnit := types.IncrementalBackup_Unit_Days
	retentionUnit := types.Retention_Unit_Days
	fullUnit := types.FullSchedule_Unit_Days

	tests := []struct {
		name           string
		policyID       string
		policyParams   *types.PolicyParams
		mockSetup      func(*MockBRSClientWrapper)
		expectError    bool
		errorContains  string
		validateResult func(*testing.T, *types.PolicyResult)
	}{
		{
			name:          "EmptyPolicyID",
			policyID:      "",
			policyParams:  &types.PolicyParams{Name: "test-policy"},
			mockSetup:     func(mockWrapper *MockBRSClientWrapper) {},
			expectError:   true,
			errorContains: "[INVALID_INPUT] policyID is required",
		},
		{
			name:          "NilParams",
			policyID:      "policy-123",
			policyParams:  nil,
			mockSetup:     func(mockWrapper *MockBRSClientWrapper) {},
			expectError:   true,
			errorContains: "[INVALID_INPUT] policyParams.Name is required",
		},
		{
			name:     "EmptyName",
			policyID: "policy-123",
			policyParams: &types.PolicyParams{
				Name: "",
			},
			mockSetup:     func(mockWrapper *MockBRSClientWrapper) {},
			expectError:   true,
			errorContains: "[INVALID_INPUT] policyParams.Name is required",
		},
		{
			name:     "IncrementalBackup_MissingUnit",
			policyID: "policy-123",
			policyParams: &types.PolicyParams{
				Name: "test-policy",
				IncrementalBackup: &types.IncrementalBackup{
					Unit: nil,
				},
			},
			mockSetup:     func(mockWrapper *MockBRSClientWrapper) {},
			expectError:   true,
			errorContains: "[UNKNOWN]",
		},
		{
			name:     "IncrementalBackup_MissingDaySchedule",
			policyID: "policy-123",
			policyParams: &types.PolicyParams{
				Name: "test-policy",
				IncrementalBackup: &types.IncrementalBackup{
					Unit:        &incUnit,
					DaySchedule: nil,
				},
			},
			mockSetup:     func(mockWrapper *MockBRSClientWrapper) {},
			expectError:   true,
			errorContains: "[UNKNOWN]",
		},
		{
			name:     "DataRetention_MissingUnit",
			policyID: "policy-123",
			policyParams: &types.PolicyParams{
				Name: "test-policy",
				DataRetention: &types.DataRetention{
					Unit:      nil,
					RetainFor: 30,
				},
			},
			mockSetup:     func(mockWrapper *MockBRSClientWrapper) {},
			expectError:   true,
			errorContains: "[PROTECTION_FAILED]",
		},
		{
			name:     "DataRetention_MissingRetainFor",
			policyID: "policy-123",
			policyParams: &types.PolicyParams{
				Name: "test-policy",
				DataRetention: &types.DataRetention{
					Unit:      &retentionUnit,
					RetainFor: 0,
				},
			},
			mockSetup:     func(mockWrapper *MockBRSClientWrapper) {},
			expectError:   true,
			errorContains: "[PROTECTION_FAILED]",
		},
		{
			name:     "FullBackup_MissingUnit",
			policyID: "policy-123",
			policyParams: &types.PolicyParams{
				Name: "test-policy",
				FullBackups: []types.FullBackup{
					{
						Unit: nil,
					},
				},
			},
			mockSetup:     func(mockWrapper *MockBRSClientWrapper) {},
			expectError:   true,
			errorContains: "[UNKNOWN]",
		},
		{
			name:     "APIError",
			policyID: "policy-123",
			policyParams: &types.PolicyParams{
				Name: "test-policy",
			},
			mockSetup: func(mockWrapper *MockBRSClientWrapper) {
				mockWrapper.Client.On("UpdateProtectionPolicyWithContext", mock.Anything, mock.Anything).Return(
					nil, nil, fmt.Errorf("API error"))
			},
			expectError:   true,
			errorContains: "[UNKNOWN] failed to update policy",
		},
		{
			name:     "Success_MinimalParams",
			policyID: "policy-123",
			policyParams: &types.PolicyParams{
				Name: "updated-policy",
			},
			mockSetup: func(mockWrapper *MockBRSClientWrapper) {
				mockWrapper.Client.On("UpdateProtectionPolicyWithContext", mock.Anything, mock.MatchedBy(
					func(opts *backuprecoveryv1.UpdateProtectionPolicyOptions) bool {
						return *opts.ID == "policy-123" &&
							*opts.Name == "updated-policy"
					})).Return(&backuprecoveryv1.ProtectionPolicyResponse{
					ID:   core.StringPtr("policy-123"),
					Name: core.StringPtr("updated-policy"),
				}, nil, nil)
			},
			expectError: false,
			validateResult: func(t *testing.T, result *types.PolicyResult) {
				assert.NotNil(t, result)
				assert.Equal(t, "policy-123", result.ID)
				assert.Equal(t, "updated-policy", result.Name)
			},
		},
		{
			name:     "Success_WithIncrementalBackup",
			policyID: "policy-123",
			policyParams: &types.PolicyParams{
				Name: "updated-policy",
				IncrementalBackup: &types.IncrementalBackup{
					Unit: &incUnit,
					DaySchedule: &types.UnitDaySchedule{
						Every: 1,
					},
				},
			},
			mockSetup: func(mockWrapper *MockBRSClientWrapper) {
				mockWrapper.Client.On("UpdateProtectionPolicyWithContext", mock.Anything, mock.MatchedBy(
					func(opts *backuprecoveryv1.UpdateProtectionPolicyOptions) bool {
						return *opts.ID == "policy-123" &&
							*opts.Name == "updated-policy" &&
							opts.BackupPolicy.Regular.Incremental != nil
					})).Return(&backuprecoveryv1.ProtectionPolicyResponse{
					ID:   core.StringPtr("policy-123"),
					Name: core.StringPtr("updated-policy"),
				}, nil, nil)
			},
			expectError: false,
			validateResult: func(t *testing.T, result *types.PolicyResult) {
				assert.NotNil(t, result)
				assert.Equal(t, "policy-123", result.ID)
				assert.Equal(t, "updated-policy", result.Name)
			},
		},
		{
			name:     "Success_WithDataRetention",
			policyID: "policy-123",
			policyParams: &types.PolicyParams{
				Name: "updated-policy",
				DataRetention: &types.DataRetention{
					Unit:      &retentionUnit,
					RetainFor: 60,
				},
			},
			mockSetup: func(mockWrapper *MockBRSClientWrapper) {
				mockWrapper.Client.On("UpdateProtectionPolicyWithContext", mock.Anything, mock.MatchedBy(
					func(opts *backuprecoveryv1.UpdateProtectionPolicyOptions) bool {
						return *opts.ID == "policy-123" &&
							*opts.Name == "updated-policy" &&
							opts.BackupPolicy.Regular.Retention != nil
					})).Return(&backuprecoveryv1.ProtectionPolicyResponse{
					ID:   core.StringPtr("policy-123"),
					Name: core.StringPtr("updated-policy"),
				}, nil, nil)
			},
			expectError: false,
			validateResult: func(t *testing.T, result *types.PolicyResult) {
				assert.NotNil(t, result)
				assert.Equal(t, "policy-123", result.ID)
			},
		},
		{
			name:     "Success_WithPrimaryBackupTarget",
			policyID: "policy-123",
			policyParams: &types.PolicyParams{
				Name: "updated-policy",
				PrimaryBackupTarget: &types.PrimaryBackupTarget{
					UseDefaultBackupTarget: true,
				},
			},
			mockSetup: func(mockWrapper *MockBRSClientWrapper) {
				mockWrapper.Client.On("UpdateProtectionPolicyWithContext", mock.Anything, mock.MatchedBy(
					func(opts *backuprecoveryv1.UpdateProtectionPolicyOptions) bool {
						return *opts.ID == "policy-123" &&
							*opts.Name == "updated-policy" &&
							opts.BackupPolicy.Regular.PrimaryBackupTarget != nil
					})).Return(&backuprecoveryv1.ProtectionPolicyResponse{
					ID:   core.StringPtr("policy-123"),
					Name: core.StringPtr("updated-policy"),
				}, nil, nil)
			},
			expectError: false,
			validateResult: func(t *testing.T, result *types.PolicyResult) {
				assert.NotNil(t, result)
				assert.Equal(t, "policy-123", result.ID)
			},
		},
		{
			name:     "Success_WithFullBackups",
			policyID: "policy-123",
			policyParams: &types.PolicyParams{
				Name: "updated-policy",
				FullBackups: []types.FullBackup{
					{
						Unit: &fullUnit,
						DaySchedule: &types.UnitDaySchedule{
							Every: 7,
						},
						Retention: &types.DataRetention{
							Unit:      &retentionUnit,
							RetainFor: 90,
						},
					},
				},
			},
			mockSetup: func(mockWrapper *MockBRSClientWrapper) {
				mockWrapper.Client.On("UpdateProtectionPolicyWithContext", mock.Anything, mock.MatchedBy(
					func(opts *backuprecoveryv1.UpdateProtectionPolicyOptions) bool {
						return *opts.ID == "policy-123" &&
							*opts.Name == "updated-policy" &&
							opts.BackupPolicy.Regular.FullBackups != nil &&
							opts.IsCBSEnabled != nil &&
							*opts.IsCBSEnabled == true
					})).Return(&backuprecoveryv1.ProtectionPolicyResponse{
					ID:   core.StringPtr("policy-123"),
					Name: core.StringPtr("updated-policy"),
				}, nil, nil)
			},
			expectError: false,
			validateResult: func(t *testing.T, result *types.PolicyResult) {
				assert.NotNil(t, result)
				assert.Equal(t, "policy-123", result.ID)
			},
		},
		{
			name:     "Success_WithRetryOptions",
			policyID: "policy-123",
			policyParams: &types.PolicyParams{
				Name: "updated-policy",
				RetryOption: &types.RetryOption{
					NumberOfRetry:       5,
					WaitInCaseOfFailure: 10,
				},
			},
			mockSetup: func(mockWrapper *MockBRSClientWrapper) {
				mockWrapper.Client.On("UpdateProtectionPolicyWithContext", mock.Anything, mock.MatchedBy(
					func(opts *backuprecoveryv1.UpdateProtectionPolicyOptions) bool {
						return *opts.ID == "policy-123" &&
							*opts.Name == "updated-policy" &&
							opts.RetryOptions != nil
					})).Return(&backuprecoveryv1.ProtectionPolicyResponse{
					ID:   core.StringPtr("policy-123"),
					Name: core.StringPtr("updated-policy"),
				}, nil, nil)
			},
			expectError: false,
			validateResult: func(t *testing.T, result *types.PolicyResult) {
				assert.NotNil(t, result)
				assert.Equal(t, "policy-123", result.ID)
			},
		},
		{
			name:     "Success_WithAllParameters",
			policyID: "policy-123",
			policyParams: &types.PolicyParams{
				Name: "updated-policy",
				IncrementalBackup: &types.IncrementalBackup{
					Unit: &incUnit,
					DaySchedule: &types.UnitDaySchedule{
						Every: 1,
					},
				},
				DataRetention: &types.DataRetention{
					Unit:      &retentionUnit,
					RetainFor: 45,
				},
				PrimaryBackupTarget: &types.PrimaryBackupTarget{
					UseDefaultBackupTarget: true,
				},
				FullBackups: []types.FullBackup{
					{
						Unit: &fullUnit,
						DaySchedule: &types.UnitDaySchedule{
							Every: 7,
						},
						Retention: &types.DataRetention{
							Unit:      &retentionUnit,
							RetainFor: 120,
						},
					},
				},
				RetryOption: &types.RetryOption{
					NumberOfRetry:       3,
					WaitInCaseOfFailure: 5,
				},
			},
			mockSetup: func(mockWrapper *MockBRSClientWrapper) {
				mockWrapper.Client.On("UpdateProtectionPolicyWithContext", mock.Anything, mock.MatchedBy(
					func(opts *backuprecoveryv1.UpdateProtectionPolicyOptions) bool {
						return *opts.ID == "policy-123" &&
							*opts.Name == "updated-policy" &&
							opts.BackupPolicy.Regular.Incremental != nil &&
							opts.BackupPolicy.Regular.Retention != nil &&
							opts.BackupPolicy.Regular.PrimaryBackupTarget != nil &&
							opts.BackupPolicy.Regular.FullBackups != nil &&
							opts.RetryOptions != nil &&
							opts.IsCBSEnabled != nil
					})).Return(&backuprecoveryv1.ProtectionPolicyResponse{
					ID:   core.StringPtr("policy-123"),
					Name: core.StringPtr("updated-policy"),
				}, nil, nil)
			},
			expectError: false,
			validateResult: func(t *testing.T, result *types.PolicyResult) {
				assert.NotNil(t, result)
				assert.Equal(t, "policy-123", result.ID)
				assert.Equal(t, "updated-policy", result.Name)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockWrapper := createMockBRSClientWrapper()
			tt.mockSetup(mockWrapper)
			taskAPI := NewTaskAPI(mockWrapper, createTestMockConfig())
			ctx := context.Background()

			result, err := taskAPI.UpdatePolicy(ctx, tt.policyID, tt.policyParams)

			if tt.expectError {
				assert.NotNil(t, err)
				assert.Nil(t, result)
				if tt.errorContains != "" {
					assert.Contains(t, err.Error(), tt.errorContains)
				}
			} else {
				assert.Nil(t, err)
				if tt.validateResult != nil {
					tt.validateResult(t, result)
				}
			}
			mockWrapper.Client.AssertExpectations(t)
		})
	}
}

func TestRunBackup(t *testing.T) {
	tests := []struct {
		name           string
		groupID        string
		backupParams   *types.BackupParams
		mockSetup      func(*MockBRSClientWrapper)
		expectError    bool
		errorContains  string
		validateResult func(*testing.T, *types.BackupResult)
	}{
		{
			name:    "APIError",
			groupID: "group-1",
			backupParams: &types.BackupParams{
				BackupType: types.BackupType_Full,
				TargetBackupObjectIDs: []types.BackupObject{
					{ID: core.Int64Ptr(101)},
				},
			},
			mockSetup: func(mockWrapper *MockBRSClientWrapper) {
				mockWrapper.Client.On("CreateProtectionGroupRunWithContext", mock.Anything, mock.MatchedBy(
					func(opts *backuprecoveryv1.CreateProtectionGroupRunOptions) bool {
						return *opts.ID == "group-1" && *opts.RunType == string(types.BackupType_Full) && len(opts.Objects) == 1
					})).Return(nil, nil, fmt.Errorf("run failed"))
			},
			expectError:   true,
			errorContains: "Failed to create protection group run",
		},
		{
			name:    "Success",
			groupID: "group-1",
			backupParams: &types.BackupParams{
				BackupType: types.BackupType_Incremental,
				TargetBackupObjectIDs: []types.BackupObject{
					{ID: core.Int64Ptr(11)},
					{ID: core.Int64Ptr(22)},
				},
			},
			mockSetup: func(mockWrapper *MockBRSClientWrapper) {
				mockWrapper.Client.On("CreateProtectionGroupRunWithContext", mock.Anything, mock.MatchedBy(
					func(opts *backuprecoveryv1.CreateProtectionGroupRunOptions) bool {
						return *opts.XIBMTenantID == "test-tenant-id" && *opts.ID == "group-1" &&
							*opts.RunType == string(types.BackupType_Incremental) && len(opts.Objects) == 2 &&
							*opts.Objects[0].ID == int64(11) && *opts.Objects[1].ID == int64(22)
					})).Return(&backuprecoveryv1.CreateProtectionGroupRunResponse{
					ProtectionGroupID: core.StringPtr("group-1"),
				}, nil, nil)
				// getActiveBackupRun calls ListBackups → GetProtectionGroupRunsWithContext
				mockWrapper.Client.On("GetProtectionGroupRunsWithContext", mock.Anything, mock.Anything).Return(
					&backuprecoveryv1.ProtectionGroupRunsResponse{
						Runs: []backuprecoveryv1.ProtectionGroupRun{
							minimalRun("run-1", "group-1", string(types.BackupRun_Status_Running)),
						},
					}, nil, nil)
				// buildBackupRunResult calls GetProtectionRunProgressWithContext
				mockWrapper.Client.On("GetProtectionRunProgressWithContext", mock.Anything, mock.Anything).Return(
					&backuprecoveryv1.GetProtectionRunProgressBody{
						ArchivalRun: []backuprecoveryv1.ArchivalTargetProgressInfo{
							{PercentageCompleted: core.Float32Ptr(float32(10))},
						},
					}, nil, nil)
			},
			expectError: false,
			validateResult: func(t *testing.T, result *types.BackupResult) {
				assert.NotNil(t, result)
				assert.Equal(t, "group-1", result.ProtectionGroupID)
				// BackupID is composite "{groupID}:{runID}"
				assert.Equal(t, "group-1:run-1", result.BackupID)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockWrapper := createMockBRSClientWrapper()
			tt.mockSetup(mockWrapper)
			taskAPI := NewTaskAPI(mockWrapper, createTestMockConfig())
			ctx := context.Background()

			result, err := taskAPI.RunBackup(ctx, tt.groupID, tt.backupParams)

			if tt.expectError {
				assert.NotNil(t, err)
				assert.Nil(t, result)
				if tt.errorContains != "" {
					assert.Contains(t, err.Error(), tt.errorContains)
				}
			} else {
				assert.Nil(t, err)
				if tt.validateResult != nil {
					tt.validateResult(t, result)
				}
			}
			mockWrapper.Client.AssertExpectations(t)
		})
	}
}

func TestGetProtectionRunProgress(t *testing.T) {
	tests := []struct {
		name           string
		runID          string
		mockSetup      func(*MockBRSClientWrapper)
		expectError    bool
		errorContains  string
		validateResult func(*testing.T, *float32)
	}{
		{
			name:  "APIError",
			runID: "backup-1",
			mockSetup: func(mockWrapper *MockBRSClientWrapper) {
				mockWrapper.Client.On("GetProtectionRunProgressWithContext", mock.Anything, mock.MatchedBy(
					func(opts *backuprecoveryv1.GetProtectionRunProgressOptions) bool {
						return *opts.RunID == "backup-1"
					})).Return(nil, nil, fmt.Errorf("progress failed"))
			},
			expectError:   true,
			errorContains: "Unable to get progress percentage for backupjob",
		},
		{
			name:  "NoArchivalProgress",
			runID: "backup-1",
			mockSetup: func(mockWrapper *MockBRSClientWrapper) {
				mockWrapper.Client.On("GetProtectionRunProgressWithContext", mock.Anything, mock.Anything).Return(
					&backuprecoveryv1.GetProtectionRunProgressBody{}, nil, nil)
			},
			expectError: false,
			validateResult: func(t *testing.T, result *float32) {
				assert.Nil(t, result)
			},
		},
		{
			name:  "Success",
			runID: "backup-1",
			mockSetup: func(mockWrapper *MockBRSClientWrapper) {
				progressInt := int64(82)
				mockWrapper.Client.On("GetProtectionRunProgressWithContext", mock.Anything, mock.Anything).Return(
					&backuprecoveryv1.GetProtectionRunProgressBody{
						ArchivalRun: []backuprecoveryv1.ArchivalTargetProgressInfo{
							{
								PercentageCompleted: core.Float32Ptr(float32(progressInt)),
							},
						},
					}, nil, nil)
			},
			expectError: false,
			validateResult: func(t *testing.T, result *float32) {
				assert.NotNil(t, result)
				assert.Equal(t, float32(82), *result)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockWrapper := createMockBRSClientWrapper()
			tt.mockSetup(mockWrapper)
			taskAPI := NewTaskAPI(mockWrapper, createTestMockConfig())
			ctx := context.Background()

			result, err := taskAPI.GetProtectionRunProgress(ctx, tt.runID)

			if tt.expectError {
				assert.NotNil(t, err)
				assert.Nil(t, result)
				if tt.errorContains != "" {
					assert.Contains(t, err.Error(), tt.errorContains)
				}
			} else {
				assert.Nil(t, err)
				if tt.validateResult != nil {
					tt.validateResult(t, result)
				}
			}
			mockWrapper.Client.AssertExpectations(t)
		})
	}
}

func TestGetBackup(t *testing.T) {
	tests := []struct {
		name           string
		backupID       string
		groupID        string
		mockSetup      func(*MockBRSClientWrapper)
		expectError    bool
		errorContains  string
		validateResult func(*testing.T, *types.BackupResult)
	}{
		{
			name:     "APIError",
			backupID: "backup-1",
			groupID:  "group-1",
			mockSetup: func(mockWrapper *MockBRSClientWrapper) {
				mockWrapper.Client.On("GetProtectionGroupRunWithContext", mock.Anything, mock.MatchedBy(
					func(opts *backuprecoveryv1.GetProtectionGroupRunOptions) bool {
						return *opts.RunID == "backup-1" && *opts.ID == "group-1"
					})).Return(nil, nil, fmt.Errorf("get failed"))
			},
			expectError:   true,
			errorContains: "Failed to fetch backup run by id",
		},
		{
			name:     "BuildResultError",
			backupID: "backup-1",
			groupID:  "group-1",
			mockSetup: func(mockWrapper *MockBRSClientWrapper) {
				mockWrapper.Client.On("GetProtectionGroupRunWithContext", mock.Anything, mock.Anything).Return(
					&backuprecoveryv1.ProtectionGroupRun{
						ID:                core.StringPtr("backup-1"),
						ProtectionGroupID: core.StringPtr("group-1"),
					}, nil, nil)
				mockWrapper.Client.On("GetProtectionRunProgressWithContext", mock.Anything, mock.Anything).Return(
					nil, nil, fmt.Errorf("progress failed"))
			},
			expectError:   true,
			errorContains: "Unable to form backupRunResult",
		},
		{
			name:     "Success",
			backupID: "backup-1",
			groupID:  "group-1",
			mockSetup: func(mockWrapper *MockBRSClientWrapper) {
				progressInt := int64(64)
				startUsecs := int64(1000000)
				endUsecs := int64(2000000)
				mockWrapper.Client.On("GetProtectionGroupRunWithContext", mock.Anything, mock.Anything).Return(
					&backuprecoveryv1.ProtectionGroupRun{
						ID:                core.StringPtr("backup-1"),
						ProtectionGroupID: core.StringPtr("group-1"),
						ArchivalInfo: &backuprecoveryv1.ArchivalRunSummary{
							ArchivalTargetResults: []backuprecoveryv1.ArchivalTargetResult{
								{
									Status:         core.StringPtr(string(types.BackupRun_Status_Succeeded)),
									StartTimeUsecs: &startUsecs,
									EndTimeUsecs:   &endUsecs,
								},
							},
						},
					}, nil, nil)
				mockWrapper.Client.On("GetProtectionRunProgressWithContext", mock.Anything, mock.Anything).Return(
					&backuprecoveryv1.GetProtectionRunProgressBody{
						ArchivalRun: []backuprecoveryv1.ArchivalTargetProgressInfo{
							{
								PercentageCompleted: core.Float32Ptr(float32(progressInt)),
							},
						},
					}, nil, nil)
			},
			expectError: false,
			validateResult: func(t *testing.T, result *types.BackupResult) {
				assert.NotNil(t, result)
				assert.Equal(t, "backup-1", result.BackupID)
				assert.Equal(t, "group-1", result.ProtectionGroupID)
				assert.Equal(t, float32(64), *result.Progress)
				assert.Equal(t, string(types.BackupRun_Status_Succeeded), result.Status)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockWrapper := createMockBRSClientWrapper()
			tt.mockSetup(mockWrapper)
			taskAPI := NewTaskAPI(mockWrapper, createTestMockConfig())
			ctx := context.Background()

			result, err := taskAPI.GetBackup(ctx, tt.backupID, tt.groupID)

			if tt.expectError {
				assert.NotNil(t, err)
				assert.Nil(t, result)
				if tt.errorContains != "" {
					assert.Contains(t, err.Error(), tt.errorContains)
				}
			} else {
				assert.Nil(t, err)
				if tt.validateResult != nil {
					tt.validateResult(t, result)
				}
			}
			mockWrapper.Client.AssertExpectations(t)
		})
	}
}

func TestListBackups(t *testing.T) {
	tests := []struct {
		name           string
		groupID        string
		mockSetup      func(*MockBRSClientWrapper)
		expectError    bool
		errorContains  string
		validateResult func(*testing.T, []*types.BackupResult)
	}{
		{
			name:    "APIError",
			groupID: "group-1",
			mockSetup: func(mockWrapper *MockBRSClientWrapper) {
				mockWrapper.Client.On("GetProtectionGroupRunsWithContext", mock.Anything, mock.MatchedBy(
					func(opts *backuprecoveryv1.GetProtectionGroupRunsOptions) bool {
						return *opts.ID == "group-1"
					})).Return(nil, nil, fmt.Errorf("list failed"))
			},
			expectError:   true,
			errorContains: "Failed to fetch backup runs",
		},
		{
			name:    "BuildResultError",
			groupID: "group-1",
			mockSetup: func(mockWrapper *MockBRSClientWrapper) {
				mockWrapper.Client.On("GetProtectionGroupRunsWithContext", mock.Anything, mock.Anything).Return(
					&backuprecoveryv1.ProtectionGroupRunsResponse{
						Runs: []backuprecoveryv1.ProtectionGroupRun{
							{
								ID:                core.StringPtr("backup-1"),
								ProtectionGroupID: core.StringPtr("group-1"),
							},
						},
					}, nil, nil)
				mockWrapper.Client.On("GetProtectionRunProgressWithContext", mock.Anything, mock.Anything).Return(
					nil, nil, fmt.Errorf("progress failed"))
			},
			expectError:   true,
			errorContains: "Unable to form backupRunResult for backupId backup-1",
		},
		{
			name:    "Success",
			groupID: "group-1",
			mockSetup: func(mockWrapper *MockBRSClientWrapper) {
				progressInt := int64(99)
				mockWrapper.Client.On("GetProtectionGroupRunsWithContext", mock.Anything, mock.Anything).Return(
					&backuprecoveryv1.ProtectionGroupRunsResponse{
						Runs: []backuprecoveryv1.ProtectionGroupRun{
							minimalRun("backup-1", "group-1", string(types.BackupRun_Status_Running)),
							minimalRun("backup-2", "group-1", string(types.BackupRun_Status_Succeeded)),
						},
					}, nil, nil)
				mockWrapper.Client.On("GetProtectionRunProgressWithContext", mock.Anything, mock.Anything).Return(
					&backuprecoveryv1.GetProtectionRunProgressBody{
						ArchivalRun: []backuprecoveryv1.ArchivalTargetProgressInfo{
							{
								PercentageCompleted: core.Float32Ptr(float32(progressInt)),
							},
						},
					}, nil, nil)
			},
			expectError: false,
			validateResult: func(t *testing.T, results []*types.BackupResult) {
				assert.NotNil(t, results)
				assert.Len(t, results, 2)
				assert.Equal(t, "backup-1", results[0].BackupID)
				assert.Equal(t, string(types.BackupRun_Status_Running), results[0].Status)
				assert.Equal(t, float32(99), *results[0].Progress)
				assert.Equal(t, "backup-2", results[1].BackupID)
				assert.Equal(t, string(types.BackupRun_Status_Succeeded), results[1].Status)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockWrapper := createMockBRSClientWrapper()
			tt.mockSetup(mockWrapper)
			taskAPI := NewTaskAPI(mockWrapper, createTestMockConfig())
			ctx := context.Background()

			result, err := taskAPI.ListBackups(ctx, tt.groupID)

			if tt.expectError {
				assert.NotNil(t, err)
				assert.Nil(t, result)
				if tt.errorContains != "" {
					assert.Contains(t, err.Error(), tt.errorContains)
				}
			} else {
				assert.Nil(t, err)
				if tt.validateResult != nil {
					tt.validateResult(t, result)
				}
			}
			mockWrapper.Client.AssertExpectations(t)
		})
	}
}

func TestGetActiveBackupRun(t *testing.T) {
	tests := []struct {
		name           string
		groupID        string
		mockSetup      func(*MockBRSClientWrapper)
		expectError    bool
		errorContains  string
		validateResult func(*testing.T, *types.BackupResult)
	}{
		{
			name:    "ListBackupsFails",
			groupID: "group-1",
			mockSetup: func(mockWrapper *MockBRSClientWrapper) {
				mockWrapper.Client.On("GetProtectionGroupRunsWithContext", mock.Anything, mock.Anything).Return(
					nil, nil, fmt.Errorf("list failed"))
			},
			expectError:   true,
			errorContains: "Failed to fetch backup runs for group Id",
		},
		{
			name:    "Success_RunningBackup",
			groupID: "group-1",
			mockSetup: func(mockWrapper *MockBRSClientWrapper) {
				progressInt := int64(50)
				mockWrapper.Client.On("GetProtectionGroupRunsWithContext", mock.Anything, mock.Anything).Return(
					&backuprecoveryv1.ProtectionGroupRunsResponse{
						Runs: []backuprecoveryv1.ProtectionGroupRun{
							minimalRun("backup-1", "group-1", string(types.BackupRun_Status_Running)),
						},
					}, nil, nil)
				mockWrapper.Client.On("GetProtectionRunProgressWithContext", mock.Anything, mock.Anything).Return(
					&backuprecoveryv1.GetProtectionRunProgressBody{
						ArchivalRun: []backuprecoveryv1.ArchivalTargetProgressInfo{
							{
								PercentageCompleted: core.Float32Ptr(float32(progressInt)),
							},
						},
					}, nil, nil)
			},
			expectError: false,
			validateResult: func(t *testing.T, result *types.BackupResult) {
				assert.NotNil(t, result)
				assert.Equal(t, "backup-1", result.BackupID)
				assert.Equal(t, string(types.BackupRun_Status_Running), result.Status)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockWrapper := createMockBRSClientWrapper()
			tt.mockSetup(mockWrapper)
			taskAPI := NewTaskAPI(mockWrapper, createTestMockConfig())
			ctx := context.Background()

			result, err := taskAPI.(*DefaultTaskAPI).getActiveBackupRun(ctx, tt.groupID)

			if tt.expectError {
				assert.NotNil(t, err)
				assert.Nil(t, result)
				if tt.errorContains != "" {
					assert.Contains(t, err.Error(), tt.errorContains)
				}
			} else {
				assert.Nil(t, err)
				if tt.validateResult != nil {
					tt.validateResult(t, result)
				}
			}
			mockWrapper.Client.AssertExpectations(t)
		})
	}
}

func TestRunRestore(t *testing.T) {
	tests := []struct {
		name           string
		groupID        string
		backupID       string
		snapshotID     int64
		restoreParams  *types.RestoreParams
		dataSource     *mockDataSource
		mockSetup      func(*MockBRSClientWrapper)
		expectError    bool
		errorContains  string
		validateResult func(*testing.T, *types.RestoreResult)
	}{
		{
			name:          "DataSourceBuildFails",
			groupID:       "group-1",
			backupID:      "backup-1",
			snapshotID:    101,
			restoreParams: &types.RestoreParams{Name: "restore-1"},
			dataSource: &mockDataSource{
				runRestoreFunc: func(context.Context, string, string, int64, *types.RestoreParams) (*backuprecoveryv1.CreateRecoveryOptions, error) {
					return nil, fmt.Errorf("build failed")
				},
			},
			mockSetup:     func(mockWrapper *MockBRSClientWrapper) {},
			expectError:   true,
			errorContains: "Failed to build recovery parameters",
		},
		{
			name:          "RecoveryAPIFails",
			groupID:       "group-1",
			backupID:      "backup-1",
			snapshotID:    101,
			restoreParams: &types.RestoreParams{Name: "restore-1"},
			dataSource:    &mockDataSource{},
			mockSetup: func(mockWrapper *MockBRSClientWrapper) {
				mockWrapper.Client.On("CreateRecoveryWithContext", mock.Anything, mock.MatchedBy(func(opts *backuprecoveryv1.CreateRecoveryOptions) bool {
					return *opts.XIBMTenantID == "test-tenant-id"
				})).Return(nil, nil, fmt.Errorf("recovery failed"))
			},
			expectError:   true,
			errorContains: "Failed to run recovery",
		},
		{
			name:          "Success",
			groupID:       "group-1",
			backupID:      "backup-1",
			snapshotID:    101,
			restoreParams: &types.RestoreParams{Name: "restore-1"},
			dataSource:    &mockDataSource{},
			mockSetup: func(mockWrapper *MockBRSClientWrapper) {
				mockWrapper.Client.On("CreateRecoveryWithContext", mock.Anything, mock.Anything).Return(
					&backuprecoveryv1.Recovery{
						ID: core.StringPtr("restore-123"),
					}, nil, nil)
			},
			expectError: false,
			validateResult: func(t *testing.T, result *types.RestoreResult) {
				assert.NotNil(t, result)
				assert.Equal(t, "restore-123", result.RestoreID)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockWrapper := createMockBRSClientWrapper()
			tt.mockSetup(mockWrapper)
			taskAPI := NewTaskAPI(mockWrapper, createTestMockConfig())
			ctx := context.Background()

			result, err := taskAPI.RunRestore(ctx, tt.groupID, tt.backupID, tt.snapshotID, tt.restoreParams, tt.dataSource)

			if tt.expectError {
				assert.NotNil(t, err)
				assert.Nil(t, result)
				if tt.errorContains != "" {
					assert.Contains(t, err.Error(), tt.errorContains)
				}
			} else {
				assert.Nil(t, err)
				if tt.validateResult != nil {
					tt.validateResult(t, result)
				}
			}
			mockWrapper.Client.AssertExpectations(t)
		})
	}
}

// =============================================================================
// CONNECTION API TESTS
// =============================================================================

func TestGetConnection(t *testing.T) {
	tests := []struct {
		name           string
		connectionID   string
		mockSetup      func(*MockBRSClientWrapper)
		expectError    bool
		errorContains  string
		validateResult func(*testing.T, *types.ConnectionResult)
	}{
		{
			name:         "Success",
			connectionID: "conn-123",
			mockSetup: func(mockWrapper *MockBRSClientWrapper) {
				mockWrapper.Client.On("GetDataSourceConnectionsWithContext", mock.Anything, mock.Anything).Return(
					&backuprecoveryv1.DataSourceConnectionList{
						Connections: []backuprecoveryv1.DataSourceConnection{
							{
								ConnectionID:      core.StringPtr("conn-123"),
								ConnectionName:    core.StringPtr("test-connection"),
								ConnectionEnvType: core.StringPtr("kKubernetes"),
								RegistrationToken: core.StringPtr("token-abc"),
							},
						},
					}, nil, nil)
				mockWrapper.Client.On("GetDataSourceConnectorsWithContext", mock.Anything, mock.Anything).Return(
					&backuprecoveryv1.DataSourceConnectorList{Connectors: []backuprecoveryv1.DataSourceConnector{}}, &core.DetailedResponse{}, nil)
			},
			expectError: false,
			validateResult: func(t *testing.T, result *types.ConnectionResult) {
				assert.NotNil(t, result)
				assert.Equal(t, "conn-123", result.ConnectionID)
				assert.Equal(t, "test-connection", result.ConnectionName)
				assert.Equal(t, "kKubernetes", result.Type)
			},
		},
		{
			name:         "NotFound",
			connectionID: "non-existent",
			mockSetup: func(mockWrapper *MockBRSClientWrapper) {
				mockWrapper.Client.On("GetDataSourceConnectionsWithContext", mock.Anything, mock.Anything).Return(
					&backuprecoveryv1.DataSourceConnectionList{
						Connections: []backuprecoveryv1.DataSourceConnection{},
					}, nil, nil)
			},
			expectError:   true,
			errorContains: "Connection details not found",
		},
		{
			name:         "APIError",
			connectionID: "conn-123",
			mockSetup: func(mockWrapper *MockBRSClientWrapper) {
				mockWrapper.Client.On("GetDataSourceConnectionsWithContext", mock.Anything, mock.Anything).Return(
					nil, nil, fmt.Errorf("API error"))
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockWrapper := createMockBRSClientWrapper()
			tt.mockSetup(mockWrapper)
			taskAPI := NewTaskAPI(mockWrapper, createTestMockConfig())
			ctx := context.Background()

			result, err := taskAPI.GetConnection(ctx, tt.connectionID)

			if tt.expectError {
				assert.NotNil(t, err)
				assert.Nil(t, result)
				if tt.errorContains != "" {
					assert.Contains(t, err.Message, tt.errorContains)
				}
			} else {
				assert.Nil(t, err)
				if tt.validateResult != nil {
					tt.validateResult(t, result)
				}
			}
			mockWrapper.Client.AssertExpectations(t)
		})
	}
}

func TestListConnections(t *testing.T) {
	tests := []struct {
		name           string
		mockSetup      func(*MockBRSClientWrapper)
		expectError    bool
		validateResult func(*testing.T, []*types.ConnectionResult)
	}{
		{
			name: "Success",
			mockSetup: func(mockWrapper *MockBRSClientWrapper) {
				mockWrapper.Client.On("GetDataSourceConnectionsWithContext", mock.Anything, mock.Anything).Return(
					&backuprecoveryv1.DataSourceConnectionList{
						Connections: []backuprecoveryv1.DataSourceConnection{
							{
								ConnectionID:      core.StringPtr("conn-1"),
								ConnectionName:    core.StringPtr("connection-1"),
								RegistrationToken: core.StringPtr("token-1"),
							},
							{
								ConnectionID:      core.StringPtr("conn-2"),
								ConnectionName:    core.StringPtr("connection-2"),
								RegistrationToken: core.StringPtr("token-2"),
							},
						},
					}, nil, nil)
				mockWrapper.Client.On("GetDataSourceConnectorsWithContext", mock.Anything, mock.Anything).Return(
					&backuprecoveryv1.DataSourceConnectorList{Connectors: []backuprecoveryv1.DataSourceConnector{}}, &core.DetailedResponse{}, nil)
			},
			expectError: false,
			validateResult: func(t *testing.T, results []*types.ConnectionResult) {
				assert.NotNil(t, results)
				assert.Equal(t, 2, len(results))
				assert.Equal(t, "conn-1", results[0].ConnectionID)
				assert.Equal(t, "conn-2", results[1].ConnectionID)
			},
		},
		{
			name: "Empty",
			mockSetup: func(mockWrapper *MockBRSClientWrapper) {
				mockWrapper.Client.On("GetDataSourceConnectionsWithContext", mock.Anything, mock.Anything).Return(
					&backuprecoveryv1.DataSourceConnectionList{
						Connections: []backuprecoveryv1.DataSourceConnection{},
					}, nil, nil)
			},
			expectError: false,
			validateResult: func(t *testing.T, results []*types.ConnectionResult) {
				assert.NotNil(t, results)
				assert.Equal(t, 0, len(results))
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockWrapper := createMockBRSClientWrapper()
			tt.mockSetup(mockWrapper)
			taskAPI := NewTaskAPI(mockWrapper, createTestMockConfig())
			ctx := commoncontext.WithAccountID(context.Background(), "test-account-123")

			results, err := taskAPI.ListConnections(ctx)

			if tt.expectError {
				assert.NotNil(t, err)
			} else {
				assert.Nil(t, err)
				if tt.validateResult != nil {
					tt.validateResult(t, results)
				}
			}
			mockWrapper.Client.AssertExpectations(t)
		})
	}
}

func TestDeleteConnection(t *testing.T) {
	tests := []struct {
		name         string
		connectionID string
		mockSetup    func(*MockBRSClientWrapper)
		expectError  bool
	}{
		{
			name:         "Success",
			connectionID: "conn-123",
			mockSetup: func(mockWrapper *MockBRSClientWrapper) {
				mockWrapper.Client.On("DeleteDataSourceConnectionWithContext", mock.Anything, mock.Anything).Return(nil, nil)
			},
			expectError: false,
		},
		{
			name:         "Error",
			connectionID: "conn-123",
			mockSetup: func(mockWrapper *MockBRSClientWrapper) {
				mockWrapper.Client.On("DeleteDataSourceConnectionWithContext", mock.Anything, mock.Anything).Return(
					nil, fmt.Errorf("delete failed"))
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockWrapper := createMockBRSClientWrapper()
			tt.mockSetup(mockWrapper)
			taskAPI := NewTaskAPI(mockWrapper, createTestMockConfig())
			ctx := context.Background()

			err := taskAPI.DeleteConnection(ctx, tt.connectionID)

			if tt.expectError {
				assert.NotNil(t, err)
			} else {
				assert.Nil(t, err)
			}
			mockWrapper.Client.AssertExpectations(t)
		})
	}
}

// =============================================================================
// REGISTRATION API TESTS
// =============================================================================

func TestGetRegistration(t *testing.T) {
	tests := []struct {
		name           string
		registrationID int64
		mockSetup      func(*MockBRSClientWrapper)
		expectError    bool
		validateResult func(*testing.T, *types.RegistrationResult)
	}{
		{
			name:           "Success",
			registrationID: 123,
			mockSetup: func(mockWrapper *MockBRSClientWrapper) {
				mockWrapper.Client.On("GetSourceRegistrationsWithContext", mock.Anything, mock.Anything).Return(
					&backuprecoveryv1.SourceRegistrations{
						Registrations: []backuprecoveryv1.SourceRegistrationResponseParams{
							{
								ID:           core.Int64Ptr(123),
								ConnectionID: core.Int64Ptr(456),
								SourceInfo: &backuprecoveryv1.Object{
									SourceName: core.StringPtr("test-source"),
								},
								AuthenticationStatus: core.StringPtr("kFinished"),
							},
						},
					}, nil, nil)
			},
			expectError: false,
			validateResult: func(t *testing.T, result *types.RegistrationResult) {
				assert.NotNil(t, result)
				assert.Equal(t, int64(123), result.RegistrationID)
				assert.Equal(t, "test-source", result.SourceName)
			},
		},
		{
			name:           "NotFound",
			registrationID: 999,
			mockSetup: func(mockWrapper *MockBRSClientWrapper) {
				mockWrapper.Client.On("GetSourceRegistrationsWithContext", mock.Anything, mock.Anything).Return(
					&backuprecoveryv1.SourceRegistrations{
						Registrations: []backuprecoveryv1.SourceRegistrationResponseParams{},
					}, nil, nil)
			},
			expectError: false,
			validateResult: func(t *testing.T, result *types.RegistrationResult) {
				assert.Nil(t, result)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockWrapper := createMockBRSClientWrapper()
			tt.mockSetup(mockWrapper)
			taskAPI := NewTaskAPI(mockWrapper, createTestMockConfig())
			ctx := commoncontext.WithAccountID(context.Background(), "test-account-123")

			result, err := taskAPI.GetRegistration(ctx, tt.registrationID)

			if tt.expectError {
				assert.NotNil(t, err)
			} else {
				assert.Nil(t, err)
				if tt.validateResult != nil {
					tt.validateResult(t, result)
				}
			}
			mockWrapper.Client.AssertExpectations(t)
		})
	}
}

func TestListRegistrations(t *testing.T) {
	tests := []struct {
		name           string
		mockSetup      func(*MockBRSClientWrapper)
		expectError    bool
		validateResult func(*testing.T, []*types.RegistrationResult)
	}{
		{
			name: "Success",
			mockSetup: func(mockWrapper *MockBRSClientWrapper) {
				mockWrapper.Client.On("GetSourceRegistrationsWithContext", mock.Anything, mock.Anything).Return(
					&backuprecoveryv1.SourceRegistrations{
						Registrations: []backuprecoveryv1.SourceRegistrationResponseParams{
							{
								ID:           core.Int64Ptr(1),
								ConnectionID: core.Int64Ptr(100),
							},
							{
								ID:           core.Int64Ptr(2),
								ConnectionID: core.Int64Ptr(200),
							},
						},
					}, nil, nil)
			},
			expectError: false,
			validateResult: func(t *testing.T, results []*types.RegistrationResult) {
				assert.NotNil(t, results)
				assert.Equal(t, 2, len(results))
				assert.Equal(t, int64(1), results[0].RegistrationID)
				assert.Equal(t, int64(2), results[1].RegistrationID)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockWrapper := createMockBRSClientWrapper()
			tt.mockSetup(mockWrapper)
			taskAPI := NewTaskAPI(mockWrapper, createTestMockConfig())
			ctx := context.Background()

			results, err := taskAPI.ListRegistrations(ctx)

			if tt.expectError {
				assert.NotNil(t, err)
			} else {
				assert.Nil(t, err)
				if tt.validateResult != nil {
					tt.validateResult(t, results)
				}
			}
			mockWrapper.Client.AssertExpectations(t)
		})
	}
}

func TestUnregisterSource(t *testing.T) {
	tests := []struct {
		name           string
		registrationID int64
		mockSetup      func(*MockBRSClientWrapper)
		expectError    bool
	}{
		{
			name:           "Success",
			registrationID: 123,
			mockSetup: func(mockWrapper *MockBRSClientWrapper) {
				mockWrapper.Client.On("DeleteProtectionSourceRegistrationWithContext", mock.Anything, mock.Anything).Return(nil, nil, nil)
			},
			expectError: false,
		},
		{
			name:           "Error",
			registrationID: 123,
			mockSetup: func(mockWrapper *MockBRSClientWrapper) {
				mockWrapper.Client.On("DeleteProtectionSourceRegistrationWithContext", mock.Anything, mock.Anything).Return(
					nil, nil, fmt.Errorf("unregister failed"))
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockWrapper := createMockBRSClientWrapper()
			tt.mockSetup(mockWrapper)
			taskAPI := NewTaskAPI(mockWrapper, createTestMockConfig())
			ctx := context.Background()

			err := taskAPI.UnregisterSource(ctx, tt.registrationID)

			if tt.expectError {
				assert.NotNil(t, err)
			} else {
				assert.Nil(t, err)
			}
			mockWrapper.Client.AssertExpectations(t)
		})
	}
}

// =============================================================================
// RESTORE API TESTS
// =============================================================================

func TestGetRestore(t *testing.T) {
	tests := []struct {
		name           string
		restoreID      string
		mockSetup      func(*MockBRSClientWrapper)
		expectError    bool
		errorContains  string
		validateResult func(*testing.T, *types.RestoreResult)
	}{
		{
			name:      "Success",
			restoreID: "restore-123",
			mockSetup: func(mockWrapper *MockBRSClientWrapper) {
				mockWrapper.Client.On("GetRecoveryByIDWithContext", mock.Anything, mock.MatchedBy(func(opts *backuprecoveryv1.GetRecoveryByIdOptions) bool {
					return *opts.XIBMTenantID == "test-tenant-id" && *opts.ID == "restore-123"
				})).Return(&backuprecoveryv1.Recovery{
					ID: core.StringPtr("restore-123"),
				}, nil, nil)
			},
			expectError: false,
			validateResult: func(t *testing.T, result *types.RestoreResult) {
				assert.NotNil(t, result)
				assert.Equal(t, "restore-123", result.RestoreID)
			},
		},
		{
			name:      "APIError",
			restoreID: "restore-123",
			mockSetup: func(mockWrapper *MockBRSClientWrapper) {
				mockWrapper.Client.On("GetRecoveryByIDWithContext", mock.Anything, mock.Anything).Return(
					nil, nil, fmt.Errorf("API error"))
			},
			expectError:   true,
			errorContains: "Failed to fetch recovery by id",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockWrapper := createMockBRSClientWrapper()
			tt.mockSetup(mockWrapper)
			taskAPI := NewTaskAPI(mockWrapper, createTestMockConfig())
			ctx := context.Background()

			result, err := taskAPI.GetRestore(ctx, tt.restoreID)

			if tt.expectError {
				assert.NotNil(t, err)
				assert.Nil(t, result)
				if tt.errorContains != "" {
					assert.Contains(t, err.Error(), tt.errorContains)
				}
			} else {
				assert.Nil(t, err)
				if tt.validateResult != nil {
					tt.validateResult(t, result)
				}
			}
			mockWrapper.Client.AssertExpectations(t)
		})
	}
}

func TestListRestores(t *testing.T) {
	tests := []struct {
		name           string
		registrationID int64
		mockSetup      func(*MockBRSClientWrapper)
		expectError    bool
		errorContains  string
		validateResult func(*testing.T, []*types.RestoreResult)
	}{
		{
			name:           "APIError",
			registrationID: 123,
			mockSetup: func(mockWrapper *MockBRSClientWrapper) {
				mockWrapper.Client.On("GetRecoveriesWithContext", mock.Anything, mock.MatchedBy(func(opts *backuprecoveryv1.GetRecoveriesOptions) bool {
					return *opts.XIBMTenantID == "test-tenant-id"
				})).Return(nil, nil, fmt.Errorf("recoveries failed"))
			},
			expectError:   true,
			errorContains: "Failed to fetch recovery by id",
		},
		{
			name:           "Success",
			registrationID: 123,
			mockSetup: func(mockWrapper *MockBRSClientWrapper) {
				mockWrapper.Client.On("GetRecoveriesWithContext", mock.Anything, mock.MatchedBy(func(opts *backuprecoveryv1.GetRecoveriesOptions) bool {
					return *opts.XIBMTenantID == "test-tenant-id"
				})).Return(&backuprecoveryv1.RecoveriesResponse{
					Recoveries: []backuprecoveryv1.Recovery{
						{ID: core.StringPtr("restore-1")},
						{ID: core.StringPtr("restore-2")},
					},
				}, nil, nil)
			},
			expectError: false,
			validateResult: func(t *testing.T, results []*types.RestoreResult) {
				assert.NotNil(t, results)
				assert.Equal(t, 2, len(results))
				assert.Equal(t, "restore-1", results[0].RestoreID)
				assert.Equal(t, "restore-2", results[1].RestoreID)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockWrapper := createMockBRSClientWrapper()
			tt.mockSetup(mockWrapper)
			taskAPI := NewTaskAPI(mockWrapper, createTestMockConfig())
			ctx := context.Background()

			results, err := taskAPI.ListRestores(ctx, tt.registrationID)

			if tt.expectError {
				assert.NotNil(t, err)
				assert.Nil(t, results)
				if tt.errorContains != "" {
					assert.Contains(t, err.Error(), tt.errorContains)
				}
			} else {
				assert.Nil(t, err)
				if tt.validateResult != nil {
					tt.validateResult(t, results)
				}
			}
			mockWrapper.Client.AssertExpectations(t)
		})
	}
}

func TestAbortRestore(t *testing.T) {
	tests := []struct {
		name        string
		restoreID   string
		forceDelete bool
		mockSetup   func(*MockBRSClientWrapper)
		expectError bool
	}{
		{
			name:        "Success",
			restoreID:   "restore-123",
			forceDelete: false,
			mockSetup: func(mockWrapper *MockBRSClientWrapper) {
				mockWrapper.Client.On("CancelRecoveryByIDWithContext", mock.Anything, mock.Anything).Return(nil, nil)
			},
			expectError: false,
		},
		{
			name:        "Error",
			restoreID:   "restore-123",
			forceDelete: false,
			mockSetup: func(mockWrapper *MockBRSClientWrapper) {
				mockWrapper.Client.On("CancelRecoveryByIDWithContext", mock.Anything, mock.Anything).Return(
					nil, fmt.Errorf("cancel failed"))
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockWrapper := createMockBRSClientWrapper()
			tt.mockSetup(mockWrapper)
			taskAPI := NewTaskAPI(mockWrapper, createTestMockConfig())
			ctx := context.Background()

			err := taskAPI.AbortRestore(ctx, tt.restoreID, tt.forceDelete)

			if tt.expectError {
				assert.NotNil(t, err)
			} else {
				assert.Nil(t, err)
			}
			mockWrapper.Client.AssertExpectations(t)
		})
	}
}

// =============================================================================
// CONVERTER FUNCTION TESTS
// =============================================================================

// func TestBuildConnectionResults(nil) {
// 	assert.NotNil(t, result)
// 	assert.Equal(t, 0, len(result))

// 	result = buildConnectionResults(&backuprecoveryv1.DataSourceConnectionList{
// 		Connections: []backuprecoveryv1.DataSourceConnection{},
// 	})
// 	assert.NotNil(t, result)
// 	assert.Equal(t, 0, len(result))
// }

func TestBuildConnectionResults(t *testing.T) {
	tests := []struct {
		name           string
		input          *backuprecoveryv1.DataSourceConnectionList
		validateResult func(*testing.T, []*types.ConnectionResult)
	}{
		{
			name: "WithData",
			input: &backuprecoveryv1.DataSourceConnectionList{
				Connections: []backuprecoveryv1.DataSourceConnection{
					{
						ConnectionID:      core.StringPtr("conn-1"),
						ConnectionName:    core.StringPtr("connection-1"),
						ConnectionEnvType: core.StringPtr("kKubernetes"),
						RegistrationToken: core.StringPtr("token-1"),
					},
					{
						ConnectionID:      core.StringPtr("conn-2"),
						ConnectionName:    core.StringPtr("connection-2"),
						ConnectionEnvType: core.StringPtr("kPhysical"),
						RegistrationToken: core.StringPtr("token-2"),
					},
				},
			},
			validateResult: func(t *testing.T, results []*types.ConnectionResult) {
				assert.NotNil(t, results)
				assert.Equal(t, 2, len(results))
				assert.Equal(t, "conn-1", results[0].ConnectionID)
				assert.Equal(t, "connection-1", results[0].ConnectionName)
				assert.Equal(t, "kKubernetes", results[0].Type)
				assert.Equal(t, "token-1", results[0].RegistrationToken)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			results := buildConnectionResults(tt.input)
			if tt.validateResult != nil {
				tt.validateResult(t, results)
			}
		})
	}
}

func TestBuildRegistrationResults(t *testing.T) {
	tests := []struct {
		name           string
		input          *backuprecoveryv1.SourceRegistrations
		validateResult func(*testing.T, []*types.RegistrationResult)
	}{
		{
			name:  "EmptyList_Nil",
			input: nil,
			validateResult: func(t *testing.T, results []*types.RegistrationResult) {
				assert.NotNil(t, results)
				assert.Equal(t, 0, len(results))
			},
		},
		{
			name: "EmptyList_EmptyArray",
			input: &backuprecoveryv1.SourceRegistrations{
				Registrations: []backuprecoveryv1.SourceRegistrationResponseParams{},
			},
			validateResult: func(t *testing.T, results []*types.RegistrationResult) {
				assert.NotNil(t, results)
				assert.Equal(t, 0, len(results))
			},
		},
		{
			name: "WithData",
			input: &backuprecoveryv1.SourceRegistrations{
				Registrations: []backuprecoveryv1.SourceRegistrationResponseParams{
					{
						ID:           core.Int64Ptr(123),
						ConnectionID: core.Int64Ptr(456),
						SourceInfo: &backuprecoveryv1.Object{
							SourceName: core.StringPtr("source-1"),
						},
						AuthenticationStatus: core.StringPtr("kFinished"),
					},
					{
						ID:           core.Int64Ptr(789),
						ConnectionID: core.Int64Ptr(101),
						SourceInfo: &backuprecoveryv1.Object{
							SourceName: core.StringPtr("source-2"),
						},
						AuthenticationStatus: core.StringPtr("kPending"),
					},
				},
			},
			validateResult: func(t *testing.T, results []*types.RegistrationResult) {
				assert.NotNil(t, results)
				assert.Equal(t, 2, len(results))
				assert.Equal(t, int64(123), results[0].RegistrationID)
				assert.Equal(t, "source-1", results[0].SourceName)
				assert.Equal(t, "kFinished", results[0].Status)
				assert.Equal(t, int64(789), results[1].RegistrationID)
				assert.Equal(t, "source-2", results[1].SourceName)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			results := buildRegistrationResults(tt.input)
			if tt.validateResult != nil {
				tt.validateResult(t, results)
			}
		})
	}
}

func TestSafeString(t *testing.T) {
	tests := []struct {
		name     string
		input    *string
		expected string
	}{
		{
			name:     "WithNil",
			input:    nil,
			expected: "",
		},
		{
			name:     "WithValue",
			input:    core.StringPtr("test-value"),
			expected: "test-value",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := safeString(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestSafeInt64(t *testing.T) {
	tests := []struct {
		name     string
		input    *int64
		expected int64
	}{
		{
			name:     "WithNil",
			input:    nil,
			expected: int64(0),
		},
		{
			name:     "WithValue",
			input:    core.Int64Ptr(12345),
			expected: int64(12345),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := safeInt64(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestExtractSourceName(t *testing.T) {
	tests := []struct {
		name     string
		input    *backuprecoveryv1.Object
		expected string
	}{
		{
			name:     "WithNil",
			input:    nil,
			expected: "",
		},
		{
			name: "WithNilSourceName",
			input: &backuprecoveryv1.Object{
				SourceName: nil,
			},
			expected: "",
		},
		{
			name: "WithValue",
			input: &backuprecoveryv1.Object{
				SourceName: core.StringPtr("test-source"),
			},
			expected: "test-source",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := extractSourceName(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// =============================================================================
// WORKFLOW API TESTS
// =============================================================================

func TestCreateConnection(t *testing.T) {
	tests := []struct {
		name           string
		params         *types.ConnectionParams
		mockSetup      func(*MockBRSClientWrapper)
		expectError    bool
		errorContains  string
		validateResult func(*testing.T, *types.ConnectionResult)
	}{
		{
			name: "Success_NewConnection",
			params: &types.ConnectionParams{
				Name: "test-connection",
				Type: types.ConnectionType_VSI,
			},
			mockSetup: func(mockWrapper *MockBRSClientWrapper) {
				mockWrapper.Client.On("GetDataSourceConnectionsWithContext", mock.Anything, mock.Anything).Return(
					&backuprecoveryv1.DataSourceConnectionList{
						Connections: []backuprecoveryv1.DataSourceConnection{},
					},
					&core.DetailedResponse{StatusCode: 200},
					nil,
				).Once()
				mockWrapper.Client.On("CreateDataSourceConnectionWithContext", mock.Anything, mock.Anything).Return(
					&backuprecoveryv1.DataSourceConnection{
						ConnectionID:      core.StringPtr("conn-123"),
						ConnectionName:    core.StringPtr("test-connection"),
						ConnectionEnvType: core.StringPtr("kPhysical"),
						RegistrationToken: core.StringPtr("token-123"),
					},
					&core.DetailedResponse{StatusCode: 201},
					nil,
				).Once()
			},
			expectError: false,
			validateResult: func(t *testing.T, result *types.ConnectionResult) {
				assert.Equal(t, "conn-123", result.ConnectionID)
				assert.Equal(t, "test-connection", result.ConnectionName)
				assert.Equal(t, "token-123", result.RegistrationToken)
			},
		},
		{
			name: "Success_ExistingConnection",
			params: &types.ConnectionParams{
				Name: "test-connection",
				Type: types.ConnectionType_VSI,
			},
			mockSetup: func(mockWrapper *MockBRSClientWrapper) {
				mockWrapper.Client.On("GetDataSourceConnectionsWithContext", mock.Anything, mock.Anything).Return(
					&backuprecoveryv1.DataSourceConnectionList{
						Connections: []backuprecoveryv1.DataSourceConnection{
							{
								ConnectionID:      core.StringPtr("existing-conn-123"),
								ConnectionName:    core.StringPtr("test-connection"),
								ConnectionEnvType: core.StringPtr("kPhysical"),
								RegistrationToken: core.StringPtr("existing-token"),
							},
						},
					},
					&core.DetailedResponse{StatusCode: 200},
					nil,
				).Once()
				mockWrapper.Client.On("GetDataSourceConnectorsWithContext", mock.Anything, mock.Anything).Return(
					&backuprecoveryv1.DataSourceConnectorList{Connectors: []backuprecoveryv1.DataSourceConnector{}}, &core.DetailedResponse{}, nil)
			},
			expectError: false,
			validateResult: func(t *testing.T, result *types.ConnectionResult) {
				assert.Equal(t, "existing-conn-123", result.ConnectionID)
				assert.Equal(t, "test-connection", result.ConnectionName)
			},
		},
		{
			name:          "NilParams",
			params:        nil,
			mockSetup:     func(mockWrapper *MockBRSClientWrapper) {},
			expectError:   true,
			errorContains: "connectionParams.Name is required",
		},
		{
			name: "EmptyName",
			params: &types.ConnectionParams{
				Name: "",
			},
			mockSetup:     func(mockWrapper *MockBRSClientWrapper) {},
			expectError:   true,
			errorContains: "connectionParams.Name is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockWrapper := createMockBRSClientWrapper()
			tt.mockSetup(mockWrapper)
			taskAPI := NewTaskAPI(mockWrapper, createTestMockConfig())
			ctx := context.Background()

			result, err := taskAPI.CreateConnection(ctx, tt.params)

			if tt.expectError {
				assert.NotNil(t, err)
				assert.Nil(t, result)
				if tt.errorContains != "" {
					assert.Contains(t, err.Error(), tt.errorContains)
				}
			} else {
				assert.Nil(t, err)
				assert.NotNil(t, result)
				if tt.validateResult != nil {
					tt.validateResult(t, result)
				}
			}
			mockWrapper.Client.AssertExpectations(t)
		})
	}
}

func TestGetConnectionByName(t *testing.T) {
	tests := []struct {
		name           string
		connectionName string
		mockSetup      func(*MockBRSClientWrapper)
		expectError    bool
		errorContains  string
		validateResult func(*testing.T, *types.ConnectionResult)
	}{
		{
			name:           "Success",
			connectionName: "test-connection",
			mockSetup: func(mockWrapper *MockBRSClientWrapper) {
				mockWrapper.Client.On("GetDataSourceConnectionsWithContext", mock.Anything, mock.Anything).Return(
					&backuprecoveryv1.DataSourceConnectionList{
						Connections: []backuprecoveryv1.DataSourceConnection{
							{
								ConnectionID:      core.StringPtr("conn-123"),
								ConnectionName:    core.StringPtr("test-connection"),
								ConnectionEnvType: core.StringPtr("kPhysical"),
								RegistrationToken: core.StringPtr("token-123"),
							},
						},
					},
					&core.DetailedResponse{StatusCode: 200},
					nil,
				).Once()
				mockWrapper.Client.On("GetDataSourceConnectorsWithContext", mock.Anything, mock.Anything).Return(
					&backuprecoveryv1.DataSourceConnectorList{Connectors: []backuprecoveryv1.DataSourceConnector{}}, &core.DetailedResponse{}, nil)
			},
			expectError: false,
			validateResult: func(t *testing.T, result *types.ConnectionResult) {
				assert.Equal(t, "conn-123", result.ConnectionID)
			},
		},
		{
			name:           "NotFound",
			connectionName: "non-existent",
			mockSetup: func(mockWrapper *MockBRSClientWrapper) {
				mockWrapper.Client.On("GetDataSourceConnectionsWithContext", mock.Anything, mock.Anything).Return(
					&backuprecoveryv1.DataSourceConnectionList{
						Connections: []backuprecoveryv1.DataSourceConnection{},
					},
					&core.DetailedResponse{StatusCode: 200},
					nil,
				)
			},
			expectError:   true,
			errorContains: "not found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockWrapper := createMockBRSClientWrapper()
			tt.mockSetup(mockWrapper)
			taskAPI := NewTaskAPI(mockWrapper, createTestMockConfig())
			ctx := context.Background()

			result, err := taskAPI.GetConnectionByName(ctx, tt.connectionName)

			if tt.expectError {
				assert.NotNil(t, err)
				assert.Nil(t, result)
				if tt.errorContains != "" {
					assert.Contains(t, err.Error(), tt.errorContains)
				}
			} else {
				assert.Nil(t, err)
				assert.NotNil(t, result)
				if tt.validateResult != nil {
					tt.validateResult(t, result)
				}
			}
			mockWrapper.Client.AssertExpectations(t)
		})
	}
}

func TestGetConnector(t *testing.T) {
	tests := []struct {
		name          string
		connectorID   string
		mockSetup     func(*MockBRSClientWrapper)
		expectError   bool
		errorContains string
	}{
		{
			name:        "Success",
			connectorID: "connector-123",
			mockSetup: func(mockWrapper *MockBRSClientWrapper) {
				mockWrapper.Client.On("GetDataSourceConnectorsWithContext", mock.Anything, mock.Anything).Return(
					&backuprecoveryv1.DataSourceConnectorList{
						Connectors: []backuprecoveryv1.DataSourceConnector{
							{
								ConnectorID:  core.StringPtr("connector-123"),
								ConnectionID: core.StringPtr("123"),
							},
						},
					},
					&core.DetailedResponse{StatusCode: 200},
					nil,
				).Once()
			},
			expectError: false,
		},
		{
			name:        "NotFound",
			connectorID: "non-existent",
			mockSetup: func(mockWrapper *MockBRSClientWrapper) {
				mockWrapper.Client.On("GetDataSourceConnectorsWithContext", mock.Anything, mock.Anything).Return(
					&backuprecoveryv1.DataSourceConnectorList{
						Connectors: []backuprecoveryv1.DataSourceConnector{},
					},
					&core.DetailedResponse{StatusCode: 200},
					nil,
				)
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockWrapper := createMockBRSClientWrapper()
			tt.mockSetup(mockWrapper)
			taskAPI := NewTaskAPI(mockWrapper, createTestMockConfig())
			ctx := context.Background()

			result, err := taskAPI.GetConnector(ctx, tt.connectorID)

			if tt.expectError {
				assert.NotNil(t, err)
				assert.Nil(t, result)
			} else {
				assert.Nil(t, err)
				assert.NotNil(t, result)
			}
			mockWrapper.Client.AssertExpectations(t)
		})
	}
}

func TestGetConnectorByConnection(t *testing.T) {
	tests := []struct {
		name           string
		connectionID   string
		mockSetup      func(*MockBRSClientWrapper)
		expectError    bool
		validateResult func(*testing.T, []*connectors.ConnectorResult)
	}{
		{
			name:         "Success",
			connectionID: "conn-123",
			mockSetup: func(mockWrapper *MockBRSClientWrapper) {
				mockWrapper.Client.On("GetDataSourceConnectorsWithContext", mock.Anything, mock.Anything).Return(
					&backuprecoveryv1.DataSourceConnectorList{
						Connectors: []backuprecoveryv1.DataSourceConnector{
							{
								ConnectorID:  core.StringPtr("connector-123"),
								ConnectionID: core.StringPtr("123"),
							},
						},
					},
					&core.DetailedResponse{StatusCode: 200},
					nil,
				).Once()
			},
			expectError: false,
			validateResult: func(t *testing.T, result []*connectors.ConnectorResult) {
				assert.Len(t, result, 1)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockWrapper := createMockBRSClientWrapper()
			tt.mockSetup(mockWrapper)
			taskAPI := NewTaskAPI(mockWrapper, createTestMockConfig())
			ctx := context.Background()

			result, err := taskAPI.GetConnectorByConnection(ctx, tt.connectionID)

			if tt.expectError {
				assert.NotNil(t, err)
				assert.Nil(t, result)
			} else {
				assert.Nil(t, err)
				if tt.validateResult != nil {
					tt.validateResult(t, result)
				}
			}
			mockWrapper.Client.AssertExpectations(t)
		})
	}
}

func TestDeleteConnector(t *testing.T) {
	tests := []struct {
		name        string
		connectorID string
		mockSetup   func(*MockBRSClientWrapper)
		expectError bool
	}{
		{
			name:        "Success",
			connectorID: "connector-123",
			mockSetup: func(mockWrapper *MockBRSClientWrapper) {
				mockWrapper.Client.On("DeleteDataSourceConnectorWithContext", mock.Anything, mock.Anything).Return(
					&core.DetailedResponse{StatusCode: 204},
					nil,
				).Once()
			},
			expectError: false,
		},
		{
			name:        "Error",
			connectorID: "connector-123",
			mockSetup: func(mockWrapper *MockBRSClientWrapper) {
				mockWrapper.Client.On("DeleteDataSourceConnectorWithContext", mock.Anything, mock.Anything).Return(
					&core.DetailedResponse{StatusCode: 500},
					fmt.Errorf("API error"),
				)
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockWrapper := createMockBRSClientWrapper()
			tt.mockSetup(mockWrapper)
			taskAPI := NewTaskAPI(mockWrapper, createTestMockConfig())
			ctx := context.Background()

			err := taskAPI.DeleteConnector(ctx, tt.connectorID)

			if tt.expectError {
				assert.NotNil(t, err)
			} else {
				assert.Nil(t, err)
			}
			mockWrapper.Client.AssertExpectations(t)
		})
	}
}

func TestGetRegistrationByConnection(t *testing.T) {
	tests := []struct {
		name           string
		connectionID   string
		mockSetup      func(*MockBRSClientWrapper)
		expectError    bool
		validateResult func(*testing.T, *types.RegistrationResult)
	}{
		{
			name:         "Success",
			connectionID: "456",
			mockSetup: func(mockWrapper *MockBRSClientWrapper) {
				mockWrapper.Client.On("GetSourceRegistrationsWithContext", mock.Anything, mock.Anything).Return(
					&backuprecoveryv1.SourceRegistrations{
						Registrations: []backuprecoveryv1.SourceRegistrationResponseParams{
							{
								ID:           core.Int64Ptr(123),
								Name:         core.StringPtr("test-source"),
								ConnectionID: core.Int64Ptr(456),
							},
						},
					},
					&core.DetailedResponse{StatusCode: 200},
					nil,
				).Once()
			},
			expectError: false,
			validateResult: func(t *testing.T, result *types.RegistrationResult) {
				assert.Equal(t, int64(123), result.RegistrationID)
			},
		},
		{
			name:         "NotFound",
			connectionID: "non-existent",
			mockSetup: func(mockWrapper *MockBRSClientWrapper) {
				mockWrapper.Client.On("GetSourceRegistrationsWithContext", mock.Anything, mock.Anything).Return(
					&backuprecoveryv1.SourceRegistrations{
						Registrations: []backuprecoveryv1.SourceRegistrationResponseParams{},
					},
					&core.DetailedResponse{StatusCode: 200},
					nil,
				).Once()
			},
			expectError: false,
			validateResult: func(t *testing.T, result *types.RegistrationResult) {
				assert.Nil(t, result)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockWrapper := createMockBRSClientWrapper()
			tt.mockSetup(mockWrapper)
			taskAPI := NewTaskAPI(mockWrapper, createTestMockConfig())
			ctx := context.Background()

			result, err := taskAPI.GetRegistrationByConnection(ctx, tt.connectionID)

			if tt.expectError {
				assert.NotNil(t, err)
				assert.Nil(t, result)
			} else {
				assert.Nil(t, err)
				if tt.validateResult != nil {
					tt.validateResult(t, result)
				}
			}
			mockWrapper.Client.AssertExpectations(t)
		})
	}
}

// mockConnectorDeployer is a mock implementation of ConnectorDeployer
type mockConnectorDeployer struct {
	deployFunc    func(context.Context, string) (*connectors.ConnectorResult, error)
	getStatusFunc func(context.Context, string) (string, error)
	deleteFunc    func(context.Context, string) error
}

func (m *mockConnectorDeployer) GetType() connectors.ConnectorType {
	return connectors.ConnectorTypeHelm
}

func (m *mockConnectorDeployer) SetDeployContext(ctx connectors.ConnectorDeployContext) {}

func (m *mockConnectorDeployer) Deploy(ctx context.Context, registrationToken string) (*connectors.ConnectorResult, error) {
	if m.deployFunc != nil {
		return m.deployFunc(ctx, registrationToken)
	}
	return &connectors.ConnectorResult{
		ConnectorID: "connector-123",
		Status:      "deployed",
	}, nil
}

func (m *mockConnectorDeployer) GetStatus(ctx context.Context, connectorID string) (string, error) {
	if m.getStatusFunc != nil {
		return m.getStatusFunc(ctx, connectorID)
	}
	return "running", nil
}

func (m *mockConnectorDeployer) Delete(ctx context.Context, connectorID string) error {
	if m.deleteFunc != nil {
		return m.deleteFunc(ctx, connectorID)
	}
	return nil
}

func TestDeployConnector(t *testing.T) {
	tests := []struct {
		name             string
		connector        *mockConnectorDeployer
		connectionResult *types.ConnectionResult
		mockSetup        func(*MockBRSClientWrapper)
		expectError      bool
		errorContains    string
		validateResult   func(*testing.T, *connectors.ConnectorResult)
	}{
		{
			name: "Success_NewConnector",
			connector: &mockConnectorDeployer{
				deployFunc: func(ctx context.Context, token string) (*connectors.ConnectorResult, error) {
					return &connectors.ConnectorResult{
						ConnectorID: "new-connector-123",
						Status:      "deployed",
						Message:     "Successfully deployed",
					}, nil
				},
			},
			connectionResult: &types.ConnectionResult{
				ConnectionID:      "conn-123",
				RegistrationToken: "test-token",
			},
			mockSetup: func(mockWrapper *MockBRSClientWrapper) {
				mockWrapper.Client.On("GetDataSourceConnectorsWithContext", mock.Anything, mock.Anything).Return(
					&backuprecoveryv1.DataSourceConnectorList{
						Connectors: []backuprecoveryv1.DataSourceConnector{},
					},
					&core.DetailedResponse{StatusCode: 200},
					nil,
				).Once()
			},
			expectError: false,
			validateResult: func(t *testing.T, result *connectors.ConnectorResult) {
				assert.Equal(t, "new-connector-123", result.ConnectorID)
				assert.Equal(t, "conn-123", result.ConnectionID)
				assert.Equal(t, "deployed", result.Status)
			},
		},
		{
			name:      "Success_ExistingConnector",
			connector: &mockConnectorDeployer{},
			connectionResult: &types.ConnectionResult{
				ConnectionID:      "conn-123",
				RegistrationToken: "test-token",
			},
			mockSetup: func(mockWrapper *MockBRSClientWrapper) {
				mockWrapper.Client.On("GetDataSourceConnectorsWithContext", mock.Anything, mock.Anything).Return(
					&backuprecoveryv1.DataSourceConnectorList{
						Connectors: []backuprecoveryv1.DataSourceConnector{
							{
								ConnectorID:  core.StringPtr("existing-connector-123"),
								ConnectionID: core.StringPtr("conn-123"),
							},
						},
					},
					&core.DetailedResponse{StatusCode: 200},
					nil,
				).Once()
			},
			expectError: false,
			validateResult: func(t *testing.T, result *connectors.ConnectorResult) {
				assert.Equal(t, "existing-connector-123", result.ConnectorID)
			},
		},
		{
			name:      "EmptyConnectionID",
			connector: &mockConnectorDeployer{},
			connectionResult: &types.ConnectionResult{
				ConnectionID: "",
			},
			mockSetup:     func(mockWrapper *MockBRSClientWrapper) {},
			expectError:   true,
			errorContains: "connectionID is required",
		},
		{
			name:      "Error_GetConnectorFails",
			connector: &mockConnectorDeployer{},
			connectionResult: &types.ConnectionResult{
				ConnectionID:      "conn-123",
				RegistrationToken: "test-token",
			},
			mockSetup: func(mockWrapper *MockBRSClientWrapper) {
				mockWrapper.Client.On("GetDataSourceConnectorsWithContext", mock.Anything, mock.Anything).Return(
					(*backuprecoveryv1.DataSourceConnectorList)(nil),
					&core.DetailedResponse{StatusCode: 500},
					fmt.Errorf("API error"),
				).Once()
			},
			expectError:   true,
			errorContains: "failed to get connector",
		},
		{
			name: "Error_DeployFails",
			connector: &mockConnectorDeployer{
				deployFunc: func(ctx context.Context, token string) (*connectors.ConnectorResult, error) {
					return nil, fmt.Errorf("deployment failed")
				},
			},
			connectionResult: &types.ConnectionResult{
				ConnectionID:      "conn-123",
				RegistrationToken: "test-token",
			},
			mockSetup: func(mockWrapper *MockBRSClientWrapper) {
				mockWrapper.Client.On("GetDataSourceConnectorsWithContext", mock.Anything, mock.Anything).Return(
					&backuprecoveryv1.DataSourceConnectorList{
						Connectors: []backuprecoveryv1.DataSourceConnector{},
					},
					&core.DetailedResponse{StatusCode: 200},
					nil,
				).Once()
			},
			expectError:   true,
			errorContains: "Failed to deploy connector",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockWrapper := createMockBRSClientWrapper()
			tt.mockSetup(mockWrapper)
			taskAPI := NewTaskAPI(mockWrapper, createTestMockConfig())
			ctx := context.Background()

			result, err := taskAPI.DeployConnector(ctx, tt.connector, tt.connectionResult)

			if tt.expectError {
				assert.NotNil(t, err)
				assert.Nil(t, result)
				if tt.errorContains != "" {
					assert.Contains(t, err.Error(), tt.errorContains)
				}
			} else {
				assert.Nil(t, err)
				assert.NotNil(t, result)
				if tt.validateResult != nil {
					tt.validateResult(t, result)
				}
			}
			mockWrapper.Client.AssertExpectations(t)
		})
	}
}

func TestRegisterSource(t *testing.T) {
	tests := []struct {
		name           string
		connectionID   string
		mockSetup      func(*MockBRSClientWrapper)
		expectError    bool
		errorContains  string
		validateResult func(*testing.T, *types.RegistrationResult)
	}{
		{
			name:         "Success_NewRegistration",
			connectionID: "123",
			mockSetup: func(mockWrapper *MockBRSClientWrapper) {
				mockWrapper.Client.On("GetSourceRegistrationsWithContext", mock.Anything, mock.Anything).Return(
					&backuprecoveryv1.SourceRegistrations{
						Registrations: []backuprecoveryv1.SourceRegistrationResponseParams{},
					},
					&core.DetailedResponse{StatusCode: 200},
					nil,
				).Once()
				mockWrapper.Client.On("RegisterProtectionSourceWithContext", mock.Anything, mock.Anything).Return(
					&backuprecoveryv1.SourceRegistrationResponseParams{
						ID:           core.Int64Ptr(456),
						ConnectionID: core.Int64Ptr(123),
					},
					&core.DetailedResponse{StatusCode: 201},
					nil,
				).Once()
			},
			expectError: false,
			validateResult: func(t *testing.T, result *types.RegistrationResult) {
				assert.Equal(t, int64(456), result.RegistrationID)
				assert.Equal(t, "123", result.ConnectionID)
				assert.Equal(t, "mock-datasource", result.SourceName)
			},
		},
		{
			name:         "Success_ExistingRegistration",
			connectionID: "123",
			mockSetup: func(mockWrapper *MockBRSClientWrapper) {
				mockWrapper.Client.On("GetSourceRegistrationsWithContext", mock.Anything, mock.Anything).Return(
					&backuprecoveryv1.SourceRegistrations{
						Registrations: []backuprecoveryv1.SourceRegistrationResponseParams{
							{
								ID:           core.Int64Ptr(789),
								ConnectionID: core.Int64Ptr(123),
							},
						},
					},
					&core.DetailedResponse{StatusCode: 200},
					nil,
				).Once()
			},
			expectError: false,
			validateResult: func(t *testing.T, result *types.RegistrationResult) {
				assert.Equal(t, int64(789), result.RegistrationID)
				assert.Equal(t, "123", result.ConnectionID)
			},
		},
		{
			name:          "EmptyConnectionID",
			connectionID:  "",
			mockSetup:     func(mockWrapper *MockBRSClientWrapper) {},
			expectError:   true,
			errorContains: "connectionID is required",
		},
		{
			name:         "Error_GetRegistrationFails",
			connectionID: "123",
			mockSetup: func(mockWrapper *MockBRSClientWrapper) {
				mockWrapper.Client.On("GetSourceRegistrationsWithContext", mock.Anything, mock.Anything).Return(
					(*backuprecoveryv1.SourceRegistrations)(nil),
					&core.DetailedResponse{StatusCode: 500},
					fmt.Errorf("API error"),
				).Once()
			},
			expectError:   true,
			errorContains: "failed to list registrations",
		},
		{
			name:         "Error_RegisterProtectionSourceFails",
			connectionID: "123",
			mockSetup: func(mockWrapper *MockBRSClientWrapper) {
				mockWrapper.Client.On("GetSourceRegistrationsWithContext", mock.Anything, mock.Anything).Return(
					&backuprecoveryv1.SourceRegistrations{
						Registrations: []backuprecoveryv1.SourceRegistrationResponseParams{},
					},
					&core.DetailedResponse{StatusCode: 200},
					nil,
				).Once()
				mockWrapper.Client.On("RegisterProtectionSourceWithContext", mock.Anything, mock.Anything).Return(
					(*backuprecoveryv1.SourceRegistrationResponseParams)(nil),
					&core.DetailedResponse{StatusCode: 500},
					fmt.Errorf("registration failed"),
				).Once()
			},
			expectError:   true,
			errorContains: "failed to create registration",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockWrapper := createMockBRSClientWrapper()
			tt.mockSetup(mockWrapper)
			taskAPI := NewTaskAPI(mockWrapper, createTestMockConfig())
			ctx := context.Background()

			mockDS := &mockDataSource{}
			result, err := taskAPI.RegisterSource(ctx, mockDS, tt.connectionID)

			if tt.expectError {
				assert.NotNil(t, err)
				assert.Nil(t, result)
				if tt.errorContains != "" {
					assert.Contains(t, err.Error(), tt.errorContains)
				}
			} else {
				assert.Nil(t, err)
				assert.NotNil(t, result)
				if tt.validateResult != nil {
					tt.validateResult(t, result)
				}
			}
			mockWrapper.Client.AssertExpectations(t)
		})
	}
}

func TestRefreshRegistration(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name           string
		registrationID int64
		setupMock      func(*MockBRSClientWrapper)
		expectError    bool
		errorContains  string
	}{
		{
			name:           "source registration not found - get fails",
			registrationID: 123,
			setupMock: func(m *MockBRSClientWrapper) {
				m.Client.On("GetSourceRegistrations", mock.MatchedBy(
					func(opts *backuprecoveryv1.GetSourceRegistrationsOptions) bool {
						return len(opts.Ids) == 1 && opts.Ids[0] == 123
					})).Return(nil, nil, fmt.Errorf("not found"))
			},
			expectError:   true,
			errorContains: "Failed to get source registration",
		},
		{
			name:           "source registration not found - empty response",
			registrationID: 123,
			setupMock: func(m *MockBRSClientWrapper) {
				m.Client.On("GetSourceRegistrations", mock.Anything).Return(
					&backuprecoveryv1.SourceRegistrations{
						Registrations: []backuprecoveryv1.SourceRegistrationResponseParams{},
					}, nil, nil)
			},
			expectError:   true,
			errorContains: "Source registration not found",
		},
		{
			name:           "refresh fails",
			registrationID: 123,
			setupMock: func(m *MockBRSClientWrapper) {
				m.Client.On("GetSourceRegistrations", mock.Anything).Return(
					&backuprecoveryv1.SourceRegistrations{
						Registrations: []backuprecoveryv1.SourceRegistrationResponseParams{
							{ID: core.Int64Ptr(123)},
						},
					}, nil, nil)
				m.Client.On("RefreshProtectionSourceByID", mock.MatchedBy(
					func(opts *backuprecoveryv1.RefreshProtectionSourceByIdOptions) bool {
						return *opts.ID == 123
					})).Return(&core.DetailedResponse{}, fmt.Errorf("refresh failed"))
			},
			expectError:   true,
			errorContains: "Failed to refresh protection source",
		},
		{
			name:           "successful refresh",
			registrationID: 123,
			setupMock: func(m *MockBRSClientWrapper) {
				m.Client.On("GetSourceRegistrations", mock.Anything).Return(
					&backuprecoveryv1.SourceRegistrations{
						Registrations: []backuprecoveryv1.SourceRegistrationResponseParams{
							{ID: core.Int64Ptr(123)},
						},
					}, nil, nil)
				m.Client.On("RefreshProtectionSourceByID", mock.MatchedBy(
					func(opts *backuprecoveryv1.RefreshProtectionSourceByIdOptions) bool {
						return *opts.ID == 123 && *opts.XIBMTenantID == "test-tenant-id"
					})).Return(&core.DetailedResponse{}, nil)
			},
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockWrapper := createMockBRSClientWrapper()
			tt.setupMock(mockWrapper)

			taskAPI := NewTaskAPI(mockWrapper, createTestMockConfig())
			err := taskAPI.RefreshRegistration(ctx, tt.registrationID)

			if tt.expectError {
				assert.NotNil(t, err)
				assert.Contains(t, err.Error(), tt.errorContains)
			} else {
				assert.Nil(t, err)
			}

			mockWrapper.Client.AssertExpectations(t)
		})
	}
}

// =============================================================================
// TESTS FOR RESTORE NAMESPACE-LEVEL PROGRESS AND MESSAGES (buildRestoreResult)
// =============================================================================

func TestGetRestore_NamespaceProgress(t *testing.T) {
	tests := []struct {
		name           string
		restoreID      string
		mockSetup      func(*MockBRSClientWrapper)
		validateResult func(*testing.T, *types.RestoreResult)
	}{
		{
			name:      "TopLevelMessages_Populated",
			restoreID: "restore-msg",
			mockSetup: func(mockWrapper *MockBRSClientWrapper) {
				mockWrapper.Client.On("GetRecoveryByIDWithContext", mock.Anything, mock.Anything).Return(
					&backuprecoveryv1.Recovery{
						ID:       core.StringPtr("restore-msg"),
						Status:   core.StringPtr("Running"),
						Messages: []string{"warning: disk nearly full", "info: resuming from checkpoint"},
					}, nil, nil)
			},
			validateResult: func(t *testing.T, result *types.RestoreResult) {
				assert.Equal(t, "restore-msg", result.RestoreID)
				assert.Equal(t, "Running", result.Status)
				assert.Equal(t, []string{"warning: disk nearly full", "info: resuming from checkpoint"}, result.Messages)
				assert.Empty(t, result.NamespaceProgress)
			},
		},
		{
			name:      "NamespaceProgress_WithStatusAndMessages",
			restoreID: "restore-ns",
			mockSetup: func(mockWrapper *MockBRSClientWrapper) {
				mockWrapper.Client.On("GetRecoveryByIDWithContext", mock.Anything, mock.Anything).Return(
					&backuprecoveryv1.Recovery{
						ID:     core.StringPtr("restore-ns"),
						Status: core.StringPtr("Running"),
						KubernetesParams: &backuprecoveryv1.RecoveryKubernetesParams{
							RecoveryAction: core.StringPtr("RecoverNamespaces"),
							RecoverNamespaceParams: &backuprecoveryv1.RecoverKubernetesParamsRecoverNamespaceParams{
								TargetEnvironment: core.StringPtr("kKubernetes"),
								KubernetesTargetParams: &backuprecoveryv1.RecoverKubernetesNamespaceParamsKubernetesTargetParams{
									Objects: []backuprecoveryv1.KubernetesRecoveryObjectParams{
										{
											SnapshotID: core.StringPtr("snap-1"),
											ObjectInfo: &backuprecoveryv1.CommonRecoverObjectSnapshotParamsObjectInfo{
												Name: core.StringPtr("busybox-app"),
											},
											Status:   core.StringPtr("Succeeded"),
											Messages: []string{"restored 42 PVCs"},
										},
										{
											SnapshotID: core.StringPtr("snap-2"),
											ObjectInfo: &backuprecoveryv1.CommonRecoverObjectSnapshotParamsObjectInfo{
												Name: core.StringPtr("nginx-prod"),
											},
											Status:   core.StringPtr("Failed"),
											Messages: []string{"failed to bind PVC: no storage class"},
										},
									},
								},
							},
						},
					}, nil, nil)
			},
			validateResult: func(t *testing.T, result *types.RestoreResult) {
				assert.Equal(t, "restore-ns", result.RestoreID)
				assert.Len(t, result.NamespaceProgress, 2)

				assert.Equal(t, "busybox-app", result.NamespaceProgress[0].NamespaceName)
				assert.Equal(t, "Succeeded", result.NamespaceProgress[0].Status)
				assert.Equal(t, []string{"restored 42 PVCs"}, result.NamespaceProgress[0].Messages)

				assert.Equal(t, "nginx-prod", result.NamespaceProgress[1].NamespaceName)
				assert.Equal(t, "Failed", result.NamespaceProgress[1].Status)
				assert.Equal(t, []string{"failed to bind PVC: no storage class"}, result.NamespaceProgress[1].Messages)
			},
		},
		{
			name:      "NamespaceProgress_WithProgressTaskID",
			restoreID: "restore-prog",
			mockSetup: func(mockWrapper *MockBRSClientWrapper) {
				pct := float32(75)
				mockWrapper.Client.On("GetRecoveryByIDWithContext", mock.Anything, mock.Anything).Return(
					&backuprecoveryv1.Recovery{
						ID:     core.StringPtr("restore-prog"),
						Status: core.StringPtr("Running"),
						KubernetesParams: &backuprecoveryv1.RecoveryKubernetesParams{
							RecoveryAction: core.StringPtr("RecoverNamespaces"),
							RecoverNamespaceParams: &backuprecoveryv1.RecoverKubernetesParamsRecoverNamespaceParams{
								TargetEnvironment: core.StringPtr("kKubernetes"),
								KubernetesTargetParams: &backuprecoveryv1.RecoverKubernetesNamespaceParamsKubernetesTargetParams{
									Objects: []backuprecoveryv1.KubernetesRecoveryObjectParams{
										{
											SnapshotID: core.StringPtr("snap-1"),
											ObjectInfo: &backuprecoveryv1.CommonRecoverObjectSnapshotParamsObjectInfo{
												Name: core.StringPtr("myapp"),
											},
											Status:         core.StringPtr("Running"),
											ProgressTaskID: core.StringPtr("task-myapp-1"),
										},
									},
								},
							},
						},
					}, nil, nil)
				// Progress monitors returns 75%
				mockWrapper.Client.On("GetProgressMonitorsWithContext", mock.Anything,
					mock.MatchedBy(func(opts *backuprecoveryv1.GetProgressMonitorsOptions) bool {
						return len(opts.TaskPathVec) == 1 && opts.TaskPathVec[0] == "task-myapp-1"
					})).Return(
					&backuprecoveryv1.GetTasksResult{
						ResultGroupVec: []backuprecoveryv1.GetTasksResultResultGroup{
							{
								TaskVec: []backuprecoveryv1.GetTasksResultResultGroupTask{
									{
										Progress: &backuprecoveryv1.TaskProgress{
											PercentFinished: &pct,
										},
									},
								},
							},
						},
					}, nil, nil)
			},
			validateResult: func(t *testing.T, result *types.RestoreResult) {
				assert.Equal(t, "restore-prog", result.RestoreID)
				assert.Len(t, result.NamespaceProgress, 1)
				assert.Equal(t, "myapp", result.NamespaceProgress[0].NamespaceName)
				assert.Equal(t, "Running", result.NamespaceProgress[0].Status)
				assert.Equal(t, 75, result.NamespaceProgress[0].Progress)
				// Overall progress is averaged from namespace progress
				assert.Equal(t, 75, result.Progress)
			},
		},
		{
			name:      "NamespaceProgress_ProgressMonitorError_IsWarningNotFatal",
			restoreID: "restore-prog-err",
			mockSetup: func(mockWrapper *MockBRSClientWrapper) {
				mockWrapper.Client.On("GetRecoveryByIDWithContext", mock.Anything, mock.Anything).Return(
					&backuprecoveryv1.Recovery{
						ID:     core.StringPtr("restore-prog-err"),
						Status: core.StringPtr("Running"),
						KubernetesParams: &backuprecoveryv1.RecoveryKubernetesParams{
							RecoveryAction: core.StringPtr("RecoverNamespaces"),
							RecoverNamespaceParams: &backuprecoveryv1.RecoverKubernetesParamsRecoverNamespaceParams{
								TargetEnvironment: core.StringPtr("kKubernetes"),
								KubernetesTargetParams: &backuprecoveryv1.RecoverKubernetesNamespaceParamsKubernetesTargetParams{
									Objects: []backuprecoveryv1.KubernetesRecoveryObjectParams{
										{
											SnapshotID: core.StringPtr("snap-1"),
											ObjectInfo: &backuprecoveryv1.CommonRecoverObjectSnapshotParamsObjectInfo{
												Name: core.StringPtr("myapp"),
											},
											Status:         core.StringPtr("Running"),
											ProgressTaskID: core.StringPtr("task-err"),
										},
									},
								},
							},
						},
					}, nil, nil)
				// GetProgressMonitors fails — should be a warn, not a fatal error
				mockWrapper.Client.On("GetProgressMonitorsWithContext", mock.Anything, mock.Anything).Return(
					nil, nil, fmt.Errorf("progress monitor unavailable"))
			},
			validateResult: func(t *testing.T, result *types.RestoreResult) {
				// Result is still returned despite progress error
				assert.NotNil(t, result)
				assert.Equal(t, "restore-prog-err", result.RestoreID)
				assert.Len(t, result.NamespaceProgress, 1)
				// Progress defaults to 0 when monitor call fails
				assert.Equal(t, 0, result.NamespaceProgress[0].Progress)
			},
		},
		{
			name:      "OverallProgress_FallsBackToTopLevelProgressTaskID",
			restoreID: "restore-toplevel-prog",
			mockSetup: func(mockWrapper *MockBRSClientWrapper) {
				pct := float32(50)
				mockWrapper.Client.On("GetRecoveryByIDWithContext", mock.Anything, mock.Anything).Return(
					&backuprecoveryv1.Recovery{
						ID:             core.StringPtr("restore-toplevel-prog"),
						Status:         core.StringPtr("Running"),
						ProgressTaskID: core.StringPtr("task-top"),
						// No KubernetesParams — non-k8s restore, no namespace objects
					}, nil, nil)
				mockWrapper.Client.On("GetProgressMonitorsWithContext", mock.Anything,
					mock.MatchedBy(func(opts *backuprecoveryv1.GetProgressMonitorsOptions) bool {
						return len(opts.TaskPathVec) == 1 && opts.TaskPathVec[0] == "task-top"
					})).Return(
					&backuprecoveryv1.GetTasksResult{
						ResultGroupVec: []backuprecoveryv1.GetTasksResultResultGroup{
							{
								TaskVec: []backuprecoveryv1.GetTasksResultResultGroupTask{
									{
										Progress: &backuprecoveryv1.TaskProgress{
											PercentFinished: &pct,
										},
									},
								},
							},
						},
					}, nil, nil)
			},
			validateResult: func(t *testing.T, result *types.RestoreResult) {
				assert.Equal(t, "restore-toplevel-prog", result.RestoreID)
				assert.Empty(t, result.NamespaceProgress)
				assert.Equal(t, 50, result.Progress)
			},
		},
		{
			name:      "NamespaceProgress_AveragedForOverallProgress",
			restoreID: "restore-avg",
			mockSetup: func(mockWrapper *MockBRSClientWrapper) {
				pct1 := float32(60)
				pct2 := float32(40)
				mockWrapper.Client.On("GetRecoveryByIDWithContext", mock.Anything, mock.Anything).Return(
					&backuprecoveryv1.Recovery{
						ID:     core.StringPtr("restore-avg"),
						Status: core.StringPtr("Running"),
						KubernetesParams: &backuprecoveryv1.RecoveryKubernetesParams{
							RecoveryAction: core.StringPtr("RecoverNamespaces"),
							RecoverNamespaceParams: &backuprecoveryv1.RecoverKubernetesParamsRecoverNamespaceParams{
								TargetEnvironment: core.StringPtr("kKubernetes"),
								KubernetesTargetParams: &backuprecoveryv1.RecoverKubernetesNamespaceParamsKubernetesTargetParams{
									Objects: []backuprecoveryv1.KubernetesRecoveryObjectParams{
										{
											SnapshotID:     core.StringPtr("snap-1"),
											ObjectInfo:     &backuprecoveryv1.CommonRecoverObjectSnapshotParamsObjectInfo{Name: core.StringPtr("ns-a")},
											Status:         core.StringPtr("Running"),
											ProgressTaskID: core.StringPtr("task-a"),
										},
										{
											SnapshotID:     core.StringPtr("snap-2"),
											ObjectInfo:     &backuprecoveryv1.CommonRecoverObjectSnapshotParamsObjectInfo{Name: core.StringPtr("ns-b")},
											Status:         core.StringPtr("Running"),
											ProgressTaskID: core.StringPtr("task-b"),
										},
									},
								},
							},
						},
					}, nil, nil)
				mockWrapper.Client.On("GetProgressMonitorsWithContext", mock.Anything,
					mock.MatchedBy(func(opts *backuprecoveryv1.GetProgressMonitorsOptions) bool {
						return len(opts.TaskPathVec) == 1 && opts.TaskPathVec[0] == "task-a"
					})).Return(
					&backuprecoveryv1.GetTasksResult{
						ResultGroupVec: []backuprecoveryv1.GetTasksResultResultGroup{{
							TaskVec: []backuprecoveryv1.GetTasksResultResultGroupTask{{
								Progress: &backuprecoveryv1.TaskProgress{PercentFinished: &pct1},
							}},
						}},
					}, nil, nil)
				mockWrapper.Client.On("GetProgressMonitorsWithContext", mock.Anything,
					mock.MatchedBy(func(opts *backuprecoveryv1.GetProgressMonitorsOptions) bool {
						return len(opts.TaskPathVec) == 1 && opts.TaskPathVec[0] == "task-b"
					})).Return(
					&backuprecoveryv1.GetTasksResult{
						ResultGroupVec: []backuprecoveryv1.GetTasksResultResultGroup{{
							TaskVec: []backuprecoveryv1.GetTasksResultResultGroupTask{{
								Progress: &backuprecoveryv1.TaskProgress{PercentFinished: &pct2},
							}},
						}},
					}, nil, nil)
			},
			validateResult: func(t *testing.T, result *types.RestoreResult) {
				assert.Len(t, result.NamespaceProgress, 2)
				assert.Equal(t, 60, result.NamespaceProgress[0].Progress)
				assert.Equal(t, 40, result.NamespaceProgress[1].Progress)
				// (60+40)/2 = 50
				assert.Equal(t, 50, result.Progress)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockWrapper := createMockBRSClientWrapper()
			tt.mockSetup(mockWrapper)
			taskAPI := NewTaskAPI(mockWrapper, createTestMockConfig())
			ctx := context.Background()

			result, err := taskAPI.GetRestore(ctx, tt.restoreID)

			assert.Nil(t, err)
			assert.NotNil(t, result)
			if tt.validateResult != nil {
				tt.validateResult(t, result)
			}
			mockWrapper.Client.AssertExpectations(t)
		})
	}
}

// =============================================================================
// TESTS FOR BACKUP NAMESPACE-LEVEL PROGRESS AND MESSAGES (buildBackupRunResult)
// =============================================================================

func TestGetBackup_NamespaceProgress(t *testing.T) {
	tests := []struct {
		name           string
		backupID       string
		groupID        string
		mockSetup      func(*MockBRSClientWrapper)
		validateResult func(*testing.T, *types.BackupResult)
	}{
		{
			name:     "NamespaceProgress_StatusAndWarnings",
			backupID: "backup-ns",
			groupID:  "group-1",
			mockSetup: func(mockWrapper *MockBRSClientWrapper) {
				mockWrapper.Client.On("GetProtectionGroupRunWithContext", mock.Anything, mock.Anything).Return(
					&backuprecoveryv1.ProtectionGroupRun{
						ID:                core.StringPtr("backup-ns"),
						ProtectionGroupID: core.StringPtr("group-1"),
						Objects: []backuprecoveryv1.ObjectRunResult{
							{
								Object: &backuprecoveryv1.ObjectSummary{Name: core.StringPtr("frontend-ns")},
								LocalSnapshotInfo: &backuprecoveryv1.BackupRun{
									SnapshotInfo: &backuprecoveryv1.SnapshotInfo{
										Status:        core.StringPtr("kSuccessful"),
										StatusMessage: core.StringPtr("completed with minor warnings"),
										Warnings:      []string{"PVC pvc-logs skipped: ReadWriteOnce"},
									},
								},
							},
							{
								Object: &backuprecoveryv1.ObjectSummary{Name: core.StringPtr("backend-ns")},
								LocalSnapshotInfo: &backuprecoveryv1.BackupRun{
									SnapshotInfo: &backuprecoveryv1.SnapshotInfo{
										Status: core.StringPtr("kFailed"),
									},
								},
							},
						},
					}, nil, nil)
				mockWrapper.Client.On("GetProtectionRunProgressWithContext", mock.Anything, mock.Anything).Return(
					&backuprecoveryv1.GetProtectionRunProgressBody{}, nil, nil)
			},
			validateResult: func(t *testing.T, result *types.BackupResult) {
				assert.Len(t, result.NamespaceProgress, 2)

				assert.Equal(t, "frontend-ns", result.NamespaceProgress[0].NamespaceName)
				assert.Equal(t, "kSuccessful", result.NamespaceProgress[0].Status)

				assert.Equal(t, "backend-ns", result.NamespaceProgress[1].NamespaceName)
				assert.Equal(t, "kFailed", result.NamespaceProgress[1].Status)
			},
		},
		{
			name:     "NamespaceProgress_WithProgressTaskID",
			backupID: "backup-prog",
			groupID:  "group-1",
			mockSetup: func(mockWrapper *MockBRSClientWrapper) {
				pct := float32(88)
				mockWrapper.Client.On("GetProtectionGroupRunWithContext", mock.Anything, mock.Anything).Return(
					&backuprecoveryv1.ProtectionGroupRun{
						ID:                core.StringPtr("backup-prog"),
						ProtectionGroupID: core.StringPtr("group-1"),
						Objects: []backuprecoveryv1.ObjectRunResult{
							{
								Object: &backuprecoveryv1.ObjectSummary{Name: core.StringPtr("myns")},
								LocalSnapshotInfo: &backuprecoveryv1.BackupRun{
									SnapshotInfo: &backuprecoveryv1.SnapshotInfo{
										Status:         core.StringPtr("kInProgress"),
										ProgressTaskID: core.StringPtr("task-backup-myns"),
									},
								},
							},
						},
					}, nil, nil)
				mockWrapper.Client.On("GetProtectionRunProgressWithContext", mock.Anything, mock.Anything).Return(
					&backuprecoveryv1.GetProtectionRunProgressBody{}, nil, nil)
				mockWrapper.Client.On("GetProgressMonitorsWithContext", mock.Anything,
					mock.MatchedBy(func(opts *backuprecoveryv1.GetProgressMonitorsOptions) bool {
						return len(opts.TaskPathVec) == 1 && opts.TaskPathVec[0] == "task-backup-myns"
					})).Return(
					&backuprecoveryv1.GetTasksResult{
						ResultGroupVec: []backuprecoveryv1.GetTasksResultResultGroup{{
							TaskVec: []backuprecoveryv1.GetTasksResultResultGroupTask{{
								Progress: &backuprecoveryv1.TaskProgress{PercentFinished: &pct},
							}},
						}},
					}, nil, nil)
			},
			validateResult: func(t *testing.T, result *types.BackupResult) {
				assert.Len(t, result.NamespaceProgress, 1)
				assert.Equal(t, "myns", result.NamespaceProgress[0].NamespaceName)
				assert.Equal(t, "kInProgress", result.NamespaceProgress[0].Status)
				assert.Equal(t, 88, result.NamespaceProgress[0].Progress)
			},
		},
		{
			name:     "NamespaceProgress_ProgressMonitorError_IsWarningNotFatal",
			backupID: "backup-prog-err",
			groupID:  "group-1",
			mockSetup: func(mockWrapper *MockBRSClientWrapper) {
				mockWrapper.Client.On("GetProtectionGroupRunWithContext", mock.Anything, mock.Anything).Return(
					&backuprecoveryv1.ProtectionGroupRun{
						ID:                core.StringPtr("backup-prog-err"),
						ProtectionGroupID: core.StringPtr("group-1"),
						Objects: []backuprecoveryv1.ObjectRunResult{
							{
								Object: &backuprecoveryv1.ObjectSummary{Name: core.StringPtr("myns")},
								LocalSnapshotInfo: &backuprecoveryv1.BackupRun{
									SnapshotInfo: &backuprecoveryv1.SnapshotInfo{
										Status:         core.StringPtr("kInProgress"),
										ProgressTaskID: core.StringPtr("task-err"),
									},
								},
							},
						},
					}, nil, nil)
				mockWrapper.Client.On("GetProtectionRunProgressWithContext", mock.Anything, mock.Anything).Return(
					&backuprecoveryv1.GetProtectionRunProgressBody{}, nil, nil)
				mockWrapper.Client.On("GetProgressMonitorsWithContext", mock.Anything, mock.Anything).Return(
					nil, nil, fmt.Errorf("monitor unavailable"))
			},
			validateResult: func(t *testing.T, result *types.BackupResult) {
				// Must still return a result — progress error is non-fatal
				assert.NotNil(t, result)
				assert.Equal(t, "backup-prog-err", result.BackupID)
				assert.Len(t, result.NamespaceProgress, 1)
				assert.Equal(t, 0, result.NamespaceProgress[0].Progress)
			},
		},
		{
			name:     "NoObjects_NoNamespaceProgress",
			backupID: "backup-plain",
			groupID:  "group-1",
			mockSetup: func(mockWrapper *MockBRSClientWrapper) {
				mockWrapper.Client.On("GetProtectionGroupRunWithContext", mock.Anything, mock.Anything).Return(
					&backuprecoveryv1.ProtectionGroupRun{
						ID:                core.StringPtr("backup-plain"),
						ProtectionGroupID: core.StringPtr("group-1"),
					}, nil, nil)
				mockWrapper.Client.On("GetProtectionRunProgressWithContext", mock.Anything, mock.Anything).Return(
					&backuprecoveryv1.GetProtectionRunProgressBody{}, nil, nil)
			},
			validateResult: func(t *testing.T, result *types.BackupResult) {
				assert.Equal(t, "backup-plain", result.BackupID)
				assert.Empty(t, result.NamespaceProgress)
			},
		},
		{
			// RealWorldCloudArchivalRunning mirrors the exact response shape from a live
			// BRS GET /protection-groups/{id}/runs/{runId} call:
			//   id:                "206327:1788786384042187"
			//   protectionGroupId: "6774032249995190:1753115492078:206327"
			//   objects:           null        (no per-namespace objects yet)
			//   archivalInfo.archivalTargetResults[0].status: "Running"
			//   archivalInfo.archivalTargetResults[0].startTimeUsecs: 1788786384042187
			//   isCloudArchivalDirect: true, environment: kKubernetes
			name:     "RealWorldCloudArchivalRunning_NullObjects",
			backupID: "206327:1788786384042187",
			groupID:  "6774032249995190:1753115492078:206327",
			mockSetup: func(mockWrapper *MockBRSClientWrapper) {
				startUsecs := int64(1788786384042187)
				targetID := int64(16522607)
				mockWrapper.Client.On("GetProtectionGroupRunWithContext", mock.Anything, mock.Anything).Return(
					&backuprecoveryv1.ProtectionGroupRun{
						ID:                    core.StringPtr("206327:1788786384042187"),
						ProtectionGroupID:     core.StringPtr("6774032249995190:1753115492078:206327"),
						IsCloudArchivalDirect: core.BoolPtr(true),
						HasLocalSnapshot:      core.BoolPtr(true),
						Environment:           core.StringPtr("kKubernetes"),
						// Objects is nil — mirrors "objects": null in the real response
						ArchivalInfo: &backuprecoveryv1.ArchivalRunSummary{
							ArchivalTargetResults: []backuprecoveryv1.ArchivalTargetResult{
								{
									TargetID:               &targetID,
									ArchivalTaskID:         core.StringPtr("6774032249995190:1753115492078:24329558"),
									TargetName:             core.StringPtr("ExtTarget-794118aa-5fae-46fd-bfb6-a41e5e310653"),
									TargetType:             core.StringPtr("Cloud"),
									UsageType:              core.StringPtr("Archival"),
									OwnershipContext:       core.StringPtr("Local"),
									RunType:                core.StringPtr("kRegular"),
									StartTimeUsecs:         &startUsecs,
									Status:                 core.StringPtr("Running"),
									IndexingTaskID:         core.StringPtr("indexing_206327_206332"),
									SuccessfulObjectsCount: core.Int64Ptr(0),
									FailedObjectsCount:     core.Int64Ptr(0),
								},
							},
						},
					}, nil, nil)
				mockWrapper.Client.On("GetProtectionRunProgressWithContext", mock.Anything, mock.Anything).Return(
					&backuprecoveryv1.GetProtectionRunProgressBody{}, nil, nil)
			},
			validateResult: func(t *testing.T, result *types.BackupResult) {
				assert.NotNil(t, result)
				assert.Equal(t, "206327:1788786384042187", result.BackupID)
				assert.Equal(t, "6774032249995190:1753115492078:206327", result.ProtectionGroupID)
				assert.Equal(t, "Running", result.Status)
				// startTimeUsecs present → StartedAt must be populated
				assert.False(t, result.StartedAt.IsZero(), "StartedAt must be set from startTimeUsecs")
				assert.Equal(t, time.UnixMicro(1788786384042187).UTC(), result.StartedAt.UTC())
				// No endTimeUsecs → CompletedAt must be nil
				assert.Nil(t, result.CompletedAt, "CompletedAt must be nil when endTimeUsecs is absent")
				// objects=null → no namespace-level progress
				assert.Empty(t, result.NamespaceProgress)
				// No archival run progress data → Progress is nil
				assert.Nil(t, result.Progress)
			},
		},
		{
			// CloudArchivalDirect_NamespaceProgress mirrors the real API response shape
			// where isCloudArchivalDirect=true and objects[].archivalInfo is populated
			// (no localSnapshotInfo).  Observed in production for environment=kKubernetes.
			//
			// Real response objects:
			//   objects[0]: name="brs-migration",  archivalInfo.status="Running", progressTaskId="backup_206332_2/task_206340"
			//   objects[1]: name="ibm-observe",    archivalInfo.status="Running", progressTaskId="backup_206332_1/task_206334"
			name:     "CloudArchivalDirect_NamespaceProgress",
			backupID: "206327:1788786384042187",
			groupID:  "6774032249995190:1753115492078:206327",
			mockSetup: func(mockWrapper *MockBRSClientWrapper) {
				startUsecs := int64(1788786384042187)
				overallPct := float32(45.0)
				nsPct1 := float32(32.67)
				nsPct2 := float32(58.0)
				mockWrapper.Client.On("GetProtectionGroupRunWithContext", mock.Anything, mock.Anything).Return(
					&backuprecoveryv1.ProtectionGroupRun{
						ID:                    core.StringPtr("206327:1788786384042187"),
						ProtectionGroupID:     core.StringPtr("6774032249995190:1753115492078:206327"),
						IsCloudArchivalDirect: core.BoolPtr(true),
						HasLocalSnapshot:      core.BoolPtr(true),
						Environment:           core.StringPtr("kKubernetes"),
						ArchivalInfo: &backuprecoveryv1.ArchivalRunSummary{
							ArchivalTargetResults: []backuprecoveryv1.ArchivalTargetResult{
								{
									StartTimeUsecs: &startUsecs,
									Status:         core.StringPtr("Running"),
								},
							},
						},
						// objects=null — always the case for cloud-archival-direct runs
						Objects: nil,
					}, nil, nil)
				// Progress API returns per-namespace objects (the real source of truth)
				mockWrapper.Client.On("GetProtectionRunProgressWithContext", mock.Anything, mock.Anything).Return(
					&backuprecoveryv1.GetProtectionRunProgressBody{
						ArchivalRun: []backuprecoveryv1.ArchivalTargetProgressInfo{
							{
								Status:              core.StringPtr("Active"),
								PercentageCompleted: &overallPct,
								Objects: []backuprecoveryv1.ObjectProgressInfo{
									{
										Name:                core.StringPtr("e2e-app-nginx"),
										Status:              core.StringPtr("Active"),
										PercentageCompleted: &nsPct1,
									},
									{
										Name:                core.StringPtr("e2e-app-busybox"),
										Status:              core.StringPtr("Active"),
										PercentageCompleted: &nsPct2,
									},
								},
							},
						},
					}, nil, nil)
			},
			validateResult: func(t *testing.T, result *types.BackupResult) {
				assert.NotNil(t, result)
				assert.Equal(t, "Running", result.Status)
				// Overall progress from ArchivalRun[0].PercentageCompleted
				assert.NotNil(t, result.Progress)
				assert.Equal(t, float32(45.0), *result.Progress)
				// Per-namespace progress from ArchivalRun[0].Objects[]
				assert.Len(t, result.NamespaceProgress, 2)

				assert.Equal(t, "e2e-app-nginx", result.NamespaceProgress[0].NamespaceName)
				assert.Equal(t, "Active", result.NamespaceProgress[0].Status)
				assert.Equal(t, 32, result.NamespaceProgress[0].Progress) // int(32.67) — Active, not forced to 100

				assert.Equal(t, "e2e-app-busybox", result.NamespaceProgress[1].NamespaceName)
				assert.Equal(t, "Active", result.NamespaceProgress[1].Status)
				assert.Equal(t, 58, result.NamespaceProgress[1].Progress) // Active, not forced to 100
			},
		},
		{
			// Validates that status=Finished forces progress=100 (server leaves PercentageCompleted=0 after completion)
			name:     "CloudArchivalDirect_FinishedNamespace_Progress100",
			backupID: "159354:1788883592073654",
			groupID:  "2712860000048009:1757348677013:159354",
			mockSetup: func(mockWrapper *MockBRSClientWrapper) {
				startUsecs := int64(1788883592073654)
				endUsecs := int64(1788883684808702)
				overallPct := float32(100.0)
				zeroPct := float32(0.0) // server returns 0 after completion
				mockWrapper.Client.On("GetProtectionGroupRunWithContext", mock.Anything, mock.Anything).Return(
					&backuprecoveryv1.ProtectionGroupRun{
						ID:                    core.StringPtr("159354:1788883592073654"),
						ProtectionGroupID:     core.StringPtr("2712860000048009:1757348677013:159354"),
						IsCloudArchivalDirect: core.BoolPtr(true),
						ArchivalInfo: &backuprecoveryv1.ArchivalRunSummary{
							ArchivalTargetResults: []backuprecoveryv1.ArchivalTargetResult{
								{StartTimeUsecs: &startUsecs, EndTimeUsecs: &endUsecs, Status: core.StringPtr("Succeeded")},
							},
						},
						Objects: nil,
					}, nil, nil)
				mockWrapper.Client.On("GetProtectionRunProgressWithContext", mock.Anything, mock.Anything).Return(
					&backuprecoveryv1.GetProtectionRunProgressBody{
						ArchivalRun: []backuprecoveryv1.ArchivalTargetProgressInfo{
							{
								Status:              core.StringPtr("Finished"),
								PercentageCompleted: &overallPct,
								Objects: []backuprecoveryv1.ObjectProgressInfo{
									{Name: core.StringPtr("e2e-app-nginx"), Status: core.StringPtr("Finished"), PercentageCompleted: &zeroPct},
									{Name: core.StringPtr("e2e-app-busybox"), Status: core.StringPtr("Finished"), PercentageCompleted: &zeroPct},
								},
							},
						},
					}, nil, nil)
			},
			validateResult: func(t *testing.T, result *types.BackupResult) {
				assert.NotNil(t, result)
				assert.Equal(t, "Succeeded", result.Status)
				assert.Len(t, result.NamespaceProgress, 2)

				assert.Equal(t, "e2e-app-nginx", result.NamespaceProgress[0].NamespaceName)
				assert.Equal(t, "Finished", result.NamespaceProgress[0].Status)
				assert.Equal(t, 100, result.NamespaceProgress[0].Progress) // forced to 100

				assert.Equal(t, "e2e-app-busybox", result.NamespaceProgress[1].NamespaceName)
				assert.Equal(t, "Finished", result.NamespaceProgress[1].Status)
				assert.Equal(t, 100, result.NamespaceProgress[1].Progress) // forced to 100
			},
		},
	}

	// Reuse the time import already present in the file to avoid an unused import
	_ = time.Now()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockWrapper := createMockBRSClientWrapper()
			tt.mockSetup(mockWrapper)
			taskAPI := NewTaskAPI(mockWrapper, createTestMockConfig())
			ctx := context.Background()

			result, err := taskAPI.GetBackup(ctx, tt.backupID, tt.groupID)

			assert.Nil(t, err)
			assert.NotNil(t, result)
			if tt.validateResult != nil {
				tt.validateResult(t, result)
			}
			mockWrapper.Client.AssertExpectations(t)
		})
	}
}

// =============================================================================
// GetMetaInfo Tests
// =============================================================================

func TestGetMetaInfo(t *testing.T) {
	tests := []struct {
		name          string
		params        *types.MetaInfoParams
		setupMocks    func(m *testmockrs.MockBRSClient)
		expectedLen   int
		expectedError string
	}{
		{
			name:          "NilParams",
			params:        nil,
			expectedError: "meta info params cannot be nil",
		},
		{
			name:          "MissingGroupID",
			params:        &types.MetaInfoParams{GroupID: ""},
			expectedError: "groupID is mandatory for GetMetaInfo",
		},
		{
			name: "DirectSnapshotIDs_Success",
			params: &types.MetaInfoParams{
				GroupID:     "group-1",
				SnapshotIDs: []string{"snap-1", "snap-2"},
			},
			setupMocks: func(m *testmockrs.MockBRSClient) {
				m.On("ConstructMetaInfoWithContext", mock.Anything, mock.MatchedBy(func(opts *backuprecoveryv1.ConstructMetaInfoOptions) bool {
					return *opts.SnapshotID == "snap-1"
				})).Return(&backuprecoveryv1.ConstructMetaInfoResult{
					Environment: core.StringPtr("kKubernetes"),
					KubernetesParams: &backuprecoveryv1.ConstructMetaInfoResultKubernetesParams{
						BackedUpResourceCount: core.Int64Ptr(10),
					},
				}, &core.DetailedResponse{StatusCode: 200}, nil)

				m.On("ConstructMetaInfoWithContext", mock.Anything, mock.MatchedBy(func(opts *backuprecoveryv1.ConstructMetaInfoOptions) bool {
					return *opts.SnapshotID == "snap-2"
				})).Return(&backuprecoveryv1.ConstructMetaInfoResult{
					Environment: core.StringPtr("kKubernetes"),
					KubernetesParams: &backuprecoveryv1.ConstructMetaInfoResultKubernetesParams{
						BackedUpResourceCount: core.Int64Ptr(20),
					},
				}, &core.DetailedResponse{StatusCode: 200}, nil)
			},
			expectedLen: 2,
		},
		{
			name: "BackupIDProvided_Success",
			params: &types.MetaInfoParams{
				GroupID:  "group-1",
				BackupID: "backup-1",
			},
			setupMocks: func(m *testmockrs.MockBRSClient) {
				m.On("GetProtectionGroupRunWithContext", mock.Anything, mock.MatchedBy(func(opts *backuprecoveryv1.GetProtectionGroupRunOptions) bool {
					return *opts.RunID == "backup-1" && *opts.ID == "group-1"
				})).Return(&backuprecoveryv1.ProtectionGroupRun{
					Objects: []backuprecoveryv1.ObjectRunResult{
						{
							Object: &backuprecoveryv1.ObjectSummary{
								Name: core.StringPtr("ns-1"),
								ID:   core.Int64Ptr(101),
							},
							ArchivalInfo: &backuprecoveryv1.ArchivalRun{
								ArchivalTargetResults: []backuprecoveryv1.ArchivalTargetResult{
									{SnapshotID: core.StringPtr("snap-ns-1")},
								},
							},
						},
						{
							Object: &backuprecoveryv1.ObjectSummary{
								Name: core.StringPtr("ns-2"),
								ID:   core.Int64Ptr(102),
							},
							ArchivalInfo: &backuprecoveryv1.ArchivalRun{
								ArchivalTargetResults: []backuprecoveryv1.ArchivalTargetResult{
									{SnapshotID: core.StringPtr("snap-ns-2")},
								},
							},
						},
					},
				}, &core.DetailedResponse{StatusCode: 200}, nil)

				m.On("ConstructMetaInfoWithContext", mock.Anything, mock.MatchedBy(func(opts *backuprecoveryv1.ConstructMetaInfoOptions) bool {
					return *opts.SnapshotID == "snap-ns-1"
				})).Return(&backuprecoveryv1.ConstructMetaInfoResult{
					Environment: core.StringPtr("kKubernetes"),
					KubernetesParams: &backuprecoveryv1.ConstructMetaInfoResultKubernetesParams{
						BackedUpResourceCount: core.Int64Ptr(10),
					},
				}, &core.DetailedResponse{StatusCode: 200}, nil)

				m.On("ConstructMetaInfoWithContext", mock.Anything, mock.MatchedBy(func(opts *backuprecoveryv1.ConstructMetaInfoOptions) bool {
					return *opts.SnapshotID == "snap-ns-2"
				})).Return(&backuprecoveryv1.ConstructMetaInfoResult{
					Environment: core.StringPtr("kKubernetes"),
					KubernetesParams: &backuprecoveryv1.ConstructMetaInfoResultKubernetesParams{
						BackedUpResourceCount: core.Int64Ptr(20),
					},
				}, &core.DetailedResponse{StatusCode: 200}, nil)
			},
			expectedLen: 2,
		},
		{
			name: "NoBackupID_LatestRun_Success",
			params: &types.MetaInfoParams{
				GroupID: "group-1",
			},
			setupMocks: func(m *testmockrs.MockBRSClient) {
				m.On("GetProtectionGroupRunsWithContext", mock.Anything, mock.MatchedBy(func(opts *backuprecoveryv1.GetProtectionGroupRunsOptions) bool {
					return *opts.ID == "group-1"
				})).Return(&backuprecoveryv1.ProtectionGroupRunsResponse{
					Runs: []backuprecoveryv1.ProtectionGroupRun{
						{
							Objects: []backuprecoveryv1.ObjectRunResult{
								{
									Object: &backuprecoveryv1.ObjectSummary{
										Name: core.StringPtr("latest-ns"),
										ID:   core.Int64Ptr(201),
									},
									ArchivalInfo: &backuprecoveryv1.ArchivalRun{
										ArchivalTargetResults: []backuprecoveryv1.ArchivalTargetResult{
											{SnapshotID: core.StringPtr("snap-latest")},
										},
									},
								},
							},
						},
					},
				}, &core.DetailedResponse{StatusCode: 200}, nil)

				m.On("ConstructMetaInfoWithContext", mock.Anything, mock.MatchedBy(func(opts *backuprecoveryv1.ConstructMetaInfoOptions) bool {
					return *opts.SnapshotID == "snap-latest"
				})).Return(&backuprecoveryv1.ConstructMetaInfoResult{
					Environment: core.StringPtr("kKubernetes"),
					KubernetesParams: &backuprecoveryv1.ConstructMetaInfoResultKubernetesParams{
						BackedUpResourceCount: core.Int64Ptr(30),
					},
				}, &core.DetailedResponse{StatusCode: 200}, nil)
			},
			expectedLen: 1,
		},
		{
			name: "BRSClientError",
			params: &types.MetaInfoParams{
				GroupID:     "group-1",
				SnapshotIDs: []string{"snap-fail"},
			},
			setupMocks: func(m *testmockrs.MockBRSClient) {
				m.On("ConstructMetaInfoWithContext", mock.Anything, mock.Anything).
					Return((*backuprecoveryv1.ConstructMetaInfoResult)(nil), &core.DetailedResponse{StatusCode: 500}, fmt.Errorf("BRS internal error"))
			},
			expectedError: "failed to fetch meta info for snapshot snap-fail",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockWrapper := createMockBRSClientWrapper()
			if tt.setupMocks != nil {
				tt.setupMocks(mockWrapper.Client)
			}
			taskAPI := NewTaskAPI(mockWrapper, createTestMockConfig())

			result, err := taskAPI.GetMetaInfo(context.Background(), tt.params)
			if tt.expectedError != "" {
				assert.NotNil(t, err)
				assert.Contains(t, err.Message, tt.expectedError)
				assert.Nil(t, result)
			} else {
				assert.Nil(t, err)
				assert.Len(t, result, tt.expectedLen)
			}
		})
	}
}
