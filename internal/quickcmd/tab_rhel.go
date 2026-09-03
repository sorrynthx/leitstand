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
}
