package kubernetes

import (
	"context"
	"errors"
	"fmt"
	"strings"

	corev1 "k8s.io/api/core/v1"
	v1meta "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type ConfigLookup interface {
	GetValue(context.Context, KubeClient, corev1.Pod) (string, error)
}

type ConfigLookups []ConfigLookup

func (c ConfigLookups) Search(ctx context.Context, client KubeClient, pod corev1.Pod) (string, error) {
	var errs []error
	for _, search := range c {
		found, err := search.GetValue(ctx, client, pod)
		if err == nil {
			return found, nil
		}
		errs = append(errs, err)
	}
	return "", errors.Join(errs...)
}

type LookupEnv []string

var (
	ErrNoEnvNames              = errors.New("dialect does not contain any env names")
	ErrEnvNoExist              = errors.New("env is not set")
	ErrSecretDoesNotHaveKey    = errors.New("secret does not have key")
	ErrConfigMapDoesNotHaveKey = errors.New("config map does not have key")
)

func (e LookupEnv) GetValue(ctx context.Context, client KubeClient, pod corev1.Pod) (string, error) {
	if len(e) == 0 {
		return "", ErrNoEnvNames
	}

	for _, envName := range e {
		for _, container := range pod.Spec.Containers {
			for _, env := range container.Env {
				if env.Name == envName {
					if env.Value != "" {
						return env.Value, nil
					}
					if env.ValueFrom != nil {
						if env.ValueFrom.SecretKeyRef != nil {
							secretKeyRef := env.ValueFrom.SecretKeyRef
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
						if env.ValueFrom.ConfigMapKeyRef != nil {
							configMapRef := env.ValueFrom.ConfigMapKeyRef
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
				}
			}

			for _, source := range container.EnvFrom {
				if source.SecretRef != nil {
					secret, err := client.Secrets().Get(ctx, source.SecretRef.Name, v1meta.GetOptions{})
					if err != nil {
						return "", err
					}
					data, ok := secret.Data[envName]
					if ok {
						return string(data), nil
					}
				}
				if source.ConfigMapRef != nil {
					configMap, err := client.ConfigMaps().Get(ctx, source.ConfigMapRef.Name, v1meta.GetOptions{})
					if err != nil {
						return "", err
					}
					data, ok := configMap.Data[envName]
					if ok {
						return data, nil
					}
				}
			}
		}
	}

	return "", fmt.Errorf("%w: %s", ErrEnvNoExist, strings.Join(e, ", "))
}

type LookupVolumeSecret struct {
	Name string
	Key  string
}

var ErrVolumeNoExist = errors.New("volume does not exist")

func (f LookupVolumeSecret) GetValue(ctx context.Context, client KubeClient, pod corev1.Pod) (string, error) {
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
			} else {
				return "", fmt.Errorf("%w: %s", ErrSecretDoesNotHaveKey, f.Key)
			}
		}
	}

	return "", fmt.Errorf("%w: %s", ErrVolumeNoExist, f.Name)
}
