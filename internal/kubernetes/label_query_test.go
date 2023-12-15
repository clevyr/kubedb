package kubernetes

import (
	"testing"

	"github.com/stretchr/testify/assert"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var pod = v1.Pod{
	ObjectMeta: metav1.ObjectMeta{
		Labels: map[string]string{
			"key":  "value",
			"key2": "value2",
		},
	},
}

var pods = []v1.Pod{pod}

func TestLabelQuery_FindPods(t *testing.T) {
	type fields struct {
		Name  string
		Value string
	}
	type args struct {
		list []v1.Pod
	}
	tests := []struct {
		name     string
		fields   fields
		args     args
		wantPods []v1.Pod
	}{
		{"1 found", fields{"key", "value"}, args{pods}, []v1.Pod{pod}},
		{"0 found", fields{"key", "wrong"}, args{pods}, []v1.Pod{}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			query := LabelQuery{
				Name:  tt.fields.Name,
				Value: tt.fields.Value,
			}
			assert.Equal(t, tt.wantPods, query.FindPods(tt.args.list))
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
			got := query.Matches(tt.args.pod)
			assert.Equal(t, tt.want, got)
		})
	}
}
