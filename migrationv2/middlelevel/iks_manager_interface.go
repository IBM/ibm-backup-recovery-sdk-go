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
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

// IksManagerInterface defines the interface for IKS/ROKS cluster management operations
type IksManagerInterface interface {
	// ApplyRBACAndGetKubeconfig applies RBAC and retrieves kubeconfig for the cluster
	ApplyRBACAndGetKubeconfig(clusterName string) ([]byte, error)

	// GetKubeApi retrieves the Kubernetes clientset and REST configuration
	GetKubeApi(clusterName string) (*kubernetes.Clientset, *rest.Config, error)

	// GetClusterBearerToken retrieves the bearer token for the cluster
	GetClusterBearerToken(clientSet *kubernetes.Clientset) (string, error)
}
