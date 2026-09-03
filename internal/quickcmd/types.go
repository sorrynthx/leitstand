package quickcmd

type OSTab int

const (
	OSTabShortcuts OSTab = iota
	OSTabCommon
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
	{Tab: OSTabShortcuts, LabelKey: "drawer_tab_shortcuts", Badge: "⌨️ Shortcuts"},
	{Tab: OSTabCommon, LabelKey: "drawer_tab_common", Badge: "🐧 Common"},
	{Tab: OSTabUbuntu, LabelKey: "drawer_tab_ubuntu", Badge: "🟠 Ubuntu/Debian"},
	{Tab: OSTabRHEL, LabelKey: "drawer_tab_rhel", Badge: "🔴 RHEL/Rocky"},
	{Tab: OSTabAlpine, LabelKey: "drawer_tab_alpine", Badge: "🏔️ Alpine"},
	{Tab: OSTabDocker, LabelKey: "drawer_tab_docker", Badge: "🐳 Docker"},
}

var Catalog = map[OSTab][]CommandItem{
	OSTabShortcuts: ShortcutsCommands,
	OSTabCommon:    CommonCommands,
	OSTabUbuntu:    UbuntuCommands,
	OSTabRHEL:      RHELCommands,
	OSTabAlpine:    AlpineCommands,
	OSTabDocker:    DockerCommands,
}
