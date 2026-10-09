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
	Region string
	APIKey string

	// Exactly one of BRSInstanceName or BRSInstanceCRN must be provided.
	BRSInstanceName string          // Looked up by name; the instance must already exist in IBM Cloud
	BRSInstanceCRN  string          // Looked up by CRN (preferred — unambiguous and faster)
	BRSEndpointType BRSEndpointType // defaults to public
	ResourceGroupID string

	// SDK configuration
	Timeout       time.Duration
	EnableLogging bool

	// API enablement flags (both optional)
	EnableWorkflowAPI bool // Enable workflow-based APIs (for CLI, migration-tool)
	EnableTaskAPI     bool // Enable task-based APIs (for IKS/ROKS e2e tests)

	// IAMEndpoint overrides the default IBM Cloud IAM endpoint (optional).
	// Example: "https://iam.test.cloud.ibm.com" for staging.
	IAMEndpoint string

	// ResourceControllerEndpoint overrides the default IBM Cloud Resource Controller endpoint (optional).
	// Example: "https://resource-controller.test.cloud.ibm.com" for staging.
	ResourceControllerEndpoint string

	// TenantId is the BRS tenant identifier.
	// WARNING: if provided manually and it does not match the BRS instance's
	// actual tenant ID, all API calls will fail with authorisation errors.
	TenantId string

	// Activity tracking configuration.
	// EnableActivityTracker=true requires ActivityTrackerConfig and AccountID to be set.
	EnableActivityTracker bool
	// ActivityTrackerConfig holds the parameters used to build the HTTP sink internally.
	// Mandatory when EnableActivityTracker=true.
	ActivityTrackerConfig *activity_tracker.HTTPSinkConfig

	// Logger configuration (three options):
	// Option 1: Provide a logger instance directly
	Logger logger.Logger
	// Option 2: Provide logger config and we'll create the logger
	LoggerConfig *logger.Config
	// Option 3: Enable default logger (will create with default settings)
	UseDefaultLogger bool

	// Metrics configuration.
	// EnableMetrics=true requires PrometheusConfig and AccountID to be set.
	EnableMetrics bool
	// PrometheusConfig holds the parameters used to build the Prometheus metrics instance internally.
	// Mandatory when EnableMetrics=true.
	PrometheusConfig *metrics.PrometheusConfig

	// AccountID for metrics and activity tracker labeling.
	// Mandatory when EnableMetrics=true or EnableActivityTracker=true.
	AccountID string

	// activityTrackerSink and metrics are the fully constructed instances.
	// They are built on first use by GetActivityTracker()/GetMetrics() and must
	// never be set directly by callers.
	activityTrackerSink activity_tracker.Sink
	metrics             metrics.Metrics
}

type BRSEndpointType string

const (
	BRSendpointType_Public  BRSEndpointType = "public"
	BRSendpointType_Private BRSEndpointType = "private"
)

