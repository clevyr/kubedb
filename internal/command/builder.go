package command

import (
	"fmt"
	"strings"

	"github.com/alessio/shellescape"
)

func NewBuilder(p ...any) *Builder {
	return &Builder{
		cmd: p,
	}
}

type Builder struct {
	cmd []any
}

type Quoter interface {
	Quote() string
}

func (j *Builder) Push(p ...any) *Builder {
	j.cmd = append(j.cmd, p...)
	return j
}

func (j *Builder) Unshift(p ...any) *Builder {
	j.cmd = append(p, j.cmd...)
	return j
}

func (j Builder) String() string {
	var buf strings.Builder
	for k, v := range j.cmd {
		switch v := v.(type) {
		case string:
			buf.WriteString(shellescape.Quote(v))
		case Quoter:
			buf.WriteString(v.Quote())
		case fmt.Stringer:
			buf.WriteString(shellescape.Quote(v.String()))
		default:
			panic(fmt.Errorf("unknown value in command: %#v", v))
		}
		if k < len(j.cmd)-1 {
			buf.WriteString(" ")
		}
	}
	return buf.String()
}
