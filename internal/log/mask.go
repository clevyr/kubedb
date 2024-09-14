package log

import (
	"fmt"
	"log/slog"
	"strings"
)

//nolint:gochecknoglobals
var masks []string

func AddMask(mask string) {
	masks = append(masks, mask)
}

func MaskAttr(_ []string, attr slog.Attr) slog.Attr {
	var newVal string
	var changed bool
	switch val := attr.Value.Any().(type) {
	case string:
		newVal, changed = replace(val)
	case fmt.Stringer:
		newVal, changed = replace(val.String())
	}
	if changed {
		return slog.String(attr.Key, newVal)
	}
	return attr
}

func replace(str string) (string, bool) {
	var changed bool
	for _, password := range masks {
		if strings.Contains(str, password) {
			changed = true
			str = strings.ReplaceAll(str, password, "***")
		}
	}
	return str, changed
}
