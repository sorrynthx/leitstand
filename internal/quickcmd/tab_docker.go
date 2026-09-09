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
		ID:          "doc_ps_exited",
		CategoryKey: "cat_system",
		TitleKey:    "cmd_docker_exited_title",
		DescKey:     "cmd_docker_exited_desc",
		Command:     "docker ps -a --filter \"status=exited\" --format \"table {{.Names}}\t{{.Status}}\"",
	},
	{
		ID:          "doc_stats",
		CategoryKey: "cat_resources",
		TitleKey:    "cmd_docker_stats_title",
		DescKey:     "cmd_docker_stats_desc",
		Command:     "docker stats --no-stream --format \"table {{.Name}}\t{{.CPUPerc}}\t{{.MemUsage}}\t{{.MemPerc}}\t{{.NetIO}}\"",
	},
	{
		ID:          "doc_logs_last",
		CategoryKey: "cat_logs",
		TitleKey:    "cmd_docker_logs_title",
		DescKey:     "cmd_docker_logs_desc",
		Command:     "docker logs --tail 50 $(docker ps -q -l)",
	},
	{
		ID:          "doc_logs_err",
		CategoryKey: "cat_logs",
		TitleKey:    "cmd_docker_err_logs_title",
		DescKey:     "cmd_docker_err_logs_desc",
		Command:     "docker logs --tail 100 $(docker ps -q -l) 2>&1 | grep -i -E \"error|warn|fatal|exception\" | tail -20",
	},
	{
		ID:          "doc_net_ip",
		CategoryKey: "cat_network",
		TitleKey:    "cmd_docker_net_ip_title",
		DescKey:     "cmd_docker_net_ip_desc",
		Command:     "docker inspect -f '{{range .NetworkSettings.Networks}}{{.IPAddress}}{{end}}' $(docker ps -q -l)",
	},
	{
		ID:          "doc_df",
		CategoryKey: "cat_disk",
		TitleKey:    "cmd_disk_space_title",
		DescKey:     "cmd_disk_space_desc",
		Command:     "docker system df",
	},
	{
		ID:          "doc_prune_all",
		CategoryKey: "cat_disk",
		TitleKey:    "cmd_docker_prune_title",
		DescKey:     "cmd_docker_prune_desc",
		Command:     "docker system prune -f",
	},
	{
		ID:          "doc_volume_prune",
		CategoryKey: "cat_disk",
		TitleKey:    "cmd_docker_vol_prune_title",
		DescKey:     "cmd_docker_vol_prune_desc",
		Command:     "docker volume prune -f",
	},
}
