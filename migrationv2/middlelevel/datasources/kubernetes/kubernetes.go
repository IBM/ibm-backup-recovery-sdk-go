/*******************************************************************************
 * IBM Confidential
 * OCO Source Materials
 * (C) Copyright IBM Corp. 2026 All Rights Reserved.
 * The source code for this program is not  published or otherwise divested of
 * its trade secrets, irrespective of what has been deposited with
 * the U.S. Copyright Office.
 ******************************************************************************/
package kubernetes

import (
	"context"
	"fmt"
	"strings"
	"time"

	"strconv"

	"github.com/IBM/go-sdk-core/v5/core"
	"github.com/IBM/ibm-backup-recovery-sdk-go/backuprecoveryv1"
	"github.com/IBM/ibm-backup-recovery-sdk-go/migrationv2/common/errors"
	"github.com/IBM/ibm-backup-recovery-sdk-go/migrationv2/common/logger"
	"github.com/IBM/ibm-backup-recovery-sdk-go/migrationv2/common/metrics"
	"github.com/IBM/ibm-backup-recovery-sdk-go/migrationv2/common/types"
	"github.com/IBM/ibm-backup-recovery-sdk-go/migrationv2/middlelevel"
	"github.com/IBM/ibm-backup-recovery-sdk-go/migrationv2/middlelevel/connectors"
	"github.com/IBM/ibm-backup-recovery-sdk-go/migrationv2/middlelevel/datasources"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

type KubernetesDataSourceType string

const (
	Kiksclassic  = backuprecoveryv1.DataSourceConnection_ConnectionEnvType_Kiksclassic
	Kiksvpc      = backuprecoveryv1.DataSourceConnection_ConnectionEnvType_Kiksvpc
	Kroksclassic = backuprecoveryv1.DataSourceConnection_ConnectionEnvType_Kroksclassic
	Kroksvpc     = backuprecoveryv1.DataSourceConnection_ConnectionEnvType_Kroksvpc
)

type KubernetesDataSourceConfig struct {
	ClusterName     string
	ClusterId       string
	ClusterType     string // IKS or ROKS
	ClusterEndpoint string

	ContainerEndpoint     string
	ContainerEndpointType string // public or private

	Authenticator core.Authenticator

	VpcID                      string
	IncludeNamespace           string
	ExcludeNamespace           string
	BrsAgent                   types.BRSAgentImage
	KubernetesDataSourceType   KubernetesDataSourceType
	KubernetesProtectionParams *types.KubernetesProtectionParams
	KubernetesRestoreParams    *types.KubernetesRestoreParams
	Logger                     logger.Logger
	Metrics                    metrics.Metrics
	AccountID                  string
}

func (kubernetesDataSourceConfig *KubernetesDataSourceConfig) GetDataSourceType() types.DataSourceType {
	return types.DataSourceKubernetes
}

func (kubernetesDataSourceConfig *KubernetesDataSourceConfig) Validate() error {
	if kubernetesDataSourceConfig == nil {
		return errors.NewInvalidDataSourceError("kubernetes data source config cannot be nil")
	}

	// Validate that either ClusterName or ClusterId is provided
	if kubernetesDataSourceConfig.ClusterName == "" && kubernetesDataSourceConfig.ClusterId == "" {
		return errors.NewInvalidDataSourceError("either cluster_name or cluster_id is required")
	}

	// Validate ContainerEndpoint is provided
	if kubernetesDataSourceConfig.ContainerEndpoint == "" {
		return errors.NewInvalidDataSourceError("container_endpoint is required")
	}

	// Validate ContainerEndpointType is valid
	if kubernetesDataSourceConfig.ContainerEndpointType != "" {
		validTypes := map[string]bool{"public": true, "private": true, "vpe": true}
		if !validTypes[kubernetesDataSourceConfig.ContainerEndpointType] {
			return errors.NewInvalidDataSourceError("container_endpoint_type must be 'public', 'private', or 'vpe'")
		}
	}

	// Validate Authenticator is provided
	if kubernetesDataSourceConfig.Authenticator == nil {
		return errors.NewInvalidDataSourceError("authenticator is required")
	}

	// Validate ClusterType if provided
	if kubernetesDataSourceConfig.ClusterType != "" {
		validClusterTypes := map[string]bool{"IKS": true, "ROKS": true, "iks": true, "roks": true}
		if !validClusterTypes[kubernetesDataSourceConfig.ClusterType] {
			return errors.NewInvalidDataSourceError("cluster_type must be 'IKS' or 'ROKS'")
		}
	}

	return nil
}

func (kubernetesDataSourceConfig *KubernetesDataSourceConfig) CreateDataSource(brsClient types.BRSClientWrapperInterface) (datasources.DataSource, error) {
	if kubernetesDataSourceConfig == nil {
		return nil, errors.NewInvalidDataSourceError("kubernetes data source config cannot be nil")
	}
	return NewKubernetesDataSource(kubernetesDataSourceConfig.ClusterName, kubernetesDataSourceConfig, brsClient)
}

// KubernetesDataSource implements DataSource for Kubernetes/IKS/ROKS
type KubernetesDataSource struct {
	Type                       types.DataSourceType
	KubernetesDataSourceType   KubernetesDataSourceType
	Name                       string
	Config                     *KubernetesDataSourceConfig
	ClusterID                  string
	ClusterName                string
	ClusterEndpoint            string
	Namespace                  string
	KubeConfig                 string
	KubernetesProtectionParams *types.KubernetesProtectionParams
	ConnectorDeployer          connectors.ConnectorDeployer

	clientSet               *kubernetes.Clientset
	restConfig              *rest.Config
	iksClient               middlelevel.IksManagerInterface
	KubernetesRestoreParams *types.KubernetesRestoreParams
	logger                  logger.Logger
	brsClient               types.BRSClientWrapperInterface
}

// NewKubernetesDataSource creates a new Kubernetes data source
func NewKubernetesDataSource(name string, config *KubernetesDataSourceConfig, brsClient types.BRSClientWrapperInterface) (*KubernetesDataSource, error) {
	ctx := context.Background()
	if config == nil {
		return nil, errors.NewInvalidDataSourceError("kubernetes data source config cannot be nil")
	}

	// Get logger from config or create default
	log := config.Logger
	if log == nil {
		logConfig := logger.DefaultConfig()
		logConfig.ServiceName = "brs-kubernetes-datasource"
		logConfig.Environment = "production"
		log = logger.New(logConfig)
	}

	log.Info(ctx, "Creating Kubernetes data source",
		"name", name,
		"clusterName", config.ClusterName,
		"clusterType", config.ClusterType)

	iksClientInstance, err := middlelevel.NewIKSClient(config.ContainerEndpoint, config.ContainerEndpointType, config.Authenticator, log, config.Metrics, config.AccountID)
	if err != nil {
		log.Error(ctx, "Failed to create IKS client", logger.Err(err), "clusterName", config.ClusterName)
		return nil, err
	}

	clientSet, restConfig, err := iksClientInstance.GetKubeApi(config.ClusterName)
	if err != nil {
		log.Error(ctx, "Failed to get Kubernetes API", logger.Err(err), "clusterName", config.ClusterName)
		return nil, err
	}

	k := &KubernetesDataSource{
		Type:                       types.DataSourceKubernetes,
		KubernetesDataSourceType:   config.KubernetesDataSourceType,
		Name:                       name,
		Config:                     config,
		ClusterID:                  config.ClusterId,
		ClusterName:                config.ClusterName,
		ClusterEndpoint:            config.ClusterEndpoint,
		Namespace:                  "config",
		KubeConfig:                 "",
		KubernetesProtectionParams: config.KubernetesProtectionParams,
		KubernetesRestoreParams:    config.KubernetesRestoreParams,
		brsClient:                  brsClient,
		clientSet:                  clientSet,
		restConfig:                 restConfig,
		iksClient:                  iksClientInstance,
		logger:                     log,
	}

	log.Info(ctx, "Kubernetes data source created successfully",
		"name", name,
		"clusterName", config.ClusterName,
		"clusterID", config.ClusterId)

	return k, nil
}

// GetType returns the data source type
func (k *KubernetesDataSource) GetType() types.DataSourceType {
	return k.Type
}

// GetName returns the data source name
func (k *KubernetesDataSource) GetName() string {
	return k.Name
}

// Validate validates Kubernetes-specific configuration
func (k *KubernetesDataSource) Validate() error {
	if k == nil {
		return errors.NewInvalidDataSourceError("kubernetes data source cannot be nil")
	}
	if k.Name == "" {
		return errors.NewInvalidDataSourceError("data source name is required")
	}

	if k.Config.ClusterId == "" && k.Config.ClusterName == "" {
		return errors.NewInvalidDataSourceError("either cluster_id or cluster_name is required")
	}

	return nil
}

// GetKubernetesDistribution returns the kubernetesDistribution based on the k8s type
func GetKubernetesDistribution(clusterType string) *string {
	if strings.ToLower(clusterType) == "iks" {
		return core.StringPtr(backuprecoveryv1.KubernetesSourceRegistrationParams_KubernetesDistribution_Kiks)
	}

	return core.StringPtr(backuprecoveryv1.KubernetesSourceRegistrationParams_KubernetesDistribution_Kroks)
}

// RegisterSource registers the Kubernetes cluster as a protection source
func (k *KubernetesDataSource) RegisterSourceParams(ctx context.Context, connectionID string) (*backuprecoveryv1.RegisterProtectionSourceOptions, error) {
	k.logger.Info(ctx, "Preparing Kubernetes source registration parameters",
		"operation", "RegisterSourceParams",
		"connectionID", connectionID,
		"clusterName", k.ClusterName,
		"clusterType", k.Config.ClusterType)

	//step1: Get Bearer Token
	k.logger.Debug(ctx, "Getting cluster bearer token", "clusterName", k.ClusterName)

	bearerToken, err := k.iksClient.GetClusterBearerToken(k.clientSet)
	if err != nil {
		k.logger.Error(ctx, "Failed to get cluster bearer token",
			"operation", "RegisterSourceParams",
			"clusterName", k.ClusterName,
			"error", err.Error())
		return nil, errors.NewRegistrationFailedError("could not get cluster bearer token", err)
	}
	k.logger.Debug(ctx, "Bearer token obtained successfully")

	//step2: Form Params for Kubernetes Registration
	k.logger.Debug(ctx, "Forming Kubernetes registration parameters",
		"clusterType", k.Config.ClusterType,
		"endpoint", k.Config.ClusterEndpoint)

	env := backuprecoveryv1.RegisterProtectionSourceOptions_Environment_Kkubernetes

	k8sClusterParams := &backuprecoveryv1.KubernetesSourceRegistrationParams{
		Endpoint:                     core.StringPtr(k.Config.ClusterEndpoint),
		ClientPrivateKey:             core.StringPtr(bearerToken),
		DataMoverImageLocation:       core.StringPtr(k.Config.BrsAgent.DataMover),
		VeleroAwsPluginImageLocation: core.StringPtr(k.Config.BrsAgent.VeleroAWSPlugin),
		VeleroImageLocation:          core.StringPtr(k.Config.BrsAgent.Velero),
		KubernetesDistribution:       GetKubernetesDistribution(k.Config.ClusterType),
	}

	if strings.ToLower(k.Config.ClusterType) == "roks" {
		k.logger.Debug(ctx, "Adding ROKS-specific plugin", "plugin", "VeleroOpenShiftPlugin")
		k8sClusterParams.VeleroOpenshiftPluginImageLocation = core.StringPtr(k.Config.BrsAgent.VeleroOpenShiftPlugin)
	}

	connectionIdInt, err := strconv.ParseInt(connectionID, 10, 64)
	if err != nil {
		k.logger.Error(ctx, "Failed to parse connection ID",
			"operation", "RegisterSourceParams",
			"connectionID", connectionID,
			"error", err.Error())
		return nil, errors.NewRegistrationFailedError("could not parse connection id", err)
	}

	registerSourceRegistrationOptions := &backuprecoveryv1.RegisterProtectionSourceOptions{
		Environment:      core.StringPtr(env),
		Name:             core.StringPtr(k.Name),
		KubernetesParams: k8sClusterParams,
		ConnectionID:     core.Int64Ptr(connectionIdInt),
	}

	k.logger.Info(ctx, "Kubernetes source registration parameters prepared successfully",
		"operation", "RegisterSourceParams",
		"connectionID", connectionID,
		"clusterName", k.ClusterName,
		"sourceName", k.Name)

	return registerSourceRegistrationOptions, nil
}

// CreateProtectionGroup creates a protection group for Kubernetes workloads
func (k *KubernetesDataSource) CreateProtectionGroup(ctx context.Context, registrationID int64, groupParams *types.ProtectionGroupParams) (*backuprecoveryv1.CreateProtectionGroupOptions, error) {
	// Validate params
	if groupParams == nil {
		k.logger.Error(ctx, "Protection group params are required",
			"operation", "CreateProtectionGroup",
			"registrationID", registrationID)
		return nil, errors.NewInvalidConfigError("protection group params are required", nil)
	}

	k.logger.Info(ctx, "Creating Kubernetes protection group",
		"operation", "CreateProtectionGroup",
		"registrationID", registrationID,
		"groupName", groupParams.Name)

	if k.KubernetesProtectionParams == nil {
		k.logger.Error(ctx, "Kubernetes protection params are required",
			"operation", "CreateProtectionGroup",
			"registrationID", registrationID)
		return nil, errors.NewInvalidConfigError("KubernetesProtectionParams are required", nil)
	}

	protectionGroupName := groupParams.Name
	if protectionGroupName == "" {
		protectionGroupName = types.DefaultProtectionGroupName + time.Now().Format("20060102150405")
		k.logger.Debug(ctx, "Using auto-generated protection group name", "generatedName", protectionGroupName)
	}

	excludeList := parseNamespaceList(k.KubernetesProtectionParams.ExcludeNamespaces)
	includeList := parseNamespaceList(k.KubernetesProtectionParams.IncludeNamespaces)

	namespaceDetails, err := k.fetchNamespaceIDsWithNames(registrationID, includeList, excludeList)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch kubernetes namespaces: %w", err)
	}
	uniqueNameSpaces := uniqueByID(namespaceDetails)

	k8sParams := k.GetK8Object(uniqueNameSpaces)

	env := backuprecoveryv1.CreateProtectionGroupOptions_Environment_Kkubernetes
	protectionGroupOptions := &backuprecoveryv1.CreateProtectionGroupOptions{
		Environment:      &env,
		KubernetesParams: k8sParams,
	}

	return protectionGroupOptions, nil
}

