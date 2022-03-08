package kubernetes

import (
	"fmt"
	v1 "k8s.io/api/core/v1"
)

type LabelQuery struct {
	Name  string
	Value string
}

func (query LabelQuery) Matches(pod v1.Pod) bool {
	labelValue, ok := pod.Labels[query.Name]
	if ok && labelValue == query.Value {
		return true
	}
	return false
}

func (query LabelQuery) FindPod(list *v1.PodList) (*v1.Pod, error) {
	for _, pod := range list.Items {
		if query.Matches(pod) {
			return &pod, nil
		}
	}
	return nil, fmt.Errorf("%w: %v", ErrPodNotFound, query)
}
