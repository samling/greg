package main

import (
	"github.com/samling/greg/cmd/greg/cli"
)

func main() {
	rootCmd := cli.NewRootCommand()
	rootCmd.Execute()
}
