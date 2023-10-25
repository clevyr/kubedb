package flags

import (
	"github.com/clevyr/kubedb/internal/consts"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

const DefaultSpinner = "sand"

func Spinner(cmd *cobra.Command, p *string) {
	cmd.Flags().StringVar(p, consts.SpinnerFlag, DefaultSpinner, "Spinner from https://jsfiddle.net/sindresorhus/2eLtsbey/embedded/result/")
	if err := cmd.Flags().MarkHidden(consts.SpinnerFlag); err != nil {
		panic(err)
	}
}

func BindSpinner(cmd *cobra.Command) {
	if err := viper.BindPFlag(consts.SpinnerKey, cmd.Flags().Lookup(consts.SpinnerFlag)); err != nil {
		panic(err)
	}
}