// RunRestore initiates a restore operation
func (k *KubernetesDataSource) RunRestore(ctx context.Context, groupId, backupId string, targetRegistrationID int64, commonRestoreParams *types.RestoreParams) (*backuprecoveryv1.CreateRecoveryOptions, error) {
	if commonRestoreParams == nil {
		k.logger.Error(ctx, "Restore params are required",
			"operation", "RunRestore",
			"groupId", groupId,
			"backupId", backupId)
		return nil, errors.NewInvalidConfigError("restore params are required", nil)
	}

	k.logger.Info(ctx, "Initiating Kubernetes restore operation",
		"operation", "RunRestore",
		"groupId", groupId,
		"backupId", backupId,
		"targetRegistrationID", targetRegistrationID,
		"restoreName", commonRestoreParams.Name)
	if k.KubernetesRestoreParams == nil {
		k.logger.Error(ctx, "Kubernetes restore params are required",
			"operation", "RunRestore",
			"groupId", groupId,
			"backupId", backupId)
		return nil, errors.NewInvalidConfigError("kubernetes restore params are required", nil)
	}

	k.logger.Debug(ctx, "Creating recover namespace parameters",
		"groupId", groupId,
		"backupId", backupId,
		"targetRegistrationID", targetRegistrationID)

	recoverNamespaceParams, err := k.RecoverNamespace(ctx, groupId, backupId, targetRegistrationID, k.KubernetesRestoreParams, commonRestoreParams)
	if err != nil {
		k.logger.Error(ctx, "Failed to create recover namespace params",
			"operation", "RunRestore",
			"groupId", groupId,
			"backupId", backupId,
			"error", err.Error())
		return nil, fmt.Errorf("unable to create recover namespace params: %w", err)
	}

	k.logger.Debug(ctx, "Building recovery options",
		"restoreName", commonRestoreParams.Name,
		"recoveryAction", "RecoverNamespaces")

	recoveryAction := backuprecoveryv1.GetObjectSnapshotsOptions_SnapshotActions_Recovernamespaces
	kubernetesParams := &backuprecoveryv1.RecoveryRequestParamsKubernetesParams{
		RecoverNamespaceParams: recoverNamespaceParams,
		RecoveryAction:         &recoveryAction,
	}
	snapshotEnv := backuprecoveryv1.Recovery_SnapshotEnvironment_Kkubernetes
	createRecoveryOptions := &backuprecoveryv1.CreateRecoveryOptions{
		SnapshotEnvironment: &snapshotEnv,
		KubernetesParams:    kubernetesParams,
		Name:                &commonRestoreParams.Name,
	}

	k.logger.Info(ctx, "Kubernetes restore options created successfully",
		"operation", "RunRestore",
		"groupId", groupId,
		"backupId", backupId,
		"restoreName", commonRestoreParams.Name)

	return createRecoveryOptions, nil
}
