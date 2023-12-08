package database

import (
	"context"
	"testing"

	"github.com/clevyr/kubedb/internal/config"
	"github.com/clevyr/kubedb/internal/database/postgres"
	"github.com/clevyr/kubedb/internal/kubernetes"
	"github.com/stretchr/testify/assert"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	kubernetesfake "k8s.io/client-go/kubernetes/fake"
)

func TestDetectDialect(t *testing.T) {
	postgresPod := v1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Labels: map[string]string{
				"app.kubernetes.io/name":      "postgresql",
				"app.kubernetes.io/component": "primary",
			},
		},
	}

	mariadbPod := v1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Labels: map[string]string{
				"app.kubernetes.io/name":      "postgresql",
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
		want    config.Database
		want1   []v1.Pod
		wantErr bool
	}{
		{
			"postgres",
			args{
				kubernetes.KubeClient{
					ClientSet: kubernetesfake.NewSimpleClientset(&postgresPod),
				},
			},
			postgres.Postgres{},
			[]v1.Pod{postgresPod},
			false,
		},
		{
			"mariadb",
			args{
				kubernetes.KubeClient{
					ClientSet: kubernetesfake.NewSimpleClientset(&mariadbPod),
				},
			},
			postgres.Postgres{},
			[]v1.Pod{mariadbPod},
			false,
		},
		{
			"no database",
			args{
				kubernetes.KubeClient{
					ClientSet: kubernetesfake.NewSimpleClientset(&v1.Pod{}),
				},
			},
			nil,
			[]v1.Pod{},
			true,
		},
		{
			"no pods in namespace",
			args{
				kubernetes.KubeClient{
					ClientSet: kubernetesfake.NewSimpleClientset(),
				},
			},
			nil,
			[]v1.Pod{},
			true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, got1, err := DetectDialect(context.Background(), tt.args.client)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
			assert.Equal(t, tt.want, got)
			assert.Equal(t, tt.want1, got1)
		})
	}
}
