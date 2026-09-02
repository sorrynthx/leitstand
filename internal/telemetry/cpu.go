package telemetry

import (
	"bufio"
	"errors"
	"strconv"
	"strings"
)

// CPUTickSnapshot holds raw CPU jiffies from /proc/stat.
type CPUTickSnapshot struct {
	User    uint64
	Nice    uint64
	System  uint64
	Idle    uint64
	IOWait  uint64
	IRQ     uint64
	SoftIRQ uint64
	Steal   uint64
}

// Total returns the sum of all CPU jiffies.
func (c *CPUTickSnapshot) Total() uint64 {
	return c.User + c.Nice + c.System + c.Idle + c.IOWait + c.IRQ + c.SoftIRQ + c.Steal
}

// TotalIdle returns the idle and iowait jiffies.
func (c *CPUTickSnapshot) TotalIdle() uint64 {
	return c.Idle + c.IOWait
}

// ParseProcStatLine parses a single 'cpu ' line into CPUTickSnapshot.
func ParseProcStatLine(line string) (*CPUTickSnapshot, error) {
	fields := strings.Fields(line)
	if len(fields) < 5 {
		return nil, errors.New("malformed /proc/stat cpu line")
	}

	var snap CPUTickSnapshot
	snap.User, _ = strconv.ParseUint(fields[1], 10, 64)
	snap.Nice, _ = strconv.ParseUint(fields[2], 10, 64)
	snap.System, _ = strconv.ParseUint(fields[3], 10, 64)
	snap.Idle, _ = strconv.ParseUint(fields[4], 10, 64)

	if len(fields) > 5 {
		snap.IOWait, _ = strconv.ParseUint(fields[5], 10, 64)
	}
	if len(fields) > 6 {
		snap.IRQ, _ = strconv.ParseUint(fields[6], 10, 64)
	}
	if len(fields) > 7 {
		snap.SoftIRQ, _ = strconv.ParseUint(fields[7], 10, 64)
	}
	if len(fields) > 8 {
		snap.Steal, _ = strconv.ParseUint(fields[8], 10, 64)
	}

	return &snap, nil
}

// ParseProcStat parses the first 'cpu' line from /proc/stat content.
func ParseProcStat(content string) (*CPUTickSnapshot, error) {
	scanner := bufio.NewScanner(strings.NewReader(content))
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if strings.HasPrefix(line, "cpu ") {
			return ParseProcStatLine(line)
		}
	}

	return nil, errors.New("no aggregate 'cpu' line found in /proc/stat")
}

// ParseDualProcStat parses two consecutive /proc/stat outputs and computes instant CPU%.
func ParseDualProcStat(content string) (float64, *CPUTickSnapshot, error) {
	scanner := bufio.NewScanner(strings.NewReader(content))
	var snaps []*CPUTickSnapshot

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if strings.HasPrefix(line, "cpu ") {
			s, err := ParseProcStatLine(line)
			if err == nil {
				snaps = append(snaps, s)
			}
		}
	}

	if len(snaps) >= 2 {
		cpuPct := CalculateCPUPercent(snaps[0], snaps[1])
		return cpuPct, snaps[1], nil
	} else if len(snaps) == 1 {
		return 0.0, snaps[0], nil
	}

	return 0.0, nil, errors.New("no cpu lines found")
}

// CalculateCPUPercent computes the CPU utilization percentage between two snapshots.
func CalculateCPUPercent(prev, curr *CPUTickSnapshot) float64 {
	if prev == nil || curr == nil {
		return 0.0
	}

	totalDelta := float64(curr.Total()) - float64(prev.Total())
	idleDelta := float64(curr.TotalIdle()) - float64(prev.TotalIdle())

	if totalDelta <= 0 {
		return 0.0
	}

	utilization := ((totalDelta - idleDelta) / totalDelta) * 100.0
	if utilization < 0.0 {
		return 0.0
	}
	if utilization > 100.0 {
		return 100.0
	}

	return utilization
}
