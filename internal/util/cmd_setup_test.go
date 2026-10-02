package util

import (
	"strings"
	"testing"

	"github.com/clevyr/kubedb/internal/config/conftypes"
	"github.com/clevyr/kubedb/internal/kubernetes"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes/fake"
	k8stesting "k8s.io/client-go/testing"
)

func Test_podPasswordRef(t *testing.T) {
	pod := func(containers ...string) corev1.Pod {
		var p corev1.Pod
		for _, name := range containers {
			p.Spec.Containers = append(p.Spec.Containers, corev1.Container{Name: name})
		}
		return p
	}

	tests := []struct {
		name  string
		pod   corev1.Pod
		found kubernetes.ConfigValue
		want  string
	}{
		{
			"env",
			pod("db"),
			kubernetes.ConfigValue{Container: "db", EnvName: "POSTGRES_PASSWORD"},
			`"${POSTGRES_PASSWORD}"`,
		},
		{
			"file",
			pod("db"),
			kubernetes.ConfigValue{Container: "db", File: "/secrets/pass word"},
			`"$(cat '/secrets/pass word')"`,
		},
		{"invalid env name", pod("db"), kubernetes.ConfigValue{Container: "db", EnvName: "PASS-WORD"}, ""},
		{"other container", pod("db"), kubernetes.ConfigValue{Container: "sidecar", EnvName: "PASSWORD"}, ""},
		{"multiple containers", pod("db", "sidecar"), kubernetes.ConfigValue{Container: "db", EnvName: "PASSWORD"}, ""},
		{"named secret", pod("db"), kubernetes.ConfigValue{Source: &corev1.EnvVarSource{}}, ""},
		{"not found", pod("db"), kubernetes.ConfigValue{}, ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, podPasswordRef(tt.pod, tt.found))
		})
	}
}

func Test_createNamedJob(t *testing.T) {
	tests := []struct {
		name      string
		conflicts int
		wantErr   require.ErrorAssertionFunc
	}{
		{"no conflict", 0, require.NoError},
		{"retries conflicts", 2, require.NoError},
		{"gives up", jobNameAttempts, require.Error},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			clientset := fake.NewClientset()
			var names []string
			clientset.PrependReactor("create", "jobs",
				func(action k8stesting.Action) (bool, runtime.Object, error) {
					job := action.(k8stesting.CreateAction).GetObject().(*batchv1.Job)
					names = append(names, job.Name)
					if len(names) <= tt.conflicts {
						return true, nil, apierrors.NewAlreadyExists(batchv1.Resource("jobs"), job.Name)
					}
					return false, nil, nil
				},
			)
			conf := &conftypes.Global{
				Client: kubernetes.KubeClient{ClientSet: clientset, Namespace: "test"}}
			secretRef := &corev1.SecretKeySelector{}

			got, err := createNamedJob(t.Context(), conf, &batchv1.Job{}, "kubedb-test-", secretRef)
			tt.wantErr(t, err)
			assert.Len(t, names, min(tt.conflicts+1, jobNameAttempts))
			for _, name := range names {
				assert.True(t, strings.HasPrefix(name, "kubedb-test-"))
				assert.Len(t, name, len("kubedb-test-")+5)
			}
			if err == nil {
				assert.Equal(t, names[len(names)-1], got.Name)
				assert.Equal(t, got.Name, secretRef.Name)
			}
		})
	}
}
