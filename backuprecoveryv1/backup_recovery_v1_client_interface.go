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

package backuprecoveryv1

import (
	"context"
	"io"
	"net/http"
	"time"

	"github.com/IBM/go-sdk-core/v5/core"
)

// BRSClientInterface defines the interface for BRS client operations.
// This interface allows for easier testing and mocking of the BRS client.
// It includes all methods implemented by the BackupRecoveryV1 struct.
type BRSClientInterface interface {
	// Service configuration methods
	Clone() *BackupRecoveryV1
	SetServiceURL(url string) error
	GetServiceURL() string
	SetDefaultHeaders(headers http.Header)
	SetEnableGzipCompression(enableGzip bool)
	GetEnableGzipCompression() bool
	EnableRetries(maxRetries int, maxRetryInterval time.Duration)
	DisableRetries()

	// Agent operations
	DownloadAgent(downloadAgentOptions *DownloadAgentOptions) (result io.ReadCloser, response *core.DetailedResponse, err error)
	DownloadAgentWithContext(ctx context.Context, downloadAgentOptions *DownloadAgentOptions) (result io.ReadCloser, response *core.DetailedResponse, err error)
	GetUpgradeTasks(getUpgradeTasksOptions *GetUpgradeTasksOptions) (result *AgentUpgradeTaskStates, response *core.DetailedResponse, err error)
	GetUpgradeTasksWithContext(ctx context.Context, getUpgradeTasksOptions *GetUpgradeTasksOptions) (result *AgentUpgradeTaskStates, response *core.DetailedResponse, err error)
	CreateUpgradeTask(createUpgradeTaskOptions *CreateUpgradeTaskOptions) (result *AgentUpgradeTaskState, response *core.DetailedResponse, err error)
	CreateUpgradeTaskWithContext(ctx context.Context, createUpgradeTaskOptions *CreateUpgradeTaskOptions) (result *AgentUpgradeTaskState, response *core.DetailedResponse, err error)

	// Protection Source operations
	ListProtectionSources(listProtectionSourcesOptions *ListProtectionSourcesOptions) (result []ProtectionSourceNodes, response *core.DetailedResponse, err error)
	ListProtectionSourcesWithContext(ctx context.Context, listProtectionSourcesOptions *ListProtectionSourcesOptions) (result []ProtectionSourceNodes, response *core.DetailedResponse, err error)
	ListProtectionSourcesRegistrationInfo(listProtectionSourcesRegistrationInfoOptions *ListProtectionSourcesRegistrationInfoOptions) (result *GetRegistrationInfoResponse, response *core.DetailedResponse, err error)
	ListProtectionSourcesRegistrationInfoWithContext(ctx context.Context, listProtectionSourcesRegistrationInfoOptions *ListProtectionSourcesRegistrationInfoOptions) (result *GetRegistrationInfoResponse, response *core.DetailedResponse, err error)

	// Connection operations
	GetDataSourceConnections(getDataSourceConnectionsOptions *GetDataSourceConnectionsOptions) (result *DataSourceConnectionList, response *core.DetailedResponse, err error)
	GetDataSourceConnectionsWithContext(ctx context.Context, getDataSourceConnectionsOptions *GetDataSourceConnectionsOptions) (result *DataSourceConnectionList, response *core.DetailedResponse, err error)
	CreateDataSourceConnection(createDataSourceConnectionOptions *CreateDataSourceConnectionOptions) (result *DataSourceConnection, response *core.DetailedResponse, err error)
	CreateDataSourceConnectionWithContext(ctx context.Context, createDataSourceConnectionOptions *CreateDataSourceConnectionOptions) (result *DataSourceConnection, response *core.DetailedResponse, err error)
	DeleteDataSourceConnection(deleteDataSourceConnectionOptions *DeleteDataSourceConnectionOptions) (response *core.DetailedResponse, err error)
	DeleteDataSourceConnectionWithContext(ctx context.Context, deleteDataSourceConnectionOptions *DeleteDataSourceConnectionOptions) (response *core.DetailedResponse, err error)
	PatchDataSourceConnection(patchDataSourceConnectionOptions *PatchDataSourceConnectionOptions) (result *DataSourceConnection, response *core.DetailedResponse, err error)
	PatchDataSourceConnectionWithContext(ctx context.Context, patchDataSourceConnectionOptions *PatchDataSourceConnectionOptions) (result *DataSourceConnection, response *core.DetailedResponse, err error)
	GenerateDataSourceConnectionRegistrationToken(generateDataSourceConnectionRegistrationTokenOptions *GenerateDataSourceConnectionRegistrationTokenOptions) (result *string, response *core.DetailedResponse, err error)
	GenerateDataSourceConnectionRegistrationTokenWithContext(ctx context.Context, generateDataSourceConnectionRegistrationTokenOptions *GenerateDataSourceConnectionRegistrationTokenOptions) (result *string, response *core.DetailedResponse, err error)

	// Connector operations
	GetDataSourceConnectors(getDataSourceConnectorsOptions *GetDataSourceConnectorsOptions) (result *DataSourceConnectorList, response *core.DetailedResponse, err error)
	GetDataSourceConnectorsWithContext(ctx context.Context, getDataSourceConnectorsOptions *GetDataSourceConnectorsOptions) (result *DataSourceConnectorList, response *core.DetailedResponse, err error)
	GetConnectorMetadata(getConnectorMetadataOptions *GetConnectorMetadataOptions) (result *ConnectorMetadata, response *core.DetailedResponse, err error)
	GetConnectorMetadataWithContext(ctx context.Context, getConnectorMetadataOptions *GetConnectorMetadataOptions) (result *ConnectorMetadata, response *core.DetailedResponse, err error)
	DeleteDataSourceConnector(deleteDataSourceConnectorOptions *DeleteDataSourceConnectorOptions) (response *core.DetailedResponse, err error)
	DeleteDataSourceConnectorWithContext(ctx context.Context, deleteDataSourceConnectorOptions *DeleteDataSourceConnectorOptions) (response *core.DetailedResponse, err error)
	PatchDataSourceConnector(patchDataSourceConnectorOptions *PatchDataSourceConnectorOptions) (result *DataSourceConnector, response *core.DetailedResponse, err error)
	PatchDataSourceConnectorWithContext(ctx context.Context, patchDataSourceConnectorOptions *PatchDataSourceConnectorOptions) (result *DataSourceConnector, response *core.DetailedResponse, err error)

	// Object Snapshot operations
	GetObjectSnapshots(getObjectSnapshotsOptions *GetObjectSnapshotsOptions) (result *GetObjectSnapshotsResponse, response *core.DetailedResponse, err error)
	GetObjectSnapshotsWithContext(ctx context.Context, getObjectSnapshotsOptions *GetObjectSnapshotsOptions) (result *GetObjectSnapshotsResponse, response *core.DetailedResponse, err error)

	// Protection Policy operations
	GetProtectionPolicies(getProtectionPoliciesOptions *GetProtectionPoliciesOptions) (result *ProtectionPoliciesResponse, response *core.DetailedResponse, err error)
	GetProtectionPoliciesWithContext(ctx context.Context, getProtectionPoliciesOptions *GetProtectionPoliciesOptions) (result *ProtectionPoliciesResponse, response *core.DetailedResponse, err error)
	CreateProtectionPolicy(createProtectionPolicyOptions *CreateProtectionPolicyOptions) (result *ProtectionPolicyResponse, response *core.DetailedResponse, err error)
	CreateProtectionPolicyWithContext(ctx context.Context, createProtectionPolicyOptions *CreateProtectionPolicyOptions) (result *ProtectionPolicyResponse, response *core.DetailedResponse, err error)
	GetProtectionPolicyByID(getProtectionPolicyByIdOptions *GetProtectionPolicyByIdOptions) (result *ProtectionPolicyResponse, response *core.DetailedResponse, err error)
	GetProtectionPolicyByIDWithContext(ctx context.Context, getProtectionPolicyByIdOptions *GetProtectionPolicyByIdOptions) (result *ProtectionPolicyResponse, response *core.DetailedResponse, err error)
	UpdateProtectionPolicy(updateProtectionPolicyOptions *UpdateProtectionPolicyOptions) (result *ProtectionPolicyResponse, response *core.DetailedResponse, err error)
	UpdateProtectionPolicyWithContext(ctx context.Context, updateProtectionPolicyOptions *UpdateProtectionPolicyOptions) (result *ProtectionPolicyResponse, response *core.DetailedResponse, err error)
	DeleteProtectionPolicy(deleteProtectionPolicyOptions *DeleteProtectionPolicyOptions) (response *core.DetailedResponse, err error)
	DeleteProtectionPolicyWithContext(ctx context.Context, deleteProtectionPolicyOptions *DeleteProtectionPolicyOptions) (response *core.DetailedResponse, err error)

	// Protection Group operations
	GetProtectionGroups(getProtectionGroupsOptions *GetProtectionGroupsOptions) (result *ProtectionGroupsResponse, response *core.DetailedResponse, err error)
	GetProtectionGroupsWithContext(ctx context.Context, getProtectionGroupsOptions *GetProtectionGroupsOptions) (result *ProtectionGroupsResponse, response *core.DetailedResponse, err error)
	CreateProtectionGroup(createProtectionGroupOptions *CreateProtectionGroupOptions) (result *ProtectionGroupResponse, response *core.DetailedResponse, err error)
	CreateProtectionGroupWithContext(ctx context.Context, createProtectionGroupOptions *CreateProtectionGroupOptions) (result *ProtectionGroupResponse, response *core.DetailedResponse, err error)
	GetProtectionGroupByID(getProtectionGroupByIdOptions *GetProtectionGroupByIdOptions) (result *ProtectionGroupResponse, response *core.DetailedResponse, err error)
	GetProtectionGroupByIDWithContext(ctx context.Context, getProtectionGroupByIdOptions *GetProtectionGroupByIdOptions) (result *ProtectionGroupResponse, response *core.DetailedResponse, err error)
	UpdateProtectionGroup(updateProtectionGroupOptions *UpdateProtectionGroupOptions) (result *ProtectionGroupResponse, response *core.DetailedResponse, err error)
	UpdateProtectionGroupWithContext(ctx context.Context, updateProtectionGroupOptions *UpdateProtectionGroupOptions) (result *ProtectionGroupResponse, response *core.DetailedResponse, err error)
	DeleteProtectionGroup(deleteProtectionGroupOptions *DeleteProtectionGroupOptions) (response *core.DetailedResponse, err error)
	DeleteProtectionGroupWithContext(ctx context.Context, deleteProtectionGroupOptions *DeleteProtectionGroupOptions) (response *core.DetailedResponse, err error)

	// Protection Group Run operations
	GetProtectionGroupRuns(getProtectionGroupRunsOptions *GetProtectionGroupRunsOptions) (result *ProtectionGroupRunsResponse, response *core.DetailedResponse, err error)
	GetProtectionGroupRunsWithContext(ctx context.Context, getProtectionGroupRunsOptions *GetProtectionGroupRunsOptions) (result *ProtectionGroupRunsResponse, response *core.DetailedResponse, err error)
	UpdateProtectionGroupRun(updateProtectionGroupRunOptions *UpdateProtectionGroupRunOptions) (result *UpdateProtectionGroupRunResponse, response *core.DetailedResponse, err error)
	UpdateProtectionGroupRunWithContext(ctx context.Context, updateProtectionGroupRunOptions *UpdateProtectionGroupRunOptions) (result *UpdateProtectionGroupRunResponse, response *core.DetailedResponse, err error)
	CreateProtectionGroupRun(createProtectionGroupRunOptions *CreateProtectionGroupRunOptions) (result *CreateProtectionGroupRunResponse, response *core.DetailedResponse, err error)
	CreateProtectionGroupRunWithContext(ctx context.Context, createProtectionGroupRunOptions *CreateProtectionGroupRunOptions) (result *CreateProtectionGroupRunResponse, response *core.DetailedResponse, err error)
	PerformActionOnProtectionGroupRun(performActionOnProtectionGroupRunOptions *PerformActionOnProtectionGroupRunOptions) (result *PerformRunActionResponse, response *core.DetailedResponse, err error)
	PerformActionOnProtectionGroupRunWithContext(ctx context.Context, performActionOnProtectionGroupRunOptions *PerformActionOnProtectionGroupRunOptions) (result *PerformRunActionResponse, response *core.DetailedResponse, err error)
	GetProtectionGroupRun(getProtectionGroupRunOptions *GetProtectionGroupRunOptions) (result *ProtectionGroupRun, response *core.DetailedResponse, err error)
	GetProtectionGroupRunWithContext(ctx context.Context, getProtectionGroupRunOptions *GetProtectionGroupRunOptions) (result *ProtectionGroupRun, response *core.DetailedResponse, err error)

	// Recovery operations
	GetRecoveries(getRecoveriesOptions *GetRecoveriesOptions) (result *RecoveriesResponse, response *core.DetailedResponse, err error)
	GetRecoveriesWithContext(ctx context.Context, getRecoveriesOptions *GetRecoveriesOptions) (result *RecoveriesResponse, response *core.DetailedResponse, err error)
	CreateRecovery(createRecoveryOptions *CreateRecoveryOptions) (result *Recovery, response *core.DetailedResponse, err error)
	CreateRecoveryWithContext(ctx context.Context, createRecoveryOptions *CreateRecoveryOptions) (result *Recovery, response *core.DetailedResponse, err error)
	CreateDownloadFilesAndFoldersRecovery(createDownloadFilesAndFoldersRecoveryOptions *CreateDownloadFilesAndFoldersRecoveryOptions) (result *Recovery, response *core.DetailedResponse, err error)
	CreateDownloadFilesAndFoldersRecoveryWithContext(ctx context.Context, createDownloadFilesAndFoldersRecoveryOptions *CreateDownloadFilesAndFoldersRecoveryOptions) (result *Recovery, response *core.DetailedResponse, err error)
	GetRecoveryByID(getRecoveryByIdOptions *GetRecoveryByIdOptions) (result *Recovery, response *core.DetailedResponse, err error)
	GetRecoveryByIDWithContext(ctx context.Context, getRecoveryByIdOptions *GetRecoveryByIdOptions) (result *Recovery, response *core.DetailedResponse, err error)
	DownloadFilesFromRecovery(downloadFilesFromRecoveryOptions *DownloadFilesFromRecoveryOptions) (response *core.DetailedResponse, err error)
	DownloadFilesFromRecoveryWithContext(ctx context.Context, downloadFilesFromRecoveryOptions *DownloadFilesFromRecoveryOptions) (response *core.DetailedResponse, err error)
	CancelRecoveryByID(cancelRecoveryByIdOptions *CancelRecoveryByIdOptions) (response *core.DetailedResponse, err error)
	CancelRecoveryByIDWithContext(ctx context.Context, cancelRecoveryByIdOptions *CancelRecoveryByIdOptions) (response *core.DetailedResponse, err error)

	// Restore Point operations
	GetRestorePointsInTimeRange(getRestorePointsInTimeRangeOptions *GetRestorePointsInTimeRangeOptions) (result *GetRestorePointsInTimeRangeResponse, response *core.DetailedResponse, err error)
	GetRestorePointsInTimeRangeWithContext(ctx context.Context, getRestorePointsInTimeRangeOptions *GetRestorePointsInTimeRangeOptions) (result *GetRestorePointsInTimeRangeResponse, response *core.DetailedResponse, err error)

	// Indexed File operations
	DownloadIndexedFile(downloadIndexedFileOptions *DownloadIndexedFileOptions) (response *core.DetailedResponse, err error)
	DownloadIndexedFileWithContext(ctx context.Context, downloadIndexedFileOptions *DownloadIndexedFileOptions) (response *core.DetailedResponse, err error)
	SearchIndexedObjects(searchIndexedObjectsOptions *SearchIndexedObjectsOptions) (result *SearchIndexedObjectsResponse, response *core.DetailedResponse, err error)
	SearchIndexedObjectsWithContext(ctx context.Context, searchIndexedObjectsOptions *SearchIndexedObjectsOptions) (result *SearchIndexedObjectsResponse, response *core.DetailedResponse, err error)

	// Object Search operations
	SearchObjects(searchObjectsOptions *SearchObjectsOptions) (result *ObjectsSearchResponseBody, response *core.DetailedResponse, err error)
	SearchObjectsWithContext(ctx context.Context, searchObjectsOptions *SearchObjectsOptions) (result *ObjectsSearchResponseBody, response *core.DetailedResponse, err error)
	SearchProtectedObjects(searchProtectedObjectsOptions *SearchProtectedObjectsOptions) (result *ProtectedObjectsSearchResponse, response *core.DetailedResponse, err error)
	SearchProtectedObjectsWithContext(ctx context.Context, searchProtectedObjectsOptions *SearchProtectedObjectsOptions) (result *ProtectedObjectsSearchResponse, response *core.DetailedResponse, err error)

	// Source Registration operations
	GetSourceRegistrations(getSourceRegistrationsOptions *GetSourceRegistrationsOptions) (result *SourceRegistrations, response *core.DetailedResponse, err error)
	GetSourceRegistrationsWithContext(ctx context.Context, getSourceRegistrationsOptions *GetSourceRegistrationsOptions) (result *SourceRegistrations, response *core.DetailedResponse, err error)
	RegisterProtectionSource(registerProtectionSourceOptions *RegisterProtectionSourceOptions) (result *SourceRegistrationResponseParams, response *core.DetailedResponse, err error)
	RegisterProtectionSourceWithContext(ctx context.Context, registerProtectionSourceOptions *RegisterProtectionSourceOptions) (result *SourceRegistrationResponseParams, response *core.DetailedResponse, err error)
	GetProtectionSourceRegistration(getProtectionSourceRegistrationOptions *GetProtectionSourceRegistrationOptions) (result *SourceRegistrationResponseParams, response *core.DetailedResponse, err error)
	GetProtectionSourceRegistrationWithContext(ctx context.Context, getProtectionSourceRegistrationOptions *GetProtectionSourceRegistrationOptions) (result *SourceRegistrationResponseParams, response *core.DetailedResponse, err error)
	UpdateProtectionSourceRegistration(updateProtectionSourceRegistrationOptions *UpdateProtectionSourceRegistrationOptions) (result *SourceRegistrationResponseParams, response *core.DetailedResponse, err error)
	UpdateProtectionSourceRegistrationWithContext(ctx context.Context, updateProtectionSourceRegistrationOptions *UpdateProtectionSourceRegistrationOptions) (result *SourceRegistrationResponseParams, response *core.DetailedResponse, err error)
	PatchProtectionSourceRegistration(patchProtectionSourceRegistrationOptions *PatchProtectionSourceRegistrationOptions) (result *SourceRegistrationResponseParams, response *core.DetailedResponse, err error)
	PatchProtectionSourceRegistrationWithContext(ctx context.Context, patchProtectionSourceRegistrationOptions *PatchProtectionSourceRegistrationOptions) (result *SourceRegistrationResponseParams, response *core.DetailedResponse, err error)
	DeleteProtectionSourceRegistration(deleteProtectionSourceRegistrationOptions *DeleteProtectionSourceRegistrationOptions) (response *core.DetailedResponse, err error)
	DeleteProtectionSourceRegistrationWithContext(ctx context.Context, deleteProtectionSourceRegistrationOptions *DeleteProtectionSourceRegistrationOptions) (response *core.DetailedResponse, err error)
	RefreshProtectionSourceByID(refreshProtectionSourceByIdOptions *RefreshProtectionSourceByIdOptions) (response *core.DetailedResponse, err error)
	RefreshProtectionSourceByIDWithContext(ctx context.Context, refreshProtectionSourceByIdOptions *RefreshProtectionSourceByIdOptions) (response *core.DetailedResponse, err error)

	// Progress Monitor operations
	GetProgressMonitors(getProgressMonitorsOptions *GetProgressMonitorsOptions) (result *GetTasksResult, response *core.DetailedResponse, err error)
	GetProgressMonitorsWithContext(ctx context.Context, getProgressMonitorsOptions *GetProgressMonitorsOptions) (result *GetTasksResult, response *core.DetailedResponse, err error)
	GetProtectionRunProgress(getProtectionRunProgressOptions *GetProtectionRunProgressOptions) (result *GetProtectionRunProgressBody, response *core.DetailedResponse, err error)
	GetProtectionRunProgressWithContext(ctx context.Context, getProtectionRunProgressOptions *GetProtectionRunProgressOptions) (result *GetProtectionRunProgressBody, response *core.DetailedResponse, err error)

	// Meta Info operations
	ConstructMetaInfo(constructMetaInfoOptions *ConstructMetaInfoOptions) (result *ConstructMetaInfoResult, response *core.DetailedResponse, err error)
	ConstructMetaInfoWithContext(ctx context.Context, constructMetaInfoOptions *ConstructMetaInfoOptions) (result *ConstructMetaInfoResult, response *core.DetailedResponse, err error)
}
