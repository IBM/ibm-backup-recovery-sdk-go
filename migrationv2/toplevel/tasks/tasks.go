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
	"net/http"
	"time"

	"github.com/IBM/go-sdk-core/v5/core"
	"github.com/IBM/ibm-backup-recovery-sdk-go/backuprecoveryv1"
	activitytracker "github.com/IBM/ibm-backup-recovery-sdk-go/migrationv2/common/activity-tracker"
	"github.com/IBM/ibm-backup-recovery-sdk-go/migrationv2/common/errors"
	"github.com/IBM/ibm-backup-recovery-sdk-go/migrationv2/common/helpers"
	"github.com/IBM/ibm-backup-recovery-sdk-go/migrationv2/common/logger"
	"github.com/IBM/ibm-backup-recovery-sdk-go/migrationv2/common/metrics"
	"github.com/IBM/ibm-backup-recovery-sdk-go/migrationv2/common/types"
	"github.com/IBM/ibm-backup-recovery-sdk-go/migrationv2/middlelevel/connectors"
	"github.com/IBM/ibm-backup-recovery-sdk-go/migrationv2/middlelevel/datasources"
)

// DefaultTaskAPI is the default implementation of TaskAPI
type DefaultTaskAPI struct {
	brsClient           types.BRSClientWrapperInterface
	config              *types.SDKConfig
	queryHelper         *helpers.BRSQueryHelper
	activityTrackerSink activitytracker.Sink
	logger              logger.Logger
	metrics             metrics.Metrics
	accountID           string
}

// NewTaskAPI creates a new task API instance
func NewTaskAPI(brsClient types.BRSClientWrapperInterface, config *types.SDKConfig) TaskAPI {
	// ActivityTrackerSink is always non-nil after ToSDKConfig (sets NoOpSink when disabled).
	// Guard against nil for tests that build SDKConfig directly.
	var sink activitytracker.Sink = activitytracker.NoOpSink{}
	if config != nil && config.ActivityTrackerSink != nil {
		sink = config.ActivityTrackerSink
	}

	// Get logger from config or create a no-op logger
	var log logger.Logger
	if config != nil && config.Logger != nil {
		log = config.Logger
	} else {
		logConfig := logger.DefaultConfig()
		logConfig.ServiceName = "brs-migration-sdk"
		logConfig.Environment = "production"
		log = logger.New(logConfig)
	}

	// Metrics is always non-nil after ToSDKConfig (sets NoopMetrics when disabled).
	// Guard against nil for tests that build SDKConfig directly.
	metricsCollector := metrics.Metrics(metrics.NewNoop())
	if config != nil && config.Metrics != nil {
		metricsCollector = config.Metrics
	}

	var accountID string
	if config != nil {
		accountID = config.AccountID
	}

	return &DefaultTaskAPI{
		brsClient:           brsClient,
		config:              config,
		queryHelper:         helpers.NewBRSQueryHelper(brsClient),
		activityTrackerSink: sink,
		logger:              log,
		metrics:             metricsCollector,
		accountID:           accountID,
	}
}

// getAccountID returns the account ID, falling back to "unknown" for metrics labels.
// AccountID is guaranteed non-empty when metrics or activity tracking is active,
// because Config.Validate() enforces it before a Client can be constructed.
func (t *DefaultTaskAPI) getAccountID(ctx context.Context) string {
	if t.accountID == "" {
		return "unknown"
	}
	return t.accountID
}

// getSourceIP gets the source IP from context.
// Returns empty string if not set.
func (t *DefaultTaskAPI) getSourceIP(ctx context.Context) string {
	if ctx == nil {
		return ""
	}
	if sourceIP, ok := ctx.Value("source_ip").(string); ok {
		return sourceIP
	}
	return ""
}

func (t *DefaultTaskAPI) emitActivityEvent(ctx context.Context, event activitytracker.Event) {
	// activityTrackerSink is always at least NoOpSink{}; Emit is a no-op when AT is disabled.
	_ = t.activityTrackerSink.Emit(ctx, event)
}

// recordOperationSuccess records metrics for successful operations
// This helper reduces code duplication across all Task API functions
func (t *DefaultTaskAPI) recordOperationSuccess(ctx context.Context, operation string, startedAt time.Time) {
	t.metrics.IncCounter(ctx, metrics.MetricOperationTotal,
		metrics.Label{Key: "accountId", Value: t.getAccountID(ctx)},
		metrics.Label{Key: "operation", Value: operation},
		metrics.Label{Key: "status", Value: "success"},
	)

	t.metrics.RecordDuration(ctx, metrics.MetricOperationDuration, time.Since(startedAt),
		metrics.Label{Key: "accountId", Value: t.getAccountID(ctx)},
		metrics.Label{Key: "operation", Value: operation},
	)
}

// recordOperationFailure records metrics for failed operations
// This helper reduces code duplication across all Task API functions
func (t *DefaultTaskAPI) recordOperationFailure(ctx context.Context, operation string, startedAt time.Time) {
	t.metrics.IncCounter(ctx, metrics.MetricOperationTotal,
		metrics.Label{Key: "accountId", Value: t.getAccountID(ctx)},
		metrics.Label{Key: "operation", Value: operation},
		metrics.Label{Key: "status", Value: "failure"},
	)

	t.metrics.RecordDuration(ctx, metrics.MetricOperationDuration, time.Since(startedAt),
		metrics.Label{Key: "accountId", Value: t.getAccountID(ctx)},
		metrics.Label{Key: "operation", Value: operation},
	)
}

// CreateConnection creates a new connection (idempotent)
// Returns existing connection if found by name, creates new if not found
// Microservice-safe: Handles race conditions gracefully
func (t *DefaultTaskAPI) CreateConnection(ctx context.Context, connectionParams *types.ConnectionParams) (*types.ConnectionResult, *errors.SDKError) {
	startedAt := time.Now()

	// Validate parameters
	if connectionParams == nil || connectionParams.Name == "" {
		t.logger.Error(ctx, "Connection creation failed: invalid parameters", "error", "connectionParams.Name is required", "operation", types.OpCreateConnection)
		sdkErr := errors.NewSDKError(errors.ErrCodeInvalidInput, "connectionParams.Name is required", nil)
		t.emitActivityEvent(ctx, activitytracker.BuildConnectionActivityEvent("", "", activitytracker.EventStateFailed, "connection validation failed", sdkErr, 1, startedAt, t.getSourceIP(ctx)))
		t.logger.Error(ctx, "Connection creation failed: invalid parameters", "error", "connectionParams.Name is required")
		t.recordOperationFailure(ctx, metrics.OperationCreateConnection, startedAt)

		return nil, sdkErr
	}

	t.logger.Info(ctx, "Creating connection", "connectionName", connectionParams.Name, "operation", types.OpCreateConnection)

	// Step 1: Check if connection already exists (idempotent check)
	t.logger.Debug(ctx, "Checking if connection already exists", "connectionName", connectionParams.Name, "operation", types.OpCreateConnection)
	t.emitActivityEvent(ctx, activitytracker.BuildConnectionActivityEvent(connectionParams.Name, "", activitytracker.EventStateStarted, "starting connection creation", nil, 1, startedAt, t.getSourceIP(ctx)))

	existing, err := t.GetConnectionByName(ctx, connectionParams.Name)
	if err != nil {
		if !errors.IsNotFoundError(err) {
			t.logger.Error(ctx, "Failed to check existing connection", logger.Err(err), "connectionName", connectionParams.Name, "operation", types.OpCreateConnection)
			t.emitActivityEvent(ctx, activitytracker.BuildConnectionActivityEvent(connectionParams.Name, "", activitytracker.EventStateFailed, "failed to lookup existing connection", err, 1, startedAt, t.getSourceIP(ctx)))
			return nil, err
		}

		t.logger.Debug(ctx, "Connection not found, proceeding with creation", "connectionName", connectionParams.Name, "operation", types.OpCreateConnection)
	}

	if existing != nil {
		// Connection already exists - check if we need to generate registration token
		if existing.RegistrationToken == "" && existing.ConnectionID != "" {
			// Generate registration token for existing connection
			token, _, apiErr := t.brsClient.GetBRSClient().GenerateDataSourceConnectionRegistrationToken(
				&backuprecoveryv1.GenerateDataSourceConnectionRegistrationTokenOptions{
					XIBMTenantID: core.StringPtr(t.brsClient.GetTenantId()),
					ConnectionID: core.StringPtr(existing.ConnectionID),
				})

			if apiErr != nil {
				return nil, errors.NewConnectionFailedError("Failed to generate registration token for existing connection", apiErr)
			}

			if token != nil {
				existing.RegistrationToken = *token
			}
		}
		// Connection already exists - return it (idempotent behavior)
		t.logger.Info(ctx, "Connection already exists, returning existing connection",
			"connectionName", connectionParams.Name,
			"connectionID", existing.ConnectionID,
			"idempotent", true,
			"operation", types.OpCreateConnection)
		t.emitActivityEvent(ctx, activitytracker.BuildConnectionActivityEvent(existing.ConnectionName, existing.ConnectionID, activitytracker.EventStateSucceeded, "connection already exists; reusing existing connection", nil, 1, startedAt, t.getSourceIP(ctx)))
		return existing, nil
	}

	// Convert user-friendly connection type to BRS internal type
	brsType := connectionParams.Type.ToBRSConnectionType()

	var connEnvType *string
	if connectionParams.Type != types.ConnectionType_VSI {
		connEnvType = &brsType
	}

	params := &backuprecoveryv1.CreateDataSourceConnectionOptions{
		ConnectionName:    &connectionParams.Name,
		ConnectionEnvType: connEnvType,
		XIBMTenantID:      core.StringPtr(t.brsClient.GetTenantId()),
	}

	t.logger.Debug(ctx, "Calling BRS API to create connection", "connectionName", connectionParams.Name, "userType", connectionParams.Type, "brsType", brsType, "operation", types.OpCreateConnection)

	// Call BRS API to create connection
	connectionResult, _, apiErr := t.brsClient.GetBRSClient().CreateDataSourceConnectionWithContext(ctx, params)
	if apiErr != nil {
		if errors.IsStatusCode(apiErr, http.StatusConflict) {
			t.logger.Warn(ctx, "Connection creation conflict, fetching existing connection", "connectionName", connectionParams.Name, "operation", types.OpCreateConnection)
			existingConnection, sdkErr := t.GetConnectionByName(ctx, connectionParams.Name)
			if sdkErr != nil {
				t.emitActivityEvent(ctx, activitytracker.BuildConnectionActivityEvent(connectionParams.Name, "", activitytracker.EventStateFailed, "connection create conflict occurred and existing connection lookup failed", sdkErr, 1, startedAt, t.getSourceIP(ctx)))
				return nil, sdkErr
			}

			t.emitActivityEvent(ctx, activitytracker.BuildConnectionActivityEvent(existingConnection.ConnectionName, existingConnection.ConnectionID, activitytracker.EventStateSucceeded, "connection already exists after conflict; reusing existing connection", nil, 1, startedAt, t.getSourceIP(ctx)))
			return existingConnection, nil
		}

		t.logger.Error(ctx, "Failed to create connection via BRS API", logger.Err(apiErr), "connectionName", connectionParams.Name, "operation", types.OpCreateConnection)
		sdkErr := errors.NewConnectionFailedError("Failed to create connection", apiErr)
		t.emitActivityEvent(ctx, activitytracker.BuildConnectionActivityEvent(connectionParams.Name, "", activitytracker.EventStateFailed, "connection creation request failed", sdkErr, 1, startedAt, t.getSourceIP(ctx)))
		t.recordOperationFailure(ctx, metrics.OperationCreateConnection, startedAt)

		return nil, sdkErr
	}

	result := &types.ConnectionResult{
		ConnectionName: *connectionResult.ConnectionName,
		Type:           *connectionResult.ConnectionEnvType,
		CreatedAt:      time.Now(),
	}

	if connectionResult.ConnectionID != nil {
		result.ConnectionID = *connectionResult.ConnectionID
	}

	if connectionResult.RegistrationToken != nil {
		result.RegistrationToken = *connectionResult.RegistrationToken
	}

	t.emitActivityEvent(ctx, activitytracker.BuildConnectionActivityEvent(result.ConnectionName, result.ConnectionID, activitytracker.EventStateSucceeded, "connection created successfully", nil, 1, startedAt, t.getSourceIP(ctx)))
	t.recordOperationSuccess(ctx, metrics.OperationCreateConnection, startedAt)

	t.logger.Info(ctx, "Connection created successfully",
		"connectionName", result.ConnectionName,
		"connectionID", result.ConnectionID,
		"type", result.Type,
		"operation", types.OpCreateConnection)

	return result, nil
}

// GetConnection retrieves connection by ID
func (t *DefaultTaskAPI) GetConnection(ctx context.Context, connectionID string) (*types.ConnectionResult, *errors.SDKError) {
	start := time.Now()
	t.logger.Debug(ctx, "Retrieving connection by ID", "connectionID", connectionID, "operation", types.OpGetConnection)

	connectionData, _, apiErr := t.brsClient.GetBRSClient().GetDataSourceConnectionsWithContext(ctx, &backuprecoveryv1.GetDataSourceConnectionsOptions{
		XIBMTenantID:  core.StringPtr(t.brsClient.GetTenantId()),
		ConnectionIds: []string{connectionID},
	})

	if apiErr != nil {
		t.logger.Error(ctx, "Failed to retrieve connection", logger.Err(apiErr), "connectionID", connectionID, "operation", types.OpGetConnection)
		t.recordOperationFailure(ctx, metrics.OperationGetConnection, start)
		return nil, errors.NewConnectionFailedError("Failed to get connection details", apiErr)
	}

	connectionsResult := buildConnectionResults(connectionData)
	if len(connectionsResult) == 0 {
		t.logger.Warn(ctx, "Connection not found", "connectionID", connectionID, "operation", types.OpGetConnection)
		t.recordOperationFailure(ctx, metrics.OperationGetConnection, start)
		return nil, errors.NewSDKError(errors.ErrCodeConnectionNotFound, fmt.Sprintf("Connection details not found with id: %v", connectionID), nil)
	}

	t.logger.Info(ctx, "Connection retrieved successfully", "connectionID", connectionID, "connectionName", connectionsResult[0].ConnectionName, "operation", types.OpGetConnection)

	t.recordOperationSuccess(ctx, metrics.OperationGetConnection, start)

	return connectionsResult[0], nil
}

