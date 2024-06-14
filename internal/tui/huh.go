package tui

import "github.com/charmbracelet/huh"

func NewForm(groups ...*huh.Group) *huh.Form {
	keymap := huh.NewDefaultKeyMap()
	keymap.Quit.SetKeys("ctrl+c", "esc")

	return huh.NewForm(groups...).WithKeyMap(keymap)
}
