package filter

import (
	"testing"

	"github.com/stretchr/testify/assert"
	v1 "k8s.io/api/core/v1"
)

func TestLabelQueryOr_Matches(t *testing.T) {
	t.Parallel()
	type args struct {
		pod v1.Pod
	}
	tests := []struct {
		name    string
		queries Or
		args    args
		want    bool
	}{
		{
			"1 found",
			Or{
				Label{Name: "key", Value: "wrong"},
				Label{Name: "key2", Value: "value2"},
			},
			args{stubPod()},
			true,
		},
		{
			"0 found",
			Or{
				Label{Name: "key", Value: "wrong"},
				Label{Name: "key2", Value: "also wrong"},
			},
			args{stubPod()},
			false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := tt.queries.Matches(tt.args.pod)
			assert.Equal(t, tt.want, got)
		})
	}
}
