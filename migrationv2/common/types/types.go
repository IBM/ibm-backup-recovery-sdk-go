/*******************************************************************************
 * IBM Confidential
 * OCO Source Materials
 * (C) Copyright IBM Corp. 2026 All Rights Reserved.
 * The source code for this program is not  published or otherwise divested of
 * its trade secrets, irrespective of what has been deposited with
 * the U.S. Copyright Office.
 ******************************************************************************/
package types

import (
	"time"

	backuprecoveryv1 "github.com/IBM/ibm-backup-recovery-sdk-go/backuprecoveryv1"
	activity_tracker "github.com/IBM/ibm-backup-recovery-sdk-go/migrationv2/common/activity-tracker"
	"github.com/IBM/ibm-backup-recovery-sdk-go/migrationv2/common/logger"
	"github.com/IBM/ibm-backup-recovery-sdk-go/migrationv2/common/metrics"
	resourcecontrollerv2 "github.com/IBM/platform-services-go-sdk/resourcecontrollerv2"
)

// BRSClientWrapperInterface defines the interface for BRS client wrapper operations.
// This interface allows for easier testing and mocking of the BRS client wrapper.
type BRSClientWrapperInterface interface {
	GetTenantId() string
	GetBRSClient() backuprecoveryv1.BRSClientInterface
	GetRegion() string
	GetCRN() string
}

// DATA SOURCE TYPES

// DataSourceType represents the type of data source
type DataSourceType string

const (
	DataSourceKubernetes DataSourceType = "kubernetes"
	DataSourceVPCVSI     DataSourceType = "vpcvsi"
	DataSourceExtensible DataSourceType = "extensible"
)

type ConnectionEnvType string

// User-friendly connection type constants
const (
	ConnectionType_IKS_CLASSIC  ConnectionEnvType = "IKS_CLASSIC"
	ConnectionType_IKS_VPC      ConnectionEnvType = "IKS_VPC"
	ConnectionType_ROKS_CLASSIC ConnectionEnvType = "ROKS_CLASSIC"
	ConnectionType_ROKS_VPC     ConnectionEnvType = "ROKS_VPC"
	ConnectionType_VSI          ConnectionEnvType = "VSI"
)

// connectionTypeMap maps user-friendly types to internal BRS types
var connectionTypeMap = map[ConnectionEnvType]string{
	ConnectionType_IKS_CLASSIC:  string(backuprecoveryv1.DataSourceConnection_ConnectionEnvType_Kiksclassic),
	ConnectionType_IKS_VPC:      string(backuprecoveryv1.DataSourceConnection_ConnectionEnvType_Kiksvpc),
	ConnectionType_ROKS_CLASSIC: string(backuprecoveryv1.DataSourceConnection_ConnectionEnvType_Kroksclassic),
	ConnectionType_ROKS_VPC:     string(backuprecoveryv1.DataSourceConnection_ConnectionEnvType_Kroksvpc),
	ConnectionType_VSI:          "VSI",
}

// ToBRSConnectionType converts user-friendly connection type to internal BRS type
func (c ConnectionEnvType) ToBRSConnectionType() string {
	if brsType, ok := connectionTypeMap[c]; ok {
		return brsType
	}
	// Return as-is if not found in map
	return string(c)
}

// ConnectionResult represents the result of a connection operation
type ConnectionResult struct {
	ConnectionID      string    `json:"connectionId"`      // BRS-assigned connection ID
	ConnectionName    string    `json:"connectionName"`    // Connection name
	RegistrationToken string    `json:"registrationToken"` // Token for connector deployment (renamed from RegistrationToken)
	Type              string    `json:"type"`              // Connection type
	Status            string    `json:"status"`            // Connection status
	Message           string    `json:"message"`           // Status message
	CreatedAt         time.Time `json:"createdAt"`         // Creation timestamp
}

// RegistrationResult represents the result of a registration operation
type RegistrationResult struct {
	RegistrationID int64     `json:"registrationId"`
	ConnectionID   string    `json:"connectionId"`
	SourceName     string    `json:"sourceName"`
	Status         string    `json:"status"`
	Message        string    `json:"message"`
	CreatedAt      time.Time `json:"createdAt"`
}

// ProtectionGroupResult represents the result of a protection group operation
type ProtectionGroupResult struct {
	ProtectionGroupID string    `json:"protectionGroupId"`
	GroupName         string    `json:"groupName"`
	RegistrationID    int64     `json:"registrationId"`
	PolicyID          string    `json:"policyId"`
	Status            string    `json:"status"`
	Message           string    `json:"message"`
	CreatedAt         time.Time `json:"createdAt"`
}

// BackupResult represents the result of a backup operation
type BackupResult struct {
	BackupID          string    `json:"backupId"`
	ProtectionGroupID string    `json:"protectionGroupId"`
	Status            string    `json:"status"`   // "running", "completed", "failed"
	Progress          *float32  `json:"progress"` // 0-100
	Message           string    `json:"message"`
	Timestamp         time.Time `json:"timestamp"`
	StartedAt         time.Time `json:"startedAt"`
	CompletedAt       time.Time `json:"completedAt,omitempty"`
}

// RestoreResult represents the result of a restore operation
type RestoreResult struct {
	RestoreID   string     `json:"restoreId"`
	BackupID    string     `json:"backupId"`
	TargetID    int64      `json:"targetId"`
	Status      string     `json:"status"`   // "running", "completed", "failed"
	Progress    int        `json:"progress"` // 0-100
	Message     string     `json:"message"`
	Timestamp   time.Time  `json:"timestamp"`
	StartedAt   time.Time  `json:"startedAt"`
	CompletedAt *time.Time `json:"completedAt,omitempty"`
}

// WORKFLOW RESULT TYPES

// SetupResult represents the result of a setup workflow
type SetupResult struct {
	ConnectionID     string    `json:"connectionId"`
	ConnectorID      string    `json:"connectorId"`
	RegistrationID   int64     `json:"registrationId"`
	Status           string    `json:"status"`
	ResourcesCreated int       `json:"resourcesCreated"`
	ResourcesReused  int       `json:"resourcesReused"`
	Message          string    `json:"message"`
	StartedAt        time.Time `json:"startedAt"`
	CompletedAt      time.Time `json:"completedAt"`
	Duration         float64   `json:"durationSeconds"`
}

// ProtectResult represents the result of a protection workflow
type ProtectResult struct {
	ProtectionGroupID string    `json:"protectionGroupId"`
	GroupName         string    `json:"groupName"`
	RegistrationID    int64     `json:"registrationId"`
	PolicyID          string    `json:"policyId"`
	Status            string    `json:"status"`
	ResourcesCreated  int       `json:"resourcesCreated"`
	ResourcesReused   int       `json:"resourcesReused"`
	Message           string    `json:"message"`
	StartedAt         time.Time `json:"startedAt"`
	CompletedAt       time.Time `json:"completedAt"`
	Duration          float64   `json:"durationSeconds"`
}

