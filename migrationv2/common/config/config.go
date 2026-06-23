/*******************************************************************************
 * IBM Confidential
 * OCO Source Materials
 * (C) Copyright IBM Corp. 2026 All Rights Reserved.
 * The source code for this program is not  published or otherwise divested of
 * its trade secrets, irrespective of what has been deposited with
 * the U.S. Copyright Office.
 ******************************************************************************/
package config

import (
	"fmt"
	"strings"
	"time"

	"github.com/IBM/go-sdk-core/v5/core"
	activity_tracker "github.com/IBM/ibm-backup-recovery-sdk-go/migrationv2/common/activity-tracker"
	"github.com/IBM/ibm-backup-recovery-sdk-go/migrationv2/common/errors"
	"github.com/IBM/ibm-backup-recovery-sdk-go/migrationv2/common/logger"
	"github.com/IBM/ibm-backup-recovery-sdk-go/migrationv2/common/metrics"
	"github.com/IBM/ibm-backup-recovery-sdk-go/migrationv2/common/types"
)

// Config represents the SDK configuration
type Config struct {
	// IBM Cloud configuration
	Region          string
	APIKey          string
	BRSInstanceName string          // Optional: if empty, will be created automatically
	BRSInstanceCRN  string          // Optional: if provided, will use existing instance
	BRSEndpointType BRSEndpointType // defaults to public
	ResourceGroupID string
	Tags            []string

	// SDK configuration
	Timeout       time.Duration
	RetryAttempts int
	RetryDelay    time.Duration
	EnableLogging bool

	// API enablement flags (both optional)
	EnableWorkflowAPI bool // Enable workflow-based APIs (for CLI, migration-tool)
	EnableTaskAPI     bool // Enable task-based APIs (for IKS/ROKS e2e tests)

	// Microservice mode configuration
	// IMPORTANT: For microservice deployments with multiple instances:
	// - MUST provide BRSInstanceCRN (shared across all instances)
	MicroserviceMode bool // Enable microservice-safe mode (validates required config)

	// Optional custom settings
	CustomSettings map[string]interface{}
	IAMEndpoint    string
	TenantId       string

	// Activity tracking configuration
	ActivityTrackerSink activity_tracker.Sink

	// Logger configuration (three options):
	// Option 1: Provide a logger instance directly
	Logger logger.Logger
	// Option 2: Provide logger config and we'll create the logger
	LoggerConfig *logger.Config
	// Option 3: Enable default logger (will create with default settings)
	UseDefaultLogger bool

	// Metrics configuration (three options):
	// Option 1: Provide a metrics instance directly
	Metrics metrics.Metrics
	// Option 2: Provide metrics config and we'll create the metrics
	MetricsConfig *metrics.PrometheusConfig
	// Option 3: Enable default metrics (will create with default settings)
	UseDefaultMetrics bool
	// Option 4: Disable metrics collection
	EnableMetrics bool

	// AccountID for metrics and activity tracker labeling
	// Required only if Metrics or ActivityTrackerSink is provided
	AccountID string
}

type BRSEndpointType string

const (
	BRSendpointType_Public  BRSEndpointType = "public"
	BRSendpointType_Private BRSEndpointType = "private"
)

// DefaultConfig returns a configuration with default values
// By default, both WorkflowAPI and TaskAPI are enabled, and default logger is enabled
func DefaultConfig() *Config {
	return &Config{
		Timeout:           30 * time.Minute,
		RetryAttempts:     3,
		RetryDelay:        5 * time.Second,
		EnableLogging:     true,
		EnableWorkflowAPI: true, // Enabled by default
		EnableTaskAPI:     true, // Enabled by default
		UseDefaultLogger:  true, // Enabled by default
		EnableMetrics:     true, // Enabled by default
		UseDefaultMetrics: true, // Enabled by default
		CustomSettings:    make(map[string]interface{}),
	}
}

