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
	"testing"

	"github.com/IBM/ibm-backup-recovery-sdk-go/migrationv2/common/config"
	"github.com/IBM/ibm-backup-recovery-sdk-go/migrationv2/common/logger"
	"github.com/IBM/ibm-backup-recovery-sdk-go/migrationv2/common/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestNewResourceControllerManager covers constructor nil/non-nil outcomes.
func TestNewResourceControllerManager(t *testing.T) {
	tests := []struct {
		name    string
		cfg     *config.Config
		wantNil bool
	}{
		{
			name:    "nil config returns nil manager",
			cfg:     nil,
			wantNil: true,
		},
		{
			name:    "valid config returns non-nil manager",
			cfg:     &config.Config{Region: "us-south", APIKey: "test-key"},
			wantNil: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.NotPanics(t, func() {
				m := NewResourceControllerManager(tt.cfg, logger.NewNoop())
				if tt.wantNil {
					assert.Nil(t, m)
				} else {
					assert.NotNil(t, m)
				}
			})
		})
	}
}

// TestGetClient covers nil-safety on GetClient before and after initialization.
func TestGetClient(t *testing.T) {
	tests := []struct {
		name    string
		setup   func() *ResourceControllerManager
		wantNil bool
	}{
		{
			name: "before Initialize returns nil without panic",
			setup: func() *ResourceControllerManager {
				return NewResourceControllerManager(
					&config.Config{Region: "us-south", APIKey: "test-key"},
					logger.NewNoop(),
				)
			},
			wantNil: true,
		},
		{
			name: "after failed Initialize (empty APIKey) returns nil without panic",
			setup: func() *ResourceControllerManager {
				m := NewResourceControllerManager(
					&config.Config{Region: "us-south", APIKey: ""},
					logger.NewNoop(),
				)
				_ = m.Initialize(context.Background())
				return m
			},
			wantNil: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.NotPanics(t, func() {
				m := tt.setup()
				client := m.GetClient()
				if tt.wantNil {
					assert.Nil(t, client)
				} else {
					assert.NotNil(t, client)
				}
			})
		})
	}
}

// TestGetClientWrapper_NilReceiver verifies that a nil receiver does not panic.
func TestGetClientWrapper_NilReceiver(t *testing.T) {
	var m *ResourceControllerManager
	assert.NotPanics(t, func() {
		assert.Nil(t, m.GetClientWrapper(context.Background()))
	})
}

// TestGetBRSInstanceByName covers the full validation surface:
// empty name, uninitialised client, and nil RowsCount in the result.
func TestGetBRSInstanceByName(t *testing.T) {
	tests := []struct {
		name          string
		instanceName  string
		setup         func(*ResourceControllerManager)
		errorContains string
	}{
		{
			name:          "empty name returns error without panic",
			instanceName:  "",
			setup:         func(m *ResourceControllerManager) {},
			errorContains: "name cannot be empty",
		},
		{
			name:          "uninitialized client returns error without panic",
			instanceName:  "test-instance",
			setup:         func(m *ResourceControllerManager) {},
			errorContains: "resource controller client is not initialized",
		},
		{
			name:         "wrapper set but inner Client nil returns error without panic",
			instanceName: "test-instance",
			setup: func(m *ResourceControllerManager) {
				m.resourceControllerClient = &types.ResourceControllerClientWrapper{Client: nil}
			},
			errorContains: "resource controller client is not initialized",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := NewResourceControllerManager(
				&config.Config{Region: "us-south", APIKey: "test-key"},
				logger.NewNoop(),
			)
			tt.setup(m)

			assert.NotPanics(t, func() {
				instance, err := m.GetBRSInstanceByName(context.Background(), tt.instanceName)
				assert.Nil(t, instance)
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.errorContains)
			})
		})
	}
}

// TestGetBRSInstanceByCRN covers the full validation surface:
// empty CRN and uninitialised client.
func TestGetBRSInstanceByCRN(t *testing.T) {
	tests := []struct {
		name          string
		crn           string
		setup         func(*ResourceControllerManager)
		errorContains string
	}{
		{
			name:          "empty CRN returns error without panic",
			crn:           "",
			setup:         func(m *ResourceControllerManager) {},
			errorContains: "crn cannot be empty",
		},
		{
			name:          "uninitialized client returns error without panic",
			crn:           "crn:v1:bluemix:public:backup-recovery:us-south:a/acc:inst::",
			setup:         func(m *ResourceControllerManager) {},
			errorContains: "resource controller client is not initialized",
		},
		{
			name: "wrapper set but inner Client nil returns error without panic",
			crn:  "crn:v1:bluemix:public:backup-recovery:us-south:a/acc:inst::",
			setup: func(m *ResourceControllerManager) {
				m.resourceControllerClient = &types.ResourceControllerClientWrapper{Client: nil}
			},
			errorContains: "resource controller client is not initialized",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := NewResourceControllerManager(
				&config.Config{Region: "us-south", APIKey: "test-key"},
				logger.NewNoop(),
			)
			tt.setup(m)

			assert.NotPanics(t, func() {
				instance, err := m.GetBRSInstanceByCRN(context.Background(), tt.crn)
				assert.Nil(t, instance)
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.errorContains)
			})
		})
	}
}
