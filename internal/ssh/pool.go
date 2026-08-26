package ssh

import (
	"fmt"
	"leitstand/internal/logger"
	"leitstand/internal/storage"
	"sync"
	"time"
)

// Pool manages persistent, multiplexed SSH connections per host.
type Pool struct {
	mu      sync.RWMutex
	clients map[int64]*Client
	timeout time.Duration
}

// NewPool creates a new SSH connection pool.
func NewPool(defaultTimeout time.Duration) *Pool {
	if defaultTimeout <= 0 {
		defaultTimeout = 10 * time.Second
	}
	return &Pool{
		clients: make(map[int64]*Client),
		timeout: defaultTimeout,
	}
}

// Get retrieves an existing active client from the pool without reconnecting.
func (p *Pool) Get(hostID int64) (*Client, bool) {
	p.mu.RLock()
	defer p.mu.RUnlock()

	client, exists := p.clients[hostID]
	if !exists {
		return nil, false
	}
	return client, true
}

// GetOrCreate returns an existing live client or establishes a new SSH connection.
// IMPORTANT: Network Dialing (ssh.Dial) is performed OUTSIDE the pool mutex
// so that a slow/unreachable host does NOT block connections to other hosts.
func (p *Pool) GetOrCreate(host *storage.Host, secret *storage.HostSecret, decryptedSecret []byte, passphrase []byte) (*Client, error) {
	// 1. Quick check if already connected and alive (Read Lock only)
	p.mu.RLock()
	if client, exists := p.clients[host.ID]; exists {
		if client.IsAlive() {
			p.mu.RUnlock()
			return client, nil
		}
	}
	p.mu.RUnlock()

	// 2. Build configuration and perform ssh.Dial outside the lock
	addr := BuildAddress(host.Address, host.Port)
	cfg, err := BuildClientConfig(host.Username, secret.AuthMethod, decryptedSecret, passphrase, p.timeout)
	if err != nil {
		return nil, fmt.Errorf("failed to build ssh config for host %s (%d): %w", host.Name, host.ID, err)
	}

	logger.Debugf("Pool: Dialing SSH to host %s (%s)...", host.Name, addr)
	newClient, err := NewClient(addr, cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to host %s (%s): %w", host.Name, addr, err)
	}

	// 3. Register client in pool (short Write Lock)
	p.mu.Lock()
	if existing, exists := p.clients[host.ID]; exists && existing.IsAlive() {
		p.mu.Unlock()
		_ = newClient.Close()
		return existing, nil
	}

	// Close previous stale client if any
	if old, exists := p.clients[host.ID]; exists {
		_ = old.Close()
	}

	p.clients[host.ID] = newClient
	p.mu.Unlock()

	logger.Infof("Pool: Established live SSH connection to host %s (%s)", host.Name, addr)
	return newClient, nil
}

// CloseHost closes and removes a specific host connection.
func (p *Pool) CloseHost(hostID int64) {
	p.mu.Lock()
	defer p.mu.Unlock()

	if client, exists := p.clients[hostID]; exists {
		_ = client.Close()
		delete(p.clients, hostID)
	}
}

// CloseAll terminates all managed connections.
func (p *Pool) CloseAll() {
	p.mu.Lock()
	defer p.mu.Unlock()

	for id, client := range p.clients {
		_ = client.Close()
		delete(p.clients, id)
	}
}

// ActiveCount returns the number of active host connections.
func (p *Pool) ActiveCount() int {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return len(p.clients)
}
