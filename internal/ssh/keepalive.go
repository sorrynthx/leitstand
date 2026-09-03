package ssh

import (
	"time"
)

// StartKeepAlive starts a background ticker that sends keepalive requests to the remote SSH server.
func (c *Client) StartKeepAlive(interval time.Duration) {
	c.mu.Lock()
	if c.keepAliveStop != nil || c.rawClient == nil {
		c.mu.Unlock()
		return
	}
	stopChan := make(chan struct{})
	c.keepAliveStop = stopChan
	c.mu.Unlock()

	if interval <= 0 {
		interval = 30 * time.Second
	}

	go func() {
		ticker := time.NewTicker(interval)
		defer ticker.Stop()

		for {
			select {
			case <-stopChan:
				return
			case <-ticker.C:
				c.mu.Lock()
				currentRaw := c.rawClient
				c.mu.Unlock()
				if currentRaw == nil {
					return
				}
				// OpenSSH standard keepalive request
				_, _, _ = currentRaw.SendRequest("keepalive@openssh.com", true, nil)
			}
		}
	}()
}

// StopKeepAlive halts the active keepalive background ticker.
func (c *Client) StopKeepAlive() {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.keepAliveStop != nil {
		close(c.keepAliveStop)
		c.keepAliveStop = nil
	}
}
