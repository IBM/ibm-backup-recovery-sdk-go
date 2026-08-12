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
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/IBM/ibm-backup-recovery-sdk-go/migrationv2/common/config"
)

// TestNewClient_NilConfig verifies that a nil config is caught before any
// internal allocation, producing a clear error and no panic.
func TestNewClient_NilConfig(t *testing.T) {
	assert.NotPanics(t, func() {
		client, err := NewClient(context.Background(), nil)
		assert.Nil(t, client)
		require.Error(t, err)
		assert.Equal(t, "config cannot be nil", err.Error())
	})
}

// TestNewClient_InvalidConfig is a table-driven test covering every mandatory
// field validation path in Config.Validate(). Each case must return an error
// (never a panic) and must NOT reach the network.
func TestNewClient_InvalidConfig(t *testing.T) {
	validCRN := "crn:v1:bluemix:public:backup-recovery:us-south:a/acc:inst::"

	tests := []struct {
		name          string
		cfg           *config.Config
		errorContains string
	}{
		{
			name:          "missing region",
			cfg:           &config.Config{APIKey: "key", ResourceGroupID: "rg", BRSInstanceCRN: validCRN},
			errorContains: "region is required",
		},
		{
			name:          "missing API key",
			cfg:           &config.Config{Region: "us-south", ResourceGroupID: "rg", BRSInstanceCRN: validCRN},
			errorContains: "API key is required",
		},
		{
			name:          "missing BRSInstanceCRN and BRSInstanceName",
			cfg:           &config.Config{Region: "us-south", APIKey: "key", ResourceGroupID: "rg"},
			errorContains: "one of BRSInstanceCRN or BRSInstanceName is required",
		},
		{
			name:          "missing resource group ID",
			cfg:           &config.Config{Region: "us-south", APIKey: "key", BRSInstanceCRN: validCRN},
			errorContains: "resource group ID is required",
		},
		{
			name: "invalid BRS CRN format",
			cfg: &config.Config{
				Region: "us-south", APIKey: "key", ResourceGroupID: "rg",
				BRSInstanceCRN: "not-a-crn",
			},
			errorContains: "BRS instance CRN is invalid",
		},
		{
			name: "EnableMetrics=true without PrometheusConfig",
			cfg: &config.Config{
				Region: "us-south", APIKey: "key", ResourceGroupID: "rg",
				BRSInstanceCRN: validCRN, EnableMetrics: true,
			},
			errorContains: "PrometheusConfig is required when EnableMetrics=true",
		},
		{
			name: "EnableActivityTracker=true without ActivityTrackerConfig",
			cfg: &config.Config{
				Region: "us-south", APIKey: "key", ResourceGroupID: "rg",
				BRSInstanceCRN: validCRN, EnableActivityTracker: true,
			},
			errorContains: "ActivityTrackerConfig is required when EnableActivityTracker=true",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.NotPanics(t, func() {
				client, err := NewClient(context.Background(), tt.cfg)
				assert.Nil(t, client)
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.errorContains)
			})
		})
	}
}

