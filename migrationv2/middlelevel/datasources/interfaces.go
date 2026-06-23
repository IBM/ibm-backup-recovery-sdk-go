/*******************************************************************************
 * IBM Confidential
 * OCO Source Materials
 * (C) Copyright IBM Corp. 2026 All Rights Reserved.
 * The source code for this program is not  published or otherwise divested of
 * its trade secrets, irrespective of what has been deposited with
 * the U.S. Copyright Office.
 ******************************************************************************/
package datasources

import (
	"context"

	"github.com/IBM/ibm-backup-recovery-sdk-go/backuprecoveryv1"
	"github.com/IBM/ibm-backup-recovery-sdk-go/migrationv2/common/types"
)

// DataSource represents a COMMON interface for ALL data sources
// This interface is implemented by Kubernetes, VPC VSI, and any other data source types
// All datasources (source and target) must implement this interface
// The BRS instance is passed to each method to ensure all operations use the same BRS instance
//
// BRS Backup/Restore Flow (Correct Order):
// 1. Create/Use BRS Instance (handled by BRSManager)
// 2. Create Connection on BRS (returns connection token)
// 3. Deploy Connector on data source using the connection token
// 4. Register data source with BRS
// 5. Create Protection Group with policy
// 6. Run Protection Group to take backup
// 7. Restore to same or different data source

// DATA SOURCE TYPES

// DataSourceType represents the type of data source

// DataSourceConfig holds configuration for creating a data source
type DataSourceConfig interface {
	GetDataSourceType() types.DataSourceType
	CreateDataSource(brsClient types.BRSClientWrapperInterface) (DataSource, error)
	Validate() error
}

type DataSource interface {
	// GetType returns the data source type (kubernetes, vpc_vsi, etc.)
	GetType() types.DataSourceType

	// GetName returns the data source name
	GetName() string

	// Validate validates the data source configuration
	Validate() error

	//===================================================================
	//
	// REGISTRATION OPERATIONS
	//
	//===================================================================
	// Step 4: Registration operations - registers the data source with BRS for protection
	RegisterSourceParams(ctx context.Context, connectionID string) (*backuprecoveryv1.RegisterProtectionSourceOptions, error)
	//===================================================================
	//
	// PROTECTION OPERATIONS
	//
	//===================================================================
	// Step 5: Protection operations - creates protection groups for the data source
	// Uses generic ProtectionGroupParams which contains datasource-specific params
	CreateProtectionGroup(ctx context.Context, registrationID int64, params *types.ProtectionGroupParams) (*backuprecoveryv1.CreateProtectionGroupOptions, error)

	//===================================================================
	//
	// RESTORE OPERATIONS
	//
	//===================================================================
	// Step 7: Restore operations - restores from backup to target data source
	RunRestore(ctx context.Context, groupId, backupId string, targetRegistrationID int64, restoreParams *types.RestoreParams) (*backuprecoveryv1.CreateRecoveryOptions, error)
}

// DataSourceFactory creates data source instances
type DataSourceFactory interface {
	// CreateDataSource creates a data source based on configuration
	CreateDataSource(config DataSourceConfig, brsClient types.BRSClientWrapperInterface) (DataSource, error)

	// SupportedTypes returns list of supported data source types
	SupportedTypes() []types.DataSourceType
}
