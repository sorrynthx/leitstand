package ssh

import (
	"fmt"
	"io"
	"leitstand/internal/logger"
	"leitstand/internal/storage"
	"net"
	"sync"
	"sync/atomic"
	"time"
)

// ActiveTunnel represents a live local port forwarding tunnel.
type ActiveTunnel struct {
	Tunnel      *storage.SSHTunnel
	client      *Client
	listener    net.Listener
	stopChan    chan struct{}
	closeOnce   sync.Once
	activeConns int64
	startedAt   time.Time
	lastErrMu   sync.RWMutex
	lastError   string
}

// NewActiveTunnel initializes and starts an ActiveTunnel.
func NewActiveTunnel(tun *storage.SSHTunnel, client *Client) (*ActiveTunnel, error) {
	if client == nil || !client.IsAlive() {
		return nil, fmt.Errorf("ssh client is not connected")
	}

	bindAddr := fmt.Sprintf("127.0.0.1:%d", tun.LocalPort)
	listener, err := net.Listen("tcp", bindAddr)
	if err != nil {
		return nil, fmt.Errorf("failed to bind local port %d: %w", tun.LocalPort, err)
	}

	at := &ActiveTunnel{
		Tunnel:    tun,
		client:    client,
		listener:  listener,
		stopChan:  make(chan struct{}),
		startedAt: time.Now(),
	}

	go at.acceptLoop()
	logger.Infof("Tunnel #%d (%s) started on %s -> %s:%d", tun.ID, tun.Name, bindAddr, tun.RemoteHost, tun.RemotePort)
	return at, nil
}

func (at *ActiveTunnel) acceptLoop() {
	defer at.listener.Close()

	for {
		conn, err := at.listener.Accept()
		if err != nil {
			select {
			case <-at.stopChan:
				return // Clean shutdown
			default:
				at.setLastError(err.Error())
				return
			}
		}

		go at.handleConn(conn)
	}
}

func (at *ActiveTunnel) handleConn(local net.Conn) {
	rawClient := at.client.RawClient()
	if rawClient == nil {
		local.Close()
		at.setLastError("ssh client connection lost")
		return
	}

	remoteAddr := fmt.Sprintf("%s:%d", at.Tunnel.RemoteHost, at.Tunnel.RemotePort)
	remote, err := rawClient.Dial("tcp", remoteAddr)
	if err != nil {
		local.Close()
		errMsg := fmt.Sprintf("failed to dial remote %s: %v", remoteAddr, err)
		at.setLastError(errMsg)
		logger.Warnf("Tunnel #%d: %s", at.Tunnel.ID, errMsg)
		return
	}

	atomic.AddInt64(&at.activeConns, 1)
	defer atomic.AddInt64(&at.activeConns, -1)

	var wg sync.WaitGroup
	wg.Add(2)

	// Local -> Remote
	go func() {
		defer wg.Done()
		_, _ = io.Copy(remote, local)
		if tc, ok := remote.(*net.TCPConn); ok {
			_ = tc.CloseWrite()
		}
	}()

	// Remote -> Local
	go func() {
		defer wg.Done()
		_, _ = io.Copy(local, remote)
		if tc, ok := local.(*net.TCPConn); ok {
			_ = tc.CloseWrite()
		}
	}()

	// Wait for completion or tunnel shutdown
	doneChan := make(chan struct{})
	go func() {
		wg.Wait()
		close(doneChan)
	}()

	select {
	case <-doneChan:
	case <-at.stopChan:
	}

	_ = local.Close()
	_ = remote.Close()
}

// Stop closes the local listener and terminates the tunnel.
func (at *ActiveTunnel) Stop() error {
	var err error
	at.closeOnce.Do(func() {
		close(at.stopChan)
		err = at.listener.Close()
		logger.Infof("Tunnel #%d (%s) stopped", at.Tunnel.ID, at.Tunnel.Name)
	})
	return err
}

// Conns returns the current number of active connections.
func (at *ActiveTunnel) Conns() int64 {
	return atomic.LoadInt64(&at.activeConns)
}

// LastError returns the most recent error message if any.
func (at *ActiveTunnel) LastError() string {
	at.lastErrMu.RLock()
	defer at.lastErrMu.RUnlock()
	return at.lastError
}

func (at *ActiveTunnel) setLastError(msg string) {
	at.lastErrMu.Lock()
	defer at.lastErrMu.Unlock()
	at.lastError = msg
}
