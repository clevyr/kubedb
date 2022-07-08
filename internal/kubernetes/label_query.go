package kubernetes

import (
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

func (query LabelQuery) FindPods(list *v1.PodList) (pods []v1.Pod, err error) {
	for _, pod := range list.Items {
		if query.Matches(pod) {
			pods = append(pods, pod)
		}
	}

	if len(pods) == 0 {
		err = ErrPodNotFound
	}
	return pods, err
}
