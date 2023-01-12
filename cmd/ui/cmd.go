package ui

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	uiConfig "github.com/clevyr/kubedb/cmd/ui/config"
	"github.com/clevyr/kubedb/cmd/ui/models/root"
	"github.com/clevyr/kubedb/internal/config"
	zone "github.com/lrstanley/bubblezone"
	"github.com/muesli/termenv"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var Command = &cobra.Command{
	Use:     "ui",
	Short:   "launch a terminal UI",
	PreRunE: preRun,
	RunE:    run,
}

var conf config.Global

func preRun(cmd *cobra.Command, args []string) error {
	if err := viper.Unmarshal(&conf); err != nil {
		return err
	}

	return nil
}

func run(cmd *cobra.Command, args []string) error {
	lipgloss.SetHasDarkBackground(termenv.HasDarkBackground())
	zone.NewGlobal()

	conf := &uiConfig.Config{
		Global:  conf,
		RootCmd: cmd.Parent(),
	}

	program := tea.NewProgram(
		root.NewModel(conf),
		tea.WithAltScreen(),
		tea.WithMouseCellMotion(),
	)
	if _, err := program.Run(); err != nil {
		return err
	}

	if conf.Run && conf.Cmd != nil {
		return conf.RootCmd.Execute()
	}
	return nil
}