// GetConnectionByName retrieves connection by name
func (t *DefaultTaskAPI) GetConnectionByName(ctx context.Context, connectionName string) (*types.ConnectionResult, *errors.SDKError) {
	start := time.Now()
	t.logger.Info(ctx, "Retrieving connection by name", "connectionName", connectionName, "operation", types.OpGetConnectionByName)

	listOptions := &backuprecoveryv1.GetDataSourceConnectionsOptions{
		ConnectionNames: []string{connectionName},
		XIBMTenantID:    core.StringPtr(t.brsClient.GetTenantId()),
	}

	connections, _, apiErr := t.brsClient.GetBRSClient().GetDataSourceConnectionsWithContext(ctx, listOptions)
	if apiErr != nil {
		t.logger.Error(ctx, "Failed to retrieve connection by name",
			logger.Err(apiErr),
			"connectionName", connectionName,
			"operation", types.OpGetConnectionByName)
		t.recordOperationFailure(ctx, metrics.OperationGetConnection, start)
		return nil, errors.NewConnectionFailedError("failed to list connections", apiErr)
	}

	if connections == nil || len(connections.Connections) == 0 {
		t.logger.Warn(ctx, "Connection not found by name",
			"connectionName", connectionName,
			"operation", types.OpGetConnectionByName)
		t.recordOperationFailure(ctx, metrics.OperationGetConnection, start)
		return nil, errors.NewSDKError(errors.ErrCodeConnectionNotFound, fmt.Sprintf("Connection details not found with name: %v", connectionName), nil)
	}

	existingConn := &connections.Connections[0]

	// Convert to ConnectionResult
	result := &types.ConnectionResult{
		ConnectionName: *existingConn.ConnectionName,
		CreatedAt:      time.Now(),
	}

	if existingConn.ConnectionID != nil {
		result.ConnectionID = *existingConn.ConnectionID
	}

	if existingConn.RegistrationToken != nil {
		result.RegistrationToken = *existingConn.RegistrationToken
	}

	t.recordOperationSuccess(ctx, metrics.OperationGetConnection, start)
	t.logger.Info(ctx, "Connection retrieved successfully by name",
		"connectionName", connectionName,
		"connectionID", result.ConnectionID,
		"operation", types.OpGetConnectionByName)
	return result, nil
}

// ListConnections lists all connections
func (t *DefaultTaskAPI) ListConnections(ctx context.Context) ([]*types.ConnectionResult, *errors.SDKError) {
	start := time.Now()
	t.logger.Debug(ctx, "Listing all connections", "operation", types.OpListConnections)

	connectionData, _, apiErr := t.brsClient.GetBRSClient().GetDataSourceConnectionsWithContext(ctx, &backuprecoveryv1.GetDataSourceConnectionsOptions{
		XIBMTenantID: core.StringPtr(t.brsClient.GetTenantId()),
	})

	if apiErr != nil {
		t.logger.Error(ctx, "Failed to list connections", logger.Err(apiErr), "operation", types.OpListConnections)
		t.recordOperationFailure(ctx, metrics.OperationListConnections, start)
		return nil, errors.NewConnectionFailedError("Failed to get connection details", apiErr)
	}

	results := buildConnectionResults(connectionData)
	t.logger.Info(ctx, "Connections listed successfully", "count", len(results), "operation", types.OpListConnections)

	// Record success
	t.recordOperationSuccess(ctx, metrics.OperationListConnections, start)

	return results, nil
}

// DeleteConnection deletes a connection
func (t *DefaultTaskAPI) DeleteConnection(ctx context.Context, connectionID string) *errors.SDKError {
	start := time.Now()
	t.logger.Info(ctx, "Deleting connection", "connectionID", connectionID, "operation", types.OpDeleteConnection)

	startedAt := time.Now()
	t.emitActivityEvent(ctx, activitytracker.BuildDeleteConnectionActivityEvent(connectionID, activitytracker.EventStateStarted, "starting connection deletion", nil, 1, startedAt, t.getSourceIP(ctx)))

	_, apiErr := t.brsClient.GetBRSClient().DeleteDataSourceConnectionWithContext(ctx, &backuprecoveryv1.DeleteDataSourceConnectionOptions{
		XIBMTenantID: core.StringPtr(t.brsClient.GetTenantId()),
		ConnectionID: &connectionID,
	})

	if apiErr != nil {
		t.logger.Error(ctx, "Failed to delete connection", logger.Err(apiErr), "connectionID", connectionID, "operation", types.OpDeleteConnection)
		sdkErr := errors.NewSDKError(errors.ErrCodeConnectionFailed, "Failed to delete connection details", apiErr)
		t.emitActivityEvent(ctx, activitytracker.BuildDeleteConnectionActivityEvent(connectionID, activitytracker.EventStateFailed, "connection deletion failed", sdkErr, 1, startedAt, t.getSourceIP(ctx)))
		t.recordOperationFailure(ctx, metrics.OperationDeleteConnection, start)

		return sdkErr
	}

	t.logger.Info(ctx, "Connection deleted successfully", "connectionID", connectionID, "operation", types.OpDeleteConnection)
	t.emitActivityEvent(ctx, activitytracker.BuildDeleteConnectionActivityEvent(connectionID, activitytracker.EventStateSucceeded, "connection deleted successfully", nil, 1, startedAt, t.getSourceIP(ctx)))

	// Record success
	t.recordOperationSuccess(ctx, metrics.OperationDeleteConnection, start)

	return nil
}

// DeployConnector deploys connector (idempotent)
// Returns existing connector if found for connection, deploys new if not found
// Microservice-safe: Handles race conditions gracefully
func (t *DefaultTaskAPI) DeployConnector(ctx context.Context, connector connectors.ConnectorDeployer, connectionResult *types.ConnectionResult) (*connectors.ConnectorResult, *errors.SDKError) {
	startedAt := time.Now()

	// Validate parameters
	if connectionResult == nil || connectionResult.ConnectionID == "" {
		t.logger.Error(ctx, "Connector deployment failed: invalid parameters", "error", "connectionID is required", "operation", types.OpDeployConnector)
		t.recordOperationFailure(ctx, metrics.OperationDeployConnector, startedAt)
		sdkErr := errors.NewSDKError(errors.ErrCodeInvalidInput, "connectionID is required", nil)
		t.emitActivityEvent(ctx, activitytracker.BuildConnectorActivityEvent("", "", activitytracker.EventStateFailed, "connector deployment validation failed", sdkErr, 1, startedAt, t.getSourceIP(ctx)))
		return nil, sdkErr
	}

	t.logger.Info(ctx, "Deploying connector",
		"connectionID", connectionResult.ConnectionID,
		"connectionName", connectionResult.ConnectionName,
		"connectorType", connector.GetType(),
		types.OpDeployConnector)

	t.emitActivityEvent(ctx, activitytracker.BuildConnectorActivityEvent(connectionResult.ConnectionID, "", activitytracker.EventStateStarted, "starting connector deployment", nil, 1, startedAt, t.getSourceIP(ctx)))

	t.logger.Debug(ctx, "Checking if connector already exists", "connectionID", connectionResult.ConnectionID, "operation", types.OpDeployConnector)
	existing, err := t.GetConnectorByConnection(ctx, connectionResult.ConnectionID)
	if err != nil && !errors.ErrorCodeMatch(err, errors.ErrCodeConnectorNotFound) {
		t.emitActivityEvent(ctx, activitytracker.BuildConnectorActivityEvent(connectionResult.ConnectionID, "", activitytracker.EventStateFailed, "failed to lookup existing connector", err, 1, startedAt, t.getSourceIP(ctx)))
		t.logger.Error(ctx, "Failed to check existing connector", logger.Err(err), "connectionID", connectionResult.ConnectionID, "operation", types.OpDeployConnector)
		t.recordOperationFailure(ctx, metrics.OperationDeployConnector, startedAt)
		return nil, err
	}

	if existing != nil {
		t.emitActivityEvent(ctx, activitytracker.BuildConnectorActivityEvent(connectionResult.ConnectionID, existing[0].ConnectorID, activitytracker.EventStateSucceeded, "connector already exists; reusing existing connector", nil, 1, startedAt, t.getSourceIP(ctx)))
		// Connector already exists - return it (idempotent behavior)
		t.logger.Info(ctx, "Connector already exists, returning existing connector",
			"connectionID", connectionResult.ConnectionID,
			"connectorID", existing[0].ConnectorID,
			"idempotent", true, "operation", types.OpDeployConnector)

		t.recordOperationSuccess(ctx, metrics.OperationDeployConnector, startedAt)
		return existing[0], nil
	}

	t.logger.Debug(ctx, "Deploying new connector", "connectionID", connectionResult.ConnectionID, "connectorType", connector.GetType(), "operation", types.OpDeployConnector)

	// Inject BRS client + platform type so the connector can resolve live chart
	// metadata from the /v2/data-source-connectors/metadata API before installing.
	// PlatformType is taken directly from ConnectionResult.Type (the raw BRS
	// k8sPlatformType string, e.g. "kRoksVpc") which is always set by CreateConnection.
	connector.SetDeployContext(connectors.ConnectorDeployContext{
		BRSClient:    t.brsClient,
		PlatformType: connectionResult.Type,
	})

	connectorResult, apiErr := connector.Deploy(ctx, connectionResult.RegistrationToken)
	if apiErr != nil {
		t.logger.Error(ctx, "Failed to deploy connector", logger.Err(apiErr), "connectionID", connectionResult.ConnectionID, "connectorType", connector.GetType(), "operation", types.OpDeployConnector)
		sdkErr := errors.NewConnectorFailedError("Failed to deploy connector", apiErr)
		t.emitActivityEvent(ctx, activitytracker.BuildConnectorActivityEvent(connectionResult.ConnectionID, "", activitytracker.EventStateFailed, "connector deployment failed", sdkErr, 1, startedAt, t.getSourceIP(ctx)))
		t.recordOperationFailure(ctx, metrics.OperationDeployConnector, startedAt)

		return nil, sdkErr
	}

	connectorResult.ConnectionID = connectionResult.ConnectionID
	t.emitActivityEvent(ctx, activitytracker.BuildConnectorActivityEvent(connectionResult.ConnectionID, connectorResult.ConnectorID, activitytracker.EventStateSucceeded, "connector deployed successfully", nil, 1, startedAt, t.getSourceIP(ctx)))

	t.logger.Info(ctx, "Connector deployed successfully",
		"connectionID", connectionResult.ConnectionID,
		"connectorID", connectorResult.ConnectorID,
		"status", connectorResult.Status)

	t.recordOperationSuccess(ctx, metrics.OperationDeployConnector, startedAt)

	return connectorResult, nil
}

// GetConnector retrieves connector by ID
func (t *DefaultTaskAPI) GetConnector(ctx context.Context, connectorID string) (*connectors.ConnectorResult, *errors.SDKError) {
	start := time.Now()
	t.logger.Debug(ctx, "Retrieving connector by ID", "connectorID", connectorID, "operation", types.OpGetConnector)

	result, _, err := t.brsClient.GetBRSClient().GetDataSourceConnectorsWithContext(ctx, &backuprecoveryv1.GetDataSourceConnectorsOptions{
		XIBMTenantID: core.StringPtr(t.brsClient.GetTenantId()),
		ConnectorIds: []string{connectorID},
	})
	if err != nil {
		t.logger.Error(ctx, "Failed to retrieve connector", logger.Err(err), "connectorID", connectorID, "operation", types.OpGetConnector)
		t.recordOperationFailure(ctx, metrics.OperationGetConnector, start)
		return nil, errors.NewConnectorFailedError("failed to get connector by connectionID", err)
	}

	if result == nil || len(result.Connectors) == 0 {
		t.logger.Warn(ctx, "Connector not found", "connectorID", connectorID, "operation", types.OpGetConnector)
		t.recordOperationFailure(ctx, metrics.OperationGetConnector, start)
		return nil, errors.NewConnectorNotFound(fmt.Sprintf("No Connectors found for connectorID: %v", connectorID), nil)
	}

	connector := convertConnectorResponse(result.Connectors)

	t.logger.Info(ctx, "Connector retrieved successfully", "connectorID", connectorID, "operation", types.OpGetConnector)

	// Record success
	t.recordOperationSuccess(ctx, metrics.OperationGetConnector, start)

	return connector[0], nil
}

// GetConnectorByConnection retrieves connector for a connection
func (t *DefaultTaskAPI) GetConnectorByConnection(ctx context.Context, connectionID string) ([]*connectors.ConnectorResult, *errors.SDKError) {
	t.logger.Debug(ctx, "Retrieving connector by connection", "connectionID", connectionID, "operation", types.OpGetConnectorByConnection)

	result, _, apiErr := t.brsClient.GetBRSClient().GetDataSourceConnectorsWithContext(ctx, &backuprecoveryv1.GetDataSourceConnectorsOptions{
		XIBMTenantID: core.StringPtr(t.brsClient.GetTenantId()),
		ConnectionID: core.StringPtr(connectionID),
	})
	if apiErr != nil {
		t.logger.Debug(ctx, "Connector not found for connection", "connectionID", connectionID, "error", apiErr.Error(), "operation", types.OpGetConnectorByConnection)
		return nil, errors.NewConnectorFailedError("failed to get connector by connectionID", apiErr)
	}

	if result == nil || len(result.Connectors) == 0 {
		return nil, errors.NewConnectorNotFound(fmt.Sprintf("No Connectors found for connectionID: %v", connectionID), nil)
	}

	results := convertConnectorResponse(result.Connectors)
	t.logger.Debug(ctx, "Connectors found for connection", "connectionID", connectionID, "count", len(results), "operation", types.OpGetConnectorByConnection)
	return results, nil
}

