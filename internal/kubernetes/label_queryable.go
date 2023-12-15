package kubernetes

import v1 "k8s.io/api/core/v1"

type LabelQueryable interface {
	Matches(pod v1.Pod) bool
	FindPods(pods []v1.Pod) []v1.Pod
}
