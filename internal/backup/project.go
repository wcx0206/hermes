package backup

import (
	"context"
	"fmt"

	"github.com/wcx0206/hermes/internal/config"
	"github.com/wcx0206/hermes/internal/rclone"
)

func RunProject(ctx context.Context, cfg *config.Config, project *config.Project) error {
	for _, remote := range project.RcloneRemotes {
		client := &rclone.Client{RemoteName: remote.Name}
		for _, src := range project.SourcePaths {
			if err := client.Copy(src, remote.Bucket); err != nil {
				return fmt.Errorf("%s -> %s: %w", src, remote.Bucket, err)
			}
		}
	}
	return nil
}
