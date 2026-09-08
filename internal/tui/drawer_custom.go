package tui

import (
	"fmt"
	"leitstand/internal/i18n"
	"leitstand/internal/quickcmd"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
)

func (d *RunbookDrawer) loadCustomCommands() {
	if d.store == nil {
		return
	}
	cmds, err := d.store.GetCustomCommands()
	if err == nil {
		d.customCmds = cmds
	}
}

func (d *RunbookDrawer) initCustomInputs() {
	d.customInputs = make([]textinput.Model, 4)
	placeholders := []string{
		"Docker live logs",
		"docker logs -f --tail 100",
		"Docker",
		"Follow recent container logs",
	}
	for i := range d.customInputs {
		t := textinput.New()
		t.Placeholder = placeholders[i]
		t.CharLimit = 256
		t.Width = 40
		d.customInputs[i] = t
	}
	d.customInputs[0].Focus()
	d.customFocused = 0
}

func (d *RunbookDrawer) getCustomItems() []quickcmd.CommandItem {
	var items []quickcmd.CommandItem
	for _, c := range d.customCmds {
		cat := c.Category
		if cat == "" {
			cat = "General"
		}
		items = append(items, quickcmd.CommandItem{
			ID:          fmt.Sprintf("%d", c.ID),
			CategoryKey: cat,
			TitleKey:    c.Title,
			DescKey:     c.Description,
			Command:     c.Command,
		})
	}
	return items
}

func (d *RunbookDrawer) openAddCustomForm() {
	d.initCustomInputs()
	d.isAdding = true
	d.isEditing = false
	d.isDeleting = false
	d.errMessage = ""
}

func (d *RunbookDrawer) openEditCustomForm() {
	if len(d.customCmds) == 0 || d.selectedIndex < 0 || d.selectedIndex >= len(d.customCmds) {
		return
	}
	cur := d.customCmds[d.selectedIndex]
	d.initCustomInputs()
	d.editingID = cur.ID
	d.customInputs[0].SetValue(cur.Title)
	d.customInputs[1].SetValue(cur.Command)
	d.customInputs[2].SetValue(cur.Category)
	d.customInputs[3].SetValue(cur.Description)
	d.isAdding = false
	d.isEditing = true
	d.isDeleting = false
	d.errMessage = ""
}

func (d *RunbookDrawer) openDeleteCustomConfirm() {
	if len(d.customCmds) == 0 || d.selectedIndex < 0 || d.selectedIndex >= len(d.customCmds) {
		return
	}
	d.isDeleting = true
	d.isAdding = false
	d.isEditing = false
	d.errMessage = ""
}

func (d *RunbookDrawer) handleCustomFormKeys(msg tea.Msg) (bool, string, tea.Cmd) {
	keyMsg, ok := msg.(tea.KeyMsg)
	if !ok {
		return false, "", nil
	}

	switch keyMsg.String() {
	case "esc":
		d.isAdding = false
		d.isEditing = false
		d.errMessage = ""
		return false, "", nil

	case "tab", "down":
		d.customFocused = (d.customFocused + 1) % len(d.customInputs)
		for i := range d.customInputs {
			if i == d.customFocused {
				d.customInputs[i].Focus()
			} else {
				d.customInputs[i].Blur()
			}
		}
		return false, "", nil

	case "shift+tab", "up":
		d.customFocused--
		if d.customFocused < 0 {
			d.customFocused = len(d.customInputs) - 1
		}
		for i := range d.customInputs {
			if i == d.customFocused {
				d.customInputs[i].Focus()
			} else {
				d.customInputs[i].Blur()
			}
		}
		return false, "", nil

	case "enter":
		title := strings.TrimSpace(d.customInputs[0].Value())
		command := strings.TrimSpace(d.customInputs[1].Value())
		category := strings.TrimSpace(d.customInputs[2].Value())
		desc := strings.TrimSpace(d.customInputs[3].Value())

		if title == "" || command == "" {
			d.errMessage = "⚠️ Title and Command cannot be empty"
			return false, "", nil
		}
		if category == "" {
			category = "General"
		}

		if d.isAdding {
			if d.store != nil {
				_, err := d.store.AddCustomCommand(title, command, category, desc)
				if err != nil {
					d.errMessage = fmt.Sprintf("Error: %v", err)
					return false, "", nil
				}
			}
			d.toastMessage = fmt.Sprintf(i18n.T("custom_cmd_saved_toast"), title)
		} else if d.isEditing {
			if d.store != nil {
				err := d.store.UpdateCustomCommand(d.editingID, title, command, category, desc)
				if err != nil {
					d.errMessage = fmt.Sprintf("Error: %v", err)
					return false, "", nil
				}
			}
			d.toastMessage = fmt.Sprintf(i18n.T("custom_cmd_updated_toast"), title)
		}

		d.isAdding = false
		d.isEditing = false
		d.loadCustomCommands()
		return false, "", nil
	}

	var cmd tea.Cmd
	d.customInputs[d.customFocused], cmd = d.customInputs[d.customFocused].Update(msg)
	return false, "", cmd
}

func (d *RunbookDrawer) handleDeleteConfirmKeys(msg tea.Msg) (bool, string, tea.Cmd) {
	keyMsg, ok := msg.(tea.KeyMsg)
	if !ok {
		return false, "", nil
	}

	switch keyMsg.String() {
	case "y", "Y", "enter":
		if len(d.customCmds) > 0 && d.selectedIndex >= 0 && d.selectedIndex < len(d.customCmds) {
			target := d.customCmds[d.selectedIndex]
			if d.store != nil {
				_ = d.store.DeleteCustomCommand(target.ID)
			}
			d.toastMessage = i18n.T("custom_cmd_deleted_toast")
			d.loadCustomCommands()
			if d.selectedIndex >= len(d.customCmds) && d.selectedIndex > 0 {
				d.selectedIndex = len(d.customCmds) - 1
			}
		}
		d.isDeleting = false
		return false, "", nil

	case "n", "N", "esc":
		d.isDeleting = false
		return false, "", nil
	}

	return false, "", nil
}
