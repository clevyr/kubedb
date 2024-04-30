package config

import (
	"io"
	"os"
	"testing"

	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_LogFormat(t *testing.T) {
	t.Parallel()
	type args struct {
		format string
	}
	tests := []struct {
		name string
		args args
		want io.Writer
	}{
		{"default", args{"auto"}, zerolog.ConsoleWriter{Out: os.Stderr, NoColor: true}},
		{"color", args{"color"}, zerolog.ConsoleWriter{Out: os.Stderr}},
		{"plain", args{"plain"}, zerolog.ConsoleWriter{Out: os.Stderr, NoColor: true}},
		{"json", args{"json"}, os.Stderr},
		{"unknown", args{""}, zerolog.ConsoleWriter{Out: os.Stderr, NoColor: true}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := LogFormat(os.Stderr, tt.args.format)
			require.IsType(t, tt.want, got)
			if want, ok := tt.want.(zerolog.ConsoleWriter); ok {
				got := got.(zerolog.ConsoleWriter)
				assert.NotNil(t, got.FormatMessage)
				want.FormatMessage = got.FormatMessage
				assert.Equal(t, want.Out, got.Out)
				assert.Equal(t, want.NoColor, got.NoColor)
			}
		})
	}
}

func Test_LogLevel(t *testing.T) {
	t.Parallel()
	type args struct {
		level string
	}
	tests := []struct {
		name string
		args args
		want zerolog.Level
	}{
		{"trace", args{"trace"}, zerolog.TraceLevel},
		{"debug", args{"debug"}, zerolog.DebugLevel},
		{"info", args{"info"}, zerolog.InfoLevel},
		{"warning", args{"warning"}, zerolog.WarnLevel},
		{"error", args{"error"}, zerolog.ErrorLevel},
		{"fatal", args{"fatal"}, zerolog.FatalLevel},
		{"panic", args{"panic"}, zerolog.PanicLevel},
		{"unknown", args{""}, zerolog.InfoLevel},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := LogLevel(tt.args.level)
			assert.Equal(t, tt.want, got)
		})
	}
}