// DeleteConnector removes connector from both BRS and the actual deployment (e.g., Kubernetes namespace)
// If connector deployer is provided, it will also clean up the deployed resources
func (t *DefaultTaskAPI) DeleteConnector(ctx context.Context, connectorID string, connector ...connectors.ConnectorDeployer) *errors.SDKError {
	startedAt := time.Now()
	t.emitActivityEvent(ctx, activitytracker.BuildDeleteConnectorActivityEvent(connectorID, activitytracker.EventStateStarted, "starting connector deletion", nil, 1, startedAt, t.getSourceIP(ctx)))
	t.logger.Info(ctx, "Deleting connector", "connectorID", connectorID, "operation", types.OpDeleteConnector)

	// Step 1: Delete deployed resources if connector deployer is provided
	if len(connector) > 0 && connector[0] != nil {
		t.logger.Debug(ctx, "Deleting connector deployed resources", "connectorID", connectorID, "connectorType", connector[0].GetType())
		if err := connector[0].Delete(ctx, connectorID); err != nil {
			t.logger.Error(ctx, "Failed to delete connector deployed resources", logger.Err(err), "connectorID", connectorID, "operation", types.OpDeleteConnector)
			// Continue to delete from BRS even if deployment cleanup fails
			t.logger.Warn(ctx, "Continuing with BRS deletion despite deployment cleanup failure", "connectorID", connectorID)
		} else {
			t.logger.Info(ctx, "Connector deployed resources deleted successfully", "connectorID", connectorID)
		}
	}

	// Step 2: Delete connector record from BRS
	_, err := t.brsClient.GetBRSClient().DeleteDataSourceConnectorWithContext(ctx, &backuprecoveryv1.DeleteDataSourceConnectorOptions{
		XIBMTenantID: core.StringPtr(t.brsClient.GetTenantId()),
		ConnectorID:  core.StringPtr(connectorID),
	})
	if err != nil {
		t.logger.Error(ctx, "Failed to delete connector from BRS", logger.Err(err), "connectorID", connectorID, "operation", types.OpDeleteConnector)
		t.recordOperationFailure(ctx, metrics.OperationDeleteConnector, startedAt)

		sdkErr := errors.NewConnectorFailedError("failed to delete connector by connectorId", err)
		t.emitActivityEvent(ctx, activitytracker.BuildDeleteConnectorActivityEvent(connectorID, activitytracker.EventStateFailed, "connector deletion failed", sdkErr, 1, startedAt, t.getSourceIP(ctx)))
		return sdkErr
	}

	t.emitActivityEvent(ctx, activitytracker.BuildDeleteConnectorActivityEvent(connectorID, activitytracker.EventStateSucceeded, "connector deleted successfully", nil, 1, startedAt, t.getSourceIP(ctx)))
	t.logger.Info(ctx, "Connector deleted successfully", "connectorID", connectorID, "operation", types.OpDeleteConnector)

	// Record success
	t.recordOperationSuccess(ctx, metrics.OperationDeleteConnector, startedAt)

	return nil
}

// RegisterSource registers a source (idempotent)
// Returns existing registration if found for connection, creates new if not found
// Microservice-safe: Handles race conditions gracefully
func (t *DefaultTaskAPI) RegisterSource(ctx context.Context, dataSource datasources.DataSource, connectionID string) (*types.RegistrationResult, *errors.SDKError) {
	start := time.Now()
	t.logger.Info(ctx, "Registering source", "connectionID", connectionID, "operation", types.OpRegisterSource)

	// Validate parameters
	if connectionID == "" {
		t.logger.Error(ctx, "Source registration failed: invalid parameters", "error", "connectionID is required", "operation", types.OpRegisterSource)
		t.recordOperationFailure(ctx, metrics.OperationRegisterSource, start)
		return nil, errors.NewSDKError(errors.ErrCodeInvalidInput, "connectionID is required", nil)
	}

	// Step 1: Check if registration already exists for this connection (idempotent check)
	t.logger.Debug(ctx, "Checking if registration already exists", "connectionID", connectionID, "operation", types.OpRegisterSource)
	existing, err := t.GetRegistrationByConnection(ctx, connectionID)
	if err != nil {
		t.logger.Error(ctx, "Failed to check existing registration", logger.Err(err), "connectionID", connectionID)
		t.recordOperationFailure(ctx, metrics.OperationRegisterSource, start)

		return nil, err
	}

	if existing != nil {
		// Registration already exists - return it (idempotent behavior)
		t.logger.Info(ctx, "Registration already exists, returning existing registration",
			"connectionID", connectionID,
			"registrationID", existing.RegistrationID,
			"idempotent", true, "operation", types.OpRegisterSource)
		t.recordOperationSuccess(ctx, metrics.OperationRegisterSource, start)
		return existing, nil
	}

	params, paramErr := dataSource.RegisterSourceParams(ctx, connectionID)
	if paramErr != nil {
		t.logger.Error(ctx, "Failed to build registration params", logger.Err(paramErr), "connectionID", connectionID, "operation", types.OpRegisterSource)
		t.recordOperationFailure(ctx, metrics.OperationRegisterSource, start)
		return nil, errors.NewSDKError(errors.ErrCodeRegistrationFailed, "failed to build registration params", paramErr)
	}

	params.XIBMTenantID = core.StringPtr(t.brsClient.GetTenantId())

	t.logger.Debug(ctx, "Calling BRS API to register source", "connectionID", connectionID, "operation", types.OpRegisterSource)
	registrationResultResult, response, sdkErr := t.brsClient.GetBRSClient().RegisterProtectionSourceWithContext(ctx, params)
	if sdkErr != nil {
		t.logger.Error(ctx, "Failed to register source via BRS API", logger.Err(sdkErr), "connectionID", connectionID, "operation", types.OpRegisterSource)
		t.recordOperationFailure(ctx, metrics.OperationRegisterSource, start)
		return nil, errors.NewRegistrationFailedError(
			fmt.Sprintf("failed to create registration: %v", sdkErr),
			sdkErr,
		).WithDetails("response", response)
	}

	result := &types.RegistrationResult{
		RegistrationID: *registrationResultResult.ID,
		ConnectionID:   connectionID,
		SourceName:     dataSource.GetName(),
		CreatedAt:      time.Now(),
	}

	t.logger.Info(ctx, "Source registered successfully",
		"connectionID", connectionID,
		"registrationID", result.RegistrationID,
		"sourceName", result.SourceName, "operation", types.OpRegisterSource)

	// Record success
	t.recordOperationSuccess(ctx, metrics.OperationRegisterSource, start)

	return result, nil
}

// GetRegistration retrieves registration by ID
func (t *DefaultTaskAPI) GetRegistration(ctx context.Context, registrationID int64) (*types.RegistrationResult, *errors.SDKError) {
	t.logger.Debug(ctx, "Retrieving registration by ID", "registrationID", registrationID, "operation", types.OpGetRegistration)
	start := time.Now()

	data, _, err := t.brsClient.GetBRSClient().GetSourceRegistrationsWithContext(ctx, &backuprecoveryv1.GetSourceRegistrationsOptions{
		XIBMTenantID: core.StringPtr(t.brsClient.GetTenantId()),
		Ids:          []int64{registrationID},
	})
	if err != nil {
		t.logger.Error(ctx, "Failed to retrieve registration", logger.Err(err), "registrationID", registrationID, "operation", types.OpGetRegistration)
		t.recordOperationFailure(ctx, metrics.OperationGetRegistration, start)
		return nil, errors.NewRegistrationFailedError("GetRegistration API call error", err)
	}

	registration := buildRegistrationResults(data)
	if registration == nil || len(registration) == 0 {
		t.logger.Warn(ctx, "Registration not found", "registrationID", registrationID, "operation", types.OpGetRegistration)
		t.recordOperationSuccess(ctx, metrics.OperationGetRegistration, start)
		return nil, nil
	}

	t.logger.Info(ctx, "Registration retrieved successfully", "registrationID", registrationID, "operation", types.OpGetRegistration)

	// Record success
	t.recordOperationSuccess(ctx, metrics.OperationGetRegistration, start)

	return registration[0], nil
}

// GetRegistrationByConnection retrieves registration for a connection
func (t *DefaultTaskAPI) GetRegistrationByConnection(ctx context.Context, connectionID string) (*types.RegistrationResult, *errors.SDKError) {
	t.logger.Debug(ctx, "Retrieving registration by connection", "connectionID", connectionID, "operation", types.OpGetRegistrationByConnection)

	// Inline query logic to make testing easier
	listOptions := &backuprecoveryv1.GetSourceRegistrationsOptions{
		XIBMTenantID: core.StringPtr(t.brsClient.GetTenantId()),
	}

	registrations, _, apiErr := t.brsClient.GetBRSClient().GetSourceRegistrationsWithContext(ctx, listOptions)
	if apiErr != nil {
		t.logger.Error(ctx, "Failed to list registrations", logger.Err(apiErr), "connectionID", connectionID, "operation", types.OpGetRegistrationByConnection)
		return nil, errors.NewRegistrationFailedError("failed to list registrations", apiErr)
	}

	// Find registration by connection ID
	var existingReg *backuprecoveryv1.SourceRegistrationResponseParams
	if registrations != nil && registrations.Registrations != nil {
		for _, reg := range registrations.Registrations {
			if reg.ConnectionID != nil && fmt.Sprintf("%v", *reg.ConnectionID) == connectionID {
				existingReg = &reg
				break
			}
		}
	}

	if existingReg == nil {
		t.logger.Debug(ctx, "Registration not found for connection", "connectionID", connectionID, "operation", types.OpGetRegistrationByConnection)
		return nil, nil // Not found (not an error)
	}

	sourceResponse := buildSingleRegistrationResult(*existingReg)
	t.logger.Info(ctx, "Registration retrieved successfully", "connectionID", connectionID, "registrationID", *existingReg.ID, "operation", types.OpGetRegistrationByConnection)
	return sourceResponse, nil
}

// ListRegistrations lists all registrations
func (t *DefaultTaskAPI) ListRegistrations(ctx context.Context) ([]*types.RegistrationResult, *errors.SDKError) {
	start := time.Now()
	t.logger.Debug(ctx, "Listing all registrations", "operation", types.OpListRegistrations)

	data, _, err := t.brsClient.GetBRSClient().GetSourceRegistrationsWithContext(ctx, &backuprecoveryv1.GetSourceRegistrationsOptions{
		XIBMTenantID: core.StringPtr(t.brsClient.GetTenantId()),
	})
	if err != nil {
		t.logger.Error(ctx, "Failed to list registrations", logger.Err(err), "operation", types.OpListRegistrations)
		t.recordOperationFailure(ctx, metrics.OperationListRegistrations, start)
		return nil, errors.NewRegistrationFailedError("ListRegistration API call error", err)
	}

	results := buildRegistrationResults(data)
	t.logger.Info(ctx, "Registrations listed successfully", "count", len(results), "operation", types.OpListRegistrations)

	// Record success
	t.recordOperationSuccess(ctx, metrics.OperationListRegistrations, start)

	return results, nil
}

// UnregisterSource unregisters a source
func (t *DefaultTaskAPI) UnregisterSource(ctx context.Context, registrationID int64) *errors.SDKError {
	startedAt := time.Now()
	t.logger.Info(ctx, "Unregistering source", "registrationID", registrationID, "operation", types.OpUnregisterSource)
	t.emitActivityEvent(ctx, activitytracker.BuildUnregisterSourceActivityEvent(registrationID, activitytracker.EventStateStarted, "starting source unregistration", nil, 1, startedAt, t.getSourceIP(ctx)))

	deleteProtectionSourceOptions := &backuprecoveryv1.DeleteProtectionSourceRegistrationOptions{
		XIBMTenantID: core.StringPtr(t.brsClient.GetTenantId()),
		ID:           core.Int64Ptr(registrationID),
	}

	_, err := t.brsClient.GetBRSClient().DeleteProtectionSourceRegistrationWithContext(ctx, deleteProtectionSourceOptions)
	if err != nil {
		t.logger.Error(ctx, "Failed to unregister source", logger.Err(err), "registrationID", registrationID, "operation", types.OpUnregisterSource)
		t.recordOperationFailure(ctx, metrics.OperationUnregisterSource, startedAt)

		sdkErr := errors.NewRegistrationFailedError("Failed to Unregister the source", err)
		t.emitActivityEvent(ctx, activitytracker.BuildUnregisterSourceActivityEvent(registrationID, activitytracker.EventStateFailed, "source unregistration failed", sdkErr, 1, startedAt, t.getSourceIP(ctx)))
		return sdkErr
	}

	t.logger.Info(ctx, "Source unregistered successfully", "registrationID", registrationID, "operation", types.OpUnregisterSource)
	t.emitActivityEvent(ctx, activitytracker.BuildUnregisterSourceActivityEvent(registrationID, activitytracker.EventStateSucceeded, "source unregistered successfully", nil, 1, startedAt, t.getSourceIP(ctx)))

	// Record success
	t.recordOperationSuccess(ctx, metrics.OperationUnregisterSource, startedAt)

	return nil
}