// MigrationResult represents the result of a complete migration workflow (V8 Enhanced)
type MigrationResult struct {
	// Resource IDs
	SourceConnectionID   string `json:"sourceConnectionId"`   // Source connection ID
	SourceConnectorID    string `json:"sourceConnectorId"`    // Source connector ID
	SourceRegistrationID int64  `json:"sourceRegistrationId"` // Source registration ID
	TargetConnectionID   string `json:"targetConnectionId"`   // Target connection ID
	TargetConnectorID    string `json:"targetConnectorId"`    // Target connector ID
	TargetRegistrationID int64  `json:"targetRegistrationId"` // Target registration ID
	ProtectionGroupID    string `json:"protectionGroupId"`    // Protection group ID
	PolicyID             string `json:"policyId"`             // Policy ID
	BackupID             string `json:"backupId"`             // Backup ID
	RestoreID            string `json:"restoreId"`            // Restore ID

	// Status Information
	Status            string `json:"status"`            // Overall status: "in_progress", "completed", "failed"
	SourceSetupStatus string `json:"sourceSetupStatus"` // Source setup status
	TargetSetupStatus string `json:"targetSetupStatus"` // Target setup status
	ProtectionStatus  string `json:"protectionStatus"`  // Protection status
	BackupStatus      string `json:"backupStatus"`      // Backup status
	RestoreStatus     string `json:"restoreStatus"`     // Restore status

	// Resource Tracking
	ResourcesCreated int  `json:"resourcesCreated"` // Number of resources created
	ResourcesReused  int  `json:"resourcesReused"`  // Number of resources reused
	PolicyCreated    bool `json:"policyCreated"`    // True if policy was auto-created

	// Timing Information
	Message     string       `json:"message"`     // Status message
	StartedAt   time.Time    `json:"startedAt"`   // Migration start time
	CompletedAt time.Time    `json:"completedAt"` // Migration completion time
	Duration    float64      `json:"duration"`    // Duration in seconds
	StepResults []StepResult `json:"stepResults"` // Detailed step-by-step results
}

// CONFIGURATION TYPES

// SDKConfig holds configuration for the migration SDK
type SDKConfig struct {
	// BRS Configuration
	BRSURL          string
	BRSAPIKey       string
	BRSAuthToken    string
	Region          string
	APIKey          string
	BRSInstanceName string
	ResourceGroupID string
	Timeout         time.Duration

	// Polling Configuration
	DefaultPollingInterval time.Duration
	DefaultPollingTimeout  time.Duration

	// Feature Flags
	EnableWorkflowAPI bool
	EnableTaskAPI     bool

	// Logging
	EnableDebugLogging bool
	Logger             logger.Logger // Structured logger instance

	// Metrics collection instance. Non-nil when EnableMetrics=true.
	Metrics metrics.Metrics

	// Activity tracking sink. Non-nil when EnableActivityTracker=true.
	ActivityTrackerSink activity_tracker.Sink

	// AccountID for metrics and activity tracker labeling.
	// Set when EnableMetrics=true or EnableActivityTracker=true.
	AccountID string
}

// PollingConfig configures async operation polling
type PollingConfig struct {
	Interval     time.Duration
	Timeout      time.Duration
	InitialDelay time.Duration
	MaxAttempts  int
}

// DefaultPollingConfig returns default polling configuration
func DefaultPollingConfig() *PollingConfig {
	return &PollingConfig{
		Interval:     10 * time.Second,
		Timeout:      30 * time.Minute,
		InitialDelay: 5 * time.Second,
		MaxAttempts:  180, // 30 minutes with 10 second intervals
	}
}

// WORKFLOW TYPES

// WorkflowType represents the type of workflow
type WorkflowType string

const (
	WorkflowSetup      WorkflowType = "setup"
	WorkflowProtection WorkflowType = "protection"
	WorkflowProtect    WorkflowType = "protect"
	WorkflowBackup     WorkflowType = "backup"
	WorkflowRestore    WorkflowType = "restore"
	WorkflowMigration  WorkflowType = "migration"
	WorkflowMigrate    WorkflowType = "migrate"
)

// ASYNC OPERATION TYPES

// OperationStatus represents the status of an async operation
type OperationStatus string

const (
	OperationStatusPending   OperationStatus = "pending"
	OperationStatusRunning   OperationStatus = "running"
	OperationStatusCompleted OperationStatus = "completed"
	OperationStatusFailed    OperationStatus = "failed"
)

// AsyncOperation represents an async operation status
type AsyncOperation struct {
	OperationID   string
	OperationType string // "backup", "restore", "connector_deploy", etc.
	Status        string // "pending", "running", "completed", "failed"
	Progress      int    // 0-100
	Message       string
	StartedAt     time.Time
	CompletedAt   *time.Time
	Error         string
}

// BRS CLIENT WRAPPER

// BRSClientWrapper wraps the BRS client for easier testing and mocking
type BRSClientWrapper struct {
	Client   backuprecoveryv1.BRSClientInterface
	Region   string
	CRN      string
	TenantId string
}

// NewBRSClientWrapper creates a new BRS client wrapper
func NewBRSClientWrapper(client backuprecoveryv1.BRSClientInterface) *BRSClientWrapper {
	return &BRSClientWrapper{
		Client: client,
	}
}

// GetTenantId returns the tenant ID
func (w *BRSClientWrapper) GetTenantId() string {
	return w.TenantId
}

// GetBRSClient returns the BRS client interface
func (w *BRSClientWrapper) GetBRSClient() backuprecoveryv1.BRSClientInterface {
	return w.Client
}

// GetRegion returns the region
func (w *BRSClientWrapper) GetRegion() string {
	return w.Region
}

// GetCRN returns the CRN
func (w *BRSClientWrapper) GetCRN() string {
	return w.CRN
}

// BRS CLIENT WRAPPER

// BRSClientWrapper wraps the BRS client for easier testing and mocking
type ResourceControllerClientWrapper struct {
	Client *resourcecontrollerv2.ResourceControllerV2
	Region string
	CRN    string
}

// NewBRSClientWrapper creates a new BRS client wrapper
func NewResourceControllerClientWrapper(client *resourcecontrollerv2.ResourceControllerV2) *ResourceControllerClientWrapper {
	return &ResourceControllerClientWrapper{
		Client: client,
	}
}

// CONNECTION PARAMETERS

