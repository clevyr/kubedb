package terminal

import (
	"golang.org/x/sys/unix"
	"os"
)

func IsTTY() bool {
	_, err := unix.IoctlGetWinsize(int(os.Stdout.Fd()), unix.TIOCGWINSZ)
	return err == nil
}