// RefreshRegistration refreshes a protection source registration
func (t *DefaultTaskAPI) RefreshRegistration(ctx context.Context, registrationID int64) *errors.SDKError {
	// Get source registration to verify it exists
	getSourceOptions := &backuprecoveryv1.GetSourceRegistrationsOptions{
		XIBMTenantID: core.StringPtr(t.brsClient.GetTenantId()),
		Ids:          []int64{registrationID},
	}

	sourceResp, _, err := t.brsClient.GetBRSClient().GetSourceRegistrations(getSourceOptions)
	if err != nil {
		return errors.NewRegistrationFailedError("Failed to get source registration", err)
	}

	if sourceResp == nil || len(sourceResp.Registrations) == 0 {
		return errors.NewSDKError(errors.ErrCodeRegistrationNotFound, "Source registration not found", nil)
	}

	// Refresh the protection source
	refreshOptions := &backuprecoveryv1.RefreshProtectionSourceByIdOptions{
		XIBMTenantID: core.StringPtr(t.brsClient.GetTenantId()),
		ID:           core.Int64Ptr(registrationID),
	}

	_, err = t.brsClient.GetBRSClient().RefreshProtectionSourceByID(refreshOptions)
	if err != nil {
		sdkErr := errors.NewRegistrationFailedError("Failed to refresh protection source", err)
		return sdkErr
	}

	return nil
}

// CreateProtectionGroup creates a protection group (idempotent)
// Returns existing protection group if found by name, creates new if not found
// Microservice-safe: Handles race conditions gracefully
func (t *DefaultTaskAPI) CreateProtectionGroup(ctx context.Context, registrationID int64, groupParams *types.ProtectionGroupParams, dataSource datasources.DataSource) (*types.ProtectionGroupResult, *errors.SDKError) {
	startedAt := time.Now()

	if groupParams == nil || groupParams.Policy == nil || groupParams.Name == "" {
		t.logger.Error(ctx, "Protection group creation failed: invalid parameters", "error", "groupParams.Name is required", "operation", types.OpCreateProtectionGroup)
		t.recordOperationFailure(ctx, metrics.OperationCreateProtectionGroup, startedAt)

		sdkErr := errors.NewSDKError(errors.ErrCodeInvalidInput, "groupParams.Name is required", nil)
		t.emitActivityEvent(ctx, activitytracker.BuildProtectionGroupActivityEvent("", "", registrationID, "", activitytracker.EventStateFailed, "protection group validation failed", sdkErr, 1, startedAt, t.getSourceIP(ctx)))
		return nil, sdkErr
	}

	t.logger.Info(ctx, "Creating protection group", "groupName", groupParams.Name, "registrationID", registrationID, "operation", types.OpCreateProtectionGroup)

	t.emitActivityEvent(ctx, activitytracker.BuildProtectionGroupActivityEvent(groupParams.Name, "", registrationID, groupParams.Policy.ID, activitytracker.EventStateStarted, "starting protection group creation", nil, 1, startedAt, t.getSourceIP(ctx)))

	t.logger.Debug(ctx, "Checking if protection group already exists", "groupName", groupParams.Name)
	existing, err := t.GetProtectionGroupByName(ctx, groupParams.Name)
	if err != nil {
		t.logger.Error(ctx, "Failed to check existing protection group", logger.Err(err), "groupName", groupParams.Name, "registrationID", registrationID, "operation", types.OpCreateProtectionGroup)
		t.emitActivityEvent(ctx, activitytracker.BuildProtectionGroupActivityEvent(groupParams.Name, "", registrationID, groupParams.Policy.ID, activitytracker.EventStateFailed, "failed to lookup existing protection group", err, 1, startedAt, t.getSourceIP(ctx)))
		return nil, err
	}

	if existing != nil {
		// Protection group already exists - return it (idempotent behavior)
		t.logger.Info(ctx, "Protection group already exists, returning existing group",
			"groupName", groupParams.Name,
			"groupID", existing.ProtectionGroupID,
			"idempotent", true, "registrationID", registrationID, "operation", types.OpCreateProtectionGroup)
		t.emitActivityEvent(ctx, activitytracker.BuildProtectionGroupActivityEvent(existing.GroupName, existing.ProtectionGroupID, registrationID, groupParams.Policy.ID, activitytracker.EventStateSucceeded, "protection group already exists; reusing existing protection group", nil, 1, startedAt, t.getSourceIP(ctx)))

		t.recordOperationSuccess(ctx, metrics.OperationCreateProtectionGroup, startedAt)
		return existing, nil
	}

	t.logger.Debug(ctx, "Building protection groupparameters", "groupName", groupParams.Name)
	params, pErr := dataSource.CreateProtectionGroup(ctx, registrationID, groupParams)
	if pErr != nil {
		t.logger.Error(ctx, "Failed to build protection group parameters", logger.Err(pErr), "groupName", groupParams.Name, "registrationID", registrationID, "operation", types.OpCreateProtectionGroup)
		sdkErr := errors.NewSDKError(errors.ErrCodeProtectionFailed, "CreateProtectionGroup BRS API call failed", pErr)
		t.emitActivityEvent(ctx, activitytracker.BuildProtectionGroupActivityEvent(groupParams.Name, "", registrationID, groupParams.Policy.ID, activitytracker.EventStateFailed, "failed to build protection group request", sdkErr, 1, startedAt, t.getSourceIP(ctx)))

		t.recordOperationFailure(ctx, metrics.OperationCreateProtectionGroup, startedAt)

		return nil, errors.NewSDKError(errors.ErrCodeProtectionFailed, "CreateProtectionGroup BRS API call failed", pErr)
	}

	params.XIBMTenantID = core.StringPtr(t.brsClient.GetTenantId())
	params.PolicyID = &groupParams.Policy.ID
	params.Name = &groupParams.Name

	t.logger.Debug(ctx, "Calling BRS API to create protection group", "groupName", groupParams.Name, "registrationID", registrationID, "operation", types.OpCreateProtectionGroup)
	protectionGroupResult, _, pErr := t.brsClient.GetBRSClient().CreateProtectionGroupWithContext(ctx, params)
	if pErr != nil {
		t.logger.Error(ctx, "Failed to create protection group via BRS API", logger.Err(pErr), "groupName", groupParams.Name, "registrationID", registrationID, "operation", types.OpCreateProtectionGroup)
		sdkErr := errors.NewSDKError(errors.ErrCodeProtectionFailed, "failed to create protection group", pErr)
		t.emitActivityEvent(ctx, activitytracker.BuildProtectionGroupActivityEvent(groupParams.Name, "", registrationID, groupParams.Policy.ID, activitytracker.EventStateFailed, "protection group creation failed", sdkErr, 1, startedAt, t.getSourceIP(ctx)))

		t.recordOperationFailure(ctx, metrics.OperationCreateProtectionGroup, startedAt)
		return nil, errors.NewSDKError(errors.ErrCodeProtectionFailed, "failed to create protection group", pErr)
	}

	result := &types.ProtectionGroupResult{
		ProtectionGroupID: *protectionGroupResult.ID,
		GroupName:         *protectionGroupResult.Name,
		RegistrationID:    registrationID,
		PolicyID:          groupParams.Policy.ID,
	}

	t.logger.Info(ctx, "Protection group created successfully",
		"groupName", result.GroupName,
		"groupID", result.ProtectionGroupID,
		"policyID", result.PolicyID, "registrationID", registrationID, "operation", types.OpCreateProtectionGroup)
	t.emitActivityEvent(ctx, activitytracker.BuildProtectionGroupActivityEvent(result.GroupName, result.ProtectionGroupID, registrationID, result.PolicyID, activitytracker.EventStateSucceeded, "protection group created successfully", nil, 1, startedAt, t.getSourceIP(ctx)))

	t.recordOperationSuccess(ctx, metrics.OperationCreateProtectionGroup, startedAt)

	return result, nil
}

// GetProtectionGroup retrieves protection group by ID
func (t *DefaultTaskAPI) GetProtectionGroup(ctx context.Context, groupID string) (*types.ProtectionGroupResult, *errors.SDKError) {
	t.logger.Debug(ctx, "Retrieving protection group by ID", "groupID", groupID, "operation", types.OpGetProtectionGroup)
	start := time.Now()

	params := &backuprecoveryv1.GetProtectionGroupByIdOptions{
		XIBMTenantID: core.StringPtr(t.brsClient.GetTenantId()),
		ID:           &groupID,
	}

	protectionGroup, _, err := t.brsClient.GetBRSClient().GetProtectionGroupByIDWithContext(ctx, params)
	if err != nil {
		t.logger.Error(ctx, "Failed to retrieve protection group", logger.Err(err), "groupID", groupID, "operation", types.OpGetProtectionGroup)
		t.recordOperationFailure(ctx, metrics.OperationGetProtectionGroup, start)
		return nil, errors.NewSDKError(errors.ErrCodeProtectionFailed, "failed to fetch protection group by Id", err)
	}

	// Convert to ProtectionGroupResult
	result := &types.ProtectionGroupResult{
		GroupName:         *protectionGroup.Name,
		ProtectionGroupID: *protectionGroup.ID,
	}

	t.logger.Info(ctx, "Protection group retrieved successfully", "groupID", groupID, "groupName", result.GroupName, "operation", types.OpGetProtectionGroup)

	t.recordOperationSuccess(ctx, metrics.OperationGetProtectionGroup, start)

	return result, nil
}

// GetProtectionGroupByName retrieves protection group by name
func (t *DefaultTaskAPI) GetProtectionGroupByName(ctx context.Context, groupName string) (*types.ProtectionGroupResult, *errors.SDKError) {
	t.logger.Debug(ctx, "Retrieving protection group by name", "groupName", groupName, "operation", types.OpGetProtectionGroupByName)

	params := &backuprecoveryv1.GetProtectionGroupsOptions{
		XIBMTenantID: core.StringPtr(t.brsClient.GetTenantId()),
		Names:        []string{groupName},
	}

	protectionGroups, _, err := t.brsClient.GetBRSClient().GetProtectionGroupsWithContext(ctx, params)
	if err != nil {
		t.logger.Error(ctx, "Failed to get protection group by name", logger.Err(err), "groupName", groupName, "operation", types.OpGetProtectionGroup)
		return nil, errors.NewSDKError(errors.ErrCodeProtectionFailed, "failed to fetch protection group by name", err)
	}
	if protectionGroups != nil && len(protectionGroups.ProtectionGroups) == 0 {
		t.logger.Debug(ctx, "Protection group not found", "groupName", groupName, "operation", types.OpGetProtectionGroup)
		return nil, nil
	}
	// Convert to ProtectionGroupResult
	result := &types.ProtectionGroupResult{
		GroupName:         *protectionGroups.ProtectionGroups[0].Name,
		ProtectionGroupID: *protectionGroups.ProtectionGroups[0].ID,
		Status:            "active",
		CreatedAt:         time.Now(),
	}

	t.logger.Info(ctx, "Protection group retrieved successfully", "groupName", groupName, "groupID", result.ProtectionGroupID, "operation", types.OpGetProtectionGroup)
	return result, nil
}

// ListProtectionGroups lists protection groups for a registration
func (t *DefaultTaskAPI) ListProtectionGroups(ctx context.Context, registrationID int64) ([]*types.ProtectionGroupResult, *errors.SDKError) {
	t.logger.Debug(ctx, "Listing protection groups", "registrationID", registrationID, "operation", types.OpListProtectionGroups)
	start := time.Now()

	params := &backuprecoveryv1.GetProtectionGroupsOptions{
		XIBMTenantID: core.StringPtr(t.brsClient.GetTenantId()),
	}

	protectionGroups, _, err := t.brsClient.GetBRSClient().GetProtectionGroupsWithContext(ctx, params)
	if err != nil {
		t.logger.Error(ctx, "Failed to list protection groups", logger.Err(err), "registrationID", registrationID, "operation", types.OpListProtectionGroups)
		t.recordOperationFailure(ctx, metrics.OperationListProtectionGroups, start)
		return nil, errors.NewSDKError(errors.ErrCodeProtectionFailed, "failed to fetch list of protection groups", err)
	}

	// Convert to ProtectionGroupResult
	result := make([]*types.ProtectionGroupResult, 0)
	for _, protectionGroup := range protectionGroups.ProtectionGroups {
		result = append(result, &types.ProtectionGroupResult{
			GroupName:         *protectionGroup.Name,
			ProtectionGroupID: *protectionGroup.ID,
		})
	}

	t.logger.Info(ctx, "Protection groups listed successfully", "registrationID", registrationID, "count", len(result), "operation", types.OpListProtectionGroups)
	t.recordOperationSuccess(ctx, metrics.OperationListProtectionGroups, start)

	return result, nil
}

// UpdateProtectionGroup updates protection group settings
func (t *DefaultTaskAPI) UpdateProtectionGroup(ctx context.Context, groupID string, updateParams *types.ProtectionGroupParams) (*types.ProtectionGroupResult, *errors.SDKError) {
	t.logger.Debug(ctx, "Updating protection group", "groupID", groupID, "operation", types.OpUpdateProtectionGroup)
	// TODO: Implement - call BRS API to update protection group
	t.logger.Warn(ctx, "UpdateProtectionGroup not yet implemented", "groupID", groupID)
	return nil, errors.NewSDKError(errors.ErrCodeUnknown, "UpdateProtectionGroup not yet implemented", nil)
}