// Validate validates the configuration
func (c *Config) Validate() error {
	if c.Region == "" {
		return errors.NewInvalidConfigError("region is required", nil)
	}

	if c.APIKey == "" {
		return errors.NewInvalidConfigError("API key is required", nil)
	}

	// Validate logger configuration only if logging is enabled
	if c.EnableLogging {
		if c.Logger == nil && c.LoggerConfig == nil && !c.UseDefaultLogger {
			return errors.NewInvalidConfigError("logger is required when EnableLogging=true: provide Logger, LoggerConfig, or set UseDefaultLogger=true", nil)
		}
	}

	// Validate AccountID if metrics or activity tracker are enabled
	c.initializeMetrics()
	metricsEnabled := c.Metrics != nil && c.Metrics != metrics.NewNoop()
	activityTrackerEnabled := c.ActivityTrackerSink != nil

	if (metricsEnabled || activityTrackerEnabled) && c.AccountID == "" {
		return errors.NewInvalidConfigError("AccountID is required when Metrics or ActivityTrackerSink is enabled", nil)
	}

	if c.BRSInstanceCRN != "" {
		if !IsValidIBMCRN(c.BRSInstanceCRN) {
			return errors.NewInvalidConfigError("BRS instance CRN is invalid", nil)
		}
	}

	// Microservice mode validation
	if c.MicroserviceMode {
		// In microservice mode, BRS instance CRN is REQUIRED
		if c.BRSInstanceCRN == "" {
			return errors.NewInvalidConfigError(
				"BRSInstanceCRN is required in microservice mode - all instances must share the same BRS instance",
				nil,
			)
		}
		// will not auto create
		if c.BRSInstanceName != "" && c.BRSInstanceCRN == "" {
			return errors.NewInvalidConfigError(
				"in microservice mode, use BRSInstanceCRN instead of BRSInstanceName to ensure all instances use the same BRS instance",
				nil,
			)
		}
	} else {
		// Single instance mode - auto-creation is allowed
		if c.BRSInstanceName == "" && c.BRSInstanceCRN == "" {
			// Will create a new instance with auto-generated name
			c.BRSInstanceName = fmt.Sprintf("brs-mig-sdk-auto-%d", time.Now().Unix())
		}
	}

	if c.ResourceGroupID == "" {
		return errors.NewInvalidConfigError("resource group ID is required", nil)
	}
	if c.Timeout <= 0 {
		return errors.NewInvalidConfigError("timeout must be positive", nil)
	}
	return nil
}

// ToSDKConfig converts to types.SDKConfig
func (c *Config) ToSDKConfig() *types.SDKConfig {
	// Ensure logger and metrics are initialized
	c.initializeLogger()
	c.initializeMetrics()

	return &types.SDKConfig{
		Region:              c.Region,
		APIKey:              c.APIKey,
		BRSInstanceName:     c.BRSInstanceName,
		ResourceGroupID:     c.ResourceGroupID,
		Tags:                c.Tags,
		Timeout:             c.Timeout,
		ActivityTrackerSink: c.ActivityTrackerSink,
		Logger:              c.Logger,
		Metrics:             c.Metrics,
		AccountID:           c.AccountID,
	}
}

// initializeLogger creates a logger if not already provided
func (c *Config) initializeLogger() {
	// If logger already exists, nothing to do
	if c.Logger != nil {
		return
	}

	// If logging is disabled, use NoOp logger
	if !c.EnableLogging {
		c.Logger = logger.NewNoOpLogger()
		return
	}

	// If LoggerConfig is provided, create logger from config
	if c.LoggerConfig != nil {
		c.Logger = logger.New(*c.LoggerConfig)
		return
	}

	// If UseDefaultLogger is true, create default logger
	if c.UseDefaultLogger {
		defaultLogConfig := logger.DefaultConfig()
		defaultLogConfig.ServiceName = "brs-migration-sdk"
		defaultLogConfig.Environment = "production"
		defaultLogConfig.Level = "info" // Default to info level
		c.Logger = logger.New(defaultLogConfig)
		return
	}

	// Fallback: if nothing is configured, use NoOp logger
	c.Logger = logger.NewNoOpLogger()
}

// GetLogger returns the logger instance, initializing it if necessary
func (c *Config) GetLogger() logger.Logger {
	c.initializeLogger()
	return c.Logger
}

// initializeMetrics creates a metrics instance if not already provided
func (c *Config) initializeMetrics() {
	// If metrics already exists, nothing to do
	if c.Metrics != nil {
		return
	}

	// If metrics is disabled, use NoOp metrics
	if !c.EnableMetrics {
		c.Metrics = metrics.NewNoop()
		return
	}

	// If MetricsConfig is provided, create metrics from config
	if c.MetricsConfig != nil {
		c.Metrics = metrics.NewPrometheus(*c.MetricsConfig)
		return
	}

	// If UseDefaultMetrics is true, create default metrics
	if c.UseDefaultMetrics {
		// Ensure logger is initialized first (metrics needs logger)
		c.initializeLogger()

		defaultMetricsConfig := metrics.PrometheusConfig{
			Namespace: "brs_migration_sdk",
			Registry:  nil, // Use default registry
			Logger:    c.Logger,
		}
		c.Metrics = metrics.NewPrometheus(defaultMetricsConfig)
		return
	}

	// Fallback: if nothing is configured, use NoOp metrics
	c.Metrics = metrics.NewNoop()
}

// GetMetrics returns the metrics instance, initializing it if necessary
func (c *Config) GetMetrics() metrics.Metrics {
	c.initializeMetrics()
	return c.Metrics
}

// WithRegion sets the region
func (c *Config) WithRegion(region string) *Config {
	c.Region = region
	return c
}

// WithAPIKey sets the API key
func (c *Config) WithAPIKey(apiKey string) *Config {
	c.APIKey = apiKey
	return c
}

// WithBRSInstanceName sets the BRS instance name (optional)
func (c *Config) WithBRSInstanceName(name string) *Config {
	c.BRSInstanceName = name
	return c
}

