package log

import (
	"log/slog"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAddMask(t *testing.T) {
	masks = nil
	t.Cleanup(func() {
		masks = nil
	})
	AddMask("test")
	assert.Equal(t, []string{"test"}, masks)
	AddMask("another")
	assert.Equal(t, []string{"test", "another"}, masks)
}

func TestMaskAttr(t *testing.T) {
	masks = []string{"test"}
	t.Cleanup(func() {
		masks = nil
	})

	var buf strings.Builder
	buf.WriteString("test value")

	type args struct {
		attr slog.Attr
	}
	tests := []struct {
		name string
		args args
		want slog.Attr
	}{
		{"int", args{slog.Int("count", 0)}, slog.Int("count", 0)},
		{"nil", args{slog.Any("error", nil)}, slog.Any("error", nil)},
		{"string", args{slog.String("key", "test value")}, slog.String("key", "*** value")},
		{"stringer", args{slog.Any("key", &buf)}, slog.String("key", "*** value")},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := MaskAttr(nil, tt.args.attr)
			assert.Equal(t, tt.want, got)
		})
	}
}

func Test_replace(t *testing.T) {
	masks = []string{"test"}
	t.Cleanup(func() {
		masks = nil
	})

	type args struct {
		str string
	}
	tests := []struct {
		name  string
		args  args
		want  string
		want1 bool
	}{
		{"no match", args{"abc"}, "abc", false},
		{"match", args{"test value"}, "*** value", true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, changed := replace(tt.args.str)
			assert.Equal(t, tt.want, got)
			assert.Equal(t, tt.want1, changed)
		})
	}
}