// DeleteProtectionGroup deletes a protection group
func (t *DefaultTaskAPI) DeleteProtectionGroup(ctx context.Context, groupID string, deleteSnapshots bool) *errors.SDKError {
	startedAt := time.Now()
	t.logger.Info(ctx, "Deleting protection group", "groupID", groupID, "deleteSnapshots", deleteSnapshots, "operation", types.OpDeleteProtectionGroup)
	t.emitActivityEvent(ctx, activitytracker.BuildDeleteProtectionGroupActivityEvent(groupID, activitytracker.EventStateStarted, "starting protection group deletion", nil, 1, startedAt, t.getSourceIP(ctx)))

	params := &backuprecoveryv1.DeleteProtectionGroupOptions{
		XIBMTenantID:    core.StringPtr(t.brsClient.GetTenantId()),
		ID:              &groupID,
		DeleteSnapshots: core.BoolPtr(deleteSnapshots),
	}

	_, err := t.brsClient.GetBRSClient().DeleteProtectionGroupWithContext(ctx, params)
	if err != nil {
		t.logger.Error(ctx, "Failed to delete protection group", logger.Err(err), "groupID", groupID, "operation", types.OpDeleteProtectionGroup)
		t.recordOperationFailure(ctx, metrics.OperationDeleteProtectionGroup, startedAt)

		sdkErr := errors.NewSDKError(errors.ErrCodeProtectionFailed, "failed to delete protection group by Id", err)
		t.emitActivityEvent(ctx, activitytracker.BuildDeleteProtectionGroupActivityEvent(groupID, activitytracker.EventStateFailed, "protection group deletion failed", sdkErr, 1, startedAt, t.getSourceIP(ctx)))
		return sdkErr
	}

	t.logger.Info(ctx, "Protection group deleted successfully", "groupID", groupID, "operation", types.OpDeleteProtectionGroup)
	t.emitActivityEvent(ctx, activitytracker.BuildDeleteProtectionGroupActivityEvent(groupID, activitytracker.EventStateSucceeded, "protection group deleted successfully", nil, 1, startedAt, t.getSourceIP(ctx)))

	// Record success
	t.recordOperationSuccess(ctx, metrics.OperationDeleteProtectionGroup, startedAt)

	return nil
}

// CreatePolicy creates a new policy (idempotent)
// Returns existing policy if found by name, creates new if not found
// Microservice-safe: Handles race conditions gracefully
func (t *DefaultTaskAPI) CreatePolicy(ctx context.Context, policyParams *types.PolicyParams) (*types.PolicyResult, *errors.SDKError) {
	startedAt := time.Now()

	if policyParams == nil || policyParams.Name == "" {
		t.logger.Error(ctx, "Policy creation failed: invalid parameters", "error", "policyParams.Name is required", "operation", types.OpCreatePolicy)
		sdkErr := errors.NewSDKError(errors.ErrCodeInvalidInput, "policyParams.Name is required", nil)
		t.emitActivityEvent(ctx, activitytracker.BuildPolicyActivityEvent("", activitytracker.EventStateFailed, "policy validation failed", sdkErr, 1, startedAt, t.getSourceIP(ctx)))
		t.recordOperationFailure(ctx, metrics.OperationCreatePolicy, startedAt)
		return nil, sdkErr
	}

	t.logger.Info(ctx, "Creating policy", "policyName", policyParams.Name, "operation", types.OpCreatePolicy)

	t.emitActivityEvent(ctx, activitytracker.BuildPolicyActivityEvent(policyParams.Name, activitytracker.EventStateStarted, "starting policy creation", nil, 1, startedAt, t.getSourceIP(ctx)))

	t.logger.Debug(ctx, "Checking if policy already exists", "policyName", policyParams.Name)
	existing, err := t.GetPolicyByName(ctx, policyParams.Name)
	if err != nil {
		t.logger.Error(ctx, "Failed to check existing policy", logger.Err(err), "policyName", policyParams.Name, "operation", types.OpCreatePolicy)
		t.emitActivityEvent(ctx, activitytracker.BuildPolicyActivityEvent(policyParams.Name, activitytracker.EventStateFailed, "failed to lookup existing policy", err, 1, startedAt, t.getSourceIP(ctx)))
		return nil, err
	}

	if existing != nil {
		t.logger.Info(ctx, "Policy already exists, returning existing policy",
			"policyName", policyParams.Name,
			"policyID", existing.ID,
			"idempotent", true, "operation", types.OpCreatePolicy)
		t.emitActivityEvent(ctx, activitytracker.BuildPolicyActivityEvent(policyParams.Name, activitytracker.EventStateSucceeded, "policy already exists; reusing existing policy", nil, 1, startedAt, t.getSourceIP(ctx)))
		return existing, nil
	}

	backupPolicy := &backuprecoveryv1.BackupPolicy{
		Regular: &backuprecoveryv1.RegularBackupPolicy{},
	}

	if policyParams.IncrementalBackup != nil {
		incremental, err := buildIncrementalSchedule(policyParams.IncrementalBackup)
		if err != nil {
			sdkErr := errors.NewSDKError(errors.ErrCodeUnknown, "failed to create incremental schedule", err)
			t.emitActivityEvent(ctx, activitytracker.BuildPolicyActivityEvent(policyParams.Name, activitytracker.EventStateFailed, "failed to build incremental schedule", sdkErr, 1, startedAt, t.getSourceIP(ctx)))
			return nil, sdkErr
		}
		backupPolicy.Regular.Incremental = incremental
	}

	if policyParams.DataRetention != nil {
		retention, err := buildRetention(policyParams.DataRetention)
		if err != nil {
			sdkErr := errors.NewSDKError(errors.ErrCodeProtectionFailed, "Unable to build DataRetention params", err)
			t.emitActivityEvent(ctx, activitytracker.BuildPolicyActivityEvent(policyParams.Name, activitytracker.EventStateFailed, "failed to build data retention settings", sdkErr, 1, startedAt, t.getSourceIP(ctx)))
			return nil, sdkErr
		}
		backupPolicy.Regular.Retention = retention
	}

	if policyParams.PrimaryBackupTarget != nil {
		backupPolicy.Regular.PrimaryBackupTarget = &backuprecoveryv1.PrimaryBackupTarget{
			UseDefaultBackupTarget: &policyParams.PrimaryBackupTarget.UseDefaultBackupTarget,
		}
	}

	if policyParams.FullBackups != nil {
		full, err := buildFullSchedule(policyParams.FullBackups)
		if err != nil {
			sdkErr := errors.NewSDKError(errors.ErrCodeUnknown, "failed to create full schedule", err)
			t.emitActivityEvent(ctx, activitytracker.BuildPolicyActivityEvent(policyParams.Name, activitytracker.EventStateFailed, "failed to build full backup schedule", sdkErr, 1, startedAt, t.getSourceIP(ctx)))
			return nil, sdkErr
		}
		backupPolicy.Regular.FullBackups = full
	}

	t.logger.Debug(ctx, "Building policy parameters", "policyName", policyParams.Name, "operation", types.OpCreatePolicy)
	CreateProtectionPolicyOptions := &backuprecoveryv1.CreateProtectionPolicyOptions{
		XIBMTenantID: core.StringPtr(t.brsClient.GetTenantId()),
		Name:         &policyParams.Name,
		BackupPolicy: backupPolicy,
	}

	if policyParams.RetryOption != nil {
		CreateProtectionPolicyOptions.RetryOptions = buildRetryOptions(*policyParams.RetryOption)
	}

	if len(policyParams.FullBackups) > 0 {
		CreateProtectionPolicyOptions.IsCBSEnabled = core.BoolPtr(true)
	}

	t.logger.Debug(ctx, "Calling BRS API to create policy", "policyName", policyParams.Name, "operation", types.OpCreatePolicy)
	result, _, serr := t.brsClient.GetBRSClient().CreateProtectionPolicyWithContext(ctx, CreateProtectionPolicyOptions)
	if serr != nil {
		t.logger.Error(ctx, "Failed to create policy via BRS API", logger.Err(serr), "policyName", policyParams.Name, "operation", types.OpCreatePolicy)
		sdkErr := errors.NewSDKError(errors.ErrCodeUnknown, "failed to create policy", serr)
		t.emitActivityEvent(ctx, activitytracker.BuildPolicyActivityEvent(policyParams.Name, activitytracker.EventStateFailed, "policy creation request failed", sdkErr, 1, startedAt, t.getSourceIP(ctx)))
		return nil, sdkErr
	}

	policyResult := &types.PolicyResult{
		ID:   *result.ID,
		Name: *result.Name,
	}

	t.logger.Info(ctx, "Policy created successfully",
		"policyName", policyResult.Name,
		"policyID", policyResult.ID, "operation", types.OpCreatePolicy)
	t.emitActivityEvent(ctx, activitytracker.BuildPolicyActivityEvent(policyResult.Name, activitytracker.EventStateSucceeded, "policy created successfully", nil, 1, startedAt, t.getSourceIP(ctx)))

	return policyResult, nil
}

// GetPolicy retrieves policy by ID
func (t *DefaultTaskAPI) GetPolicy(ctx context.Context, policyID string) (*types.PolicyResult, *errors.SDKError) {
	t.logger.Debug(ctx, "Retrieving policy by ID", "policyID", policyID, "operation", types.OpGetPolicy)
	startedAt := time.Now()

	params := &backuprecoveryv1.GetProtectionPolicyByIdOptions{
		XIBMTenantID: core.StringPtr(t.brsClient.GetTenantId()),
		ID:           &policyID,
	}

	protectionPolicy, _, err := t.brsClient.GetBRSClient().GetProtectionPolicyByIDWithContext(ctx, params)
	if err != nil {
		t.logger.Error(ctx, "Failed to get policy", logger.Err(err), "policyID", policyID)
		t.recordOperationFailure(ctx, metrics.OperationGetPolicy, startedAt)
		return nil, errors.NewSDKError(errors.ErrCodeProtectionFailed, "failed to fetch protection policy by id", err)
	}

	// Convert to PolicyResult
	result := &types.PolicyResult{
		Name:      safeString(protectionPolicy.Name),
		ID:        safeString(protectionPolicy.ID),
		CreatedAt: time.Now(),
	}

	if protectionPolicy.BackupPolicy != nil && protectionPolicy.BackupPolicy.Regular != nil {
		if protectionPolicy.BackupPolicy.Regular.Retention != nil && protectionPolicy.BackupPolicy.Regular.Retention.Duration != nil {
			result.RetentionDays = *protectionPolicy.BackupPolicy.Regular.Retention.Duration
		}
	}

	t.logger.Info(ctx, "Policy retrieved successfully", "policyID", policyID, "policyName", result.Name, "operation", types.OpGetPolicy)

	// Record success
	t.recordOperationSuccess(ctx, metrics.OperationGetPolicy, startedAt)

	return result, nil
}

// GetPolicyByName retrieves policy by name (for idempotency checks)
func (t *DefaultTaskAPI) GetPolicyByName(ctx context.Context, policyName string) (*types.PolicyResult, *errors.SDKError) {
	t.logger.Debug(ctx, "Retrieving policy by name", "policyName", policyName, "operation", types.OpGetPolicyByName)

	params := &backuprecoveryv1.GetProtectionPoliciesOptions{
		XIBMTenantID: core.StringPtr(t.brsClient.GetTenantId()),
		PolicyNames:  []string{policyName},
	}

	protectionPolicies, _, err := t.brsClient.GetBRSClient().GetProtectionPoliciesWithContext(ctx, params)
	if err != nil {
		t.logger.Error(ctx, "Failed to get policy by name", logger.Err(err), "policyName", policyName, "operation", types.OpGetPolicyByName)
		return nil, errors.NewSDKError(errors.ErrCodeProtectionFailed, "failed to fetch protection policy by name", err)
	}
	if protectionPolicies != nil && len(protectionPolicies.Policies) == 0 {
		t.logger.Debug(ctx, "Policy not found", "policyName", policyName)
		return nil, nil
	}
	result := &types.PolicyResult{
		Name:      safeString(protectionPolicies.Policies[0].Name),
		ID:        safeString(protectionPolicies.Policies[0].ID),
		Status:    "active",
		CreatedAt: time.Now(),
	}

	if protectionPolicies.Policies[0].BackupPolicy != nil && protectionPolicies.Policies[0].BackupPolicy.Regular != nil {
		if protectionPolicies.Policies[0].BackupPolicy.Regular.Retention != nil && protectionPolicies.Policies[0].BackupPolicy.Regular.Retention.Duration != nil {
			result.RetentionDays = *protectionPolicies.Policies[0].BackupPolicy.Regular.Retention.Duration
		}
	}

	t.logger.Info(ctx, "Policy retrieved successfully", "policyName", policyName, "policyID", result.ID, "operation", types.OpGetPolicyByName)
	return result, nil
}

// ListPolicies lists all policies
func (t *DefaultTaskAPI) ListPolicies(ctx context.Context) ([]*types.PolicyResult, *errors.SDKError) {
	start := time.Now()
	t.logger.Debug(ctx, "Listing all policies", "operation", types.OpListPolicies)

	params := &backuprecoveryv1.GetProtectionPoliciesOptions{
		XIBMTenantID: core.StringPtr(t.brsClient.GetTenantId()),
	}

	protectionPolicies, _, err := t.brsClient.GetBRSClient().GetProtectionPoliciesWithContext(ctx, params)
	if err != nil {
		t.logger.Error(ctx, "Failed to list policies", logger.Err(err), "operation", types.OpListPolicies)
		t.recordOperationFailure(ctx, metrics.OperationListPolicies, start)
		return nil, errors.NewSDKError(errors.ErrCodeProtectionFailed, "failed to fetch protection policies", err)
	}

	if protectionPolicies == nil || len(protectionPolicies.Policies) == 0 {
		t.logger.Info(ctx, "No policies found", "operation", types.OpListPolicies)
		return []*types.PolicyResult{}, nil
	}

	result := make([]*types.PolicyResult, 0, len(protectionPolicies.Policies))

	for _, protectionPolicy := range protectionPolicies.Policies {
		policyResult := &types.PolicyResult{
			Name:      safeString(protectionPolicy.Name),
			ID:        safeString(protectionPolicy.ID),
			Status:    "active",
			CreatedAt: time.Now(),
		}

		if protectionPolicy.BackupPolicy != nil && protectionPolicy.BackupPolicy.Regular != nil {
			if protectionPolicy.BackupPolicy.Regular.Retention != nil && protectionPolicy.BackupPolicy.Regular.Retention.Duration != nil {
				policyResult.RetentionDays = *protectionPolicy.BackupPolicy.Regular.Retention.Duration
			}

			if protectionPolicy.BackupPolicy.Regular.Incremental != nil {
				policyResult.BackupType = "incremental"
			}

			if len(protectionPolicy.BackupPolicy.Regular.FullBackups) > 0 {
				if policyResult.BackupType == "" {
					policyResult.BackupType = "full"
				} else {
					policyResult.BackupType = "incremental+full"
				}
			}
		}

		result = append(result, policyResult)
	}

	t.logger.Info(ctx, "Policies listed successfully", "count", len(result), "operation", types.OpListPolicies)
	// Record success
	t.recordOperationSuccess(ctx, metrics.OperationListPolicies, start)

	return result, nil
}

