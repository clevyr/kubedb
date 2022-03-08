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

func (queries LabelQueryAnd) FindPod(list *v1.PodList) (*v1.Pod, error) {
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
			return &pod, nil
		}
	}
	return nil, fmt.Errorf("%w: %v", ErrPodNotFound, queries)
}