// WithBRSInstanceCRN sets the BRS instance CRN (optional, use existing instance)
func (c *Config) WithBRSInstanceCRN(crn string) *Config {
	c.BRSInstanceCRN = crn
	return c
}

// WithResourceGroupID sets the resource group ID
func (c *Config) WithResourceGroupID(id string) *Config {
	c.ResourceGroupID = id
	return c
}

// WithTags sets the tags
func (c *Config) WithTags(tags []string) *Config {
	c.Tags = tags
	return c
}

// WithTimeout sets the timeout
func (c *Config) WithTimeout(timeout time.Duration) *Config {
	c.Timeout = timeout
	return c
}

// WithRetryAttempts sets the retry attempts
func (c *Config) WithRetryAttempts(attempts int) *Config {
	c.RetryAttempts = attempts
	return c
}

// WithRetryDelay sets the retry delay
func (c *Config) WithRetryDelay(delay time.Duration) *Config {
	c.RetryDelay = delay
	return c
}

// WithLogging enables or disables logging
func (c *Config) WithLogging(enabled bool) *Config {
	c.EnableLogging = enabled
	return c
}

// WithLogger sets the logger instance directly
func (c *Config) WithLogger(log logger.Logger) *Config {
	c.Logger = log
	c.LoggerConfig = nil
	c.UseDefaultLogger = false
	return c
}

// WithLoggerConfig sets the logger configuration (logger will be created from this config)
func (c *Config) WithLoggerConfig(logConfig *logger.Config) *Config {
	c.LoggerConfig = logConfig
	c.Logger = nil
	c.UseDefaultLogger = false
	return c
}

// WithDefaultLogger enables the default logger
func (c *Config) WithDefaultLogger(enabled bool) *Config {
	c.UseDefaultLogger = enabled
	if enabled {
		c.Logger = nil
		c.LoggerConfig = nil
	}
	return c
}

// WithCustomSetting adds a custom setting
func (c *Config) WithCustomSetting(key string, value interface{}) *Config {
	c.CustomSettings[key] = value
	return c
}

// WithWorkflowAPI enables or disables workflow-based APIs
func (c *Config) WithWorkflowAPI(enabled bool) *Config {
	c.EnableWorkflowAPI = enabled
	return c
}

// WithTaskAPI enables or disables task-based APIs
func (c *Config) WithTaskAPI(enabled bool) *Config {
	c.EnableTaskAPI = enabled
	return c
}

// WithMicroserviceMode enables microservice-safe mode
// WithMetrics sets the metrics instance directly
func (c *Config) WithMetrics(m metrics.Metrics) *Config {
	c.Metrics = m
	c.MetricsConfig = nil
	c.UseDefaultMetrics = false
	return c
}

// WithMetricsConfig sets the metrics configuration (metrics will be created from this config)
func (c *Config) WithMetricsConfig(metricsConfig *metrics.PrometheusConfig) *Config {
	c.MetricsConfig = metricsConfig
	c.Metrics = nil
	c.UseDefaultMetrics = false
	return c
}

// WithDefaultMetrics enables the default metrics
func (c *Config) WithDefaultMetrics(enabled bool) *Config {
	c.UseDefaultMetrics = enabled
	if enabled {
		c.Metrics = nil
		c.MetricsConfig = nil
	}
	return c
}

// WithMetricsEnabled enables or disables metrics collection
func (c *Config) WithMetricsEnabled(enabled bool) *Config {
	c.EnableMetrics = enabled
	return c
}

// WithAccountID sets the account ID (required if metrics or activity tracker are enabled)
func (c *Config) WithAccountID(accountID string) *Config {
	c.AccountID = accountID
	return c
}

// This enforces that BRSInstanceCRN is provided and validates microservice requirements
func (c *Config) WithMicroserviceMode(enabled bool) *Config {
	c.MicroserviceMode = enabled
	return c
}

func (c *Config) GetAuth() core.Authenticator {
	if c.APIKey != "" {
		return &core.IamAuthenticator{
			ApiKey: c.APIKey,
		}
	}

	return nil
}

func IsValidIBMCRN(crn string) bool {
	if crn == "" {
		return false
	}

	parts := strings.Split(crn, ":")

	// IBM CRN can have trailing empty segments → allow >= 8
	if len(parts) < 8 {
		return false
	}

	// Mandatory fixed parts
	if parts[0] != "crn" {
		return false
	}

	if parts[1] != "v1" {
		return false
	}

	// Required non-empty fields
	if parts[2] == "" || // cloud name
		parts[3] == "" || // cloud type
		parts[4] == "" || // service name
		parts[6] == "" || // account id
		parts[7] == "" { // resource type
		return false
	}

	// Account ID must start with "a/"
	if !strings.HasPrefix(parts[6], "a/") {
		return false
	}

	return true
}

func DefaultIfEmpty(value string, defaultValue string) string {
	if value == "" {
		return defaultValue
	}

	return value
}
