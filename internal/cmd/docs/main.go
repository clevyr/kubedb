package main

import (
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
	log.Println(`generating docs in "` + output + `"`)

	log.Println("removing existing directory")
	err = os.RemoveAll(output)
	if err != nil {
		log.Fatal(err)
	}

	log.Println("making directory")
	err = os.MkdirAll(output, 0755)
	if err != nil {
		log.Fatal(err)
	}

	log.Println("generating markdown")
	rootCmd := cmd.Command
	err = doc.GenMarkdownTree(rootCmd, output)
	if err != nil {
		log.Fatal(err)
	}

	log.Println("finished")
}
