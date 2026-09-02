package tui

import (
	"os"
	"path/filepath"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
)

func (f *HostForm) Update(msg tea.Msg) (bool, *HostFormData, tea.Cmd) {
	if f.filePicker != nil {
		done, pickedPath, cmd := f.filePicker.Update(msg)
		if done {
			if pickedPath != "" {
				f.inputs[5].SetValue(pickedPath)
				f.inputs[5].SetCursor(len(pickedPath))
			}
			f.filePicker = nil
		}
		return false, nil, cmd
	}

	order := f.getFieldOrder()

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "esc":
			return true, nil, nil

		case "f1", "alt+1":
			f.authType = AuthTypePassword
			f.errMessage = ""
			return false, nil, nil

		case "f2", "alt+2":
			f.authType = AuthTypeKey
			f.errMessage = ""
			return false, nil, nil

		case "1":
			if order[f.focusIndex] == -1 {
				f.authType = AuthTypePassword
				f.errMessage = ""
				return false, nil, nil
			}
		case "2":
			if order[f.focusIndex] == -1 {
				f.authType = AuthTypeKey
				f.errMessage = ""
				return false, nil, nil
			}

		case "left", "right", "space":
			if order[f.focusIndex] == -1 {
				if f.authType == AuthTypePassword {
					f.authType = AuthTypeKey
				} else {
					f.authType = AuthTypePassword
				}
				f.errMessage = ""
				return false, nil, nil
			}

		case "b", "B", "ctrl+f", "f", "F":
			if f.authType == AuthTypeKey && (order[f.focusIndex] == 5 || order[f.focusIndex] == -1) {
				currVal := strings.TrimSpace(f.inputs[5].Value())
				var initDir string
				if currVal != "" {
					expanded := expandHomePath(currVal)
					if fi, err := os.Stat(expanded); err == nil {
						if fi.IsDir() {
							initDir = expanded
						} else {
							initDir = filepath.Dir(expanded)
						}
					}
				}
				f.filePicker = NewFilePickerModal(initDir, f.width, f.height)
				return false, nil, nil
			}

		case "tab", "down", "enter":
			if msg.String() == "enter" && f.focusIndex == len(order)-1 {
				data, err := f.validateAndExtract()
				if err != nil {
					f.errMessage = err.Error()
					return false, nil, nil
				}
				return true, data, nil
			}

			if msg.String() == "enter" && f.authType == AuthTypeKey && order[f.focusIndex] == 5 && strings.TrimSpace(f.inputs[5].Value()) == "" {
				f.filePicker = NewFilePickerModal("", f.width, f.height)
				return false, nil, nil
			}

			currField := order[f.focusIndex]
			if currField >= 0 && currField < len(f.inputs) {
				f.inputs[currField].Blur()
			}

			f.focusIndex = (f.focusIndex + 1) % len(order)

			nextField := order[f.focusIndex]
			if nextField >= 0 && nextField < len(f.inputs) {
				f.inputs[nextField].Focus()
			}
			f.errMessage = ""
			return false, nil, nil

		case "shift+tab", "up":
			currField := order[f.focusIndex]
			if currField >= 0 && currField < len(f.inputs) {
				f.inputs[currField].Blur()
			}

			f.focusIndex--
			if f.focusIndex < 0 {
				f.focusIndex = len(order) - 1
			}

			nextField := order[f.focusIndex]
			if nextField >= 0 && nextField < len(f.inputs) {
				f.inputs[nextField].Focus()
			}
			f.errMessage = ""
			return false, nil, nil
		}
	}

	currField := order[f.focusIndex]
	if currField >= 0 && currField < len(f.inputs) {
		var cmd tea.Cmd
		f.inputs[currField], cmd = f.inputs[currField].Update(msg)
		return false, nil, cmd
	}

	return false, nil, nil
}
