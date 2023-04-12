package kubernetes

import (
	"context"
	"errors"
	"fmt"

	v1 "k8s.io/api/core/v1"
	v1meta "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var ErrEnvVarNotFound = errors.New("env var not found")

func FindEnvVar(pod v1.Pod, envName string) (*v1.EnvVar, error) {
	for _, container := range pod.Spec.Containers {
		for _, env := range container.Env {
			if env.Name == envName {
				return &env, nil
			}
		}
	}
	return nil, fmt.Errorf("%v: %s", ErrEnvVarNotFound, envName)
}

var (
	ErrNoEnvNames           = errors.New("dialect does not contain any env names")
	ErrSecretDoesNotHaveKey = errors.New("secret does not have key")
)

func (client KubeClient) GetValueFromEnv(ctx context.Context, pod v1.Pod, envNames []string) (string, error) {
	if len(envNames) == 0 {
		return "", ErrNoEnvNames
	}

	var err error
	var envVar *v1.EnvVar
	for _, envName := range envNames {
		envVar, err = FindEnvVar(pod, envName)
		if err == nil {
			break
		}
	}
	if err != nil {
		return "", err
	}

	if envVar.ValueFrom != nil && envVar.ValueFrom.SecretKeyRef != nil {
		secretKeyRef := envVar.ValueFrom.SecretKeyRef
		secret, err := client.Secrets().Get(ctx, secretKeyRef.Name, v1meta.GetOptions{})
		if err != nil {
			return "", err
		}
		data, ok := secret.Data[secretKeyRef.Key]
		if !ok {
			return "", fmt.Errorf("%w: %v", ErrSecretDoesNotHaveKey, secretKeyRef)
		}
		return string(data), nil
	}
	return envVar.Value, nil
}
