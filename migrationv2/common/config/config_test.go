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
	"testing"
	"time"

	activity_tracker "github.com/IBM/ibm-backup-recovery-sdk-go/migrationv2/common/activity-tracker"
	"github.com/IBM/ibm-backup-recovery-sdk-go/migrationv2/common/metrics"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// baseConfig returns a minimal valid config (metrics and activity tracking both off).
func baseConfig() *Config {
	return &Config{
		Region:          "us-south",
		APIKey:          "test-api-key",
		ResourceGroupID: "test-rg-id",
		BRSInstanceCRN:  "crn:v1:bluemix:public:backup-recovery:us-south:a/account123:instance-123::",
		Timeout:         30 * time.Second,
	}
}

func TestValidate_Metrics(t *testing.T) {
	tests := []struct {
		name          string
		setup         func(*Config)
		expectError   bool
		errorContains string
	}{
		{
			name:        "EnableMetrics=false, no PrometheusConfig → valid",
			setup:       func(cfg *Config) { cfg.EnableMetrics = false },
			expectError: false,
		},
		{
			name: "EnableMetrics=true, no PrometheusConfig → error",
			setup: func(cfg *Config) {
				cfg.EnableMetrics = true
			},
			expectError:   true,
			errorContains: "PrometheusConfig is required when EnableMetrics=true",
		},
		{
			name: "EnableMetrics=true, PrometheusConfig provided, no AccountID → error",
			setup: func(cfg *Config) {
				cfg.EnableMetrics = true
				cfg.PrometheusConfig = &metrics.PrometheusConfig{Namespace: "test"}
			},
			expectError:   true,
			errorContains: "AccountID is required when EnableMetrics=true",
		},
		{
			name: "EnableMetrics=true, PrometheusConfig provided, AccountID provided → valid",
			setup: func(cfg *Config) {
				cfg.EnableMetrics = true
				cfg.PrometheusConfig = &metrics.PrometheusConfig{Namespace: "test"}
				cfg.AccountID = "account-123"
			},
			expectError: false,
		},
		{
			name: "EnableMetrics=false, PrometheusConfig set, no AccountID → valid (flag off)",
			setup: func(cfg *Config) {
				cfg.EnableMetrics = false
				cfg.PrometheusConfig = &metrics.PrometheusConfig{Namespace: "test"}
			},
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := baseConfig()
			tt.setup(cfg)
			err := cfg.Validate()
			if tt.expectError {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errorContains)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestValidate_ActivityTracker(t *testing.T) {
	tests := []struct {
		name          string
		setup         func(*Config)
		expectError   bool
		errorContains string
	}{
		{
			name:        "EnableActivityTracker=false, no config → valid",
			setup:       func(cfg *Config) { cfg.EnableActivityTracker = false },
			expectError: false,
		},
		{
			name: "EnableActivityTracker=true, no ActivityTrackerConfig → error",
			setup: func(cfg *Config) {
				cfg.EnableActivityTracker = true
			},
			expectError:   true,
			errorContains: "ActivityTrackerConfig is required when EnableActivityTracker=true",
		},
		{
			name: "EnableActivityTracker=true, config provided, no AccountID → error",
			setup: func(cfg *Config) {
				cfg.EnableActivityTracker = true
				cfg.ActivityTrackerConfig = &activity_tracker.HTTPSinkConfig{
					IngestionEndpoint: "https://logs.example.com",
				}
			},
			expectError:   true,
			errorContains: "AccountID is required when EnableMetrics=true or EnableActivityTracker=true",
		},
		{
			name: "EnableActivityTracker=true, config provided, AccountID provided → valid",
			setup: func(cfg *Config) {
				cfg.EnableActivityTracker = true
				cfg.ActivityTrackerConfig = &activity_tracker.HTTPSinkConfig{
					IngestionEndpoint: "https://logs.example.com",
				}
				cfg.AccountID = "account-123"
			},
			expectError: false,
		},
		{
			name: "EnableActivityTracker=false, config set, no AccountID → valid (flag off)",
			setup: func(cfg *Config) {
				cfg.EnableActivityTracker = false
				cfg.ActivityTrackerConfig = &activity_tracker.HTTPSinkConfig{
					IngestionEndpoint: "https://logs.example.com",
				}
			},
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := baseConfig()
			tt.setup(cfg)
			err := cfg.Validate()
			if tt.expectError {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errorContains)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestGetMetrics(t *testing.T) {
	t.Run("No PrometheusConfig, EnableMetrics=false → NoOp returned", func(t *testing.T) {
		cfg := &Config{}
		m := cfg.GetMetrics()
		assert.NotNil(t, m)
		_, isNoop := m.(*metrics.NoopMetrics)
		assert.True(t, isNoop)
	})

	t.Run("PrometheusConfig provided, EnableMetrics=true → Prometheus returned", func(t *testing.T) {
		cfg := &Config{
			EnableMetrics:    true,
			PrometheusConfig: &metrics.PrometheusConfig{Namespace: "test"},
		}
		m := cfg.GetMetrics()
		assert.NotNil(t, m)
		_, isPrometheus := m.(*metrics.PrometheusMetrics)
		assert.True(t, isPrometheus)
	})

	t.Run("PrometheusConfig provided, EnableMetrics=false → NoOp (flag off)", func(t *testing.T) {
		cfg := &Config{
			EnableMetrics:    false,
			PrometheusConfig: &metrics.PrometheusConfig{Namespace: "test"},
		}
		m := cfg.GetMetrics()
		_, isNoop := m.(*metrics.NoopMetrics)
		assert.True(t, isNoop)
	})

	t.Run("Subsequent calls return the same instance", func(t *testing.T) {
		cfg := &Config{
			EnableMetrics:    true,
			PrometheusConfig: &metrics.PrometheusConfig{Namespace: "test2"},
		}
		first := cfg.GetMetrics()
		second := cfg.GetMetrics()
		assert.Same(t, first, second, "GetMetrics must return the same instance on repeated calls")
	})
}

func TestGetActivityTracker(t *testing.T) {
	t.Run("No config, EnableActivityTracker=false → NoOpSink returned", func(t *testing.T) {
		cfg := &Config{}
		sink := cfg.GetActivityTracker()
		assert.NotNil(t, sink)
		_, isNoop := sink.(activity_tracker.NoOpSink)
		assert.True(t, isNoop)
	})

	t.Run("Valid config, EnableActivityTracker=true → HTTPSink returned", func(t *testing.T) {
		cfg := &Config{
			EnableActivityTracker: true,
			ActivityTrackerConfig: &activity_tracker.HTTPSinkConfig{
				IngestionEndpoint: "https://logs.example.com",
			},
		}
		sink := cfg.GetActivityTracker()
		assert.NotNil(t, sink)
		_, isHTTP := sink.(*activity_tracker.HTTPSink)
		assert.True(t, isHTTP)
	})

	t.Run("Invalid config (empty endpoint), EnableActivityTracker=true → falls back to NoOpSink", func(t *testing.T) {
		cfg := &Config{
			EnableActivityTracker: true,
			ActivityTrackerConfig: &activity_tracker.HTTPSinkConfig{
				IngestionEndpoint: "", // invalid – triggers fallback
			},
		}
		sink := cfg.GetActivityTracker()
		assert.NotNil(t, sink)
		_, isNoop := sink.(activity_tracker.NoOpSink)
		assert.True(t, isNoop)
	})

	t.Run("Subsequent calls return the same instance", func(t *testing.T) {
		cfg := &Config{
			EnableActivityTracker: true,
			ActivityTrackerConfig: &activity_tracker.HTTPSinkConfig{
				IngestionEndpoint: "https://logs.example.com",
			},
		}
		first := cfg.GetActivityTracker()
		second := cfg.GetActivityTracker()
		assert.Equal(t, first, second, "GetActivityTracker must return the same instance on repeated calls")
	})
}

func TestConfigSetters(t *testing.T) {
	tests := []struct {
		name         string
		setterFunc   func(*Config)
		validateFunc func(*testing.T, *Config)
	}{
		{
			name: "WithMetricsEnabled sets EnableMetrics",
			setterFunc: func(cfg *Config) {
				cfg.WithMetricsEnabled(true)
			},
			validateFunc: func(t *testing.T, cfg *Config) {
				assert.True(t, cfg.EnableMetrics)
			},
		},
		{
			name: "WithPrometheusConfig sets PrometheusConfig",
			setterFunc: func(cfg *Config) {
				cfg.WithPrometheusConfig(&metrics.PrometheusConfig{Namespace: "test"})
			},
			validateFunc: func(t *testing.T, cfg *Config) {
				require.NotNil(t, cfg.PrometheusConfig)
				assert.Equal(t, "test", cfg.PrometheusConfig.Namespace)
			},
		},
		{
			name: "WithAccountID sets AccountID",
			setterFunc: func(cfg *Config) {
				cfg.WithAccountID("account-123")
			},
			validateFunc: func(t *testing.T, cfg *Config) {
				assert.Equal(t, "account-123", cfg.AccountID)
			},
		},
		{
			name: "WithActivityTrackerEnabled sets EnableActivityTracker",
			setterFunc: func(cfg *Config) {
				cfg.WithActivityTrackerEnabled(true)
			},
			validateFunc: func(t *testing.T, cfg *Config) {
				assert.True(t, cfg.EnableActivityTracker)
			},
		},
		{
			name: "WithActivityTrackerConfig sets ActivityTrackerConfig",
			setterFunc: func(cfg *Config) {
				cfg.WithActivityTrackerConfig(&activity_tracker.HTTPSinkConfig{
					IngestionEndpoint: "https://logs.example.com",
				})
			},
			validateFunc: func(t *testing.T, cfg *Config) {
				require.NotNil(t, cfg.ActivityTrackerConfig)
				assert.Equal(t, "https://logs.example.com", cfg.ActivityTrackerConfig.IngestionEndpoint)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := &Config{Region: "us-south", APIKey: "test-api-key"}
			tt.setterFunc(cfg)
			tt.validateFunc(t, cfg)
		})
	}
}
