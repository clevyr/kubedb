package sqlformat

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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
		wantErr require.ErrorAssertionFunc
	}{
		{"gzip", Format(0), args{"gzip"}, require.NoError},
		{"gz", Format(0), args{"gz"}, require.NoError},
		{"g", Format(0), args{"g"}, require.NoError},
		{"plain", Format(0), args{"plain"}, require.NoError},
		{"sql", Format(0), args{"sql"}, require.NoError},
		{"p", Format(0), args{"p"}, require.NoError},
		{"custom", Format(0), args{"custom"}, require.NoError},
		{"c", Format(0), args{"c"}, require.NoError},
		{"png", Format(0), args{"png"}, require.Error},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			err := tt.i.Set(tt.args.s)
			tt.wantErr(t, err)
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
