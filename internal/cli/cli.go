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
	"gopkg.in/yaml.v3"

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
	return cmd
}

func newProjectListCmd(opts *projectOpts) *cobra.Command {
	return &cobra.Command{
		Use:     "list",
		Aliases: []string{"-l"},
		Short:   "List configured projects",
		RunE: func(cmd *cobra.Command, _ []string) error {
			cfg, err := loadConfig(opts.configPath)
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
			}
			return nil
		},
	}
}

// 添加一个项目我希望可以逐步输入
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
			cfg, err := loadConfig(opts.configPath)
			if err != nil {
				return err
			}

			replace := config.Project{
				Name:          name,
				SourcePaths:   sourcePaths,
				Cron:          cronExpr,
				RcloneRemotes: rcloneRemotes,
			}
			// overwrite if exists
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
			if err := saveConfig(opts.configPath, cfg); err != nil {
				return err
			}
			fmt.Printf("project %s saved\n", name)
			return nil
		},
	}
	cmd.Flags().StringVar(&name, "name", "", "project name")
	cmd.Flags().StringSliceVar(&sourcePaths, "source", nil, "source path (repeatable)")
	cmd.Flags().StringVar(&cronExpr, "cron", "", "cron expression (optional)")
	return cmd
}

func newProjectDeleteCmd(opts *projectOpts) *cobra.Command {
	return &cobra.Command{
		Use:   "delete <project>",
		Short: "Delete a project by name",
		Args:  cobra.ExactArgs(1),
		RunE: func(_ *cobra.Command, args []string) error {
			name := args[0]
			cfg, err := loadConfig(opts.configPath)
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
			if err := saveConfig(opts.configPath, cfg); err != nil {
				return err
			}
			fmt.Printf("project %s deleted\n", name)
			return nil
		},
	}
}

func parseRemoteSpecs(specs []string, fallbackBucket string) ([]config.RcloneRemote, error) {
	if len(specs) == 0 {
		if fallbackBucket == "" {
			return nil, errors.New("no remote provided and defaults.bucket empty")
		}
		return []config.RcloneRemote{
			{Name: "default", Bucket: fallbackBucket},
		}, nil
	}
	remotes := make([]config.RcloneRemote, 0, len(specs))
	for _, spec := range specs {
		parts := strings.SplitN(spec, ":", 2)
		if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
			return nil, fmt.Errorf("invalid remote spec %q (expected name:bucket)", spec)
		}
		remotes = append(remotes, config.RcloneRemote{
			Name:   parts[0],
			Bucket: parts[1],
		})
	}
	return remotes, nil
}

func loadConfig(path string) (*config.Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var cfg config.Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}
	return &cfg, nil
}

func saveConfig(path string, cfg *config.Config) error {
	out, err := yaml.Marshal(cfg)
	if err != nil {
		return err
	}
	return os.WriteFile(path, out, 0o644)
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

func promptString(r *bufio.Reader, label string) string {
	fmt.Printf("%s: ", label)
	text, _ := r.ReadString('\n')
	return strings.TrimSpace(text)
}

func promptList(r *bufio.Reader, label string) []string {
	val := promptString(r, label)
	if val == "" {
		return nil
	}
	parts := strings.Split(val, ",")
	for i := range parts {
		parts[i] = strings.TrimSpace(parts[i])
	}
	return parts
}

func promptRemotes(r *bufio.Reader) []config.RcloneRemote {
	var remotes []config.RcloneRemote
	for {
		name := promptString(r, "rclone config name, can not be empty")
		if name == "" {
			fmt.Println("remote name cannot be empty, please retry.")
			continue
		}
		var bucket string
		for {
			bucket = promptString(r, "Bucket for "+name)
			if bucket == "" {
				fmt.Println("bucket required, please retry.")
				continue
			}
			break
		}
		remotes = append(remotes, config.RcloneRemote{
			Name:   name,
			Bucket: bucket,
		})
		more := promptString(r, "Add more remotes? (y/n)")
		if strings.ToLower(more) == "y" {
			continue
		}
		break
	}
	return remotes
}
