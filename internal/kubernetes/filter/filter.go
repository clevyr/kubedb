package filter

import v1 "k8s.io/api/core/v1"

type Filter interface {
	Matches(pod v1.Pod) bool
}

func Pods(pods []v1.Pod, filter Filter) []v1.Pod {
	matched := make([]v1.Pod, 0, len(pods))

	for _, pod := range pods {
		if filter.Matches(pod) {
			matched = append(matched, pod)
		}
	}

	return matched
}
