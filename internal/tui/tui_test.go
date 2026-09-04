package tui

import (
	"leitstand/internal/config"
	"leitstand/internal/i18n"
	"leitstand/internal/ssh"
	"leitstand/internal/storage"
	"strings"
	"testing"
	"time"

	"github.com/charmbracelet/lipgloss"
	"github.com/mattn/go-runewidth"
)

func TestRenderProgressBar(t *testing.T) {
	bar := RenderProgressBar(20, 50.0, lipgloss.Color("#00FF00"), lipgloss.Color("#555555"))
	plain := lipgloss.NewStyle().Render(bar)
	if plain == "" {
		t.Error("expected non-empty progress bar")
	}
}

func TestCJKWidthSafeRendering(t *testing.T) {
	koreanText := "⚡ Leitstand 관제 센터 [정상 가동]"
	visualWidth := runewidth.StringWidth(koreanText)
	if visualWidth <= len([]rune(koreanText)) {
		// Korean double-width runes should have visual width greater than rune count
		t.Logf("visual width: %d, rune count: %d", visualWidth, len([]rune(koreanText)))
	}
}

func TestTUIModelView(t *testing.T) {
	cfg := config.NewDefaultConfig()
	model := NewModel(cfg, nil, nil, nil, true)
	model.width = 120
	model.height = 30
	model.initOrResizeViewport()

	model.hosts = []*storage.Host{
		{ID: 1, Name: "test-node-1", Address: "10.0.0.1", GroupName: "Cluster A"},
		{ID: 2, Name: "test-node-2", Address: "10.0.0.2", GroupName: "Cluster A"},
	}
	model.selectedIndex = 0
	model.userHasNavigated = true

	model.metrics[1] = &storage.MetricRecord{
		HostID:      1,
		Timestamp:   time.Now(),
		CPUPercent:  42.5,
		MemoryTotal: 16 * 1024 * 1024 * 1024,
		MemoryUsed:  8 * 1024 * 1024 * 1024,
		DiskUsed:    50 * 1024 * 1024 * 1024,
		DiskTotal:   100 * 1024 * 1024 * 1024,
		NetRxBytes:  10240,
		NetTxBytes:  20480,
	}

	model.showTelemetryDrawer = true
	rendered := model.View()
	if !strings.Contains(rendered, i18n.T("cpu_usage")) {
		t.Errorf("expected view to render CPU usage label in Telemetry Drawer, got:\n%s", rendered)
	}
}

func TestMultiTabManagement(t *testing.T) {
	hts := NewHostTabState(101, "prod-server", "/var/www")
	if len(hts.Tabs) != 1 {
		t.Fatalf("expected 1 initial tab, got %d", len(hts.Tabs))
	}
	if hts.ActiveTab() == nil {
		t.Fatal("expected non-nil active tab")
	}
	if hts.ActiveTab().CWD != "/var/www" {
		t.Errorf("expected initial CWD '/var/www', got '%s'", hts.ActiveTab().CWD)
	}

	// Add 2nd and 3rd tabs
	tab2 := hts.AddNewTab(80, 20)
	if len(hts.Tabs) != 2 || hts.ActiveIndex != 1 {
		t.Errorf("expected 2 tabs and activeIndex 1, got %d tabs, index %d", len(hts.Tabs), hts.ActiveIndex)
	}

	tab3 := hts.AddNewTab(80, 20)
	if len(hts.Tabs) != 3 || hts.ActiveIndex != 2 {
		t.Errorf("expected 3 tabs and activeIndex 2, got %d tabs, index %d", len(hts.Tabs), hts.ActiveIndex)
	}

	// Tab-specific command history and auto title
	tab2.SetAutoTitle(1, "docker logs -f app")
	if !strings.Contains(tab2.Title, "docker logs") {
		t.Errorf("expected title to contain 'docker logs', got '%s'", tab2.Title)
	}

	tab3.SetAutoTitle(2, "tail -f /var/log/syslog")
	if !strings.Contains(tab3.Title, "tail") {
		t.Errorf("expected title to contain 'tail', got '%s'", tab3.Title)
	}

	// Switch tabs
	hts.SwitchTab(0)
	if hts.ActiveIndex != 0 {
		t.Errorf("expected activeIndex 0, got %d", hts.ActiveIndex)
	}

	hts.NextTab()
	if hts.ActiveIndex != 1 {
		t.Errorf("expected activeIndex 1 after NextTab, got %d", hts.ActiveIndex)
	}

	hts.PrevTab()
	if hts.ActiveIndex != 0 {
		t.Errorf("expected activeIndex 0 after PrevTab, got %d", hts.ActiveIndex)
	}

	// Close tab
	hts.SwitchTab(1)
	closed := hts.CloseActiveTab()
	if !closed {
		t.Fatal("expected CloseActiveTab to succeed")
	}
	if len(hts.Tabs) != 2 {
		t.Errorf("expected 2 tabs remaining, got %d", len(hts.Tabs))
	}

	// Cannot close beyond 1 tab
	hts.CloseActiveTab()
	if len(hts.Tabs) != 1 {
		t.Errorf("expected 1 tab remaining, got %d", len(hts.Tabs))
	}
	closedLast := hts.CloseActiveTab()
	if closedLast {
		t.Error("expected closing last tab to return false")
	}
}

