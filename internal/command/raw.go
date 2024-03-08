package command

type Raw string

func (r Raw) Quote() string {
	return string(r)
}

const Pipe = Raw("|")
