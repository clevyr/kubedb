package mask

import (
	"fmt"
	"log/slog"
	"strings"
	"sync"
)

//nolint:gochecknoglobals
var Default = &Masker{}

type Masker struct {
	masks []string
	mu    sync.RWMutex
}

func (m *Masker) Add(masks ...string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.masks = append(m.masks, masks...)
}

func Add(masks ...string) {
	Default.Add(masks...)
}

func (m *Masker) MaskAttr(_ []string, attr slog.Attr) slog.Attr {
	var newVal string
	var changed bool
	switch val := attr.Value.Any().(type) {
	case string:
		newVal, changed = m.replace(val)
	case fmt.Stringer:
		newVal, changed = m.replace(val.String())
	}
	if changed {
		return slog.String(attr.Key, newVal)
	}
	return attr
}

func MaskAttr(groups []string, attr slog.Attr) slog.Attr {
	return Default.MaskAttr(groups, attr)
}

func (m *Masker) replace(str string) (string, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	var changed bool
	for _, password := range m.masks {
		if strings.Contains(str, password) {
			changed = true
			str = strings.ReplaceAll(str, password, "***")
		}
	}
	return str, changed
}
