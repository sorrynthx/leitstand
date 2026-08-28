package tui

import (
	"errors"
	"fmt"
	"leitstand/internal/i18n"
	"leitstand/internal/storage"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"golang.org/x/crypto/ssh"
)

type AuthType int

const (
	AuthTypePassword AuthType = 0
	AuthTypeKey      AuthType = 1
)

type HostFormData struct {
	Name       string
	Address    string
	Port       int
	Username   string
	AuthMethod string // "password" or "private_key"
	Password   string
	KeyPath    string
	KeyContent string
	Passphrase string
	Group      string
}

type HostForm struct {
	isEditMode   bool
	hostID       int64
	existingKey  string
	existingPass string
	authType     AuthType
	inputs       []textinput.Model
	focusIndex   int
	width        int
	height       int
	errMessage   string
	filePicker   *FilePickerModal
}

func NewHostForm() *HostForm {
	// Total 8 possible text input fields
	// 0: Name, 1: Address, 2: Port, 3: Username, 4: Password, 5: KeyPath, 6: KeyPassphrase, 7: Group
	inputs := make([]textinput.Model, 8)

	// 0: Host Name
	inputs[0] = textinput.New()
	inputs[0].Placeholder = "e.g. prod-api-01"
	inputs[0].Focus()
	inputs[0].Prompt = "Host Name:       "
	inputs[0].Width = 38

	// 1: IP / Address
	inputs[1] = textinput.New()
	inputs[1].Placeholder = "e.g. 192.168.0.50 or host.example.com"
	inputs[1].Prompt = "IP / Address:    "
	inputs[1].Width = 38

	// 2: Port
	inputs[2] = textinput.New()
	inputs[2].Placeholder = "22"
	inputs[2].SetValue("22")
	inputs[2].Prompt = "SSH Port:        "
	inputs[2].Width = 12

	// 3: Username
	inputs[3] = textinput.New()
	inputs[3].Placeholder = "ubuntu or root"
	inputs[3].Prompt = "SSH Username:    "
	inputs[3].Width = 38

	// 4: Password
	inputs[4] = textinput.New()
	inputs[4].Placeholder = "SSH Password"
	inputs[4].EchoMode = textinput.EchoPassword
	inputs[4].EchoCharacter = '•'
	inputs[4].Prompt = "Password:        "
	inputs[4].Width = 38

	// 5: Key Path
	inputs[5] = textinput.New()
	inputs[5].Placeholder = "Press [b] to browse or enter path"
	inputs[5].Prompt = "Key File Path:   "
	inputs[5].Width = 38

	// 6: Key Passphrase (optional)
	inputs[6] = textinput.New()
	inputs[6].Placeholder = "Leave empty if unencrypted"
	inputs[6].EchoMode = textinput.EchoPassword
	inputs[6].EchoCharacter = '•'
	inputs[6].Prompt = "Key Passphrase:  "
	inputs[6].Width = 38

	// 7: Group
	inputs[7] = textinput.New()
	inputs[7].Placeholder = "e.g. Production Lab"
	inputs[7].SetValue("General")
	inputs[7].Prompt = "Server Group:    "
	inputs[7].Width = 38

	return &HostForm{
		isEditMode: false,
		authType:   AuthTypePassword,
		inputs:     inputs,
		focusIndex: 0,
	}
}

// NewEditHostForm creates a pre-populated form for editing an existing server.
func NewEditHostForm(h *storage.Host, secret *storage.HostSecret, payload *storage.SecretPayload) *HostForm {
	f := NewHostForm()
	f.isEditMode = true
	f.hostID = h.ID

	f.inputs[0].SetValue(h.Name)
	f.inputs[1].SetValue(h.Address)
	f.inputs[2].SetValue(strconv.Itoa(h.Port))
	f.inputs[3].SetValue(h.Username)
	f.inputs[7].SetValue(h.GroupName)

	if secret != nil && secret.AuthMethod == "private_key" {
		f.authType = AuthTypeKey
		if payload != nil {
			f.existingKey = payload.PrivateKey
			f.existingPass = payload.Passphrase
			f.inputs[5].SetValue("(Vault Key Loaded)")
			f.inputs[6].SetValue(payload.Passphrase)
		}
	} else {
		f.authType = AuthTypePassword
		if payload != nil && payload.Password != "" {
			f.inputs[4].SetValue(payload.Password)
		}
	}

	return f
}

// getFieldOrder returns the virtual navigation indices for the current authType.
// Virtual index -1 is the Auth Method Selector.
func (f *HostForm) getFieldOrder() []int {
	if f.authType == AuthTypePassword {
		return []int{0, 1, 2, 3, -1, 4, 7}
	}
	return []int{0, 1, 2, 3, -1, 5, 6, 7}
}

