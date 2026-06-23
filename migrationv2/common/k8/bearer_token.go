package k8

import (
	"context"
	"fmt"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

func GetBearerToken(ctx context.Context, clientSet *kubernetes.Clientset, name string, Namespace string) (string, error) {
	secretData, err := clientSet.CoreV1().Secrets(Namespace).Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		return "", fmt.Errorf("failed to get secret: %v", err)
	}

	token, exists := secretData.Data["token"]
	if !exists {
		return "", fmt.Errorf("token not found in secret '%s'", name)
	}

	return string(token), nil
}

func ClusterBearerToken(clientSet *kubernetes.Clientset) (string, error) {
	ctx := context.TODO()

	_, err := GetServiceAccountToken(ctx, clientSet, SaName, Namespace)
	if err != nil {
		err = CreateServiceAccount(ctx, clientSet, SaName, Namespace)
		if err != nil {
			return "", err
		}
	}

	_, err = GetCRbac(ctx, clientSet, CrbName)
	if err != nil {
		err = CreateCRbac(ctx, clientSet, CrbName, Namespace)
		if err != nil {
			return "", err
		}
	}

	_, err = GetSecret(ctx, clientSet, SecretName, Namespace)
	if err != nil {
		err = CreateSecret(ctx, clientSet, SecretName, Namespace)
		if err != nil {
			return "", err
		}
	}

	token, err := GetBearerToken(ctx, clientSet, SecretName, Namespace)
	if err != nil {
		return "", err
	}

	return token, err
}
