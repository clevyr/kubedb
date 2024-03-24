package filter

import (
	v1 "k8s.io/api/core/v1"
)

type And []Filter

func (filters And) Matches(pod v1.Pod) bool {
	for _, filter := range filters {
		if !filter.Matches(pod) {
			return false
		}
	}
	return true
}
