/**
 * (C) Copyright IBM Corp. 2026.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *      http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package testing

import (
	"context"
	"io"
	"net/http"
	"time"

	"github.com/IBM/go-sdk-core/v5/core"
	"github.com/IBM/ibm-backup-recovery-sdk-go/backuprecoveryv1"
	"github.com/stretchr/testify/mock"
)

// MockBRSClient provides a complete mock implementation of BRSClientInterface using testify/mock
// This provides better assertion capabilities including call verification, argument matching, and more
type MockBRSClient struct {
	mock.Mock
}

// Service configuration methods
func (m *MockBRSClient) Clone() *backuprecoveryv1.BackupRecoveryV1 {
	args := m.Called()
	if args.Get(0) == nil {
		return nil
	}
	return args.Get(0).(*backuprecoveryv1.BackupRecoveryV1)
}

func (m *MockBRSClient) SetServiceURL(url string) error {
	args := m.Called(url)
	return args.Error(0)
}

func (m *MockBRSClient) GetServiceURL() string {
	args := m.Called()
	return args.String(0)
}

func (m *MockBRSClient) SetDefaultHeaders(headers http.Header) {
	m.Called(headers)
}

func (m *MockBRSClient) SetEnableGzipCompression(enableGzip bool) {
	m.Called(enableGzip)
}

func (m *MockBRSClient) GetEnableGzipCompression() bool {
	args := m.Called()
	return args.Bool(0)
}

func (m *MockBRSClient) EnableRetries(maxRetries int, maxRetryInterval time.Duration) {
	m.Called(maxRetries, maxRetryInterval)
}

func (m *MockBRSClient) DisableRetries() {
	m.Called()
}

// Agent operations
func (m *MockBRSClient) DownloadAgent(opts *backuprecoveryv1.DownloadAgentOptions) (io.ReadCloser, *core.DetailedResponse, error) {
	args := m.Called(opts)
	if args.Get(0) == nil {
		return nil, args.Get(1).(*core.DetailedResponse), args.Error(2)
	}
	return args.Get(0).(io.ReadCloser), args.Get(1).(*core.DetailedResponse), args.Error(2)
}

func (m *MockBRSClient) DownloadAgentWithContext(ctx context.Context, opts *backuprecoveryv1.DownloadAgentOptions) (io.ReadCloser, *core.DetailedResponse, error) {
	args := m.Called(ctx, opts)
	if args.Get(0) == nil {
		return nil, args.Get(1).(*core.DetailedResponse), args.Error(2)
	}
	return args.Get(0).(io.ReadCloser), args.Get(1).(*core.DetailedResponse), args.Error(2)
}

func (m *MockBRSClient) GetUpgradeTasks(opts *backuprecoveryv1.GetUpgradeTasksOptions) (*backuprecoveryv1.AgentUpgradeTaskStates, *core.DetailedResponse, error) {
	args := m.Called(opts)
	if args.Get(0) == nil {
		return nil, args.Get(1).(*core.DetailedResponse), args.Error(2)
	}
	return args.Get(0).(*backuprecoveryv1.AgentUpgradeTaskStates), args.Get(1).(*core.DetailedResponse), args.Error(2)
}

func (m *MockBRSClient) GetUpgradeTasksWithContext(ctx context.Context, opts *backuprecoveryv1.GetUpgradeTasksOptions) (*backuprecoveryv1.AgentUpgradeTaskStates, *core.DetailedResponse, error) {
	args := m.Called(ctx, opts)
	if args.Get(0) == nil {
		return nil, args.Get(1).(*core.DetailedResponse), args.Error(2)
	}
	return args.Get(0).(*backuprecoveryv1.AgentUpgradeTaskStates), args.Get(1).(*core.DetailedResponse), args.Error(2)
}

func (m *MockBRSClient) CreateUpgradeTask(opts *backuprecoveryv1.CreateUpgradeTaskOptions) (*backuprecoveryv1.AgentUpgradeTaskState, *core.DetailedResponse, error) {
	args := m.Called(opts)
	if args.Get(0) == nil {
		return nil, args.Get(1).(*core.DetailedResponse), args.Error(2)
	}
	return args.Get(0).(*backuprecoveryv1.AgentUpgradeTaskState), args.Get(1).(*core.DetailedResponse), args.Error(2)
}

func (m *MockBRSClient) CreateUpgradeTaskWithContext(ctx context.Context, opts *backuprecoveryv1.CreateUpgradeTaskOptions) (*backuprecoveryv1.AgentUpgradeTaskState, *core.DetailedResponse, error) {
	args := m.Called(ctx, opts)
	if args.Get(0) == nil {
		return nil, args.Get(1).(*core.DetailedResponse), args.Error(2)
	}
	return args.Get(0).(*backuprecoveryv1.AgentUpgradeTaskState), args.Get(1).(*core.DetailedResponse), args.Error(2)
}

// Protection Source operations
func (m *MockBRSClient) ListProtectionSources(opts *backuprecoveryv1.ListProtectionSourcesOptions) ([]backuprecoveryv1.ProtectionSourceNodes, *core.DetailedResponse, error) {
	args := m.Called(opts)
	if args.Get(0) == nil {
		return nil, args.Get(1).(*core.DetailedResponse), args.Error(2)
	}
	return args.Get(0).([]backuprecoveryv1.ProtectionSourceNodes), args.Get(1).(*core.DetailedResponse), args.Error(2)
}

func (m *MockBRSClient) ListProtectionSourcesWithContext(ctx context.Context, opts *backuprecoveryv1.ListProtectionSourcesOptions) ([]backuprecoveryv1.ProtectionSourceNodes, *core.DetailedResponse, error) {
	args := m.Called(ctx, opts)
	if args.Get(0) == nil {
		return nil, args.Get(1).(*core.DetailedResponse), args.Error(2)
	}
	return args.Get(0).([]backuprecoveryv1.ProtectionSourceNodes), args.Get(1).(*core.DetailedResponse), args.Error(2)
}

func (m *MockBRSClient) ListProtectionSourcesRegistrationInfo(opts *backuprecoveryv1.ListProtectionSourcesRegistrationInfoOptions) (*backuprecoveryv1.GetRegistrationInfoResponse, *core.DetailedResponse, error) {
	args := m.Called(opts)
	if args.Get(0) == nil {
		return nil, args.Get(1).(*core.DetailedResponse), args.Error(2)
	}
	return args.Get(0).(*backuprecoveryv1.GetRegistrationInfoResponse), args.Get(1).(*core.DetailedResponse), args.Error(2)
}

func (m *MockBRSClient) ListProtectionSourcesRegistrationInfoWithContext(ctx context.Context, opts *backuprecoveryv1.ListProtectionSourcesRegistrationInfoOptions) (*backuprecoveryv1.GetRegistrationInfoResponse, *core.DetailedResponse, error) {
	args := m.Called(ctx, opts)
	if args.Get(0) == nil {
		return nil, args.Get(1).(*core.DetailedResponse), args.Error(2)
	}
	return args.Get(0).(*backuprecoveryv1.GetRegistrationInfoResponse), args.Get(1).(*core.DetailedResponse), args.Error(2)
}

// Connection operations
func (m *MockBRSClient) GetDataSourceConnections(opts *backuprecoveryv1.GetDataSourceConnectionsOptions) (*backuprecoveryv1.DataSourceConnectionList, *core.DetailedResponse, error) {
	args := m.Called(opts)
	var resp *backuprecoveryv1.DataSourceConnectionList
	var detailedResp *core.DetailedResponse

	if args.Get(0) != nil {
		resp = args.Get(0).(*backuprecoveryv1.DataSourceConnectionList)
	}
	if args.Get(1) != nil {
		detailedResp = args.Get(1).(*core.DetailedResponse)
	}

	return resp, detailedResp, args.Error(2)
}

func (m *MockBRSClient) GetDataSourceConnectionsWithContext(ctx context.Context, opts *backuprecoveryv1.GetDataSourceConnectionsOptions) (*backuprecoveryv1.DataSourceConnectionList, *core.DetailedResponse, error) {
	args := m.Called(ctx, opts)
	var resp *backuprecoveryv1.DataSourceConnectionList
	var detailedResp *core.DetailedResponse

	if args.Get(0) != nil {
		resp = args.Get(0).(*backuprecoveryv1.DataSourceConnectionList)
	}
	if args.Get(1) != nil {
		detailedResp = args.Get(1).(*core.DetailedResponse)
	}

	return resp, detailedResp, args.Error(2)
}

func (m *MockBRSClient) CreateDataSourceConnection(opts *backuprecoveryv1.CreateDataSourceConnectionOptions) (*backuprecoveryv1.DataSourceConnection, *core.DetailedResponse, error) {
	args := m.Called(opts)
	var resp *backuprecoveryv1.DataSourceConnection
	var detailedResp *core.DetailedResponse

	if args.Get(0) != nil {
		resp = args.Get(0).(*backuprecoveryv1.DataSourceConnection)
	}
	if args.Get(1) != nil {
		detailedResp = args.Get(1).(*core.DetailedResponse)
	}

	return resp, detailedResp, args.Error(2)
}

func (m *MockBRSClient) CreateDataSourceConnectionWithContext(ctx context.Context, opts *backuprecoveryv1.CreateDataSourceConnectionOptions) (*backuprecoveryv1.DataSourceConnection, *core.DetailedResponse, error) {
	args := m.Called(ctx, opts)
	var resp *backuprecoveryv1.DataSourceConnection
	var detailedResp *core.DetailedResponse

	if args.Get(0) != nil {
		resp = args.Get(0).(*backuprecoveryv1.DataSourceConnection)
	}
	if args.Get(1) != nil {
		detailedResp = args.Get(1).(*core.DetailedResponse)
	}

	return resp, detailedResp, args.Error(2)
}

func (m *MockBRSClient) DeleteDataSourceConnection(opts *backuprecoveryv1.DeleteDataSourceConnectionOptions) (*core.DetailedResponse, error) {
	args := m.Called(opts)
	var detailedResp *core.DetailedResponse

	if args.Get(0) != nil {
		detailedResp = args.Get(1).(*core.DetailedResponse)
	}

	return detailedResp, args.Error(1)
}

func (m *MockBRSClient) DeleteDataSourceConnectionWithContext(ctx context.Context, opts *backuprecoveryv1.DeleteDataSourceConnectionOptions) (*core.DetailedResponse, error) {
	args := m.Called(ctx, opts)
	var detailedResp *core.DetailedResponse

	if args.Get(0) != nil {
		detailedResp = args.Get(1).(*core.DetailedResponse)
	}

	return detailedResp, args.Error(1)
}

func (m *MockBRSClient) PatchDataSourceConnection(opts *backuprecoveryv1.PatchDataSourceConnectionOptions) (*backuprecoveryv1.DataSourceConnection, *core.DetailedResponse, error) {
	args := m.Called(opts)
	if args.Get(0) == nil {
		return nil, args.Get(1).(*core.DetailedResponse), args.Error(2)
	}
	return args.Get(0).(*backuprecoveryv1.DataSourceConnection), args.Get(1).(*core.DetailedResponse), args.Error(2)
}

func (m *MockBRSClient) PatchDataSourceConnectionWithContext(ctx context.Context, opts *backuprecoveryv1.PatchDataSourceConnectionOptions) (*backuprecoveryv1.DataSourceConnection, *core.DetailedResponse, error) {
	args := m.Called(ctx, opts)
	if args.Get(0) == nil {
		return nil, args.Get(1).(*core.DetailedResponse), args.Error(2)
	}
	return args.Get(0).(*backuprecoveryv1.DataSourceConnection), args.Get(1).(*core.DetailedResponse), args.Error(2)
}

func (m *MockBRSClient) GenerateDataSourceConnectionRegistrationToken(opts *backuprecoveryv1.GenerateDataSourceConnectionRegistrationTokenOptions) (*string, *core.DetailedResponse, error) {
	args := m.Called(opts)
	if args.Get(0) == nil {
		return nil, args.Get(1).(*core.DetailedResponse), args.Error(2)
	}
	return args.Get(0).(*string), args.Get(1).(*core.DetailedResponse), args.Error(2)
}

func (m *MockBRSClient) GenerateDataSourceConnectionRegistrationTokenWithContext(ctx context.Context, opts *backuprecoveryv1.GenerateDataSourceConnectionRegistrationTokenOptions) (*string, *core.DetailedResponse, error) {
	args := m.Called(ctx, opts)
	if args.Get(0) == nil {
		return nil, args.Get(1).(*core.DetailedResponse), args.Error(2)
	}
	return args.Get(0).(*string), args.Get(1).(*core.DetailedResponse), args.Error(2)
}

// Connector operations
func (m *MockBRSClient) GetDataSourceConnectors(opts *backuprecoveryv1.GetDataSourceConnectorsOptions) (*backuprecoveryv1.DataSourceConnectorList, *core.DetailedResponse, error) {
	args := m.Called(opts)
	if args.Get(0) == nil {
		return nil, args.Get(1).(*core.DetailedResponse), args.Error(2)
	}
	return args.Get(0).(*backuprecoveryv1.DataSourceConnectorList), args.Get(1).(*core.DetailedResponse), args.Error(2)
}

func (m *MockBRSClient) GetDataSourceConnectorsWithContext(ctx context.Context, opts *backuprecoveryv1.GetDataSourceConnectorsOptions) (*backuprecoveryv1.DataSourceConnectorList, *core.DetailedResponse, error) {
	args := m.Called(ctx, opts)
	if args.Get(0) == nil {
		return nil, args.Get(1).(*core.DetailedResponse), args.Error(2)
	}
	return args.Get(0).(*backuprecoveryv1.DataSourceConnectorList), args.Get(1).(*core.DetailedResponse), args.Error(2)
}

func (m *MockBRSClient) GetConnectorMetadata(opts *backuprecoveryv1.GetConnectorMetadataOptions) (*backuprecoveryv1.ConnectorMetadata, *core.DetailedResponse, error) {
	args := m.Called(opts)
	if args.Get(0) == nil {
		return nil, args.Get(1).(*core.DetailedResponse), args.Error(2)
	}
	return args.Get(0).(*backuprecoveryv1.ConnectorMetadata), args.Get(1).(*core.DetailedResponse), args.Error(2)
}

func (m *MockBRSClient) GetConnectorMetadataWithContext(ctx context.Context, opts *backuprecoveryv1.GetConnectorMetadataOptions) (*backuprecoveryv1.ConnectorMetadata, *core.DetailedResponse, error) {
	args := m.Called(ctx, opts)
	if args.Get(0) == nil {
		return nil, args.Get(1).(*core.DetailedResponse), args.Error(2)
	}
	return args.Get(0).(*backuprecoveryv1.ConnectorMetadata), args.Get(1).(*core.DetailedResponse), args.Error(2)
}

func (m *MockBRSClient) DeleteDataSourceConnector(opts *backuprecoveryv1.DeleteDataSourceConnectorOptions) (*core.DetailedResponse, error) {
	args := m.Called(opts)
	return args.Get(0).(*core.DetailedResponse), args.Error(1)
}

func (m *MockBRSClient) DeleteDataSourceConnectorWithContext(ctx context.Context, opts *backuprecoveryv1.DeleteDataSourceConnectorOptions) (*core.DetailedResponse, error) {
	args := m.Called(ctx, opts)
	return args.Get(0).(*core.DetailedResponse), args.Error(1)
}

func (m *MockBRSClient) PatchDataSourceConnector(opts *backuprecoveryv1.PatchDataSourceConnectorOptions) (*backuprecoveryv1.DataSourceConnector, *core.DetailedResponse, error) {
	args := m.Called(opts)
	if args.Get(0) == nil {
		return nil, args.Get(1).(*core.DetailedResponse), args.Error(2)
	}
	return args.Get(0).(*backuprecoveryv1.DataSourceConnector), args.Get(1).(*core.DetailedResponse), args.Error(2)
}

func (m *MockBRSClient) PatchDataSourceConnectorWithContext(ctx context.Context, opts *backuprecoveryv1.PatchDataSourceConnectorOptions) (*backuprecoveryv1.DataSourceConnector, *core.DetailedResponse, error) {
	args := m.Called(ctx, opts)
	if args.Get(0) == nil {
		return nil, args.Get(1).(*core.DetailedResponse), args.Error(2)
	}
	return args.Get(0).(*backuprecoveryv1.DataSourceConnector), args.Get(1).(*core.DetailedResponse), args.Error(2)
}

// Object Snapshot operations
func (m *MockBRSClient) GetObjectSnapshots(opts *backuprecoveryv1.GetObjectSnapshotsOptions) (*backuprecoveryv1.GetObjectSnapshotsResponse, *core.DetailedResponse, error) {
	args := m.Called(opts)
	if args.Get(0) == nil {
		return nil, args.Get(1).(*core.DetailedResponse), args.Error(2)
	}
	return args.Get(0).(*backuprecoveryv1.GetObjectSnapshotsResponse), args.Get(1).(*core.DetailedResponse), args.Error(2)
}

func (m *MockBRSClient) GetObjectSnapshotsWithContext(ctx context.Context, opts *backuprecoveryv1.GetObjectSnapshotsOptions) (*backuprecoveryv1.GetObjectSnapshotsResponse, *core.DetailedResponse, error) {
	args := m.Called(ctx, opts)
	if args.Get(0) == nil {
		return nil, args.Get(1).(*core.DetailedResponse), args.Error(2)
	}
	return args.Get(0).(*backuprecoveryv1.GetObjectSnapshotsResponse), args.Get(1).(*core.DetailedResponse), args.Error(2)
}

// Protection Policy operations
func (m *MockBRSClient) GetProtectionPolicies(opts *backuprecoveryv1.GetProtectionPoliciesOptions) (*backuprecoveryv1.ProtectionPoliciesResponse, *core.DetailedResponse, error) {
	args := m.Called(opts)
	var resp *backuprecoveryv1.ProtectionPoliciesResponse
	var detailedResp *core.DetailedResponse

	if args.Get(0) != nil {
		resp = args.Get(0).(*backuprecoveryv1.ProtectionPoliciesResponse)
	}
	if args.Get(1) != nil {
		detailedResp = args.Get(1).(*core.DetailedResponse)
	}

	return resp, detailedResp, args.Error(2)
}

func (m *MockBRSClient) GetProtectionPoliciesWithContext(ctx context.Context, opts *backuprecoveryv1.GetProtectionPoliciesOptions) (*backuprecoveryv1.ProtectionPoliciesResponse, *core.DetailedResponse, error) {
	args := m.Called(ctx, opts)
	var resp *backuprecoveryv1.ProtectionPoliciesResponse
	var detailedResp *core.DetailedResponse

	if args.Get(0) != nil {
		resp = args.Get(0).(*backuprecoveryv1.ProtectionPoliciesResponse)
	}
	if args.Get(1) != nil {
		detailedResp = args.Get(1).(*core.DetailedResponse)
	}

	return resp, detailedResp, args.Error(2)
}

func (m *MockBRSClient) CreateProtectionPolicy(opts *backuprecoveryv1.CreateProtectionPolicyOptions) (*backuprecoveryv1.ProtectionPolicyResponse, *core.DetailedResponse, error) {
	args := m.Called(opts)
	var resp *backuprecoveryv1.ProtectionPolicyResponse
	var detailedResp *core.DetailedResponse

	if args.Get(0) != nil {
		resp = args.Get(0).(*backuprecoveryv1.ProtectionPolicyResponse)
	}
	if args.Get(1) != nil {
		detailedResp = args.Get(1).(*core.DetailedResponse)
	}

	return resp, detailedResp, args.Error(2)
}

func (m *MockBRSClient) CreateProtectionPolicyWithContext(ctx context.Context, opts *backuprecoveryv1.CreateProtectionPolicyOptions) (*backuprecoveryv1.ProtectionPolicyResponse, *core.DetailedResponse, error) {
	args := m.Called(ctx, opts)
	var resp *backuprecoveryv1.ProtectionPolicyResponse
	var detailedResp *core.DetailedResponse

	if args.Get(0) != nil {
		resp = args.Get(0).(*backuprecoveryv1.ProtectionPolicyResponse)
	}
	if args.Get(1) != nil {
		detailedResp = args.Get(1).(*core.DetailedResponse)
	}

	return resp, detailedResp, args.Error(2)
}

func (m *MockBRSClient) GetProtectionPolicyByID(opts *backuprecoveryv1.GetProtectionPolicyByIdOptions) (*backuprecoveryv1.ProtectionPolicyResponse, *core.DetailedResponse, error) {
	args := m.Called(opts)
	var resp *backuprecoveryv1.ProtectionPolicyResponse
	var detailedResp *core.DetailedResponse

	if args.Get(0) != nil {
		resp = args.Get(0).(*backuprecoveryv1.ProtectionPolicyResponse)
	}
	if args.Get(1) != nil {
		detailedResp = args.Get(1).(*core.DetailedResponse)
	}

	return resp, detailedResp, args.Error(2)
}

func (m *MockBRSClient) GetProtectionPolicyByIDWithContext(ctx context.Context, opts *backuprecoveryv1.GetProtectionPolicyByIdOptions) (*backuprecoveryv1.ProtectionPolicyResponse, *core.DetailedResponse, error) {
	args := m.Called(ctx, opts)
	var resp *backuprecoveryv1.ProtectionPolicyResponse
	var detailedResp *core.DetailedResponse

	if args.Get(0) != nil {
		resp = args.Get(0).(*backuprecoveryv1.ProtectionPolicyResponse)
	}
	if args.Get(1) != nil {
		detailedResp = args.Get(1).(*core.DetailedResponse)
	}

	return resp, detailedResp, args.Error(2)
}

func (m *MockBRSClient) UpdateProtectionPolicy(opts *backuprecoveryv1.UpdateProtectionPolicyOptions) (*backuprecoveryv1.ProtectionPolicyResponse, *core.DetailedResponse, error) {
	args := m.Called(opts)
	var resp *backuprecoveryv1.ProtectionPolicyResponse
	var detailedResp *core.DetailedResponse

	if args.Get(0) != nil {
		resp = args.Get(0).(*backuprecoveryv1.ProtectionPolicyResponse)
	}
	if args.Get(1) != nil {
		detailedResp = args.Get(1).(*core.DetailedResponse)
	}

	return resp, detailedResp, args.Error(2)
}

func (m *MockBRSClient) UpdateProtectionPolicyWithContext(ctx context.Context, opts *backuprecoveryv1.UpdateProtectionPolicyOptions) (*backuprecoveryv1.ProtectionPolicyResponse, *core.DetailedResponse, error) {
	args := m.Called(ctx, opts)
	var resp *backuprecoveryv1.ProtectionPolicyResponse
	var detailedResp *core.DetailedResponse

	if args.Get(0) != nil {
		resp = args.Get(0).(*backuprecoveryv1.ProtectionPolicyResponse)
	}
	if args.Get(1) != nil {
		detailedResp = args.Get(1).(*core.DetailedResponse)
	}

	return resp, detailedResp, args.Error(2)
}

func (m *MockBRSClient) DeleteProtectionPolicy(opts *backuprecoveryv1.DeleteProtectionPolicyOptions) (*core.DetailedResponse, error) {
	args := m.Called(opts)
	var detailedResp *core.DetailedResponse

	if args.Get(0) != nil {
		detailedResp = args.Get(1).(*core.DetailedResponse)
	}

	return detailedResp, args.Error(1)
}

func (m *MockBRSClient) DeleteProtectionPolicyWithContext(ctx context.Context, opts *backuprecoveryv1.DeleteProtectionPolicyOptions) (*core.DetailedResponse, error) {
	args := m.Called(ctx, opts)
	var detailedResp *core.DetailedResponse

	if args.Get(0) != nil {
		detailedResp = args.Get(1).(*core.DetailedResponse)
	}

	return detailedResp, args.Error(1)
}

// Protection Group operations
func (m *MockBRSClient) GetProtectionGroups(opts *backuprecoveryv1.GetProtectionGroupsOptions) (*backuprecoveryv1.ProtectionGroupsResponse, *core.DetailedResponse, error) {
	args := m.Called(opts)

	var resp *backuprecoveryv1.ProtectionGroupsResponse
	var detailedResp *core.DetailedResponse

	if args.Get(0) != nil {
		resp = args.Get(0).(*backuprecoveryv1.ProtectionGroupsResponse)
	}
	if args.Get(1) != nil {
		detailedResp = args.Get(1).(*core.DetailedResponse)
	}

	return resp, detailedResp, args.Error(2)
}

func (m *MockBRSClient) GetProtectionGroupsWithContext(ctx context.Context, opts *backuprecoveryv1.GetProtectionGroupsOptions) (*backuprecoveryv1.ProtectionGroupsResponse, *core.DetailedResponse, error) {
	args := m.Called(ctx, opts)
	var resp *backuprecoveryv1.ProtectionGroupsResponse
	var detailedResp *core.DetailedResponse

	if args.Get(0) != nil {
		resp = args.Get(0).(*backuprecoveryv1.ProtectionGroupsResponse)
	}
	if args.Get(1) != nil {
		detailedResp = args.Get(1).(*core.DetailedResponse)
	}

	return resp, detailedResp, args.Error(2)
}

func (m *MockBRSClient) CreateProtectionGroup(opts *backuprecoveryv1.CreateProtectionGroupOptions) (*backuprecoveryv1.ProtectionGroupResponse, *core.DetailedResponse, error) {
	args := m.Called(opts)
	var resp *backuprecoveryv1.ProtectionGroupResponse
	var detailedResp *core.DetailedResponse

	if args.Get(0) != nil {
		resp = args.Get(0).(*backuprecoveryv1.ProtectionGroupResponse)
	}
	if args.Get(1) != nil {
		detailedResp = args.Get(1).(*core.DetailedResponse)
	}

	return resp, detailedResp, args.Error(2)

}

func (m *MockBRSClient) CreateProtectionGroupWithContext(ctx context.Context, opts *backuprecoveryv1.CreateProtectionGroupOptions) (*backuprecoveryv1.ProtectionGroupResponse, *core.DetailedResponse, error) {
	args := m.Called(ctx, opts)
	var resp *backuprecoveryv1.ProtectionGroupResponse
	var detailedResp *core.DetailedResponse

	if args.Get(0) != nil {
		resp = args.Get(0).(*backuprecoveryv1.ProtectionGroupResponse)
	}
	if args.Get(1) != nil {
		detailedResp = args.Get(1).(*core.DetailedResponse)
	}

	return resp, detailedResp, args.Error(2)
}

func (m *MockBRSClient) GetProtectionGroupByID(opts *backuprecoveryv1.GetProtectionGroupByIdOptions) (*backuprecoveryv1.ProtectionGroupResponse, *core.DetailedResponse, error) {
	args := m.Called(opts)
	var resp *backuprecoveryv1.ProtectionGroupResponse
	var detailedResp *core.DetailedResponse

	if args.Get(0) != nil {
		resp = args.Get(0).(*backuprecoveryv1.ProtectionGroupResponse)
	}
	if args.Get(1) != nil {
		detailedResp = args.Get(1).(*core.DetailedResponse)
	}

	return resp, detailedResp, args.Error(2)
}

func (m *MockBRSClient) GetProtectionGroupByIDWithContext(ctx context.Context, opts *backuprecoveryv1.GetProtectionGroupByIdOptions) (*backuprecoveryv1.ProtectionGroupResponse, *core.DetailedResponse, error) {
	args := m.Called(ctx, opts)
	var resp *backuprecoveryv1.ProtectionGroupResponse
	var detailedResp *core.DetailedResponse

	if args.Get(0) != nil {
		resp = args.Get(0).(*backuprecoveryv1.ProtectionGroupResponse)
	}
	if args.Get(1) != nil {
		detailedResp = args.Get(1).(*core.DetailedResponse)
	}

	return resp, detailedResp, args.Error(2)
}

func (m *MockBRSClient) UpdateProtectionGroup(opts *backuprecoveryv1.UpdateProtectionGroupOptions) (*backuprecoveryv1.ProtectionGroupResponse, *core.DetailedResponse, error) {
	args := m.Called(opts)
	var resp *backuprecoveryv1.ProtectionGroupResponse
	var detailedResp *core.DetailedResponse

	if args.Get(0) != nil {
		resp = args.Get(0).(*backuprecoveryv1.ProtectionGroupResponse)
	}
	if args.Get(1) != nil {
		detailedResp = args.Get(1).(*core.DetailedResponse)
	}

	return resp, detailedResp, args.Error(2)
}

func (m *MockBRSClient) UpdateProtectionGroupWithContext(ctx context.Context, opts *backuprecoveryv1.UpdateProtectionGroupOptions) (*backuprecoveryv1.ProtectionGroupResponse, *core.DetailedResponse, error) {
	args := m.Called(ctx, opts)
	var resp *backuprecoveryv1.ProtectionGroupResponse
	var detailedResp *core.DetailedResponse

	if args.Get(0) != nil {
		resp = args.Get(0).(*backuprecoveryv1.ProtectionGroupResponse)
	}
	if args.Get(1) != nil {
		detailedResp = args.Get(1).(*core.DetailedResponse)
	}

	return resp, detailedResp, args.Error(2)
}

func (m *MockBRSClient) DeleteProtectionGroup(opts *backuprecoveryv1.DeleteProtectionGroupOptions) (*core.DetailedResponse, error) {
	args := m.Called(opts)
	var detailedResp *core.DetailedResponse
	if args.Get(0) != nil {
		detailedResp = args.Get(0).(*core.DetailedResponse)
	}
	return detailedResp, args.Error(1)
}

func (m *MockBRSClient) DeleteProtectionGroupWithContext(ctx context.Context, opts *backuprecoveryv1.DeleteProtectionGroupOptions) (*core.DetailedResponse, error) {
	args := m.Called(ctx, opts)
	var detailedResp *core.DetailedResponse
	if args.Get(0) != nil {
		detailedResp = args.Get(0).(*core.DetailedResponse)
	}
	return detailedResp, args.Error(1)
}

// Protection Group Run operations
func (m *MockBRSClient) GetProtectionGroupRuns(opts *backuprecoveryv1.GetProtectionGroupRunsOptions) (*backuprecoveryv1.ProtectionGroupRunsResponse, *core.DetailedResponse, error) {
	args := m.Called(opts)
	var resp *backuprecoveryv1.ProtectionGroupRunsResponse
	var detailedResp *core.DetailedResponse

	if args.Get(0) != nil {
		resp = args.Get(0).(*backuprecoveryv1.ProtectionGroupRunsResponse)
	}
	if args.Get(1) != nil {
		detailedResp = args.Get(1).(*core.DetailedResponse)
	}

	return resp, detailedResp, args.Error(2)
}

func (m *MockBRSClient) GetProtectionGroupRunsWithContext(ctx context.Context, opts *backuprecoveryv1.GetProtectionGroupRunsOptions) (*backuprecoveryv1.ProtectionGroupRunsResponse, *core.DetailedResponse, error) {
	args := m.Called(ctx, opts)
	var resp *backuprecoveryv1.ProtectionGroupRunsResponse
	var detailedResp *core.DetailedResponse

	if args.Get(0) != nil {
		resp = args.Get(0).(*backuprecoveryv1.ProtectionGroupRunsResponse)
	}
	if args.Get(1) != nil {
		detailedResp = args.Get(1).(*core.DetailedResponse)
	}

	return resp, detailedResp, args.Error(2)
}

func (m *MockBRSClient) UpdateProtectionGroupRun(opts *backuprecoveryv1.UpdateProtectionGroupRunOptions) (*backuprecoveryv1.UpdateProtectionGroupRunResponse, *core.DetailedResponse, error) {
	args := m.Called(opts)
	var resp *backuprecoveryv1.UpdateProtectionGroupRunResponse
	var detailedResp *core.DetailedResponse

	if args.Get(0) != nil {
		resp = args.Get(0).(*backuprecoveryv1.UpdateProtectionGroupRunResponse)
	}
	if args.Get(1) != nil {
		detailedResp = args.Get(1).(*core.DetailedResponse)
	}

	return resp, detailedResp, args.Error(2)
}

func (m *MockBRSClient) UpdateProtectionGroupRunWithContext(ctx context.Context, opts *backuprecoveryv1.UpdateProtectionGroupRunOptions) (*backuprecoveryv1.UpdateProtectionGroupRunResponse, *core.DetailedResponse, error) {
	args := m.Called(ctx, opts)
	var resp *backuprecoveryv1.UpdateProtectionGroupRunResponse
	var detailedResp *core.DetailedResponse

	if args.Get(0) != nil {
		resp = args.Get(0).(*backuprecoveryv1.UpdateProtectionGroupRunResponse)
	}
	if args.Get(1) != nil {
		detailedResp = args.Get(1).(*core.DetailedResponse)
	}

	return resp, detailedResp, args.Error(2)
}

func (m *MockBRSClient) CreateProtectionGroupRun(opts *backuprecoveryv1.CreateProtectionGroupRunOptions) (*backuprecoveryv1.CreateProtectionGroupRunResponse, *core.DetailedResponse, error) {
	args := m.Called(opts)

	var resp *backuprecoveryv1.CreateProtectionGroupRunResponse
	var detailedResp *core.DetailedResponse

	if args.Get(0) != nil {
		resp = args.Get(0).(*backuprecoveryv1.CreateProtectionGroupRunResponse)
	}
	if args.Get(1) != nil {
		detailedResp = args.Get(1).(*core.DetailedResponse)
	}

	return resp, detailedResp, args.Error(2)
}

func (m *MockBRSClient) CreateProtectionGroupRunWithContext(ctx context.Context, opts *backuprecoveryv1.CreateProtectionGroupRunOptions) (*backuprecoveryv1.CreateProtectionGroupRunResponse, *core.DetailedResponse, error) {
	args := m.Called(ctx, opts)
	var resp *backuprecoveryv1.CreateProtectionGroupRunResponse
	var detailedResp *core.DetailedResponse

	if args.Get(0) != nil {
		resp = args.Get(0).(*backuprecoveryv1.CreateProtectionGroupRunResponse)
	}
	if args.Get(1) != nil {
		detailedResp = args.Get(1).(*core.DetailedResponse)
	}

	return resp, detailedResp, args.Error(2)
}

func (m *MockBRSClient) PerformActionOnProtectionGroupRun(opts *backuprecoveryv1.PerformActionOnProtectionGroupRunOptions) (*backuprecoveryv1.PerformRunActionResponse, *core.DetailedResponse, error) {
	args := m.Called(opts)
	var resp *backuprecoveryv1.PerformRunActionResponse
	var detailedResp *core.DetailedResponse

	if args.Get(0) != nil {
		resp = args.Get(0).(*backuprecoveryv1.PerformRunActionResponse)
	}
	if args.Get(1) != nil {
		detailedResp = args.Get(1).(*core.DetailedResponse)
	}

	return resp, detailedResp, args.Error(2)
}

func (m *MockBRSClient) PerformActionOnProtectionGroupRunWithContext(ctx context.Context, opts *backuprecoveryv1.PerformActionOnProtectionGroupRunOptions) (*backuprecoveryv1.PerformRunActionResponse, *core.DetailedResponse, error) {
	args := m.Called(ctx, opts)
	var resp *backuprecoveryv1.PerformRunActionResponse
	var detailedResp *core.DetailedResponse

	if args.Get(0) != nil {
		resp = args.Get(0).(*backuprecoveryv1.PerformRunActionResponse)
	}
	if args.Get(1) != nil {
		detailedResp = args.Get(1).(*core.DetailedResponse)
	}

	return resp, detailedResp, args.Error(2)
}

func (m *MockBRSClient) GetProtectionGroupRun(opts *backuprecoveryv1.GetProtectionGroupRunOptions) (*backuprecoveryv1.ProtectionGroupRun, *core.DetailedResponse, error) {
	args := m.Called(opts)
	var resp *backuprecoveryv1.ProtectionGroupRun
	var detailedResp *core.DetailedResponse

	if args.Get(0) != nil {
		resp = args.Get(0).(*backuprecoveryv1.ProtectionGroupRun)
	}
	if args.Get(1) != nil {
		detailedResp = args.Get(1).(*core.DetailedResponse)
	}

	return resp, detailedResp, args.Error(2)
}

func (m *MockBRSClient) GetProtectionGroupRunWithContext(ctx context.Context, opts *backuprecoveryv1.GetProtectionGroupRunOptions) (*backuprecoveryv1.ProtectionGroupRun, *core.DetailedResponse, error) {
	args := m.Called(ctx, opts)
	var resp *backuprecoveryv1.ProtectionGroupRun
	var detailedResp *core.DetailedResponse

	if args.Get(0) != nil {
		resp = args.Get(0).(*backuprecoveryv1.ProtectionGroupRun)
	}
	if args.Get(1) != nil {
		detailedResp = args.Get(1).(*core.DetailedResponse)
	}

	return resp, detailedResp, args.Error(2)
}

// Recovery operations
func (m *MockBRSClient) GetRecoveries(opts *backuprecoveryv1.GetRecoveriesOptions) (*backuprecoveryv1.RecoveriesResponse, *core.DetailedResponse, error) {
	args := m.Called(opts)
	var resp *backuprecoveryv1.RecoveriesResponse
	var detailedResp *core.DetailedResponse

	if args.Get(0) != nil {
		resp = args.Get(0).(*backuprecoveryv1.RecoveriesResponse)
	}
	if args.Get(1) != nil {
		detailedResp = args.Get(1).(*core.DetailedResponse)
	}

	return resp, detailedResp, args.Error(2)
}

func (m *MockBRSClient) GetRecoveriesWithContext(ctx context.Context, opts *backuprecoveryv1.GetRecoveriesOptions) (*backuprecoveryv1.RecoveriesResponse, *core.DetailedResponse, error) {
	args := m.Called(ctx, opts)
	var resp *backuprecoveryv1.RecoveriesResponse
	var detailedResp *core.DetailedResponse

	if args.Get(0) != nil {
		resp = args.Get(0).(*backuprecoveryv1.RecoveriesResponse)
	}
	if args.Get(1) != nil {
		detailedResp = args.Get(1).(*core.DetailedResponse)
	}

	return resp, detailedResp, args.Error(2)
}

func (m *MockBRSClient) CreateRecovery(opts *backuprecoveryv1.CreateRecoveryOptions) (*backuprecoveryv1.Recovery, *core.DetailedResponse, error) {
	args := m.Called(opts)
	var resp *backuprecoveryv1.Recovery
	var detailedResp *core.DetailedResponse

	if args.Get(0) != nil {
		resp = args.Get(0).(*backuprecoveryv1.Recovery)
	}
	if args.Get(1) != nil {
		detailedResp = args.Get(1).(*core.DetailedResponse)
	}

	return resp, detailedResp, args.Error(2)
}

func (m *MockBRSClient) CreateRecoveryWithContext(ctx context.Context, opts *backuprecoveryv1.CreateRecoveryOptions) (*backuprecoveryv1.Recovery, *core.DetailedResponse, error) {
	args := m.Called(ctx, opts)
	var resp *backuprecoveryv1.Recovery
	var detailedResp *core.DetailedResponse

	if args.Get(0) != nil {
		resp = args.Get(0).(*backuprecoveryv1.Recovery)
	}
	if args.Get(1) != nil {
		detailedResp = args.Get(1).(*core.DetailedResponse)
	}

	return resp, detailedResp, args.Error(2)
}

func (m *MockBRSClient) CreateDownloadFilesAndFoldersRecovery(opts *backuprecoveryv1.CreateDownloadFilesAndFoldersRecoveryOptions) (*backuprecoveryv1.Recovery, *core.DetailedResponse, error) {
	args := m.Called(opts)
	var resp *backuprecoveryv1.Recovery
	var detailedResp *core.DetailedResponse

	if args.Get(0) != nil {
		resp = args.Get(0).(*backuprecoveryv1.Recovery)
	}
	if args.Get(1) != nil {
		detailedResp = args.Get(1).(*core.DetailedResponse)
	}

	return resp, detailedResp, args.Error(2)
}

func (m *MockBRSClient) CreateDownloadFilesAndFoldersRecoveryWithContext(ctx context.Context, opts *backuprecoveryv1.CreateDownloadFilesAndFoldersRecoveryOptions) (*backuprecoveryv1.Recovery, *core.DetailedResponse, error) {
	args := m.Called(ctx, opts)
	var resp *backuprecoveryv1.Recovery
	var detailedResp *core.DetailedResponse

	if args.Get(0) != nil {
		resp = args.Get(0).(*backuprecoveryv1.Recovery)
	}
	if args.Get(1) != nil {
		detailedResp = args.Get(1).(*core.DetailedResponse)
	}

	return resp, detailedResp, args.Error(2)
}

func (m *MockBRSClient) GetRecoveryByID(opts *backuprecoveryv1.GetRecoveryByIdOptions) (*backuprecoveryv1.Recovery, *core.DetailedResponse, error) {
	args := m.Called(opts)
	var resp *backuprecoveryv1.Recovery
	var detailedResp *core.DetailedResponse

	if args.Get(0) != nil {
		resp = args.Get(0).(*backuprecoveryv1.Recovery)
	}
	if args.Get(1) != nil {
		detailedResp = args.Get(1).(*core.DetailedResponse)
	}

	return resp, detailedResp, args.Error(2)
}

func (m *MockBRSClient) GetRecoveryByIDWithContext(ctx context.Context, opts *backuprecoveryv1.GetRecoveryByIdOptions) (*backuprecoveryv1.Recovery, *core.DetailedResponse, error) {
	args := m.Called(ctx, opts)
	var resp *backuprecoveryv1.Recovery
	var detailedResp *core.DetailedResponse

	if args.Get(0) != nil {
		resp = args.Get(0).(*backuprecoveryv1.Recovery)
	}
	if args.Get(1) != nil {
		detailedResp = args.Get(1).(*core.DetailedResponse)
	}

	return resp, detailedResp, args.Error(2)
}

func (m *MockBRSClient) DownloadFilesFromRecovery(opts *backuprecoveryv1.DownloadFilesFromRecoveryOptions) (*core.DetailedResponse, error) {
	args := m.Called(opts)
	return args.Get(0).(*core.DetailedResponse), args.Error(1)
}

func (m *MockBRSClient) DownloadFilesFromRecoveryWithContext(ctx context.Context, opts *backuprecoveryv1.DownloadFilesFromRecoveryOptions) (*core.DetailedResponse, error) {
	args := m.Called(ctx, opts)
	return args.Get(0).(*core.DetailedResponse), args.Error(1)
}

func (m *MockBRSClient) CancelRecoveryByID(opts *backuprecoveryv1.CancelRecoveryByIdOptions) (*core.DetailedResponse, error) {
	args := m.Called(opts)
	var detailedResp *core.DetailedResponse

	if args.Get(0) != nil {
		detailedResp = args.Get(0).(*core.DetailedResponse)
	}

	return detailedResp, args.Error(1)
}

func (m *MockBRSClient) CancelRecoveryByIDWithContext(ctx context.Context, opts *backuprecoveryv1.CancelRecoveryByIdOptions) (*core.DetailedResponse, error) {
	args := m.Called(ctx, opts)
	var detailedResp *core.DetailedResponse

	if args.Get(0) != nil {
		detailedResp = args.Get(0).(*core.DetailedResponse)
	}

	return detailedResp, args.Error(1)
}

// Restore Point operations
func (m *MockBRSClient) GetRestorePointsInTimeRange(opts *backuprecoveryv1.GetRestorePointsInTimeRangeOptions) (*backuprecoveryv1.GetRestorePointsInTimeRangeResponse, *core.DetailedResponse, error) {
	args := m.Called(opts)
	if args.Get(0) == nil {
		return nil, args.Get(1).(*core.DetailedResponse), args.Error(2)
	}
	return args.Get(0).(*backuprecoveryv1.GetRestorePointsInTimeRangeResponse), args.Get(1).(*core.DetailedResponse), args.Error(2)
}

func (m *MockBRSClient) GetRestorePointsInTimeRangeWithContext(ctx context.Context, opts *backuprecoveryv1.GetRestorePointsInTimeRangeOptions) (*backuprecoveryv1.GetRestorePointsInTimeRangeResponse, *core.DetailedResponse, error) {
	args := m.Called(ctx, opts)
	if args.Get(0) == nil {
		return nil, args.Get(1).(*core.DetailedResponse), args.Error(2)
	}
	return args.Get(0).(*backuprecoveryv1.GetRestorePointsInTimeRangeResponse), args.Get(1).(*core.DetailedResponse), args.Error(2)
}

// Indexed File operations
func (m *MockBRSClient) DownloadIndexedFile(opts *backuprecoveryv1.DownloadIndexedFileOptions) (*core.DetailedResponse, error) {
	args := m.Called(opts)
	return args.Get(0).(*core.DetailedResponse), args.Error(1)
}

func (m *MockBRSClient) DownloadIndexedFileWithContext(ctx context.Context, opts *backuprecoveryv1.DownloadIndexedFileOptions) (*core.DetailedResponse, error) {
	args := m.Called(ctx, opts)
	return args.Get(0).(*core.DetailedResponse), args.Error(1)
}

func (m *MockBRSClient) SearchIndexedObjects(opts *backuprecoveryv1.SearchIndexedObjectsOptions) (*backuprecoveryv1.SearchIndexedObjectsResponse, *core.DetailedResponse, error) {
	args := m.Called(opts)
	if args.Get(0) == nil {
		return nil, args.Get(1).(*core.DetailedResponse), args.Error(2)
	}
	return args.Get(0).(*backuprecoveryv1.SearchIndexedObjectsResponse), args.Get(1).(*core.DetailedResponse), args.Error(2)
}

func (m *MockBRSClient) SearchIndexedObjectsWithContext(ctx context.Context, opts *backuprecoveryv1.SearchIndexedObjectsOptions) (*backuprecoveryv1.SearchIndexedObjectsResponse, *core.DetailedResponse, error) {
	args := m.Called(ctx, opts)
	if args.Get(0) == nil {
		return nil, args.Get(1).(*core.DetailedResponse), args.Error(2)
	}
	return args.Get(0).(*backuprecoveryv1.SearchIndexedObjectsResponse), args.Get(1).(*core.DetailedResponse), args.Error(2)
}

// Object Search operations
func (m *MockBRSClient) SearchObjects(opts *backuprecoveryv1.SearchObjectsOptions) (*backuprecoveryv1.ObjectsSearchResponseBody, *core.DetailedResponse, error) {
	args := m.Called(opts)
	if args.Get(0) == nil {
		return nil, args.Get(1).(*core.DetailedResponse), args.Error(2)
	}
	return args.Get(0).(*backuprecoveryv1.ObjectsSearchResponseBody), args.Get(1).(*core.DetailedResponse), args.Error(2)
}

func (m *MockBRSClient) SearchObjectsWithContext(ctx context.Context, opts *backuprecoveryv1.SearchObjectsOptions) (*backuprecoveryv1.ObjectsSearchResponseBody, *core.DetailedResponse, error) {
	args := m.Called(ctx, opts)
	if args.Get(0) == nil {
		return nil, args.Get(1).(*core.DetailedResponse), args.Error(2)
	}
	return args.Get(0).(*backuprecoveryv1.ObjectsSearchResponseBody), args.Get(1).(*core.DetailedResponse), args.Error(2)
}

func (m *MockBRSClient) SearchProtectedObjects(opts *backuprecoveryv1.SearchProtectedObjectsOptions) (*backuprecoveryv1.ProtectedObjectsSearchResponse, *core.DetailedResponse, error) {
	args := m.Called(opts)
	if args.Get(0) == nil {
		return nil, args.Get(1).(*core.DetailedResponse), args.Error(2)
	}
	return args.Get(0).(*backuprecoveryv1.ProtectedObjectsSearchResponse), args.Get(1).(*core.DetailedResponse), args.Error(2)
}

func (m *MockBRSClient) SearchProtectedObjectsWithContext(ctx context.Context, opts *backuprecoveryv1.SearchProtectedObjectsOptions) (*backuprecoveryv1.ProtectedObjectsSearchResponse, *core.DetailedResponse, error) {
	args := m.Called(ctx, opts)
	if args.Get(0) == nil {
		return nil, args.Get(1).(*core.DetailedResponse), args.Error(2)
	}
	return args.Get(0).(*backuprecoveryv1.ProtectedObjectsSearchResponse), args.Get(1).(*core.DetailedResponse), args.Error(2)
}

// Source Registration operations
func (m *MockBRSClient) GetSourceRegistrations(opts *backuprecoveryv1.GetSourceRegistrationsOptions) (*backuprecoveryv1.SourceRegistrations, *core.DetailedResponse, error) {
	args := m.Called(opts)
	var resp *backuprecoveryv1.SourceRegistrations
	var detailedResp *core.DetailedResponse

	if args.Get(0) != nil {
		resp = args.Get(0).(*backuprecoveryv1.SourceRegistrations)
	}
	if args.Get(1) != nil {
		detailedResp = args.Get(1).(*core.DetailedResponse)
	}

	return resp, detailedResp, args.Error(2)
}

func (m *MockBRSClient) GetSourceRegistrationsWithContext(ctx context.Context, opts *backuprecoveryv1.GetSourceRegistrationsOptions) (*backuprecoveryv1.SourceRegistrations, *core.DetailedResponse, error) {
	args := m.Called(ctx, opts)
	var resp *backuprecoveryv1.SourceRegistrations
	var detailedResp *core.DetailedResponse

	if args.Get(0) != nil {
		resp = args.Get(0).(*backuprecoveryv1.SourceRegistrations)
	}
	if args.Get(1) != nil {
		detailedResp = args.Get(1).(*core.DetailedResponse)
	}

	return resp, detailedResp, args.Error(2)
}

func (m *MockBRSClient) RegisterProtectionSource(opts *backuprecoveryv1.RegisterProtectionSourceOptions) (*backuprecoveryv1.SourceRegistrationResponseParams, *core.DetailedResponse, error) {
	args := m.Called(opts)
	var resp *backuprecoveryv1.SourceRegistrationResponseParams
	var detailedResp *core.DetailedResponse

	if args.Get(0) != nil {
		resp = args.Get(0).(*backuprecoveryv1.SourceRegistrationResponseParams)
	}
	if args.Get(1) != nil {
		detailedResp = args.Get(1).(*core.DetailedResponse)
	}

	return resp, detailedResp, args.Error(2)
}

func (m *MockBRSClient) RegisterProtectionSourceWithContext(ctx context.Context, opts *backuprecoveryv1.RegisterProtectionSourceOptions) (*backuprecoveryv1.SourceRegistrationResponseParams, *core.DetailedResponse, error) {
	args := m.Called(ctx, opts)
	var resp *backuprecoveryv1.SourceRegistrationResponseParams
	var detailedResp *core.DetailedResponse

	if args.Get(0) != nil {
		resp = args.Get(0).(*backuprecoveryv1.SourceRegistrationResponseParams)
	}
	if args.Get(1) != nil {
		detailedResp = args.Get(1).(*core.DetailedResponse)
	}

	return resp, detailedResp, args.Error(2)
}

func (m *MockBRSClient) GetProtectionSourceRegistration(opts *backuprecoveryv1.GetProtectionSourceRegistrationOptions) (*backuprecoveryv1.SourceRegistrationResponseParams, *core.DetailedResponse, error) {
	args := m.Called(opts)
	var resp *backuprecoveryv1.SourceRegistrationResponseParams
	var detailedResp *core.DetailedResponse

	if args.Get(0) != nil {
		resp = args.Get(0).(*backuprecoveryv1.SourceRegistrationResponseParams)
	}
	if args.Get(1) != nil {
		detailedResp = args.Get(1).(*core.DetailedResponse)
	}

	return resp, detailedResp, args.Error(2)
}

func (m *MockBRSClient) GetProtectionSourceRegistrationWithContext(ctx context.Context, opts *backuprecoveryv1.GetProtectionSourceRegistrationOptions) (*backuprecoveryv1.SourceRegistrationResponseParams, *core.DetailedResponse, error) {
	args := m.Called(ctx, opts)
	var resp *backuprecoveryv1.SourceRegistrationResponseParams
	var detailedResp *core.DetailedResponse

	if args.Get(0) != nil {
		resp = args.Get(0).(*backuprecoveryv1.SourceRegistrationResponseParams)
	}
	if args.Get(1) != nil {
		detailedResp = args.Get(1).(*core.DetailedResponse)
	}

	return resp, detailedResp, args.Error(2)
}

func (m *MockBRSClient) UpdateProtectionSourceRegistration(opts *backuprecoveryv1.UpdateProtectionSourceRegistrationOptions) (*backuprecoveryv1.SourceRegistrationResponseParams, *core.DetailedResponse, error) {
	args := m.Called(opts)
	var resp *backuprecoveryv1.SourceRegistrationResponseParams
	var detailedResp *core.DetailedResponse

	if args.Get(0) != nil {
		resp = args.Get(0).(*backuprecoveryv1.SourceRegistrationResponseParams)
	}
	if args.Get(1) != nil {
		detailedResp = args.Get(1).(*core.DetailedResponse)
	}

	return resp, detailedResp, args.Error(2)
}

func (m *MockBRSClient) UpdateProtectionSourceRegistrationWithContext(ctx context.Context, opts *backuprecoveryv1.UpdateProtectionSourceRegistrationOptions) (*backuprecoveryv1.SourceRegistrationResponseParams, *core.DetailedResponse, error) {
	args := m.Called(ctx, opts)
	var resp *backuprecoveryv1.SourceRegistrationResponseParams
	var detailedResp *core.DetailedResponse

	if args.Get(0) != nil {
		resp = args.Get(0).(*backuprecoveryv1.SourceRegistrationResponseParams)
	}
	if args.Get(1) != nil {
		detailedResp = args.Get(1).(*core.DetailedResponse)
	}

	return resp, detailedResp, args.Error(2)
}

func (m *MockBRSClient) PatchProtectionSourceRegistration(opts *backuprecoveryv1.PatchProtectionSourceRegistrationOptions) (*backuprecoveryv1.SourceRegistrationResponseParams, *core.DetailedResponse, error) {
	args := m.Called(opts)
	var resp *backuprecoveryv1.SourceRegistrationResponseParams
	var detailedResp *core.DetailedResponse

	if args.Get(0) != nil {
		resp = args.Get(0).(*backuprecoveryv1.SourceRegistrationResponseParams)
	}
	if args.Get(1) != nil {
		detailedResp = args.Get(1).(*core.DetailedResponse)
	}

	return resp, detailedResp, args.Error(2)
}

func (m *MockBRSClient) PatchProtectionSourceRegistrationWithContext(ctx context.Context, opts *backuprecoveryv1.PatchProtectionSourceRegistrationOptions) (*backuprecoveryv1.SourceRegistrationResponseParams, *core.DetailedResponse, error) {
	args := m.Called(ctx, opts)
	var resp *backuprecoveryv1.SourceRegistrationResponseParams
	var detailedResp *core.DetailedResponse

	if args.Get(0) != nil {
		resp = args.Get(0).(*backuprecoveryv1.SourceRegistrationResponseParams)
	}
	if args.Get(1) != nil {
		detailedResp = args.Get(1).(*core.DetailedResponse)
	}

	return resp, detailedResp, args.Error(2)
}

func (m *MockBRSClient) DeleteProtectionSourceRegistration(opts *backuprecoveryv1.DeleteProtectionSourceRegistrationOptions) (*core.DetailedResponse, error) {
	args := m.Called(opts)
	var detailedResp *core.DetailedResponse

	if args.Get(0) != nil {
		detailedResp = args.Get(1).(*core.DetailedResponse)
	}

	return detailedResp, args.Error(1)
}

func (m *MockBRSClient) DeleteProtectionSourceRegistrationWithContext(ctx context.Context, opts *backuprecoveryv1.DeleteProtectionSourceRegistrationOptions) (*core.DetailedResponse, error) {
	args := m.Called(ctx, opts)
	if args.Get(0) == nil {
		return nil, args.Error(2)
	}
	return args.Get(1).(*core.DetailedResponse), args.Error(2)
}

func (m *MockBRSClient) RefreshProtectionSourceByID(opts *backuprecoveryv1.RefreshProtectionSourceByIdOptions) (*core.DetailedResponse, error) {
	args := m.Called(opts)
	return args.Get(0).(*core.DetailedResponse), args.Error(1)
}

func (m *MockBRSClient) RefreshProtectionSourceByIDWithContext(ctx context.Context, opts *backuprecoveryv1.RefreshProtectionSourceByIdOptions) (*core.DetailedResponse, error) {
	args := m.Called(ctx, opts)
	if args.Get(0) == nil {
		return nil, args.Error(2)
	}
	return args.Get(1).(*core.DetailedResponse), args.Error(2)
}

// Progress Monitor operations
func (m *MockBRSClient) GetProgressMonitors(opts *backuprecoveryv1.GetProgressMonitorsOptions) (*backuprecoveryv1.GetTasksResult, *core.DetailedResponse, error) {
	args := m.Called(opts)
	var resp *backuprecoveryv1.GetTasksResult
	var detailedResp *core.DetailedResponse
	if args.Get(0) != nil {
		resp = args.Get(0).(*backuprecoveryv1.GetTasksResult)
	}
	if args.Get(1) != nil {
		detailedResp = args.Get(1).(*core.DetailedResponse)
	}

	return resp, detailedResp, args.Error(2)
}

func (m *MockBRSClient) GetProgressMonitorsWithContext(ctx context.Context, opts *backuprecoveryv1.GetProgressMonitorsOptions) (*backuprecoveryv1.GetTasksResult, *core.DetailedResponse, error) {
	args := m.Called(ctx, opts)
	var resp *backuprecoveryv1.GetTasksResult
	var detailedResp *core.DetailedResponse
	if args.Get(0) != nil {
		resp = args.Get(0).(*backuprecoveryv1.GetTasksResult)
	}
	if args.Get(1) != nil {
		detailedResp = args.Get(1).(*core.DetailedResponse)
	}

	return resp, detailedResp, args.Error(2)
}

func (m *MockBRSClient) GetProtectionRunProgress(opts *backuprecoveryv1.GetProtectionRunProgressOptions) (*backuprecoveryv1.GetProtectionRunProgressBody, *core.DetailedResponse, error) {
	args := m.Called(opts)
	var resp *backuprecoveryv1.GetProtectionRunProgressBody
	var detailedResp *core.DetailedResponse

	if args.Get(0) != nil {
		resp = args.Get(0).(*backuprecoveryv1.GetProtectionRunProgressBody)
	}
	if args.Get(1) != nil {
		detailedResp = args.Get(1).(*core.DetailedResponse)
	}

	return resp, detailedResp, args.Error(2)
}

func (m *MockBRSClient) GetProtectionRunProgressWithContext(ctx context.Context, opts *backuprecoveryv1.GetProtectionRunProgressOptions) (*backuprecoveryv1.GetProtectionRunProgressBody, *core.DetailedResponse, error) {
	args := m.Called(ctx, opts)
	var resp *backuprecoveryv1.GetProtectionRunProgressBody
	var detailedResp *core.DetailedResponse

	if args.Get(0) != nil {
		resp = args.Get(0).(*backuprecoveryv1.GetProtectionRunProgressBody)
	}
	if args.Get(1) != nil {
		detailedResp = args.Get(1).(*core.DetailedResponse)
	}

	return resp, detailedResp, args.Error(2)
}

// Meta Info operations
func (m *MockBRSClient) ConstructMetaInfo(opts *backuprecoveryv1.ConstructMetaInfoOptions) (*backuprecoveryv1.ConstructMetaInfoResult, *core.DetailedResponse, error) {
	args := m.Called(opts)
	if args.Get(0) == nil {
		return nil, args.Get(1).(*core.DetailedResponse), args.Error(2)
	}
	return args.Get(0).(*backuprecoveryv1.ConstructMetaInfoResult), args.Get(1).(*core.DetailedResponse), args.Error(2)
}

func (m *MockBRSClient) ConstructMetaInfoWithContext(ctx context.Context, opts *backuprecoveryv1.ConstructMetaInfoOptions) (*backuprecoveryv1.ConstructMetaInfoResult, *core.DetailedResponse, error) {
	args := m.Called(ctx, opts)
	if args.Get(0) == nil {
		return nil, args.Get(1).(*core.DetailedResponse), args.Error(2)
	}
	return args.Get(0).(*backuprecoveryv1.ConstructMetaInfoResult), args.Get(1).(*core.DetailedResponse), args.Error(2)
}