func (f *HostForm) Update(msg tea.Msg) (bool, *HostFormData, tea.Cmd) {
	// If FilePicker sub-modal is open, route messages to it
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
			return true, nil, nil // Close modal without saving

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

		case "left", "right":
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
			// Open File Picker if in Key Path field or Auth selector
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
				// Submit form
				data, err := f.validateAndExtract()
				if err != nil {
					f.errMessage = err.Error()
					return false, nil, nil
				}
				return true, data, nil
			}

			// Open file picker if enter pressed on empty Key Path field
			if msg.String() == "enter" && f.authType == AuthTypeKey && order[f.focusIndex] == 5 && strings.TrimSpace(f.inputs[5].Value()) == "" {
				f.filePicker = NewFilePickerModal("", f.width, f.height)
				return false, nil, nil
			}

			// Move to next field
			currField := order[f.focusIndex]
			if currField >= 0 && currField < len(f.inputs) {
				f.inputs[currField].Blur()
			}

			f.focusIndex = (f.focusIndex + 1) % len(order)

			nextField := order[f.focusIndex]
			if nextField >= 0 && nextField < len(f.inputs) {
				f.inputs[nextField].Focus()
			}
			return false, nil, textinput.Blink

		case "shift+tab", "up":
			currField := order[f.focusIndex]
			if currField >= 0 && currField < len(f.inputs) {
				f.inputs[currField].Blur()
			}

			f.focusIndex--
			if f.focusIndex < 0 {
				f.focusIndex = len(order) - 1
			}

			prevField := order[f.focusIndex]
			if prevField >= 0 && prevField < len(f.inputs) {
				f.inputs[prevField].Focus()
			}
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

	// Update active textinput
	currField := order[f.focusIndex]
	if currField >= 0 && currField < len(f.inputs) {
		var cmd tea.Cmd
		f.inputs[currField], cmd = f.inputs[currField].Update(msg)
		return false, nil, cmd
	}

	return false, nil, nil
}

func expandHomePath(p string) string {
	if strings.HasPrefix(p, "~/") || strings.HasPrefix(p, "~\\") || p == "~" {
		home, err := os.UserHomeDir()
		if err == nil {
			if p == "~" {
				return home
			}
			return filepath.Join(home, p[2:])
		}
	}
	return p
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

	group := strings.TrimSpace(f.inputs[7].Value())
	if group == "" {
		group = "General"
	}

	if f.authType == AuthTypePassword {
		pass := f.inputs[4].Value()
		if pass == "" {
			return nil, fmt.Errorf("password is required")
		}
		return &HostFormData{
			Name:       name,
			Address:    addr,
			Port:       port,
			Username:   user,
			AuthMethod: "password",
			Password:   pass,
			Group:      group,
		}, nil
	}

	// AuthTypeKey
	keyPathInput := strings.TrimSpace(f.inputs[5].Value())
	var keyBytes []byte
	var keyContent string
	passphrase := f.inputs[6].Value()

	if f.isEditMode && f.existingKey != "" && (keyPathInput == "(Vault Key Loaded)" || keyPathInput == "") {
		// Use existing key from vault
		keyBytes = []byte(f.existingKey)
		keyContent = f.existingKey
		if passphrase == "" {
			passphrase = f.existingPass
		}
	} else {
		if keyPathInput == "" {
			return nil, fmt.Errorf("private key file path is required")
		}

		expandedPath := expandHomePath(keyPathInput)
		var err error
		keyBytes, err = os.ReadFile(expandedPath)
		if err != nil {
			return nil, fmt.Errorf("failed to read private key file: %w", err)
		}
		keyContent = string(keyBytes)
	}

	// Validate private key structure immediately
	var err error
	if passphrase != "" {
		_, err = ssh.ParsePrivateKeyWithPassphrase(keyBytes, []byte(passphrase))
	} else {
		_, err = ssh.ParsePrivateKey(keyBytes)
	}

	if err != nil {
		var missingPassErr *ssh.PassphraseMissingError
		if errors.As(err, &missingPassErr) {
			return nil, fmt.Errorf("passphrase is required for this encrypted private key")
		}
		return nil, fmt.Errorf("invalid private key file: %w", err)
	}

	return &HostFormData{
		Name:       name,
		Address:    addr,
		Port:       port,
		Username:   user,
		AuthMethod: "private_key",
		KeyPath:    keyPathInput,
		KeyContent: keyContent,
		Passphrase: passphrase,
		Group:      group,
	}, nil
}

