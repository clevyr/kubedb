package command

import (
	"fmt"
	"gopkg.in/alessio/shellescape.v1"
	"strings"
)

func NewBuilder(p ...any) *Builder {
	return &Builder{
		cmd: p,
	}
}

type Builder struct {
	cmd []any
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
		case Env:
			buf.WriteString(v.String())
		case Raw:
			buf.WriteString(string(v))
		default:
			panic(fmt.Errorf("unknown value in command: %#v", v))
		}
		if k < len(j.cmd)-1 {
			buf.WriteString(" ")
		}
	}
	return buf.String()
}
