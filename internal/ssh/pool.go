package ssh

import (
	"fmt"
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
func (p *Pool) GetOrCreate(host *storage.Host, secret *storage.HostSecret, decryptedSecret []byte, passphrase []byte) (*Client, error) {
	p.mu.Lock()
	defer p.mu.Unlock()

	// Check if already in pool and healthy
	if client, exists := p.clients[host.ID]; exists {
		if client.IsAlive() {
			return client, nil
		}
		// Connection dropped, close and remove from pool
		_ = client.Close()
		delete(p.clients, host.ID)
	}

	// Build address and client config
	addr := BuildAddress(host.Address, host.Port)
	cfg, err := BuildClientConfig(host.Username, secret.AuthMethod, decryptedSecret, passphrase, p.timeout)
	if err != nil {
		return nil, fmt.Errorf("failed to build ssh config for host %s (%d): %w", host.Name, host.ID, err)
	}

	client, err := NewClient(addr, cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to host %s (%s): %w", host.Name, addr, err)
	}

	p.clients[host.ID] = client
	return client, nil
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
