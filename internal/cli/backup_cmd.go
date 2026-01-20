package cli

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"github.com/wcx0206/hermes/internal/backup"
	"github.com/wcx0206/hermes/internal/config"
)

type backupOpts struct {
	configPath string
}

func NewBackupCmd() *cobra.Command {
	opts := &backupOpts{}
	cmd := &cobra.Command{
		Use:   "backup",
		Short: "Trigger backups manually",
	}
	cmd.PersistentFlags().StringVar(&opts.configPath, "config", "config.yaml", "config file path")

	cmd.AddCommand(
		newBackupRunCmd(opts),
	)
	return cmd
}

func newBackupRunCmd(opts *backupOpts) *cobra.Command {
	var projects string
	cmd := &cobra.Command{
		Use:   "run",
		Short: "Run backup now",
		RunE: func(_ *cobra.Command, _ []string) error {
			cfg, err := config.LoadConfig(opts.configPath)
			if err != nil {
				return err
			}
			projectSet := make(map[string]struct{})
			if projects != "" {
				for _, name := range strings.Split(projects, ",") {
					if trimmed := strings.TrimSpace(name); trimmed != "" {
						projectSet[trimmed] = struct{}{}
					}
				}
			}
			projectList := make([]config.Project, 0, len(cfg.Projects))
			for _, p := range cfg.Projects {
				// 如果没有指定项目，则全部添加
				if len(projectSet) == 0 {
					projectList = append(projectList, p)
					continue
				}
				if _, ok := projectSet[p.Name]; ok {
					projectList = append(projectList, p)
				}
			}
			ctx := context.Background()
			for i := range projectList {
				now := time.Now()
				if err := backup.RunProject(ctx, cfg, &projectList[i]); err != nil {
					return err
				}
				cost := time.Since(now)
				fmt.Printf("Backup for project '%s' completed successfully, cost '%s'\n", projectList[i].Name, cost)
			}
			return nil
		},
	}
	cmd.Flags().StringVar(&projects, "projects", "", "Comma-separated list of projects to back up (default: all)")
	return cmd
}
