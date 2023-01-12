package error_view

import "github.com/charmbracelet/bubbles/key"

type keyMap struct {
	quit key.Binding
	back key.Binding
}

func (k keyMap) ShortHelp() []key.Binding {
	return []key.Binding{k.quit, k.back}
}

func (k keyMap) FullHelp() [][]key.Binding {
	return nil
}

var keys = keyMap{
	quit: key.NewBinding(
		key.WithKeys("q", "esc", "ctrl+c"),
		key.WithHelp("q", "quit"),
	),
	back: key.NewBinding(
		key.WithKeys("backspace"),
		key.WithHelp("backspace", "back")),
}
