package quickcmd

var AlpineCommands = []CommandItem{
	{
		ID:          "alp_apk_update",
		CategoryKey: "cat_system",
		TitleKey:    "cmd_apk_update_title",
		DescKey:     "cmd_apk_update_desc",
		Command:     "apk update && apk upgrade --dry-run",
	},
	{
		ID:          "alp_rc_status",
		CategoryKey: "cat_system",
		TitleKey:    "cmd_rc_status_title",
		DescKey:     "cmd_rc_status_desc",
		Command:     "rc-status",
	},
	{
		ID:          "alp_dmesg",
		CategoryKey: "cat_logs",
		TitleKey:    "cmd_dmesg_title",
		DescKey:     "cmd_dmesg_desc",
		Command:     "dmesg | tail -30",
	},
}
