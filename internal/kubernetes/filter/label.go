package filter

import (
	"github.com/rs/zerolog/log"
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
		log.Panic().Str("op", string(label.Operator)).Msg("Filter operator not implemented")
	}
	return false
}
