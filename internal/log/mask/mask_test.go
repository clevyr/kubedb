package mask

import (
	"log/slog"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMasker_Add(t *testing.T) {
	masker := Masker{}
	masker.Add("test")
	assert.Equal(t, []string{"test"}, masker.masks)
	masker.Add("another")
	assert.Equal(t, []string{"test", "another"}, masker.masks)
}

func TestMasker_MaskAttr(t *testing.T) {
	masker := &Masker{masks: []string{"test"}}
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
			got := masker.MaskAttr(nil, tt.args.attr)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestMasker_replace(t *testing.T) {
	masker := &Masker{masks: []string{"test"}}

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
			got, changed := masker.replace(tt.args.str)
			assert.Equal(t, tt.want, got)
			assert.Equal(t, tt.want1, changed)
		})
	}
}
