package kubernetes

import (
	log "github.com/sirupsen/logrus"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/selection"
)

type LabelQuery struct {
	Name     string
	Operator selection.Operator
	Value    string
}

func (query LabelQuery) Matches(pod v1.Pod) bool {
	labelValue, ok := pod.Labels[query.Name]
	switch query.Operator {
	case selection.Exists:
		return ok
	case "":
		return ok && labelValue == query.Value
	default:
		log.Panicf("Query operator not implemented: %q", query.Operator)
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
