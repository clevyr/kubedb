package filter

import (
	corev1 "k8s.io/api/core/v1"
)

func stubPod() corev1.Pod {
	return corev1.Pod{
		Labels: map[string]string{
			"key":  "value",
			"key2": "value2",
		},
	}
}
