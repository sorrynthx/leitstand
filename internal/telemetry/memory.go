package telemetry

import (
	"bufio"
	"errors"
	"strconv"
	"strings"
)

// MemoryStats holds parsed memory metrics in bytes.
type MemoryStats struct {
	Total     uint64 // bytes
	Used      uint64 // bytes
	Available uint64 // bytes
	Free      uint64 // bytes
	Buffers   uint64 // bytes
	Cached    uint64 // bytes
}

// ParseMeminfo parses /proc/meminfo content and calculates memory usage.
func ParseMeminfo(content string) (*MemoryStats, error) {
	scanner := bufio.NewScanner(strings.NewReader(content))
	values := make(map[string]uint64)

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}
		parts := strings.SplitN(line, ":", 2)
		if len(parts) != 2 {
			continue
		}

		key := strings.TrimSpace(parts[0])
		valFields := strings.Fields(parts[1])
		if len(valFields) == 0 {
			continue
		}

		valKb, err := strconv.ParseUint(valFields[0], 10, 64)
		if err == nil {
			values[key] = valKb * 1024 // Convert kB to bytes
		}
	}

	total, hasTotal := values["MemTotal"]
	if !hasTotal || total == 0 {
		return nil, errors.New("missing or invalid MemTotal in /proc/meminfo")
	}

	free := values["MemFree"]
	buffers := values["Buffers"]
	cached := values["Cached"]

	// Prefer MemAvailable if present (Linux 3.14+)
	available, hasAvailable := values["MemAvailable"]
	var used uint64
	if hasAvailable {
		if total >= available {
			used = total - available
		}
	} else {
		// Fallback for older kernels: used = total - free - buffers - cached
		if total >= (free + buffers + cached) {
			used = total - free - buffers - cached
		}
		available = free + buffers + cached
	}

	return &MemoryStats{
		Total:     total,
		Used:      used,
		Available: available,
		Free:      free,
		Buffers:   buffers,
		Cached:    cached,
	}, nil
}
