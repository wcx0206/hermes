package main

import (
	"log"

	"github.com/spf13/cobra"

	"github.com/wcx0206/hermes/internal/cli"
)

func main() {

	// cfg, err := config.Load()
	// if err != nil {
	// 	log.Fatalf("load config: %v", err)
	// }
	// // Initialize logging
	// err = logging.Init(cfg.Logging)
	// if err != nil {
	// 	log.Fatalf("init logger: %v", err)
	// }
	// defer logging.Sync()

	root := &cobra.Command{
		Use:   "hermes",
		Short: "Manage hermes backup projects",
	}

	root.AddCommand(
		cli.NewProjectCmd(),
		cli.NewServerCmd(),
		cli.NewConfigCmd(),
		cli.NewBackupCmd(),
	)

	if err := root.Execute(); err != nil {
		log.Fatal(err)
	}
}
