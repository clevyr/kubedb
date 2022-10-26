package database

import (
	"context"
	"github.com/clevyr/kubedb/internal/config"
	"github.com/clevyr/kubedb/internal/database/dialect"
	"github.com/clevyr/kubedb/internal/kubernetes"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	kubernetesfake "k8s.io/client-go/kubernetes/fake"
	"reflect"
	"testing"
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
		want    config.Databaser
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
			dialect.Postgres{},
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
			dialect.Postgres{},
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
			if (err != nil) != tt.wantErr {
				t.Errorf("DetectDialect() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("DetectDialect() got = %v, want %v", got, tt.want)
			}
			if !reflect.DeepEqual(got1, tt.want1) {
				t.Errorf("DetectDialect() got1 = %v, want %v", got1, tt.want1)
			}
		})
	}
}
