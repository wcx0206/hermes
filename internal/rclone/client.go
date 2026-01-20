package rclone

import (
	"fmt"
	"os/exec"
)

type Client struct {
	RemoteName string
}

type Options struct {
	RemoteName string
}

func NewRcloneClient(opts Options) *Client {
	return &Client{
		RemoteName: opts.RemoteName,
	}
}

func (c *Client) Sync(localPath, remotePath string) error {
	cmd := exec.Command(
		"rclone",
		"sync",
		localPath,
		fmt.Sprintf("%s:%s", c.RemoteName, remotePath),
		"--transfers=4",
		"--checkers=4",
	)
	return cmd.Run()
}

func (c *Client) Copy(localPaths []string, remotePath string) error {
	for _, lp := range localPaths {
		cmd := exec.Command(
			"rclone",
			"copy",
			lp,
			fmt.Sprintf("%s:%s", c.RemoteName, remotePath),
			"--transfers=4",
			"--checkers=4",
		)
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("rclone copy %s to %s:%s failed: %w", lp, c.RemoteName, remotePath, err)
		}
	}
	return nil
}
