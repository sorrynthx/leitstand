package tui

import (
	"context"
	"fmt"
	"leitstand/internal/ssh"
	"path/filepath"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
)

// Demo constants
const (
	DemoDefaultUser = "ubuntu"
	DemoDefaultHost = "demo-node-1"
	DemoDefaultHome = "/home/ubuntu"
)

// GetDemoCompletionFiles returns a list of mock files for tab autocomplete in demo mode.
func GetDemoCompletionFiles() []string {
	return []string{
		"app/", "app.py", "logs/", "config.yaml", "README.md",
		"docker-compose.yml", "Dockerfile", "package.json",
	}
}

// GetDemoFileContent returns mock content for remote file editing in demo mode.
func GetDemoFileContent(filePath string) string {
	return fmt.Sprintf(`# Demo configuration file: %s
version: '3.8'
services:
  web:
    image: nginx:alpine
    ports:
      - '80:80'
    restart: always
  app:
    image: node:20-alpine
    environment:
      NODE_ENV: production
`, filePath)
}

// GenerateDemoTopFrame generates a realistic live top frame with fluctuating CPU/Mem and process metrics.
func GenerateDemoTopFrame(hostName string, tickCount int) string {
	now := time.Now().Format("15:04:05")
	load1 := 0.25 + float64(tickCount%7)*0.08
	load5 := 0.35 + float64(tickCount%5)*0.04
	load15 := 0.40

	cpuUsr := 12.5 + float64((tickCount*7)%25)
	cpuSys := 4.2 + float64((tickCount*3)%10)
	cpuIdl := 100.0 - cpuUsr - cpuSys

	memTotal := 16384
	memUsed := 4210 + (tickCount*13)%300
	memFree := memTotal - memUsed

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("top - %s up 14 days,  2 users,  load average: %.2f, %.2f, %.2f\n", now, load1, load5, load15))
	sb.WriteString(fmt.Sprintf("Tasks: 142 total,   1 running, 141 sleeping,   0 stopped,   0 zombie\n"))
	sb.WriteString(fmt.Sprintf("%%Cpu(s): %4.1f us, %4.1f sy,  0.0 ni, %4.1f id,  0.2 wa,  0.0 hi,  0.1 si\n", cpuUsr, cpuSys, cpuIdl))
	sb.WriteString(fmt.Sprintf("MiB Mem : %7d total, %7d free, %7d used,    4128 buff/cache\n", memTotal, memFree, memUsed))
	sb.WriteString(fmt.Sprintf("MiB Swap:    2048 total,    2048 free,       0 used.   11782 avail Mem\n\n"))
	sb.WriteString(fmt.Sprintf("  PID USER      PR  NI    VIRT    RES    SHR S  %%CPU  %%MEM     TIME+ COMMAND\n"))

	p1Cpu := 15.2 + float64((tickCount*11)%30)
	p2Cpu := 8.4 + float64((tickCount*5)%15)
	p3Cpu := 4.1 + float64((tickCount*3)%8)
	p4Cpu := 2.0 + float64((tickCount)%5)

	sb.WriteString(fmt.Sprintf("%5d %-8s %3d %3d %7s %6s %6s %s %5.1f %5.1f   %2d:%02d.%02d %-20s\n", 1042, "root", 20, 0, "1.4g", "412m", "32m", "S", p1Cpu, 2.5, 4, (tickCount*2)%60, tickCount%60, "mysqld"))
	sb.WriteString(fmt.Sprintf("%5d %-8s %3d %3d %7s %6s %6s %s %5.1f %5.1f   %2d:%02d.%02d %-20s\n", 1891, "www-data", 20, 0, "850m", "198m", "18m", "S", p2Cpu, 1.2, 2, (tickCount*3)%60, tickCount%60, "nginx: worker"))
	sb.WriteString(fmt.Sprintf("%5d %-8s %3d %3d %7s %6s %6s %s %5.1f %5.1f   %2d:%02d.%02d %-20s\n", 2210, "ubuntu", 20, 0, "620m", "145m", "22m", "S", p3Cpu, 0.9, 1, tickCount%60, tickCount%60, "node /app/server.js"))
	sb.WriteString(fmt.Sprintf("%5d %-8s %3d %3d %7s %6s %6s %s %5.1f %5.1f   %2d:%02d.%02d %-20s\n", 782, "root", 20, 0, "1.8g", "84m", "12m", "S", p4Cpu, 0.5, 0, tickCount%60, tickCount%60, "dockerd"))
	sb.WriteString(fmt.Sprintf("%5d %-8s %3d %3d %7s %6s %6s %s %5.1f %5.1f   %2d:%02d.%02d %-20s\n", 1, "root", 20, 0, "168m", "14m", "9m", "S", 0.0, 0.1, 0, 12, 45, "systemd"))
	sb.WriteString(fmt.Sprintf("%5d %-8s %3d %3d %7s %6s %6s %s %5.1f %5.1f   %2d:%02d.%02d %-20s\n", 412, "root", 20, 0, "92m", "8m", "6m", "S", 0.0, 0.1, 0, 4, 10, "systemd-journald"))
	sb.WriteString(fmt.Sprintf("%5d %-8s %3d %3d %7s %6s %6s %s %5.1f %5.1f   %2d:%02d.%02d %-20s\n", 915, "root", 20, 0, "18m", "4m", "3m", "S", 0.0, 0.0, 0, 1, 15, "sshd: /usr/sbin/sshd"))
	sb.WriteString(fmt.Sprintf("%5d %-8s %3d %3d %7s %6s %6s %s %5.1f %5.1f   %2d:%02d.%02d %-20s\n", 3120, "ubuntu", 20, 0, "22m", "5m", "4m", "R", 0.3, 0.0, 0, 0, tickCount%60, "top"))
	return sb.String()
}

