package telemetry

import (
	"fmt"
	"strings"
	"time"
)

const (
	MetricSplitDelimiter = "===LEITSTAND_SPLIT==="

	// MetricExtractScript runs POSIX commands in one single roundtrip.
	MetricExtractScript = `cat /proc/stat; echo '` + MetricSplitDelimiter + `'; cat /proc/meminfo; echo '` + MetricSplitDelimiter + `'; df -k /; echo '` + MetricSplitDelimiter + `'; cat /proc/net/dev; echo '` + MetricSplitDelimiter + `'; uname -srm; cat /etc/os-release 2>/dev/null | grep PRETTY_NAME; uptime -p 2>/dev/null || cat /proc/uptime 2>/dev/null; nproc 2>/dev/null || grep -c ^processor /proc/cpuinfo 2>/dev/null`
)

// RawTelemetryBundle holds parsed data components from a single extraction script run.
type RawTelemetryBundle struct {
	Timestamp time.Time
	CPUTick   *CPUTickSnapshot
	Memory    *MemoryStats
	Disk      *DiskStats
	Network   *NetworkStats
	SysInfo   *SysInfo
}

// ParseRawBundle parses the composite output from MetricExtractScript.
func ParseRawBundle(rawOutput string) (*RawTelemetryBundle, error) {
	sections := strings.Split(rawOutput, MetricSplitDelimiter)
	if len(sections) < 4 {
		return nil, fmt.Errorf("unexpected telemetry output format: expected at least 4 sections, got %d", len(sections))
	}

	cpuSnap, err := ParseProcStat(sections[0])
	if err != nil {
		return nil, fmt.Errorf("cpu parse failed: %w", err)
	}

	memStats, err := ParseMeminfo(sections[1])
	if err != nil {
		return nil, fmt.Errorf("meminfo parse failed: %w", err)
	}

	diskStats, err := ParseDF(sections[2])
	if err != nil {
		return nil, fmt.Errorf("df parse failed: %w", err)
	}

	netStats, err := ParseProcNetDev(sections[3])
	if err != nil {
		return nil, fmt.Errorf("network parse failed: %w", err)
	}

	var sysInfo *SysInfo
	if len(sections) >= 5 {
		sysInfo = ParseSysInfo(sections[4])
	} else {
		sysInfo = &SysInfo{OSDistro: "Linux", Kernel: "POSIX", Uptime: "Active", CPUCores: 1}
	}

	return &RawTelemetryBundle{
		Timestamp: time.Now(),
		CPUTick:   cpuSnap,
		Memory:    memStats,
		Disk:      diskStats,
		Network:   netStats,
		SysInfo:   sysInfo,
	}, nil
}
