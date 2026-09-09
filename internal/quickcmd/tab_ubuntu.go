package quickcmd

var UbuntuCommands = []CommandItem{
	{
		ID:          "ubu_apt_update",
		CategoryKey: "cat_system",
		TitleKey:    "cmd_apt_upg_title",
		DescKey:     "cmd_apt_upg_desc",
		Command:     "apt update && apt list --upgradable",
	},
	{
		ID:          "ubu_svc_failed",
		CategoryKey: "cat_system",
		TitleKey:    "cmd_svc_failed_title",
		DescKey:     "cmd_svc_failed_desc",
		Command:     "systemctl list-units --type=service --state=failed",
	},
	{
		ID:          "ubu_svc_running",
		CategoryKey: "cat_system",
		TitleKey:    "cmd_svc_running_title",
		DescKey:     "cmd_svc_running_desc",
		Command:     "systemctl list-units --type=service --state=running",
	},
	{
		ID:          "ubu_blame",
		CategoryKey: "cat_system",
		TitleKey:    "cmd_systemd_blame_title",
		DescKey:     "cmd_systemd_blame_desc",
		Command:     "systemd-analyze blame | head -5",
	},
	{
		ID:          "ubu_journal_err",
		CategoryKey: "cat_logs",
		TitleKey:    "cmd_journal_err_title",
		DescKey:     "cmd_journal_err_desc",
		Command:     "journalctl -xe -p 3 -n 50 --no-pager",
	},
	{
		ID:          "ubu_journal_recent",
		CategoryKey: "cat_logs",
		TitleKey:    "cmd_journal_recent_title",
		DescKey:     "cmd_journal_recent_desc",
		Command:     "journalctl -p 3 --since \"1 hour ago\" --no-pager",
	},
	{
		ID:          "ubu_ufw_status",
		CategoryKey: "cat_network",
		TitleKey:    "cmd_ufw_title",
		DescKey:     "cmd_ufw_desc",
		Command:     "ufw status verbose",
	},
	{
		ID:          "ubu_apt_clean",
		CategoryKey: "cat_disk",
		TitleKey:    "cmd_pkg_clean_title",
		DescKey:     "cmd_pkg_clean_desc",
		Command:     "apt clean && apt autoremove -y --dry-run",
	},
	{
		ID:          "ubu_reboot_required",
		CategoryKey: "cat_system",
		TitleKey:    "cmd_reboot_check_title",
		DescKey:     "cmd_reboot_check_desc",
		Command:     "[ -f /var/run/reboot-required ] && echo '⚠️ Reboot Required' || echo '✅ No Reboot Required'",
	},
}
