/*******************************************************************************
 * IBM Confidential
 * OCO Source Materials
 * (C) Copyright IBM Corp. 2026 All Rights Reserved.
 * The source code for this program is not  published or otherwise divested of
 * its trade secrets, irrespective of what has been deposited with
 * the U.S. Copyright Office.
 ******************************************************************************/
package vpcvsi

import (
	"context"
	"fmt"

	"github.com/IBM/go-sdk-core/v5/core"
	"github.com/IBM/ibm-backup-recovery-sdk-go/backuprecoveryv1"
	"github.com/IBM/ibm-backup-recovery-sdk-go/migrationv2/common/errors"
	"github.com/IBM/ibm-backup-recovery-sdk-go/migrationv2/common/types"
	"github.com/IBM/ibm-backup-recovery-sdk-go/migrationv2/middlelevel/connectors"
	"github.com/IBM/ibm-backup-recovery-sdk-go/migrationv2/middlelevel/datasources"
)

type PhysicalDataSourceConfig struct {
	Name                   string
	VsiEndpoint            string
	ExcludeFolders         string
	IncludeFolders         string
	VPCVSIProtectionParams *types.VPCVSIProtectionParams
	VPCVSIRestoreParams    *types.VsiVpcRestoreParams
	// VPCVSIParams     *VPCVSIProtectionParams     `json:"vpcvsiParams,omitempty"`     // VPC VSI-specific params
	// etc ..
}

func (physicalDataSourceConfig *PhysicalDataSourceConfig) GetDataSourceType() types.DataSourceType {
	return types.DataSourceVPCVSI
}

func (physicalDataSourceConfig *PhysicalDataSourceConfig) CreateDataSource(brsClient types.BRSClientWrapperInterface) (datasources.DataSource, error) {
	return NewVPCVSIDataSource(physicalDataSourceConfig.Name, physicalDataSourceConfig, brsClient)
}

func (physicalDataSourceConfig *PhysicalDataSourceConfig) Validate() error {
	return nil
}

// VPCVSIDataSource implements DataSource for VPC Virtual Server Instances
type VPCVSIDataSource struct {
	Type                   types.DataSourceType
	Name                   string
	Config                 *PhysicalDataSourceConfig
	VsiEndpoint            string
	VPCID                  string
	VSIIDs                 []string
	Region                 string
	Subnet                 string
	SecurityGroup          string
	VPCVSIProtectionParams *types.VPCVSIProtectionParams
	VPCVSIRestoreParams    *types.VsiVpcRestoreParams
	ConnectorDeployer      connectors.ConnectorDeployer
	brsClient              types.BRSClientWrapperInterface
}

// NewVPCVSIDataSource creates a new VPC VSI data source
func NewVPCVSIDataSource(name string, config *PhysicalDataSourceConfig, brsClient types.BRSClientWrapperInterface) (*VPCVSIDataSource, error) {
	if config == nil {
		return nil, errors.NewInvalidDataSourceError("physical data source config cannot be nil")
	}

	v := &VPCVSIDataSource{
		Type:                   types.DataSourceVPCVSI,
		Name:                   name,
		Config:                 config,
		VsiEndpoint:            config.VsiEndpoint,
		VPCVSIProtectionParams: config.VPCVSIProtectionParams,
		VPCVSIRestoreParams:    config.VPCVSIRestoreParams,
		brsClient:              brsClient,
	}

	return v, nil
}

// GetType returns the data source type
func (v *VPCVSIDataSource) GetType() types.DataSourceType {
	return v.Type
}

// GetName returns the data source name
func (v *VPCVSIDataSource) GetName() string {
	return v.Name
}

// Validate validates VPC VSI-specific configuration
func (v *VPCVSIDataSource) Validate() error {
	if v.Name == "" {
		return errors.NewInvalidDataSourceError("data source name is required")
	}
	if v.VPCID == "" {
		return errors.NewInvalidDataSourceError("vpc_id is required")
	}
	if len(v.VSIIDs) == 0 { // one or more can be passed
		return errors.NewInvalidDataSourceError("at least one vsi_id is required")
	}
	return nil
}

