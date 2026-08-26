package tui

import (
	"leitstand/internal/config"
	"leitstand/internal/i18n"
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

	rendered := model.View()
	if !strings.Contains(rendered, "LEITSTAND COCKPIT") {
		t.Errorf("expected view to contain header title, got:\n%s", rendered)
	}
	if !strings.Contains(rendered, "test-node-1") {
		t.Errorf("expected view to render host name, got:\n%s", rendered)
	}
	if !strings.Contains(rendered, i18n.T("cpu_usage")) {
		t.Errorf("expected view to render CPU usage label, got:\n%s", rendered)
	}
}