func (f *HostForm) View(termWidth, termHeight int) string {
	// If file picker is active, render file picker instead
	if f.filePicker != nil {
		return f.filePicker.View(termWidth, termHeight)
	}

	var b strings.Builder

	titleText := "➕ REGISTER TARGET SERVER (SSH)"
	if f.isEditMode {
		titleText = "✏️ EDIT TARGET SERVER (SSH)"
	}
	title := lipgloss.NewStyle().Bold(true).Foreground(ColorPrimary).Render(titleText)
	b.WriteString(title + "\n\n")

	if IsCapsLockOn() {
		capsBadge := lipgloss.NewStyle().
			Bold(true).
			Background(ColorWarning).
			Foreground(lipgloss.Color("#000000")).
			Padding(0, 1).
			Render(i18n.T("badge_caps_lock"))
		b.WriteString(capsBadge + "\n\n")
	}

	// Check password / passphrase for non-ASCII
	hasNonASCII := false
	if f.authType == AuthTypePassword {
		for _, r := range f.inputs[4].Value() {
			if r > 127 {
				hasNonASCII = true
				break
			}
		}
	} else {
		for _, r := range f.inputs[6].Value() {
			if r > 127 {
				hasNonASCII = true
				break
			}
		}
	}
	if hasNonASCII {
		warnBox := lipgloss.NewStyle().
			Foreground(ColorDanger).
			Bold(true).
			Render(i18n.T("warn_non_ascii"))
		b.WriteString(warnBox + "\n\n")
	}

	if f.errMessage != "" {
		b.WriteString(lipgloss.NewStyle().Foreground(ColorDanger).Bold(true).Render("⚠️ "+f.errMessage) + "\n\n")
	}

	order := f.getFieldOrder()

	// Section 1: Target Host Connection Info
	b.WriteString(f.inputs[0].View() + "\n")
	b.WriteString(f.inputs[1].View() + "\n")
	b.WriteString(f.inputs[2].View() + "\n")
	b.WriteString(f.inputs[3].View() + "\n\n")

	// Section 2: Auth Method Selector (index -1)
	isAuthFocus := (order[f.focusIndex] == -1)
	authLabelStyle := lipgloss.NewStyle().Foreground(ColorSecondary)
	if isAuthFocus {
		authLabelStyle = authLabelStyle.Bold(true).Foreground(ColorPrimary)
	}

	var pwdBtn, keyBtn string
	if f.authType == AuthTypePassword {
		pwdBtn = lipgloss.NewStyle().Bold(true).Background(ColorPrimary).Foreground(lipgloss.Color("#FFFFFF")).Padding(0, 1).Render("[1] 🔒 Password")
		keyBtn = lipgloss.NewStyle().Foreground(ColorMuted).Padding(0, 1).Render("[2] 🔑 Private Key")
	} else {
		pwdBtn = lipgloss.NewStyle().Foreground(ColorMuted).Padding(0, 1).Render("[1] 🔒 Password")
		keyBtn = lipgloss.NewStyle().Bold(true).Background(ColorPrimary).Foreground(lipgloss.Color("#FFFFFF")).Padding(0, 1).Render("[2] 🔑 Private Key")
	}

	focusArrow := "  "
	if isAuthFocus {
		focusArrow = lipgloss.NewStyle().Bold(true).Foreground(ColorPrimary).Render("❯ ")
	}
	b.WriteString(fmt.Sprintf("%s%s %s   %s\n\n", focusArrow, authLabelStyle.Render("Auth Method:     "), pwdBtn, keyBtn))

	// Section 3: Credentials Input
	if f.authType == AuthTypePassword {
		b.WriteString(f.inputs[4].View() + "\n\n")
	} else {
		// Key Path with Browse Hint
		browseBadge := lipgloss.NewStyle().
			Bold(true).
			Background(ColorSecondary).
			Foreground(lipgloss.Color("#FFFFFF")).
			Padding(0, 1).
			Render("[b] 📂 Browse File")
		b.WriteString(f.inputs[5].View() + "  " + browseBadge + "\n")
		b.WriteString(f.inputs[6].View() + "\n\n")
	}

	// Section 4: Group
	b.WriteString(f.inputs[7].View() + "\n\n")

	// Divider line
	divider := lipgloss.NewStyle().Foreground(lipgloss.Color("#334155")).Render(strings.Repeat("─", 70))
	b.WriteString(divider + "\n")

	var hints string
	if f.authType == AuthTypeKey {
		hints = lipgloss.NewStyle().Foreground(ColorMuted).Render("[Tab/↓] Next   [Shift+Tab/↑] Prev   [1/2] Auth   [b] 📂 Browse File\n[Enter/Ctrl+S] Save & Connect   [Esc] Cancel")
	} else {
		hints = lipgloss.NewStyle().Foreground(ColorMuted).Render("[Tab/↓] Next   [Shift+Tab/↑] Prev   [1/2] Auth   [Enter/Ctrl+S] Save & Connect   [Esc] Cancel")
	}
	b.WriteString(hints)

	boxStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(ColorPrimary).
		Padding(1, 4).
		Width(80)

	content := boxStyle.Render(b.String())

	return lipgloss.Place(termWidth, termHeight, lipgloss.Center, lipgloss.Center, content)
}
