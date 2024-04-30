package config

import (
	"fmt"
	"os"

	"github.com/charmbracelet/lipgloss"
	"github.com/clevyr/kubedb/internal/consts"
	"github.com/mattn/go-isatty"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func InitLog(cmd *cobra.Command) {
	logLevel := viper.GetString(consts.LogLevelKey)
	parsedLevel, err := zerolog.ParseLevel(logLevel)
	if err != nil {
		if logLevel == "warning" {
			parsedLevel = zerolog.WarnLevel
		} else {
			log.Warn().Str("level", logLevel).Msg("invalid log level. defaulting to info.")
			viper.Set(consts.LogLevelKey, zerolog.InfoLevel.String())
			parsedLevel = zerolog.InfoLevel
		}
	}
	zerolog.SetGlobalLevel(parsedLevel)

	logFormat := viper.GetString(consts.LogFormatKey)
	switch logFormat {
	case "text", "txt", "t":
		var useColor bool
		baseFormatter := func(i interface{}) string {
			return fmt.Sprintf("%-45s", i)
		}
		formatter := baseFormatter
		errOut := cmd.ErrOrStderr()
		if w, ok := errOut.(*os.File); ok {
			useColor = isatty.IsTerminal(w.Fd())
			if useColor {
				boldStyle := lipgloss.NewStyle().Bold(true)
				formatter = func(i interface{}) string {
					return boldStyle.Render(baseFormatter(i))
				}
			}
		}

		log.Logger = log.Output(zerolog.ConsoleWriter{
			Out:           errOut,
			NoColor:       !useColor,
			FormatMessage: formatter,
		})
	case "json", "j":
		// default
	default:
		log.Warn().Str("format", logFormat).Msg("invalid log formatter. defaulting to text.")
		viper.Set(consts.LogFormatKey, "text")
		InitLog(cmd)
	}
}
