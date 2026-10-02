package kubernetes

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/client-go/kubernetes/fake"
)

func TestLookupEnv_GetValue(t *testing.T) {
	const namespace = "test"
	client := KubeClient{
		Namespace: namespace,
		ClientSet: fake.NewClientset(
			&corev1.Secret{
				Name: "creds", Namespace: namespace,
				Data: map[string][]byte{"password": []byte("secret"), "PASSWORD": []byte("from-env")},
			},
			&corev1.ConfigMap{
				Name: "config", Namespace: namespace,
				Data: map[string]string{"password": "configmap"},
			},
		),
	}
	secretRef := corev1.LocalObjectReference{Name: "creds"}
	configMapRef := corev1.LocalObjectReference{Name: "config"}

	pod := func(container corev1.Container) corev1.Pod {
		container.Name = "db"
		return corev1.Pod{Spec: corev1.PodSpec{
			Containers: []corev1.Container{container},
			Volumes: []corev1.Volume{{
				Name:   "secrets",
				Secret: &corev1.SecretVolumeSource{SecretName: "creds"},
			}},
		}}
	}

	tests := []struct {
		name    string
		pod     corev1.Pod
		want    ConfigValue
		wantErr require.ErrorAssertionFunc
	}{
		{
			"value",
			pod(corev1.Container{Env: []corev1.EnvVar{{Name: "PASSWORD", Value: "literal"}}}),
			ConfigValue{Value: "literal", Container: "db", EnvName: "PASSWORD"},
			require.NoError,
		},
		{
			"secret key ref",
			pod(corev1.Container{Env: []corev1.EnvVar{{
				Name: "PASSWORD",
				ValueFrom: &corev1.EnvVarSource{SecretKeyRef: &corev1.SecretKeySelector{
					LocalObjectReference: secretRef, Key: "password",
				}},
			}}}),
			ConfigValue{
				Value: "secret", Container: "db", EnvName: "PASSWORD",
				Source: &corev1.EnvVarSource{SecretKeyRef: &corev1.SecretKeySelector{
					LocalObjectReference: secretRef, Key: "password",
				}},
			},
			require.NoError,
		},
		{
			"config map key ref",
			pod(corev1.Container{Env: []corev1.EnvVar{{
				Name: "PASSWORD",
				ValueFrom: &corev1.EnvVarSource{ConfigMapKeyRef: &corev1.ConfigMapKeySelector{
					LocalObjectReference: configMapRef, Key: "password",
				}},
			}}}),
			ConfigValue{
				Value: "configmap", Container: "db", EnvName: "PASSWORD",
				Source: &corev1.EnvVarSource{ConfigMapKeyRef: &corev1.ConfigMapKeySelector{
					LocalObjectReference: configMapRef, Key: "password",
				}},
			},
			require.NoError,
		},
		{
			"file",
			pod(corev1.Container{
				Env:          []corev1.EnvVar{{Name: "PASSWORD_FILE", Value: "/secrets/password"}},
				VolumeMounts: []corev1.VolumeMount{{Name: "secrets", MountPath: "/secrets"}},
			}),
			ConfigValue{
				Value: "secret", Container: "db", File: "/secrets/password",
				Source: &corev1.EnvVarSource{SecretKeyRef: &corev1.SecretKeySelector{
					LocalObjectReference: secretRef, Key: "password",
				}},
			},
			require.NoError,
		},
		{
			"env from secret",
			pod(corev1.Container{EnvFrom: []corev1.EnvFromSource{{
				Prefix:    "DB_",
				SecretRef: &corev1.SecretEnvSource{LocalObjectReference: secretRef},
			}}}),
			ConfigValue{
				Value: "from-env", Container: "db", EnvName: "DB_PASSWORD",
				Source: &corev1.EnvVarSource{SecretKeyRef: &corev1.SecretKeySelector{
					LocalObjectReference: secretRef, Key: "PASSWORD",
				}},
			},
			require.NoError,
		},
		{"not found", pod(corev1.Container{}), ConfigValue{}, require.Error},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := LookupEnv{"PASSWORD"}.GetValue(t.Context(), client, tt.pod)
			tt.wantErr(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}
