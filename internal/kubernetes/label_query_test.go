package kubernetes

import (
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"reflect"
	"testing"
)

var pod = v1.Pod{
	ObjectMeta: metav1.ObjectMeta{
		Labels: map[string]string{
			"key":  "value",
			"key2": "value2",
		},
	},
}

var podList = v1.PodList{Items: []v1.Pod{pod}}

func TestLabelQuery_FindPods(t *testing.T) {
	type fields struct {
		Name  string
		Value string
	}
	type args struct {
		list *v1.PodList
	}
	tests := []struct {
		name     string
		fields   fields
		args     args
		wantPods []v1.Pod
		wantErr  bool
	}{
		{"1 found", fields{"key", "value"}, args{&podList}, []v1.Pod{pod}, false},
		{"0 found", fields{"key", "wrong"}, args{&podList}, nil, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			query := LabelQuery{
				Name:  tt.fields.Name,
				Value: tt.fields.Value,
			}
			gotPods, err := query.FindPods(tt.args.list)
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

func TestLabelQuery_Matches(t *testing.T) {
	type fields struct {
		Name  string
		Value string
	}
	type args struct {
		pod v1.Pod
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   bool
	}{
		{"1 found", fields{"key", "value"}, args{pod}, true},
		{"0 found", fields{"key", "wrong"}, args{pod}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			query := LabelQuery{
				Name:  tt.fields.Name,
				Value: tt.fields.Value,
			}
			if got := query.Matches(tt.args.pod); got != tt.want {
				t.Errorf("Matches() = %v, want %v", got, tt.want)
			}
		})
	}
}
