package flags

import (
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

const DefaultSpinner = "sand"

func Spinner(cmd *cobra.Command, p *string) {
	cmd.Flags().StringVar(p, "spinner", DefaultSpinner, "Spinner from https://jsfiddle.net/sindresorhus/2eLtsbey/embedded/result/")
	if err := cmd.Flags().MarkHidden("spinner"); err != nil {
		panic(err)
	}
}

func BindSpinner(cmd *cobra.Command) {
	if err := viper.BindPFlag("spinner.name", cmd.Flags().Lookup("spinner")); err != nil {
		panic(err)
	}
}
