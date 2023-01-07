package main

import (
	"fmt"
	"github.com/clevyr/kubedb/cmd"
	"github.com/spf13/cobra/doc"
	"log"
	"os"
)

func main() {
	var err error
	output := "./docs"

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
