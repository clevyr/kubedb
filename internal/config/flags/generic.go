package flags

import (
	"gabe565.com/utils/must"
	"github.com/clevyr/kubedb/internal/consts"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

const DefaultSpinner = "dots14"

func Spinner(cmd *cobra.Command, p *string) {
	cmd.Flags().StringVar(p, consts.SpinnerFlag, DefaultSpinner, "Spinner from https://jsfiddle.net/sindresorhus/2eLtsbey/embedded/result/")
	must.Must(cmd.Flags().MarkHidden(consts.SpinnerFlag))
}

func BindSpinner(cmd *cobra.Command) {
	must.Must(viper.BindPFlag(consts.SpinnerKey, cmd.Flags().Lookup(consts.SpinnerFlag)))
}
