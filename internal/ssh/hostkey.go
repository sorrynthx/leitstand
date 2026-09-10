package ssh

import (
	"errors"
	"fmt"
	"net"
	"os"
	"path/filepath"
	"sync"

	"golang.org/x/crypto/ssh"
	"golang.org/x/crypto/ssh/knownhosts"
)

var (
	knownHostsMu sync.Mutex
)

// DefaultKnownHostsPath returns the standard path to the user's ~/.ssh/known_hosts file.
func DefaultKnownHostsPath() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return filepath.Join(".", ".ssh", "known_hosts")
	}
	return filepath.Join(home, ".ssh", "known_hosts")
}

// GetHostKeyCallback returns an ssh.HostKeyCallback enforcing the given policy.
// Supported policies:
// - "accept-new" (default, TOFU): auto-registers new hosts; blocks changed keys (MITM defense).
// - "strict": strictly rejects unknown hosts and mismatched keys.
// - "insecure": ignores host keys (not recommended for public networks).
func GetHostKeyCallback(knownHostsPath string, policy string) (ssh.HostKeyCallback, error) {
	if policy == "insecure" {
		return ssh.InsecureIgnoreHostKey(), nil
	}

	if policy == "" {
		policy = "accept-new"
	}

	if knownHostsPath == "" {
		knownHostsPath = DefaultKnownHostsPath()
	}

	// Ensure ~/.ssh directory and known_hosts file exist
	dir := filepath.Dir(knownHostsPath)
	if err := os.MkdirAll(dir, 0700); err != nil {
		return nil, fmt.Errorf("failed to create ssh directory %s: %w", dir, err)
	}

	knownHostsMu.Lock()
	if _, err := os.Stat(knownHostsPath); os.IsNotExist(err) {
		if f, createErr := os.OpenFile(knownHostsPath, os.O_CREATE|os.O_WRONLY, 0600); createErr == nil {
			_ = f.Close()
		}
	}
	knownHostsMu.Unlock()

	return func(hostname string, remote net.Addr, key ssh.PublicKey) error {
		knownHostsMu.Lock()
		defer knownHostsMu.Unlock()

		khCallback, err := knownhosts.New(knownHostsPath)
		if err != nil {
			return fmt.Errorf("failed to read known_hosts file (%s): %w", knownHostsPath, err)
		}

		checkErr := khCallback(hostname, remote, key)
		if checkErr == nil {
			// Host is known and verified!
			return nil
		}

		var keyErr *knownhosts.KeyError
		if errors.As(checkErr, &keyErr) {
			if len(keyErr.Want) == 0 {
				// Host is unknown (first connection)
				if policy == "accept-new" {
					return appendHostKey(knownHostsPath, hostname, remote, key)
				}
				return fmt.Errorf("host %s is unknown and policy is strict: %w", hostname, checkErr)
			}

			// Host key mismatch! Possible MITM attack!
			return fmt.Errorf("🚨 HOST KEY MISMATCH DETECTED for %s! Possible Man-In-The-Middle (MITM) attack or host key changed. Existing key in %s does not match remote key", hostname, knownHostsPath)
		}

		var revokedErr *knownhosts.RevokedError
		if errors.As(checkErr, &revokedErr) {
			return fmt.Errorf("🚨 REVOKED HOST KEY for %s! The host key has been explicitly revoked in %s", hostname, knownHostsPath)
		}

		return checkErr
	}, nil
}

// appendHostKey appends a new host's public key entry to the known_hosts file.
func appendHostKey(knownHostsPath string, hostname string, remote net.Addr, key ssh.PublicKey) error {
	f, err := os.OpenFile(knownHostsPath, os.O_APPEND|os.O_WRONLY, 0600)
	if err != nil {
		return fmt.Errorf("failed to open known_hosts (%s) for writing: %w", knownHostsPath, err)
	}
	defer f.Close()

	addresses := []string{knownhosts.Normalize(hostname)}
	if remote != nil {
		remoteAddr := remote.String()
		if remoteAddr != hostname {
			normalizedRemote := knownhosts.Normalize(remoteAddr)
			if normalizedRemote != addresses[0] {
				addresses = append(addresses, normalizedRemote)
			}
		}
	}

	line := knownhosts.Line(addresses, key)
	if _, err := f.WriteString(line + "\n"); err != nil {
		return fmt.Errorf("failed to write host key to %s: %w", knownHostsPath, err)
	}
	return nil
}
