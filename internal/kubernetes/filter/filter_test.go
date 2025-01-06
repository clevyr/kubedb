package filter

import (
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func stubPod() corev1.Pod {
	return corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Labels: map[string]string{
				"key":  "value",
				"key2": "value2",
			},
		},
	}
}
