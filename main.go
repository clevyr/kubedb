package main

import (
	"os"

	"github.com/clevyr/kubedb/cmd"
)

var (
	version = "next"
	commit  = ""
)

func main() {
	rootCmd := cmd.NewCommand(version, commit)
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
