package kubernetes

import (
	"fmt"
	v1 "k8s.io/api/core/v1"
)

type LabelQueryAnd []LabelQuery

func (queries LabelQueryAnd) Matches(pod v1.Pod) bool {
	for _, query := range queries {
		if query.Matches(pod) {
			return true
		}
	}
	return false
}

func (queries LabelQueryAnd) FindPods(list *v1.PodList) (pods []v1.Pod, err error) {
	for _, pod := range list.Items {
		var match bool
		for _, query := range queries {
			if query.Matches(pod) {
				match = true
			} else {
				match = false
				break
			}
		}
		if match {
			pods = append(pods, pod)
		}
	}

	if len(pods) == 0 {
		err = fmt.Errorf("%w: %v", ErrPodNotFound, queries)
	}
	return pods, err
}
