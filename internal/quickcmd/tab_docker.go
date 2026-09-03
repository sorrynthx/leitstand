package quickcmd

var DockerCommands = []CommandItem{
	{
		ID:          "doc_ps_all",
		CategoryKey: "cat_system",
		TitleKey:    "cmd_docker_ps_title",
		DescKey:     "cmd_docker_ps_desc",
		Command:     "docker ps -a --format 'table {{.Names}}\t{{.Status}}\t{{.Ports}}'",
	},
	{
		ID:          "doc_stats",
		CategoryKey: "cat_resources",
		TitleKey:    "cmd_docker_stats_title",
		DescKey:     "cmd_docker_stats_desc",
		Command:     "docker stats --no-stream",
	},
	{
		ID:          "doc_logs_last",
		CategoryKey: "cat_logs",
		TitleKey:    "cmd_docker_logs_title",
		DescKey:     "cmd_docker_logs_desc",
		Command:     "docker logs --tail 50 $(docker ps -q -l)",
	},
	{
		ID:          "doc_df",
		CategoryKey: "cat_disk",
		TitleKey:    "cmd_disk_space_title",
		DescKey:     "cmd_disk_space_desc",
		Command:     "docker system df",
	},
}
