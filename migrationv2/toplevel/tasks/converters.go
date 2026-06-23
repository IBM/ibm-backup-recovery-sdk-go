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
	"fmt"
	"time"

	"github.com/IBM/ibm-backup-recovery-sdk-go/backuprecoveryv1"
	"github.com/IBM/ibm-backup-recovery-sdk-go/migrationv2/common/types"
	"github.com/IBM/ibm-backup-recovery-sdk-go/migrationv2/middlelevel/connectors"
)

// buildConnectionResults converts DataSourceConnectionList into a slice of ConnectionResult
func buildConnectionResults(list *backuprecoveryv1.DataSourceConnectionList) []*types.ConnectionResult {
	if list == nil || len(list.Connections) == 0 {
		return []*types.ConnectionResult{}
	}

	results := make([]*types.ConnectionResult, len(list.Connections))

	for i, conn := range list.Connections {
		results[i] = &types.ConnectionResult{
			ConnectionName:    safeString(conn.ConnectionName),
			ConnectionID:      safeString(conn.ConnectionID),
			Type:              safeString(conn.ConnectionEnvType),
			RegistrationToken: safeString(conn.RegistrationToken),
			CreatedAt:         time.Now(),
		}
	}

	return results
}

// buildRegistrationResults converts SourceRegistrations into a slice of RegistrationResult
func buildRegistrationResults(src *backuprecoveryv1.SourceRegistrations) []*types.RegistrationResult {
	if src == nil || src.Registrations == nil || len(src.Registrations) == 0 {
		return []*types.RegistrationResult{}
	}

	results := make([]*types.RegistrationResult, 0, len(src.Registrations))

	for _, reg := range src.Registrations {
		results = append(results, buildSingleRegistrationResult(reg))
	}

	return results
}

// buildSingleRegistrationResult maps a single SourceRegistrationResponseParams to RegistrationResult
func buildSingleRegistrationResult(data backuprecoveryv1.SourceRegistrationResponseParams) *types.RegistrationResult {
	return &types.RegistrationResult{
		RegistrationID: safeInt64(data.ID),
		SourceName:     extractSourceName(data.SourceInfo),
		Status:         safeString(data.AuthenticationStatus),
		ConnectionID:   fmt.Sprintf("%d", *data.ConnectionID),
		CreatedAt:      time.Now(),
	}
}

// extractSourceName safely retrieves source name from SourceInfo
func extractSourceName(info *backuprecoveryv1.Object) string {
	if info != nil && info.SourceName != nil {
		return *info.SourceName
	}

	return ""
}

// safeString safely dereferences a string pointer
func safeString(ptr *string) string {
	if ptr != nil {
		return *ptr
	}

	return ""
}

// safeInt64 safely dereferences an int64 pointer
func safeInt64(ptr *int64) int64 {
	if ptr != nil {
		return *ptr
	}

	return 0
}

func convertConnectorResponse(connectorsList []backuprecoveryv1.DataSourceConnector) []*connectors.ConnectorResult {
	results := make([]*connectors.ConnectorResult, 0, len(connectorsList))

	for _, connector := range connectorsList {
		res := &connectors.ConnectorResult{}

		if connector.ConnectionID != nil {
			res.ConnectionID = *connector.ConnectionID
		}

		if connector.ConnectorID != nil {
			res.ConnectorID = *connector.ConnectorID
		}

		res.Status = getStatus(&connector)

		results = append(results, res)
	}

	return results
}

func getStatus(connector *backuprecoveryv1.DataSourceConnector) string {
	if connector == nil || connector.ConnectivityStatus == nil {
		return "Unknown"
	}

	if *connector.ConnectivityStatus.IsConnected {
		return "Healthy"
	}

	return "Unhealthy"
}
