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
	"fmt"

	"github.com/IBM/ibm-backup-recovery-sdk-go/migrationv2/common/errors"
)

// DefaultConnectorFactory implements ConnectorFactory
type DefaultConnectorFactory struct {
	supportedTypes map[ConnectorType]bool
}

// NewConnectorFactory creates a new connector factory
func NewConnectorFactory() ConnectorFactory {
	return &DefaultConnectorFactory{
		supportedTypes: map[ConnectorType]bool{
			ConnectorTypeHelm:      true,
			ConnectorTypeOperator:  false, // Not yet implemented
			ConnectorTypeAgent:     true,
			ConnectorTypeAgentless: false, // Not yet implemented
		},
	}
}

// CreateDeployer creates a connector deployer for the specified type
func (f *DefaultConnectorFactory) CreateConnectorDeployer(config ConnectorDeployConfig) (ConnectorDeployer, error) {
	// Check if type is supported
	connectorType := config.GetType()
	if supported, ok := f.supportedTypes[connectorType]; !ok || !supported {
		return nil, errors.NewInvalidConfigError(
			fmt.Sprintf("connector type '%s' is not supported or not yet implemented", connectorType),
			nil,
		)
	}
	return config.CreateConnectorDeployer()

}

// SupportedTypes returns list of supported connector types
func (f *DefaultConnectorFactory) SupportedTypes() []ConnectorType {
	types := make([]ConnectorType, 0)
	for connectorType, supported := range f.supportedTypes {
		if supported {
			types = append(types, connectorType)
		}
	}
	return types
}

// GetSupportedTypesForDataSource returns connector types supported for a specific datasource
func GetSupportedTypesForDataSource(dataSourceType string) []ConnectorType {
	switch dataSourceType {
	case "kubernetes":
		return []ConnectorType{
			ConnectorTypeHelm,
			// ConnectorTypeOperator,  // Not yet implemented
		}
	case "vpcvsi":
		return []ConnectorType{
			ConnectorTypeAgent,
			// ConnectorTypeAgentless, // Not yet implemented
		}
	default:
		return []ConnectorType{}
	}
}

// GetDefaultConnectorType returns the default connector type for a datasource
func GetDefaultConnectorType(dataSourceType string) ConnectorType {
	switch dataSourceType {
	case "kubernetes":
		return ConnectorTypeHelm
	case "vpcvsi":
		return ConnectorTypeAgent
	default:
		return ""
	}
}

// ValidateConnectorTypeForDataSource validates if a connector type is valid for a datasource
func ValidateConnectorTypeForDataSource(dataSourceType string, connectorType ConnectorType) error {
	supportedTypes := GetSupportedTypesForDataSource(dataSourceType)

	for _, supported := range supportedTypes {
		if supported == connectorType {
			return nil
		}
	}

	return errors.NewInvalidConfigError(
		fmt.Sprintf("connector type '%s' is not supported for datasource type '%s'", connectorType, dataSourceType),
		nil,
	)
}
