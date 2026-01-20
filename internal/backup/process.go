package backup

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
)

func getBackupPidFilePath() string {
	execPath, err := os.Executable()
	if err != nil {
		return "~/.cache/hermes/hermes-backup.pid"
	}
	return filepath.Join(filepath.Dir(execPath), "hermes-backup.pid")
}

func GetPid() (int, error) {
	data, err := os.ReadFile(getBackupPidFilePath())
	pidStr := string(data)
	if err != nil || pidStr == "" {
		return 0, err
	}
	if pid, convErr := strconv.Atoi(pidStr); convErr == nil {
		return pid, nil
	}
	return 0, fmt.Errorf("invsalid pid file content: %s", pidStr)
}

func SavePid() error {
	pid := os.Getpid()

	return os.WriteFile(getBackupPidFilePath(), []byte(fmt.Sprint(pid)), 0o644)
}

func RemovePid() error {
	return os.Remove(getBackupPidFilePath())
}
