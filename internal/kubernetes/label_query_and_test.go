package kubernetes

import (
	"testing"

	"github.com/stretchr/testify/assert"
	v1 "k8s.io/api/core/v1"
)

func TestLabelQueryAnd_FindPods(t *testing.T) {
	type args struct {
		list []v1.Pod
	}
	tests := []struct {
		name     string
		queries  LabelQueryAnd
		args     args
		wantPods []v1.Pod
	}{
		{
			"1 found",
			LabelQueryAnd{
				LabelQuery{Name: "key", Value: "value"},
				LabelQuery{Name: "key2", Value: "value2"},
			},
			args{pods},
			[]v1.Pod{pod},
		},
		{
			"0 found",
			LabelQueryAnd{
				LabelQuery{Name: "key", Value: "value"},
				LabelQuery{Name: "key2", Value: "wrong"},
			},
			args{pods},
			[]v1.Pod{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.wantPods, tt.queries.FindPods(tt.args.list))
		})
	}
}

func TestLabelQueryAnd_Matches(t *testing.T) {
	type args struct {
		pod v1.Pod
	}
	tests := []struct {
		name    string
		queries LabelQueryAnd
		args    args
		want    bool
	}{
		{
			"1 found",
			LabelQueryAnd{
				LabelQuery{Name: "key", Value: "value"},
				LabelQuery{Name: "key2", Value: "value2"},
			},
			args{pod},
			true,
		},
		{
			"0 found",
			LabelQueryAnd{
				LabelQuery{Name: "key", Value: "value"},
				LabelQuery{Name: "key2", Value: "wrong"},
			},
			args{pod},
			false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.queries.Matches(tt.args.pod)
			assert.Equal(t, tt.want, got)
		})
	}
}
