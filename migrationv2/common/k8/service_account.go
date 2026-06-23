package k8

import (
	"context"
	"fmt"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

func GetServiceAccountToken(ctx context.Context, clientSet *kubernetes.Clientset, name string, namespace string) (*corev1.ServiceAccount, error) {
	sa, err := clientSet.CoreV1().ServiceAccounts(namespace).Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to get service account: %v", err)
	}

	return sa, nil
}

func CreateServiceAccount(ctx context.Context, clientSet *kubernetes.Clientset, name string, namespace string) error {
	serviceAccount := &corev1.ServiceAccount{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
	}

	_, err := clientSet.CoreV1().ServiceAccounts(namespace).Create(ctx, serviceAccount, metav1.CreateOptions{})
	if err != nil {
		return fmt.Errorf("failed to create service account: %v", err)
	}

	return nil
}
