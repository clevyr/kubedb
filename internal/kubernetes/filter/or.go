package filter

import (
	v1 "k8s.io/api/core/v1"
)

type Or []Filter

func (filters Or) Matches(pod v1.Pod) bool {
	for _, filter := range filters {
		if filter.Matches(pod) {
			return true
		}
	}
	return false
}
