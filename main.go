package main

import (
	"errors"
	"fmt"
	"io"
	"os"
	"runtime/debug"

	"github.com/clevyr/kubedb/cmd"
	"github.com/clevyr/kubedb/internal/util"
)

var errPanic = errors.New("panic")

func main() {
	defer func() {
		var err error
		if msg := recover(); msg != nil {
			err = fmt.Errorf("%w: %v\n\n%s", errPanic, msg, string(debug.Stack()))
			_, _ = io.WriteString(os.Stderr, err.Error())
		}
		util.PostRun(err)
	}()

	rootCmd := cmd.NewCommand()
	if err := rootCmd.Execute(); err != nil {
		util.PostRun(err)
		//nolint:gocritic
		os.Exit(1)
	}
}
