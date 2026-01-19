package rclone

import (
	"fmt"
	"os/exec"
)

type Client struct {
	RemoteName string
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

func (c *Client) Copy(localPath, remotePath string) error {
	cmd := exec.Command(
		"rclone",
		"copy",
		localPath,
		fmt.Sprintf("%s:%s", c.RemoteName, remotePath),
		"--transfers=4",
		"--checkers=4",
	)
	return cmd.Run()
}
