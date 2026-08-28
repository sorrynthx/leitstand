package telemetry

import (
	"context"
	"fmt"
	"leitstand/internal/logger"
	"leitstand/internal/ssh"
	"leitstand/internal/storage"
	"leitstand/internal/vault"
	"sync"
	"time"
)

// HostTelemetryState tracks historical tick snapshots to compute deltas.
type HostTelemetryState struct {
	LastCPUTick   *CPUTickSnapshot
	LastNetwork   *NetworkStats
	LastTimestamp time.Time
	LastRecord    *storage.MetricRecord
	SysInfo       *SysInfo
	LastError     error
}

// Collector coordinates telemetry collection across managed hosts.
type Collector struct {
	mu         sync.RWMutex
	store      *storage.Storage
	pool       *ssh.Pool
	vault      *vault.Vault
	hostStates map[int64]*HostTelemetryState
}

// NewCollector creates a telemetry collector.
func NewCollector(store *storage.Storage, pool *ssh.Pool, v *vault.Vault) *Collector {
	return &Collector{
		store:      store,
		pool:       pool,
		vault:      v,
		hostStates: make(map[int64]*HostTelemetryState),
	}
}

// Pool returns the underlying SSH connection pool.
func (c *Collector) Pool() *ssh.Pool {
	return c.pool
}

// CollectHost polls telemetry for a single host.
func (c *Collector) CollectHost(host *storage.Host) (*storage.MetricRecord, error) {
	if !c.vault.IsUnlocked() {
		return nil, vault.ErrVaultLocked
	}

	secret, err := c.store.GetHostSecret(host.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to load secret for host %d: %w", host.ID, err)
	}

	decrypted, err := c.vault.Decrypt(secret.Nonce, secret.Ciphertext)
	if err != nil {
		return nil, fmt.Errorf("failed to decrypt secret for host %d: %w", host.ID, err)
	}
	defer vault.ZeroBytes(decrypted)

	payload, err := storage.ParseSecretPayload(decrypted, secret.AuthMethod)
	if err != nil {
		return nil, fmt.Errorf("failed to parse secret payload for host %d: %w", host.ID, err)
	}

	client, err := c.pool.GetOrCreateFromPayload(host, secret, payload)
	if err != nil {
		logger.Warnf("CollectHost: SSH connection failed for host %s (%d): %v", host.Name, host.ID, err)
		c.recordHostError(host.ID, err)
		return nil, fmt.Errorf("ssh connection failed: %w", err)
	}

	stdout, stderr, err := client.Exec(MetricExtractScript)
	if err != nil {
		logger.Warnf("CollectHost: Exec failed for host %s (%d): %v (stderr: %s)", host.Name, host.ID, err, string(stderr))
		c.recordHostError(host.ID, fmt.Errorf("exec error: %w (stderr: %s)", err, string(stderr)))
		return nil, err
	}

	bundle, err := ParseRawBundle(string(stdout))
	if err != nil {
		logger.Warnf("CollectHost: ParseRawBundle failed for host %s (%d): %v", host.Name, host.ID, err)
		c.recordHostError(host.ID, err)
		return nil, fmt.Errorf("telemetry parse error: %w", err)
	}

	c.mu.Lock()
	state, exists := c.hostStates[host.ID]
	if !exists {
		state = &HostTelemetryState{}
		c.hostStates[host.ID] = state
	}

	var cpuPct float64
	if state.LastCPUTick != nil {
		cpuPct = CalculateCPUPercent(state.LastCPUTick, bundle.CPUTick)
	}

	var rxRate, txRate uint64
	if state.LastNetwork != nil && !state.LastTimestamp.IsZero() {
		elapsed := bundle.Timestamp.Sub(state.LastTimestamp).Seconds()
		if elapsed > 0 {
			if bundle.Network.RxBytes >= state.LastNetwork.RxBytes {
				rxRate = uint64(float64(bundle.Network.RxBytes-state.LastNetwork.RxBytes) / elapsed)
			}
			if bundle.Network.TxBytes >= state.LastNetwork.TxBytes {
				txRate = uint64(float64(bundle.Network.TxBytes-state.LastNetwork.TxBytes) / elapsed)
			}
		}
	}

	state.LastCPUTick = bundle.CPUTick
	state.LastNetwork = bundle.Network
	state.LastTimestamp = bundle.Timestamp
	state.SysInfo = bundle.SysInfo
	state.LastError = nil

	record := &storage.MetricRecord{
		HostID:      host.ID,
		Timestamp:   bundle.Timestamp,
		CPUPercent:  cpuPct,
		MemoryTotal: bundle.Memory.Total,
		MemoryUsed:  bundle.Memory.Used,
		DiskUsed:    bundle.Disk.Used,
		DiskTotal:   bundle.Disk.Total,
		NetRxBytes:  rxRate,
		NetTxBytes:  txRate,
	}
	state.LastRecord = record
	c.mu.Unlock()

	return record, nil
}

func (c *Collector) recordHostError(hostID int64, err error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	state, exists := c.hostStates[hostID]
	if !exists {
		state = &HostTelemetryState{}
		c.hostStates[hostID] = state
	}
	state.LastError = err
}

// GetHostState returns the latest cached state of a host.
func (c *Collector) GetHostState(hostID int64) (*HostTelemetryState, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	state, exists := c.hostStates[hostID]
	if !exists {
		return nil, false
	}
	return state, true
}

// StartPollingLoop runs periodic polling across all active hosts.
func (c *Collector) StartPollingLoop(ctx context.Context, interval time.Duration, onUpdate func(hostID int64, record *storage.MetricRecord, err error)) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	// Initial poll immediately
	c.pollAll(onUpdate)

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			c.pollAll(onUpdate)
		}
	}
}

func (c *Collector) pollAll(onUpdate func(hostID int64, record *storage.MetricRecord, err error)) {
	hosts, err := c.store.ListHosts()
	if err != nil {
		return
	}

	var wg sync.WaitGroup
	for _, h := range hosts {
		wg.Add(1)
		go func(host *storage.Host) {
			defer wg.Done()
			rec, err := c.CollectHost(host)
			if onUpdate != nil {
				onUpdate(host.ID, rec, err)
			}
		}(h)
	}
	wg.Wait()
}
