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
	"fmt"
	"os"
	"time"

	"helm.sh/helm/v4/pkg/action"
	"helm.sh/helm/v4/pkg/chart/loader"
	"helm.sh/helm/v4/pkg/cli"
	"helm.sh/helm/v4/pkg/kube"
	releasev1 "helm.sh/helm/v4/pkg/release/v1"
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

	return actionConfig.Init(restGetter, namespace, os.Getenv("HELM_DRIVER"))
}

// Install installs a Helm chart with the given configuration
func (h *DefaultHelmClient) Install(actionConfig *action.Configuration, config *HelmInstallConfig) (*releasev1.Release, error) {
	install := action.NewInstall(actionConfig)
	install.ReleaseName = config.ReleaseName
	install.Namespace = config.Namespace
	install.CreateNamespace = false
	// Helm v4 requires WaitStrategy to be set even when Wait=false because it is
	// used unconditionally for hook execution. HookOnlyStrategy is the safe
	// default (waits only for hook Pods/Jobs, not for chart resources).
	// Upgrade to StatusWatcherStrategy when the caller asked for a full wait.
	install.WaitStrategy = kube.HookOnlyStrategy
	if config.Wait {
		install.WaitStrategy = kube.StatusWatcherStrategy
	}
	install.Timeout = 20 * time.Minute
	install.ChartPathOptions.Version = config.Version
	install.DisableOpenAPIValidation = true

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

	releaser, err := install.Run(ch, config.Values)
	if err != nil {
		return nil, err
	}
	rel, ok := releaser.(*releasev1.Release)
	if !ok {
		return nil, fmt.Errorf("unexpected release type returned from Helm install: %T", releaser)
	}
	return rel, nil
}
