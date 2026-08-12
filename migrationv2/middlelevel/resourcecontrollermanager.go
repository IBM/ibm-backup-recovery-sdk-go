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
	"net/url"
	"strings"

	"github.com/IBM/go-sdk-core/v5/core"
	"github.com/IBM/ibm-backup-recovery-sdk-go/migrationv2/common/config"
	"github.com/IBM/ibm-backup-recovery-sdk-go/migrationv2/common/logger"
	"github.com/IBM/ibm-backup-recovery-sdk-go/migrationv2/common/types"
	"github.com/IBM/platform-services-go-sdk/resourcecontrollerv2"
	resourcecontroller "github.com/IBM/platform-services-go-sdk/resourcecontrollerv2"
)

// ResourceControllerManager manages IBM Cloud resource controller operations
// for BRS (Backup Recovery Service) instances. It provides methods to list,
// retrieve, and manage BRS service instances through the IBM Cloud Resource Controller API.
type ResourceControllerManager struct {
	config                   *config.Config
	resourceControllerClient *types.ResourceControllerClientWrapper
	logger                   logger.Logger
}

// NewResourceControllerManager creates a new ResourceControllerManager instance
// with the provided configuration.
//
// Parameters:
//   - cfg: Configuration containing authentication and service settings
//   - log: Logger instance for structured logging
//
// Returns:
//   - *ResourceControllerManager: A new manager instance
func NewResourceControllerManager(cfg *config.Config, log logger.Logger) *ResourceControllerManager {
	if cfg == nil {
		return nil
	}
	return &ResourceControllerManager{
		config: cfg,
		logger: log,
	}
}

// GetClientWrapper returns the wrapped resource controller client.
// This method provides access to the underlying IBM Cloud Resource Controller client.
//
// Parameters:
//   - ctx: Context for the operation
//
// Returns:
//   - *types.ResourceControllerClientWrapper: The wrapped client instance, or nil if not initialized
func (m *ResourceControllerManager) GetClientWrapper(ctx context.Context) *types.ResourceControllerClientWrapper {
	if m == nil {
		return nil
	}
	return m.resourceControllerClient
}

// Initialize initializes the resource controller client with authentication.
// This method must be called before using any other methods that interact with IBM Cloud.
//
// Parameters:
//   - ctx: Context for the operation
//
// Returns:
//   - error: An error if initialization fails, nil otherwise
func (m *ResourceControllerManager) Initialize(ctx context.Context) error {
	m.logger.Info(ctx, "Initializing Resource Controller Manager", "operation", "Initialize")

	if m.config == nil {
		m.logger.Error(ctx, "Configuration is nil")
		return fmt.Errorf("config is nil")
	}

	m.logger.Debug(ctx, "Creating resource controller client with authentication")
	
	// Create resource controller client with authentication
	resourceControllerClient, err := resourcecontroller.NewResourceControllerV2(&resourcecontroller.ResourceControllerV2Options{
		Authenticator: m.config.GetAuth(),
	})
	if err != nil {
		m.logger.Error(ctx, "Failed to create resource controller client", logger.Err(err))
		return fmt.Errorf("failed to create resource controller client: %w", err)
	}

	m.resourceControllerClient = &types.ResourceControllerClientWrapper{
		Client: resourceControllerClient,
	}

	m.logger.Info(ctx, "Resource Controller Manager initialized successfully")
	return nil
}

// isBRSInstance checks if a resource instance is a BRS (Backup Recovery Service) instance
// by examining its CRN for BRS service identifiers.
//
// Parameters:
//   - instance: The resource instance to check
//
// Returns:
//   - bool: true if the instance is a BRS instance, false otherwise
func (m *ResourceControllerManager) GetClient() *resourcecontrollerv2.ResourceControllerV2 {
	if m.resourceControllerClient == nil {
		return nil
	}
	return m.resourceControllerClient.Client
}