// ConnectionParams holds INPUT parameters for creating a connection
// Note: NO Credentials field - credentials are managed by DataSource
// Note: NO Status or ConnectionToken - these are OUTPUT fields in ConnectionResult
type ConnectionParams struct {
	Name     string                 `json:"name"`     // Connection name (INPUT - required)
	Type     ConnectionEnvType      `json:"type"`     // Connection type: "kubernetes", "vpcvsi", etc. (INPUT - required)
	Metadata map[string]interface{} `json:"metadata"` // Additional metadata (INPUT - optional)
}

// RegistrationParams holds parameters for registering a source
type RegistrationParams struct {
	SourceName  string                 `json:"sourceName"`
	SourceType  string                 `json:"sourceType"`
	Environment string                 `json:"environment"`
	Metadata    map[string]interface{} `json:"metadata"`
}

type BRSAgentImage struct {
	DataMover             string `ini:"dataMover"`
	Velero                string `ini:"velero"`
	VeleroAWSPlugin       string `ini:"veleroAwsPlugin"`
	DataPlugin            string `ini:"dataPlugin"`
	VeleroOpenShiftPlugin string `ini:"veleroOpenshiftPlugin"`
}

type BRSInstance struct {
	ID         string `json:"id"`
	Name       string `json:"name"`
	TenantID   string `json:"tenantID"`
	PrivateURL string `json:"privateURL"`
	PublicURL  string `json:"publicURL"`
	CRN        string `json:"crn"`
	Region     string `json:"region"`
}

// PROTECTION GROUP DETAILED TYPES

// ProtectionGroup represents a complete protection group configuration
type ProtectionGroup struct {
	Name              string              `json:"name"`
	NumberOfBackups   int64               `json:"numberOfBackups"`
	Frequency         Frequency           `json:"frequency,omitempty"`
	IncludeNamespaces string              `json:"includeNamespaces,omitempty"`
	ExcludeNamespaces string              `json:"excludeNamespaces,omitempty"`
	NamespacesSetting []NamespacesSetting `json:"namespacesSetting,omitempty"`
	Settings          *ProtectionSetting  `json:"settings,omitempty"`
	Policy            Policy              `json:"policy,omitempty"`
}

// Frequency defines backup frequency
type Frequency struct {
	Value int64  `json:"value,omitempty"`
	Unit  string `json:"unit,omitempty"` // "hours", "days", "weeks"
}

// NamespacesSetting defines per-namespace protection settings
type NamespacesSetting struct {
	Namespace              string          `json:"namespace,omitempty"`
	PersistentVolumeClaims PVCSection      `json:"persistentVolumeClaims,omitempty"`
	Resources              ResourceSection `json:"resources,omitempty"`
	Hooks                  HookSection     `json:"hooks,omitempty"`
}

// PVCSection defines PVC inclusion/exclusion
type PVCSection struct {
	Inclusion []string `json:"inclusion,omitempty"`
	Exclusion []string `json:"exclusion,omitempty"`
}

// ResourceSection defines resource type inclusion/exclusion
type ResourceSection struct {
	Inclusion *ResourceTypes `json:"inclusion,omitempty"`
	Exclusion *ResourceTypes `json:"exclusion,omitempty"`
}

// ResourceTypes lists resource types
type ResourceTypes struct {
	ResourceTypes []string `json:"resourceTypes,omitempty"`
}

// HookSection defines backup hooks
type HookSection struct {
	FailBackupIfHookFailed bool   `json:"failBackupIfHookFailed,omitempty"`
	RulesApplyMode         string `json:"rulesApplyMode,omitempty"`
	Rules                  []Rule `json:"rules,omitempty"`
}

type Rule struct {
	Rule RuleDetail `yaml:"rule,omitempty"`
}

// Rule defines a backup hook rule
type RuleDetail struct {
	PodLabels  string `yaml:"podLabels,omitempty"`
	PreScript  string `yaml:"preScript,omitempty"`
	PostScript string `yaml:"postScript,omitempty"`
	Container  string `yaml:"container,omitempty"`
}

type NamespaceData struct {
	Id     int64
	Name   string
	PvcIDs map[string]int64
}

// ProtectionSetting defines protection settings
type ProtectionSetting struct {
	// Add protection settings fields as needed
	RetentionDays int64 `json:"retentionDays,omitempty"`
	// StartTime    string     `yaml:"startTime,omitempty"`
	EndDate          string `yaml:"endDate,omitempty"`
	CSISnapshot      bool   `yaml:"CSISnapshot,omitempty"`
	Labels           Labels `yaml:"labels,omitempty"`
	PauseFutureRuns  bool   `yaml:"pauseFutureRuns,omitempty" json:"pauseFutureRuns,omitempty"`
}

type Labels struct {
	PersistentVolumeClaim KeyValueFilter `yaml:"persistentVolumeClaim"`
}

// Policy defines backup policy
type Policy struct {
	Name        string `json:"name,omitempty"`
	ID          string `json:"id,omitempty"`
	Description string `json:"description,omitempty"`
}

// ProtectionGroupParams holds INPUT parameters for creating a protection group
// This is a simplified version for API input, full ProtectionGroup is used for results
type ProtectionGroupParams struct {
	Name            string             `json:"name"`            // Protection group name
	NumberOfBackups int64              `json:"numberOfBackups"` // Number of backups to retain
	Frequency       *Frequency         `json:"frequency"`       // Backup frequency
	Settings        *ProtectionSetting `json:"settings"`        // Protection settings
	Policy          *Policy            `json:"policy"`          // Policy configuration

	// DataSource-specific params (only one should be set based on datasource type)
	// KubernetesParams *KubernetesProtectionParams `json:"kubernetesParams,omitempty"` // Kubernetes-specific params
	// VPCVSIParams     *VPCVSIProtectionParams     `json:"vpcvsiParams,omitempty"`     // VPC VSI-specific params
}

// KubernetesProtectionParams holds Kubernetes-specific protection parameters
type KubernetesProtectionParams struct {
	// Namespace selection
	IncludeNamespaces string              `json:"includeNamespaces,omitempty"` // Comma-separated or "*"
	ExcludeNamespaces string              `json:"excludeNamespaces,omitempty"` // Comma-separated
	NamespacesSetting []NamespacesSetting `json:"namespacesSetting,omitempty"` // Per-namespace config

	// Label-based selection
	LabelSelector *LabelSelector `json:"labelSelector,omitempty"` // Label-based filtering

	// Resource filtering
	IncludeResourceTypes []string `json:"includeResourceTypes,omitempty"` // e.g., ["Deployment", "StatefulSet"]
	ExcludeResourceTypes []string `json:"excludeResourceTypes,omitempty"` // e.g., ["Job", "Pod"]

	// PVC settings
	PVCBackupMode string   `json:"pvcBackupMode,omitempty"` // "snapshot", "filesystem", "both"
	IncludePVCs   []string `json:"includePVCs,omitempty"`   // PVC names to include
	ExcludePVCs   []string `json:"excludePVCs,omitempty"`   // PVC names to exclude

	// Backup hooks
	PreBackupHooks  []BackupHook `json:"preBackupHooks,omitempty"`  // Pre-backup hooks
	PostBackupHooks []BackupHook `json:"postBackupHooks,omitempty"` // Post-backup hooks
	FailOnHookError bool         `json:"failOnHookError"`           // Fail backup if hook fails

	Settings *ProtectionSetting `yaml:"settings,omitempty"`

	// Application awareness
	ApplicationConsistent bool   `json:"applicationConsistent"` // Use app-aware backup
	QuiesceTimeout        int64  `json:"quiesceTimeout"`        // Quiesce timeout in seconds
	IncludeAppLabels      string `yaml:"includeAppLabels"`
}

