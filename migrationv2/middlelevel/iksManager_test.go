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
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/IBM/go-sdk-core/v5/core"
	"github.com/IBM/ibm-backup-recovery-sdk-go/migrationv2/common/logger"
	"github.com/IBM/ibm-backup-recovery-sdk-go/migrationv2/common/metrics"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

func testLogger() logger.Logger {
	return logger.New(logger.Config{
		Level:       "info",
		ServiceName: "test-iks-manager",
		Environment: "test",
	})
}

// iamTokenHandler is a minimal /identity/token stub that returns a bearer token
// so IamAuthenticator.GetToken() succeeds without hitting the real IAM endpoint.
func iamTokenHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]interface{}{
		"access_token":  "test-bearer-token",
		"refresh_token": "test-refresh-token",
		"token_type":    "Bearer",
		"expires_in":    3600,
	})
}

// newTestIksManager builds an IksManager whose BaseService URL and IAM endpoint
// both point at the supplied httptest server so no real network calls are made.
func newTestIksManagerWithServer(srv *httptest.Server) *IksManager {
	auth := &core.IamAuthenticator{
		ApiKey: "test-api-key",
		URL:    srv.URL + "/identity/token",
	}
	service, _ := core.NewBaseService(&core.ServiceOptions{
		URL:           srv.URL,
		Authenticator: auth,
	})
	return &IksManager{
		iksClient:     service,
		authenticator: auth,
		endpointType:  "public",
		logger:        testLogger(),
		metrics:       metrics.NewNoop(),
		accountID:     "test-account-id",
	}
}

// validKubeconfig is a minimal YAML kubeconfig that clientcmd can parse.
var validKubeconfig = []byte(`apiVersion: v1
kind: Config
clusters:
- cluster:
    server: https://127.0.0.1:6443
  name: test-cluster
contexts:
- context:
    cluster: test-cluster
    user: test-user
  name: test-context
current-context: test-context
users:
- name: test-user
  user:
    token: test-bearer-token
`)

// ---------------------------------------------------------------------------
// TestNewIKSClient
// ---------------------------------------------------------------------------

func TestNewIKSClient(t *testing.T) {
	log := testLogger()
	m := metrics.NewNoop()

	t.Run("Success with valid authenticator and endpoint", func(t *testing.T) {
		auth := &core.IamAuthenticator{ApiKey: "test-key"}
		mgr, err := NewIKSClient("https://containers.cloud.ibm.com", "public", auth, log, m, "acct-1")
		assert.NoError(t, err)
		assert.NotNil(t, mgr)
		assert.Equal(t, "public", mgr.endpointType)
		assert.Equal(t, "acct-1", mgr.accountID)
	})

	t.Run("Success with nil metrics falls back to noop", func(t *testing.T) {
		auth := &core.IamAuthenticator{ApiKey: "test-key"}
		mgr, err := NewIKSClient("https://containers.cloud.ibm.com", "private", auth, log, nil, "acct-2")
		assert.NoError(t, err)
		assert.NotNil(t, mgr)
		assert.NotNil(t, mgr.metrics)
	})

	t.Run("Failure with nil authenticator", func(t *testing.T) {
		mgr, err := NewIKSClient("https://containers.cloud.ibm.com", "public", nil, log, m, "acct-3")
		assert.Error(t, err)
		assert.Nil(t, mgr)
		assert.Contains(t, err.Error(), "failed to initialize kube client")
	})
}

// ---------------------------------------------------------------------------
// TestApplyRBACAndGetKubeconfig
// ---------------------------------------------------------------------------

func TestApplyRBACAndGetKubeconfig(t *testing.T) {
	tests := []struct {
		name        string
		handler     http.HandlerFunc
		expectErr   bool
		errContains string
		validateBody func(*testing.T, []byte)
	}{
		{
			name: "Success returns kubeconfig bytes",
			handler: func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
				_, _ = w.Write(validKubeconfig)
			},
			expectErr: false,
			validateBody: func(t *testing.T, b []byte) {
				assert.Equal(t, validKubeconfig, b)
			},
		},
		{
			name: "404 response yields not-found message",
			handler: func(w http.ResponseWriter, r *http.Request) {
				http.Error(w, "cluster not found", http.StatusNotFound)
			},
			expectErr:   true,
			errContains: "not found",
		},
		{
			name: "401 response yields authentication-failed message",
			handler: func(w http.ResponseWriter, r *http.Request) {
				http.Error(w, "Unauthorized", http.StatusUnauthorized)
			},
			expectErr:   true,
			errContains: "authentication failed",
		},
		{
			name: "403 response yields authentication-failed message",
			handler: func(w http.ResponseWriter, r *http.Request) {
				http.Error(w, "Forbidden", http.StatusForbidden)
			},
			expectErr:   true,
			errContains: "authentication failed",
		},
		{
			name: "500 response yields generic error with status code",
			handler: func(w http.ResponseWriter, r *http.Request) {
				http.Error(w, "internal server error", http.StatusInternalServerError)
			},
			expectErr:   true,
			errContains: "failed to access cluster",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mux := http.NewServeMux()
			mux.HandleFunc("/identity/token", iamTokenHandler)
			mux.HandleFunc("/v2/applyRBACAndGetKubeconfig", tt.handler)
			srv := httptest.NewServer(mux)
			defer srv.Close()

			mgr := newTestIksManagerWithServer(srv)
			result, err := mgr.ApplyRBACAndGetKubeconfig("my-cluster")

			if tt.expectErr {
				assert.Error(t, err)
				assert.Nil(t, result)
				assert.Contains(t, err.Error(), tt.errContains)
			} else {
				require.NoError(t, err)
				if tt.validateBody != nil {
					tt.validateBody(t, result)
				}
			}
		})
	}
}

