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
)

// -----Recovery-------

const (
	DefaultSnapahotWait_TotalTimeout    time.Duration = 10 * time.Minute // 10 minutes total wait time
	DefaultSnapahotWait_PollingInterval time.Duration = 30 * time.Second // Poll every 30 seconds
)

// -----Backup Runs-------

const (
	// Default time to get active backup
	GetBackup_totalTimeout = 5 * time.Minute // 1 minute total wait time
	// Default polling interval to get active backup
	GetBackup_pollingInterval = 15 * time.Second // Poll every 10 seconds
)

type BackupRun_StatusType string

// Status of the backupRuns
const (
	BackupRun_Status_Running              BackupRun_StatusType = backuprecoveryv1.ArchivalTargetResult_Status_Running
	BackupRun_Status_Canceling            BackupRun_StatusType = backuprecoveryv1.ArchivalTargetResult_Status_Canceling
	BackupRun_Status_Canceled             BackupRun_StatusType = backuprecoveryv1.ArchivalTargetResult_Status_Canceled
	BackupRun_Status_Failed               BackupRun_StatusType = backuprecoveryv1.ArchivalTargetResult_Status_Failed
	BackupRun_Status_Missed               BackupRun_StatusType = backuprecoveryv1.ArchivalTargetResult_Status_Missed
	BackupRun_Status_Succeeded            BackupRun_StatusType = backuprecoveryv1.ArchivalTargetResult_Status_Succeeded
	BackupRun_Status_SucceededWithWarning BackupRun_StatusType = backuprecoveryv1.ArchivalTargetResult_Status_Succeededwithwarning
	BackupRun_Status_Skipped              BackupRun_StatusType = backuprecoveryv1.ArchivalTargetResult_Status_Skipped
	BackupRun_Status_Accepted             BackupRun_StatusType = backuprecoveryv1.ArchivalTargetResult_Status_Accepted
)

type BackupType string

const (
	BackupType_Incremental BackupType = backuprecoveryv1.CreateProtectionGroupRunOptions_RunType_Kregular
	BackupType_Full        BackupType = backuprecoveryv1.CreateProtectionGroupRunOptions_RunType_Kfull
)

// --- protection groups------

const (
	DefaultProtectionGroupName string = "auto-protection-group-sdk-"
)

// -----protection policy-------

// custom type for all the units supported for incremental Schedule
type IncrementalBackup_Unit string

// Constants associated with the IncrementalBackup Unit
const (
	IncrementalBackup_Unit_Days    IncrementalBackup_Unit = "Days"
	IncrementalBackup_Unit_Hours   IncrementalBackup_Unit = "Hours"
	IncrementalBackup_Unit_Minutes IncrementalBackup_Unit = "Minutes"
	IncrementalBackup_Unit_Months  IncrementalBackup_Unit = "Months"
	IncrementalBackup_Unit_Weeks   IncrementalBackup_Unit = "Weeks"
	IncrementalBackup_Unit_Years   IncrementalBackup_Unit = "Years"
)

// custom type for all the days of the week
type DayOfWeek string

// Constants associated with the WeekDay Unit
const (
	WeekDay_Monday    DayOfWeek = "Monday"
	WeekDay_Sunday    DayOfWeek = "Sunday"
	WeekDay_Tuesday   DayOfWeek = "Tuesday"
	WeekDay_Wednesday DayOfWeek = "Wednesday"
	WeekDay_Thursday  DayOfWeek = "Thursday"
	WeekDay_Friday    DayOfWeek = "Friday"
	WeekDay_Saturday  DayOfWeek = "Saturday"
)

// custom type for supported days of the year
type DayOfYear string

const (
	YearDay_First DayOfYear = "First"
	YearDay_Last  DayOfYear = "Last"
)

// custom type for units supported for Full Schedule
type FullSchedule_Unit string

// Constants associated with the FullBackup Unit
const (
	FullSchedule_Unit_Days   FullSchedule_Unit = "Days"
	FullSchedule_Unit_Weeks  FullSchedule_Unit = "Weeks"
	FullSchedule_Unit_Months FullSchedule_Unit = "Months"
	FullSchedule_Unit_Years  FullSchedule_Unit = "Years"
)

// custom type for units supported for Full Schedule
type Retention_Unit string

// Constants associated with the Retention Unit
const (
	Retention_Unit_Days   Retention_Unit = "Days"
	Retention_Unit_Months Retention_Unit = "Months"
	Retention_Unit_Weeks  Retention_Unit = "Weeks"
	Retention_Unit_Years  Retention_Unit = "Years"
)