// LabelSelector defines label-based selection criteria
type LabelSelector struct {
	MatchLabels      map[string]string `json:"matchLabels,omitempty"`      // Labels that must match
	MatchExpressions []LabelExpression `json:"matchExpressions,omitempty"` // Label expressions
}

// LabelExpression defines a label selector expression
type LabelExpression struct {
	Key      string   `json:"key"`              // Label key
	Operator string   `json:"operator"`         // "In", "NotIn", "Exists", "DoesNotExist"
	Values   []string `json:"values,omitempty"` // Values for In/NotIn operators
}

// BackupHook defines a backup hook (pre/post backup command)
type BackupHook struct {
	Name      string   `json:"name"`              // Hook name
	Container string   `json:"container"`         // Container name to run hook in
	Command   []string `json:"command"`           // Command to execute
	OnError   string   `json:"onError"`           // "fail" or "continue"
	Timeout   int64    `json:"timeout,omitempty"` // Timeout in seconds
}

// VPCVSIProtectionParams holds VPC VSI-specific protection parameters
// This maps to PhysicalFileProtectionGroupParams in the BRS SDK
type VPCVSIProtectionParams struct {
	// File-level backup paths
	IncludePaths []string `json:"includePaths,omitempty"` // Paths to include in backup
	ExcludePaths []string `json:"excludePaths,omitempty"` // Paths to exclude from backup

	// VSS Writers (Windows-specific)
	ExcludedVssWriters []string `json:"excludedVssWriters,omitempty"` // VSS writer names to exclude

	// Indexing settings
	EnableIndexing bool `json:"enableIndexing"` // Enable file indexing for search and recovery

	// Deduplication settings
	PerformSourceSideDeduplication bool    `json:"performSourceSideDeduplication"`    // Enable source-side dedup
	PerformBrickBasedDeduplication bool    `json:"performBrickBasedDeduplication"`    // Enable brick-based dedup
	DedupExclusionSourceIds        []int64 `json:"dedupExclusionSourceIds,omitempty"` // Source IDs to exclude from dedup

	// Quiesce settings
	QuiesceFilesystem        bool `json:"quiesceFilesystem"`        // Quiesce filesystem before backup
	ContinueOnQuiesceFailure bool `json:"continueOnQuiesceFailure"` // Continue if quiesce fails

	// CoBMR (Cohesity Bare Metal Recovery)
	CobmrBackup bool `json:"cobmrBackup"` // Enable CoBMR backup

	// Pre/Post scripts
	PreBackupScript   string `json:"preBackupScript,omitempty"`  // Script to run before backup
	PostBackupScript  string `json:"postBackupScript,omitempty"` // Script to run after backup
	ScriptTimeout     int64  `json:"scriptTimeout,omitempty"`    // Script timeout in seconds
	FailOnScriptError bool   `json:"failOnScriptError"`          // Fail backup if script fails

	// Global exclude settings
	GlobalExcludePaths []string `json:"globalExcludePaths,omitempty"` // Global paths to exclude
	GlobalExcludeFS    []string `json:"globalExcludeFS,omitempty"`    // Global filesystems to exclude

	// Error handling
	IgnorableErrors []string `json:"ignorableErrors,omitempty"` // Errors to ignore (e.g., "kEOF", "kNonExistent")

	// Parallel execution
	AllowParallelRuns bool `json:"allowParallelRuns"` // Allow parallel backup runs

	// Symlink handling
	FollowSymlinks bool `json:"followSymlinks"` // Follow symbolic links
}

// BackupParams holds parameters for running a backup
type BackupParams struct {
	BackupType            BackupType     `json:"backupType"`                  // "full", "incremental"
	TargetBackupObjectIDs []BackupObject `json:"target_object_ids,omitempty"` // If empty, all objects are protected.
}
type BackupObject struct {
	ID *int64 `json:"id" validate:"required"`
}

// RESTORE DETAILED TYPES

// Common  RestoreParams holds INPUT parameters for running a restore
type RestoreParams struct {
	Name                    string `json:"name"`                    // Restore name
	BackupPositionFromFirst int64  `json:"backupPositionFromFirst"` // Backup position (0=latest)
}

type PhysicalRecoveryAction string

const (
	RecoveryAction_RecoverPhysicalVolumes PhysicalRecoveryAction = "RecoverPhysicalVolumes"
	RecoveryAction_RecoverFiles           PhysicalRecoveryAction = "RecoverFiles"
	RecoveryAction_InstantVolumeMount     PhysicalRecoveryAction = "InstantVolumeMount"
	RecoveryAction_RecoverSystem          PhysicalRecoveryAction = "RecoverSystem"
)

// VsiVpcRestoreParams holds INPUT parameters for running a restore on VSI (Physical environment).
// This maps to RecoverPhysicalParams in the BRS SDK
type VsiVpcRestoreParams struct {
	// Recovery action type - determines which recovery operation to perform
	RecoveryAction PhysicalRecoveryAction `json:"recoveryAction"` //

	// Target registration ID for recovery
	TargetRegistrationID int64 `json:"targetRegistrationID,omitempty"`

	// Volume recovery parameters (for RecoverPhysicalVolumes action)
	RecoverVolumeParams *RecoverVolumeParams `json:"recoverVolumeParams,omitempty"`

	// File and folder recovery parameters (for RecoverFiles action)
	RecoverFileAndFolderParams *RecoverFileAndFolderParams `json:"recoverFileAndFolderParams,omitempty"`

	// Mount volume parameters (for InstantVolumeMount action)
	MountVolumeParams *MountVolumeParams `json:"mountVolumeParams,omitempty"`

	// System recovery parameters (for RecoverSystem action)
	SystemRecoveryParams *SystemRecoveryParams `json:"systemRecoveryParams,omitempty"`
}

