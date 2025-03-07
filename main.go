package main

import (
	"log/slog"
	"os"

	"gabe565.com/utils/slogx"
	"github.com/clevyr/kubedb/cmd"
	"github.com/clevyr/kubedb/internal/log"
)

func main() {
	log.Init(os.Stderr, slogx.LevelInfo, slogx.FormatAuto)
	if err := cmd.Execute(cmd.New()); err != nil {
		slog.Error(err.Error())
		os.Exit(1)
	}
}
