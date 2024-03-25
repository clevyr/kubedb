package sqlformat

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFormat_Set(t *testing.T) {
	t.Parallel()
	type args struct {
		s string
	}
	tests := []struct {
		name    string
		i       Format
		args    args
		wantErr bool
	}{
		{"gzip", Format(0), args{"gzip"}, false},
		{"gz", Format(0), args{"gz"}, false},
		{"g", Format(0), args{"g"}, false},
		{"plain", Format(0), args{"plain"}, false},
		{"sql", Format(0), args{"sql"}, false},
		{"p", Format(0), args{"p"}, false},
		{"custom", Format(0), args{"custom"}, false},
		{"c", Format(0), args{"c"}, false},
		{"png", Format(0), args{"png"}, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			err := tt.i.Set(tt.args.s)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestFormat_Type(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name string
		i    Format
		want string
	}{
		{"simple", Format(0), "string"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := tt.i.Type()
			assert.Equal(t, tt.want, got)
		})
	}
}
