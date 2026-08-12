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
	"helm.sh/helm/v4/pkg/action"
	"helm.sh/helm/v4/pkg/registry"
	releasev1 "helm.sh/helm/v4/pkg/release/v1"
)

// HelmInstallConfig contains configuration for Helm chart installation
type HelmInstallConfig struct {
	// ReleaseName is the name of the Helm release
	ReleaseName string

	// Namespace is the Kubernetes namespace to install into
	Namespace string

	// ChartRef is the chart reference (e.g., oci://...) or local path
	ChartRef string

	// Version is the chart version (optional - empty string gets latest)
	Version string

	// Values are the Helm values to pass to the chart
	Values map[string]interface{}

	// Wait indicates whether to wait for resources to be ready
	Wait bool
}

// HelmClient interface abstracts Helm operations for better testability
// Only includes methods that are actually used in production code
type HelmClient interface {
	// Init initializes the Helm action configuration
	Init(actionConfig *action.Configuration, namespace string) error

	// Install installs a Helm chart with the given configuration
	// Handles chart location internally using the registry client from actionConfig
	Install(actionConfig *action.Configuration, config *HelmInstallConfig) (*releasev1.Release, error)
}

// RegistryClient interface abstracts registry operations
type RegistryClient interface {
	// Login logs into a container registry
	Login(host string, username, password string) error

	// Logout logs out from a container registry
	Logout(host string) error
}

// HelmRegistryClientWrapper wraps the actual Helm registry client
type HelmRegistryClientWrapper struct {
	client *registry.Client
}

// NewHelmRegistryClient creates a new registry client wrapper
func NewHelmRegistryClient() (*HelmRegistryClientWrapper, error) {
	client, err := registry.NewClient()
	if err != nil {
		return nil, err
	}
	return &HelmRegistryClientWrapper{client: client}, nil
}

// Login logs into a container registry
func (w *HelmRegistryClientWrapper) Login(host string, username, password string) error {
	return w.client.Login(host, registry.LoginOptBasicAuth(username, password))
}

// Logout logs out from a container registry
func (w *HelmRegistryClientWrapper) Logout(host string) error {
	return w.client.Logout(host)
}

// GetClient returns the underlying registry client
func (w *HelmRegistryClientWrapper) GetClient() *registry.Client {
	return w.client
}
