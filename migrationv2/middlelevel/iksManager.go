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
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/IBM/go-sdk-core/v5/core"
	"github.com/IBM/ibm-backup-recovery-sdk-go/migrationv2/common/k8"
	"github.com/IBM/ibm-backup-recovery-sdk-go/migrationv2/common/logger"
	"github.com/IBM/ibm-backup-recovery-sdk-go/migrationv2/common/metrics"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

type IksManager struct {
	iksClient     *core.BaseService
	endpointType  string
	authenticator core.Authenticator
	logger        logger.Logger
	metrics       metrics.Metrics
	accountID     string
}

func NewIKSClient(containerEndpoint, containerEndpointType string, authenticator core.Authenticator, log logger.Logger, m metrics.Metrics, accountID string) (*IksManager, error) {
	ctx := context.Background()

	log.Debug(ctx, "Initializing IKS client", "endpoint", containerEndpoint, "endpointType", containerEndpointType)

	service, err := core.NewBaseService(&core.ServiceOptions{
		URL:           containerEndpoint,
		Authenticator: authenticator,
	})
	if err != nil {
		log.Error(ctx, "Failed to initialize IKS client", logger.Err(err), "endpoint", containerEndpoint)
		return nil, fmt.Errorf("failed to initialize kube client: %v", err)
	}

	if m == nil {
		m = metrics.NewNoop()
	}

	log.Info(ctx, "IKS Manager initialized successfully", "endpoint", containerEndpoint, "endpointType", containerEndpointType)

	return &IksManager{
		iksClient:     service,
		endpointType:  containerEndpointType,
		authenticator: authenticator,
		logger:        log,
		metrics:       m,
		accountID:     accountID,
	}, nil
}

// iamAuthenticator validates that the authenticator is non-nil and is an
// *core.IamAuthenticator, returning it ready to use.  A direct type assertion
// without this check would panic on a nil or wrong-typed authenticator.
func (c *IksManager) iamAuthenticator() (*core.IamAuthenticator, error) {
	if c.authenticator == nil {
		return nil, fmt.Errorf("authenticator is not configured")
	}
	iamAuth, ok := c.authenticator.(*core.IamAuthenticator)
	if !ok {
		return nil, fmt.Errorf("authenticator must be *core.IamAuthenticator, got %T", c.authenticator)
	}
	return iamAuth, nil
}

