package command

import "gopkg.in/alessio/shellescape.v1"

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

func (e Env) String() string {
	return e.Key + "=" + shellescape.Quote(e.Value)
}
