package telemetry

import (
	"bufio"
	"errors"
	"strconv"
	"strings"
)

// DiskStats holds disk storage usage in bytes.
type DiskStats struct {
	Total     uint64 // bytes
	Used      uint64 // bytes
	Available uint64 // bytes
}

// ParseDF parses `df -k /` output to extract root filesystem metrics.
func ParseDF(content string) (*DiskStats, error) {
	scanner := bufio.NewScanner(strings.NewReader(content))
	var lastLine string

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line != "" && !strings.HasPrefix(line, "Filesystem") {
			lastLine = line
		}
	}

	if lastLine == "" {
		return nil, errors.New("no disk data rows found in df output")
	}

	fields := strings.Fields(lastLine)
	if len(fields) < 4 {
		return nil, errors.New("malformed df output line")
	}

	totalKb, err := strconv.ParseUint(fields[1], 10, 64)
	if err != nil {
		return nil, err
	}
	usedKb, err := strconv.ParseUint(fields[2], 10, 64)
	if err != nil {
		return nil, err
	}
	availKb, err := strconv.ParseUint(fields[3], 10, 64)
	if err != nil {
		return nil, err
	}

	return &DiskStats{
		Total:     totalKb * 1024,
		Used:      usedKb * 1024,
		Available: availKb * 1024,
	}, nil
}
