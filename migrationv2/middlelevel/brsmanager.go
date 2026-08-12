/*******************************************************************************
 * IBM Confidential
 * OCO Source Materials
 * (C) Copyright IBM Corp. 2026 All Rights Reserved.
 * The source code for this program is not  published or otherwise divested of
 * its trade secrets, irrespective of what has been deposited with
 * the U.S. Copyright Office.
 ******************************************************************************/
package middlelevel

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/IBM/go-sdk-core/v5/core"
	"github.com/IBM/ibm-backup-recovery-sdk-go/backuprecoveryv1"
	"github.com/IBM/ibm-backup-recovery-sdk-go/migrationv2/common/config"
	"github.com/IBM/ibm-backup-recovery-sdk-go/migrationv2/common/errors"
	"github.com/IBM/ibm-backup-recovery-sdk-go/migrationv2/common/logger"
	"github.com/IBM/ibm-backup-recovery-sdk-go/migrationv2/common/metrics"
	"github.com/IBM/ibm-backup-recovery-sdk-go/migrationv2/common/types"
)

// BRSManager manages BRS instance lifecycle and datasource registration
type BRSManager struct {
	config                    *config.Config
	brsClient                 types.BRSClientWrapperInterface
	ResourceControllerManager *ResourceControllerManager
	instanceCRN               string
	isAutoCreated             bool
	TenantId                  string
	logger                    logger.Logger
	metrics                   metrics.Metrics
}

// NewBRSManager creates a new BRS manager
func NewBRSManager(cfg *config.Config, resourceControllerManager *ResourceControllerManager, log logger.Logger, m metrics.Metrics) *BRSManager {
	return &BRSManager{
		config:                    cfg,
		ResourceControllerManager: resourceControllerManager,
		logger:                    log,
		metrics:                   m,
	}
}

// Initialize initializes the BRS instance (creates if needed or uses existing)
func (m *BRSManager) Initialize(ctx context.Context) error {
	start := time.Now()
	m.logger.Info(ctx, "Initializing BRS Manager", "operation", "Initialize")

	var err error
	// Case 1: BRS Instance CRN is provided - use existing instance
	if m.config.BRSInstanceCRN != "" {
		m.logger.Info(ctx, "Using existing BRS instance by CRN", "crn", m.config.BRSInstanceCRN)
		err = m.useExistingInstance(ctx, m.config.BRSInstanceCRN)
	} else if m.config.BRSInstanceName != "" {
		// Case 2: BRS Instance Name is provided - check if exists, create if not
		m.logger.Info(ctx, "Getting or creating BRS instance by name", "instanceName", m.config.BRSInstanceName)
		err = m.getOrCreateInstance(ctx, m.config.BRSInstanceName)
	} else {
		// Case 3: Neither provided - create with auto-generated name
		autoName := fmt.Sprintf("brs-auto-%d", time.Now().Unix())
		m.logger.Info(ctx, "Creating BRS instance with auto-generated name", "instanceName", autoName, "autoCreated", true)
		m.isAutoCreated = true
		err = m.createNewInstance(ctx, autoName)
	}

	duration := time.Since(start)
	accountID := m.config.AccountID

	if err != nil {
		// Record failed initialization
		m.metrics.IncCounter(ctx, metrics.MetricOperationTotal,
			metrics.Label{Key: "accountId", Value: accountID},
			metrics.Label{Key: "operation", Value: metrics.OperationBRSInitialize},
			metrics.Label{Key: "status", Value: "failure"},
		)
		m.metrics.RecordDuration(ctx, metrics.MetricOperationDuration, duration,
			metrics.Label{Key: "accountId", Value: accountID},
			metrics.Label{Key: "operation", Value: metrics.OperationBRSInitialize},
		)
		return err
	}

	// Record successful initialization
	m.metrics.IncCounter(ctx, metrics.MetricOperationTotal,
		metrics.Label{Key: "accountId", Value: accountID},
		metrics.Label{Key: "operation", Value: metrics.OperationBRSInitialize},
		metrics.Label{Key: "status", Value: "success"},
	)
	m.metrics.RecordDuration(ctx, metrics.MetricOperationDuration, duration,
		metrics.Label{Key: "accountId", Value: accountID},
		metrics.Label{Key: "operation", Value: metrics.OperationBRSInitialize},
	)

	return nil
}

// useExistingInstance uses an existing BRS instance by CRN
func (m *BRSManager) useExistingInstance(ctx context.Context, crn string) error {
	m.logger.Debug(ctx, "Fetching BRS instance by CRN", "crn", crn)

	var brsInstance *types.BRSInstance
	brsInstance, err := m.ResourceControllerManager.GetBRSInstanceByCRN(ctx, crn)
	if err != nil {
		m.logger.Error(ctx, "Failed to fetch BRS instance", logger.Err(err), "crn", crn)
		return errors.NewInvalidConfigError(
			fmt.Sprintf("failed to fetch tenantId %s", crn),
			err,
		)
	}

	m.config.TenantId = ensureTrailingSlash(brsInstance.TenantID)
	m.logger.Debug(ctx, "Initializing BRS client", "tenantId", m.config.TenantId, "crn", crn)

	brsClient, err := m.initializeBRSClient(brsInstance)
	if err != nil {
		m.logger.Error(ctx, "Failed to initialize BRS client", logger.Err(err), "crn", crn)
		return errors.NewInvalidConfigError(
			fmt.Sprintf("failed to initialize BRS client with CRN %s", crn),
			err,
		)
	}

	m.brsClient = brsClient
	m.instanceCRN = crn
	m.isAutoCreated = false

	m.logger.Info(ctx, "Successfully initialized BRS Manager with existing instance",
		"crn", crn,
		"tenantId", m.config.TenantId,
		"autoCreated", false)

	return nil
}

