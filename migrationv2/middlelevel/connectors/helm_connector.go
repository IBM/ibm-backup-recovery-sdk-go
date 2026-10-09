/*******************************************************************************
 * IBM Confidential
 * OCO Source Materials
 * (C) Copyright IBM Corp. 2026 All Rights Reserved.
 * The source code for this program is not  published or otherwise divested of
 * its trade secrets, irrespective of what has been deposited with
 * the U.S. Copyright Office.
 ******************************************************************************/
package connectors

import (
	"context"
	"fmt"
	"time"

	"github.com/IBM/go-sdk-core/v5/core"
	backuprecoveryv1 "github.com/IBM/ibm-backup-recovery-sdk-go/backuprecoveryv1"
	"github.com/IBM/ibm-backup-recovery-sdk-go/migrationv2/common/config"
	"github.com/IBM/ibm-backup-recovery-sdk-go/migrationv2/common/errors"
	"strings"

	"github.com/IBM/ibm-backup-recovery-sdk-go/migrationv2/common/logger"
	"github.com/IBM/ibm-backup-recovery-sdk-go/migrationv2/common/metrics"
	"github.com/IBM/ibm-backup-recovery-sdk-go/migrationv2/middlelevel"
	"helm.sh/helm/v4/pkg/action"
	corev1 "k8s.io/api/core/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

type HelmKubeConnectorConfig struct {
	// Namespace where the connector will be deployed
	Namespace string `json:"namespace"`

	ClusterName string `json:"clusterName"`

	ContainerEndpoint     string `json:"containerEndpoint,omitempty"`
	ContainerEndpointType string `json:"containerEndpointType,omitempty"` // public or private

	// StorageClass specifies the storage class to use for persistent volumes (optional).
	// If not provided, Helm will use the default storage class.
	StorageClass string `json:"storageClass,omitempty"`

	// Replicas specifies the number of connector pods to run
	// Default: 1 for Helm/Operator, ignored for DaemonSet
	Replicas int32 `json:"replicas"`

	// ChartVersion specifies the Helm chart version (for Helm connector)
	ChartVersion string `json:"chartVersion,omitempty"`

	ReleaseName string `json:"releaseName,omitempty"`

	ChartName string `json:"chartName,omitempty"`

	ChartReference string `json:"chartReference,omitempty"`

	WaitTillDeploy bool `json:"waitTillDeploy,omitempty"` // defaults to false

	// OperatorVersion specifies the operator version (for Operator connector)
	OperatorVersion string `json:"operatorVersion,omitempty"`

	// ImagePullPolicy specifies when to pull the connector image
	// Options: Always, IfNotPresent, Never
	ImagePullPolicy string `json:"imagePullPolicy,omitempty"`

	// Resources specifies resource requests and limits for connector pods
	Resources *ResourceRequirements `json:"resources,omitempty"`

	// NodeSelector specifies node selection constraints
	NodeSelector map[string]string `json:"nodeSelector,omitempty"`

	// Tolerations specifies pod tolerations
	Tolerations []map[string]interface{} `json:"tolerations,omitempty"`

	// Affinity specifies pod affinity/anti-affinity rules
	Affinity map[string]interface{} `json:"affinity,omitempty"`

	// ServiceAccount specifies the service account to use
	ServiceAccount string `json:"serviceAccount,omitempty"`

	// ImagePullSecrets specifies the list of secret names for pulling images (e.g. ["all-icr-io"])
	// Defaults to ["all-icr-io"] if not specified.
	ImagePullSecrets []string `json:"imagePullSecrets,omitempty"`

	// SCCEnabled specifies whether OpenShift Security Context Constraints (SCC) is enabled.
	// Defaults to false if not specified.
	SCCEnabled *bool `json:"sccEnabled,omitempty"`

	// CustomValues allows passing additional Helm values or operator config
	CustomValues map[string]interface{} `json:"customValues,omitempty"`

	// ConnectorType specifies which type of connector to deploy
	ConnectorType ConnectorType `json:"connectorType"`

	// // RegistrationToken is the token from BRS connection creation
	// RegistrationToken string `json:"registrationToken"`

	// AuthConfig contains authentication configuration for accessing the datasource
	// For Kubernetes: KubernetesAuthConfig (kubeconfig, token, or certificate)
	// For VPC VSI: VSIAuthConfig (SSH key, password)
	AuthConfig AuthConfig `json:"authConfig"`

	// DataSourceConfig contains datasource-specific configuration
	// For Kubernetes: cluster_id, cluster_endpoint, etc.
	// For VPC VSI: vpc_id, vsi_ids, region, etc.
	DataSourceConfig map[string]interface{} `json:"dataSourceConfig"`

	// DeploymentParams contains connector-specific deployment parameters
	// For Helm: namespace, chart_version, values, etc.
	// For Agent: install_path, service_account, etc.
	DeploymentParams map[string]interface{} `json:"deploymentParams"`

	// Logger for structured logging
	Logger logger.Logger `json:"-"`

	// Metrics for monitoring (optional)
	Metrics metrics.Metrics `json:"-"`

	// AccountID for metrics labeling (optional)
	AccountID string `json:"-"`
}

// ResourceRequirements specifies compute resource requirements
type ResourceRequirements struct {
	// Requests specifies minimum resources required
	Requests *ResourceList `json:"requests,omitempty"`

	// Limits specifies maximum resources allowed
	Limits *ResourceList `json:"limits,omitempty"`
}

// ResourceList specifies CPU and memory resources
type ResourceList struct {
	// CPU in cores (e.g., "500m" for 0.5 cores, "2" for 2 cores)
	CPU string `json:"cpu,omitempty"`

	// Memory in bytes (e.g., "512Mi", "2Gi")
	Memory string `json:"memory,omitempty"`
}

// GetType returns the connector type
func (config *HelmKubeConnectorConfig) CreateConnectorDeployer() (ConnectorDeployer, error) {
	return NewHelmConnector(config)
}

// GetType returns the connector type
func (config *HelmKubeConnectorConfig) GetType() ConnectorType {
	return ConnectorTypeHelm
}

// Validate validates the deployment configuration
// func (config *HelmKubeConnectorConfig) GetRegistrationToken() string {
// 	return config.RegistrationToken
// }

// Validate validates the deployment configuration
func (config *HelmKubeConnectorConfig) Validate() error {
	if config == nil {
		return errors.NewInvalidConfigError("connector deploy config is required", nil)
	}

	// Validate required fields
	if config.Namespace == "" {
		return errors.NewInvalidConfigError("namespace is required", nil)
	}

	if config.ClusterName == "" {
		return errors.NewInvalidConfigError("clusterName is required", nil)
	}

	// Validate authentication config
	if config.AuthConfig == nil {
		return errors.NewInvalidConfigError("authentication config is required", nil)
	}

	// Validate the auth config itself
	if err := config.AuthConfig.Validate(); err != nil {
		return err
	}

	return nil
}

// GetRequirements returns deployment requirements for Helm connector
func (config *HelmKubeConnectorConfig) GetRequirements() *ConnectorRequirements {
	return &ConnectorRequirements{
		MinimumVersion: "3.0.0",
		RequiredTools:  []string{"helm", "kubectl"},
		RequiredPermissions: []string{
			"create/update/delete deployments",
			"create/update/delete services",
			"create/update/delete configmaps",
			"create/update/delete secrets",
			"create/update/delete serviceaccounts",
			"create/update/delete roles",
			"create/update/delete rolebindings",
		},
	}
}

// HelmConnector implements ConnectorDeployer for Helm-based Kubernetes connector
type HelmConnector struct {
	// Configuration
	namespace       string
	releaseName     string
	chartRepo       string
	chartVersion    string
	ImagePullPolicy string
	replicas        int32
	clientSet       kubernetes.Interface
	restConfig      *rest.Config
	iksClient       *middlelevel.IksManager
	config          *HelmKubeConnectorConfig
	helmClient      HelmClient
	logger          logger.Logger

	// deployCtx is injected by the task layer via SetDeployContext before Deploy.
	// It carries the BRS client and the platform type from ConnectionResult.Type.
	deployCtx ConnectorDeployContext

	// State
	deployed bool
}

// NewHelmConnector creates a new Helm connector deployer
func NewHelmConnector(kubernetesConnectorConfig *HelmKubeConnectorConfig) (*HelmConnector, error) {
	if kubernetesConnectorConfig == nil {
		return nil, errors.NewInvalidConfigError("connector config is required", nil)
	}

	// Get logger from config or create default
	var log logger.Logger
	if kubernetesConnectorConfig.Logger != nil {
		log = kubernetesConnectorConfig.Logger
	} else {
		logConfig := logger.DefaultConfig()
		logConfig.ServiceName = "brs-helm-connector"
		logConfig.Environment = "production"
		log = logger.New(logConfig)
	}

	// Default metrics to NoOp if not provided
	if kubernetesConnectorConfig.Metrics == nil {
		kubernetesConnectorConfig.Metrics = metrics.NewNoop()
	}

	auth := &core.IamAuthenticator{
		URL:    kubernetesConnectorConfig.AuthConfig.GetIamURL(),
		ApiKey: kubernetesConnectorConfig.AuthConfig.GetAPIKey(),
	}

	iksClientInstance, err := middlelevel.NewIKSClient(
		kubernetesConnectorConfig.ContainerEndpoint,
		kubernetesConnectorConfig.ContainerEndpointType,
		auth,
		log,
		kubernetesConnectorConfig.Metrics,
		kubernetesConnectorConfig.AccountID,
	)
	if err != nil {
		return nil, err
	}

	clientSet, restConfig, err := iksClientInstance.GetKubeApi(kubernetesConnectorConfig.ClusterName)
	if err != nil {
		log.Error(context.Background(), "Failed to get kubeconfig", logger.Err(err), "clusterName", kubernetesConnectorConfig.ClusterName)
		return nil, err
	}

	// Create Helm client
	helmClient := NewDefaultHelmClient(restConfig, kubernetesConnectorConfig.Namespace)

	return &HelmConnector{
		namespace:       kubernetesConnectorConfig.Namespace,
		releaseName:     config.DefaultIfEmpty(kubernetesConnectorConfig.ReleaseName, "brs-connector"),
		chartRepo:       config.DefaultIfEmpty(kubernetesConnectorConfig.ChartReference, "oci://icr.io/ext/brs/brs-ds-connector-chart"),
		chartVersion:    kubernetesConnectorConfig.ChartVersion,
		ImagePullPolicy: config.DefaultIfEmpty(kubernetesConnectorConfig.ImagePullPolicy, "IfNotPresent"),
		config:          kubernetesConnectorConfig,
		clientSet:       clientSet,
		replicas:        kubernetesConnectorConfig.Replicas,
		restConfig:      restConfig,
		iksClient:       iksClientInstance,
		logger:          log,
		helmClient:      helmClient,
	}, nil
}

// GetType returns the connector type
func (h *HelmConnector) GetType() ConnectorType {
	return ConnectorTypeHelm
}

// SetDeployContext satisfies ConnectorDeployer. The task layer calls this before
// Deploy, injecting the BRS client and the platform type from ConnectionResult.Type
// so that live chart metadata can be fetched.
func (h *HelmConnector) SetDeployContext(ctx ConnectorDeployContext) {
	h.deployCtx = ctx
}

// resolveChartRef calls the BRS metadata API to fetch the live chart OCI
// reference for the given platform type. It only fills in the fields that the
// user left empty in the config:
//   - chartRepo  is resolved only when config.ChartReference == ""
//   - chartVersion is resolved only when config.ChartVersion == ""
//
// If both fields were provided by the user the API call is skipped entirely.
// A non-fatal error is logged and the caller continues with the static values.
func (h *HelmConnector) resolveChartRef(ctx context.Context) {
	needRepo := h.config.ChartReference == ""
	needVersion := h.config.ChartVersion == ""

	// Nothing to resolve — user supplied both values explicitly.
	if !needRepo && !needVersion {
		h.logger.Debug(ctx, "Chart reference and version provided by user; skipping metadata fetch",
			"chartReference", h.config.ChartReference, "chartVersion", h.config.ChartVersion)
		return
	}

	brsClient := h.deployCtx.BRSClient
	platformType := h.deployCtx.PlatformType

	if brsClient == nil || platformType == "" {
		return
	}

	opts := &backuprecoveryv1.GetConnectorMetadataOptions{
		XIBMTenantID: core.StringPtr(brsClient.GetTenantId()),
	}
	metadata, _, err := brsClient.GetBRSClient().GetConnectorMetadata(opts)
	if err != nil {
		h.logger.Warn(ctx, "Failed to fetch connector metadata from BRS; using static chart reference",
			"platformType", platformType, "error", err)
		return
	}

	for _, info := range metadata.K8sConnectorInfoList {
		if info.K8sPlatformType == nil || *info.K8sPlatformType != platformType {
			continue
		}
		ref := info.HelmChartOciRef
		if ref == nil {
			break
		}

		// Resolve chartRepo only if the user did not provide one.
		if needRepo {
			registryHost := ""
			if ref.RegistryHost != nil {
				registryHost = *ref.RegistryHost
			}
			namespace := ""
			if ref.Namespace != nil {
				namespace = *ref.Namespace + "/"
			}
			repository := ""
			if ref.Repository != nil {
				repository = *ref.Repository
			}
			chartRef := registryHost + namespace + repository
			if chartRef != "" {
				if !strings.HasPrefix(chartRef, "oci://") {
					chartRef = "oci://" + chartRef
				}
				h.chartRepo = chartRef
				h.logger.Info(ctx, "Resolved chart reference from BRS metadata",
					"platformType", platformType, "chartRef", chartRef)
			}
		}

		// Resolve chartVersion only if the user did not provide one.
		if needVersion && ref.Tag != nil && *ref.Tag != "" {
			h.chartVersion = *ref.Tag
			h.logger.Info(ctx, "Resolved chart version from BRS metadata",
				"platformType", platformType, "chartVersion", h.chartVersion)
		}

		return
	}

	h.logger.Warn(ctx, "No matching platform type in BRS metadata; using static chart reference",
		"platformType", platformType)
}

// Deploy deploys the Helm-based connector to Kubernetes cluster
func (h *HelmConnector) Deploy(ctx context.Context, registrationToken string) (*ConnectorResult, error) {
	if h.deployed {
		return nil, fmt.Errorf("connector is already deployed in namespace %q; call Delete first", h.namespace)
	}

	h.logger.Info(ctx, "Starting Helm connector deployment",
		"namespace", h.namespace,
		"releaseName", h.releaseName,
		"chartVersion", h.chartVersion,
		"operation", "Deploy")

	if err := h.config.AuthConfig.Validate(); err != nil {
		h.logger.Error(ctx, "Authentication config validation failed", logger.Err(err))
		return nil, err
	}

	// Resolve chart OCI reference from BRS metadata API when the user left
	// ChartReference or ChartVersion empty. PlatformType comes from ConnectionResult.Type.
	h.resolveChartRef(ctx)

	// Step 1: Ensure namespace exists
	h.logger.Debug(ctx, "Checking if namespace exists", "namespace", h.namespace)
	if _, err := h.clientSet.CoreV1().Namespaces().Get(ctx, h.namespace, v1.GetOptions{}); err != nil {
		if !k8serrors.IsNotFound(err) {
			h.logger.Error(ctx, "Failed to check namespace existence", logger.Err(err), "namespace", h.namespace)
			return nil, fmt.Errorf("failed to check namespace %q: %w", h.namespace, err)
		}
		h.logger.Info(ctx, "Namespace not found, creating new namespace", "namespace", h.namespace)
		_, err = h.clientSet.CoreV1().Namespaces().Create(ctx, &corev1.Namespace{
			ObjectMeta: v1.ObjectMeta{Name: h.namespace},
		}, v1.CreateOptions{})
		if err != nil {
			h.logger.Error(ctx, "Failed to create namespace", logger.Err(err), "namespace", h.namespace)
			return nil, err
		}
		h.logger.Info(ctx, "Namespace created successfully", "namespace", h.namespace)
	} else {
		h.logger.Debug(ctx, "Namespace already exists", "namespace", h.namespace)
	}

	// Step 2: Setup Helm
	h.logger.Debug(ctx, "Setting up Helm action configuration", "namespace", h.namespace)
	actionConfig := new(action.Configuration)
	if err := h.helmClient.Init(actionConfig, h.namespace); err != nil {
		h.logger.Error(ctx, "Failed to initialize Helm action configuration", logger.Err(err))
		return nil, err
	}

	// Step 3: Wire registry client into action config
	h.logger.Debug(ctx, "Setting up registry client for Helm chart repository")
	if err := h.setupRegistryClient(actionConfig); err != nil {
		h.logger.Error(ctx, "Failed to setup registry client", logger.Err(err))
		return nil, err
	}

	// Step 4: Prepare values
	// Default image pull secrets to ["all-icr-io"] if not configured
	imagePullSecrets := h.config.ImagePullSecrets
	if len(imagePullSecrets) == 0 {
		imagePullSecrets = []string{"all-icr-io"}
	}

	// Default SCC to false if not configured
	sccEnabled := false
	if h.config.SCCEnabled != nil {
		sccEnabled = *h.config.SCCEnabled
	}

	values := map[string]interface{}{
		"secrets": map[string]interface{}{
			"registrationToken": registrationToken,
		},
		"image": map[string]interface{}{
			"pullPolicy": h.ImagePullPolicy,
		},
		"imagePullSecrets": imagePullSecrets,
		"deploymentPlatform": map[string]interface{}{
			"rocp": map[string]interface{}{
				"sccEnabled": sccEnabled,
			},
		},
		"replicaCount":     h.replicas,
		"fullnameOverride": h.releaseName,
	}

	// Add storage class if provided by user
	if h.config.StorageClass != "" {
		h.logger.Debug(ctx, "Configuring storage class", "storageClass", h.config.StorageClass)
		values["volumeClaimTemplate"] = map[string]interface{}{
			"storageClass": h.config.StorageClass,
		}
	}

	h.logger.Info(ctx, "Installing Helm release",
		"releaseName", h.releaseName,
		"namespace", h.namespace,
		"replicas", h.replicas,
		"waitForDeploy", h.config.WaitTillDeploy)

	// Add NodeSelector if specified
	if len(h.config.NodeSelector) > 0 {
		values["nodeSelector"] = h.config.NodeSelector
	}

	// Add Tolerations if specified
	if len(h.config.Tolerations) > 0 {
		values["tolerations"] = h.config.Tolerations
	}

	// Add ServiceAccount if specified
	if h.config.ServiceAccount != "" {
		values["serviceAccount"] = map[string]interface{}{
			"name": h.config.ServiceAccount,
		}
	}

	// Add Resources if specified
	if h.config.Resources != nil {
		resourcesMap := make(map[string]interface{})
		if h.config.Resources.Requests != nil {
			requests := make(map[string]interface{})
			if h.config.Resources.Requests.CPU != "" {
				requests["cpu"] = h.config.Resources.Requests.CPU
			}
			if h.config.Resources.Requests.Memory != "" {
				requests["memory"] = h.config.Resources.Requests.Memory
			}
			if len(requests) > 0 {
				resourcesMap["requests"] = requests
			}
		}
		if h.config.Resources.Limits != nil {
			limits := make(map[string]interface{})
			if h.config.Resources.Limits.CPU != "" {
				limits["cpu"] = h.config.Resources.Limits.CPU
			}
			if h.config.Resources.Limits.Memory != "" {
				limits["memory"] = h.config.Resources.Limits.Memory
			}
			if len(limits) > 0 {
				resourcesMap["limits"] = limits
			}
		}
		if len(resourcesMap) > 0 {
			values["resources"] = resourcesMap
		}
	}

	// Merge CustomValues if specified (allows overriding any value)
	if len(h.config.CustomValues) > 0 {
		for key, value := range h.config.CustomValues {
			values[key] = value
		}
	}

	// Step 5: Install Helm chart (handles chart location internally with registry client)
	installConfig := &HelmInstallConfig{
		ReleaseName: h.releaseName,
		Namespace:   h.namespace,
		ChartRef:    h.chartRepo,
		Version:     h.chartVersion,
		Values:      values,
		Wait:        h.config.WaitTillDeploy,
	}

	release, err := h.helmClient.Install(actionConfig, installConfig)
	if err != nil {
		h.logger.Error(ctx, "Helm release installation failed",
			logger.Err(err),
			"releaseName", h.releaseName,
			"namespace", h.namespace)
		return nil, err
	}

	// Step 6: Build result
	result := &ConnectorResult{
		Status:    "deployed",
		Message:   fmt.Sprintf("Deployed in namespace %s", release.Namespace),
		CreatedAt: time.Now(),
	}

	h.deployed = true

	h.logger.Info(ctx, "Helm connector deployed successfully",
		"releaseName", h.releaseName,
		"namespace", release.Namespace,
		"status", result.Status,
		"chartVersion", h.chartVersion)

	return result, nil
}

// GetStatus retrieves the status of the Helm connector
func (h *HelmConnector) GetStatus(ctx context.Context, connectorID string) (string, error) {
	h.logger.Debug(ctx, "Retrieving Helm connector status",
		"connectorID", connectorID,
		"releaseName", h.releaseName,
		"namespace", h.namespace,
		"operation", "GetStatus")

	// TODO: Implement actual status check
	// This would typically:
	// 1. Check Helm release status
	// 2. Check pod status in namespace
	// 3. Verify connectivity to BRS
	// 4. Check connector health endpoint

	if !h.deployed {
		h.logger.Info(ctx, "Connector not deployed", "connectorID", connectorID, "status", "not_deployed")
		return "not_deployed", nil
	}

	h.logger.Info(ctx, "Connector status retrieved", "connectorID", connectorID, "status", "running")
	return "running", nil
}

// Delete removes the Helm connector from the cluster by deleting the namespace
func (h *HelmConnector) Delete(ctx context.Context, connectorID string) error {
	h.logger.Info(ctx, "Deleting Helm connector",
		"connectorID", connectorID,
		"releaseName", h.releaseName,
		"namespace", h.namespace,
		"operation", "Delete")

	// Delete the namespace - this will remove all resources including the Helm release
	h.logger.Debug(ctx, "Deleting namespace", "namespace", h.namespace)
	if h.clientSet == nil {
		return fmt.Errorf("kubernetes client not initialized")
	}
	err := h.clientSet.CoreV1().Namespaces().Delete(ctx, h.namespace, v1.DeleteOptions{})
	if err != nil {
		h.logger.Error(ctx, "Failed to delete namespace",
			logger.Err(err),
			"namespace", h.namespace,
			"connectorID", connectorID)
		return err
	}

	h.deployed = false

	h.logger.Info(ctx, "Helm connector deleted successfully",
		"connectorID", connectorID,
		"releaseName", h.releaseName,
		"namespace", h.namespace)

	return nil
}

func (h *HelmConnector) setupRegistryClient(actionConfig *action.Configuration) error {
	switch h.config.AuthConfig.GetAuthMethod() {
	case AuthMethodAPIKey, AuthMethodToken:
		regClient, err := NewHelmRegistryClient()
		if err != nil {
			return err
		}
		actionConfig.RegistryClient = regClient.GetClient()
		return nil
	case AuthMethodKubeconfig:
		return nil
	default:
		return errors.NewInvalidConfigError("unsupported auth method", nil)
	}
}
