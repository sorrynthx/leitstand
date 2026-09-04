package storage

import (
	"encoding/json"
	"time"
)

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

// SecretPayload represents the decrypted sensitive credentials for a host.
type SecretPayload struct {
	Password   string `json:"password,omitempty"`
	PrivateKey string `json:"private_key,omitempty"`
	KeyPath    string `json:"key_path,omitempty"`
	Passphrase string `json:"passphrase,omitempty"`
}

// Encode marshals the SecretPayload to JSON bytes.
func (s *SecretPayload) Encode() ([]byte, error) {
	return json.Marshal(s)
}

// ParseSecretPayload decodes decrypted raw bytes into a SecretPayload with fallback for legacy plain secrets.
func ParseSecretPayload(raw []byte, authMethod string) (*SecretPayload, error) {
	var payload SecretPayload
	if err := json.Unmarshal(raw, &payload); err == nil && (payload.Password != "" || payload.PrivateKey != "") {
		return &payload, nil
	}

	// Backward compatibility fallback for legacy plain text passwords/keys
	if authMethod == "private_key" {
		return &SecretPayload{PrivateKey: string(raw)}, nil
	}
	return &SecretPayload{Password: string(raw)}, nil
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

// SSHTunnel represents a port forwarding configuration.
type SSHTunnel struct {
	ID         int64     `json:"id"`
	HostID     int64     `json:"host_id"`
	Name       string    `json:"name"`
	LocalPort  int       `json:"local_port"`
	RemoteHost string    `json:"remote_host"`
	RemotePort int       `json:"remote_port"`
	AutoStart  bool      `json:"auto_start"`
	CreatedAt  time.Time `json:"created_at"`
}

// AIChatMessage represents a single conversation message in the AI Copilot.
type AIChatMessage struct {
	ID        int64     `json:"id"`
	HostID    int64     `json:"host_id"`
	Role      string    `json:"role"` // "user", "assistant", "system"
	Content   string    `json:"content"`
	CreatedAt time.Time `json:"created_at"`
}