// RecoverVolumeParams holds parameters for volume recovery
type RecoverVolumeParams struct {
	MountTargetID      int64           `json:"mountTargetId"`                // Target entity ID where volumes are mounted
	VolumeMapping      []VolumeMapping `json:"volumeMapping,omitempty"`      // Source to destination volume mapping
	ForceUnmountVolume bool            `json:"forceUnmountVolume,omitempty"` // Force unmount if already mounted
	VlanConfigID       *int64          `json:"vlanConfigId,omitempty"`       // VLAN configuration ID
}

// VolumeMapping defines source to destination volume mapping
type VolumeMapping struct {
	SourceVolumeGUID      string `json:"sourceVolumeGuid"`      // Source volume GUID
	DestinationVolumeGUID string `json:"destinationVolumeGuid"` // Destination volume GUID
}

// RecoverFileAndFolderParams holds parameters for file/folder recovery
type RecoverFileAndFolderParams struct {
	FilesAndFolders           []FileAndFolderInfo `json:"filesAndFolders"`                     // Files and folders to recover
	RestoreToOriginalPaths    bool                `json:"restoreToOriginalPaths,omitempty"`    // Restore to original paths
	OverwriteExisting         bool                `json:"overwriteExisting,omitempty"`         // Overwrite existing files
	AlternateRestoreDirectory string              `json:"alternateRestoreDirectory,omitempty"` // Alternate restore directory
	PreserveAttributes        bool                `json:"preserveAttributes,omitempty"`        // Preserve file attributes
	PreserveTimestamps        bool                `json:"preserveTimestamps,omitempty"`        // Preserve timestamps
	PreserveACLs              bool                `json:"preserveAcls,omitempty"`              // Preserve ACLs
	ContinueOnError           bool                `json:"continueOnError,omitempty"`           // Continue on error
	SaveSuccessFiles          bool                `json:"saveSuccessFiles,omitempty"`
	RestoreEntityType         string              `json:"restoreEntityType,omitempty"`
	VlanConfigID              *int64              `json:"vlanConfigId,omitempty"` // VLAN configuration ID
}

// FileAndFolderInfo defines file or folder to recover
type FileAndFolderInfo struct {
	AbsolutePath       string `json:"absolutePath"`                 // Absolute path of file/folder
	IsDirectory        bool   `json:"isDirectory,omitempty"`        // Is this a directory
	IsViewFileRecovery bool   `json:"isViewFileRecovery,omitempty"` // Is view file recovery
}

// MountVolumeParams holds parameters for mounting volumes
type MountVolumeParams struct {
	MountToOriginalTarget bool     `json:"mountToOriginalTarget"`   // Mount to original target
	VolumeNames           []string `json:"volumeNames,omitempty"`   // Volume names to mount
	ReadOnlyMount         bool     `json:"readOnlyMount,omitempty"` // Read-only mount
	VlanConfigID          *int64   `json:"vlanConfigId,omitempty"`  // VLAN configuration ID
}

// SystemRecoveryParams holds parameters for system recovery
type SystemRecoveryParams struct {
	FullNasPath string `json:"fullNasPath"` // Full NAS path for recovery
}

// KubernetesRestoreParams holds INPUT parameters for running a Kubernetes restore operation.
// This struct supports both simple restores (all namespaces with same settings) and
// complex restores (different settings per namespace).
//
// Simple restore example (all namespaces use same storage class mapping):
//
//	{
//	  "RecoverObjectSpec": {
//	    "storageClasses": [{"old": "class-A", "new": "class-B"}]
//	  }
//	}
//
// Complex restore example (namespace-specific overrides):
//
//	{
//	  "RecoverObjectSpec": {
//	    "storageClasses": [{"old": "class-A", "new": "class-B"}]  // Default for all
//	  },
//	  "RecoverMultipleObjects": [
//	    {
//	      "SnapshotInfo": {"NamespaceInfo": {"namespace": "app-1"}},
//	      "RecoverObjectSpec": {
//	        "storageClasses": [{"old": "class-A", "new": "class-C"}]  // Override for app-1
//	      }
//	    }
//	  ],
//	  "IncludeNamespaces": ["app-1", "app-2"]  // Only restore these 2 namespaces
//	}
type KubernetesRestoreParams struct {
	CleanupRequired bool `json:"cleanupRequired,omitempty"`
	Expectation     bool `json:"expectation,omitempty"`

	// RecoverMultipleObjects allows specifying different restore settings for different namespaces.
	// Each entry specifies a namespace (by name) and its specific recovery settings.
	// Namespace-level settings OVERRIDE top-level RecoverObjectSpec settings.
	RecoverMultipleObjects []RecoverObject `json:"recoverObjects,omitempty"`

	// RecoverObjectSpec provides DEFAULT restore settings that apply to ALL namespaces.
	// Individual namespaces can override these settings via RecoverMultipleObjects.
	// Example: Set default storage class mapping that applies to all namespaces.
	RecoverObjectSpec *RecoverObjectSpec `json:"RecoverObjectSpec,omitempty"`

	// IncludeNamespaces filters which namespaces to restore from the backup run.
	// If empty, ALL namespaces are restored.
	// If specified, ONLY the listed namespaces are restored.
	// Example: ["app-1", "app-2"] will restore only these 2 namespaces.
	IncludeNamespaces []string `json:"includeNamespaces,omitempty"`

	SkipClusterCompatibilityCheck   bool                             `json:"skipClusterCompatibilityCheck,omitempty"`
	RenameRecoveredNamespacesParams *RenameRecoveredNamespacesParams `json:"renameRecoveredNamespacesParams,omitempty"`
	RecoverToNewTarget              bool                             `json:"recoverToNewTarget,omitempty"`
	ZoneMappings                    []MigrationMapParams             `json:"zoneMapping,omitempty"`
	RegionMapping                   *MigrationMapParams              `json:"regionMapping,omitempty"`
}

type MigrationMapParams struct {
	Source string `json:"source,omitempty"`
	Target string `json:"target,omitempty"`
}

// RecoverObject represents a single namespace to be recovered with its specific settings.
// Used in RecoverMultipleObjects array to specify namespace-level overrides.
//
// Example: Restore namespace "app-1" with specific storage class mapping
//
//	{
//	  "SnapshotInfo": {
//	    "NamespaceInfo": {"namespace": "app-1"}
//	  },
//	  "RecoverObjectSpec": {
//	    "storageClasses": [{"old": "class-A", "new": "class-C"}]
//	  }
//	}
type RecoverObject struct {
	// SnapshotInfo identifies which namespace to recover.
	// Users typically provide the namespace NAME (not snapshot ID).
	// The system will resolve the namespace name to its snapshot ID internally.
	SnapshotInfo *SnapshotInfo `json:"snapshotInfo,omitempty"`

	// RecoverObjectSpec defines HOW to recover this specific namespace.
	// This OVERRIDES the top-level RecoverObjectSpec for this namespace only.
	// Example: Different storage class mapping for this namespace.
	RecoverObjectSpec *RecoverObjectSpec `json:"recoverObjectSpec,omitempty"`
}

