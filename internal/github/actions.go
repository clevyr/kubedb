package github

import (
	"fmt"
	"os"
)

func SetOutput(name, value string) error {
	if filename := os.Getenv("GITHUB_OUTPUT"); filename != "" {
		f, err := os.OpenFile(filename, os.O_APPEND|os.O_WRONLY, 0755)
		if err != nil {
			return err
		}
		if _, err := f.WriteString(name + "=" + value + "\n"); err != nil {
			return err
		}
		return f.Close()
	} else {
		fmt.Println("::set-output name=" + name + "::" + value)
	}
	return nil
}
