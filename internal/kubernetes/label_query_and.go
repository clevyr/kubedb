package kubernetes

import (
	v1 "k8s.io/api/core/v1"
)

type LabelQueryAnd []LabelQueryable

func (queries LabelQueryAnd) Matches(pod v1.Pod) bool {
	for _, query := range queries {
		if !query.Matches(pod) {
			return false
		}
	}
	return true
}

func (queries LabelQueryAnd) FindPods(pods []v1.Pod) []v1.Pod {
	matched := make([]v1.Pod, 0, len(pods))

	for _, pod := range pods {
		if queries.Matches(pod) {
			matched = append(matched, pod)
		}
	}

	return matched
}
