package backup

import (
	"fmt"
	"os"
	"strconv"
)

func GetPid() (int, error) {
	data, err := os.ReadFile("/var/run/hermes-backup.pid")
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
	pid := os.Getegid()
	pidFile := "/var/run/hermes-backup.pid"
	return os.WriteFile(pidFile, []byte(fmt.Sprint(pid)), 0o644)
}
