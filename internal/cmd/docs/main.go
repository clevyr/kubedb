package main

import (
	"fmt"
	"github.com/clevyr/kubedb/cmd"
	"github.com/spf13/cobra/doc"
	flag "github.com/spf13/pflag"
	"log"
	"os"
	"path/filepath"
)

func main() {
	var output string
	flag.StringVarP(&output, "directory", "C", "./docs", "dir to hold the generated config")
	flag.Parse()

	var err error

	output = filepath.Join(".", filepath.Join("/", output))

	err = os.RemoveAll(output)
	if err != nil {
		log.Fatal(fmt.Errorf("failed to remove existing dia: %w", err))
	}

	err = os.MkdirAll(output, 0755)
	if err != nil {
		log.Fatal(fmt.Errorf("failed to mkdir: %w", err))
	}

	rootCmd := cmd.Command
	err = doc.GenMarkdownTree(rootCmd, output)
	if err != nil {
		log.Fatal(fmt.Errorf("failed to generate markdown: %w", err))
	}
}