// RecoverObjectSpec defines the recovery settings for a namespace.
// This can be used at two levels:
// 1. Top-level (applies to ALL namespaces as default)
// 2. Namespace-level (overrides top-level for specific namespace)
//
// Example: Map old storage class to new storage class
//
//	{
//	  "storageClasses": [
//	    {"old": "ibmc-vpc-block-5iops-tier", "new": "ibmc-vpc-block-10iops-tier"}
//	  ]
//	}
type RecoverObjectSpec struct {
	IncludeObjects         *K8sObject            `json:"includeObjects,omitempty"`         // Kubernetes resources to include in restore
	ExcludeObjects         *K8sObject            `json:"excludeObjects,omitempty"`         // Kubernetes resources to exclude from restore
	UseStorageClassMapping *bool                 `json:"useStorageClassMapping,omitempty"` // Whether to use storage class mapping
	StorageClasses         []StorageClassMapping `json:"storageClasses,omitempty"`         // Storage class mappings (old -> new)
	RestoreOnlyPvc         bool                  `json:"restoreOnlyPvc,omitempty"`         // Restore only PVCs (not other resources)
}

// info for the snapshot to be recovered. User can provide either NamespaceInfo or can directly provide SnapshotID
type SnapshotInfo struct {
	NamespaceInfo       *NamespaceInfo `json:"namespaceInfo,omitempty"`       // namespaceinfo to be restored
	SnapshotID          string         `json:"snapshotID,omitempty"`          // snapshot to be restored
	ProtectionGroupId   string         `json:"protectionGroupId,omitempty"`   // protectionGroupId of the snapshot
	ProtectionGroupName string         `json:"protectionGroupName,omitempty"` // protectionGroup nameof the snapshot
}

// Namespace info to fetch the snapshot to be recovered
type NamespaceInfo struct {
	Namespace               string    `json:"namespace,omitempty"`     // namespace to be restored
	BackupPositionFromFirst int64     `json:"backupPositionFromFirst"` // snapshot to be recovered
	Timestamp               time.Time `json:"timestamp"`
}

// kubernetes object that defines labels and resources to be filtered
type K8sObject struct {
	SelectedLabels    *SelectedLabels `json:"selectedLabels,omitempty"`    // selected k8s labels
	SelectedResources []ResourceInfo  `json:"selectedResources,omitempty"` // selected k8s resources
}

type SelectedLabels struct {
	LabelCombination string     `json:"labelCombination,omitempty"` // "AND" or "OR"
	Labels           []K8sLabel `json:"labels,omitempty"`
}
type K8sLabel struct {
	Key   string `json:"key,omitempty"`
	Value string `json:"value,omitempty"`
}

type ResourceInfo struct {
	ResourceKind     string             `json:"resourceType,omitempty"`
	ResourceApiGroup string             `json:"resourceApiGroup,omitempty"`
	ResourceList     []ResourceInstance `json:"resourceList,omitempty"`
}

type ResourceInstance struct {
	Id   int64  `json:"id,omitempty"`
	Name string `json:"name,omitempty"`
}

type RenameRecoveredNamespacesParams struct {
	Prefix string `json:"prefix,omitempty"`
	Suffix string `json:"suffix,omitempty"`
}

// resolve namespace name (to be recovered) to namespace Id, parent Id and group ID of the snapshot
type ResolveNamespaceResult struct {
	NamespaceId int64
	SourceId    int64
	GroupId     string
}

// StorageClassMapping maps old storage class to new (for cross-cloud migration)
type StorageClassMapping struct {
	Old string `json:"old"` // Old storage class name (e.g., "gp2" on AWS)
	New string `json:"new"` // New storage class name (e.g., "standard-rwo" on GCP)
}

// RestoreParams holds INPUT parameters for running a restore on VpcVsi
type PhysicalRestoreParams struct {
	Files string `json:"files"` // files to include
}

// RestoreLabels defines label-based filtering for restore
// type RestoreLabels struct {
// 	PersistentVolumeClaim KeyValueFilter `json:"persistentVolumeClaim"`
// }

// KeyValueFilter defines key-value based filtering
type KeyValueFilter struct {
	LogicRule string `json:"logicRule"` // "AND", "OR"
	Inclusion string `json:"inclusion"` // Comma-separated key=value pairs
	Exclusion string `json:"exclusion"` // Comma-separated key=value pairs
}

// RestoreNamespaceResource defines namespace-level restore configuration
// type RestoreNamespaceResource struct {
// 	Resources      RestoreResources      `json:"resources"`      // Resource filtering
// 	StorageClasses []StorageClassMapping `json:"storageClasses"` // Storage class mappings

// }

// type FilterMode string

// const (
// 	FilterMode_Include FilterMode = "Include"
// 	FilterMode_Exclude FilterMode = "Exclude"
// )

// RestoreResources defines resource-level filtering
// type RestoreResources struct {
// 	PersistentVolumeClaim KeyValueFilter  `json:"persistentVolumeClaim"` // PVC filtering
// 	K8sObjects            K8sObjectFilter `json:"k8sObjects"`            // K8s object filtering
// }

// K8sObjectFilter defines K8s object type filtering
// type K8sObjectFilter struct {
// 	FilterMode FilterMode         `json:"filterMode"` // "Include", "Exclude"
// 	Inclusion  []ResourceTypeList `json:"inclusion"`  // Resource types to include
// 	Exclusion  []ResourceTypeList `json:"exclusion"`  // Resource types to exclude
// }

// ResourceTypeList lists resource types
type ResourceTypeList struct {
	ResourceTypes []string `json:"resourceTypes"` // e.g., ["Deployment", "Service"]
}

// MIGRATION PARAMETERS

// MigrationParams holds INPUT parameters for complete migration workflow
type MigrationParams struct {
	// Connection Configuration
	SourceConnectionParams *ConnectionParams `json:"sourceConnectionParams"` // Source connection config
	TargetConnectionParams *ConnectionParams `json:"targetConnectionParams"` // Target connection config

	// Protection Configuration
	ProtectionGroupName   string                 `json:"protectionGroupName"`   // Protection group name
	ProtectionGroupParams *ProtectionGroupParams `json:"protectionGroupParams"` // Full protection config
	CreateDefaultPolicy   bool                   `json:"createDefaultPolicy"`   // Auto-create policy if true
	PolicyName            string                 `json:"policyName"`            // Policy name (if creating)

	// Backup Configuration
	BackupParams  *BackupParams `json:"backupParams"`  // Backup configuration
	WaitForBackup bool          `json:"waitForBackup"` // Wait for backup completion

	// Restore Configuration
	RestoreParams  *RestoreParams `json:"restoreParams"`  // Restore configuration
	WaitForRestore bool           `json:"waitForRestore"` // Wait for restore completion
}

