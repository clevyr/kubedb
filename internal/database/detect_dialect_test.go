package database

import (
	"testing"

	"github.com/clevyr/kubedb/internal/database/mariadb"
	"github.com/clevyr/kubedb/internal/database/postgres"
	"github.com/clevyr/kubedb/internal/kubernetes"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	kubernetesfake "k8s.io/client-go/kubernetes/fake"
)

func TestDetectDialect(t *testing.T) {
	postgresPod := corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Labels: map[string]string{
				"app.kubernetes.io/name":      "postgresql",
				"app.kubernetes.io/component": "primary",
			},
		},
	}

	mariadbPod := corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Labels: map[string]string{
				"app.kubernetes.io/name":      "mariadb",
				"app.kubernetes.io/component": "primary",
			},
		},
	}

	type args struct {
		client kubernetes.KubeClient
	}
	tests := []struct {
		name    string
		args    args
		want    []DetectResult
		wantErr require.ErrorAssertionFunc
	}{
		{
			"postgres",
			args{
				kubernetes.KubeClient{
					ClientSet: kubernetesfake.NewClientset(&postgresPod),
				},
			},
			[]DetectResult{{postgres.Postgres{}, []corev1.Pod{postgresPod}}},
			require.NoError,
		},
		{
			"mariadb",
			args{
				kubernetes.KubeClient{
					ClientSet: kubernetesfake.NewClientset(&mariadbPod),
				},
			},
			[]DetectResult{{mariadb.MariaDB{}, []corev1.Pod{mariadbPod}}},
			require.NoError,
		},
		{
			"no database",
			args{
				kubernetes.KubeClient{
					ClientSet: kubernetesfake.NewClientset(&corev1.Pod{}),
				},
			},
			nil,
			require.Error,
		},
		{
			"no pods in namespace",
			args{
				kubernetes.KubeClient{
					ClientSet: kubernetesfake.NewClientset(),
				},
			},
			nil,
			require.Error,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := DetectDialect(t.Context(), tt.args.client)
			tt.wantErr(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}