// UpdatePolicy updates policy settings
func (t *DefaultTaskAPI) UpdatePolicy(ctx context.Context, policyID string, policyParams *types.PolicyParams) (*types.PolicyResult, *errors.SDKError) {
	startedAt := time.Now()

	if policyID == "" {
		t.logger.Error(ctx, "Policy update failed: invalid parameters", "error", "policyID is required", "operation", types.OpUpdatePolicy)
		sdkErr := errors.NewSDKError(errors.ErrCodeInvalidInput, "policyID is required", nil)
		t.emitActivityEvent(ctx, activitytracker.BuildPolicyActivityEvent("", activitytracker.EventStateFailed, "policy validation failed", sdkErr, 1, startedAt, t.getSourceIP(ctx)))
		t.recordOperationFailure(ctx, metrics.OperationUpdatePolicy, startedAt)
		return nil, sdkErr
	}

	if policyParams == nil || policyParams.Name == "" {
		t.logger.Error(ctx, "Policy update failed: invalid parameters", "error", "policyParams.Name is required", "operation", types.OpUpdatePolicy)
		sdkErr := errors.NewSDKError(errors.ErrCodeInvalidInput, "policyParams.Name is required", nil)
		t.emitActivityEvent(ctx, activitytracker.BuildPolicyActivityEvent("", activitytracker.EventStateFailed, "policy validation failed", sdkErr, 1, startedAt, t.getSourceIP(ctx)))
		t.recordOperationFailure(ctx, metrics.OperationUpdatePolicy, startedAt)
		return nil, sdkErr
	}

	t.logger.Info(ctx, "Updating policy", "policyID", policyID, "policyName", policyParams.Name, "operation", types.OpUpdatePolicy)

	t.emitActivityEvent(ctx, activitytracker.BuildPolicyActivityEvent(policyParams.Name, activitytracker.EventStateStarted, "starting policy update", nil, 1, startedAt, t.getSourceIP(ctx)))

	backupPolicy := &backuprecoveryv1.BackupPolicy{
		Regular: &backuprecoveryv1.RegularBackupPolicy{},
	}

	if policyParams.IncrementalBackup != nil {
		incremental, err := buildIncrementalSchedule(policyParams.IncrementalBackup)
		if err != nil {
			sdkErr := errors.NewSDKError(errors.ErrCodeUnknown, "failed to create incremental schedule", err)
			t.emitActivityEvent(ctx, activitytracker.BuildPolicyActivityEvent(policyParams.Name, activitytracker.EventStateFailed, "failed to build incremental schedule", sdkErr, 1, startedAt, t.getSourceIP(ctx)))
			return nil, sdkErr
		}
		backupPolicy.Regular.Incremental = incremental
	}

	if policyParams.DataRetention != nil {
		retention, err := buildRetention(policyParams.DataRetention)
		if err != nil {
			sdkErr := errors.NewSDKError(errors.ErrCodeProtectionFailed, "Unable to build DataRetention params", err)
			t.emitActivityEvent(ctx, activitytracker.BuildPolicyActivityEvent(policyParams.Name, activitytracker.EventStateFailed, "failed to build data retention settings", sdkErr, 1, startedAt, t.getSourceIP(ctx)))
			return nil, sdkErr
		}
		backupPolicy.Regular.Retention = retention
	}

	if policyParams.PrimaryBackupTarget != nil {
		backupPolicy.Regular.PrimaryBackupTarget = &backuprecoveryv1.PrimaryBackupTarget{
			UseDefaultBackupTarget: &policyParams.PrimaryBackupTarget.UseDefaultBackupTarget,
		}
	}

	if policyParams.FullBackups != nil {
		full, err := buildFullSchedule(policyParams.FullBackups)
		if err != nil {
			sdkErr := errors.NewSDKError(errors.ErrCodeUnknown, "failed to create full schedule", err)
			t.emitActivityEvent(ctx, activitytracker.BuildPolicyActivityEvent(policyParams.Name, activitytracker.EventStateFailed, "failed to build full backup schedule", sdkErr, 1, startedAt, t.getSourceIP(ctx)))
			return nil, sdkErr
		}
		backupPolicy.Regular.FullBackups = full
	}

	t.logger.Debug(ctx, "Building policy update parameters", "policyID", policyID, "policyName", policyParams.Name, "operation", types.OpUpdatePolicy)
	UpdateProtectionPolicyOptions := &backuprecoveryv1.UpdateProtectionPolicyOptions{
		ID:           core.StringPtr(policyID),
		XIBMTenantID: core.StringPtr(t.brsClient.GetTenantId()),
		Name:         &policyParams.Name,
		BackupPolicy: backupPolicy,
	}

	if policyParams.RetryOption != nil {
		UpdateProtectionPolicyOptions.RetryOptions = buildRetryOptions(*policyParams.RetryOption)
	}

	if len(policyParams.FullBackups) > 0 {
		UpdateProtectionPolicyOptions.IsCBSEnabled = core.BoolPtr(true)
	}

	t.logger.Debug(ctx, "Calling BRS API to update policy", "policyID", policyID, "policyName", policyParams.Name, "operation", types.OpUpdatePolicy)
	result, _, serr := t.brsClient.GetBRSClient().UpdateProtectionPolicyWithContext(ctx, UpdateProtectionPolicyOptions)
	if serr != nil {
		t.logger.Error(ctx, "Failed to update policy via BRS API", logger.Err(serr), "policyID", policyID, "policyName", policyParams.Name, "operation", types.OpUpdatePolicy)
		sdkErr := errors.NewSDKError(errors.ErrCodeUnknown, "failed to update policy", serr)
		t.emitActivityEvent(ctx, activitytracker.BuildPolicyActivityEvent(policyParams.Name, activitytracker.EventStateFailed, "policy update request failed", sdkErr, 1, startedAt, t.getSourceIP(ctx)))
		return nil, sdkErr
	}

	policyResult := &types.PolicyResult{
		ID:   *result.ID,
		Name: *result.Name,
	}

	t.logger.Info(ctx, "Policy updated successfully",
		"policyName", policyResult.Name,
		"policyID", policyResult.ID, "operation", types.OpUpdatePolicy)
	t.emitActivityEvent(ctx, activitytracker.BuildPolicyActivityEvent(policyResult.Name, activitytracker.EventStateSucceeded, "policy updated successfully", nil, 1, startedAt, t.getSourceIP(ctx)))
	t.recordOperationSuccess(ctx, metrics.OperationUpdatePolicy, startedAt)

	return policyResult, nil
}

// DeletePolicy deletes a policy
func (t *DefaultTaskAPI) DeletePolicy(ctx context.Context, policyID string) *errors.SDKError {
	startedAt := time.Now()
	t.logger.Info(ctx, "Deleting policy", "policyID", policyID, "operation", types.OpDeletePolicy)
	t.emitActivityEvent(ctx, activitytracker.BuildDeletePolicyActivityEvent(policyID, activitytracker.EventStateStarted, "starting policy deletion", nil, 1, startedAt, t.getSourceIP(ctx)))

	params := &backuprecoveryv1.DeleteProtectionPolicyOptions{
		XIBMTenantID: core.StringPtr(t.brsClient.GetTenantId()),
		ID:           &policyID,
	}

	_, err := t.brsClient.GetBRSClient().DeleteProtectionPolicyWithContext(ctx, params)
	if err != nil {
		t.logger.Error(ctx, "Failed to delete policy", logger.Err(err), "policyID", policyID, "operation", types.OpDeletePolicy)
		t.recordOperationFailure(ctx, metrics.OperationDeletePolicy, startedAt)

		sdkErr := errors.NewSDKError(errors.ErrCodeProtectionFailed, "failed to delete protection POlicy by Id", err)
		t.emitActivityEvent(ctx, activitytracker.BuildDeletePolicyActivityEvent(policyID, activitytracker.EventStateFailed, "policy deletion failed", sdkErr, 1, startedAt, t.getSourceIP(ctx)))
		return sdkErr
	}

	t.logger.Info(ctx, "Policy deleted successfully", "policyID", policyID, "operation", types.OpDeletePolicy)

	// Record success
	t.recordOperationSuccess(ctx, metrics.OperationDeletePolicy, startedAt)

	t.emitActivityEvent(ctx, activitytracker.BuildDeletePolicyActivityEvent(policyID, activitytracker.EventStateSucceeded, "policy deleted successfully", nil, 1, startedAt, t.getSourceIP(ctx)))
	return nil
}

// RunBackup runs a backup job (async operation)
func (t *DefaultTaskAPI) RunBackup(ctx context.Context, groupID string, backupParams *types.BackupParams) (*types.BackupResult, *errors.SDKError) {
	startedAt := time.Now()
	backupType := ""
	if backupParams != nil {
		backupType = string(backupParams.BackupType)
	}
	t.logger.Info(ctx, "Running backup", "groupID", groupID, "backupType", backupType, "operation", types.OpRunBackup)
	t.emitActivityEvent(ctx, activitytracker.BuildBackupActivityEvent(groupID, "", backupType, activitytracker.EventStateStarted, "starting backup run", nil, 1, startedAt, t.getSourceIP(ctx)))

	createProtectionGroupRunOptions := &backuprecoveryv1.CreateProtectionGroupRunOptions{
		XIBMTenantID: core.StringPtr(t.brsClient.GetTenantId()),
		ID:           core.StringPtr(groupID),
		RunType:      core.StringPtr(string(backupParams.BackupType)),
	}

	targetBackupObjects := make([]backuprecoveryv1.RunObject, 0)
	for _, object := range backupParams.TargetBackupObjectIDs {
		targetBackupObjects = append(targetBackupObjects, backuprecoveryv1.RunObject{ID: object.ID})
	}

	createProtectionGroupRunOptions.Objects = targetBackupObjects

	t.logger.Debug(ctx, "Calling BRS API to create backup run", "groupID", groupID, "backupType", backupType)
	result, _, err := t.brsClient.GetBRSClient().CreateProtectionGroupRunWithContext(ctx, createProtectionGroupRunOptions)
	if err != nil {
		t.logger.Error(ctx, "Failed to create backup run", logger.Err(err), "groupID", groupID, "backupType", backupType, "operation", types.OpRunBackup)
		t.recordOperationFailure(ctx, metrics.OperationRunBackup, startedAt)

		sdkErr := errors.NewSDKError(errors.ErrCodeUnknown, "Failed to create protection group run", err)
		t.emitActivityEvent(ctx, activitytracker.BuildBackupActivityEvent(groupID, "", backupType, activitytracker.EventStateFailed, "backup run creation failed", sdkErr, 1, startedAt, t.getSourceIP(ctx)))
		return nil, sdkErr
	}

	backupResult := &types.BackupResult{
		ProtectionGroupID: *result.ProtectionGroupID,
	}

	t.logger.Info(ctx, "Backup run created successfully", "groupID", groupID, "backupType", backupType, "operation", types.OpRunBackup)
	t.emitActivityEvent(ctx, activitytracker.BuildBackupActivityEvent(groupID, backupResult.BackupID, backupType, activitytracker.EventStateSucceeded, "backup run created successfully", nil, 1, startedAt, t.getSourceIP(ctx)))

	// Record success
	t.recordOperationSuccess(ctx, metrics.OperationRunBackup, startedAt)

	return backupResult, nil
}

// Get active run on protection group.
func (t *DefaultTaskAPI) getActiveBackupRun(ctx context.Context, groupID string) (*types.BackupResult, *errors.SDKError) {

	var totalTimeout time.Duration = types.GetBackup_totalTimeout
	var pollingInterval time.Duration = types.GetBackup_pollingInterval

	deadline := time.Now().Add(totalTimeout)

	for time.Now().Before(deadline) {
		backupRuns, sdkErr := t.ListBackups(ctx, groupID)
		if sdkErr != nil {
			return nil, errors.NewSDKError(errors.ErrCodeUnknown, "Failed to fetch backup runs for group Id", sdkErr)
		}

		if len(backupRuns) > 0 && backupRuns[0].Status == string(types.BackupRun_Status_Running) {
			return backupRuns[0], nil
		}
		time.Sleep(pollingInterval)
	}
	return nil, nil
}

// GetBackup retrieves backup by ID
func (t *DefaultTaskAPI) GetBackup(ctx context.Context, backupID, groupID string) (*types.BackupResult, *errors.SDKError) {
	start := time.Now()
	t.logger.Info(ctx, "Retrieving backup by ID", "backupID", backupID, "groupID", groupID, "operation", types.OpGetBackup)

	getProtectionGroupRunOptions := &backuprecoveryv1.GetProtectionGroupRunOptions{
		XIBMTenantID: core.StringPtr(t.brsClient.GetTenantId()),
		RunID:        core.StringPtr(backupID),
		ID:           core.StringPtr(groupID),
	}

	result, _, err := t.brsClient.GetBRSClient().GetProtectionGroupRunWithContext(ctx, getProtectionGroupRunOptions)
	if err != nil {
		t.logger.Error(ctx, "Failed to get backup", logger.Err(err), "backupID", backupID, "groupID", groupID, "operation", types.OpGetBackup)
		t.recordOperationFailure(ctx, metrics.OperationGetBackup, start)
		return nil, errors.NewSDKError(errors.ErrCodeUnknown, "Failed to fetch backup run by id", err)
	}

	backupRunResult, err := t.buildBackupRunResult(ctx, result)
	if err != nil {
		t.logger.Error(ctx, "Failed to build backup result", logger.Err(err), "backupID", backupID, "groupID", groupID, "operation", types.OpGetBackup)
		t.recordOperationFailure(ctx, metrics.OperationGetBackup, start)
		return nil, errors.NewSDKError(errors.ErrCodeUnknown, fmt.Sprintf("Failed to fetch backup run: Unable to form backupRunResult for backupId %v", *result.ID), err)
	}

	t.recordOperationSuccess(ctx, metrics.OperationGetBackup, start)
	t.logger.Info(ctx, "Backup retrieved successfully", "backupID", backupID, "groupID", groupID, "status", backupRunResult.Status, "operation", types.OpGetBackup)
	return backupRunResult, nil

}

