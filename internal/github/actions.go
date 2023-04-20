package github

import (
	"fmt"
	"io"
	"os"
)

var output io.Writer = os.Stdout

func SetOutput(name, value string) error {
	if filename := os.Getenv("GITHUB_OUTPUT"); filename != "" {
		f, err := os.OpenFile(filename, os.O_APPEND|os.O_WRONLY, 0o755)
		if err != nil {
			return err
		}
		if _, err := f.WriteString(name + "=" + value + "\n"); err != nil {
			return err
		}
		return f.Close()
	} else {
		if _, err := fmt.Fprint(output, "::set-output name="+name+"::"+value); err != nil {
			return err
		}
	}
	return nil
}
