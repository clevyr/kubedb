package github

import (
	"os"
)

const OutputEnv = "GITHUB_OUTPUT"

func SetOutput(name, value string) error {
	if filename := os.Getenv(OutputEnv); filename != "" {
		f, err := os.OpenFile(filename, os.O_APPEND|os.O_WRONLY, 0o755)
		if err != nil {
			return err
		}
		if _, err := f.WriteString(name + "=" + value + "\n"); err != nil {
			return err
		}
		return f.Close()
	}
	return nil
}
