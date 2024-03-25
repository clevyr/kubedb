package filter

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

func TestLabel_Matches(t *testing.T) {
	t.Parallel()
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
			t.Parallel()
			query := Label{
				Name:  tt.fields.Name,
				Value: tt.fields.Value,
			}
			got := query.Matches(tt.args.pod)
			assert.Equal(t, tt.want, got)
		})
	}
}