// DefaultConfig returns a configuration with default values.
// Metrics and activity tracking are disabled by default; opt in via EnableMetrics/EnableActivityTracker.
func DefaultConfig() *Config {
	return &Config{
		Timeout:           30 * time.Minute,
		EnableLogging:     true,
		EnableWorkflowAPI: true, // Enabled by default
		EnableTaskAPI:     true, // Enabled by default
		UseDefaultLogger:  true, // Enabled by default
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

	// When EnableMetrics=true, PrometheusConfig is mandatory.
	if c.EnableMetrics && c.PrometheusConfig == nil {
		return errors.NewInvalidConfigError("PrometheusConfig is required when EnableMetrics=true", nil)
	}

	// When EnableActivityTracker=true, ActivityTrackerConfig is mandatory.
	if c.EnableActivityTracker && c.ActivityTrackerConfig == nil {
		return errors.NewInvalidConfigError("ActivityTrackerConfig is required when EnableActivityTracker=true", nil)
	}

	// AccountID is mandatory whenever metrics or activity tracking is enabled.
	if (c.EnableMetrics || c.EnableActivityTracker) && c.AccountID == "" {
		return errors.NewInvalidConfigError("AccountID is required when EnableMetrics=true or EnableActivityTracker=true", nil)
	}

	if c.BRSInstanceCRN != "" {
		if !IsValidIBMCRN(c.BRSInstanceCRN) {
			return errors.NewInvalidConfigError("BRS instance CRN is invalid", nil)
		}
	}

	// Exactly one of BRSInstanceCRN or BRSInstanceName is required.
	if c.BRSInstanceName == "" && c.BRSInstanceCRN == "" {
		return errors.NewInvalidConfigError(
			"one of BRSInstanceCRN or BRSInstanceName is required; the BRS instance must already exist in IBM Cloud",
			nil,
		)
	}

	if c.ResourceGroupID == "" {
		return errors.NewInvalidConfigError("resource group ID is required", nil)
	}
	if c.Timeout <= 0 {
		return errors.NewInvalidConfigError("timeout must be positive", nil)
	}
	return nil
}

// ToSDKConfig converts to types.SDKConfig.
// Caller (NewClient) guarantees all three initialize* methods have already run,
// so Logger, metrics, and activityTrackerSink are non-nil here.
func (c *Config) ToSDKConfig() *types.SDKConfig {
	return &types.SDKConfig{
		Region:              c.Region,
		APIKey:              c.APIKey,
		BRSInstanceName:     c.BRSInstanceName,
		ResourceGroupID:     c.ResourceGroupID,
		Timeout:             c.Timeout,
		ActivityTrackerSink: c.activityTrackerSink,
		Logger:              c.Logger,
		Metrics:             c.metrics,
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

// initializeMetrics builds a Prometheus metrics instance from PrometheusConfig
// when EnableMetrics=true, or assigns a NoOp instance so callers never receive nil.
func (c *Config) initializeMetrics() {
	if c.metrics != nil {
		return
	}
	if c.EnableMetrics && c.PrometheusConfig != nil {
		c.metrics = metrics.NewPrometheus(*c.PrometheusConfig)
		return
	}
	// EnableMetrics=false or no config: use NoOp so callers never receive nil.
	c.metrics = metrics.NewNoop()
}

// initializeActivityTracker builds an HTTPSink from ActivityTrackerConfig when
// EnableActivityTracker=true, or assigns a NoOpSink so callers never receive nil.
func (c *Config) initializeActivityTracker() {
	if c.activityTrackerSink != nil {
		return
	}
	if c.EnableActivityTracker && c.ActivityTrackerConfig != nil {
		// Reuse the config's own API key so the user never has to supply a
		// separate IAMAuthenticator inside ActivityTrackerConfig.
		sinkCfg := *c.ActivityTrackerConfig
		if sinkCfg.IAMAuthenticator == nil {
			if auth, ok := c.GetAuth().(*core.IamAuthenticator); ok {
				sinkCfg.IAMAuthenticator = auth
			}
		}
		sink, err := activity_tracker.NewHTTPSink(sinkCfg)
		if err == nil {
			c.activityTrackerSink = sink
			return
		}
		// If sink construction fails, fall through to NoOpSink so the SDK
		// remains functional; the error will surface when the first event is emitted.
	}
	// EnableActivityTracker=false or construction failed: NoOpSink so callers never receive nil.
	c.activityTrackerSink = activity_tracker.NoOpSink{}
}

// GetMetrics returns the fully initialised metrics instance, initializing it
// lazily on first call if not already set via SetMetrics.
func (c *Config) GetMetrics() metrics.Metrics {
	c.initializeMetrics()
	return c.metrics
}

// SetMetrics stores a pre-built metrics instance. Called by NewClient after
// building the instance itself so the value is available via ToSDKConfig.
func (c *Config) SetMetrics(m metrics.Metrics) {
	c.metrics = m
}

// GetActivityTracker returns the fully initialised activity tracker sink,
// initializing it lazily on first call if not already set via SetActivityTrackerSink.
func (c *Config) GetActivityTracker() activity_tracker.Sink {
	c.initializeActivityTracker()
	return c.activityTrackerSink
}

// SetActivityTrackerSink stores a pre-built activity tracker sink. Called by
// NewClient after building the instance itself so the value is available via ToSDKConfig.
func (c *Config) SetActivityTrackerSink(sink activity_tracker.Sink) {
	c.activityTrackerSink = sink
}

// ---------------------------------------------------------------------------
// Builder (With*) methods
// ---------------------------------------------------------------------------

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

// WithIAMEndpoint overrides the IAM endpoint (e.g. staging).
func (c *Config) WithIAMEndpoint(url string) *Config {
	c.IAMEndpoint = url
	return c
}

// WithResourceControllerEndpoint overrides the Resource Controller endpoint (e.g. staging).
func (c *Config) WithResourceControllerEndpoint(url string) *Config {
	c.ResourceControllerEndpoint = url
	return c
}

// WithTimeout sets the timeout
func (c *Config) WithTimeout(timeout time.Duration) *Config {
	c.Timeout = timeout
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

// WithMetricsEnabled enables or disables metrics collection.
// When true, a PrometheusConfig and AccountID must also be provided.
func (c *Config) WithMetricsEnabled(enabled bool) *Config {
	c.EnableMetrics = enabled
	return c
}

// WithPrometheusConfig sets the Prometheus configuration used to build the
// metrics instance internally. Must be paired with WithMetricsEnabled(true) and WithAccountID.
func (c *Config) WithPrometheusConfig(cfg *metrics.PrometheusConfig) *Config {
	c.PrometheusConfig = cfg
	return c
}

// WithActivityTrackerEnabled enables or disables activity tracking.
// When true, ActivityTrackerConfig and AccountID must also be provided.
func (c *Config) WithActivityTrackerEnabled(enabled bool) *Config {
	c.EnableActivityTracker = enabled
	return c
}

// WithActivityTrackerConfig sets the HTTP sink configuration used to build the
// activity tracker sink internally. Must be paired with WithActivityTrackerEnabled(true) and WithAccountID.
func (c *Config) WithActivityTrackerConfig(cfg *activity_tracker.HTTPSinkConfig) *Config {
	c.ActivityTrackerConfig = cfg
	return c
}

// WithAccountID sets the account ID.
// Required when EnableMetrics=true or EnableActivityTracker=true.
func (c *Config) WithAccountID(accountID string) *Config {
	c.AccountID = accountID
	return c
}

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

// GetAuth returns an IamAuthenticator for the configured API key.
// When IAMEndpoint is set it is propagated to the authenticator so every
// SDK client that uses this authenticator (Resource Controller, BRS, IKS, …)
// hits the correct IAM endpoint rather than the production default.
func (c *Config) GetAuth() core.Authenticator {
	if c.APIKey != "" {
		return &core.IamAuthenticator{
			ApiKey: c.APIKey,
			URL:    c.IAMEndpoint, // empty string is fine; SDK uses default when empty
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
