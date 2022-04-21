package kubernetes

import (
	v1 "k8s.io/api/core/v1"
	"reflect"
	"testing"
)

func TestLabelQueryAnd_FindPods(t *testing.T) {
	type args struct {
		list *v1.PodList
	}
	tests := []struct {
		name     string
		queries  LabelQueryAnd
		args     args
		wantPods []v1.Pod
		wantErr  bool
	}{
		{
			"1 found",
			LabelQueryAnd{
				{"key", "value"},
				{"key2", "value2"},
			},
			args{&podList},
			[]v1.Pod{pod},
			false,
		},
		{
			"0 found",
			LabelQueryAnd{
				{"key", "value"},
				{"key2", "wrong"},
			},
			args{&podList},
			nil,
			true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotPods, err := tt.queries.FindPods(tt.args.list)
			if (err != nil) != tt.wantErr {
				t.Errorf("FindPods() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(gotPods, tt.wantPods) {
				t.Errorf("FindPods() gotPods = %v, want %v", gotPods, tt.wantPods)
			}
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
				{"key", "value"},
				{"key2", "value2"},
			},
			args{pod},
			true,
		},
		{
			"0 found",
			LabelQueryAnd{
				{"key", "value"},
				{"key2", "wrong"},
			},
			args{pod},
			false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.queries.Matches(tt.args.pod); got != tt.want {
				t.Errorf("Matches() = %v, want %v", got, tt.want)
			}
		})
	}
}
