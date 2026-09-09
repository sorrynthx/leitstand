package quickcmd

var RHELCommands = []CommandItem{
	{
		ID:          "rhel_dnf_check",
		CategoryKey: "cat_system",
		TitleKey:    "cmd_dnf_check_title",
		DescKey:     "cmd_dnf_check_desc",
		Command:     "dnf check-update",
	},
	{
		ID:          "rhel_svc_failed",
		CategoryKey: "cat_system",
		TitleKey:    "cmd_svc_failed_title",
		DescKey:     "cmd_svc_failed_desc",
		Command:     "systemctl list-units --type=service --state=failed",
	},
	{
		ID:          "rhel_blame",
		CategoryKey: "cat_system",
		TitleKey:    "cmd_systemd_blame_title",
		DescKey:     "cmd_systemd_blame_desc",
		Command:     "systemd-analyze blame | head -5",
	},
	{
		ID:          "rhel_journal_recent",
		CategoryKey: "cat_logs",
		TitleKey:    "cmd_journal_recent_title",
		DescKey:     "cmd_journal_recent_desc",
		Command:     "journalctl -p 3 --since \"1 hour ago\" --no-pager",
	},
	{
		ID:          "rhel_firewall",
		CategoryKey: "cat_network",
		TitleKey:    "cmd_firewall_title",
		DescKey:     "cmd_firewall_desc",
		Command:     "firewall-cmd --list-all",
	},
	{
		ID:          "rhel_selinux",
		CategoryKey: "cat_system",
		TitleKey:    "cmd_selinux_title",
		DescKey:     "cmd_selinux_desc",
		Command:     "sestatus",
	},
	{
		ID:          "rhel_dnf_clean",
		CategoryKey: "cat_disk",
		TitleKey:    "cmd_pkg_clean_title",
		DescKey:     "cmd_pkg_clean_desc",
		Command:     "dnf clean all",
	},
}