// ListBackups lists backups for a protection group
func (t *DefaultTaskAPI) ListBackups(ctx context.Context, groupID string) ([]*types.BackupResult, *errors.SDKError) {
	t.logger.Debug(ctx, "Listing backups", "groupID", groupID, "operation", types.OpListBackups)

	getProtectionGroupRunsOptions := &backuprecoveryv1.GetProtectionGroupRunsOptions{
		XIBMTenantID: core.StringPtr(t.brsClient.GetTenantId()),
		ID:           core.StringPtr(groupID),
	}

	result, _, err := t.brsClient.GetBRSClient().GetProtectionGroupRunsWithContext(ctx, getProtectionGroupRunsOptions)
	if err != nil {
		t.logger.Error(ctx, "Failed to list backups", logger.Err(err), "groupID", groupID, "operation", types.OpListBackups)
		return nil, errors.NewSDKError(errors.ErrCodeUnknown, "Failed to fetch backup runs", err)
	}
	backupRuns := make([]*types.BackupResult, 0)

	for _, backuprun := range result.Runs {
		backupRunResult, err := t.buildBackupRunResult(ctx, &backuprun)
		if err != nil {
			t.logger.Error(ctx, "Failed to build backup result", logger.Err(err), "backupID", *backuprun.ID, "groupID", groupID, "operation", types.OpListBackups)
			return nil, errors.NewSDKError(errors.ErrCodeUnknown, fmt.Sprintf("Failed to fetch list of backup runs: Unable to form backupRunResult for backupId %v", *backuprun.ID), err)
		}
		backupRuns = append(backupRuns, backupRunResult)
	}

	t.logger.Info(ctx, "Backups listed successfully", "groupID", groupID, "count", len(backupRuns), "operation", types.OpListBackups)
	return backupRuns, nil
}

func (t *DefaultTaskAPI) GetProtectionRunProgress(ctx context.Context, backupID string) (*float32, *errors.SDKError) {
	start := time.Now()
	t.logger.Debug(ctx, "Getting backup progress", "backupID", backupID, "operation", types.OpGetProtectionRunProgress)

	getProtectionRunProgressOptions := &backuprecoveryv1.GetProtectionRunProgressOptions{
		XIBMTenantID: core.StringPtr(t.brsClient.GetTenantId()),
		RunID:        core.StringPtr(backupID),
	}
	result, _, err := t.brsClient.GetBRSClient().GetProtectionRunProgressWithContext(ctx, getProtectionRunProgressOptions)
	if err != nil {
		t.logger.Error(ctx, "Failed to get backup progress", logger.Err(err), "backupID", backupID, "operation", types.OpGetProtectionRunProgress)
		t.recordOperationFailure(ctx, metrics.OperationGetBackup, start)
		return nil, errors.NewSDKError(errors.ErrCodeUnknown, "Unable to get progress percentage for backupjob", err)
	}

	if result == nil || len(result.ArchivalRun) == 0 || result.ArchivalRun[0].PercentageCompleted == nil {
		t.logger.Debug(ctx, "No progress data available", "backupID", backupID, "operation", types.OpGetProtectionRunProgress)
		t.recordOperationSuccess(ctx, metrics.OperationGetBackup, start)
		return nil, nil
	}

	t.recordOperationSuccess(ctx, metrics.OperationGetBackup, start)
	t.logger.Info(ctx, "Backup progress retrieved", "backupID", backupID, "progress", *result.ArchivalRun[0].PercentageCompleted, "operation", types.OpGetProtectionRunProgress)
	return result.ArchivalRun[0].PercentageCompleted, nil
}

// WaitForBackup waits for backup to complete
func (t *DefaultTaskAPI) WaitForBackup(ctx context.Context, backupID string, timeout time.Duration) (*types.BackupResult, *errors.SDKError) {
	t.logger.Debug(ctx, "Waiting for backup completion", "backupID", backupID, "timeout", timeout, "operation", types.OpWaitForBackup)
	// TODO: Implement - poll BRS until backup completes
	t.logger.Warn(ctx, "WaitForBackup not yet implemented", "backupID", backupID)
	return nil, errors.NewSDKError(errors.ErrCodeUnknown, "WaitForBackup not yet implemented", nil)
}

// RunRestore runs a restore job (async operation)
func (t *DefaultTaskAPI) RunRestore(ctx context.Context, groupId, backupId string, targetRegistrationID int64, restoreParams *types.RestoreParams, datasource datasources.DataSource) (*types.RestoreResult, *errors.SDKError) {
	startedAt := time.Now()
	t.logger.Info(ctx, "Running restore", "groupID", groupId, "backupID", backupId, "targetRegistrationID", targetRegistrationID, "operation", types.OpRunRestore)
	t.emitActivityEvent(ctx, activitytracker.BuildRestoreActivityEvent(groupId, backupId, "", targetRegistrationID, activitytracker.EventStateStarted, "starting restore run", nil, 1, startedAt, t.getSourceIP(ctx)))

	t.logger.Debug(ctx, "Building restore parameters", "groupID", groupId, "backupID", backupId, "operation", types.OpRunRestore)
	recoveryParams, err := datasource.RunRestore(ctx, groupId, backupId, targetRegistrationID, restoreParams)
	if err != nil {
		t.logger.Error(ctx, "Failed to build restore parameters", logger.Err(err), "groupID", groupId, "backupID", backupId, "operation", types.OpRunRestore)
		sdkErr := errors.NewSDKError(errors.ErrCodeUnknown, "Failed to build recovery parameters", err)
		t.emitActivityEvent(ctx, activitytracker.BuildRestoreActivityEvent(groupId, backupId, "", targetRegistrationID, activitytracker.EventStateFailed, "failed to build restore request", sdkErr, 1, startedAt, t.getSourceIP(ctx)))
		t.recordOperationFailure(ctx, metrics.OperationRunRestore, startedAt)
		return nil, sdkErr
	}
	recoveryParams.XIBMTenantID = core.StringPtr(t.brsClient.GetTenantId())

	t.logger.Debug(ctx, "Calling BRS API to create restore run", "groupID", groupId, "backupID", backupId)
	result, _, err := t.brsClient.GetBRSClient().CreateRecoveryWithContext(ctx, recoveryParams)
	if err != nil {
		t.logger.Error(ctx, "Failed to create restore run", logger.Err(err), "groupID", groupId, "backupID", backupId, "operation", types.OpRunRestore)
		sdkErr := errors.NewSDKError(errors.ErrCodeUnknown, "Failed to run recovery", err)
		t.emitActivityEvent(ctx, activitytracker.BuildRestoreActivityEvent(groupId, backupId, "", targetRegistrationID, activitytracker.EventStateFailed, "restore run creation failed", sdkErr, 1, startedAt, t.getSourceIP(ctx)))
		t.recordOperationFailure(ctx, metrics.OperationRunRestore, startedAt)
		return nil, sdkErr
	}

	restoreResult := &types.RestoreResult{
		RestoreID: *result.ID,
	}

	t.recordOperationSuccess(ctx, metrics.OperationRunRestore, startedAt)
	t.logger.Info(ctx, "Restore run created successfully", "groupID", groupId, "backupID", backupId, "restoreID", restoreResult.RestoreID, "operation", types.OpRunRestore)
	t.emitActivityEvent(ctx, activitytracker.BuildRestoreActivityEvent(groupId, backupId, restoreResult.RestoreID, targetRegistrationID, activitytracker.EventStateSucceeded, "restore run created successfully", nil, 1, startedAt, t.getSourceIP(ctx)))
	return restoreResult, nil
}

// GetRestore retrieves restore by ID
func (t *DefaultTaskAPI) GetRestore(ctx context.Context, recoveryId string) (*types.RestoreResult, *errors.SDKError) {
	start := time.Now()
	t.logger.Info(ctx, "Retrieving restore by ID", "restoreID", recoveryId, "operation", types.OpGetRestore)

	getRecoveryByIdOptions := &backuprecoveryv1.GetRecoveryByIdOptions{
		XIBMTenantID: core.StringPtr(t.brsClient.GetTenantId()),
		ID:           core.StringPtr(recoveryId),
	}

	result, _, err := t.brsClient.GetBRSClient().GetRecoveryByIDWithContext(ctx, getRecoveryByIdOptions)
	if err != nil {
		t.logger.Error(ctx, "Failed to get restore", logger.Err(err), "restoreID", recoveryId, "operation", types.OpGetRestore)
		t.recordOperationFailure(ctx, metrics.OperationGetRestore, start)
		return nil, errors.NewSDKError(errors.ErrCodeUnknown, "Failed to fetch recovery by id", err)
	}

	// Use the common builder function to create RestoreResult
	restoreResult, buildErr := t.buildRestoreResult(ctx, result)
	if buildErr != nil {
		t.logger.Error(ctx, "Failed to build restore result", logger.Err(buildErr), "restoreID", recoveryId, "operation", types.OpGetRestore)
		t.recordOperationFailure(ctx, metrics.OperationGetRestore, start)
		return nil, buildErr
	}

	t.recordOperationSuccess(ctx, metrics.OperationGetRestore, start)
	t.logger.Info(ctx, "Restore retrieved successfully", "restoreID", recoveryId, "status", restoreResult.Status, "operation", types.OpGetRestore)
	return restoreResult, nil
}

// ListRestores lists restores for a registration
func (t *DefaultTaskAPI) ListRestores(ctx context.Context, registrationID int64) ([]*types.RestoreResult, *errors.SDKError) {
	start := time.Now()
	t.logger.Info(ctx, "Listing restores", "registrationID", registrationID, "operation", types.OpListRestores)

	getRecoveriesOptions := &backuprecoveryv1.GetRecoveriesOptions{
		XIBMTenantID: core.StringPtr(t.brsClient.GetTenantId()),
	}

	result, _, err := t.brsClient.GetBRSClient().GetRecoveriesWithContext(ctx, getRecoveriesOptions)
	if err != nil {
		t.logger.Error(ctx, "Failed to list restores", logger.Err(err), "registrationID", registrationID, "operation", types.OpListRestores)
		t.recordOperationFailure(ctx, metrics.OperationListRestores, start)
		return nil, errors.NewSDKError(errors.ErrCodeUnknown, "Failed to fetch recovery by id", err)
	}

	recoveryList := make([]*types.RestoreResult, 0)
	for _, recovery := range result.Recoveries {
		// Use the common builder function to create RestoreResult
		restoreResult, buildErr := t.buildRestoreResult(ctx, &recovery)
		if buildErr != nil {
			// Log error but continue processing other recoveries
			t.logger.Warn(ctx, "Failed to build restore result, skipping", logger.Err(buildErr), "recoveryID", *recovery.ID)
			continue
		}
		recoveryList = append(recoveryList, restoreResult)
	}

	t.recordOperationSuccess(ctx, metrics.OperationListRestores, start)
	t.logger.Info(ctx, "Restores listed successfully", "registrationID", registrationID, "count", len(recoveryList), "operation", types.OpListRestores)
	return recoveryList, nil
}

// WaitForRestore waits for restore to complete
func (t *DefaultTaskAPI) WaitForRestore(ctx context.Context, restoreID string, timeout time.Duration) (*types.RestoreResult, *errors.SDKError) {
	t.logger.Debug(ctx, "Waiting for restore completion", "restoreID", restoreID, "timeout", timeout, "operation", types.OpWaitForRestore)
	// TODO: Implement - poll BRS until restore completes
	t.logger.Warn(ctx, "WaitForRestore not yet implemented", "restoreID", restoreID)
	return nil, errors.NewSDKError(errors.ErrCodeUnknown, "WaitForRestore not yet implemented", nil)
}

// PauseBackup pauses a running backup
func (t *DefaultTaskAPI) PauseBackup(ctx context.Context, backupID, groupID string, force bool) *errors.SDKError {
	startedAt := time.Now()
	t.logger.Info(ctx, "Pausing backup", "backupID", backupID, "groupID", groupID, "force", force, "operation", types.OpPauseBackup)

	params := backuprecoveryv1.PauseProtectionRunActionParams{
		RunID: core.StringPtr(backupID),
	}

	options := backuprecoveryv1.PerformActionOnProtectionGroupRunOptions{
		ID:           core.StringPtr(groupID),
		XIBMTenantID: core.StringPtr(t.brsClient.GetTenantId()),
		Action:       core.StringPtr("Pause"),
		PauseParams:  []backuprecoveryv1.PauseProtectionRunActionParams{params},
	}

	_, _, err := t.brsClient.GetBRSClient().PerformActionOnProtectionGroupRunWithContext(ctx, &options)
	if err != nil {
		t.logger.Error(ctx, "Failed to pause backup", logger.Err(err), "backupID", backupID, "groupID", groupID, "operation", types.OpPauseBackup)
		t.recordOperationFailure(ctx, metrics.OperationPauseBackup, startedAt)
		return errors.NewProtectionFailedError("PauseBackup Failed", err)
	}

	t.logger.Info(ctx, "Backup paused successfully", "backupID", backupID, "groupID", groupID, "operation", types.OpPauseBackup)
	t.recordOperationSuccess(ctx, metrics.OperationPauseBackup, startedAt)
	return nil
}

