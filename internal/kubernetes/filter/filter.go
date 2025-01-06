package filter

import corev1 "k8s.io/api/core/v1"

type Filter interface {
	Matches(pod corev1.Pod) bool
}

func Pods(pods []corev1.Pod, filter Filter) []corev1.Pod {
	matched := make([]corev1.Pod, 0, len(pods))

	for _, pod := range pods {
		if filter.Matches(pod) {
			matched = append(matched, pod)
		}
	}

	return matched
}
