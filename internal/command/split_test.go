package command

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSplit_String(t *testing.T) {
	tests := []struct {
		name string
		s    Split
		want string
	}{
		{"1 param", Split("echo"), "echo"},
		{"2 params", Split("echo hello"), "echo hello"},
		{"escaped params", Split("echo hello | cat;"), "echo hello '|' 'cat;'"},
		{"single quoted", Split("echo 'hello world'"), "echo 'hello world'"},
		{"double quoted", Split(`echo "hello world"`), `echo 'hello world'`},
		{"inline space", Split(`echo hello\ world`), `echo 'hello world'`},
		{"escaped quote", Split(`echo \'hello\ world\'`), `echo ''"'"'hello world'"'"''`},
		{"escape in string", Split(`echo 'hello \' world'`), `echo 'hello '"'"' world'`},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equalf(t, tt.want, tt.s.String(), "String()")
		})
	}
}
