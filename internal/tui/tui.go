package tui

import (
	"os"

	"gabe565.com/utils/slogx"
	"gabe565.com/utils/termx"
	"github.com/charmbracelet/lipgloss"
	"github.com/muesli/termenv"
)

//nolint:gochecknoglobals
var Renderer *lipgloss.Renderer

func InitRenderer(format slogx.Format) {
	var color bool
	switch format {
	case slogx.FormatAuto:
		color = termx.IsColor(os.Stdout)
	case slogx.FormatColor:
		color = true
	}
	Renderer = lipgloss.NewRenderer(os.Stdout, termenv.WithTTY(color))
	if color {
		Renderer.SetColorProfile(termenv.ANSI256)
	} else {
		Renderer.SetColorProfile(termenv.Ascii)
	}
	Renderer.SetHasDarkBackground(lipgloss.HasDarkBackground())
}

func init() { //nolint:gochecknoinits
	// Make lipgloss cache the current background color.
	// Attempting to use stdin after remotecommand.Executor results in an EOF,
	// so this value must be cached early to allow KubeDB to print tables post-exec.
	_ = lipgloss.HasDarkBackground()
}