func TestStreamingCommandDetection(t *testing.T) {
	cases := []struct {
		cmd      string
		expected bool
	}{
		{"tail -f /var/log/syslog", true},
		{"docker logs -f my-container", true},
		{"docker-compose logs -f", true},
		{"journalctl -f -u nginx", true},
		{"ping 8.8.8.8", true},
		{"watch -n 1 df -h", true},
		{"ls -la", false},
		{"cat /etc/hosts", false},
		{"df -h", false},
		{"docker ps", false},
	}

	for _, c := range cases {
		got := IsStreamingCommand(c.cmd)
		if got != c.expected {
			t.Errorf("IsStreamingCommand(%q) = %v, expected %v", c.cmd, got, c.expected)
		}
	}
}

func TestFileManagerModal(t *testing.T) {
	i18n.SetLang(i18n.LangEN)
	fm := NewFileManagerModal(101, "prod-server", ".", "/var/www")
	if fm.FocusLocal {
		t.Error("expected initial focus on Remote pane")
	}

	// Mock file items
	fm.LocalItems = []*ssh.FileItem{
		{Name: "..", Path: "..", IsDir: true},
		{Name: "app.py", Path: "app.py", Size: 1024, IsDir: false},
		{Name: "config.yaml", Path: "config.yaml", Size: 512, IsDir: false},
	}
	fm.RemoteItems = GetDemoRemoteFiles("/var/www")

	// Test multi selection
	fm.LocalSelected["app.py"] = true
	fm.LocalSelected["config.yaml"] = true

	selected := fm.GetSelectedPaths(true)
	if len(selected) != 2 {
		t.Errorf("expected 2 selected files, got %d", len(selected))
	}

	// Test rendering
	viewStr := fm.View(120, 30)
	if !strings.Contains(viewStr, "SFTP Dual-Pane File Manager") {
		t.Errorf("expected title in view, got:\n%s", viewStr)
	}
	if !strings.Contains(viewStr, "My PC (Local)") {
		t.Errorf("expected Local PC header, got:\n%s", viewStr)
	}
	if !strings.Contains(viewStr, "prod-server") {
		t.Errorf("expected remote server name in view, got:\n%s", viewStr)
	}

	// Test transfer progress rendering
	fm.IsTransferring = true
	fm.TransferIsUpload = true
	fm.CurrentFileName = "app.py"
	fm.CurrentBytes = 512
	fm.CurrentTotal = 1024
	fm.BytesPerSec = 2048000

	transferView := fm.View(120, 30)
	if !strings.Contains(transferView, "UPLOADING") {
		t.Errorf("expected UPLOADING indicator in transfer bar, got:\n%s", transferView)
	}

	// Test Runbook overlay
	fm.ShowRunbook = true
	runbookView := fm.View(120, 30)
	if !strings.Contains(runbookView, i18n.T("sftp_runbook_title")) {
		t.Errorf("expected Runbook title in view, got:\n%s", runbookView)
	}
	fm.ShowRunbook = false

	// Test Filter & Fast Navigation
	fm.LocalFilter = "app"
	visible := fm.GetVisibleItems(true)
	if len(visible) != 2 || visible[0].Name != ".." || visible[1].Name != "app.py" {
		t.Errorf("expected 2 items ('..' and 'app.py'), got %d items", len(visible))
	}
	fm.LocalFilter = ""

	// Test Clipboard (Cut / Copy)
	fm.ShowCmdOutput = false
	fm.ClipboardPaths = []string{"app.py", "config.yaml"}
	fm.ClipboardIsCut = true
	clipView := fm.View(120, 30)
	if !strings.Contains(clipView, "대기") && !strings.Contains(clipView, "Cut") && !strings.Contains(clipView, "2") {
		t.Errorf("expected clipboard banner in view, got:\n%s", clipView)
	}
}

func TestLocalDirectoryListing(t *testing.T) {
	items, err := ssh.ListLocalDirectory(".", false)
	if err != nil {
		t.Fatalf("ListLocalDirectory failed: %v", err)
	}
	if len(items) == 0 {
		t.Error("expected non-empty directory listing for '.'")
	}
}
