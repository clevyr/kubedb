package filter

import (
	log "github.com/sirupsen/logrus"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/selection"
)

type Label struct {
	Name     string
	Operator selection.Operator
	Value    string
}

func (label Label) Matches(pod v1.Pod) bool {
	labelValue, ok := pod.Labels[label.Name]
	switch label.Operator {
	case selection.Exists:
		return ok
	case "":
		return ok && labelValue == label.Value
	default:
		log.Panicf("Filter operator not implemented: %q", label.Operator)
	}
	return false
}