// SimulateDemoCmd simulates command outputs for Leitstand's demo mode.
func SimulateDemoCmd(cmd string, hostName string) string {
	trimmed := strings.TrimSpace(cmd)

	switch {
	case trimmed == "ls" || trimmed == "ls -l" || trimmed == "ls -la" || trimmed == "dir":
		return `total 36
drwxr-xr-x 4 ubuntu ubuntu 4096 Aug 28 14:00 .
drwxr-xr-x 3 root   root   4096 Aug 20 09:00 ..
-rw-r--r-- 1 ubuntu ubuntu  420 Aug 28 12:00 README.md
drwxr-xr-x 2 ubuntu ubuntu 4096 Aug 28 13:30 app
-rw-r--r-- 1 ubuntu ubuntu 1204 Aug 28 11:15 config.yaml
-rw-r--r-- 1 ubuntu ubuntu  890 Aug 28 10:20 docker-compose.yml
drwxr-xr-x 2 ubuntu ubuntu 4096 Aug 28 14:00 logs`

	case trimmed == "docker ps" || trimmed == "docker ps -a":
		return `CONTAINER ID   IMAGE          COMMAND                  CREATED        STATUS        PORTS                NAMES
a1b2c3d4e5f6   nginx:alpine   "/docker-entrypoint.…"   2 hours ago    Up 2 hours    0.0.0.0:80->80/tcp   web-proxy
f6e5d4c3b2a1   redis:7-alpine "docker-entrypoint.s…"   5 hours ago    Up 5 hours    6379/tcp             cache-redis
9876543210ab   node:20-alpine "docker-entrypoint.s…"   24 hours ago   Up 24 hours   3000/tcp             api-server`

	case trimmed == "free -m" || trimmed == "free -h":
		return `               total        used        free      shared  buff/cache   available
Mem:           16384        4210        8046         120        4128       11782
Swap:           2048           0        2048`

	case trimmed == "df -h" || trimmed == "df":
		return `Filesystem      Size  Used Avail Use% Mounted on
/dev/root       100G   38G   62G  38% /
tmpfs           7.8G     0  7.8G   0% /dev/shm
/dev/nvme0n1p1  512M  6.2M  506M   2% /boot/efi`

	case trimmed == "uname -a":
		return fmt.Sprintf("Linux %s 6.8.0-40-generic #40-Ubuntu SMP PREEMPT_DYNAMIC x86_64 GNU/Linux", hostName)

	case trimmed == "whoami":
		return "ubuntu"

	case trimmed == "uptime":
		return " 14:00:00 up 14 days,  2 users,  load average: 0.42, 0.38, 0.35"

	case strings.HasPrefix(trimmed, "systemctl status"):
		service := strings.TrimSpace(strings.TrimPrefix(trimmed, "systemctl status"))
		if service == "" {
			service = "nginx"
		}
		return fmt.Sprintf(`● %s.service - %s Service
     Loaded: loaded (/lib/systemd/system/%s.service; enabled; preset: enabled)
     Active: active (running) since Sun 2026-08-28 08:00:00 KST; 6h ago
   Main PID: 1042 (%s)
      Tasks: 4 (limit: 18942)
     Memory: 48.2M
        CPU: 1.250s
     CGroup: /system.slice/%s.service
             ├─1042 %s: master process
             └─1043 %s: worker process`, service, service, service, service, service, service, service)

	case strings.HasPrefix(trimmed, "cat "):
		file := strings.TrimSpace(strings.TrimPrefix(trimmed, "cat "))
		if file == "config.yaml" {
			return `server:
  port: 8080
  host: 0.0.0.0
database:
  driver: postgres
  pool_size: 20`
		}
		return fmt.Sprintf("# Content of %s\n# (Simulated demo file content)\nstatus: ok\nenvironment: demo", file)

	default:
		return fmt.Sprintf("[Demo Mode] Command executed: '%s'\nResult: 0 (Success)", cmd)
	}
}

