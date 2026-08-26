package telemetry

import (
	"bufio"
	"fmt"
	"strconv"
	"strings"
)

// SysInfo holds host system specification and runtime metadata.
type SysInfo struct {
	OSDistro string `json:"os_distro"`
	Kernel   string `json:"kernel"`
	Uptime   string `json:"uptime"`
	CPUCores int    `json:"cpu_cores"`
}

// ParseSysInfo parses the system information section output.
func ParseSysInfo(content string) *SysInfo {
	info := &SysInfo{
		OSDistro: "Linux",
		Kernel:   "POSIX",
		Uptime:   "Unknown",
		CPUCores: 1,
	}

	scanner := bufio.NewScanner(strings.NewReader(content))
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}

		if strings.HasPrefix(line, "PRETTY_NAME=") {
			distro := strings.TrimPrefix(line, "PRETTY_NAME=")
			distro = strings.Trim(distro, `"'`)
			info.OSDistro = distro
		} else if strings.HasPrefix(line, "Linux ") || strings.HasPrefix(line, "Darwin ") {
			info.Kernel = line
		} else if strings.HasPrefix(line, "up ") {
			info.Uptime = line
		} else if cores, err := strconv.Atoi(line); err == nil && cores > 0 {
			info.CPUCores = cores
		} else if strings.Contains(line, ".") && !strings.Contains(line, " ") {
			// /proc/uptime in seconds
			parts := strings.Fields(line)
			if len(parts) > 0 {
				if secFloat, err := strconv.ParseFloat(parts[0], 64); err == nil {
					totalSecs := int64(secFloat)
					days := totalSecs / 86400
					hours := (totalSecs % 86400) / 3600
					mins := (totalSecs % 3600) / 60
					if days > 0 {
						info.Uptime = fmt.Sprintf("up %d days, %d hrs", days, hours)
					} else {
						info.Uptime = fmt.Sprintf("up %d hrs, %d mins", hours, mins)
					}
				}
			}
		}
	}

	return info
}
