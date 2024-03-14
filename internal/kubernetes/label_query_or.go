package kubernetes

import (
	v1 "k8s.io/api/core/v1"
)

type LabelQueryOr []LabelQueryable

func (queries LabelQueryOr) Matches(pod v1.Pod) bool {
	for _, query := range queries {
		if query.Matches(pod) {
			return true
		}
	}
	return false
}

func (queries LabelQueryOr) FindPods(pods []v1.Pod) []v1.Pod {
	matched := make([]v1.Pod, 0, len(pods))

	for _, pod := range pods {
		if queries.Matches(pod) {
			matched = append(matched, pod)
		}
	}

	return matched
}