type DataLockConfig_Mode string

const (
	DataLockConfig_Mode_Administrative DataLockConfig_Mode = "Administrative"
	DataLockConfig_Mode_Compliance     DataLockConfig_Mode = "Compliance"
)

type DataLockConfig_Unit string

// Constants associated with the DataLockConfig. Unit property.
const (
	DataLockConfig_Unit_Days   DataLockConfig_Unit = "Days"
	DataLockConfig_Unit_Months DataLockConfig_Unit = "Months"
	DataLockConfig_Unit_Weeks  DataLockConfig_Unit = "Weeks"
	DataLockConfig_Unit_Years  DataLockConfig_Unit = "Years"
)

// custom type for aprimary target type
type PrimaryBackupTarget_TargetType string

const (
	PrimaryBackupTarget_TargetType_Archival PrimaryBackupTarget_TargetType = "Archival"
	PrimaryBackupTarget_TargetType_Local    PrimaryBackupTarget_TargetType = "Local"
)

// Default Retry Options
const (
	DefaultPolicyRetries  int64 = 1
	DefaultPolicyInterval int64 = 5
)

// Default Policy Incremental Schedule
const (
	DefaultPolicyIncrementalUnit            = IncrementalBackup_Unit_Days
	DefaultPolicyIncrementalFrequency int64 = 1
)

// -----Operation Names for Logging-------

// Task API Operation Names
const (
	// Connection Operations
	OpCreateConnection    = "CreateConnection"
	OpGetConnection       = "GetConnection"
	OpGetConnectionByName = "GetConnectionByName"
	OpListConnections     = "ListConnections"
	OpDeleteConnection    = "DeleteConnection"

	// Connector Operations
	OpDeployConnector          = "DeployConnector"
	OpGetConnector             = "GetConnector"
	OpGetConnectorByConnection = "GetConnectorByConnection"
	OpDeleteConnector          = "DeleteConnector"

	// Source Registration Operations
	OpRegisterSource              = "RegisterSource"
	OpGetRegistration             = "GetRegistration"
	OpGetRegistrationByConnection = "GetRegistrationByConnection"
	OpListRegistrations           = "ListRegistrations"
	OpUnregisterSource            = "UnregisterSource"

	// Protection Group Operations
	OpCreateProtectionGroup    = "CreateProtectionGroup"
	OpGetProtectionGroup       = "GetProtectionGroup"
	OpGetProtectionGroupByName = "GetProtectionGroupByName"
	OpListProtectionGroups     = "ListProtectionGroups"
	OpUpdateProtectionGroup    = "UpdateProtectionGroup"
	OpDeleteProtectionGroup    = "DeleteProtectionGroup"

	// Policy Operations
	OpCreatePolicy    = "CreatePolicy"
	OpGetPolicy       = "GetPolicy"
	OpGetPolicyByName = "GetPolicyByName"
	OpListPolicies    = "ListPolicies"
	OpUpdatePolicy    = "UpdatePolicy"
	OpDeletePolicy    = "DeletePolicy"

	// Backup Operations
	OpRunBackup                = "RunBackup"
	OpGetBackup                = "GetBackup"
	OpListBackups              = "ListBackups"
	OpGetProtectionRunProgress = "GetProtectionRunProgress"
	OpWaitForBackup            = "WaitForBackup"
	OpPauseBackup              = "PauseBackup"
	OpResumeBackup             = "ResumeBackup"
	OpAbortBackup              = "AbortBackup"

	// Restore Operations
	OpRunRestore     = "RunRestore"
	OpGetRestore     = "GetRestore"
	OpListRestores   = "ListRestores"
	OpWaitForRestore = "WaitForRestore"
	OpAbortRestore   = "AbortRestore"

	// Helper Operations
	OpBuildBackupRunResult = "buildBackupRunResult"
	OpGetRestoreProgress   = "getRestoreProgress"
	// MetaInfo Operations
	OpGetMetaInfo = "GetMetaInfo"
)

// Kubernetes DataSource Operation Names
const (
	OpRegisterSourceParams      = "RegisterSourceParams"
	OpCreateProtectionGroupK8s  = "CreateProtectionGroup"
	OpRunRestoreK8s             = "RunRestore"
	OpConstructMetaInfoK8s      = "ConstructMetaInfoOptions"
)