// ListBRSInstances retrieves all active BRS (Backup Recovery Service) instances
// from IBM Cloud Resource Controller. It handles pagination automatically and
// filters results to include only BRS service instances.
//
// Parameters:
//   - ctx: Context for the operation
//
// Returns:
//   - []types.BRSInstance: A slice of BRS instances found
//   - error: An error if the operation fails, nil otherwise
func (m *ResourceControllerManager) ListBRSInstances(ctx context.Context) ([]types.BRSInstance, error) {
	if m.resourceControllerClient == nil || m.resourceControllerClient.Client == nil {
		return nil, fmt.Errorf("resource controller client is not initialized")
	}

	start := ""
	var instances []resourcecontroller.ResourceInstance

	// Paginate through all resource instances
	for {
		listOptions := &resourcecontroller.ListResourceInstancesOptions{
			Type:            core.StringPtr(ResourceTypeServiceInstance.String()),
			State:           core.StringPtr(ResourceStateActive.String()),
			ResourceGroupID: core.StringPtr(m.config.ResourceGroupID),
		}

		if start != "" {
			listOptions.Start = core.StringPtr(start)
		}

		result, _, err := m.resourceControllerClient.Client.ListResourceInstancesWithContext(ctx, listOptions)
		if err != nil {
			return nil, fmt.Errorf("%s: %w", "failed to list resource instances", err)
		}

		if result != nil {
			instances = append(instances, result.Resources...)
		}

		// Check if there are more pages
		if result == nil || result.NextURL == nil {
			break
		}

		start, err = ExtractStartToken(*result.NextURL)
		if err != nil {
			return nil, fmt.Errorf("failed to extract start token: %w", err)
		}
	}

	// Filter and build BRS instances
	var brsInstances []types.BRSInstance

	for _, instance := range instances {
		// Check if this is a BRS instance by examining the CRN
		if !isBRSInstance(instance) {
			continue
		}

		brsInstance := buildBRSInstance(instance)
		if brsInstance != nil {
			brsInstances = append(brsInstances, *brsInstance)
		}
	}

	return brsInstances, nil
}

// GetBRSInstanceByName retrieves a specific BRS instance by its name.
// The search is case-insensitive and handles pagination automatically.
//
// Parameters:
//   - ctx: Context for the operation
//   - name: The name of the BRS instance to find
//
// Returns:
//   - *types.BRSInstance: The BRS instance if found
//   - error: An error if the operation fails or instance is not found
func (m *ResourceControllerManager) GetBRSInstanceByName(ctx context.Context, name string) (*types.BRSInstance, error) {
	if name == "" {
		return nil, fmt.Errorf("name cannot be empty")
	}
	if m.resourceControllerClient == nil || m.resourceControllerClient.Client == nil {
		return nil, fmt.Errorf("resource controller client is not initialized")
	}

	listOptions := &resourcecontroller.ListResourceInstancesOptions{
		Type:            core.StringPtr(ResourceTypeServiceInstance.String()),
		State:           core.StringPtr(ResourceStateActive.String()),
		ResourceGroupID: core.StringPtr(m.config.ResourceGroupID),
		Name:            core.StringPtr(name),
	}

	result, _, err := m.resourceControllerClient.Client.ListResourceInstancesWithContext(ctx, listOptions)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", "failed to list resource instances", err)
	}

	if result == nil || result.RowsCount == nil || *result.RowsCount == 0 || len(result.Resources) == 0 {
		return nil, fmt.Errorf("No BRS instance found for name: %v", name)
	}

	instance := buildBRSInstance(result.Resources[0])

	if instance != nil {
		return instance, nil
	}

	return nil, fmt.Errorf("no BRS instance found with name: %s", name)
}

func (m *ResourceControllerManager) checkInstanceCRN(crn string, result *resourcecontroller.ResourceInstancesList) *types.BRSInstance {
	if result == nil {
		return nil
	}
	for _, instance := range result.Resources {
		// Skip instances with missing required fields
		if instance.Name == nil {
			continue
		}

		// Check if this is a BRS instance
		if !isBRSInstance(instance) {
			continue
		}

		// Case-insensitive name comparison
		if strings.EqualFold(*instance.CRN, crn) {
			return buildBRSInstance(instance)
		}
	}
	return nil
}

