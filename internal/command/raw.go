package command

type Raw string

func (r Raw) Quote() string {
	return string(r)
}

const (
	Pipe = Raw("|")
	// EndOfOpts marks the end of options. All following values are treated as positional arguments.
	EndOfOpts = Raw("--")
)
