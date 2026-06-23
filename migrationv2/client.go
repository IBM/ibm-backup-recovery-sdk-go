/*******************************************************************************
 * IBM Confidential
 * OCO Source Materials
 * (C) Copyright IBM Corp. 2026 All Rights Reserved.
 * The source code for this program is not  published or otherwise divested of
 * its trade secrets, irrespective of what has been deposited with
 * the U.S. Copyright Office.
 ******************************************************************************/
package migrationv2

import (
	"context"
	"fmt"

	activity_tracker "github.com/IBM/ibm-backup-recovery-sdk-go/migrationv2/common/activity-tracker"
	"github.com/IBM/ibm-backup-recovery-sdk-go/migrationv2/common/config"
	"github.com/IBM/ibm-backup-recovery-sdk-go/migrationv2/common/errors"
	"github.com/IBM/ibm-backup-recovery-sdk-go/migrationv2/middlelevel"
	"github.com/IBM/ibm-backup-recovery-sdk-go/migrationv2/middlelevel/connectors"
	"github.com/IBM/ibm-backup-recovery-sdk-go/migrationv2/middlelevel/datasources"
	"github.com/IBM/ibm-backup-recovery-sdk-go/migrationv2/toplevel/tasks"
	"github.com/IBM/ibm-backup-recovery-sdk-go/migrationv2/toplevel/workflow"
)

// Client is the main SDK client that manages BRS instance and provides access to all workflow and tasks APIs
// WorkflowAPI and TaskAPI are optional - they will only be initialized if enabled in config
// All APIs provide idempotent and async(implementation based) operations
type Client struct {
	config                    *config.Config
	brsManager                *middlelevel.BRSManager
	resourceControllerManager *middlelevel.ResourceControllerManager

	// API interfaces (optional - can be nil if not enabled)
	WorkflowAPI workflow.WorkflowAPI // nil if EnableWorkflowAPI is false
	TaskAPI     tasks.TaskAPI        // nil if EnableTaskAPI is false

	// Factory for preparing datasources
	DataSourceFactory datasources.DataSourceFactory
	ConnectorFactory  connectors.ConnectorFactory
}

// NewClient creates a new SDK client
// BRS instance is optional:
// - If BRSInstanceCRN is provided, it will use the existing instance
// - If BRSInstanceName is provided, it will get or create that instance
// - If neither is provided, it will auto-create an instance with generated name
func NewClient(ctx context.Context, cfg *config.Config) (*Client, error) {
	// Validate configuration (includes logger validation)
	if err := cfg.Validate(); err != nil {
		return nil, err
	}

	// Get logger and metrics from config (will initialize if needed)
	log := cfg.GetLogger()
	metricsCollector := cfg.GetMetrics()

	// Create resource controller manager with logger
	resourceControllerManager := middlelevel.NewResourceControllerManager(cfg, log)

	// Initialize resource controller instance (create)
	if err := resourceControllerManager.Initialize(ctx); err != nil {
		return nil, fmt.Errorf(
			"failed to initialize Resource Controller instance: %v",
			err,
		)
	}

	// Create BRS manager with logger and metrics
	brsManager := middlelevel.NewBRSManager(cfg, resourceControllerManager, log, metricsCollector)

	// Initialize BRS instance (create if needed or use existing)
	if err := brsManager.Initialize(ctx); err != nil {
		return nil, errors.NewConnectionFailedError(
			"failed to initialize BRS instance",
			err,
		)
	}

	// Get BRS client wrapper
	brsClient := brsManager.GetBRSClient()

	sdkConfig := cfg.ToSDKConfig()
	if sdkConfig.ActivityTrackerSink == nil {
		sdkConfig.ActivityTrackerSink = activity_tracker.NoOpSink{}
	}

	// Create client
	client := &Client{
		config:                    cfg,
		brsManager:                brsManager,
		resourceControllerManager: resourceControllerManager,
		DataSourceFactory:         datasources.NewDataSourceFactory(),
		ConnectorFactory:          connectors.NewConnectorFactory(),
	}

	// Initialize WorkflowAPI only if enabled
	if cfg.EnableWorkflowAPI {
		client.WorkflowAPI = workflow.NewWorkflowAPI(brsClient, sdkConfig)
	}

	// Initialize TaskAPI only if enabled
	if cfg.EnableTaskAPI {
		client.TaskAPI = tasks.NewTaskAPI(brsClient, sdkConfig)
	}

	return client, nil
}

func (c *Client) CreateDataSource(config datasources.DataSourceConfig) (datasources.DataSource, error) {
	return c.DataSourceFactory.CreateDataSource(config, c.brsManager.GetBRSClient())
}

// RegisterDataSource registers a datasource with the BRS instance
// This ensures all datasources (source and target) are registered with the same BRS instance
func (c *Client) RegisterDataSource(ctx context.Context, dataSource datasources.DataSource) error {
	if dataSource == nil {
		return errors.NewInvalidDataSourceError("datasource cannot be nil")
	}

	// Validate datasource
	if err := dataSource.Validate(); err != nil {
		return errors.NewInvalidDataSourceError(
			fmt.Sprintf("datasource validation failed: %v", err),
		)
	}

	// Register with BRS manager to track
	if err := c.brsManager.RegisterDataSource(ctx, dataSource.GetName(), dataSource.GetType()); err != nil {
		return err
	}

	return nil
}

// GetBRSInstanceCRN returns the CRN of the BRS instance being used
func (c *Client) GetBRSInstanceCRN() string {
	return c.brsManager.GetInstanceCRN()
}

// IsAutoCreatedInstance returns whether the BRS instance was auto-created
func (c *Client) IsAutoCreatedInstance() bool {
	return c.brsManager.IsAutoCreated()
}

// Close cleans up resources
// If the BRS instance was auto-created, it will be deleted
func (c *Client) Close(ctx context.Context) error {
	return c.brsManager.Cleanup(ctx)
}

// GetBRSClient returns the BRS client wrapper for advanced usage
func (c *Client) GetBRSClient() interface{} {
	return c.brsManager.GetBRSClient()
}
