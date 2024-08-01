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
