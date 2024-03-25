package flags

import (
	"github.com/clevyr/kubedb/internal/consts"
	"github.com/clevyr/kubedb/internal/util"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func RemoteGzip(cmd *cobra.Command) {
	cmd.Flags().Bool(consts.RemoteGzipFlag, true, "Compress data over the wire. Results in lower bandwidth usage, but higher database load. May improve speed on slow connections.")
	if err := cmd.RegisterFlagCompletionFunc(consts.RemoteGzipFlag, util.BoolCompletion); err != nil {
		panic(err)
	}
}

func BindRemoteGzip(cmd *cobra.Command) {
	if err := viper.BindPFlag(consts.RemoteGzipKey, cmd.Flags().Lookup(consts.RemoteGzipFlag)); err != nil {
		panic(err)
	}
}
