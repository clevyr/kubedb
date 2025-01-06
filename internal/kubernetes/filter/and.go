package filter

import (
	corev1 "k8s.io/api/core/v1"
)

type And []Filter

func (filters And) Matches(pod corev1.Pod) bool {
	for _, filter := range filters {
		if !filter.Matches(pod) {
			return false
		}
	}
	return true
}
