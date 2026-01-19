package main

import (
	"log"

	"github.com/spf13/cobra"

	"github.com/wcx0206/hermes/internal/cli"
)

func main() {
	root := &cobra.Command{
		Use:   "hermes",
		Short: "Manage hermes backup projects",
	}

	root.AddCommand(
		cli.NewProjectCmd(),
		cli.NewServiceCmd(),
	)

	if err := root.Execute(); err != nil {
		log.Fatal(err)
	}
}
