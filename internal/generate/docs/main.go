package main

import (
	"fmt"
	"log"
	"os"

	"github.com/clevyr/kubedb/cmd"
	"github.com/clevyr/kubedb/internal/config/flags"
	"github.com/spf13/cobra/doc"
)

func main() {
	var err error
	output := "./docs"

	err = os.RemoveAll(output)
	if err != nil {
		log.Fatal(fmt.Errorf("failed to remove existing dia: %w", err))
	}

	err = os.MkdirAll(output, 0o755)
	if err != nil {
		log.Fatal(fmt.Errorf("failed to mkdir: %w", err))
	}

	if err := os.Setenv(flags.KubeconfigEnv, ""); err != nil {
		log.Fatal(err)
	}

	rootCmd := cmd.NewCommand()

	err = doc.GenMarkdownTree(rootCmd, output)
	if err != nil {
		log.Fatal(fmt.Errorf("failed to generate markdown: %w", err))
	}
}
