package storage

import "time"

// Host represents a registered target server.
type Host struct {
	ID        int64     `json:"id"`
	Name      string    `json:"name"`
	Address   string    `json:"address"`
	Port      int       `json:"port"`
	Username  string    `json:"username"`
	GroupName string    `json:"group_name"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// HostSecret holds encrypted authentication credentials for a host.
type HostSecret struct {
	HostID     int64  `json:"host_id"`
	AuthMethod string `json:"auth_method"` // "password", "private_key", "agent"
	Nonce      []byte `json:"nonce"`
	Ciphertext []byte `json:"ciphertext"`
}

// MetricRecord represents a single telemetry snapshot for a host.
type MetricRecord struct {
	HostID      int64     `json:"host_id"`
	Timestamp   time.Time `json:"timestamp"`
	CPUPercent  float64   `json:"cpu_percent"`
	MemoryTotal uint64    `json:"memory_total"`
	MemoryUsed  uint64    `json:"memory_used"`
	DiskUsed    uint64    `json:"disk_used"`
	DiskTotal   uint64    `json:"disk_total"`
	NetRxBytes  uint64    `json:"net_rx_bytes"`
	NetTxBytes  uint64    `json:"net_tx_bytes"`
}
