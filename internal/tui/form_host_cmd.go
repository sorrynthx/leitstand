package tui

import (
	"os"
	"path/filepath"
	"strings"

	"leitstand/internal/storage"

	tea "github.com/charmbracelet/bubbletea"
)

func expandHomePath(p string) string {
	p = strings.Trim(p, "\"'` ")
	if strings.HasPrefix(p, "~/") || strings.HasPrefix(p, "~\\") || p == "~" {
		home, err := os.UserHomeDir()
		if err == nil {
			if p == "~" {
				return home
			}
			return filepath.Join(home, p[2:])
		}
	}
	return p
}

type HostSavedMsg struct {
	HostName string
}

func (m *Model) saveNewHostCmd(data *HostFormData) tea.Cmd {
	return func() tea.Msg {
		if m.store == nil || m.vault == nil {
			return nil
		}

		h := &storage.Host{
			Name:      data.Name,
			Address:   data.Address,
			Port:      data.Port,
			Username:  data.Username,
			GroupName: data.Group,
		}

		hostID, err := m.store.CreateHost(h)
		if err != nil {
			return nil
		}

		payload := &storage.SecretPayload{
			Password:   data.Password,
			PrivateKey: data.KeyContent,
			KeyPath:    data.KeyPath,
			Passphrase: data.Passphrase,
		}

		encoded, err := payload.Encode()
		if err != nil {
			return nil
		}

		nonce, ciphertext, err := m.vault.Encrypt(encoded)
		if err != nil {
			return nil
		}

		secret := &storage.HostSecret{
			HostID:     hostID,
			AuthMethod: data.AuthMethod,
			Nonce:      nonce,
			Ciphertext: ciphertext,
		}

		_ = m.store.SaveHostSecret(secret)
		return HostSavedMsg{HostName: data.Name}
	}
}

func (m *Model) updateExistingHostCmd(hostID int64, data *HostFormData) tea.Cmd {
	return func() tea.Msg {
		if m.store == nil || m.vault == nil {
			return nil
		}

		h, err := m.store.GetHost(hostID)
		if err != nil || h == nil {
			return nil
		}

		h.Name = data.Name
		h.Address = data.Address
		h.Port = data.Port
		h.Username = data.Username
		h.GroupName = data.Group

		_ = m.store.UpdateHost(h)

		payload := &storage.SecretPayload{
			Password:   data.Password,
			PrivateKey: data.KeyContent,
			KeyPath:    data.KeyPath,
			Passphrase: data.Passphrase,
		}

		encoded, err := payload.Encode()
		if err != nil {
			return nil
		}

		nonce, ciphertext, err := m.vault.Encrypt(encoded)
		if err != nil {
			return nil
		}

		secret := &storage.HostSecret{
			HostID:     hostID,
			AuthMethod: data.AuthMethod,
			Nonce:      nonce,
			Ciphertext: ciphertext,
		}

		_ = m.store.SaveHostSecret(secret)
		return HostSavedMsg{HostName: data.Name}
	}
}
