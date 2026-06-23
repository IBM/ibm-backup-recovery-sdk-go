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
	"log/slog"
	"os"
	"time"

	"helm.sh/helm/v3/pkg/action"
	"helm.sh/helm/v3/pkg/chart/loader"
	"helm.sh/helm/v3/pkg/cli"
	"helm.sh/helm/v3/pkg/release"
	"k8s.io/cli-runtime/pkg/genericclioptions"
	"k8s.io/client-go/rest"
)

// DefaultHelmClient is the default implementation of HelmClient interface
type DefaultHelmClient struct {
	restConfig *rest.Config
	namespace  string
}

// NewDefaultHelmClient creates a new default Helm client
func NewDefaultHelmClient(restConfig *rest.Config, namespace string) *DefaultHelmClient {
	return &DefaultHelmClient{
		restConfig: restConfig,
		namespace:  namespace,
	}
}

// Init initializes the Helm action configuration
func (h *DefaultHelmClient) Init(actionConfig *action.Configuration, namespace string) error {
	restGetter := &genericclioptions.ConfigFlags{
		Namespace: &namespace,
		WrapConfigFn: func(*rest.Config) *rest.Config {
			return h.restConfig
		},
	}

	return actionConfig.Init(restGetter, namespace, os.Getenv("HELM_DRIVER"), slog.Debug)
}

// Install installs a Helm chart with the given configuration
func (h *DefaultHelmClient) Install(actionConfig *action.Configuration, config *HelmInstallConfig) (*release.Release, error) {
	install := action.NewInstall(actionConfig)
	install.ReleaseName = config.ReleaseName
	install.Namespace = config.Namespace
	install.CreateNamespace = false
	install.Wait = config.Wait
	install.Timeout = 20 * time.Minute
	install.ChartPathOptions.Version = config.Version

	// Locate the chart (handles OCI registries with registry client)
	chartPath, err := install.ChartPathOptions.LocateChart(config.ChartRef, cli.New())
	if err != nil {
		return nil, err
	}

	// Load the chart
	ch, err := loader.Load(chartPath)
	if err != nil {
		return nil, err
	}

	return install.Run(ch, config.Values)
}
