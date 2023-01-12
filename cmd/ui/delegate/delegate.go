package delegate

import (
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/lipgloss"
)

func NewDelegate(keys *KeyMap) list.DefaultDelegate {
	d := list.NewDefaultDelegate()

	primaryColor := lipgloss.AdaptiveColor{Light: "#C46F00", Dark: "#FC8F02"}
	d.Styles.SelectedTitle.Foreground(primaryColor)
	d.Styles.SelectedTitle.BorderForeground(primaryColor)

	secondaryColor := lipgloss.AdaptiveColor{Light: "#D48B2C", Dark: "#CF7502"}
	d.Styles.SelectedDesc.Foreground(secondaryColor)
	d.Styles.SelectedDesc.BorderForeground(primaryColor)

	help := []key.Binding{keys.Back, keys.Choose}

	d.ShortHelpFunc = func() []key.Binding {
		return help
	}

	d.FullHelpFunc = func() [][]key.Binding {
		return [][]key.Binding{help}
	}

	return d
}

type KeyMap struct {
	Back   key.Binding
	Choose key.Binding
}

func (d KeyMap) ShortHelp() []key.Binding {
	return []key.Binding{
		d.Back,
		d.Choose,
	}
}

func (d KeyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{
		{
			d.Back,
			d.Choose,
		},
	}
}

func NewKeyMap() *KeyMap {
	return &KeyMap{
		Back: key.NewBinding(
			key.WithKeys("backspace"),
			key.WithHelp("backspace", "back")),
		Choose: key.NewBinding(
			key.WithKeys("enter"),
			key.WithHelp("enter", "choose"),
		),
	}
}
