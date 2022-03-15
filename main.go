package main

import (
	"github.com/clevyr/kubedb/cmd"
	"os"
)

//go:generate go run internal/cmd/docs/main.go --directory=docs

func main() {
	if err := cmd.Execute(); err != nil {
		os.Exit(1)
	}
}
