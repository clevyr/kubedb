package log

import (
	"log/slog"
	"os"
	"testing"

	"github.com/clevyr/kubedb/internal/consts"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
)

func TestInitFromCmd(t *testing.T) {
	tests := []struct {
		name       string
		levelStr   string
		formatStr  string
		wantLevel  slog.Level
		wantFormat Format
	}{
		{"info/auto", "info", "auto", slog.LevelInfo, FormatAuto},
		{"warn/json", "warn", "json", slog.LevelWarn, FormatJSON},
		{"trace/auto", "trace", "auto", LevelTrace, FormatAuto},
		{"numeric/auto", "-4", "auto", slog.LevelDebug, FormatAuto},
		{"invalid level", "abc", "json", slog.LevelInfo, FormatJSON},
		{"invalid format", "info", "abc", slog.LevelInfo, FormatAuto},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := &cobra.Command{}
			cmd.SetOut(os.Stderr)

			viper.Set(consts.LogLevelKey, tt.levelStr)
			viper.Set(consts.LogFormatKey, tt.formatStr)

			level, format := InitFromCmd(cmd)
			assert.Equal(t, tt.wantLevel, level)
			assert.Equal(t, tt.wantFormat, format)
		})
	}
}
