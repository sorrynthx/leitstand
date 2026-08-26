package tui

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type HostFormData struct {
	Name     string
	Address  string
	Port     int
	Username string
	Password string
	Group    string
}

type HostForm struct {
	inputs     []textinput.Model
	focusIndex int
	width      int
	height     int
	errMessage string
}

func NewHostForm() *HostForm {
	inputs := make([]textinput.Model, 6)

	// 0: Host Name
	inputs[0] = textinput.New()
	inputs[0].Placeholder = "e.g. office-test-01"
	inputs[0].Focus()
	inputs[0].Prompt = "Host Name:  "
	inputs[0].Width = 30

	// 1: IP / Address
	inputs[1] = textinput.New()
	inputs[1].Placeholder = "e.g. 192.168.0.50"
	inputs[1].Prompt = "IP/Address: "
	inputs[1].Width = 30

	// 2: Port
	inputs[2] = textinput.New()
	inputs[2].Placeholder = "22"
	inputs[2].SetValue("22")
	inputs[2].Prompt = "SSH Port:   "
	inputs[2].Width = 10

	// 3: Username
	inputs[3] = textinput.New()
	inputs[3].Placeholder = "ubuntu or root"
	inputs[3].Prompt = "Username:   "
	inputs[3].Width = 30

	// 4: Password
	inputs[4] = textinput.New()
	inputs[4].Placeholder = "SSH Password"
	inputs[4].EchoMode = textinput.EchoPassword
	inputs[4].EchoCharacter = '•'
	inputs[4].Prompt = "Password:   "
	inputs[4].Width = 30

	// 5: Group
	inputs[5] = textinput.New()
	inputs[5].Placeholder = "e.g. Office Lab"
	inputs[5].SetValue("General")
	inputs[5].Prompt = "Group:      "
	inputs[5].Width = 30

	return &HostForm{
		inputs:     inputs,
		focusIndex: 0,
	}
}

func (f *HostForm) Update(msg tea.Msg) (bool, *HostFormData, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "esc":
			return true, nil, nil // Close modal without saving

		case "tab", "down", "enter":
			if msg.String() == "enter" && f.focusIndex == len(f.inputs)-1 {
				// Submit form
				data, err := f.validateAndExtract()
				if err != nil {
					f.errMessage = err.Error()
					return false, nil, nil
				}
				return true, data, nil
			}

			// Move to next field
			f.inputs[f.focusIndex].Blur()
			f.focusIndex = (f.focusIndex + 1) % len(f.inputs)
			f.inputs[f.focusIndex].Focus()
			return false, nil, textinput.Blink

		case "shift+tab", "up":
			f.inputs[f.focusIndex].Blur()
			f.focusIndex--
			if f.focusIndex < 0 {
				f.focusIndex = len(f.inputs) - 1
			}
			f.inputs[f.focusIndex].Focus()
			return false, nil, textinput.Blink

		case "ctrl+s":
			data, err := f.validateAndExtract()
			if err != nil {
				f.errMessage = err.Error()
				return false, nil, nil
			}
			return true, data, nil
		}
	}

	// Update active input model
	cmds := make([]tea.Cmd, len(f.inputs))
	for i := range f.inputs {
		f.inputs[i], cmds[i] = f.inputs[i].Update(msg)
	}

	return false, nil, tea.Batch(cmds...)
}

func (f *HostForm) validateAndExtract() (*HostFormData, error) {
	name := strings.TrimSpace(f.inputs[0].Value())
	if name == "" {
		return nil, fmt.Errorf("host name is required")
	}

	addr := strings.TrimSpace(f.inputs[1].Value())
	if addr == "" {
		return nil, fmt.Errorf("address/IP is required")
	}

	portStr := strings.TrimSpace(f.inputs[2].Value())
	port := 22
	if portStr != "" {
		if p, err := strconv.Atoi(portStr); err == nil && p > 0 {
			port = p
		}
	}

	user := strings.TrimSpace(f.inputs[3].Value())
	if user == "" {
		return nil, fmt.Errorf("username is required")
	}

	pass := f.inputs[4].Value()
	if pass == "" {
		return nil, fmt.Errorf("password is required")
	}

	group := strings.TrimSpace(f.inputs[5].Value())
	if group == "" {
		group = "General"
	}

	return &HostFormData{
		Name:     name,
		Address:  addr,
		Port:     port,
		Username: user,
		Password: pass,
		Group:    group,
	}, nil
}

func (f *HostForm) View(termWidth, termHeight int) string {
	var b strings.Builder

	title := lipgloss.NewStyle().Bold(true).Foreground(ColorPrimary).Render("➕ REGISTER TARGET SERVER (SSH)")
	b.WriteString(title + "\n")
	b.WriteString(lipgloss.NewStyle().Foreground(ColorMuted).Render("MobaXterm-style instant connection & monitoring") + "\n\n")

	if f.errMessage != "" {
		b.WriteString(lipgloss.NewStyle().Foreground(ColorDanger).Render("⚠️ "+f.errMessage) + "\n\n")
	}

	for i := range f.inputs {
		b.WriteString(f.inputs[i].View() + "\n")
	}

	b.WriteString("\n")
	hints := lipgloss.NewStyle().Foreground(ColorMuted).Render("[Tab/↓] Next  [Shift+Tab/↑] Prev  [Enter/Ctrl+S] Save & Connect  [Esc] Cancel")
	b.WriteString(hints)

	boxStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(ColorPrimary).
		Padding(1, 2).
		Width(65)

	content := boxStyle.Render(b.String())

	return lipgloss.Place(termWidth, termHeight, lipgloss.Center, lipgloss.Center, content)
}