// CreateConnection creates a connection for VPC VSI
func (v *VPCVSIDataSource) CreateConnection(ctx context.Context, connectionName string) (*backuprecoveryv1.CreateDataSourceConnectionOptions, error) {
	// Build connection parameters for VPC VSI
	connectionParams := &backuprecoveryv1.CreateDataSourceConnectionOptions{
		ConnectionName: &connectionName,
	}

	// TODO: Add VPC VSI-specific connection parameters

	return connectionParams, nil
}

// GetConnection retrieves connection details
func (v *VPCVSIDataSource) GetConnection(ctx context.Context, connectionID string) (*backuprecoveryv1.GetDataSourceConnectionsOptions, error) {
	// TODO: Implement get connection logic, use low-level APIs
	return nil, fmt.Errorf("GetConnection not yet implemented for VPC VSI")
}

// DeleteConnection deletes a connection
func (v *VPCVSIDataSource) DeleteConnection(ctx context.Context, connectionID string) error {
	// TODO: Implement delete connection logic, by uisng low level APIs
	return fmt.Errorf("DeleteConnection not yet implemented for VPC VSI")
}

// RegisterSource registers the VPC VSI as a protection source
func (v *VPCVSIDataSource) RegisterSourceParams(ctx context.Context, connectionID string) (*backuprecoveryv1.RegisterProtectionSourceOptions, error) {
	// TODO: Implement registration logic for VPC VSI, by using low-level APIs
	env := backuprecoveryv1.RegisterProtectionSourceOptions_Environment_Kphysical
	registrationOptions := &backuprecoveryv1.RegisterProtectionSourceOptions{
		Environment: &env,
		PhysicalParams: &backuprecoveryv1.PhysicalSourceRegistrationParams{
			Endpoint: core.StringPtr(v.VsiEndpoint),
		},
	}
	return registrationOptions, fmt.Errorf("RegisterSource not yet implemented for VPC VSI")
}

// GetRegistration retrieves registration details
func (v *VPCVSIDataSource) GetRegistration(ctx context.Context, registrationID int64) (*backuprecoveryv1.GetSourceRegistrationsOptions, error) {
	// TODO: Implement get registration logic, by using low-level APIs

	return nil, fmt.Errorf("GetRegistration not yet implemented for VPC VSI")
}

// UnregisterSource unregisters the protection source
func (v *VPCVSIDataSource) UnregisterSource(ctx context.Context, registrationID int64) error {
	// TODO: Implement unregister logic, by using low-level APIs
	return fmt.Errorf("UnregisterSource not yet implemented for VPC VSI")
}

// CreateProtectionGroup creates a protection group for VPC VSI
func (v *VPCVSIDataSource) CreateProtectionGroup(ctx context.Context, registrationID int64, params *types.ProtectionGroupParams) (*backuprecoveryv1.CreateProtectionGroupOptions, error) {
	// Validate params
	if registrationID == 0 {
		return nil, errors.NewSDKError(errors.ErrCodeInvalidInput, "registrationID is required", nil)
	}
	if params == nil {
		return nil, errors.NewInvalidConfigError("protection group params are required", nil)
	}

	// Validate VPC VSI-specific params are provided
	if v.VPCVSIProtectionParams == nil {
		return nil, errors.NewInvalidConfigError("VPCVSIProtectionParams are required", nil)
	}

	// Build physical protection group params using the helper
	physicalParams, err := v.GetPhysicalProtectionGroupParams(registrationID)
	if err != nil {
		return nil, fmt.Errorf("failed to build physical protection group params: %w", err)
	}

	// Set environment to physical
	env := backuprecoveryv1.CreateProtectionGroupOptions_Environment_Kphysical

	// Create the protection group options
	protectionGroupOptions := &backuprecoveryv1.CreateProtectionGroupOptions{
		Environment:    &env,
		PhysicalParams: physicalParams,
	}

	return protectionGroupOptions, nil
}

// GetProtectionGroup retrieves protection group details
func (v *VPCVSIDataSource) GetProtectionGroup(ctx context.Context, groupID string) (*backuprecoveryv1.GetProtectionGroupByIdOptions, error) {
	// TODO: Implement get protection group logic, use low level api
	return nil, fmt.Errorf("GetProtectionGroup not yet implemented for VPC VSI")
}

