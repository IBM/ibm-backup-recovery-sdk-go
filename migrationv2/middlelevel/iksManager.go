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

	"github.com/IBM-Cloud/container-services-go-sdk/kubernetesserviceapiv1"
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

	// Ensure metrics is never nil - use NoopMetrics if not provided
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

func (c *IksManager) ApplyRBACAndGetKubeconfig(clusterName string) ([]byte, error) {
	ctx := context.Background()
	start := time.Now()
	c.logger.Info(ctx, "Applying RBAC and retrieving kubeconfig", "clusterName", clusterName, "endpointType", c.endpointType)

	endpoint := "/v2/applyRBACAndGetKubeconfig"
	method := core.POST

	payload := map[string]interface{}{
		"cluster":      clusterName,
		"admin":        true,
		"format":       "yaml",
		"endpointType": c.endpointType,
	}

	builder := core.NewRequestBuilder(method)
	builder.WithContext(context.Background())

	if _, err := builder.SetBodyContentJSON(payload); err != nil {
		c.logger.Error(ctx, "Failed to set request body", logger.Err(err), "clusterName", clusterName)
		return nil, fmt.Errorf("failed to set request body: %v", err)
	}

	if _, err := builder.ResolveRequestURL(c.iksClient.Options.URL, endpoint, nil); err != nil {
		c.logger.Error(ctx, "Failed to resolve URL", logger.Err(err), "endpoint", endpoint)
		return nil, fmt.Errorf("failed to resolve URL: %v", err)
	}

	builder.AddHeader("Accept", "text/yaml")
	builder.AddHeader("Authorization", fmt.Sprintf("Bearer %s", c.iksClient.Options.Authenticator.(*core.IamAuthenticator).ApiKey))

	request, err := builder.Build()
	if err != nil {
		c.logger.Error(ctx, "Failed to build request", logger.Err(err), "clusterName", clusterName)
		return nil, fmt.Errorf("failed to build request: %v", err)
	}

	c.logger.Debug(ctx, "Sending request to IKS API", "endpoint", endpoint, "clusterName", clusterName)

	// Record IBM Cloud Container Service call metric
	c.metrics.IncCounter(ctx, metrics.MetricIBMCloudContainerServiceCallsTotal,
		metrics.Label{Key: "accountId", Value: c.accountID},
	)

	var rawBody io.ReadCloser
	response, err := c.iksClient.Request(request, &rawBody)
	duration := time.Since(start)

	if err != nil {
		c.logger.Error(ctx, "IKS API request failed", logger.Err(err), "clusterName", clusterName)
		// Record failed operation
		c.metrics.IncCounter(ctx, metrics.MetricOperationTotal,
			metrics.Label{Key: "operation", Value: metrics.OperationIKSApplyRBAC},
			metrics.Label{Key: "status", Value: "failure"},
			metrics.Label{Key: "accountId", Value: c.accountID},
		)
		c.metrics.RecordDuration(ctx, metrics.MetricOperationDuration, duration,
			metrics.Label{Key: "operation", Value: metrics.OperationIKSApplyRBAC},
			metrics.Label{Key: "accountId", Value: c.accountID},
		)
		return nil, fmt.Errorf("request failed: %v", err)
	}

	defer func() {
		rawBody.Close()
	}()

	body, err := io.ReadAll(rawBody)
	if err != nil {
		c.logger.Error(ctx, "Failed to read response body", logger.Err(err), "clusterName", clusterName)
		// Record failed operation
		c.metrics.IncCounter(ctx, metrics.MetricOperationTotal,
			metrics.Label{Key: "operation", Value: metrics.OperationIKSApplyRBAC},
			metrics.Label{Key: "status", Value: "failure"},
			metrics.Label{Key: "accountId", Value: c.accountID},
		)
		c.metrics.RecordDuration(ctx, metrics.MetricOperationDuration, duration,
			metrics.Label{Key: "operation", Value: metrics.OperationIKSApplyRBAC},
			metrics.Label{Key: "accountId", Value: c.accountID},
		)
		return nil, fmt.Errorf("failed to read response body: %v", err)
	}

	if response.StatusCode != http.StatusOK {
		c.logger.Error(ctx, "IKS API returned error", "statusCode", response.StatusCode, "clusterName", clusterName, "response", string(body))
		// Record failed operation
		c.metrics.IncCounter(ctx, metrics.MetricOperationTotal,
			metrics.Label{Key: "operation", Value: metrics.OperationIKSApplyRBAC},
			metrics.Label{Key: "status", Value: "failure"},
			metrics.Label{Key: "accountId", Value: c.accountID},
		)
		c.metrics.RecordDuration(ctx, metrics.MetricOperationDuration, duration,
			metrics.Label{Key: "operation", Value: metrics.OperationIKSApplyRBAC},
			metrics.Label{Key: "accountId", Value: c.accountID},
		)

		// Provide user-friendly error messages based on status code
		switch response.StatusCode {
		case http.StatusNotFound:
			return nil, fmt.Errorf("cluster '%s' not found. Please verify the cluster name and ensure it exists in your account", clusterName)
		case http.StatusUnauthorized, http.StatusForbidden:
			return nil, fmt.Errorf("authentication failed for cluster '%s'. Please verify your API key has access to this cluster", clusterName)
		default:
			return nil, fmt.Errorf("failed to access cluster '%s': HTTP %d: %s", clusterName, response.StatusCode, string(body))
		}
	}

	// Record successful operation
	c.metrics.IncCounter(ctx, metrics.MetricOperationTotal,
		metrics.Label{Key: "operation", Value: metrics.OperationIKSApplyRBAC},
		metrics.Label{Key: "status", Value: "success"},
		metrics.Label{Key: "accountId", Value: c.accountID},
	)
	c.metrics.RecordDuration(ctx, metrics.MetricOperationDuration, duration,
		metrics.Label{Key: "operation", Value: metrics.OperationIKSApplyRBAC},
		metrics.Label{Key: "accountId", Value: c.accountID},
	)

	c.logger.Info(ctx, "Successfully retrieved kubeconfig", "clusterName", clusterName, "configSize", len(body))
	return body, nil
}

