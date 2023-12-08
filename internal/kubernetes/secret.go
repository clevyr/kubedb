package kubernetes

import (
	"errors"
	"fmt"

	log "github.com/sirupsen/logrus"
	v1 "k8s.io/api/core/v1"
)

func FindEnvVar(pod v1.Pod, envName string) (*v1.EnvVar, error) {
	for _, container := range pod.Spec.Containers {
		for _, env := range container.Env {
			if env.Name == envName {
				return &env, nil
			}
		}
	}
	log.WithField("name", envName).Trace("failed to find env")
	return nil, fmt.Errorf("%w: %s", ErrEnvVarNotFound, envName)
}

var (
	ErrNoEnvNames              = errors.New("dialect does not contain any env names")
	ErrSecretDoesNotHaveKey    = errors.New("secret does not have key")
	ErrConfigMapDoesNotHaveKey = errors.New("config map does not have key")
)
