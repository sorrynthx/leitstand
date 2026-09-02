package tui

import (
	"context"
	"fmt"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
)

func SimulateDemoCmd(cmd, host string) string {
	trimmed := strings.TrimSpace(cmd)
	lower := strings.ToLower(trimmed)

	switch {
	case lower == "ls" || lower == "ls -la" || lower == "ll":
		return "drwxr-xr-x 4 ubuntu ubuntu 4096 Jan  1 00:00 .\ndrwxr-xr-x 3 root   root   4096 Jan  1 00:00 ..\n-rw-r--r-- 1 ubuntu ubuntu 12540 Jan  1 00:00 app.py\n-rw-r--r-- 1 ubuntu ubuntu  1240 Jan  1 00:00 config.yaml\n-rw-r--r-- 1 ubuntu ubuntu   890 Jan  1 00:00 docker-compose.yml"

	case lower == "uptime":
		return fmt.Sprintf(" %s up 10 days, 2:15,  2 users,  load average: 0.45, 0.32, 0.28", time.Now().Format("15:04:05"))

	case lower == "free -m" || lower == "free -h":
		return "               total        used        free      shared  buff/cache   available\nMem:           16384        6420        9964         128        2400        9400\nSwap:           2048           0        2048"

	case lower == "df -h":
		return "Filesystem      Size  Used Avail Use% Mounted on\n/dev/sda1       512G  185G  328G  36% /\ntmpfs            16G  128M   16G   1% /dev/shm"

	case strings.HasPrefix(lower, "systemctl status"):
		service := strings.TrimSpace(strings.TrimPrefix(lower, "systemctl status"))
		return fmt.Sprintf("● %s.service - %s Service\n   Loaded: loaded (/etc/systemd/system/%s.service; enabled)\n   Active: active (running) since Mon 2026-08-31 00:00:00 KST\n Main PID: 1245 (%s)\n    Tasks: 8\n   Memory: 48.5M\n   CGroup: /system.slice/%s.service", service, service, service, service, service)

	default:
		return fmt.Sprintf("Simulated execution output for '%s' on %s.\nResult: Command completed successfully with status 0.", cmd, host)
	}
}

func SimulateDemoFileTransfer(ctx context.Context, hostID int64, isUpload bool, srcPaths []string, destDir string, msgChan chan<- tea.Msg) {
	defer close(msgChan)
	total := len(srcPaths)

	for i, src := range srcPaths {
		fileName := strings.TrimPrefix(src, "./")
		if idx := strings.LastIndex(fileName, "/"); idx != -1 {
			fileName = fileName[idx+1:]
		}

		for pct := 10; pct <= 100; pct += 30 {
			select {
			case <-ctx.Done():
				return
			case <-time.After(200 * time.Millisecond):
				msgChan <- FileTransferProgressMsg{
					HostID:       hostID,
					FileName:     fileName,
					FileIndex:    i + 1,
					FileTotal:    total,
					CurrentBytes: int64(pct * 1000),
					TotalBytes:   100000,
					BytesPerSec:  204800,
					IsDone:       (pct == 100 && i == total-1),
					msgChan:      nil,
				}
			}
		}
	}
}
