package util

import "errors"

var (
	ErrNoDump        = errors.New("database does not support dump")
	ErrNoExec        = errors.New("database does not support exec")
	ErrNoPortForward = errors.New("database does not support port forwarding")
	ErrNoRestore     = errors.New("database does not support restore")
)
