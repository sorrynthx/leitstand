package tui

import (
	"fmt"
	"leitstand/internal/i18n"
	"strings"

	"github.com/charmbracelet/bubbles/textarea"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// EditorModal manages an in-app file editor overlay.
type EditorModal struct {
	HostID   int64
	HostName string
	FilePath string
	textarea textarea.Model
	isSaving bool
	err      error
}

// NewEditorModal creates an initialized editor modal for a specific file.
func NewEditorModal(hostID int64, hostName string, filePath string, initialContent string, width, height int) *EditorModal {
	ta := textarea.New()
	ta.Placeholder = "Empty file. Start typing..."
	ta.Focus()

	// Adjust dimensions for modal
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

// Update handles editor modal key events.
// Returns (done bool, saveRequested bool, updatedContent string, cmd tea.Cmd)
func (e *EditorModal) Update(msg tea.Msg) (bool, bool, string, tea.Cmd) {
	if keyMsg, ok := msg.(tea.KeyMsg); ok {
		switch keyMsg.String() {
		case "esc":
			return true, false, "", nil

		case "ctrl+s":
			e.isSaving = true
			return false, true, e.textarea.Value(), nil
		}
	}

	var cmd tea.Cmd
	e.textarea, cmd = e.textarea.Update(msg)
	return false, false, "", cmd
}

// SetError displays an error inside the modal.
func (e *EditorModal) SetError(err error) {
	e.err = err
	e.isSaving = false
}

// View renders the modal dialog.
func (e *EditorModal) View(screenWidth, screenHeight int) string {
	var b strings.Builder

	// Header
	header := lipgloss.NewStyle().Bold(true).Foreground(ColorPrimary).Render(fmt.Sprintf("%s ── %s (%s)", i18n.T("editor_title"), e.FilePath, e.HostName))
	b.WriteString(header + "\n\n")

	// Error banner if any
	if e.err != nil {
		errBox := lipgloss.NewStyle().Foreground(ColorDanger).Bold(true).Render(fmt.Sprintf("⚠️  Save Error: %v", e.err))
		b.WriteString(errBox + "\n")
	}

	// Textarea
	b.WriteString(e.textarea.View() + "\n\n")

	// Footer instructions
	saveHint := lipgloss.NewStyle().Bold(true).Foreground(ColorSuccess).Render(i18n.T("editor_save_hint")) + "  "
	exitHint := lipgloss.NewStyle().Bold(true).Foreground(ColorDanger).Render(i18n.T("editor_exit_hint")) + "  "
	lineHint := lipgloss.NewStyle().Foreground(ColorMuted).Render(fmt.Sprintf("Line: %d", e.textarea.Line()+1))

	b.WriteString(saveHint + exitHint + lineHint)

	boxWidth := screenWidth - 8
	if boxWidth < 45 {
		boxWidth = 45
	}
	boxHeight := screenHeight - 6
	if boxHeight < 10 {
		boxHeight = 10
	}

	modalBox := lipgloss.NewStyle().
		Border(lipgloss.DoubleBorder()).
		BorderForeground(ColorPrimary).
		Padding(1, 2).
		Width(boxWidth).
		Height(boxHeight)

	return lipgloss.Place(screenWidth, screenHeight, lipgloss.Center, lipgloss.Center, modalBox.Render(b.String()))
}
