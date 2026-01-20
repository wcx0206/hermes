package cli

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"syscall"
	"time"

	"github.com/spf13/cobra"
	"github.com/wcx0206/hermes/internal/backup"
)

type serverOpts struct {
	binaryPath string // 备份服务的二进制文件路径
	configPath string // 配置文件路径
}

func NewServerCmd() *cobra.Command {
	opts := &serverOpts{}
	cmd := &cobra.Command{
		Use:   "server",
		Short: "Manage the backup server",
	}
	cliPath, err := os.Executable()
	if err != nil {
		fmt.Fprintln(os.Stderr, "failed to get executable path:", err)
	}

	backupBinaryPath := filepath.Join(filepath.Dir(cliPath), "hermes-backup") // 假设备份服务与CLI在同一目录下
	configPath := filepath.Join(filepath.Dir(cliPath), "config.yaml")

	cmd.PersistentFlags().StringVar(&opts.configPath, "config", configPath, "config file path")
	cmd.PersistentFlags().StringVar(&opts.binaryPath, "binary", backupBinaryPath, "backup server binary path")

	cmd.AddCommand(newServerStartCmd(opts))
	cmd.AddCommand(newServerStopCmd(opts))
	cmd.AddCommand(newServerRestartCmd(opts))
	return cmd
}

// 启动 Backup Server
func newServerStartCmd(opts *serverOpts) *cobra.Command {
	return &cobra.Command{
		Use:   "start",
		Short: "Start the backup server",
		RunE: func(c *cobra.Command, args []string) error {
			// 校验备份二进制文件是否存在
			if _, err := os.Stat(opts.binaryPath); err != nil {
				return fmt.Errorf("backup binary not found: %w", err)
			}
			// 校验当前配置文件是否存在
			if _, err := os.Stat(opts.configPath); err != nil {
				return fmt.Errorf("config file not found: %w", err)
			}
			// 校验当前备份进程是否在运行 基于存储在 /var/run/hermes-backup.pid 中的pid
			pid, err := backup.GetPid()
			switch {
			case err == nil:
				if pid != 0 {
					if proc, findErr := os.FindProcess(pid); findErr == nil {
						if proc.Signal(syscall.Signal(0)) == nil {
							return fmt.Errorf("backup server already running (pid=%d)", pid)
						}
					}
				}
			case os.IsNotExist(err):
				// PID 文件不存在 说明没有运行中的进程
			default:
				return fmt.Errorf("failed to check backup server status: %w", err)
			}

			// 启动备份服务器进程
			proc := exec.Command(opts.binaryPath, "--config", opts.configPath)
			proc.Stdout = c.OutOrStdout()
			proc.Stderr = c.ErrOrStderr()

			if err := proc.Start(); err != nil {
				return fmt.Errorf("failed start server: %w", err)
			}
			fmt.Fprintln(c.OutOrStdout(), "Hermes Backup server started")
			return nil
		},
	}
}

// 停止 Backup Server
func newServerStopCmd(_ *serverOpts) *cobra.Command {
	return &cobra.Command{
		Use:   "stop",
		Short: "Gracefully stop backup server",
		RunE: func(_ *cobra.Command, _ []string) error {
			pid, err := backup.GetPid()
			fmt.Printf("Stopping Hermes Backup server (pid=%d)...\n", pid)
			if err != nil {
				return err
			}
			proc, err := os.FindProcess(pid)
			if err != nil {
				return err
			}
			signalErr := proc.Signal(syscall.SIGTERM)
			if signalErr != nil && !errors.Is(signalErr, os.ErrProcessDone) {
				return fmt.Errorf("send stop signal: %w", signalErr)
			}
			deadline := time.Now().Add(10 * time.Second)
			for {
				// 判断进程退出过程是否超时
				if time.Now().After(deadline) {
					fmt.Printf("backup server (pid=%d) did not exit in time\n", pid)
					//退出超时再次发送 SIGKILL 信号强制终止
					_ = proc.Signal(syscall.SIGKILL)
					break
				}
				if _, chkErr := backup.GetPid(); os.IsNotExist(chkErr) {
					break
				}
				time.Sleep(500 * time.Millisecond)
			}
			fmt.Println("Hermes Backup server stopped")
			return nil
		},
	}
}

func newServerRestartCmd(opts *serverOpts) *cobra.Command {
	return &cobra.Command{
		Use:   "restart",
		Short: "Restart the backup server",
		RunE: func(c *cobra.Command, _ []string) error {
			pid, err := backup.GetPid()
			switch {
			case err == nil && pid > 0:
				proc, findErr := os.FindProcess(pid)
				if findErr == nil {
					if sigErr := proc.Signal(syscall.SIGTERM); sigErr != nil && !errors.Is(sigErr, os.ErrProcessDone) {
						return fmt.Errorf("send stop signal: %w", sigErr)
					}
				}
				deadline := time.Now().Add(10 * time.Second)
				for {
					if time.Now().After(deadline) {
						return fmt.Errorf("backup server (pid=%d) did not exit in time", pid)
					}
					if _, chkErr := backup.GetPid(); os.IsNotExist(chkErr) {
						break
					}
					time.Sleep(500 * time.Millisecond)
				}
			case os.IsNotExist(err):
				// not running, proceed to start
			case err != nil:
				return fmt.Errorf("check pid: %w", err)
			}

			if _, err := os.Stat(opts.binaryPath); err != nil {
				return fmt.Errorf("backup binary not found: %w", err)
			}
			if _, err := os.Stat(opts.configPath); err != nil {
				return fmt.Errorf("config file not found: %w", err)
			}

			proc := exec.Command(opts.binaryPath, "--config", opts.configPath)
			proc.Stdout = c.OutOrStdout()
			proc.Stderr = c.ErrOrStderr()

			if err := proc.Start(); err != nil {
				return fmt.Errorf("failed start server: %w", err)
			}
			fmt.Fprintln(c.OutOrStdout(), "Hermes Backup server restarted")
			return nil
		},
	}
}
