package cli

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"sort"
	"strconv"
	"strings"
	"syscall"

	"github.com/spf13/cobra"

	"github.com/wcx0206/hermes/internal/config"
)

type projectOpts struct {
	configPath string
}

func NewProjectCmd() *cobra.Command {
	opts := &projectOpts{}
	cmd := &cobra.Command{
		Use:   "project",
		Short: "Manage backup projects defined in config.yaml",
	}
	cmd.PersistentFlags().StringVar(&opts.configPath, "config", "config.yaml", "config file path")

	cmd.AddCommand(newProjectListCmd(opts))
	cmd.AddCommand(newProjectAddCmd(opts))
	cmd.AddCommand(newProjectDeleteCmd(opts))
	cmd.AddCommand(newProjectUpdateCmd(opts))
	return cmd
}

func newProjectListCmd(opts *projectOpts) *cobra.Command {
	return &cobra.Command{
		Use:     "list",
		Aliases: []string{"ls"},
		Short:   "List configured projects",
		RunE: func(cmd *cobra.Command, _ []string) error {
			cfg, err := config.LoadConfig(opts.configPath)
			if err != nil {
				fmt.Fprint(cmd.OutOrStderr(), "failed to load config:", err)
				return err
			}
			if len(cfg.Projects) == 0 {
				fmt.Fprintln(cmd.OutOrStdout(), "no projects configured")
				return nil
			}
			sort.Slice(cfg.Projects, func(i, j int) bool {
				return cfg.Projects[i].Name < cfg.Projects[j].Name
			})
			for _, p := range cfg.Projects {
				fmt.Fprintf(cmd.OutOrStdout(), "- %s (cron=%s, sources=%v)\n", p.Name, p.Cron, p.SourcePaths)
				for _, r := range p.RcloneRemotes {
					fmt.Fprintf(cmd.OutOrStdout(), "    - rclone remote: %s -> %s\n", r.Name, r.Bucket)
				}
			}
			return nil
		},
	}
}

func newProjectAddCmd(opts *projectOpts) *cobra.Command {
	var (
		name          string
		sourcePaths   []string
		rcloneRemotes []config.RcloneRemote
		cronExpr      string
	)
	cmd := &cobra.Command{
		Use:   "add",
		Short: "Add or replace a project",
		RunE: func(_ *cobra.Command, _ []string) error {
			reader := bufio.NewReader(os.Stdin)
			if name == "" {
				name = promptString(reader, "Project name")
				if name == "" {
					return errors.New("project name is required\n")
				}
			}
			if len(sourcePaths) == 0 {
				sourcePaths = promptList(reader, "Source paths (comma separated)")
				if len(sourcePaths) == 0 {
					return errors.New("at least one source path is required")
				}
			}
			if len(rcloneRemotes) == 0 {
				rcloneRemotes = promptRemotes(reader)
			}
			if cronExpr == "" {
				cronExpr = promptString(reader, "Cron expression (optional)")
			}

			if len(sourcePaths) == 0 {
				return errors.New("at least one source path is required")
			}
			cfg, err := config.LoadConfig(opts.configPath)
			if err != nil {
				return err
			}

			replace := config.Project{
				Name:          name,
				SourcePaths:   sourcePaths,
				Cron:          cronExpr,
				RcloneRemotes: rcloneRemotes,
			}

			replaced := false
			for i, p := range cfg.Projects {
				if p.Name == name {
					cfg.Projects[i] = replace
					replaced = true
					break
				}
			}
			if !replaced {
				cfg.Projects = append(cfg.Projects, replace)
			} else {
				fmt.Printf("project %s replaced\n", name)
			}
			if err := config.SaveConfig(opts.configPath, cfg); err != nil {
				fmt.Println("failed to save config:", err)
				return err
			}
			fmt.Printf("project %s saved:\n", name)
			printProject(replace)
			return nil
		},
	}
	cmd.Flags().StringVar(&name, "name", "", "project name")
	cmd.Flags().StringSliceVar(&sourcePaths, "source", nil, "source path (repeatable)")
	cmd.Flags().StringVar(&cronExpr, "cron", "", "cron expression (optional)")
	return cmd
}