// GetKubeApi retrieves the Kubernetes clientset and REST configuration.
func (c *IksManager) GetKubeApi(clusterName string) (*kubernetes.Clientset, *rest.Config, error) {
	ctx := context.Background()
	start := time.Now()
	c.logger.Info(ctx, "Getting Kubernetes API client", "clusterName", clusterName, "operation", "GetKubeApi")

	_, err := kubernetesserviceapiv1.NewKubernetesServiceApiV1(
		&kubernetesserviceapiv1.KubernetesServiceApiV1Options{
			Authenticator: c.authenticator,
		},
	)
	if err != nil {
		c.logger.Error(ctx, "Failed to create Kubernetes service API client", logger.Err(err), "clusterName", clusterName)
		// Record failed operation
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
		return nil, nil, err
	}

	c.logger.Debug(ctx, "Fetching kubeconfig for cluster", "clusterName", clusterName)
	kubeconfig, err := c.ApplyRBACAndGetKubeconfig(clusterName)
	if err != nil {
		c.logger.Error(ctx, "Failed to get kubeconfig", logger.Err(err), "clusterName", clusterName)
		// Record failed operation
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

	c.logger.Debug(ctx, "Creating REST config from kubeconfig", "clusterName", clusterName)
	restConfig, err := clientcmd.RESTConfigFromKubeConfig(kubeconfig)
	if err != nil {
		c.logger.Error(ctx, "Failed to load REST config", logger.Err(err), "clusterName", clusterName)
		// Record failed operation
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

	c.logger.Debug(ctx, "Creating Kubernetes clientset", "clusterName", clusterName)
	clientSet, err := kubernetes.NewForConfig(restConfig)
	if err != nil {
		c.logger.Error(ctx, "Failed to create Kubernetes client", logger.Err(err), "clusterName", clusterName)
		// Record failed operation
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

	// Record successful operation
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

// GetClusterBearerToken retrieves the bearer token for the cluster
func (c *IksManager) GetClusterBearerToken(clientSet *kubernetes.Clientset) (string, error) {
	ctx := context.Background()
	start := time.Now()
	c.logger.Info(ctx, "Getting cluster bearer token", "operation", "GetClusterBearerToken")

	token, err := k8.ClusterBearerToken(clientSet)
	duration := time.Since(start)

	if err != nil {
		c.logger.Error(ctx, "Failed to get cluster bearer token", logger.Err(err))
		// Record failed operation
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

	// Record successful operation
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
