package kubernetes

import (
	"context"
	"errors"
	"fmt"
	"path"
	"strings"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type ConfigLookup interface {
	GetValue(ctx context.Context, client KubeClient, pod corev1.Pod) (string, error)
}

type ConfigLookups []ConfigLookup

func (c ConfigLookups) Search(ctx context.Context, client KubeClient, pod corev1.Pod) (string, error) {
	if len(c) == 0 {
		return "", nil
	}

	errs := make([]error, 0, len(c))
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

//nolint:gocognit,funlen
func (e LookupEnv) GetValue(ctx context.Context, client KubeClient, pod corev1.Pod) (string, error) {
	if len(e) == 0 {
		return "", ErrNoEnvNames
	}

	for _, lookupName := range e {
		for _, container := range pod.Spec.Containers {
			for _, env := range container.Env {
				switch env.Name {
				case lookupName:
					if env.Value != "" {
						return env.Value, nil
					}
					if env.ValueFrom != nil {
						if env.ValueFrom.SecretKeyRef != nil {
							secretKeyRef := env.ValueFrom.SecretKeyRef
							secret, err := client.Secrets().Get(ctx, secretKeyRef.Name, metav1.GetOptions{})
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
							configMap, err := client.ConfigMaps().Get(ctx, configMapRef.Name, metav1.GetOptions{})
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
				case lookupName + "_FILE":
					if env.Value != "" {
						base, key := path.Split(env.Value)
						base = path.Clean(base)
						for _, volume := range container.VolumeMounts {
							mountPath := path.Clean(volume.MountPath)
							if mountPath == base || (volume.SubPath == key && mountPath == path.Clean(env.Value)) {
								return LookupSecretVolume{Name: volume.Name, Key: key}.GetValue(ctx, client, pod)
							}
						}
					}
				}
			}

			for _, source := range container.EnvFrom {
				if source.SecretRef != nil {
					secret, err := client.Secrets().Get(ctx, source.SecretRef.Name, metav1.GetOptions{})
					if err != nil {
						return "", err
					}
					data, ok := secret.Data[lookupName]
					if ok {
						return string(data), nil
					}
				}
				if source.ConfigMapRef != nil {
					configMap, err := client.ConfigMaps().Get(ctx, source.ConfigMapRef.Name, metav1.GetOptions{})
					if err != nil {
						return "", err
					}
					data, ok := configMap.Data[lookupName]
					if ok {
						return data, nil
					}
				}
			}
		}
	}

	return "", fmt.Errorf("%w: %s", ErrEnvNoExist, strings.Join(e, ", "))
}

type LookupNamedSecret struct {
	Name string
	Key  string
}

func (f LookupNamedSecret) GetValue(ctx context.Context, client KubeClient, _ corev1.Pod) (string, error) {
	if f.Name == "" || f.Key == "" {
		return "", ErrNoEnvNames
	}

	secret, err := client.Secrets().Get(ctx, f.Name, metav1.GetOptions{})
	if err != nil {
		return "", err
	}

	if value, ok := secret.Data[f.Key]; ok {
		return string(value), nil
	}
	return "", fmt.Errorf("%w: %s", ErrSecretDoesNotHaveKey, f.Key)
}

type LookupSecretVolume struct {
	Name string
	Key  string
}

var (
	ErrNotASecretVolume = errors.New("matched volume is not a secret")
	ErrNoSecretVolume   = errors.New("secret volume does not exist")
)

func (f LookupSecretVolume) GetValue(ctx context.Context, client KubeClient, pod corev1.Pod) (string, error) {
	if f.Name == "" || f.Key == "" {
		return "", ErrNoEnvNames
	}

	var secretName string
	for _, volume := range pod.Spec.Volumes {
		if volume.Name == f.Name {
			if volume.Secret == nil {
				return "", fmt.Errorf("%w: %v", ErrNotASecretVolume, f.Name)
			}
			secretName = volume.Secret.SecretName
			break
		}
	}
	if secretName == "" {
		return "", fmt.Errorf("%w: %v", ErrNoSecretVolume, f.Name)
	}

	return LookupNamedSecret{Name: secretName, Key: f.Key}.GetValue(ctx, client, pod)
}

type LookupDefault string

func (l LookupDefault) GetValue(context.Context, KubeClient, corev1.Pod) (string, error) {
	return string(l), nil
}