// DeleteProtectionGroup deletes a protection group
func (v *VPCVSIDataSource) DeleteProtectionGroup(ctx context.Context, groupID string) error {
	// TODO: Implement delete protection group, need to use low-level apis
	return fmt.Errorf("DeleteProtectionGroup not yet implemented for VPC VSI")
}

// RunBackup initiates a backup for VPC VSI
func (v *VPCVSIDataSource) RunBackup(ctx context.Context, groupID string) (*types.BackupResult, error) {
	// TODO: Implement backup logic for VPC VSI, use low-level apis
	return nil, fmt.Errorf("RunBackup not yet implemented for VPC VSI")
}

// GetBackupStatus retrieves backup status
func (v *VPCVSIDataSource) GetBackupStatus(ctx context.Context, backupID string) (interface{}, error) {
	// TODO: Implement get backup status logic
	return nil, fmt.Errorf("GetBackupStatus not yet implemented for VPC VSI")
}

// RunRestore initiates a restore operation for VPC VSI (Physical environment)
func (v *VPCVSIDataSource) RunRestore(ctx context.Context, groupId, backupId string, targetRegistrationID int64, commonRestoreParams *types.RestoreParams) (*backuprecoveryv1.CreateRecoveryOptions, error) {
	if commonRestoreParams == nil {
		return nil, errors.NewInvalidConfigError("restore params are required", nil)
	}
	if v.VPCVSIRestoreParams == nil {
		return nil, errors.NewInvalidConfigError("VPC VSI restore params are required", nil)
	}

	var recoverPhysicalParams *backuprecoveryv1.RecoverPhysicalParams
	var err error

	// Determine which recovery action to perform based on the recovery action type
	switch v.VPCVSIRestoreParams.RecoveryAction {
	case backuprecoveryv1.RecoverPhysicalParams_RecoveryAction_Recoverphysicalvolumes:
		// Volume recovery
		recoverPhysicalParams, err = v.RecoverPhysicalVolumes(groupId, backupId, targetRegistrationID, v.VPCVSIRestoreParams, commonRestoreParams)
		if err != nil {
			return nil, fmt.Errorf("failed to create volume recovery params: %w", err)
		}

	case backuprecoveryv1.RecoverPhysicalParams_RecoveryAction_Recoverfiles:
		// File and folder recovery
		recoverPhysicalParams, err = v.RecoverPhysicalFiles(groupId, backupId, targetRegistrationID, v.VPCVSIRestoreParams, commonRestoreParams)
		if err != nil {
			return nil, fmt.Errorf("failed to create file recovery params: %w", err)
		}

	case backuprecoveryv1.RecoverPhysicalParams_RecoveryAction_Instantvolumemount:
		// Mount volume
		recoverPhysicalParams, err = v.MountPhysicalVolumes(groupId, backupId, targetRegistrationID, v.VPCVSIRestoreParams, commonRestoreParams)
		if err != nil {
			return nil, fmt.Errorf("failed to create mount volume params: %w", err)
		}

	case backuprecoveryv1.RecoverPhysicalParams_RecoveryAction_Recoversystem:
		// System recovery
		recoverPhysicalParams, err = v.RecoverPhysicalSystem(groupId, backupId, targetRegistrationID, v.VPCVSIRestoreParams, commonRestoreParams)
		if err != nil {
			return nil, fmt.Errorf("failed to create system recovery params: %w", err)
		}

	default:
		return nil, fmt.Errorf("unsupported recovery action: %s", v.VPCVSIRestoreParams.RecoveryAction)
	}

	// Build CreateRecoveryOptions
	snapshotEnv := backuprecoveryv1.Recovery_SnapshotEnvironment_Kphysical
	createRecoveryOptions := &backuprecoveryv1.CreateRecoveryOptions{
		SnapshotEnvironment: &snapshotEnv,
		PhysicalParams:      recoverPhysicalParams,
		Name:                &commonRestoreParams.Name,
	}

	return createRecoveryOptions, nil
}

// GetRestoreStatus retrieves restore status
func (v *VPCVSIDataSource) GetRestoreStatus(ctx context.Context, restoreID string) (interface{}, error) {
	// TODO: Implement get restore status logic, use low-level apis
	return nil, fmt.Errorf("GetRestoreStatus not yet implemented for VPC VSI")
}
