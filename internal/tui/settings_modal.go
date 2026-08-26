package tui

import (
	"fmt"
	"leitstand/internal/i18n"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type SettingsTab int

const (
	TabGeneral SettingsTab = iota
	TabAbout
)

type SettingsField int

const (
	FieldLanguage SettingsField = iota
	FieldInterval
	FieldCurrentPass
	FieldNewPass
	FieldConfirmPass
	FieldSubmitBtn
	FieldCount
)

// SettingsModal provides user configuration for language, interval, vault password, and creator profile.
type SettingsModal struct {
	activeTab         SettingsTab
	selectedLangIndex int
	intervalIndex     int
	inputs            []textinput.Model
	focusField        SettingsField
	errMessage        string
	successMessage    string
}

var intervalOptions = []struct {
	Label    string
	Duration time.Duration
}{
	{Label: "5 Seconds (Recommended)", Duration: 5 * time.Second},
	{Label: "10 Seconds (Relaxed)", Duration: 10 * time.Second},
	{Label: "30 Seconds (Minimal)", Duration: 30 * time.Second},
	{Label: "60 Seconds (Slow / Eco)", Duration: 60 * time.Second},
	{Label: "Off (Disabled / Full Console)", Duration: 0},
}

// NewSettingsModal initializes settings modal with current configuration.
func NewSettingsModal(currLang i18n.Lang, currInterval time.Duration) *SettingsModal {
	langIdx := 0
	for i, opt := range i18n.SupportedLangs {
		if opt.Code == currLang {
			langIdx = i
			break
		}
	}

	intervalIdx := 0 // 5s default
	for i, opt := range intervalOptions {
		if opt.Duration == currInterval {
			intervalIdx = i
			break
		}
	}

	inputs := make([]textinput.Model, 3)
	// Current Password
	inputs[0] = textinput.New()
	inputs[0].Placeholder = "Required only if changing password"
	inputs[0].EchoMode = textinput.EchoPassword
	inputs[0].EchoCharacter = '•'
	inputs[0].Prompt = "Current Password: "
	inputs[0].Width = 25

	// New Password
	inputs[1] = textinput.New()
	inputs[1].Placeholder = "Enter new master password"
	inputs[1].EchoMode = textinput.EchoPassword
	inputs[1].EchoCharacter = '•'
	inputs[1].Prompt = "New Password:     "
	inputs[1].Width = 25

	// Confirm New Password
	inputs[2] = textinput.New()
	inputs[2].Placeholder = "Confirm new master password"
	inputs[2].EchoMode = textinput.EchoPassword
	inputs[2].EchoCharacter = '•'
	inputs[2].Prompt = "Confirm Password: "
	inputs[2].Width = 25

	return &SettingsModal{
		activeTab:         TabGeneral,
		selectedLangIndex: langIdx,
		intervalIndex:     intervalIdx,
		inputs:            inputs,
		focusField:        FieldLanguage,
	}
}

// Update handles key navigation inside settings modal.
// Returns (done bool, saveRequested bool, lang i18n.Lang, interval time.Duration, currPass, newPass string, cmd tea.Cmd)
func (s *SettingsModal) Update(msg tea.Msg) (bool, bool, i18n.Lang, time.Duration, string, string, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "esc":
			return true, false, "", 0, "", "", nil

		case "1":
			s.activeTab = TabGeneral
			return false, false, "", 0, "", "", nil

		case "2":
			s.activeTab = TabAbout
			return false, false, "", 0, "", "", nil

		case "tab":
			if s.activeTab == TabAbout {
				s.activeTab = TabGeneral
				return false, false, "", 0, "", "", nil
			}
			s.blurCurrent()
			s.focusField = (s.focusField + 1) % FieldCount
			s.focusCurrent()
			return false, false, "", 0, "", "", textinput.Blink

		case "shift+tab":
			if s.activeTab == TabAbout {
				s.activeTab = TabGeneral
				return false, false, "", 0, "", "", nil
			}
			s.blurCurrent()
			s.focusField--
			if s.focusField < 0 {
				s.focusField = FieldCount - 1
			}
			s.focusCurrent()
			return false, false, "", 0, "", "", textinput.Blink

		case "down":
			if s.activeTab == TabGeneral {
				s.blurCurrent()
				s.focusField = (s.focusField + 1) % FieldCount
				s.focusCurrent()
				return false, false, "", 0, "", "", textinput.Blink
			}

		case "up":
			if s.activeTab == TabGeneral {
				s.blurCurrent()
				s.focusField--
				if s.focusField < 0 {
					s.focusField = FieldCount - 1
				}
				s.focusCurrent()
				return false, false, "", 0, "", "", textinput.Blink
			}

		case "left", "h":
			if s.activeTab == TabGeneral {
				if s.focusField == FieldLanguage {
					s.selectedLangIndex--
					if s.selectedLangIndex < 0 {
						s.selectedLangIndex = len(i18n.SupportedLangs) - 1
					}
					i18n.SetLang(i18n.SupportedLangs[s.selectedLangIndex].Code)
					return false, false, "", 0, "", "", nil
				} else if s.focusField == FieldInterval {
					s.intervalIndex--
					if s.intervalIndex < 0 {
						s.intervalIndex = len(intervalOptions) - 1
					}
					return false, false, "", 0, "", "", nil
				}
			}

		case "right", "l":
			if s.activeTab == TabGeneral {
				if s.focusField == FieldLanguage {
					s.selectedLangIndex = (s.selectedLangIndex + 1) % len(i18n.SupportedLangs)
					i18n.SetLang(i18n.SupportedLangs[s.selectedLangIndex].Code)
					return false, false, "", 0, "", "", nil
				} else if s.focusField == FieldInterval {
					s.intervalIndex = (s.intervalIndex + 1) % len(intervalOptions)
					return false, false, "", 0, "", "", nil
				}
			}

		case "enter":
			if s.activeTab == TabAbout {
				s.activeTab = TabGeneral
				return false, false, "", 0, "", "", nil
			}

			selectedLang := i18n.SupportedLangs[s.selectedLangIndex].Code
			selectedInterval := intervalOptions[s.intervalIndex].Duration

			currPass := s.inputs[0].Value()
			newPass := s.inputs[1].Value()
			confirmPass := s.inputs[2].Value()

			if newPass != "" {
				if len(newPass) < 4 {
					s.errMessage = "New password must be at least 4 characters"
					return false, false, "", 0, "", "", nil
				}
				if newPass != confirmPass {
					s.errMessage = i18n.T("settings_pass_mismatch")
					return false, false, "", 0, "", "", nil
				}
				if currPass == "" {
					s.errMessage = "Current password is required to set new password"
					return false, false, "", 0, "", "", nil
				}
			}

			return true, true, selectedLang, selectedInterval, currPass, newPass, nil
		}
	}

	var cmd tea.Cmd
	if s.activeTab == TabGeneral {
		inputIdx := s.inputIndexForField(s.focusField)
		if inputIdx >= 0 && inputIdx < len(s.inputs) {
			s.inputs[inputIdx], cmd = s.inputs[inputIdx].Update(msg)
		}
	}
	return false, false, "", 0, "", "", cmd
}