func (c *IksManager) ApplyRBACAndGetKubeconfig(clusterName string) ([]byte, error) {
	ctx := context.Background()
	start := time.Now()
	c.logger.Info(ctx, "Applying RBAC and retrieving kubeconfig", "clusterName", clusterName, "endpointType", c.endpointType)

	// Obtain bearer token via the authenticator (IAM token exchange).
	iamAuth, err := c.iamAuthenticator()
	if err != nil {
		c.logger.Error(ctx, "Invalid authenticator", logger.Err(err), "clusterName", clusterName)
		c.recordApplyRBACMetrics(ctx, time.Since(start), "failure")
		return nil, err
	}
	bearerToken, err := iamAuth.GetToken()
	if err != nil {
		c.logger.Error(ctx, "Failed to obtain IAM bearer token", logger.Err(err), "clusterName", clusterName)
		c.recordApplyRBACMetrics(ctx, time.Since(start), "failure")
		return nil, fmt.Errorf("failed to obtain IAM bearer token: %v", err)
	}

	endpoint := "/v2/applyRBACAndGetKubeconfig"
	payload := map[string]interface{}{
		"cluster":      clusterName,
		"admin":        true,
		"format":       "yaml",
		"endpointType": c.endpointType,
	}

	builder := core.NewRequestBuilder(core.POST)
	builder.WithContext(ctx)

	if _, err := builder.SetBodyContentJSON(payload); err != nil {
		c.logger.Error(ctx, "Failed to set request body", logger.Err(err), "clusterName", clusterName)
		return nil, fmt.Errorf("failed to set request body: %v", err)
	}

	if _, err := builder.ResolveRequestURL(c.iksClient.Options.URL, endpoint, nil); err != nil {
		c.logger.Error(ctx, "Failed to resolve URL", logger.Err(err), "endpoint", endpoint)
		return nil, fmt.Errorf("failed to resolve URL: %v", err)
	}

	builder.AddHeader("Accept", "text/yaml")
	builder.AddHeader("Authorization", "Bearer "+bearerToken)

	request, err := builder.Build()
	if err != nil {
		c.logger.Error(ctx, "Failed to build request", logger.Err(err), "clusterName", clusterName)
		return nil, fmt.Errorf("failed to build request: %v", err)
	}

	c.metrics.IncCounter(ctx, metrics.MetricIBMCloudContainerServiceCallsTotal,
		metrics.Label{Key: "accountId", Value: c.accountID},
	)

	var rawBody io.ReadCloser
	response, err := c.iksClient.Request(request, &rawBody)
	duration := time.Since(start)

	if err != nil {
		c.logger.Error(ctx, "IKS API request failed", logger.Err(err), "clusterName", clusterName)
		c.recordApplyRBACMetrics(ctx, duration, "failure")
		errMsg := strings.ToLower(err.Error())
		if strings.Contains(errMsg, "not found") {
			return nil, fmt.Errorf("cluster '%s' not found. Please verify the cluster name and ensure it exists in your account", clusterName)
		}
		if strings.Contains(errMsg, "unauthorized") || strings.Contains(errMsg, "forbidden") {
			return nil, fmt.Errorf("authentication failed for cluster '%s'. Please verify your API key has access to this cluster", clusterName)
		}
		return nil, fmt.Errorf("failed to access cluster '%s': %v", clusterName, err)
	}
	defer rawBody.Close()

	body, err := io.ReadAll(rawBody)
	if err != nil {
		c.logger.Error(ctx, "Failed to read response body", logger.Err(err), "clusterName", clusterName)
		c.recordApplyRBACMetrics(ctx, duration, "failure")
		return nil, fmt.Errorf("failed to read response body: %v", err)
	}

	if response.StatusCode != http.StatusOK {
		c.logger.Error(ctx, "IKS API returned error", "statusCode", response.StatusCode, "clusterName", clusterName, "response", string(body))
		c.recordApplyRBACMetrics(ctx, duration, "failure")
		switch response.StatusCode {
		case http.StatusNotFound:
			return nil, fmt.Errorf("cluster '%s' not found. Please verify the cluster name and ensure it exists in your account", clusterName)
		case http.StatusUnauthorized, http.StatusForbidden:
			return nil, fmt.Errorf("authentication failed for cluster '%s'. Please verify your API key has access to this cluster", clusterName)
		default:
			return nil, fmt.Errorf("failed to access cluster '%s': HTTP %d: %s", clusterName, response.StatusCode, string(body))
		}
	}

	c.recordApplyRBACMetrics(ctx, duration, "success")
	c.logger.Info(ctx, "Successfully retrieved kubeconfig", "clusterName", clusterName, "configSize", len(body))
	return body, nil
}

func (c *IksManager) recordApplyRBACMetrics(ctx context.Context, duration time.Duration, status string) {
	c.metrics.IncCounter(ctx, metrics.MetricOperationTotal,
		metrics.Label{Key: "operation", Value: metrics.OperationIKSApplyRBAC},
		metrics.Label{Key: "status", Value: status},
		metrics.Label{Key: "accountId", Value: c.accountID},
	)
	c.metrics.RecordDuration(ctx, metrics.MetricOperationDuration, duration,
		metrics.Label{Key: "operation", Value: metrics.OperationIKSApplyRBAC},
		metrics.Label{Key: "accountId", Value: c.accountID},
	)
}

