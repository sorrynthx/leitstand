package tui

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"

	"golang.org/x/crypto/ssh"
)

func (f *HostForm) validateAndExtract() (*HostFormData, error) {
	name := strings.TrimSpace(f.inputs[0].Value())
	if name == "" {
		return nil, errors.New("host name is required")
	}

	addr := strings.TrimSpace(f.inputs[1].Value())
	if addr == "" {
		return nil, errors.New("IP or address is required")
	}

	portStr := strings.TrimSpace(f.inputs[2].Value())
	port := 22
	if portStr != "" {
		p, err := strconv.Atoi(portStr)
		if err != nil || p <= 0 || p > 65535 {
			return nil, errors.New("port must be a valid number between 1 and 65535")
		}
		port = p
	}

	user := strings.TrimSpace(f.inputs[3].Value())
	if user == "" {
		return nil, errors.New("username is required")
	}

	group := strings.TrimSpace(f.inputs[7].Value())
	if group == "" {
		group = "Default"
	}

	data := &HostFormData{
		Name:     name,
		Address:  addr,
		Port:     port,
		Username: user,
		Group:    group,
	}

	if f.authType == AuthTypeKey {
		data.AuthMethod = "private_key"
		rawVal := strings.TrimSpace(f.inputs[5].Value())

		var keyBytes []byte
		var err error

		if strings.HasPrefix(rawVal, "[Encrypted") || rawVal == "" {
			if f.existingKey != "" {
				keyBytes = []byte(f.existingKey)
			} else {
				return nil, errors.New("private key path or key content is required")
			}
		} else if strings.HasPrefix(rawVal, "-----BEGIN ") {
			keyBytes = []byte(rawVal)
		} else {
			expandedKeyPath := expandHomePath(rawVal)
			keyBytes, err = os.ReadFile(expandedKeyPath)
			if err != nil {
				if f.existingKey != "" {
					keyBytes = []byte(f.existingKey)
				} else {
					return nil, fmt.Errorf("failed to read private key file '%s': %w", expandedKeyPath, err)
				}
			} else {
				data.KeyPath = expandedKeyPath
			}
		}

		keyBytes = bytes.TrimSpace(keyBytes)
		if len(keyBytes) == 0 {
			return nil, errors.New("private key content is empty")
		}

		data.KeyContent = string(keyBytes)

		passphrase := strings.TrimSpace(f.inputs[6].Value())
		if passphrase != "" {
			_, err = ssh.ParseRawPrivateKeyWithPassphrase(keyBytes, []byte(passphrase))
			if err != nil {
				return nil, fmt.Errorf("invalid private key or passphrase: %w", err)
			}
		} else {
			_, err = ssh.ParseRawPrivateKey(keyBytes)
			if err != nil {
				return nil, fmt.Errorf("invalid private key file (passphrase required?): %w", err)
			}
		}
		data.Passphrase = passphrase
	} else {
		data.AuthMethod = "password"
		pass := strings.TrimSpace(f.inputs[4].Value())
		if pass == "" && f.existingPass == "" {
			return nil, errors.New("password is required")
		}
		if pass == "" {
			pass = f.existingPass
		}
		data.Password = pass
	}

	return data, nil
}
