package kubernetes

import (
	"testing"

	"github.com/stretchr/testify/assert"
	v1 "k8s.io/api/core/v1"
)

func TestLabelQueryOr_FindPods(t *testing.T) {
	type args struct {
		list []v1.Pod
	}
	tests := []struct {
		name     string
		queries  LabelQueryOr
		args     args
		wantPods []v1.Pod
	}{
		{
			"1 found",
			LabelQueryOr{
				{Name: "key", Value: "wrong"},
				{Name: "key2", Value: "value2"},
			},
			args{pods},
			[]v1.Pod{pod},
		},
		{
			"0 found",
			LabelQueryOr{
				{Name: "key", Value: "wrong"},
				{Name: "key2", Value: "also wrong"},
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

func TestLabelQueryOr_Matches(t *testing.T) {
	type args struct {
		pod v1.Pod
	}
	tests := []struct {
		name    string
		queries LabelQueryOr
		args    args
		want    bool
	}{
		{
			"1 found",
			LabelQueryOr{
				{Name: "key", Value: "wrong"},
				{Name: "key2", Value: "value2"},
			},
			args{pod},
			true,
		},
		{
			"0 found",
			LabelQueryOr{
				{Name: "key", Value: "wrong"},
				{Name: "key2", Value: "also wrong"},
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
