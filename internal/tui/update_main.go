package tui

import (
	"fmt"
	"leitstand/internal/i18n"
	"leitstand/internal/storage"
	"leitstand/internal/vault"
	"time"

	tea "github.com/charmbracelet/bubbletea"
)

func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if winMsg, ok := msg.(tea.WindowSizeMsg); ok {
		m.width = winMsg.Width
		m.height = winMsg.Height
		m.initOrResizeViewport()
	}

	if updatedModel, cmd, handled := m.updateActiveModals(msg); handled {
		return updatedModel, cmd
	}

	switch msg := msg.(type) {

	case HostSavedMsg:
		m.statusMessage = fmt.Sprintf("✨ Host '%s' saved successfully!", msg.HostName)
		return m, m.loadHostsCmd()

	case tea.WindowSizeMsg:
		return m, nil

	case TickMsg:
		return m, tea.Batch(
			m.pollActiveHostsCmd(),
			m.tickCmd(),
		)

	case tea.KeyMsg:
		keyStr := msg.String()

		if keyStr == "f5" || keyStr == "F5" {
			m.showDrawer = false
			m.showTelemetryDrawer = !m.showTelemetryDrawer
			if m.showTelemetryDrawer {
				m.statusMessage = i18n.T("telemetry_deck_expanded")
			} else {
				m.statusMessage = i18n.T("telemetry_deck_collapsed")
			}
			m.updateViewportContent()
			return m, nil
		}

		if m.showTelemetryDrawer {
			if keyStr == "esc" || keyStr == "m" || keyStr == "M" || keyStr == "q" {
				m.showTelemetryDrawer = false
				m.updateViewportContent()
				return m, nil
			}
			if keyStr == "?" {
				m.showTelemetryDrawer = false
				m.updateViewportContent()
			}
		}

		if updatedModel, cmd, handled := m.updateRunbookDrawer(msg); handled {
			return updatedModel, cmd
		}

		if m.activePane == PaneConsole && len(m.hosts) > 0 && m.selectedIndex >= 0 && m.selectedIndex < len(m.hosts) {
			return m.updateConsoleKeys(msg, keyStr)
		}

		if updatedModel, cmd, handled := m.updateHostListNavigation(keyStr); handled {
			return updatedModel, cmd
		}

	case []*storage.Host:
		m.hosts = msg
		if len(m.hosts) > 0 {
			if m.selectedIndex >= len(m.hosts) {
				m.selectedIndex = 0
			}
			m.updateViewportContent()
			return m, m.pollActiveHostsCmd()
		}
		m.updateViewportContent()
		return m, nil

	case TerminalExitedMsg:
		m.statusMessage = "💻 Terminal session ended."
		return m, nil

	case SessionLogExportedMsg:
		if msg.Err != nil {
			m.statusMessage = fmt.Sprintf("❌ 세션 로그 저장 실패: %v", msg.Err)
		} else {
			m.statusMessage = fmt.Sprintf("💾 세션 로그 저장 완료: %s", msg.FilePath)
			if len(m.hosts) > 0 && m.selectedIndex >= 0 && m.selectedIndex < len(m.hosts) {
				curHost := m.hosts[m.selectedIndex]
				hts := m.GetOrCreateHostTabs(curHost.ID, curHost.Name)
				if tab := hts.ActiveTab(); tab != nil {
					tab.AppendLog(fmt.Sprintf("💾 [Session Log Exported] -> %s", msg.FilePath))
				}
			}
			if m.store != nil {
				_ = m.store.SetSetting("last_session_log_path", msg.FilePath)
				_ = m.store.SetSetting("last_session_log_time", time.Now().Format("2006-01-02 15:04:05"))
			}
		}
		m.updateViewportContent()
		return m, nil

	default:
		return m.handleMessage(msg)
	}

	return m, nil
}

func (m *Model) handleMessage(msg tea.Msg) (tea.Model, tea.Cmd) {
	if model, cmd, handled := m.tryHandleTelemetryMessage(msg); handled {
		return model, cmd
	}
	if model, cmd, handled := m.tryHandleSFTPMessage(msg); handled {
		return model, cmd
	}
	return m.handleCommandResultMessage(msg)
}

func (m *Model) changeVaultPassword(currPassword, newPassword string) error {
	if m.store == nil || m.vault == nil {
		return fmt.Errorf("vault or storage not available")
	}

	tempVerifyVault := vault.New()
	err := m.store.UnlockVault(tempVerifyVault, currPassword)
	if err != nil {
		return fmt.Errorf("current password verification failed: %w", err)
	}
	tempVerifyVault.Lock()

	newVault := vault.New()
	err = m.store.RekeyVault(m.vault, newVault, newPassword)
	if err != nil {
		return fmt.Errorf("failed to change master password: %w", err)
	}

	m.vault.Lock()
	m.vault = newVault
	return nil
}
