package tui

import (
	"fmt"
	"leitstand/internal/i18n"
	"os"
	"path/filepath"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

func (fp *FilePickerModal) Update(msg tea.Msg) (bool, string, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "esc":
			return true, "", nil

		case "up", "k":
			if fp.selectedIndex > 0 {
				fp.selectedIndex--
			}
			return false, "", nil

		case "down", "j":
			if fp.selectedIndex < len(fp.items)-1 {
				fp.selectedIndex++
			}
			return false, "", nil

		case "pgup":
			fp.selectedIndex -= 10
			if fp.selectedIndex < 0 {
				fp.selectedIndex = 0
			}
			return false, "", nil

		case "pgdown":
			fp.selectedIndex += 10
			if fp.selectedIndex >= len(fp.items) {
				fp.selectedIndex = len(fp.items) - 1
			}
			if fp.selectedIndex < 0 {
				fp.selectedIndex = 0
			}
			return false, "", nil

		case "home":
			fp.selectedIndex = 0
			return false, "", nil

		case "end":
			if len(fp.items) > 0 {
				fp.selectedIndex = len(fp.items) - 1
			}
			return false, "", nil

		case "~":
			home, err := os.UserHomeDir()
			if err == nil {
				sshDir := filepath.Join(home, ".ssh")
				if _, err := os.Stat(sshDir); err == nil {
					fp.loadDirectory(sshDir)
				} else {
					fp.loadDirectory(home)
				}
			}
			return false, "", nil

		case "backspace":
			parent := filepath.Dir(fp.currentDir)
			if parent != fp.currentDir {
				fp.loadDirectory(parent)
			}
			return false, "", nil

		case " ":
			if fp.PickDir {
				return true, fp.currentDir, nil
			}
			return false, "", nil

		case "enter":
			if len(fp.items) == 0 || fp.selectedIndex >= len(fp.items) {
				return false, "", nil
			}

			selected := fp.items[fp.selectedIndex]
			if selected.IsDir {
				fp.loadDirectory(selected.Path)
				return false, "", nil
			}

			return true, selected.Path, nil
		}
	}
	return false, "", nil
}

func (fp *FilePickerModal) View(termWidth, termHeight int) string {
	var b strings.Builder

	title := i18n.T("picker_title_key")
	if fp.PickDir {
		title = i18n.T("picker_title_dir")
	}
	b.WriteString(lipgloss.NewStyle().Bold(true).Foreground(ColorPrimary).Render(title) + "\n")
	dirLabel := fmt.Sprintf(i18n.T("picker_current_dir"), fp.currentDir)
	b.WriteString(lipgloss.NewStyle().Foreground(ColorMuted).Render(dirLabel) + "\n\n")

	if fp.errMessage != "" {
		b.WriteString(lipgloss.NewStyle().Foreground(ColorDanger).Render("❌ "+fp.errMessage) + "\n\n")
	}

	maxVisible := 10
	startIdx := 0
	if fp.selectedIndex >= maxVisible {
		startIdx = fp.selectedIndex - maxVisible + 1
	}
	endIdx := startIdx + maxVisible
	if endIdx > len(fp.items) {
		endIdx = len(fp.items)
	}

	if len(fp.items) == 0 {
		b.WriteString(lipgloss.NewStyle().Foreground(ColorMuted).Render(i18n.T("picker_empty_dir") + "\n"))
	} else {
		for i := startIdx; i < endIdx; i++ {
			item := fp.items[i]
			isSelected := (i == fp.selectedIndex)

			var icon string
			var nameStyle lipgloss.Style

			if item.IsDir {
				icon = "📁 "
				nameStyle = lipgloss.NewStyle().Foreground(ColorSecondary).Bold(true)
			} else if item.IsKey {
				icon = "🔑 "
				nameStyle = lipgloss.NewStyle().Foreground(ColorSuccess).Bold(true)
			} else {
				icon = "📄 "
				nameStyle = lipgloss.NewStyle().Foreground(ColorMuted)
			}

			displayName := item.Name
			if isSelected {
				arrow := lipgloss.NewStyle().Bold(true).Foreground(ColorPrimary).Render("❯ ")
				line := lipgloss.NewStyle().
					Bold(true).
					Background(ColorPrimary).
					Foreground(ColorWhite).
					Padding(0, 1).
					Render(icon + displayName)
				b.WriteString(arrow + line + "\n")
			} else {
				b.WriteString("  " + icon + nameStyle.Render(displayName) + "\n")
			}
		}
	}

	b.WriteString("\n")
	hintsText := i18n.T("picker_hints_file")
	if fp.PickDir {
		hintsText = i18n.T("picker_hints_dir")
	}
	hints := lipgloss.NewStyle().Foreground(ColorMuted).Render(hintsText)
	b.WriteString(hints)

	return RenderModalContainer(b.String(), 72, ColorPrimary, termWidth, termHeight)
}
