package command

type Raw string

func (r Raw) String() string {
	return string(r)
}

const Pipe = Raw("|")
