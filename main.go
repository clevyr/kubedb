package main

import (
	"errors"
	"fmt"
	"io"
	"log/slog"
	"os"
	"runtime/debug"

	"github.com/clevyr/kubedb/cmd"
	"github.com/clevyr/kubedb/internal/finalizer"
	"github.com/clevyr/kubedb/internal/log"
)

var errPanic = errors.New("panic")

func main() {
	defer func() {
		var err error
		var status int
		if msg := recover(); msg != nil {
			status = 1
			slog.Error("Recovered from panic", "error", msg)
			err = fmt.Errorf("%w: %v\n\n%s", errPanic, msg, string(debug.Stack()))
			_, _ = io.WriteString(os.Stderr, err.Error())
		}
		finalizer.PostRun(err)
		os.Exit(status)
	}()

	log.Init(os.Stderr, slog.LevelInfo, log.FormatAuto)
	rootCmd := cmd.NewCommand()
	if err := rootCmd.Execute(); err != nil {
		if rootCmd.SilenceErrors {
			slog.Error(err.Error())
		}
		finalizer.PostRun(err)
		os.Exit(1) //nolint:gocritic
	}
}
