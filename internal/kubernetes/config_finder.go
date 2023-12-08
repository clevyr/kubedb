package kubernetes

import (
	"context"
	"errors"
	"fmt"
	"strings"

	corev1 "k8s.io/api/core/v1"
	v1meta "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var (
	ErrEnvVarNotFound  = errors.New("env var not found")
	ErrNoDiscoveryEnvs = errors.New("failed to find config")
)

type ConfigFinder interface {
	GetValue(context.Context, KubeClient, corev1.Pod) (string, error)
}

type ConfigFinders []ConfigFinder

func (c ConfigFinders) Search(ctx context.Context, client KubeClient, pod corev1.Pod) (string, error) {
	for _, search := range c {
		found, err := search.GetValue(ctx, client, pod)
		if err == nil {
			return found, nil
		}
	}
	return "", ErrNoDiscoveryEnvs
}

type ConfigFromEnv []string

func (e ConfigFromEnv) GetValue(ctx context.Context, client KubeClient, pod corev1.Pod) (string, error) {
	if len(e) == 0 {
		return "", ErrNoEnvNames
	}

	var err error
	var envVar *corev1.EnvVar
	for _, envName := range e {
		envVar, err = FindEnvVar(pod, envName)
		if err == nil {
			break
		}
	}
	if err != nil {
		if errors.Is(err, ErrEnvVarNotFound) {
			return "", fmt.Errorf("%w: %s", ErrNoDiscoveryEnvs, strings.Join(e, ", "))
		}
		return "", err
	}

	if envVar.ValueFrom != nil {
		switch {
		case envVar.ValueFrom.SecretKeyRef != nil:
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
		case envVar.ValueFrom.ConfigMapKeyRef != nil:
			configMapRef := envVar.ValueFrom.ConfigMapKeyRef
			configMap, err := client.ConfigMaps().Get(ctx, configMapRef.Name, v1meta.GetOptions{})
			if err != nil {
				return "", err
			}
			data, ok := configMap.Data[configMapRef.Key]
			if !ok {
				return "", fmt.Errorf("%w: %v", ErrConfigMapDoesNotHaveKey, configMapRef)
			}
			return data, nil
		}
	}
	return envVar.Value, nil
}

type ConfigFromVolumeSecret struct {
	Name string
	Key  string
}

func (f ConfigFromVolumeSecret) GetValue(ctx context.Context, client KubeClient, pod corev1.Pod) (string, error) {
	if f.Name == "" || f.Key == "" {
		return "", ErrNoEnvNames
	}

	for _, volume := range pod.Spec.Volumes {
		if volume.Name == f.Name && volume.Secret != nil {
			secret, err := client.Secrets().Get(ctx, volume.Secret.SecretName, v1meta.GetOptions{})
			if err != nil {
				return "", err
			}

			if value, ok := secret.Data[f.Key]; ok {
				return string(value), nil
			}
		}
	}

	return "", ErrNoDiscoveryEnvs
}
