package tui

import (
	"fmt"
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

	b.WriteString(lipgloss.NewStyle().Bold(true).Foreground(ColorPrimary).Render("🔑 Select SSH Private Key File") + "\n")
	b.WriteString(lipgloss.NewStyle().Foreground(ColorMuted).Render(fmt.Sprintf("Directory: %s", fp.currentDir)) + "\n\n")

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
		b.WriteString(lipgloss.NewStyle().Foreground(ColorMuted).Render("  (Directory is empty)\n"))
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
	hints := lipgloss.NewStyle().Foreground(ColorMuted).Render("[↑/↓] Navigate  [Enter] Select File / Enter Dir  [Backspace] Up  [~] ~/.ssh  [Esc] Cancel")
	b.WriteString(hints)

	return RenderModalContainer(b.String(), 72, ColorPrimary, termWidth, termHeight)
}
