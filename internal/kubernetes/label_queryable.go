package kubernetes

import v1 "k8s.io/api/core/v1"

type LabelQueryable interface {
	Matches(pod v1.Pod) bool
	FindPod(list *v1.PodList) (*v1.Pod, error)
}
