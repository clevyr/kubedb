package log

import (
	"io"
	"log/slog"
	"time"

	"gabe565.com/utils/slogx"
	"gabe565.com/utils/termx"
	"github.com/clevyr/kubedb/internal/config"
	"github.com/clevyr/kubedb/internal/log/mask"
	"github.com/clevyr/kubedb/internal/tui"
	"github.com/lmittmann/tint"
	"github.com/spf13/cobra"
)

func InitGlobal(cmd *cobra.Command) {
	Init(cmd.ErrOrStderr(), config.Global.Log.Level, config.Global.Log.Format)
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
			color = termx.IsColor(w)
		case slogx.FormatColor:
			color = true
		}

		tui.InitRenderer(format)

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
