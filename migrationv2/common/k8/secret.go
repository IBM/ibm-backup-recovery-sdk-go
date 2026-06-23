package k8

import (
	"context"
	"fmt"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

func GetSecret(ctx context.Context, clientSet *kubernetes.Clientset, name string, namespace string) (*corev1.Secret, error) {
	secret, err := clientSet.CoreV1().Secrets(namespace).Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to get secret: %v", err)
	}

	return secret, nil
}

func CreateSecret(ctx context.Context, clientSet *kubernetes.Clientset, name string, Namespace string) error {
	secret := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: Namespace,
			Annotations: map[string]string{
				"kubernetes.io/service-account.name": SaName,
			},
		},
		Type: corev1.SecretTypeServiceAccountToken,
	}

	_, err := clientSet.CoreV1().Secrets(Namespace).Create(ctx, secret, metav1.CreateOptions{})
	if err != nil {
		return fmt.Errorf("failed to create secret: %v", err)
	}

	return nil
}
