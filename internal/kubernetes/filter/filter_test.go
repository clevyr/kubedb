package filter

import (
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func stubPod() v1.Pod {
	return v1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Labels: map[string]string{
				"key":  "value",
				"key2": "value2",
			},
		},
	}
}
