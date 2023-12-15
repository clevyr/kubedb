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

func (query LabelQuery) FindPods(pods []v1.Pod) []v1.Pod {
	matched := make([]v1.Pod, 0, len(pods))

	for _, pod := range pods {
		if query.Matches(pod) {
			matched = append(matched, pod)
		}
	}

	return matched
}
