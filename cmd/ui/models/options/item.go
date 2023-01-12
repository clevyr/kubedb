package options

import (
	zone "github.com/lrstanley/bubblezone"
	"github.com/spf13/pflag"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

type Item struct {
	*pflag.Flag
}

func (i Item) Title() string {
	return zone.Mark(i.Name, cases.Title(language.English).String(i.Name))
}
func (i Item) Description() string { return i.Usage }
func (i Item) FilterValue() string { return i.Name }
