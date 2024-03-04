package main

import (
	"os"

	"github.com/clevyr/kubedb/cmd"
)

func main() {
	rootCmd := cmd.NewCommand()
	if err := rootCmd.Execute(); err != nil {
		cmd.PostRun(err)
		os.Exit(1)
	}
	cmd.PostRun(nil)
}
