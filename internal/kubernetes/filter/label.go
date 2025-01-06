package filter

import (
	"log/slog"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/selection"
)

type Label struct {
	Name     string
	Operator selection.Operator
	Value    string
}

func (label Label) Matches(pod corev1.Pod) bool {
	labelValue, ok := pod.Labels[label.Name]
	switch label.Operator {
	case selection.Exists:
		return ok
	case "":
		return ok && labelValue == label.Value
	default:
		slog.Error("Filter operator not implemented", "op", string(label.Operator))
		panic("Filter operator not implemented")
	}
}
