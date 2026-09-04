package tui

import (
	"fmt"
	"leitstand/internal/i18n"

	tea "github.com/charmbracelet/bubbletea"
)

func (m *Model) updateActiveModals(msg tea.Msg) (tea.Model, tea.Cmd, bool) {
	// Only intercept UI inputs (KeyMsg, MouseMsg, WindowSizeMsg) for modals.
	// Allow background stream messages (StreamChunkMsg, StreamFinishedMsg, TelemetryMsg, etc.) to pass through!
	switch msg.(type) {
	case tea.KeyMsg, tea.MouseMsg, tea.WindowSizeMsg:
		// Modal UI event processing continues below
	default:
		return m, nil, false
	}
	keyMsg, isKey := msg.(tea.KeyMsg)
	keyStr := ""
	if isKey {
		keyStr = keyMsg.String()
	}
	// 1. Settings / Preferences Modal
	if m.showSettingsModal && m.settingsModal != nil {
		res, cmd := m.settingsModal.Update(msg)
		if res.Done && !res.SaveReq {
			m.showSettingsModal = false
			m.settingsModal = nil
			m.statusMessage = "Settings closed."
			return m, nil, true
		}
		if res.SaveReq {
			m.applyAndPersistSettings(res)
			m.showSettingsModal = false
			m.settingsModal = nil
			m.statusMessage = "✨ " + i18n.T("settings_saved_msg")
			m.updateViewportContent()
			return m, nil, true
		}

		return m, cmd, true
	}
	// 2. In-app File Editor Modal
	if m.showEditorModal && m.editorModal != nil {
		done, saveReq, saveContent, cmd := m.editorModal.Update(msg)
		if done {
			m.showEditorModal = false
			m.editorModal = nil
			m.statusMessage = "Editor closed."
			return m, nil, true
		}
		if saveReq {
			m.statusMessage = fmt.Sprintf("💾 Saving '%s'...", m.editorModal.FilePath)
			return m, m.saveRemoteFileCmd(m.editorModal.HostID, m.editorModal.FilePath, saveContent), true
		}
		return m, cmd, true
	}

	// 3. Vault Unlock/Init Modal
	if m.showVaultModal {
		vf := m.vaultModal
		if vf == nil {
			vf = m.vaultForm
		}
		if vf != nil {
			prevLang := i18n.GetLang()
			done, pass, cmd := vf.Update(msg)
			if i18n.GetLang() != prevLang && m.store != nil {
				_ = m.store.SetSetting("language", string(i18n.GetLang()))
			}
			if done {
				if pass == "" {
					m.cancel()
					return m, tea.Quit, true
				}

				isInit, _ := m.store.IsVaultInitialized()
				var err error
				if !isInit {
					err = m.store.InitVault(m.vault, pass)
				} else {
					err = m.store.UnlockVault(m.vault, pass)
				}

				if err == nil && m.store != nil {
					_ = m.store.SetSetting("language", string(i18n.GetLang()))
				}

				if err != nil {
					vf.SetError(err)
					return m, nil, true
				}

				m.showVaultModal = false
				m.activePane = PaneHostList
				m.selectedIndex = 0
				m.userHasNavigated = true
				m.statusMessage = "✨ Vault unlocked successfully!"
				return m, tea.Batch(m.loadHostsCmd(), m.pollActiveHostsCmd()), true
			}
			return m, cmd, true
		}
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
					delete(m.hostTabs, m.hostToDelete.ID)
					m.statusMessage = fmt.Sprintf("🗑️ Host '%s' removed successfully.", m.hostToDelete.Name)
					m.showDeleteModal = false
					m.hostToDelete = nil
					return m, m.loadHostsCmd(), true
				}
				m.showDeleteModal = false
				m.hostToDelete = nil
				return m, nil, true

			case "n", "N", "esc":
				m.showDeleteModal = false
				m.hostToDelete = nil
				m.statusMessage = "Deletion canceled."
				return m, nil, true
			}
		}
		return m, nil, true
	}

	// 5. Add / Edit Host Modal
	if m.showAddModal && m.addForm != nil {
		done, formData, cmd := m.addForm.Update(msg)
		if done {
			if formData == nil {
				m.showAddModal = false
				m.addForm = nil
				m.statusMessage = "Add host canceled."
				return m, nil, true
			}
			m.showAddModal = false
			m.addForm = nil
			m.statusMessage = fmt.Sprintf("⏳ Saving host '%s'...", formData.Name)
			return m, m.saveNewHostCmd(formData), true
		}
		return m, cmd, true
	}

	if m.showEditModal && m.editForm != nil {
		done, formData, cmd := m.editForm.Update(msg)
		if done {
			if formData == nil {
				m.showEditModal = false
				m.editForm = nil
				m.hostToEdit = nil
				m.statusMessage = "Edit host canceled."
				return m, nil, true
			}
			hostID := m.hostToEdit.ID
			m.showEditModal = false
			m.editForm = nil
			m.hostToEdit = nil
			m.statusMessage = fmt.Sprintf("⏳ Updating host '%s'...", formData.Name)
			return m, m.updateExistingHostCmd(hostID, formData), true
		}
		return m, cmd, true
	}

	// 6. Root/Sudo Elevation Password Modal
	if m.showSudoModal && m.sudoModal != nil {
		done, pass, _, remember, cmd := m.sudoModal.Update(msg)
		if done {
			m.showSudoModal = false
			m.sudoModal = nil

			if pass == "" {
				m.statusMessage = "Elevation canceled."
				return m, nil, true
			}

			if remember && m.selectedIndex >= 0 && m.selectedIndex < len(m.hosts) {
				curHost := m.hosts[m.selectedIndex]
				m.sudoCache[curHost.ID] = pass
			}

			if m.pendingSudoCmd == "su" {
				curHost := m.hosts[m.selectedIndex]
				hts := m.GetOrCreateHostTabs(curHost.ID, curHost.Name)
				activeTab := hts.ActiveTab()
				m.statusMessage = i18n.Tf("sudo_auth_in_progress", curHost.Name)
				return m, m.execSudoValidateAndElevateCmd(curHost, activeTab, pass, remember), true
			}

			cmdToRun := m.pendingSudoCmd
			m.pendingSudoCmd = ""
			curHost := m.hosts[m.selectedIndex]
			m.statusMessage = fmt.Sprintf("⏳ Executing: '%s'...", cmdToRun)
			return m, m.execSudoCmd(curHost, cmdToRun, pass), true
		}
		return m, cmd, true
	}

	// 7. File Manager Modal
	if m.showFileManager && m.fileManager != nil {
		done, cmd := m.fileManager.Update(msg)
		if done {
			m.showFileManager = false
			if m.fileManager.IsTransferring {
				m.statusMessage = i18n.T("sftp_bg_transfer_switched")
			} else {
				m.fileManager = nil
				m.statusMessage = "📂 File manager closed."
			}
			return m, nil, true
		}
		return m, cmd, true
	}

	// 8. Tunnel Manager Modal
	if model, cmd, handled := m.updateTunnelModal(msg); handled {
		return model, cmd, true
	}

	// 9. AI Copilot Modal
	if model, cmd, handled := m.UpdateAICopilot(msg); handled {
		return model, cmd, true
	}

	return m, nil, false
}

