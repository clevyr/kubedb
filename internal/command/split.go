package command

import (
	"strings"

	"github.com/alessio/shellescape"
)

type Split string

func (s Split) String() string {
	var inSingleString bool
	var inDoubleString bool
	var wasEscaped bool
	split := strings.FieldsFunc(string(s), func(r rune) bool {
		if wasEscaped {
			defer func() {
				wasEscaped = false
			}()
		}
		switch r {
		case ' ':
			return !wasEscaped && !inSingleString && !inDoubleString
		case '\'':
			if !wasEscaped && !inDoubleString {
				inSingleString = !inSingleString
				return true
			}
		case '"':
			if !wasEscaped && !inSingleString {
				inDoubleString = !inDoubleString
				return true
			}
		case '\\':
			wasEscaped = true
		}
		return false
	})
	for i, val := range split {
		val = strings.ReplaceAll(val, `\`, "")
		split[i] = shellescape.Quote(val)
	}
	return strings.Join(split, " ")
}