// ResumeBackup resumes a paused backup
func (t *DefaultTaskAPI) ResumeBackup(ctx context.Context, backupID, groupID string) *errors.SDKError {
	startedAt := time.Now()
	t.logger.Info(ctx, "Resuming backup", "backupID", backupID, "groupID", groupID, "operation", types.OpResumeBackup)

	resumeParams := backuprecoveryv1.ResumeProtectionRunActionParams{
		RunID: core.StringPtr(backupID),
	}

	options := backuprecoveryv1.PerformActionOnProtectionGroupRunOptions{
		ID:           core.StringPtr(groupID),
		XIBMTenantID: core.StringPtr(t.brsClient.GetTenantId()),
		Action:       core.StringPtr("Resume"),
		ResumeParams: []backuprecoveryv1.ResumeProtectionRunActionParams{resumeParams},
	}

	_, _, err := t.brsClient.GetBRSClient().PerformActionOnProtectionGroupRunWithContext(ctx, &options)
	if err != nil {
		t.logger.Error(ctx, "Failed to resume backup", logger.Err(err), "backupID", backupID, "groupID", groupID, "operation", types.OpResumeBackup)
		t.recordOperationFailure(ctx, metrics.OperationResumeBackup, startedAt)
		return errors.NewProtectionFailedError("ResumeBackup Failed", err)
	}

	t.logger.Info(ctx, "Backup resumed successfully", "backupID", backupID, "groupID", groupID, "operation", types.OpResumeBackup)
	t.recordOperationSuccess(ctx, metrics.OperationResumeBackup, startedAt)
	return nil
}

// AbortBackup aborts a running backup
func (t *DefaultTaskAPI) AbortBackup(ctx context.Context, backupID, groupID string, force bool) *errors.SDKError {
	startedAt := time.Now()
	t.logger.Info(ctx, "Aborting backup", "backupID", backupID, "groupID", groupID, "force", force, "operation", types.OpAbortBackup)

	runs, err := t.GetBackupRunDetails(ctx, backupID, groupID)
	if err != nil {
		t.logger.Error(ctx, "Failed to get backup runs", logger.Err(err), "backupID", backupID, "groupID", groupID, "operation", types.OpAbortBackup)
		t.recordOperationFailure(ctx, metrics.OperationAbortBackup, startedAt)
		return errors.NewProtectionFailedError("GetBackupRuns failed", err)
	}

	if len(runs) == 0 {
		t.logger.Warn(ctx, "No backup runs found", "backupID", backupID, "groupID", groupID, "operation", types.OpAbortBackup)
		t.recordOperationFailure(ctx, metrics.OperationAbortBackup, startedAt)
		return errors.NewProtectionFailedError("No backup runs found", nil)
	}

	run := findRunningRun(runs)
	if run == nil {
		t.logger.Warn(ctx, "No running runs found", "backupID", backupID, "groupID", groupID, "operation", types.OpAbortBackup)
		t.recordOperationFailure(ctx, metrics.OperationAbortBackup, startedAt)
		return errors.NewProtectionFailedError("No running runs found", nil)
	}

	repIDs := extractReplicationTaskIDs(run)
	arcIDs := extractArchivalTaskIDs(run)

	t.logger.Debug(ctx, "Canceling backup run", "backupID", backupID, "groupID", groupID, "replicationTasks", len(repIDs), "archivalTasks", len(arcIDs))

	cancelReq := backuprecoveryv1.CancelProtectionGroupRunRequest{
		RunID:             run.ID,
		ReplicationTaskID: repIDs,
		ArchivalTaskID:    arcIDs,
	}

	options := backuprecoveryv1.PerformActionOnProtectionGroupRunOptions{
		ID:           core.StringPtr(groupID),
		XIBMTenantID: core.StringPtr(t.brsClient.GetTenantId()),
		Action:       core.StringPtr("Cancel"),
		CancelParams: []backuprecoveryv1.CancelProtectionGroupRunRequest{cancelReq},
	}

	_, _, apiErr := t.brsClient.GetBRSClient().PerformActionOnProtectionGroupRunWithContext(ctx, &options)
	if apiErr != nil {
		t.logger.Error(ctx, "Failed to abort backup", logger.Err(apiErr), "backupID", backupID, "groupID", groupID, "operation", types.OpAbortBackup)
		t.recordOperationFailure(ctx, metrics.OperationAbortBackup, startedAt)
		return errors.NewProtectionFailedError("AbortBackup Failed", apiErr)
	}

	t.logger.Info(ctx, "Backup aborted successfully", "backupID", backupID, "groupID", groupID, "operation", types.OpAbortBackup)
	t.recordOperationSuccess(ctx, metrics.OperationAbortBackup, startedAt)
	return nil
}

// GetBackupRunDetails retrieves backup run details
func (t *DefaultTaskAPI) GetBackupRunDetails(ctx context.Context, backupID, groupID string) ([]backuprecoveryv1.ProtectionGroupRun, error) {
	t.logger.Debug(ctx, "Getting backup run details", "backupID", backupID, "groupID", groupID)

	options := backuprecoveryv1.GetProtectionGroupRunsOptions{
		XIBMTenantID: core.StringPtr(t.brsClient.GetTenantId()),
		RunID:        core.StringPtr(backupID),
		ID:           core.StringPtr(groupID),
	}

	resp, _, err := t.brsClient.GetBRSClient().GetProtectionGroupRunsWithContext(ctx, &options)
	if err != nil {
		t.logger.Error(ctx, "Failed to get protection group runs", logger.Err(err), "backupID", backupID, "groupID", groupID)
		return nil, err
	}

	if resp == nil || len(resp.Runs) == 0 {
		t.logger.Debug(ctx, "No runs found", "backupID", backupID, "groupID", groupID)
		return []backuprecoveryv1.ProtectionGroupRun{}, nil
	}

	t.logger.Debug(ctx, "Retrieved backup runs", "backupID", backupID, "groupID", groupID, "count", len(resp.Runs))
	return resp.Runs, nil
}

// extractReplicationTaskIDs extracts replication task IDs from a run
func extractReplicationTaskIDs(run *backuprecoveryv1.ProtectionGroupRun) []string {
	var ids []string
	if run.ReplicationInfo != nil {
		for _, v := range run.ReplicationInfo.ReplicationTargetResults {
			if v.ReplicationTaskID != nil {
				ids = append(ids, *v.ReplicationTaskID)
			}
		}
	}
	return ids
}

// extractArchivalTaskIDs extracts archival task IDs from a run
func extractArchivalTaskIDs(run *backuprecoveryv1.ProtectionGroupRun) []string {
	var ids []string
	if run.ArchivalInfo != nil {
		for _, v := range run.ArchivalInfo.ArchivalTargetResults {
			if v.ArchivalTaskID != nil {
				ids = append(ids, *v.ArchivalTaskID)
			}
		}
	}
	return ids
}

// findRunningRun finds a running backup run
func findRunningRun(runs []backuprecoveryv1.ProtectionGroupRun) *backuprecoveryv1.ProtectionGroupRun {
	for _, run := range runs {
		if run.ArchivalInfo != nil &&
			len(run.ArchivalInfo.ArchivalTargetResults) > 0 &&
			run.ArchivalInfo.ArchivalTargetResults[0].Status != nil &&
			*run.ArchivalInfo.ArchivalTargetResults[0].Status == "Running" {
			return &run
		}
	}
	return nil
}

// AbortRestore aborts a running restore
func (t *DefaultTaskAPI) AbortRestore(ctx context.Context, restoreID string, force bool) *errors.SDKError {
	startedAt := time.Now()
	t.logger.Info(ctx, "Aborting restore", "restoreID", restoreID, "force", force, "operation", types.OpAbortRestore)
	t.emitActivityEvent(ctx, activitytracker.BuildAbortRestoreActivityEvent(restoreID, force, activitytracker.EventStateStarted, "starting restore abort", nil, 1, startedAt, t.getSourceIP(ctx)))

	getRecoveryByIdOptions := &backuprecoveryv1.CancelRecoveryByIdOptions{
		XIBMTenantID: core.StringPtr(t.brsClient.GetTenantId()),
		ID:           core.StringPtr(restoreID),
	}

	_, err := t.brsClient.GetBRSClient().CancelRecoveryByIDWithContext(ctx, getRecoveryByIdOptions)
	if err != nil {
		t.logger.Error(ctx, "Failed to abort restore", logger.Err(err), "restoreID", restoreID, "operation", types.OpAbortRestore)
		sdkErr := errors.NewSDKError(errors.ErrCodeUnknown, "Failed to cancel recovery by id", err)
		t.emitActivityEvent(ctx, activitytracker.BuildAbortRestoreActivityEvent(restoreID, force, activitytracker.EventStateFailed, "restore abort failed", sdkErr, 1, startedAt, t.getSourceIP(ctx)))
		t.recordOperationFailure(ctx, metrics.OperationAbortRestore, startedAt)
		return sdkErr
	}

	t.recordOperationSuccess(ctx, metrics.OperationAbortRestore, startedAt)
	t.logger.Info(ctx, "Restore aborted successfully", "restoreID", restoreID, "operation", types.OpAbortRestore)
	t.emitActivityEvent(ctx, activitytracker.BuildAbortRestoreActivityEvent(restoreID, force, activitytracker.EventStateSucceeded, "restore aborted successfully", nil, 1, startedAt, t.getSourceIP(ctx)))
	return nil
}

// buildRestoreResult is a helper function that builds a RestoreResult from a Recovery object
// It populates all fields including status, timestamps, and progress
func (t *DefaultTaskAPI) buildRestoreResult(ctx context.Context, recovery *backuprecoveryv1.Recovery) (*types.RestoreResult, *errors.SDKError) {
	if recovery == nil || recovery.ID == nil {
		return nil, errors.NewSDKError(errors.ErrCodeInvalidInput, "Recovery object or ID is nil", nil)
	}

	restoreResult := &types.RestoreResult{
		RestoreID: *recovery.ID,
	}

	// Populate status
	if recovery.Status != nil {
		restoreResult.Status = *recovery.Status
	}

	// Populate StartedAt from StartTimeUsecs (convert microseconds to time.Time)
	if recovery.StartTimeUsecs != nil {
		restoreResult.StartedAt = time.Unix(0, *recovery.StartTimeUsecs*1000)
	}

	// Populate CompletedAt from EndTimeUsecs (convert microseconds to time.Time)
	if recovery.EndTimeUsecs != nil {
		completedAt := time.Unix(0, *recovery.EndTimeUsecs*1000)
		restoreResult.CompletedAt = &completedAt
	}

	// Populate progress field using ProgressTaskID from KubernetesParams.Objects[0]
	// The ProgressTaskID is nested inside KubernetesParams -> Objects[0] -> ProgressTaskID
	var progressTaskID string
	if recovery.KubernetesParams != nil &&
		len(recovery.KubernetesParams.Objects) > 0 &&
		recovery.KubernetesParams.Objects[0].ProgressTaskID != nil &&
		*recovery.KubernetesParams.Objects[0].ProgressTaskID != "" {
		progressTaskID = *recovery.KubernetesParams.Objects[0].ProgressTaskID
	}

	// Fetch progress if we have a valid ProgressTaskID
	if progressTaskID != "" {
		progress, progressErr := t.getRestoreProgress(ctx, progressTaskID)
		if progressErr != nil {
			// Log error but don't fail the entire operation
			// Progress is optional information
			t.logger.Warn(ctx, "Failed to fetch restore progress", "recoveryID", *recovery.ID, logger.Err(progressErr))
		} else if progress != nil {
			restoreResult.Progress = *progress
		}
	}

	return restoreResult, nil
}

// getRestoreProgress is a helper function that retrieves the progress percentage for a restore operation
// using the GetProgressMonitors API with the ProgressTaskID from GetRecoveryByID output
func (t *DefaultTaskAPI) getRestoreProgress(ctx context.Context, progressTaskID string) (*int, *errors.SDKError) {
	t.logger.Debug(ctx, "Fetching restore progress", "progressTaskID", progressTaskID, "operation", types.OpGetRestoreProgress)

	// Create options for GetProgressMonitors API
	getProgressMonitorsOptions := &backuprecoveryv1.GetProgressMonitorsOptions{
		XIBMTenantID: core.StringPtr(t.brsClient.GetTenantId()),
		TaskPathVec:  []string{progressTaskID},
	}

	// Call GetProgressMonitors API
	t.logger.Debug(ctx, "Calling BRS GetProgressMonitors API", "progressTaskID", progressTaskID)
	progressResult, _, err := t.brsClient.GetBRSClient().GetProgressMonitorsWithContext(ctx, getProgressMonitorsOptions)
	if err != nil {
		t.logger.Error(ctx, "Failed to fetch progress monitors", logger.Err(err), "progressTaskID", progressTaskID)
		return nil, errors.NewSDKError(errors.ErrCodeUnknown, "Failed to fetch progress monitors", err)
	}

	// Extract progress percentage from the result
	if progressResult != nil && len(progressResult.ResultGroupVec) > 0 {
		resultGroup := progressResult.ResultGroupVec[0]
		if len(resultGroup.TaskVec) > 0 {
			task := resultGroup.TaskVec[0]
			if task.Progress != nil && task.Progress.PercentFinished != nil {
				// Convert float32 percentage to int (0-100)
				progressInt := int(*task.Progress.PercentFinished)
				t.logger.Debug(ctx, "Restore progress retrieved successfully", "progressTaskID", progressTaskID, "progress", progressInt)
				return &progressInt, nil
			}
		}
	}
	t.logger.Warn(ctx, "No progress information available for recovery", "progressTaskID", progressTaskID)
	// No progress information available
	return nil, nil
}