// PolicyParams holds INPUT parameters for creating a policy
type PolicyParams struct {
	Name                string               `json:"name"`        // Policy name
	Description         string               `json:"description"` // Policy description
	CreateIfMissing     bool                 `yaml:"createIfMissing"`
	CleanupRequired     bool                 `yaml:"cleanupRequired"`
	IncrementalBackup   *IncrementalBackup   `yaml:"incrementalBackup"`
	FullBackups         []FullBackup         `yaml:"FullBackup"`
	PeriodicFullBackup  PeriodicFullBackup   `yaml:"periodicFullBackup"`
	BackupType          string               `json:"backupType"` // Backup type: "full", "incremental"
	RetryOption         *RetryOption         `yaml:"retryOption"`
	DataRetention       *DataRetention       `yaml:"DataRetention"` // Specifies the retention of a backup
	Expectation         bool                 `yaml:"expectation"`
	PrimaryBackupTarget *PrimaryBackupTarget `yaml:"primaryBackupTarget"`
}

// FullScheduleAndRetention : Specifies the settings to schedule the full backup and retention for each schedule.
type FullBackup struct {
	Unit          *FullSchedule_Unit `yaml:"unit"`
	DaySchedule   *UnitDaySchedule   `yaml:"daySchedule"`
	WeekSchedule  *WeekSchedule      `yaml:"weekSchedule"`
	MonthSchedule *MonthSchedule     `yaml:"monthSchedule"`
	YearSchedule  *YearSchedule      `yaml:"yearSchedule"`
	Retention     *DataRetention     `yaml:"retention"` // Specifies the retention of a backup.
}

type UnitDaySchedule struct {
	Every int64 `yaml:"every"`
}

type WeekSchedule struct {
	OnDayOfWeek []DayOfWeek `yaml:"onDayOfWeek"`
}

type MonthSchedule struct {
	OnDate int64 `yaml:"onDate"`
}

type YearSchedule struct {
	OnDayOfYear *DayOfYear `yaml:"onDayOfYear"`
}

type RetryOption struct {
	NumberOfRetry       int64 `yaml:"numberOfRetry"`
	WaitInCaseOfFailure int64 `yaml:"waitInCaseOfFailure"`
}

type DataRetention struct {
	RetainFor      int64           `yaml:"retainFor"`
	Unit           *Retention_Unit `yaml:"unit"`
	DataLockConfig *DataLockConfig `yaml:"dataLockConfig"` // DataLockConfig : Specifies WORM retention type for the snapshots.
}

type DataLockConfig struct {
	Mode                       *DataLockConfig_Mode `json:"mode" validate:"required"`             // Specifies the type of WORM retention type.
	Unit                       *DataLockConfig_Unit `json:"unit" validate:"required"`             // Specificies the Retention Unit of a dataLock
	Duration                   int64                `json:"duration" validate:"required"`         // Specifies the duration for a dataLock.
	EnableWormOnExternalTarget bool                 `json:"enableWormOnExternalTarget,omitempty"` // Specifies whether objects in the external target associated with this policy need to be made immutable.
}

type IncrementalBackup struct {
	Unit           *IncrementalBackup_Unit `yaml:"unit"`
	DaySchedule    *UnitDaySchedule        `yaml:"daySchedule"`
	MinuteSchedule *UnitDaySchedule        `yaml:"minuteSchedule"`
	HourSchedule   *UnitDaySchedule        `yaml:"hourSchedule"`
	WeekSchedule   *WeekSchedule           `yaml:"hourSchedule"`
	MonthSchedule  *MonthSchedule          `yaml:"monthchedule"`
	YearSchedule   *YearSchedule           `yaml:"yearSchedule"`
}

// Specifies the primary backup target settings for regular backups.
type PrimaryBackupTarget struct {
	TargetType             *PrimaryBackupTarget_TargetType `json:"targetType,omitempty"`             // Specifies the primary backup location where backups will be stored
	ArchivalTargetSettings *PrimaryArchivalTarget          `json:"archivalTargetSettings,omitempty"` // Specifies the primary archival settings.
	UseDefaultBackupTarget bool                            `json:"useDefaultBackupTarget,omitempty"` // Specifies if the default primary backup target must be used for backups.
}

// PrimaryArchivalTarget : Specifies the primary archival settings.
type PrimaryArchivalTarget struct {
	// Specifies the Archival target id to take primary backup.
	TargetID *int64 `json:"targetId" validate:"required"`

	// Specifies the Archival target name where Snapshots are copied.
	TargetName *string `json:"targetName,omitempty"`
}

type PeriodicFullBackup struct {
	Every string `yaml:"every"`
	For   int    `yaml:"for"`
	Unit  string `yaml:"unit"`
}

// PolicyResult represents the result of a policy operation
type PolicyResult struct {
	ID            string    `json:"id"`            // Policy ID
	Name          string    `json:"name"`          // Policy name
	Description   string    `json:"description"`   // Policy description
	RetentionDays int64     `json:"retentionDays"` // Retention period
	BackupType    string    `json:"backupType"`    // Backup type
	Status        string    `json:"status"`        // Policy status
	CreatedAt     time.Time `json:"createdAt"`     // Creation timestamp
}

// StepResult represents the result of a migration step
type StepResult struct {
	Step        string    `json:"step"`        // Step name: "source_setup", "target_setup", etc.
	Status      string    `json:"status"`      // Step status: "completed", "failed", "skipped"
	Message     string    `json:"message"`     // Step message
	StartedAt   time.Time `json:"startedAt"`   // Step start time
	CompletedAt time.Time `json:"completedAt"` // Step completion time
	Duration    float64   `json:"duration"`    // Step duration in seconds
}

// CLEANUP PARAMETERS AND RESULTS

// CleanupScope defines the scope of cleanup operation
type CleanupScope string

const (
	CleanupScopeAll      CleanupScope = "all"      // Clean everything
	CleanupScopeSource   CleanupScope = "source"   // Only source resources
	CleanupScopeTarget   CleanupScope = "target"   // Only target resources
	CleanupScopeSpecific CleanupScope = "specific" // Specific resources by ID
)

// ResourceType defines the type of resource
type ResourceType string

const (
	ResourceTypeConnection      ResourceType = "connection"
	ResourceTypeConnector       ResourceType = "connector"
	ResourceTypeRegistration    ResourceType = "registration"
	ResourceTypeProtectionGroup ResourceType = "protection_group"
	ResourceTypePolicy          ResourceType = "policy"
	ResourceTypeBackup          ResourceType = "backup"
	ResourceTypeRestore         ResourceType = "restore"
)