func (s *SettingsModal) blurCurrent() {
	idx := s.inputIndexForField(s.focusField)
	if idx >= 0 && idx < len(s.inputs) {
		s.inputs[idx].Blur()
	}
}

func (s *SettingsModal) focusCurrent() {
	idx := s.inputIndexForField(s.focusField)
	if idx >= 0 && idx < len(s.inputs) {
		s.inputs[idx].Focus()
	}
}

func (s *SettingsModal) inputIndexForField(f SettingsField) int {
	switch f {
	case FieldCurrentPass:
		return 0
	case FieldNewPass:
		return 1
	case FieldConfirmPass:
		return 2
	default:
		return -1
	}
}

func (s *SettingsModal) SetError(err error) {
	s.errMessage = fmt.Sprintf("❌ %v", err)
	s.successMessage = ""
}

// View renders the settings dialog in full screen size.
func (s *SettingsModal) View(screenWidth, screenHeight int) string {
	var b strings.Builder

	// Header Title
	title := lipgloss.NewStyle().Bold(true).Foreground(ColorPrimary).Render(i18n.T("settings_title"))
	b.WriteString(title + "\n\n")

	// Enhanced Tab Bar (Clearly shows [1] and [2] keys)
	var tabGenStyle, tabAboutStyle lipgloss.Style
	if s.activeTab == TabGeneral {
		tabGenStyle = lipgloss.NewStyle().Bold(true).Background(ColorPrimary).Foreground(lipgloss.Color("#000000")).Padding(0, 2)
		tabAboutStyle = lipgloss.NewStyle().Bold(true).Background(ColorBorder).Foreground(lipgloss.Color("#B0BEC5")).Padding(0, 2)
	} else {
		tabGenStyle = lipgloss.NewStyle().Bold(true).Background(ColorBorder).Foreground(lipgloss.Color("#B0BEC5")).Padding(0, 2)
		tabAboutStyle = lipgloss.NewStyle().Bold(true).Background(ColorPrimary).Foreground(lipgloss.Color("#000000")).Padding(0, 2)
	}

	tabBar := lipgloss.JoinHorizontal(lipgloss.Top,
		tabGenStyle.Render(i18n.T("tab_general")),
		"   ",
		tabAboutStyle.Render(i18n.T("tab_about")),
	)
	b.WriteString(tabBar + "\n")

	// Localized Grey Sub-hint for switching tabs (shown on both tabs)
	tabHint := lipgloss.NewStyle().Foreground(ColorMuted).Render(i18n.T("settings_tab_hint"))
	b.WriteString(tabHint + "\n\n")

	if s.errMessage != "" {
		b.WriteString(lipgloss.NewStyle().Foreground(ColorDanger).Bold(true).Render(s.errMessage) + "\n\n")
	}

	contentWidth := screenWidth - 12
	if contentWidth < 50 {
		contentWidth = 50
	}

	if s.activeTab == TabGeneral {
		// 1. Language Row
		langLabel := i18n.T("settings_lang_label")
		langVal := fmt.Sprintf("◄  %s  ►", i18n.SupportedLangs[s.selectedLangIndex].Label)
		langStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#FFFFFF"))
		if s.focusField == FieldLanguage {
			langStyle = lipgloss.NewStyle().Bold(true).Background(ColorPrimary).Foreground(lipgloss.Color("#000000")).Padding(0, 1)
		}
		b.WriteString(fmt.Sprintf("%-26s %s\n\n", langLabel, langStyle.Render(langVal)))

		// 2. Telemetry Interval Row
		intervalLabel := i18n.T("settings_interval_label")
		intervalVal := fmt.Sprintf("◄  %s  ►", intervalOptions[s.intervalIndex].Label)
		intervalStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#FFFFFF"))
		if s.focusField == FieldInterval {
			intervalStyle = lipgloss.NewStyle().Bold(true).Background(ColorPrimary).Foreground(lipgloss.Color("#000000")).Padding(0, 1)
		}
		b.WriteString(fmt.Sprintf("%-26s %s\n\n", intervalLabel, intervalStyle.Render(intervalVal)))

		// Divider
		dividerLen := contentWidth
		if dividerLen > 80 {
			dividerLen = 80
		}
		b.WriteString(lipgloss.NewStyle().Foreground(ColorBorder).Render(strings.Repeat("─", dividerLen)) + "\n")
		b.WriteString(lipgloss.NewStyle().Bold(true).Foreground(ColorSecondary).Render("🔐 MASTER PASSWORD SECURITY") + "\n\n")

		if IsCapsLockOn() {
			capsBadge := lipgloss.NewStyle().
				Bold(true).
				Background(ColorWarning).
				Foreground(lipgloss.Color("#000000")).
				Padding(0, 1).
				Render(i18n.T("badge_caps_lock"))
			b.WriteString(capsBadge + "\n\n")
		}

		hasNonASCII := false
		for _, input := range s.inputs {
			for _, r := range input.Value() {
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

		// 3. Password inputs
		for _, input := range s.inputs {
			b.WriteString(input.View() + "\n\n")
		}

		// Actions Footer (General Tab)
		saveBtn := i18n.T("settings_btn_save")
		if s.focusField == FieldSubmitBtn {
			saveBtn = lipgloss.NewStyle().Bold(true).Background(ColorSuccess).Foreground(lipgloss.Color("#000000")).Padding(0, 1).Render(saveBtn)
		} else {
			saveBtn = lipgloss.NewStyle().Bold(true).Foreground(ColorSuccess).Render(saveBtn)
		}

		cancelBtn := lipgloss.NewStyle().Bold(true).Foreground(ColorDanger).Render(i18n.T("settings_btn_cancel"))
		navHint := lipgloss.NewStyle().Foreground(ColorMuted).Render(i18n.T("settings_footer_gen"))
		b.WriteString(saveBtn + "    " + cancelBtn + "    " + navHint)

	} else {
		// ABOUT & CREATOR TAB (Spacious layout)
		appName := lipgloss.NewStyle().Bold(true).Foreground(ColorPrimary).Render(i18n.T("about_app_name"))
		tagline := lipgloss.NewStyle().Foreground(ColorMuted).Render(i18n.T("about_tagline"))
		b.WriteString(appName + "  ──  " + tagline + "\n\n")

		// Creator section
		cTitle := lipgloss.NewStyle().Bold(true).Foreground(ColorSecondary).Render(i18n.T("about_creator_title"))
		cName := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#FFFFFF")).Render(i18n.T("about_creator_name"))
		b.WriteString(cTitle + ":  " + cName + "\n\n")

		// Vision section
		vTitle := lipgloss.NewStyle().Bold(true).Foreground(ColorWarning).Render(i18n.T("about_vision_title"))
		vDescWidth := contentWidth - 4
		if vDescWidth > 80 {
			vDescWidth = 80
		}
		vDesc := lipgloss.NewStyle().Foreground(lipgloss.Color("#ECEFF1")).Width(vDescWidth).Render(i18n.T("about_vision_desc"))
		b.WriteString(vTitle + "\n" + vDesc + "\n\n")

		// Divider
		b.WriteString(lipgloss.NewStyle().Foreground(ColorBorder).Render(strings.Repeat("─", vDescWidth)) + "\n\n")

		// Links & Socials section
		lTitle := lipgloss.NewStyle().Bold(true).Foreground(ColorSuccess).Render(i18n.T("about_links_title"))
		b.WriteString(lTitle + "\n")
		b.WriteString("  💼 LinkedIn:  " + lipgloss.NewStyle().Foreground(ColorPrimary).Bold(true).Render("https://www.linkedin.com/in/kyunggon-kim") + "\n")
		b.WriteString("  🧵 Threads:   " + lipgloss.NewStyle().Foreground(ColorSecondary).Bold(true).Render("@kyunggon.dev") + "\n")
		b.WriteString("  🐙 GitHub:    " + lipgloss.NewStyle().Foreground(ColorWarning).Bold(true).Render("https://github.com/sorrynthx/leitstand") + "\n")
		b.WriteString("  📧 Email:     " + lipgloss.NewStyle().Foreground(ColorSuccess).Bold(true).Render("sorrynthx@gmail.com") + "\n\n")

		// Actions Footer (About Tab - Consistent structure)
		backBtn := lipgloss.NewStyle().Bold(true).Foreground(ColorPrimary).Render(i18n.T("settings_footer_about"))
		b.WriteString(backBtn)
	}

	boxWidth := screenWidth - 4
	if boxWidth < 50 {
		boxWidth = 50
	}
	boxHeight := screenHeight - 4
	if boxHeight < 15 {
		boxHeight = 15
	}

	boxStyle := lipgloss.NewStyle().
		Border(lipgloss.DoubleBorder()).
		BorderForeground(ColorPrimary).
		Padding(1, 3).
		Width(boxWidth).
		Height(boxHeight)

	return lipgloss.Place(screenWidth, screenHeight, lipgloss.Center, lipgloss.Center, boxStyle.Render(b.String()))
}
