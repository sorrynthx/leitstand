package tui

import (
	"fmt"
	"leitstand/internal/config"
	"leitstand/internal/i18n"
	"os"
	"path/filepath"
	"time"
)

// handleDirectExport exports all custom commands with 1-click directly into the OS standard runbooks directory.
func (d *RunbookDrawer) handleDirectExport() {
	if d.store == nil {
		return
	}

	targetDir := config.DefaultRunbookDir()
	if err := os.MkdirAll(targetDir, 0755); err != nil {
		d.bannerMessage = fmt.Sprintf(i18n.T("custom_cmd_export_fail"), err.Error())
		return
	}

	ts := time.Now().Format("20060102_150405")
	targetFile := filepath.Join(targetDir, fmt.Sprintf("leitstand_runbook_%s.json", ts))

	count, err := d.store.ExportCustomCommandsJSON(targetFile)
	if err != nil {
		d.bannerMessage = fmt.Sprintf(i18n.T("custom_cmd_export_fail"), err.Error())
		return
	}

	d.bannerMessage = fmt.Sprintf(i18n.T("custom_cmd_export_success"), count, targetFile)
}

// openImportPicker opens the file picker directly in the OS standard runbooks directory.
func (d *RunbookDrawer) openImportPicker(w, h int) {
	initDir := config.DefaultRunbookDir()
	_ = os.MkdirAll(initDir, 0755)

	fp := NewFilePickerModal(initDir, w, h)
	fp.CustomTitle = "📥 " + i18n.T("custom_cmd_import_title")
	fp.CustomHints = "[↑/↓, PgUp/Dn] 이동  [Enter] 런북 파일 가져오기  [Esc] 취소"
	d.filePicker = fp
}

func (d *RunbookDrawer) handleFilePicked(pickedPath string) {
	if d.store == nil || pickedPath == "" {
		return
	}

	imported, skipped, err := d.store.ImportCustomCommandsJSON(pickedPath)
	if err != nil {
		d.bannerMessage = fmt.Sprintf(i18n.T("custom_cmd_import_fail"), err.Error())
		return
	}
	d.bannerMessage = fmt.Sprintf(i18n.T("custom_cmd_import_success"), imported, skipped) + " (" + filepath.Base(pickedPath) + ")"
	d.loadCustomCommands()
}
