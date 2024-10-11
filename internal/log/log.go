package log

import (
	"io"
	"log/slog"
	"strconv"
	"strings"
	"time"

	"github.com/clevyr/kubedb/internal/consts"
	"github.com/lmittmann/tint"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"k8s.io/kubectl/pkg/util/term"
)

const LevelTrace slog.Level = -5

//go:generate go run github.com/dmarkham/enumer -type Format -trimprefix Format -transform lower -text

type Format uint8

const (
	FormatAuto Format = iota
	FormatColor
	FormatPlain
	FormatJSON
)

func InitFromCmd(cmd *cobra.Command) (slog.Level, Format) {
	var level slog.Level
	levelStr := viper.GetString(consts.LogLevelKey)
	if val, err := strconv.Atoi(levelStr); err == nil {
		level = slog.Level(val)
	} else if err := level.UnmarshalText([]byte(levelStr)); err != nil {
		switch levelStr {
		case "trace":
			level = LevelTrace
		default:
			defer func() {
				slog.Warn("Invalid log level. Defaulting to info.", "value", levelStr)
			}()
			level = slog.LevelInfo
			viper.Set(consts.LogLevelKey, level.String())
		}
	}

	var format Format
	formatStr := viper.GetString(consts.LogFormatKey)
	if err := format.UnmarshalText([]byte(formatStr)); err != nil {
		defer func() {
			slog.Warn("Invalid log format. Defaulting to auto.", "value", formatStr)
		}()
		format = FormatAuto
		viper.Set(consts.LogFormatKey, format.String())
	}

	Init(cmd.ErrOrStderr(), level, format)
	return level, format
}

func Init(w io.Writer, level slog.Level, format Format) {
	switch format {
	case FormatJSON:
		slog.SetDefault(slog.New(slog.NewJSONHandler(w, &slog.HandlerOptions{
			Level:       level,
			ReplaceAttr: MaskAttr,
		})))
	default:
		var color bool
		switch format {
		case FormatAuto:
			color = term.TTY{Out: w}.IsTerminalOut()
		case FormatColor:
			color = true
		}

		slog.SetDefault(slog.New(
			tint.NewHandler(w, &tint.Options{
				Level:       level,
				TimeFormat:  time.Kitchen,
				NoColor:     !color,
				ReplaceAttr: MaskAttr,
			}),
		))
	}
}

func LevelStrings() []string {
	return []string{
		"trace",
		strings.ToLower(slog.LevelDebug.String()),
		strings.ToLower(slog.LevelInfo.String()),
		strings.ToLower(slog.LevelWarn.String()),
		strings.ToLower(slog.LevelError.String()),
	}
}
