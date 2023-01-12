package namespaces

import zone "github.com/lrstanley/bubblezone"

type Item string

func (i Item) Title() string       { return zone.Mark(string(i), string(i)) }
func (i Item) Description() string { return "" }
func (i Item) FilterValue() string { return i.Title() }
