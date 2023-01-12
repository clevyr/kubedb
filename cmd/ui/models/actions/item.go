package actions

import (
	zone "github.com/lrstanley/bubblezone"
	"github.com/spf13/cobra"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

type Item struct {
	*cobra.Command
}

func (i Item) Title() string {
	return zone.Mark(i.Name(), cases.Title(language.English).String(i.Name()))
}
func (i Item) Description() string { return i.Short }
func (i Item) FilterValue() string { return i.Title() }
