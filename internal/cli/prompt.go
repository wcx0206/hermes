package cli

import (
	"bufio"
	"fmt"
	"strings"

	"github.com/wcx0206/hermes/internal/config"
)

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

func promptDefault(r *bufio.Reader, label, current string) string {
	fmt.Printf("%s [%s]: ", label, current)
	text, _ := r.ReadString('\n')
	text = strings.TrimSpace(text)
	if text == "" {
		return current
	}
	return text
}

func splitCSV(val string) []string {
	if val == "" {
		return nil
	}
	parts := strings.Split(val, ",")
	for i := range parts {
		parts[i] = strings.TrimSpace(parts[i])
	}
	return parts
}

func promptRemotesWithDefaults(r *bufio.Reader, current []config.RcloneRemote) []config.RcloneRemote {
	var result []config.RcloneRemote
	for i := 0; ; i++ {
		var existing config.RcloneRemote
		if i < len(current) {
			existing = current[i]
		}
		name := promptDefault(r, fmt.Sprintf("Remote #%d name", i+1), existing.Name)
		if name == "" {
			break
		}
		bucket := promptDefault(r, fmt.Sprintf("Remote #%d bucket", i+1), existing.Bucket)
		result = append(result, config.RcloneRemote{
			Name:   name,
			Bucket: bucket,
		})
		more := promptString(r, "Add/keep another remote? (y/n)")
		if strings.ToLower(more) != "y" {
			break
		}
	}
	return result
}
