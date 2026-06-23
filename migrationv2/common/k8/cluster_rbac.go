package k8

import (
	"context"
	"fmt"

	v1 "k8s.io/api/rbac/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

func GetCRbac(ctx context.Context, clientSet *kubernetes.Clientset, name string) (*v1.ClusterRoleBinding, error) {
	rbac, err := clientSet.RbacV1().ClusterRoleBindings().Get(ctx, name, metav1.GetOptions{})
	return rbac, err
}

func CreateCRbac(ctx context.Context, clientSet *kubernetes.Clientset, name string, Namespace string) error {
	clusterRoleBinding := &v1.ClusterRoleBinding{
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
		},
		Subjects: []v1.Subject{
			{
				Kind:      "ServiceAccount",
				Name:      SaName,
				Namespace: Namespace,
			},
		},
		RoleRef: v1.RoleRef{
			APIGroup: "rbac.authorization.k8s.io",
			Kind:     "ClusterRole",
			Name:     "cluster-admin",
		},
	}

	_, err := clientSet.RbacV1().ClusterRoleBindings().Create(ctx, clusterRoleBinding, metav1.CreateOptions{})
	if err != nil {
		return fmt.Errorf("failed to create cluster role binding: %v", err)
	}

	return nil
}
