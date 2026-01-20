package backup

import (
	"context"

	"github.com/wcx0206/hermes/internal/config"
	"github.com/wcx0206/hermes/internal/rclone"
)

func RunProject(ctx context.Context, cfg *config.Config, project *config.Project) error {
	for _, remote := range project.RcloneRemotes {
		client := &rclone.Client{RemoteName: remote.Name}
		if project.Mode == "sync" {
			if err := client.Sync(project.SourcePaths, remote.Bucket); err != nil {
				return err
			}
		} else {
			if err := client.Copy(project.SourcePaths, remote.Bucket); err != nil {
				return err
			}
		}
	}
	return nil
}
