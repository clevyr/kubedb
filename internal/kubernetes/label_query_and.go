package kubernetes

import (
	v1 "k8s.io/api/core/v1"
)

type LabelQueryAnd []LabelQuery

func (queries LabelQueryAnd) Matches(pod v1.Pod) bool {
	for _, query := range queries {
		if !query.Matches(pod) {
			return false
		}
	}
	return true
}

func (queries LabelQueryAnd) FindPods(list *v1.PodList) (pods []v1.Pod, err error) {
	for _, pod := range list.Items {
		if queries.Matches(pod) {
			pods = append(pods, pod)
		}
	}

	if len(pods) == 0 {
		err = ErrPodNotFound
	}
	return pods, err
}
