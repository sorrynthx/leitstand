package tui

import (
	"fmt"
	"leitstand/internal/i18n"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// View renders the complete TunnelModal overlay.
func (tm *TunnelModal) View(width, height int) string {
	modalWidth := width - 10
	if modalWidth > 86 {
		modalWidth = 86
	}
	if modalWidth < 65 {
		modalWidth = 65
	}

	var b strings.Builder

	// 1. Header
	title := TitleStyle.Render(i18n.T("tunnel_modal_title"))
	hostBadge := BadgeStyle.Copy().Background(ColorHighlight).Render(
		fmt.Sprintf("🌐 %s (%s)", tm.host.Name, tm.host.Address),
	)
	b.WriteString(lipgloss.JoinHorizontal(lipgloss.Center, title, "  ", hostBadge) + "\n\n")

	// 2. Banner: Delete confirmation or Error
	if tm.isDeleting && len(tm.tunnels) > tm.selectedIndex {
		tun := tm.tunnels[tm.selectedIndex]
		confirmStyle := lipgloss.NewStyle().
			Background(ColorDanger).
			Foreground(ColorWhite).
			Bold(true).
			Padding(0, 1)
		b.WriteString(confirmStyle.Render(fmt.Sprintf(i18n.T("tunnel_delete_confirm"), tun.Name)) + "\n\n")
	} else if tm.errMessage != "" {
		errStyle := lipgloss.NewStyle().Foreground(ColorDanger).Bold(true).Padding(0, 1)
		b.WriteString(errStyle.Render(tm.errMessage) + "\n\n")
	}

	// 3. Main Content: Table or Add Form
	if tm.isAdding {
		b.WriteString(tm.renderAddForm(modalWidth - 4))
	} else {
		b.WriteString(tm.renderTunnelList(modalWidth - 4))
	}

	// 4. Footer shortcuts
	b.WriteString("\n" + tm.renderFooter(modalWidth-4))

	modalBox := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(ColorPrimary).
		Background(ColorBg).
		Padding(1, 2).
		Width(modalWidth).
		Render(b.String())

	return lipgloss.Place(width, height, lipgloss.Center, lipgloss.Center, modalBox)
}

func (tm *TunnelModal) renderTunnelList(width int) string {
	var b strings.Builder

	if len(tm.tunnels) == 0 {
		emptyStyle := lipgloss.NewStyle().Foreground(ColorMuted).Italic(true).Padding(2, 2)
		b.WriteString(emptyStyle.Render(i18n.T("tunnel_no_tunnels")) + "\n")
		return b.String()
	}

	// Table Header
	hdrStatus := lipgloss.NewStyle().Bold(true).Width(10).Render(i18n.T("tunnel_col_status"))
	hdrName := lipgloss.NewStyle().Bold(true).Width(20).Render(i18n.T("tunnel_col_name"))
	hdrLocal := lipgloss.NewStyle().Bold(true).Width(14).Render(i18n.T("tunnel_col_local"))
	hdrRemote := lipgloss.NewStyle().Bold(true).Width(22).Render(i18n.T("tunnel_col_remote"))
	hdrConns := lipgloss.NewStyle().Bold(true).Width(8).Render(i18n.T("tunnel_col_conns"))

	headerLine := lipgloss.JoinHorizontal(lipgloss.Left, hdrStatus, hdrName, hdrLocal, hdrRemote, hdrConns)
	b.WriteString(lipgloss.NewStyle().Foreground(ColorPrimary).Render(headerLine) + "\n")
	b.WriteString(lipgloss.NewStyle().Foreground(ColorBorder).Render(strings.Repeat("─", width)) + "\n")

	// Table Rows
	for i, tun := range tm.tunnels {
		isActive := tm.tunnelMgr.IsActive(tun.ID)
		var statusStr string
		var connsCount int64

		if isActive {
			statusStr = lipgloss.NewStyle().Bold(true).Foreground(ColorSuccess).Render("🟢 ON ")
			if at, ok := tm.tunnelMgr.GetActive(tun.ID); ok {
				connsCount = at.Conns()
			}
		} else {
			statusStr = lipgloss.NewStyle().Foreground(ColorMuted).Render("🔴 OFF")
		}

		colStat := lipgloss.NewStyle().Width(10).Render(statusStr)
		colName := lipgloss.NewStyle().Width(20).Render(truncate(tun.Name, 18))
		colLoc := lipgloss.NewStyle().Foreground(ColorPrimary).Width(14).Render(fmt.Sprintf(":%d", tun.LocalPort))
		colRem := lipgloss.NewStyle().Foreground(ColorText).Width(22).Render(fmt.Sprintf("%s:%d", tun.RemoteHost, tun.RemotePort))
		colConns := lipgloss.NewStyle().Foreground(ColorMuted).Width(8).Render(fmt.Sprintf("%d", connsCount))

		row := lipgloss.JoinHorizontal(lipgloss.Left, colStat, colName, colLoc, colRem, colConns)

		if i == tm.selectedIndex {
			row = lipgloss.NewStyle().Background(ColorHighlight).Bold(true).Render(row)
		}
		b.WriteString(row + "\n")
	}

	return b.String()
}

func (tm *TunnelModal) renderAddForm(width int) string {
	var b strings.Builder

	b.WriteString(lipgloss.NewStyle().Bold(true).Foreground(ColorWarning).Render(i18n.T("tunnel_form_title_add")) + "\n\n")

	labels := []string{
		i18n.T("tunnel_form_name"),
		i18n.T("tunnel_form_local_port"),
		i18n.T("tunnel_form_remote_host"),
		i18n.T("tunnel_form_remote_port"),
	}

	for i, input := range tm.inputs {
		labelStyle := lipgloss.NewStyle().Bold(true).Width(26)
		if i == tm.focusedInput {
			labelStyle = labelStyle.Foreground(ColorPrimary)
		} else {
			labelStyle = labelStyle.Foreground(ColorText)
		}

		b.WriteString(lipgloss.JoinHorizontal(lipgloss.Left, labelStyle.Render(labels[i]+": "), input.View()) + "\n")
	}

	b.WriteString("\n" + lipgloss.NewStyle().Foreground(ColorMuted).Render(i18n.T("tunnel_form_hint")) + "\n")
	return b.String()
}

func (tm *TunnelModal) renderFooter(width int) string {
	if tm.isAdding {
		return ""
	}

	if tm.isDeleting {
		btnConfirm := lipgloss.NewStyle().Bold(true).Foreground(ColorDanger).Render(i18n.T("tunnel_btn_confirm_delete"))
		btnCancel := lipgloss.NewStyle().Bold(true).Foreground(ColorMuted).Render(i18n.T("tunnel_btn_cancel_delete"))
		line := lipgloss.JoinHorizontal(lipgloss.Left, btnConfirm, "        ", btnCancel)
		return lipgloss.NewStyle().Foreground(ColorBorder).Render(strings.Repeat("─", width)) + "\n" + line
	}

	btnToggle := lipgloss.NewStyle().Bold(true).Foreground(ColorPrimary).Render(i18n.T("tunnel_btn_toggle"))
	btnAdd := lipgloss.NewStyle().Bold(true).Foreground(ColorSuccess).Render(i18n.T("tunnel_btn_add"))
	btnDel := lipgloss.NewStyle().Bold(true).Foreground(ColorDanger).Render(i18n.T("tunnel_btn_delete"))
	btnClose := lipgloss.NewStyle().Bold(true).Foreground(ColorMuted).Render(i18n.T("tunnel_btn_close"))

	line := lipgloss.JoinHorizontal(lipgloss.Left, btnToggle, "    ", btnAdd, "    ", btnDel, "    ", btnClose)
	return lipgloss.NewStyle().Foreground(ColorBorder).Render(strings.Repeat("─", width)) + "\n" + line
}

func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-2] + ".."
}
