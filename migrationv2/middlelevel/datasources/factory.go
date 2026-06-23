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
	"fmt"

	"github.com/IBM/ibm-backup-recovery-sdk-go/migrationv2/common/types"
)

// DefaultDataSourceFactory is the default implementation of DataSourceFactory
type DefaultDataSourceFactory struct{}

// NewDataSourceFactory creates a new data source factory
func NewDataSourceFactory() DataSourceFactory {
	return &DefaultDataSourceFactory{}
}

// CreateDataSource creates a data source based on configuration
func (f *DefaultDataSourceFactory) CreateDataSource(config DataSourceConfig, brsClient types.BRSClientWrapperInterface) (DataSource, error) {
	if config == nil {
		return nil, fmt.Errorf("data source config cannot be nil")
	}

	if err := config.Validate(); err != nil {
		return nil, err
	}

	return config.CreateDataSource(brsClient)
}

// SupportedTypes returns list of supported data source types
func (f *DefaultDataSourceFactory) SupportedTypes() []types.DataSourceType {
	return []types.DataSourceType{
		types.DataSourceKubernetes,
		types.DataSourceVPCVSI,
	}
}
