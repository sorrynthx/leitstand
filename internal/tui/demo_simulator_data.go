package tui

import (
	appssh "leitstand/internal/ssh"
	appstorage "leitstand/internal/storage"
	apptelemetry "leitstand/internal/telemetry"
	"strings"
	"time"
)

func CreateDemoHosts() []*appstorage.Host {
	return []*appstorage.Host{
		{ID: 101, Name: "prod-web-01", Address: "192.168.1.10", Port: 22, Username: "ubuntu", GroupName: "Production Web"},
		{ID: 102, Name: "prod-db-master", Address: "192.168.1.20", Port: 22, Username: "postgres", GroupName: "Database Layer"},
		{ID: 103, Name: "staging-api", Address: "10.0.4.15", Port: 2222, Username: "devops", GroupName: "Staging Environment"},
	}
}

func CreateDemoMetrics(hostID int64) (*appstorage.MetricRecord, *apptelemetry.SysInfo) {
	rec := &appstorage.MetricRecord{
		HostID:      hostID,
		Timestamp:   time.Now(),
		CPUPercent:  18.5,
		MemoryTotal: 16384,
		MemoryUsed:  6420,
		DiskTotal:   512,
		DiskUsed:    184,
		NetRxBytes:  1024500,
		NetTxBytes:  2048900,
	}

	sysInfo := &apptelemetry.SysInfo{
		OSDistro: "Ubuntu 22.04 LTS (Jammy Jellyfish)",
		Kernel:   "5.15.0-88-generic x86_64",
		Uptime:   "10 days, 2 hours",
		CPUCores: 8,
	}

	return rec, sysInfo
}

func GetDemoRemoteFiles(remotePath string) []*appssh.FileItem {
	now := time.Now()

	return []*appssh.FileItem{
		{Name: "..", Path: filepathDir(remotePath), IsDir: true, ModTime: now},
		{Name: "app", Path: remotePath + "/app", IsDir: true, ModTime: now},
		{Name: "logs", Path: remotePath + "/logs", IsDir: true, ModTime: now},
		{Name: "nginx", Path: remotePath + "/nginx", IsDir: true, ModTime: now},
		{Name: "app.py", Path: remotePath + "/app.py", Size: 12540, IsDir: false, ModTime: now},
		{Name: "config.yaml", Path: remotePath + "/config.yaml", Size: 1240, IsDir: false, ModTime: now},
		{Name: "docker-compose.yml", Path: remotePath + "/docker-compose.yml", Size: 890, IsDir: false, ModTime: now},
		{Name: "server.js", Path: remotePath + "/server.js", Size: 18400, IsDir: false, ModTime: now},
		{Name: "bundle.tar.gz", Path: remotePath + "/bundle.tar.gz", Size: 84500000, IsDir: false, ModTime: now},
		{Name: "README.md", Path: remotePath + "/README.md", Size: 420, IsDir: false, ModTime: now},
	}
}

func filepathDir(p string) string {
	if idx := strings.LastIndex(p, "/"); idx != -1 && idx > 0 {
		return p[:idx]
	}
	return "."
}
