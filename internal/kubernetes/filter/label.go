package filter

import (
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/selection"
)

type Label struct {
	Name     string
	Operator selection.Operator
	Value    string
	Values   []string
}

func (label Label) Matches(pod corev1.Pod) bool {
	if label.Operator == "" {
		label.Operator = selection.In
	}

	if label.Value != "" {
		if len(label.Values) != 0 {
			panic("kubernetes.Label selector Value and Values are mutually exclusive")
		}

		label.Values = []string{label.Value}
	}

	r, err := labels.NewRequirement(label.Name, label.Operator, label.Values)
	if err != nil {
		panic(err)
	}

	return r.Matches(labels.Set(pod.Labels))
}