func newProjectDeleteCmd(opts *projectOpts) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "delete <project-name>",
		Aliases: []string{"d"},
		Short:   "Delete a project by name",
		Args:    cobra.ExactArgs(1),
		RunE: func(_ *cobra.Command, args []string) error {
			name := args[0]
			if name == "" {
				fmt.Println("project name is required")
				return nil
			}
			cfg, err := config.LoadConfig(opts.configPath)
			if err != nil {
				return err
			}
			filtered := cfg.Projects[:0]
			for _, p := range cfg.Projects {
				if p.Name != name {
					filtered = append(filtered, p)
				}
			}
			if len(filtered) == len(cfg.Projects) {
				return fmt.Errorf("project %s not found", name)
			}
			cfg.Projects = filtered
			if err := config.SaveConfig(opts.configPath, cfg); err != nil {
				fmt.Println("failed to save config:", err)
				return err
			}
			fmt.Printf("project %s deleted\n", name)
			return nil
		},
	}
	return cmd
}

func newProjectUpdateCmd(opts *projectOpts) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "update <project-name>",
		Aliases: []string{"u"},
		Short:   "Interactively update a project",
		Args:    cobra.ExactArgs(1),
		RunE: func(_ *cobra.Command, args []string) error {
			name := args[0]
			if name == "" {
				fmt.Println("project name is required")
				return nil
			}
			cfg, err := config.LoadConfig(opts.configPath)
			if err != nil {
				return err
			}
			var project *config.Project
			for i := range cfg.Projects {
				if cfg.Projects[i].Name == name {
					project = &cfg.Projects[i]
					break
				}
			}
			if project == nil {
				return fmt.Errorf("project %s not found", name)
			}

			reader := bufio.NewReader(os.Stdin)

			if v := promptDefault(reader, "Project name", project.Name); v != "" {
				project.Name = v
			}
			if sources := promptDefault(reader, "Source paths (comma separated)", strings.Join(project.SourcePaths, ",")); sources != "" {
				project.SourcePaths = splitCSV(sources)
			}
			if cron := promptDefault(reader, "Cron expression", project.Cron); cron != "" {
				project.Cron = cron
			}
			project.RcloneRemotes = promptRemotesWithDefaults(reader, project.RcloneRemotes)

			if err := config.SaveConfig(opts.configPath, cfg); err != nil {
				fmt.Println("failed to save config:", err)
				return err
			}
			fmt.Printf("project %s updated\n", project.Name)
			return nil
		},
	}
	return cmd
}

func NewServiceCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "service",
		Short: "Interact with the running hermes service",
	}
	var pidFile string
	restart := &cobra.Command{
		Use:   "restart",
		Short: "Send SIGHUP to the service for safe restart",
		RunE: func(_ *cobra.Command, _ []string) error {
			pidBytes, err := os.ReadFile(pidFile)
			if err != nil {
				return err
			}
			pid, err := strconv.Atoi(strings.TrimSpace(string(pidBytes)))
			if err != nil {
				return err
			}
			process, err := os.FindProcess(pid)
			if err != nil {
				return err
			}
			return process.Signal(syscall.SIGHUP)
		},
	}
	restart.Flags().StringVar(&pidFile, "pid-file", "/var/run/hermes.pid", "service PID file")
	cmd.AddCommand(restart)
	return cmd
}

func printProject(p config.Project) {
	fmt.Printf("Name: %s\n", p.Name)
	fmt.Printf("Source Paths: %v\n", p.SourcePaths)
	fmt.Printf("Cron: %s\n", p.Cron)
	fmt.Printf("Rclone Remotes:\n")
	for _, r := range p.RcloneRemotes {
		fmt.Printf("  - Name: %s, Bucket: %s\n", r.Name, r.Bucket)
	}
}
