package log

import (
	"io"
	"log/slog"
	"time"

	"gabe565.com/utils/slogx"
	"github.com/clevyr/kubedb/internal/consts"
	"github.com/clevyr/kubedb/internal/log/mask"
	"github.com/lmittmann/tint"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"k8s.io/kubectl/pkg/util/term"
)

func InitFromCmd(cmd *cobra.Command) error {
	var level slogx.Level
	if err := level.UnmarshalText([]byte(viper.GetString(consts.LogLevelKey))); err != nil {
		return err
	}

	var format slogx.Format
	if err := format.UnmarshalText([]byte(viper.GetString(consts.LogFormatKey))); err != nil {
		return err
	}

	Init(cmd.ErrOrStderr(), level, format)
	return nil
}

func Init(w io.Writer, level slogx.Level, format slogx.Format) {
	switch format {
	case slogx.FormatJSON:
		slog.SetDefault(slog.New(slog.NewJSONHandler(w, &slog.HandlerOptions{
			Level:       level,
			ReplaceAttr: mask.MaskAttr,
		})))
	default:
		var color bool
		switch format {
		case slogx.FormatAuto:
			color = term.TTY{Out: w}.IsTerminalOut()
		case slogx.FormatColor:
			color = true
		}

		slog.SetDefault(slog.New(
			tint.NewHandler(w, &tint.Options{
				Level:       level,
				TimeFormat:  time.Kitchen,
				NoColor:     !color,
				ReplaceAttr: mask.MaskAttr,
			}),
		))
	}
}
