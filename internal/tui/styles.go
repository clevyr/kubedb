package tui

import (
	"regexp"

	"github.com/charmbracelet/lipgloss"
	"github.com/clevyr/kubedb/internal/consts"
	"github.com/spf13/viper"
)

func HeaderStyle(r *lipgloss.Renderer) lipgloss.Style {
	return lipgloss.NewStyle().Renderer(r).
		Bold(true).
		Foreground(lipgloss.AdaptiveColor{Light: "#5A56E0", Dark: "#7571F9"})
}

func NamespaceStyle(r *lipgloss.Renderer, namespace string) lipgloss.Style {
	style := lipgloss.NewStyle().Renderer(r).SetString(namespace)

	colors := viper.GetStringMapString(consts.NamespaceColorKey)
	for k, v := range colors {
		if regexp.MustCompile(k).MatchString(namespace) {
			style = style.Foreground(lipgloss.Color(v))
			break
		}
	}

	return style
}
