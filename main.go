package main

import (
	"errors"
	"fmt"
	"io"
	"os"
	"runtime/debug"

	"github.com/clevyr/kubedb/cmd"
	"github.com/clevyr/kubedb/internal/util"
	"github.com/rs/zerolog/log"
)

var errPanic = errors.New("panic")

func main() {
	defer func() {
		var err error
		var status int
		if msg := recover(); msg != nil {
			status = 1
			log.Error().Interface("error", msg).Msg("recovered from panic")
			err = fmt.Errorf("%w: %v\n\n%s", errPanic, msg, string(debug.Stack()))
			_, _ = io.WriteString(os.Stderr, err.Error())
		}
		util.PostRun(err)
		os.Exit(status)
	}()

	rootCmd := cmd.NewCommand()
	if err := rootCmd.Execute(); err != nil {
		log.Error().Msg(err.Error())
		util.PostRun(err)
		//nolint:gocritic
		os.Exit(1)
	}
}
