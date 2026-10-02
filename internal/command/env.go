package command

import "al.essio.dev/pkg/shellescape"

func NewEnv(k, v string) Env {
	return Env{
		Key:   k,
		Value: v,
	}
}

type Env struct {
	Key   string
	Value string
}

func (e Env) Quote() string {
	return e.Key + "=" + shellescape.Quote(e.Value)
}

// Var references a shell variable. It is expanded by the shell.
type Var string

func (v Var) Quote() string {
	return `"${` + string(v) + `}"`
}