// getOrCreateInstance gets existing instance or creates new one
func (m *BRSManager) getOrCreateInstance(ctx context.Context, name string) error {
	m.logger.Debug(ctx, "Fetching BRS instance by name", "instanceName", name)

	brsInstance, err := m.ResourceControllerManager.GetBRSInstanceByName(ctx, name)
	if err != nil {
		m.logger.Error(ctx, "Failed to fetch BRS instance", logger.Err(err), "instanceName", name)
		return errors.NewInvalidConfigError(
			fmt.Sprintf("failed to fetch tenantId %s", name),
			err,
		)
	}
	if brsInstance == nil {
		m.logger.Error(ctx, "BRS instance not found and auto-creation not implemented", "instanceName", name)
		return errors.NewInvalidConfigError(
			fmt.Sprintf("BRS instance '%s' not found. Please create the instance first or provide a valid BRSInstanceCRN", name),
			nil,
		)
	}

	m.config.TenantId = ensureTrailingSlash(brsInstance.TenantID)
	m.logger.Debug(ctx, "Initializing BRS client", "tenantId", m.config.TenantId, "instanceName", name)

	brsClient, err := m.initializeBRSClient(brsInstance)
	if err != nil {
		m.logger.Error(ctx, "Failed to initialize BRS client", logger.Err(err), "instanceName", name)
		return errors.NewInvalidConfigError(
			fmt.Sprintf("failed to initialize BRS client with CRN %s", name),
			err,
		)
	}

	m.brsClient = brsClient
	m.instanceCRN = brsInstance.CRN
	m.isAutoCreated = false

	m.logger.Info(ctx, "Successfully initialized BRS Manager",
		"instanceName", name,
		"crn", brsInstance.CRN,
		"tenantId", m.config.TenantId,
		"autoCreated", false)

	return nil
}

func ensureTrailingSlash(tenantID string) string {
	if len(tenantID) == 0 {
		return tenantID
	}

	if tenantID[len(tenantID)-1] == '/' {
		return tenantID
	}

	return tenantID + "/"
}

// createNewInstance creates a new BRS instance
func (m *BRSManager) createNewInstance(ctx context.Context, name string) error {
	// TODO: Implement actual BRS instance creation using low-level APIs
	// This is a skeleton implementation

	// Generate CRN for the new instance
	// TODO:  Get the CRN of BRS instance
	crn := "CRN:TODO"

	// Initialize BRS client
	// brsClient, err := m.initializeBRSClient()
	// if err != nil {
	// 	return errors.NewConnectionFailedError(
	// 		fmt.Sprintf("failed to create BRS instance %s", name),
	// 		err,
	// 	)
	// }

	// m.brsClient = brsClient
	m.instanceCRN = crn
	m.isAutoCreated = true

	return nil
}

// initializeBRSClient initializes the BRS client
func (m *BRSManager) initializeBRSClient(brsInstance *types.BRSInstance) (types.BRSClientWrapperInterface, error) {
	// TODO: Initialize actual BRS client with proper authentication
	// This is a skeleton implementation
	if brsInstance == nil {
		return nil, fmt.Errorf("Cannot initialize nil brsInstance")
	}

	endpoint := brsInstance.PublicURL
	if m.config.BRSEndpointType == config.BRSendpointType_Private {
		endpoint = brsInstance.PrivateURL
	}
	authenticator := &core.IamAuthenticator{
		URL:    m.config.IAMEndpoint,
		ApiKey: m.config.APIKey,
	}
	endpoint = fmt.Sprintf("https://%v/v2", endpoint)
	backupRecoveryClientOptions := &backuprecoveryv1.BackupRecoveryV1Options{
		Authenticator: authenticator,
		URL:           endpoint,
	}

	// Construct the service client.
	client, err := backuprecoveryv1.NewBackupRecoveryV1(backupRecoveryClientOptions)
	if err != nil {
		return nil, fmt.Errorf("unable to create BRS Client: %w", err)
	}

	return &types.BRSClientWrapper{
		Client:   client,
		Region:   m.config.Region,
		CRN:      brsInstance.CRN,
		TenantId: m.config.TenantId,
	}, nil
}

// GetBRSClient returns the BRS client wrapper
func (m *BRSManager) GetBRSClient() types.BRSClientWrapperInterface {
	return m.brsClient
}

// GetInstanceCRN returns the BRS instance CRN
func (m *BRSManager) GetInstanceCRN() string {
	return m.instanceCRN
}

// GetInstanceId returns the BRS instance Id extracted from a CRN.
// Returns an empty string if crn is empty or has fewer than 3 colon-separated segments.
func (m *BRSManager) GetInstanceId(crn string) string {
	ids := strings.Split(crn, ":")
	if len(ids) < 3 {
		return ""
	}
	return ids[len(ids)-3]
}

// IsAutoCreated returns whether the instance was auto-created
func (m *BRSManager) IsAutoCreated() bool {
	return m.isAutoCreated
}

// RegisterDataSource registers a datasource with the BRS instance
// This ensures all datasources go through the same BRS instance
func (m *BRSManager) RegisterDataSource(ctx context.Context, dataSourceName string, dataSourceType types.DataSourceType) error {
	if m.brsClient == nil {
		return errors.NewInvalidConfigError("BRS client not initialized", nil)
	}

	// TODO: Implement actual datasource registration, by using low-level APIs
	// This is where we ensure all datasources are registered with this BRS instance

	return nil
}

// Cleanup cleans up resources (deletes auto-created instance if needed)
func (m *BRSManager) Cleanup(ctx context.Context) error {
	if !m.isAutoCreated {
		// Don't delete user-provided instances
		return nil
	}

	// TODO: Implement cleanup of auto-created instance, by using low-level APIs
	return nil
}
