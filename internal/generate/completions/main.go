package main

import (
	"bytes"
	"io"
	"os"
	"path/filepath"

	"github.com/clevyr/kubedb/cmd"
)

const (
	shellBash = "bash"
	shellZsh  = "zsh"
	shellFish = "fish"
)

func main() {
	rootCmd := cmd.NewCommand()
	var buf bytes.Buffer
	rootCmd.SetOut(&buf)

	if err := os.RemoveAll("completions"); err != nil {
		panic(err)
	}

	for _, shell := range []string{shellBash, shellZsh, shellFish} {
		rootCmd.SetArgs([]string{"completion", shell})
		if err := rootCmd.Execute(); err != nil {
			panic(err)
		}

		path := filepath.Join("completions", shell)
		if err := os.MkdirAll(path, 0o755); err != nil {
			panic(err)
		}

		switch shell {
		case shellBash:
			path = filepath.Join(path, rootCmd.Name())
		case shellZsh:
			path = filepath.Join(path, "_"+rootCmd.Name())
		case shellFish:
			path = filepath.Join(path, rootCmd.Name()+".fish")
		}

		f, err := os.Create(path)
		if err != nil {
			panic(err)
		}

		if _, err := io.Copy(f, &buf); err != nil {
			panic(err)
		}

		if err := f.Close(); err != nil {
			panic(err)
		}
	}
}
