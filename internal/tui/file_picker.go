package tui

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type FileItem struct {
	Name  string
	IsDir bool
	IsKey bool
	Size  int64
	Path  string
}

type FilePickerModal struct {
	currentDir    string
	items         []FileItem
	selectedIndex int
	width         int
	height        int
	errMessage    string
}

func NewFilePickerModal(initialDir string, width, height int) *FilePickerModal {
	targetDir := initialDir
	if targetDir == "" {
		home, err := os.UserHomeDir()
		if err == nil {
			sshDir := filepath.Join(home, ".ssh")
			if fi, err := os.Stat(sshDir); err == nil && fi.IsDir() {
				targetDir = sshDir
			} else {
				targetDir = home
			}
		} else {
			targetDir = "."
		}
	}

	fp := &FilePickerModal{
		currentDir: targetDir,
		width:      width,
		height:     height,
	}
	fp.readDirectory()
	return fp
}

func isKnownKeyFile(name string) bool {
	lower := strings.ToLower(name)
	if strings.HasSuffix(lower, ".pub") {
		return false // Public keys are not private keys
	}
	if strings.HasSuffix(lower, ".pem") ||
		strings.HasSuffix(lower, ".key") ||
		strings.HasSuffix(lower, ".pkcs8") ||
		strings.HasSuffix(lower, ".id_rsa") ||
		strings.HasSuffix(lower, ".id_ed25519") ||
		strings.HasSuffix(lower, ".id_ecdsa") ||
		strings.HasSuffix(lower, ".id_dsa") {
		return true
	}
	if lower == "id_rsa" || lower == "id_ed25519" || lower == "id_ecdsa" || lower == "id_dsa" || lower == "test_key" {
		return true
	}
	return false
}

func (fp *FilePickerModal) readDirectory() {
	fp.items = nil
	fp.selectedIndex = 0
	fp.errMessage = ""

	entries, err := os.ReadDir(fp.currentDir)
	if err != nil {
		fp.errMessage = fmt.Sprintf("Failed to read directory: %v", err)
		return
	}

	var dirs []FileItem
	var files []FileItem

	// Add parent directory entry if not at root
	parent := filepath.Dir(fp.currentDir)
	if parent != fp.currentDir {
		dirs = append(dirs, FileItem{
			Name:  ".. (Parent Directory)",
			IsDir: true,
			Path:  parent,
		})
	}

	for _, e := range entries {
		name := e.Name()
		fullPath := filepath.Join(fp.currentDir, name)
		if e.IsDir() {
			dirs = append(dirs, FileItem{
				Name:  name,
				IsDir: true,
				Path:  fullPath,
			})
		} else {
			info, err := e.Info()
			var size int64
			if err == nil {
				size = info.Size()
			}
			isKey := isKnownKeyFile(name)
			files = append(files, FileItem{
				Name:  name,
				IsDir: false,
				IsKey: isKey,
				Size:  size,
				Path:  fullPath,
			})
		}
	}

	// Sort directories and files alphabetically (keys prioritized)
	sort.Slice(dirs, func(i, j int) bool {
		if dirs[i].Name == ".. (Parent Directory)" {
			return true
		}
		if dirs[j].Name == ".. (Parent Directory)" {
			return false
		}
		return strings.ToLower(dirs[i].Name) < strings.ToLower(dirs[j].Name)
	})

	sort.Slice(files, func(i, j int) bool {
		if files[i].IsKey != files[j].IsKey {
			return files[i].IsKey // Key files first
		}
		return strings.ToLower(files[i].Name) < strings.ToLower(files[j].Name)
	})

	fp.items = append(dirs, files...)
}

func (fp *FilePickerModal) Update(msg tea.Msg) (done bool, selectedPath string, cmd tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "esc":
			return true, "", nil // Cancel

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

		case "pageup":
			fp.selectedIndex -= 8
			if fp.selectedIndex < 0 {
				fp.selectedIndex = 0
			}
			return false, "", nil

		case "pagedown":
			fp.selectedIndex += 8
			if fp.selectedIndex >= len(fp.items) {
				fp.selectedIndex = len(fp.items) - 1
			}
			return false, "", nil

		case "backspace", "h":
			parent := filepath.Dir(fp.currentDir)
			if parent != fp.currentDir {
				fp.currentDir = parent
				fp.readDirectory()
			}
			return false, "", nil

		case "~":
			home, err := os.UserHomeDir()
			if err == nil {
				sshDir := filepath.Join(home, ".ssh")
				if fi, err := os.Stat(sshDir); err == nil && fi.IsDir() {
					fp.currentDir = sshDir
				} else {
					fp.currentDir = home
				}
				fp.readDirectory()
			}
			return false, "", nil

		case "enter", "l":
			if len(fp.items) == 0 {
				return false, "", nil
			}
			item := fp.items[fp.selectedIndex]
			if item.IsDir {
				fp.currentDir = item.Path
				fp.readDirectory()
				return false, "", nil
			}
			// File selected!
			return true, item.Path, nil
		}
	}
	return false, "", nil
}

func (fp *FilePickerModal) View(termWidth, termHeight int) string {
	var b strings.Builder

	title := lipgloss.NewStyle().Bold(true).Foreground(ColorPrimary).Render("📂 SELECT SSH PRIVATE KEY FILE")
	b.WriteString(title + "\n\n")

	// Current Path Bar
	pathBadge := lipgloss.NewStyle().
		Bold(true).
		Foreground(ColorSecondary).
		Render(fmt.Sprintf("Current Directory: %s", fp.currentDir))
	b.WriteString(pathBadge + "\n\n")

	if fp.errMessage != "" {
		b.WriteString(lipgloss.NewStyle().Foreground(ColorDanger).Bold(true).Render("⚠️ "+fp.errMessage) + "\n\n")
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
		b.WriteString(lipgloss.NewStyle().Foreground(ColorMuted).Render("  (Directory is empty)") + "\n")
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
					Foreground(lipgloss.Color("#FFFFFF")).
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

	boxStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(ColorPrimary).
		Padding(1, 3).
		Width(72)

	content := boxStyle.Render(b.String())

	return lipgloss.Place(termWidth, termHeight, lipgloss.Center, lipgloss.Center, content)
}
