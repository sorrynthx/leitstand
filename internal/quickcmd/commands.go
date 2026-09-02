package quickcmd

import (
	"strings"
)

type OSTab int

const (
	OSTabCommon OSTab = iota
	OSTabUbuntu
	OSTabRHEL
	OSTabAlpine
	OSTabDocker
	OSTabCount
)

type CommandItem struct {
	ID          string
	CategoryKey string
	TitleKey    string
	DescKey     string
	Command     string
}

type TabInfo struct {
	Tab      OSTab
	LabelKey string
	Badge    string
}

var Tabs = []TabInfo{
	{Tab: OSTabCommon, LabelKey: "drawer_tab_common", Badge: "🐧 Common"},
	{Tab: OSTabUbuntu, LabelKey: "drawer_tab_ubuntu", Badge: "🟠 Ubuntu/Debian"},
	{Tab: OSTabRHEL, LabelKey: "drawer_tab_rhel", Badge: "🔴 RHEL/Rocky"},
	{Tab: OSTabAlpine, LabelKey: "drawer_tab_alpine", Badge: "🏔️ Alpine"},
	{Tab: OSTabDocker, LabelKey: "drawer_tab_docker", Badge: "🐳 Docker"},
}

func DetectOSTab(distro string) OSTab {
	d := strings.ToLower(distro)
	switch {
	case strings.Contains(d, "ubuntu"), strings.Contains(d, "debian"), strings.Contains(d, "mint"):
		return OSTabUbuntu
	case strings.Contains(d, "rhel"), strings.Contains(d, "red hat"), strings.Contains(d, "centos"), strings.Contains(d, "rocky"), strings.Contains(d, "almalinux"), strings.Contains(d, "fedora"):
		return OSTabRHEL
	case strings.Contains(d, "alpine"):
		return OSTabAlpine
	default:
		return OSTabCommon
	}
}

var Catalog = map[OSTab][]CommandItem{
	OSTabCommon: {
		{
			ID:          "com_su_root",
			CategoryKey: "cat_interactive",
			TitleKey:    "cmd_su_root_title",
			DescKey:     "cmd_su_root_desc",
			Command:     "su -",
		},
		{
			ID:          "com_sudo_i",
			CategoryKey: "cat_interactive",
			TitleKey:    "cmd_sudo_i_title",
			DescKey:     "cmd_sudo_i_desc",
			Command:     "sudo -i",
		},
		{
			ID:          "com_htop",
			CategoryKey: "cat_interactive",
			TitleKey:    "cmd_htop_title",
			DescKey:     "cmd_htop_desc",
			Command:     "htop",
		},
		{
			ID:          "com_top_cpu",
			CategoryKey: "cat_resources",
			TitleKey:    "cmd_top_cpu_title",
			DescKey:     "cmd_top_cpu_desc",
			Command:     "ps aux --sort=-%cpu | head -10",
		},
		{
			ID:          "com_top_mem",
			CategoryKey: "cat_resources",
			TitleKey:    "cmd_top_mem_title",
			DescKey:     "cmd_top_mem_desc",
			Command:     "ps aux --sort=-%mem | head -10",
		},
		{
			ID:          "com_free_m",
			CategoryKey: "cat_resources",
			TitleKey:    "cmd_free_title",
			DescKey:     "cmd_free_desc",
			Command:     "free -h",
		},
		{
			ID:          "com_open_ports",
			CategoryKey: "cat_network",
			TitleKey:    "cmd_ports_title",
			DescKey:     "cmd_ports_desc",
			Command:     "ss -tulpn",
		},
		{
			ID:          "com_ip_brief",
			CategoryKey: "cat_network",
			TitleKey:    "cmd_ip_brief_title",
			DescKey:     "cmd_ip_brief_desc",
			Command:     "ip -br a",
		},
		{
			ID:          "com_df_h",
			CategoryKey: "cat_disk",
			TitleKey:    "cmd_df_title",
			DescKey:     "cmd_df_desc",
			Command:     "df -hT",
		},
		{
			ID:          "com_du_top",
			CategoryKey: "cat_disk",
			TitleKey:    "cmd_du_top_title",
			DescKey:     "cmd_du_top_desc",
			Command:     "du -h --max-depth=1 /var 2>/dev/null | sort -hr | head -10",
		},
		{
			ID:          "com_dmesg_err",
			CategoryKey: "cat_logs",
			TitleKey:    "cmd_dmesg_title",
			DescKey:     "cmd_dmesg_desc",
			Command:     "dmesg -T --level=err,warn | tail -20",
		},
		{
			ID:          "com_systemctl_failed",
			CategoryKey: "cat_system",
			TitleKey:    "cmd_svc_failed_title",
			DescKey:     "cmd_svc_failed_desc",
			Command:     "systemctl list-units --type=service --state=failed",
		},
		{
			ID:          "com_systemctl_running",
			CategoryKey: "cat_system",
			TitleKey:    "cmd_svc_running_title",
			DescKey:     "cmd_svc_running_desc",
			Command:     "systemctl list-units --type=service --state=running",
		},
		{
			ID:          "com_journalctl_err",
			CategoryKey: "cat_logs",
			TitleKey:    "cmd_journal_err_title",
			DescKey:     "cmd_journal_err_desc",
			Command:     "journalctl -xe -p 3 -n 50 --no-pager",
		},
		{
			ID:          "com_syslog_tail",
			CategoryKey: "cat_logs",
			TitleKey:    "cmd_syslog_tail_title",
			DescKey:     "cmd_syslog_tail_desc",
			Command:     "tail -n 50 /var/log/syslog",
		},
		{
			ID:          "com_ufw_status",
			CategoryKey: "cat_network",
			TitleKey:    "cmd_ufw_title",
			DescKey:     "cmd_ufw_desc",
			Command:     "ufw status verbose",
		},
		{
			ID:          "com_apt_upgradable",
			CategoryKey: "cat_system",
			TitleKey:    "cmd_apt_upg_title",
			DescKey:     "cmd_apt_upg_desc",
			Command:     "apt list --upgradable",
		},
		{
			ID:          "com_reboot_required",
			CategoryKey: "cat_system",
			TitleKey:    "cmd_reboot_check_title",
			DescKey:     "cmd_reboot_check_desc",
			Command:     "[ -f /var/run/reboot-required ] && echo '⚠️ Reboot Required' || echo '✅ No Reboot Required'",
		},
	},
}
