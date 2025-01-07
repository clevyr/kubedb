package tui

import "github.com/charmbracelet/lipgloss"

func init() { //nolint:gochecknoinits
	// Make lipgloss cache the current background color.
	// Attempting to use stdin after remotecommand.Executor results in an EOF,
	// so this value must be cached early to allow KubeDB to print tables post-exec.
	_ = lipgloss.HasDarkBackground()
}
