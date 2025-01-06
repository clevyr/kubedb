package filter

import (
	corev1 "k8s.io/api/core/v1"
)

type Or []Filter

func (filters Or) Matches(pod corev1.Pod) bool {
	for _, filter := range filters {
		if filter.Matches(pod) {
			return true
		}
	}
	return false
}
