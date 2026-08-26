package tui

import (
	"io"
	"os"
	"strings"

	"charm.land/lipgloss/v2"
	"gabe565.com/utils/slogx"
	"gabe565.com/utils/termx"
	"github.com/charmbracelet/colorprofile"
)

// HasDarkBackground reports whether the terminal has a dark background.
//
// The value is set during init because remotecommand.Executor breaks
// detection of the stdin stream.
//
//nolint:gochecknoglobals
var HasDarkBackground = lipgloss.HasDarkBackground(os.Stdin, os.Stdout)

//nolint:gochecknoglobals
var lightDark = lipgloss.LightDark(HasDarkBackground)

func InitColorProfile(format slogx.Format) {
	var color bool
	switch format {
	case slogx.FormatAuto:
		color = termx.IsColor(os.Stdout)
	case slogx.FormatColor:
		color = true
	}
	if color {
		lipgloss.Writer.Profile = colorprofile.ANSI256
	} else {
		lipgloss.Writer.Profile = colorprofile.Ascii
	}
}

// NewWriter wraps w so that colors are downsampled to the configured profile.
func NewWriter(w io.Writer) *colorprofile.Writer {
	return &colorprofile.Writer{Forward: w, Profile: lipgloss.Writer.Profile}
}

// Plain strips all styling from a rendered string.
func Plain(s string) string {
	var buf strings.Builder
	w := colorprofile.Writer{Forward: &buf, Profile: colorprofile.NoTTY}
	_, _ = w.WriteString(s)
	return buf.String()
}
