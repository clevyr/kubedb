package flags

import (
	"github.com/spf13/cobra"
)

func Force(cmd *cobra.Command, p *bool) {
	cmd.Flags().BoolVarP(p, "force", "f", false, "Do not prompt before restore")
}
