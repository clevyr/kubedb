package flags

import (
	"gabe565.com/utils/must"
	"github.com/clevyr/kubedb/internal/completion"
	"github.com/clevyr/kubedb/internal/consts"
	"github.com/spf13/cobra"
)

func RemoteGzip(cmd *cobra.Command) {
	cmd.Flags().Bool(consts.FlagRemoteGzip, true, "Compress data over the wire. Results in lower bandwidth usage, but higher database load. May improve speed on slow connections.")
	must.Must(cmd.RegisterFlagCompletionFunc(consts.FlagRemoteGzip, completion.BoolCompletion))
}