// GetKubeApi retrieves the Kubernetes clientset and REST configuration.
func (c *IksManager) GetKubeApi(clusterName string) (*kubernetes.Clientset, *rest.Config, error) {
	ctx := context.Background()
	start := time.Now()
	c.logger.Info(ctx, "Getting Kubernetes API client", "clusterName", clusterName, "operation", "GetKubeApi")

	kubeconfig, err := c.ApplyRBACAndGetKubeconfig(clusterName)
	if err != nil {
		c.logger.Error(ctx, "Failed to get kubeconfig", logger.Err(err), "clusterName", clusterName)
		duration := time.Since(start)
		c.metrics.IncCounter(ctx, metrics.MetricOperationTotal,
			metrics.Label{Key: "accountId", Value: c.accountID},
			metrics.Label{Key: "operation", Value: metrics.OperationIKSGetKubeApi},
			metrics.Label{Key: "status", Value: "failure"},
		)
		c.metrics.RecordDuration(ctx, metrics.MetricOperationDuration, duration,
			metrics.Label{Key: "accountId", Value: c.accountID},
			metrics.Label{Key: "operation", Value: metrics.OperationIKSGetKubeApi},
		)
		errMsg := err.Error()
		if strings.Contains(errMsg, "Not Found") || strings.Contains(errMsg, "not found") {
			return nil, nil, fmt.Errorf("cluster '%s' not found", clusterName)
		}
		return nil, nil, fmt.Errorf("failed to get kubeconfig for cluster '%s': %v", clusterName, err)
	}

	restConfig, err := clientcmd.RESTConfigFromKubeConfig(kubeconfig)
	if err != nil {
		c.logger.Error(ctx, "Failed to load REST config", logger.Err(err), "clusterName", clusterName)
		duration := time.Since(start)
		c.metrics.IncCounter(ctx, metrics.MetricOperationTotal,
			metrics.Label{Key: "accountId", Value: c.accountID},
			metrics.Label{Key: "operation", Value: metrics.OperationIKSGetKubeApi},
			metrics.Label{Key: "status", Value: "failure"},
		)
		c.metrics.RecordDuration(ctx, metrics.MetricOperationDuration, duration,
			metrics.Label{Key: "accountId", Value: c.accountID},
			metrics.Label{Key: "operation", Value: metrics.OperationIKSGetKubeApi},
		)
		return nil, nil, fmt.Errorf("failed to load REST config: %v", err)
	}

	clientSet, err := kubernetes.NewForConfig(restConfig)
	if err != nil {
		c.logger.Error(ctx, "Failed to create Kubernetes client", logger.Err(err), "clusterName", clusterName)
		duration := time.Since(start)
		c.metrics.IncCounter(ctx, metrics.MetricOperationTotal,
			metrics.Label{Key: "accountId", Value: c.accountID},
			metrics.Label{Key: "operation", Value: metrics.OperationIKSGetKubeApi},
			metrics.Label{Key: "status", Value: "failure"},
		)
		c.metrics.RecordDuration(ctx, metrics.MetricOperationDuration, duration,
			metrics.Label{Key: "accountId", Value: c.accountID},
			metrics.Label{Key: "operation", Value: metrics.OperationIKSGetKubeApi},
		)
		return nil, nil, fmt.Errorf("failed to create Kubernetes client: %v", err)
	}

	duration := time.Since(start)
	c.metrics.IncCounter(ctx, metrics.MetricOperationTotal,
		metrics.Label{Key: "accountId", Value: c.accountID},
		metrics.Label{Key: "operation", Value: metrics.OperationIKSGetKubeApi},
		metrics.Label{Key: "status", Value: "success"},
	)
	c.metrics.RecordDuration(ctx, metrics.MetricOperationDuration, duration,
		metrics.Label{Key: "accountId", Value: c.accountID},
		metrics.Label{Key: "operation", Value: metrics.OperationIKSGetKubeApi},
	)

	c.logger.Info(ctx, "Kubernetes API client created successfully", "clusterName", clusterName)
	return clientSet, restConfig, nil
}

// GetClusterBearerToken retrieves the bearer token for the cluster.
func (c *IksManager) GetClusterBearerToken(clientSet *kubernetes.Clientset) (string, error) {
	ctx := context.Background()
	start := time.Now()
	c.logger.Info(ctx, "Getting cluster bearer token", "operation", "GetClusterBearerToken")

	token, err := k8.ClusterBearerToken(clientSet)
	duration := time.Since(start)

	if err != nil {
		c.logger.Error(ctx, "Failed to get cluster bearer token", logger.Err(err))
		c.metrics.IncCounter(ctx, metrics.MetricOperationTotal,
			metrics.Label{Key: "accountId", Value: c.accountID},
			metrics.Label{Key: "operation", Value: metrics.OperationIKSGetBearerToken},
			metrics.Label{Key: "status", Value: "failure"},
		)
		c.metrics.RecordDuration(ctx, metrics.MetricOperationDuration, duration,
			metrics.Label{Key: "accountId", Value: c.accountID},
			metrics.Label{Key: "operation", Value: metrics.OperationIKSGetBearerToken},
		)
		return "", err
	}

	c.metrics.IncCounter(ctx, metrics.MetricOperationTotal,
		metrics.Label{Key: "accountId", Value: c.accountID},
		metrics.Label{Key: "operation", Value: metrics.OperationIKSGetBearerToken},
		metrics.Label{Key: "status", Value: "success"},
	)
	c.metrics.RecordDuration(ctx, metrics.MetricOperationDuration, duration,
		metrics.Label{Key: "accountId", Value: c.accountID},
		metrics.Label{Key: "operation", Value: metrics.OperationIKSGetBearerToken},
	)

	c.logger.Info(ctx, "Successfully retrieved cluster bearer token")
	return token, nil
}
