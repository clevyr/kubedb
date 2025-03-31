package command

import (
	"errors"
	"fmt"
	"strings"

	"al.essio.dev/pkg/shellescape"
)

func NewBuilder(p ...any) *Builder {
	if err := checkType(p...); err != nil {
		panic(err)
	}
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

var ErrInvalidType = errors.New("invalid type")

func checkType(p ...any) error {
	for i := range p {
		switch p[i].(type) {
		case string, Quoter, fmt.Stringer:
			return nil
		default:
			return fmt.Errorf("%w: %#v", ErrInvalidType, p)
		}
	}
	return nil
}

func (j *Builder) Push(p ...any) *Builder {
	if err := checkType(p...); err != nil {
		panic(err)
	}
	j.cmd = append(j.cmd, p...)
	return j
}

func (j *Builder) Unshift(p ...any) *Builder {
	if err := checkType(p...); err != nil {
		panic(err)
	}
	j.cmd = append(p, j.cmd...)
	return j
}

func (j *Builder) String() string {
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
			panic(fmt.Errorf("%w: %#v", ErrInvalidType, v))
		}
		if k < len(j.cmd)-1 {
			buf.WriteString(" ")
		}
	}
	return buf.String()
}
