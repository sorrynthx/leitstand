package tui

import (
	"leitstand/internal/quickcmd"
	"leitstand/internal/storage"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
)

// RunbookDrawer provides an OS-aware quick command runbook.
type RunbookDrawer struct {
	activeTab           quickcmd.OSTab
	selectedIndex       int
	detectedDistro      string
	store               *storage.Storage
	customCmds          []*storage.CustomCommand
	isAdding            bool
	isEditing           bool
	isDeleting          bool
	editingID           int64
	customInputs        []textinput.Model
	customFocused       int
	filePicker          *FilePickerModal
	pendingPickerAction int // 1 = Export, 2 = Import
	bannerMessage       string
	errMessage          string
	toastMessage        string
}

// NewRunbookDrawer creates a new runbook drawer, prioritizing Custom commands if present, or auto-focusing on the detected host OS.
func NewRunbookDrawer(detectedDistro string, store *storage.Storage) *RunbookDrawer {
	autoTab := quickcmd.DetectOSTab(detectedDistro)
	d := &RunbookDrawer{
		activeTab:      autoTab,
		selectedIndex:  0,
		detectedDistro: detectedDistro,
		store:          store,
	}
	d.loadCustomCommands()
	if len(d.customCmds) > 0 {
		d.activeTab = quickcmd.OSTabCustom
	}
	return d
}

func (d *RunbookDrawer) getItems() []quickcmd.CommandItem {
	if d.activeTab == quickcmd.OSTabCustom {
		return d.getCustomItems()
	}
	return quickcmd.Catalog[d.activeTab]
}

// Update handles key navigation inside the Runbook Drawer.
func (d *RunbookDrawer) Update(msg tea.Msg) (bool, string, tea.Cmd) {
	if d.filePicker != nil {
		done, pickedPath, cmd := d.filePicker.Update(msg)
		if done {
			d.filePicker = nil
			if pickedPath != "" {
				d.handleFilePicked(pickedPath)
			}
		}
		return false, "", cmd
	}

	if d.isAdding || d.isEditing {
		return d.handleCustomFormKeys(msg)
	}
	if d.isDeleting {
		return d.handleDeleteConfirmKeys(msg)
	}

	items := d.getItems()

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "esc", "?", "ctrl+k":
			return true, "", nil

		case "1", "2", "3", "4", "5", "6", "7":
			tabIdx := int(msg.String()[0] - '1')
			d.switchTab(quickcmd.OSTab(tabIdx))
			return false, "", nil

		case "tab", "right", "l":
			nextTab := (d.activeTab + 1) % quickcmd.OSTabCount
			d.switchTab(nextTab)
			return false, "", nil

		case "shift+tab", "left", "h":
			prevTab := d.activeTab - 1
			if prevTab < 0 {
				prevTab = quickcmd.OSTabCount - 1
			}
			d.switchTab(prevTab)
			return false, "", nil

		case "up", "k":
			if len(items) > 0 {
				d.selectedIndex--
				if d.selectedIndex < 0 {
					d.selectedIndex = len(items) - 1
				}
			}
			return false, "", nil

		case "down", "j":
			if len(items) > 0 {
				d.selectedIndex++
				if d.selectedIndex >= len(items) {
					d.selectedIndex = 0
				}
			}
			return false, "", nil

		case "pgup", "pageup", "ctrl+u", "ctrl+b":
			if len(items) > 0 {
				d.selectedIndex -= 5
				if d.selectedIndex < 0 {
					d.selectedIndex = 0
				}
			}
			return false, "", nil

		case "pgdown", "pagedown", "ctrl+d":
			if len(items) > 0 {
				d.selectedIndex += 5
				if d.selectedIndex >= len(items) {
					d.selectedIndex = len(items) - 1
				}
			}
			return false, "", nil

		case "a", "A":
			if d.activeTab == quickcmd.OSTabCustom {
				d.openAddCustomForm()
				return false, "", nil
			}

		case "e", "E":
			if d.activeTab == quickcmd.OSTabCustom && len(items) > 0 {
				d.openEditCustomForm()
				return false, "", nil
			}

		case "d", "D", "delete":
			if d.activeTab == quickcmd.OSTabCustom && len(items) > 0 {
				d.openDeleteCustomConfirm()
				return false, "", nil
			}

		case "x", "X":
			if d.activeTab == quickcmd.OSTabCustom && len(d.customCmds) > 0 {
				d.handleDirectExport()
				return false, "", nil
			}

		case "i", "I":
			if d.activeTab == quickcmd.OSTabCustom {
				d.openImportPicker(80, 24)
				return false, "", nil
			}

		case "enter":
			if d.activeTab == quickcmd.OSTabShortcuts {
				return true, "", nil
			}
			if len(items) > 0 && d.selectedIndex >= 0 && d.selectedIndex < len(items) {
				return true, items[d.selectedIndex].Command, nil
			}
			return true, "", nil
		}
	}

	return false, "", nil
}

func (d *RunbookDrawer) switchTab(tab quickcmd.OSTab) {
	d.activeTab = tab
	d.selectedIndex = 0
	d.bannerMessage = ""
	if tab == quickcmd.OSTabCustom {
		d.loadCustomCommands()
	}
}

