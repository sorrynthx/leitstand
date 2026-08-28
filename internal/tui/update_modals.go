package tui

import (
	"fmt"
	"leitstand/internal/i18n"

	tea "github.com/charmbracelet/bubbletea"
)

// updateActiveModals processes events for any currently open modal dialogs.
// Returns (updatedModel, teaCmd, wasModalActive).
func (m *Model) updateActiveModals(msg tea.Msg) (tea.Model, tea.Cmd, bool) {
	keyMsg, isKey := msg.(tea.KeyMsg)
	keyStr := ""
	if isKey {
		keyStr = keyMsg.String()
	}

	// 1. Settings / Preferences Modal
	if m.showSettingsModal && m.settingsModal != nil {
		done, saveReq, lang, interval, currPass, newPass, cmd := m.settingsModal.Update(msg)
		if done && !saveReq {
			m.showSettingsModal = false
			m.settingsModal = nil
			m.statusMessage = "Settings closed."
			return m, nil, true
		}
		if saveReq {
			i18n.SetLang(lang)
			if m.cfg != nil {
				m.cfg.Telemetry.PollingInterval = interval
			}
			if m.store != nil {
				_ = m.store.SetSetting("language", string(lang))
				_ = m.store.SetSetting("polling_interval", interval.String())
			}

			if newPass != "" {
				err := m.changeVaultPassword(currPass, newPass)
				if err != nil {
					m.settingsModal.SetError(err)
					return m, nil, true
				}
			}

			m.showSettingsModal = false
			m.settingsModal = nil
			m.statusMessage = i18n.T("settings_saved")
			return m, m.tickCmd(), true
		}
		return m, cmd, true
	}

	// 1.1 Dual-Pane SFTP File Manager Modal
	if m.showFileManager && m.fileManager != nil {
		done, cmd := m.fileManager.Update(msg)
		if done {
			m.showFileManager = false
			m.fileManager = nil
			m.statusMessage = "📂 File manager closed."
			return m, nil, true
		}
		return m, cmd, true
	}

	// 2. In-app File Editor Modal
	if m.showEditorModal && m.editorModal != nil {
		done, saveReq, updatedContent, cmd := m.editorModal.Update(msg)
		if done {
			m.showEditorModal = false
			m.editorModal = nil
			m.statusMessage = "Editor closed."
			return m, nil, true
		}
		if saveReq {
			m.statusMessage = "⏳ Saving file to remote server..."
			return m, m.saveRemoteFileCmd(m.editorModal.HostID, m.editorModal.FilePath, updatedContent), true
		}
		return m, cmd, true
	}

	// 3. Vault Unlock/Init Modal
	if m.showVaultModal && m.vaultForm != nil {
		done, pass, cmd := m.vaultForm.Update(msg)
		if done {
			if pass == "" {
				m.cancel()
				return m, tea.Quit, true
			}

			isInit, _ := m.store.IsVaultInitialized()
			var err error
			if !isInit {
				err = m.store.InitVault(m.vault, pass)
				if err == nil && m.store != nil {
					_ = m.store.SetSetting("language", string(i18n.GetLang()))
				}
			} else {
				err = m.store.UnlockVault(m.vault, pass)
			}

			if err != nil {
				m.vaultForm.SetError(err)
				return m, nil, true
			}

			m.showVaultModal = false
			m.statusMessage = "✨ Vault unlocked successfully!"
			return m, tea.Batch(m.loadHostsCmd(), m.pollActiveHostsCmd()), true
		}
		return m, cmd, true
	}

	// 4. Delete Host Confirmation Modal
	if m.showDeleteModal {
		if isKey {
			switch keyStr {
			case "y", "Y", "enter":
				if m.hostToDelete != nil && m.store != nil {
					_ = m.store.DeleteHost(m.hostToDelete.ID)
					delete(m.metrics, m.hostToDelete.ID)
					delete(m.errors, m.hostToDelete.ID)
					delete(m.consoleLogs, m.hostToDelete.ID)
					delete(m.hostTabs, m.hostToDelete.ID)
					m.statusMessage = fmt.Sprintf("🗑️ Host '%s' removed successfully.", m.hostToDelete.Name)
					m.showDeleteModal = false
					m.hostToDelete = nil
					return m, m.loadHostsCmd(), true
				}
				m.showDeleteModal = false
				return m, nil, true

			case "n", "N", "esc":
				m.showDeleteModal = false
				m.hostToDelete = nil
				m.statusMessage = "Delete cancelled."
				return m, nil, true
			}
		}
		return m, nil, true
	}

	// 5. Add Host Modal
	if m.showAddModal && m.addForm != nil {
		done, data, cmd := m.addForm.Update(msg)
		if done {
			m.showAddModal = false
			if data != nil {
				return m, m.saveNewHostCmd(data), true
			}
		}
		return m, cmd, true
	}

	// 6. Edit Host Modal
	if m.showEditModal && m.editForm != nil {
		done, data, cmd := m.editForm.Update(msg)
		if done {
			m.showEditModal = false
			if data != nil && m.hostToEdit != nil {
				return m, m.updateExistingHostCmd(m.hostToEdit.ID, data), true
			}
		}
		return m, cmd, true
	}

	// 7. Sudo Password Modal
	if m.showSudoModal && m.sudoModal != nil {
		done, pass, remember, cmd := m.sudoModal.Update(msg)
		if done {
			m.showSudoModal = false
			if pass != "" && len(m.hosts) > 0 {
				curHost := m.hosts[m.selectedIndex]
				if remember {
					m.sudoCache[curHost.ID] = pass
				}
				if m.pendingSudoCmd == "su" {
					tab := m.CurrentActiveTab()
					if tab != nil {
						tab.IsRoot = true
						hts := m.GetOrCreateHostTabs(curHost.ID, curHost.Name)
						tab.SetAutoTitle(hts.ActiveIndex, "")
						tab.AppendLog(fmt.Sprintf("✨ [ROOT Session] Authenticated with root privileges on %s.\nCommands in this tab run as root (root#). Type 'exit' to log out.", curHost.Name))
						m.updateViewportContent()
						m.statusMessage = "👑 Root session activated for this tab."
						return m, nil, true
					}
				}
				m.statusMessage = fmt.Sprintf("⏳ Executing with elevated privilege: '%s'...", m.pendingSudoCmd)
				return m, m.execSudoCmd(curHost, m.pendingSudoCmd, pass), true
			}
			m.statusMessage = "Sudo command cancelled."
		}
		return m, cmd, true
	}

	return m, nil, false
}