// ---------------------------------------------------------------------------
// TestApplyRBACAndGetKubeconfig_RequestShape
// Verifies the HTTP request sent to the IKS endpoint has the correct shape.
// ---------------------------------------------------------------------------

func TestApplyRBACAndGetKubeconfig_RequestShape(t *testing.T) {
	var capturedReq *http.Request
	var capturedBody map[string]interface{}

	mux := http.NewServeMux()
	mux.HandleFunc("/identity/token", iamTokenHandler)
	mux.HandleFunc("/v2/applyRBACAndGetKubeconfig", func(w http.ResponseWriter, r *http.Request) {
		capturedReq = r
		_ = json.NewDecoder(r.Body).Decode(&capturedBody)
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write(validKubeconfig)
	})
	srv := httptest.NewServer(mux)
	defer srv.Close()

	mgr := newTestIksManagerWithServer(srv)
	mgr.endpointType = "private"

	_, err := mgr.ApplyRBACAndGetKubeconfig("verify-cluster")
	require.NoError(t, err)
	require.NotNil(t, capturedReq)

	// Method and path
	assert.Equal(t, http.MethodPost, capturedReq.Method)
	assert.Equal(t, "/v2/applyRBACAndGetKubeconfig", capturedReq.URL.Path)

	// Authorization header must be a Bearer token, never the raw API key
	authHeader := capturedReq.Header.Get("Authorization")
	assert.True(t, len(authHeader) > 7 && authHeader[:7] == "Bearer ", "Authorization must be Bearer token")
	assert.NotContains(t, authHeader, "test-api-key", "raw API key must not appear in Authorization header")

	// Body fields
	assert.Equal(t, "verify-cluster", capturedBody["cluster"])
	assert.Equal(t, true, capturedBody["admin"])
	assert.Equal(t, "yaml", capturedBody["format"])
	assert.Equal(t, "private", capturedBody["endpointType"])
}

// ---------------------------------------------------------------------------
// TestGetKubeApi
// ---------------------------------------------------------------------------

func TestGetKubeApi(t *testing.T) {
	t.Run("Success: valid kubeconfig produces clientset and restConfig", func(t *testing.T) {
		mux := http.NewServeMux()
		mux.HandleFunc("/identity/token", iamTokenHandler)
		mux.HandleFunc("/v2/applyRBACAndGetKubeconfig", func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write(validKubeconfig)
		})
		srv := httptest.NewServer(mux)
		defer srv.Close()

		mgr := newTestIksManagerWithServer(srv)
		clientSet, restConfig, err := mgr.GetKubeApi("test-cluster")

		assert.NoError(t, err)
		assert.NotNil(t, clientSet)
		assert.NotNil(t, restConfig)
		assert.Equal(t, "https://127.0.0.1:6443", restConfig.Host)
	})

	t.Run("Failure: kubeconfig fetch fails propagates error", func(t *testing.T) {
		mux := http.NewServeMux()
		mux.HandleFunc("/identity/token", iamTokenHandler)
		mux.HandleFunc("/v2/applyRBACAndGetKubeconfig", func(w http.ResponseWriter, r *http.Request) {
			http.Error(w, "cluster not found", http.StatusNotFound)
		})
		srv := httptest.NewServer(mux)
		defer srv.Close()

		mgr := newTestIksManagerWithServer(srv)
		clientSet, restConfig, err := mgr.GetKubeApi("bad-cluster")

		assert.Error(t, err)
		assert.Nil(t, clientSet)
		assert.Nil(t, restConfig)
		assert.Contains(t, err.Error(), "not found")
	})

	t.Run("Failure: invalid kubeconfig YAML fails REST config parsing", func(t *testing.T) {
		mux := http.NewServeMux()
		mux.HandleFunc("/identity/token", iamTokenHandler)
		mux.HandleFunc("/v2/applyRBACAndGetKubeconfig", func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte("not: valid: kubeconfig: yaml: !!!"))
		})
		srv := httptest.NewServer(mux)
		defer srv.Close()

		mgr := newTestIksManagerWithServer(srv)
		clientSet, restConfig, err := mgr.GetKubeApi("test-cluster")

		assert.Error(t, err)
		assert.Nil(t, clientSet)
		assert.Nil(t, restConfig)
		assert.Contains(t, err.Error(), "failed to load REST config")
	})
}
