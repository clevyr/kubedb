package kubernetes

import (
	"context"
	"errors"
	"fmt"
	v1 "k8s.io/api/core/v1"
	v1meta "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func FindSecretKeyRef(pod v1.Pod, envName string) *v1.SecretKeySelector {
	for _, container := range pod.Spec.Containers {
		for _, env := range container.Env {
			if env.Name == envName {
				return env.ValueFrom.SecretKeyRef
			}
		}
	}
	return nil
}

var ErrNoSecret = errors.New("secret not found")

var ErrSecretDoesNotHaveKey = errors.New("secret does not have key")

func (client KubeClient) GetSecretFromEnv(pod v1.Pod, envNames []string) (string, error) {
	var secretKeyRef *v1.SecretKeySelector
	for _, envName := range envNames {
		secretKeyRef = FindSecretKeyRef(pod, envName)
		if secretKeyRef != nil {
			break
		}
	}
	if secretKeyRef == nil {
		return "", ErrNoSecret
	}
	secret, err := client.Secrets().Get(context.Background(), secretKeyRef.Name, v1meta.GetOptions{})
	if err != nil {
		return "", err
	}
	data, ok := secret.Data[secretKeyRef.Key]
	if !ok {
		return "", fmt.Errorf("%w: %v", ErrSecretDoesNotHaveKey, secretKeyRef)
	}
	return string(data), nil
}