// CleanupParams holds INPUT parameters for cleanup operation
type CleanupParams struct {
	// What to clean up
	CleanupScope CleanupScope        `json:"cleanupScope"` // Scope of cleanup
	ResourceIDs  *CleanupResourceIDs `json:"resourceIds"`  // Specific resource IDs

	// Safety options
	DryRun       bool `json:"dryRun"`       // Preview without deleting
	Force        bool `json:"force"`        // Skip safety checks
	KeepBackups  bool `json:"keepBackups"`  // Don't delete backups
	KeepPolicies bool `json:"keepPolicies"` // Don't delete policies

	// Selective cleanup
	CleanupTypes []ResourceType `json:"cleanupTypes"` // Which resource types to clean

	// Confirmation
	ConfirmationToken string `json:"confirmationToken"` // Required for actual deletion
}

// CleanupResourceIDs holds specific resource IDs to clean
type CleanupResourceIDs struct {
	SourceConnectionID   string   `json:"sourceConnectionId"`
	SourceConnectorID    string   `json:"sourceConnectorId"`
	SourceRegistrationID int64    `json:"sourceRegistrationId"`
	TargetConnectionID   string   `json:"targetConnectionId"`
	TargetConnectorID    string   `json:"targetConnectorId"`
	TargetRegistrationID int64    `json:"targetRegistrationId"`
	ProtectionGroupID    string   `json:"protectionGroupId"`
	PolicyID             string   `json:"policyId"`
	BackupIDs            []string `json:"backupIds"`
	RestoreIDs           []string `json:"restoreIds"`
}

// CleanupResult represents the result of a cleanup operation
type CleanupResult struct {
	// Summary
	Status           string `json:"status"`           // "completed", "partial", "failed"
	DryRun           bool   `json:"dryRun"`           // Was this a dry run?
	ResourcesDeleted int    `json:"resourcesDeleted"` // Number deleted
	ResourcesSkipped int    `json:"resourcesSkipped"` // Number skipped
	ResourcesFailed  int    `json:"resourcesFailed"`  // Number failed

	// Detailed results
	DeletedResources []DeletedResource `json:"deletedResources"` // What was deleted
	SkippedResources []SkippedResource `json:"skippedResources"` // What was skipped
	FailedResources  []FailedResource  `json:"failedResources"`  // What failed

	// Timing
	StartedAt   time.Time `json:"startedAt"`   // Cleanup start time
	CompletedAt time.Time `json:"completedAt"` // Cleanup completion time
	Duration    float64   `json:"duration"`    // Duration in seconds

	// Steps
	StepResults []StepResult `json:"stepResults"` // Step-by-step results
}

// DeletedResource represents a successfully deleted resource
type DeletedResource struct {
	ResourceType ResourceType `json:"resourceType"` // Type of resource
	ResourceID   string       `json:"resourceId"`   // Resource ID
	ResourceName string       `json:"resourceName"` // Resource name
	DeletedAt    time.Time    `json:"deletedAt"`    // Deletion timestamp
}

// SkippedResource represents a resource that was skipped
type SkippedResource struct {
	ResourceType ResourceType `json:"resourceType"` // Type of resource
	ResourceID   string       `json:"resourceId"`   // Resource ID
	ResourceName string       `json:"resourceName"` // Resource name
	Reason       string       `json:"reason"`       // Why skipped: "in_use", "not_found", "keep_requested"
}

// FailedResource represents a resource that failed to delete
type FailedResource struct {
	ResourceType ResourceType `json:"resourceType"` // Type of resource
	ResourceID   string       `json:"resourceId"`   // Resource ID
	ResourceName string       `json:"resourceName"` // Resource name
	Error        string       `json:"error"`        // Error message
}

// CONTROL PARAMETERS AND RESULTS

// ControlAction defines the control action to perform
type ControlAction string

const (
	ControlActionPause  ControlAction = "pause"  // Pause operation
	ControlActionAbort  ControlAction = "abort"  // Abort and cleanup
	ControlActionResume ControlAction = "resume" // Resume paused
	ControlActionCancel ControlAction = "cancel" // Cancel without cleanup
)

// OperationType defines the type of operation to control
type OperationType string

const (
	OperationTypeMigration OperationType = "migration"
	OperationTypeBackup    OperationType = "backup"
	OperationTypeRestore   OperationType = "restore"
	OperationTypeSetup     OperationType = "setup"
	OperationTypeProtect   OperationType = "protect"
)

// ControlParams holds INPUT parameters for control operation
type ControlParams struct {
	// What to control
	ControlAction   ControlAction `json:"controlAction"`   // Action: "pause", "abort", "resume", "cancel"
	TargetOperation OperationType `json:"targetOperation"` // Operation type
	OperationID     string        `json:"operationId"`     // ID of operation to control
	GroupID         string        `json:"groupId"`         // ID of operation to control

	// How to control
	Force             bool          `json:"force"`             // Force immediate stop (not graceful)
	CleanupOnAbort    bool          `json:"cleanupOnAbort"`    // Cleanup resources on abort
	WaitForCompletion bool          `json:"waitForCompletion"` // Wait for control action to complete
	Timeout           time.Duration `json:"timeout"`           // How long to wait for graceful stop

	// Confirmation
	ConfirmationToken string `json:"confirmationToken"` // Required for abort/cancel
}

// ControlResult represents the result of a control operation
type ControlResult struct {
	// Summary
	Status        string        `json:"status"`        // "paused", "aborted", "resumed", "cancelled"
	OperationID   string        `json:"operationId"`   // ID of controlled operation
	OperationType OperationType `json:"operationType"` // Type of operation
	PreviousState string        `json:"previousState"` // State before control action
	CurrentState  string        `json:"currentState"`  // State after control action

	// Actions taken
	ActionsTaken     []ControlAction `json:"actionsTaken"`     // Actions performed
	ResourcesCleaned int             `json:"resourcesCleaned"` // Resources cleaned (if abort)

	// Cleanup details (if abort)
	CleanupResult *CleanupResult `json:"cleanupResult,omitempty"` // Detailed cleanup results

	// Timing
	StartedAt   time.Time `json:"startedAt"`   // Control start time
	CompletedAt time.Time `json:"completedAt"` // Control completion time
	Duration    float64   `json:"duration"`    // Duration in seconds

	// Steps
	StepResults []StepResult `json:"stepResults"` // Step-by-step results
}

// TimeRange represents a time range for filtering
type TimeRange struct {
	Start time.Time `json:"start"` // Range start time
	End   time.Time `json:"end"`   // Range end time
}
