package filter

import (
	"testing"

	"github.com/stretchr/testify/assert"
	corev1 "k8s.io/api/core/v1"
)

func TestLabel_Matches(t *testing.T) {
	type fields struct {
		Name  string
		Value string
	}
	type args struct {
		pod corev1.Pod
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   bool
	}{
		{"1 found", fields{"key", "value"}, args{stubPod()}, true},
		{"0 found", fields{"key", "wrong"}, args{stubPod()}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			query := Label{
				Name:  tt.fields.Name,
				Value: tt.fields.Value,
			}
			got := query.Matches(tt.args.pod)
			assert.Equal(t, tt.want, got)
		})
	}
}