// GetBRSInstanceByCRN retrieves a specific BRS instance by its CRN (Cloud Resource Name).
// The search is case-insensitive and handles pagination automatically.
//
// Parameters:
//   - ctx: Context for the operation
//   - crn: The CRN of the BRS instance to find
//
// Returns:
//   - *types.BRSInstance: The BRS instance if found
//   - error: An error if the operation fails or instance is not found
func (m *ResourceControllerManager) GetBRSInstanceByCRN(ctx context.Context, crn string) (*types.BRSInstance, error) {
	if crn == "" {
		return nil, fmt.Errorf("crn cannot be empty")
	}
	if m.resourceControllerClient == nil || m.resourceControllerClient.Client == nil {
		return nil, fmt.Errorf("resource controller client is not initialized")
	}

	start := ""

	// Paginate through all resource instances
	for {
		listOptions := &resourcecontroller.ListResourceInstancesOptions{
			Type:            core.StringPtr(ResourceTypeServiceInstance.String()),
			State:           core.StringPtr(ResourceStateActive.String()),
			ResourceGroupID: core.StringPtr(m.config.ResourceGroupID),
		}

		if start != "" {
			listOptions.Start = core.StringPtr(start)
		}

		result, _, err := m.resourceControllerClient.Client.ListResourceInstancesWithContext(ctx, listOptions)
		if err != nil {
			return nil, fmt.Errorf("%s: %w", "failed to list resource instances", err)
		}

		instance := m.checkInstanceCRN(crn, result)
		if instance != nil {
			return instance, nil
		}
		// Check if there are more pages
		if result == nil || result.NextURL == nil {
			break
		}

		nextStart, err := ExtractStartToken(*result.NextURL)
		if err != nil {
			return nil, fmt.Errorf("%s: %w", "failed to extract start token", err)
		}
		start = nextStart
	}

	return nil, fmt.Errorf("no BRS instance found with CRN: %s", crn)
}

// ExtractStartToken extracts the pagination start token from a next URL.
// It handles both absolute and relative URLs.
//
// Parameters:
//   - nextURL: The next URL from pagination response
//
// Returns:
//   - string: The extracted start token
//   - error: An error if URL parsing fails
func ExtractStartToken(nextURL string) (string, error) {
	if nextURL == "" {
		return "", nil
	}

	// Convert relative URLs to absolute for parsing
	if strings.HasPrefix(nextURL, URLPathPrefix) {
		nextURL = BaseURL + nextURL
	}

	u, err := url.Parse(nextURL)
	if err != nil {
		return "", fmt.Errorf("failed to parse URL: %w", err)
	}

	return u.Query().Get(QueryParamStart), nil
}

// isBRSInstance checks if a resource instance is a BRS (Backup Recovery Service) instance
// by examining its CRN for BRS service identifiers.
//
// Parameters:
//   - instance: The resource instance to check
//
// Returns:
//   - bool: true if the instance is a BRS instance, false otherwise
func isBRSInstance(instance resourcecontroller.ResourceInstance) bool {
	if instance.CRN == nil {
		return false
	}

	crnLower := strings.ToLower(*instance.CRN)
	return strings.Contains(crnLower, CRNServiceBRSTest) ||
		strings.Contains(crnLower, CRNServiceBRS)
}

// buildBRSInstance constructs a BRSInstance from a resource controller ResourceInstance.
// It safely extracts all relevant fields with nil pointer checks.
//
// Parameters:
//   - instance: The resource controller instance to convert
//
// Returns:
//   - *types.BRSInstance: The constructed BRS instance, or nil if instance is invalid
func buildBRSInstance(instance resourcecontroller.ResourceInstance) *types.BRSInstance {
	brs := &types.BRSInstance{}

	// Extract basic instance information with nil checks
	if instance.GUID != nil {
		brs.ID = *instance.GUID
	}

	if instance.Name != nil {
		brs.Name = *instance.Name
	}

	if instance.CRN != nil {
		brs.CRN = *instance.CRN
	}

	if instance.RegionID != nil {
		brs.Region = *instance.RegionID
	}

	// Extract extension information with type assertions and nil checks
	if instance.Extensions != nil {
		// Extract tenant ID
		if tenantID, ok := instance.Extensions[ExtensionKeyTenantID].(string); ok {
			brs.TenantID = tenantID
		}

		// Extract endpoints
		if endpoints, ok := instance.Extensions[ExtensionKeyEndpoints].(map[string]interface{}); ok {
			if privateURL, ok := endpoints[ExtensionKeyEndpointPrivate].(string); ok {
				brs.PrivateURL = privateURL
			}
			if publicURL, ok := endpoints[ExtensionKeyEndpointPublic].(string); ok {
				brs.PublicURL = publicURL
			}
		}
	}

	return brs
}
