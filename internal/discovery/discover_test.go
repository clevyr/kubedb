package discovery

import (
	"testing"

	"github.com/clevyr/kubedb/internal/database/postgres"
	"github.com/clevyr/kubedb/internal/kubernetes"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	kubernetesfake "k8s.io/client-go/kubernetes/fake"
)

func TestFindReplica(t *testing.T) {
	newPod := func(name, cluster, role string, ready bool) *corev1.Pod {
		status := corev1.ConditionFalse
		if ready {
			status = corev1.ConditionTrue
		}
		return &corev1.Pod{
			Name:      name,
			Namespace: "default",
			Labels:    map[string]string{"cnpg.io/cluster": cluster, "cnpg.io/instanceRole": role},
			Status: corev1.PodStatus{
				Phase:      corev1.PodRunning,
				Conditions: []corev1.PodCondition{{Type: corev1.PodReady, Status: status}},
			},
		}
	}

	primary := newPod("postgresql-1", "postgresql", "primary", true)
	notReady := newPod("postgresql-2", "postgresql", "replica", false)
	replica := newPod("postgresql-3", "postgresql", "replica", true)
	otherReplica := newPod("other-2", "other", "replica", true)

	tests := []struct {
		name   string
		pods   []*corev1.Pod
		want   string
		wantOk bool
	}{
		{"ready replica", []*corev1.Pod{primary, notReady, replica, otherReplica}, replica.Name, true},
		{"no ready replica", []*corev1.Pod{primary, notReady}, "", false},
		{"ignores other clusters", []*corev1.Pod{primary, notReady, otherReplica}, "", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			clientset := kubernetesfake.NewClientset()
			for _, pod := range tt.pods {
				_, err := clientset.CoreV1().Pods("default").Create(t.Context(), pod, metav1.CreateOptions{})
				require.NoError(t, err)
			}
			client := kubernetes.KubeClient{ClientSet: clientset, Namespace: "default"}

			got, ok := FindReplica(t.Context(), client, postgres.Postgres{}, *primary)
			assert.Equal(t, tt.wantOk, ok)
			assert.Equal(t, tt.want, got.Name)
		})
	}
}
