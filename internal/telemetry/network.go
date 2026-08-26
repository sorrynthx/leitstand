package telemetry

import (
	"bufio"
	"strconv"
	"strings"
)

// NetworkStats holds aggregated network traffic across physical interfaces.
type NetworkStats struct {
	RxBytes uint64
	TxBytes uint64
}

// ParseProcNetDev parses /proc/net/dev content, summing non-loopback interface counters.
func ParseProcNetDev(content string) (*NetworkStats, error) {
	scanner := bufio.NewScanner(strings.NewReader(content))
	var totalRx, totalTx uint64

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if !strings.Contains(line, ":") {
			continue
		}

		parts := strings.SplitN(line, ":", 2)
		if len(parts) != 2 {
			continue
		}

		iface := strings.TrimSpace(parts[0])
		// Ignore loopback interface
		if iface == "lo" {
			continue
		}

		fields := strings.Fields(parts[1])
		if len(fields) < 9 {
			continue
		}

		rx, _ := strconv.ParseUint(fields[0], 10, 64)
		tx, _ := strconv.ParseUint(fields[8], 10, 64)

		totalRx += rx
		totalTx += tx
	}

	return &NetworkStats{
		RxBytes: totalRx,
		TxBytes: totalTx,
	}, nil
}
