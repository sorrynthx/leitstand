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

// DetectOSTab maps host distro string to the most suitable OSTab.
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

// Catalog holds all curated commands categorized by OS tab.
var Catalog = map[OSTab][]CommandItem{
	OSTabCommon: {
		// Resources
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
		// Network
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
		// Disk
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
		// Logs
		{
			ID:          "com_dmesg_err",
			CategoryKey: "cat_logs",
			TitleKey:    "cmd_dmesg_title",
			DescKey:     "cmd_dmesg_desc",
			Command:     "dmesg -T --level=err,warn | tail -20",
		},
	},

	OSTabUbuntu: {
		// Services (systemd)
		{
			ID:          "ubu_failed_svc",
			CategoryKey: "cat_services",
			TitleKey:    "cmd_failed_svc_title",
			DescKey:     "cmd_failed_svc_desc",
			Command:     "systemctl --failed",
		},
		{
			ID:          "ubu_active_svc",
			CategoryKey: "cat_services",
			TitleKey:    "cmd_active_svc_title",
			DescKey:     "cmd_active_svc_desc",
			Command:     "systemctl list-units --type=service --state=running",
		},
		// Logs (journalctl)
		{
			ID:          "ubu_journal_err",
			CategoryKey: "cat_logs",
			TitleKey:    "cmd_journal_err_title",
			DescKey:     "cmd_journal_err_desc",
			Command:     "journalctl -xe -p 3 -n 50 --no-pager",
		},
		{
			ID:          "ubu_syslog_tail",
			CategoryKey: "cat_logs",
			TitleKey:    "cmd_syslog_tail_title",
			DescKey:     "cmd_syslog_tail_desc",
			Command:     "tail -n 50 /var/log/syslog",
		},
		// Packages (APT / UFW)
		{
			ID:          "ubu_ufw_status",
			CategoryKey: "cat_security",
			TitleKey:    "cmd_ufw_title",
			DescKey:     "cmd_ufw_desc",
			Command:     "ufw status verbose",
		},
		{
			ID:          "ubu_upgradable",
			CategoryKey: "cat_packages",
			TitleKey:    "cmd_apt_upgradable_title",
			DescKey:     "cmd_apt_upgradable_desc",
			Command:     "apt list --upgradable",
		},
		{
			ID:          "ubu_reboot_req",
			CategoryKey: "cat_packages",
			TitleKey:    "cmd_reboot_req_title",
			DescKey:     "cmd_reboot_req_desc",
			Command:     "[ -f /var/run/reboot-required ] && echo '⚠️ Reboot Required' || echo '✅ No Reboot Required'",
		},
	},

	OSTabRHEL: {
		// Services (systemd)
		{
			ID:          "rhel_failed_svc",
			CategoryKey: "cat_services",
			TitleKey:    "cmd_failed_svc_title",
			DescKey:     "cmd_failed_svc_desc",
			Command:     "systemctl --failed",
		},
		{
			ID:          "rhel_journal_err",
			CategoryKey: "cat_logs",
			TitleKey:    "cmd_journal_err_title",
			DescKey:     "cmd_journal_err_desc",
			Command:     "journalctl -xe -p 3 -n 50 --no-pager",
		},
		// Security / Firewall
		{
			ID:          "rhel_firewall",
			CategoryKey: "cat_security",
			TitleKey:    "cmd_firewalld_title",
			DescKey:     "cmd_firewalld_desc",
			Command:     "firewall-cmd --list-all",
		},
		{
			ID:          "rhel_selinux",
			CategoryKey: "cat_security",
			TitleKey:    "cmd_selinux_title",
			DescKey:     "cmd_selinux_desc",
			Command:     "sestatus",
		},
		// Packages (DNF)
		{
			ID:          "rhel_dnf_check",
			CategoryKey: "cat_packages",
			TitleKey:    "cmd_dnf_check_title",
			DescKey:     "cmd_dnf_check_desc",
			Command:     "dnf check-update",
		},
	},

	OSTabAlpine: {
		// Services (OpenRC)
		{
			ID:          "alp_rc_status",
			CategoryKey: "cat_services",
			TitleKey:    "cmd_rc_status_title",
			DescKey:     "cmd_rc_status_desc",
			Command:     "rc-status",
		},
		{
			ID:          "alp_rc_service",
			CategoryKey: "cat_services",
			TitleKey:    "cmd_rc_list_title",
			DescKey:     "cmd_rc_list_desc",
			Command:     "rc-service --list",
		},
		// Logs & Packages
		{
			ID:          "alp_dmesg",
			CategoryKey: "cat_logs",
			TitleKey:    "cmd_dmesg_title",
			DescKey:     "cmd_dmesg_desc",
			Command:     "dmesg | tail -30",
		},
		{
			ID:          "alp_apk_upgradable",
			CategoryKey: "cat_packages",
			TitleKey:    "cmd_apk_upgradable_title",
			DescKey:     "cmd_apk_upgradable_desc",
			Command:     "apk version -v",
		},
	},

	OSTabDocker: {
		// Containers
		{
			ID:          "doc_ps_all",
			CategoryKey: "cat_docker",
			TitleKey:    "cmd_docker_ps_title",
			DescKey:     "cmd_docker_ps_desc",
			Command:     `docker ps -a --format "table {{.ID}}\t{{.Names}}\t{{.Status}}\t{{.Ports}}"`,
		},
		{
			ID:          "doc_stats",
			CategoryKey: "cat_docker",
			TitleKey:    "cmd_docker_stats_title",
			DescKey:     "cmd_docker_stats_desc",
			Command:     "docker stats --no-stream",
		},
		{
			ID:          "doc_df",
			CategoryKey: "cat_docker",
			TitleKey:    "cmd_docker_df_title",
			DescKey:     "cmd_docker_df_desc",
			Command:     "docker system df",
		},
		{
			ID:          "doc_compose_ps",
			CategoryKey: "cat_docker",
			TitleKey:    "cmd_docker_compose_title",
			DescKey:     "cmd_docker_compose_desc",
			Command:     "docker compose ps",
		},
		{
			ID:          "doc_tail_logs",
			CategoryKey: "cat_docker",
			TitleKey:    "cmd_docker_logs_title",
			DescKey:     "cmd_docker_logs_desc",
			Command:     "docker logs --tail 50 $(docker ps -q | head -1)",
		},
	},
}