// StartDemoStream creates a mock live stream ticker for demo mode.
func StartDemoStream(ctx context.Context, hostID int64, tabID string, hostName string, isScreenApp bool, msgChan chan tea.Msg) {
	tickerInterval := 1000 * time.Millisecond
	if !isScreenApp {
		tickerInterval = 750 * time.Millisecond
	}

	ticker := time.NewTicker(tickerInterval)
	defer ticker.Stop()

	counter := 1
	for {
		select {
		case <-ctx.Done():
			msgChan <- StreamFinishedMsg{HostID: hostID, TabID: tabID, Err: nil}
			return
		case <-ticker.C:
			var chunk string
			if isScreenApp {
				chunk = GenerateDemoTopFrame(hostName, counter)
			} else {
				now := time.Now().Format("15:04:05.000")
				chunk = fmt.Sprintf("[%s] INFO demo-service: Processed request #%d - 200 OK (latency: %dms)",
					now, counter, 12+counter%35)
			}
			counter++
			msgChan <- StreamChunkMsg{HostID: hostID, TabID: tabID, Chunk: chunk, msgChan: msgChan}
			if counter > 300 {
				msgChan <- StreamFinishedMsg{HostID: hostID, TabID: tabID, Err: nil}
				return
			}
		}
	}
}

// GetDemoRemoteFiles returns simulated remote filesystem entries.
func GetDemoRemoteFiles(remotePath string) []*ssh.FileItem {
	var items []*ssh.FileItem

	if remotePath != "/" {
		items = append(items, &ssh.FileItem{
			Name:  "..",
			Path:  "/home",
			IsDir: true,
		})
	}

	items = append(items,
		&ssh.FileItem{Name: "app", Path: remotePath + "/app", IsDir: true},
		&ssh.FileItem{Name: "logs", Path: remotePath + "/logs", IsDir: true},
		&ssh.FileItem{Name: "nginx", Path: remotePath + "/nginx", IsDir: true},
		&ssh.FileItem{Name: "README.md", Path: remotePath + "/README.md", Size: 420, IsDir: false},
		&ssh.FileItem{Name: "config.yaml", Path: remotePath + "/config.yaml", Size: 1204, IsDir: false},
		&ssh.FileItem{Name: "docker-compose.yml", Path: remotePath + "/docker-compose.yml", Size: 890, IsDir: false},
		&ssh.FileItem{Name: "server.js", Path: remotePath + "/server.js", Size: 18432, IsDir: false},
		&ssh.FileItem{Name: "bundle.tar.gz", Path: remotePath + "/bundle.tar.gz", Size: 84520100, IsDir: false},
		&ssh.FileItem{Name: "app.py", Path: remotePath + "/app.py", Size: 12450, IsDir: false},
	)

	return items
}

// SimulateDemoFileTransfer simulates animated multi-file upload/download progress in demo mode.
func SimulateDemoFileTransfer(ctx context.Context, hostID int64, isUpload bool, srcPaths []string, destDirPath string, msgChan chan tea.Msg) {
	defer close(msgChan)

	totalFiles := len(srcPaths)
	for i, src := range srcPaths {
		fileName := filepath.Base(src)
		fileSize := int64(15 * 1024 * 1024) // 15MB mock
		if strings.HasSuffix(fileName, ".tar.gz") {
			fileSize = 85 * 1024 * 1024 // 85MB
		} else if strings.HasSuffix(fileName, ".py") || strings.HasSuffix(fileName, ".yaml") {
			fileSize = 120 * 1024 // 120KB
		}

		steps := 8
		stepBytes := fileSize / int64(steps)
		var currentBytes int64

		for s := 1; s <= steps; s++ {
			select {
			case <-ctx.Done():
				msgChan <- FileTransferProgressMsg{HostID: hostID, IsDone: true, Err: ctx.Err()}
				return
			case <-time.After(150 * time.Millisecond):
				currentBytes += stepBytes
				if currentBytes > fileSize {
					currentBytes = fileSize
				}
				speed := float64(stepBytes) / 0.15 // Bytes/sec

				isDone := (s == steps && i == totalFiles-1)
				msgChan <- FileTransferProgressMsg{
					HostID:       hostID,
					FileName:     fileName,
					FileIndex:    i + 1,
					FileTotal:    totalFiles,
					CurrentBytes: currentBytes,
					TotalBytes:   fileSize,
					BytesPerSec:  speed,
					IsDone:       isDone,
					Err:          nil,
					msgChan:      msgChan,
				}
			}
		}
	}
}
