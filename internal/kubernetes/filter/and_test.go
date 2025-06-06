package filter

import (
	"testing"

	"github.com/stretchr/testify/assert"
	corev1 "k8s.io/api/core/v1"
)

func TestAnd_Matches(t *testing.T) {
	type args struct {
		pod corev1.Pod
	}
	tests := []struct {
		name    string
		queries And
		args    args
		want    bool
	}{
		{
			"1 found",
			And{
				Label{Name: "key", Value: "value"},
				Label{Name: "key2", Value: "value2"},
			},
			args{stubPod()},
			true,
		},
		{
			"0 found",
			And{
				Label{Name: "key", Value: "value"},
				Label{Name: "key2", Value: "wrong"},
			},
			args{stubPod()},
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
