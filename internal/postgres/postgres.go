package postgres

import (
	"context"
	"github.com/clevyr/kubedb/internal/kubernetes"
	v1core "k8s.io/api/core/v1"
	v1meta "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/selection"
)

func GetPod(client kubernetes.KubeClient) (v1core.Pod, error) {
	return kubernetes.GetPodByLabel(client, "app", selection.Equals, []string{"postgresql"})
}

func GetSecret(client kubernetes.KubeClient) (string, error) {
	secret, err := client.Secrets().Get(context.TODO(), "postgresql", v1meta.GetOptions{})
	if err != nil {
		return "", err
	}
	return string(secret.Data["postgresql-password"]), err
}
