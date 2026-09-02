package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/textarea"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type EditorModal struct {
	HostID    int64
	HostName  string
	FilePath  string
	textarea  textarea.Model
	isSaving  bool
	err       error
	StatusMsg string
}

func NewEditorModal(hostID int64, hostName string, filePath string, initialContent string, width, height int) *EditorModal {
	ta := textarea.New()
	ta.Placeholder = "Empty file. Start typing..."
	ta.Focus()

	modalWidth := width - 10
	if modalWidth < 40 {
		modalWidth = 40
	}
	modalHeight := height - 8
	if modalHeight < 8 {
		modalHeight = 8
	}

	ta.SetWidth(modalWidth - 4)
	ta.SetHeight(modalHeight - 4)
	ta.SetValue(initialContent)
	ta.ShowLineNumbers = true

	return &EditorModal{
		HostID:   hostID,
		HostName: hostName,
		FilePath: filePath,
		textarea: ta,
	}
}

func (em *EditorModal) Update(msg tea.Msg) (bool, bool, string, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "esc":
			return true, false, "", nil

		case "ctrl+s", "f2", "alt+s":
			em.isSaving = true
			return false, true, em.textarea.Value(), nil

		case "pgup", "pageup":
			h := em.textarea.Height()
			if h <= 0 {
				h = 10
			}
			for i := 0; i < h; i++ {
				em.textarea, _ = em.textarea.Update(tea.KeyMsg{Type: tea.KeyUp})
			}
			return false, false, "", nil

		case "pgdown", "pagedown":
			h := em.textarea.Height()
			if h <= 0 {
				h = 10
			}
			for i := 0; i < h; i++ {
				em.textarea, _ = em.textarea.Update(tea.KeyMsg{Type: tea.KeyDown})
			}
			return false, false, "", nil
		}
	}

	var cmd tea.Cmd
	em.textarea, cmd = em.textarea.Update(msg)
	return false, false, "", cmd
}

func (em *EditorModal) View(width, height int) string {
	var b strings.Builder

	title := fmt.Sprintf("✏️ Editing: %s (%s)", em.FilePath, em.HostName)
	b.WriteString(lipgloss.NewStyle().Bold(true).Foreground(ColorPrimary).Render(title) + "\n\n")

	b.WriteString(em.textarea.View() + "\n\n")

	if em.err != nil {
		b.WriteString(lipgloss.NewStyle().Foreground(ColorDanger).Render("❌ "+em.err.Error()) + "\n\n")
	} else if em.StatusMsg != "" {
		b.WriteString(lipgloss.NewStyle().Foreground(ColorSuccess).Render(em.StatusMsg) + "\n\n")
	}

	hints := lipgloss.NewStyle().Foreground(ColorMuted).Render("[F2 / Ctrl+S] Save Remote File    [PgUp / PgDn] Page Scroll    [Esc] Cancel / Close")
	b.WriteString(hints)

	boxStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(ColorPrimary).
		Padding(1, 2)

	content := boxStyle.Render(b.String())
	return lipgloss.Place(width, height, lipgloss.Center, lipgloss.Center, content)
}
