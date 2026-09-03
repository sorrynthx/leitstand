package tui

import (
	"fmt"
	"leitstand/internal/i18n"
	"os"
	"path/filepath"
	"time"

	tea "github.com/charmbracelet/bubbletea"
)

func (s *SettingsModal) updateSettingsDatabaseTab(msg tea.KeyMsg) (tea.Cmd, bool) {
	if s.showRekeyModal {
		return s.updateRekeyModal(msg)
	}

	if s.dbConfirmAction > 0 {
		switch msg.String() {
		case "enter", "y", "Y":
			action := s.dbConfirmAction
			s.dbConfirmAction = 0
			return s.executeDatabaseAction(action)
		case "esc", "n", "N":
			s.dbConfirmAction = 0
			return nil, true
		}
		return nil, true
	}

	switch msg.String() {
	case "up", "k":
		if s.dbActionIndex > 0 {
			s.dbActionIndex--
		}
		return nil, true

	case "down", "j":
		if s.dbActionIndex < len(dbActions)-1 {
			s.dbActionIndex++
		}
		return nil, true

	case "left", "h":
		s.retentionIndex = (s.retentionIndex - 1 + len(RetentionDays)) % len(RetentionDays)
		days := RetentionDays[s.retentionIndex]
		if s.store != nil {
			_ = s.store.SetSetting("metrics_retention_days", fmt.Sprintf("%d", days))
		}
		s.errMessage = ""
		s.successMessage = fmt.Sprintf("⏱️ 메트릭 보관 주기가 %d일로 변경되었습니다.", days)
		return nil, true

	case "right", "l":
		s.retentionIndex = (s.retentionIndex + 1) % len(RetentionDays)
		days := RetentionDays[s.retentionIndex]
		if s.store != nil {
			_ = s.store.SetSetting("metrics_retention_days", fmt.Sprintf("%d", days))
		}
		s.errMessage = ""
		s.successMessage = fmt.Sprintf("⏱️ 메트릭 보관 주기가 %d일로 변경되었습니다.", days)
		return nil, true

	case "enter", " ":
		s.triggerDatabaseActionConfirm(s.dbActionIndex + 1)
		return nil, true
	}

	return nil, false
}

func (s *SettingsModal) triggerDatabaseActionConfirm(action int) {
	s.dbConfirmAction = action
	switch action {
	case 1:
		s.dbConfirmPrompt = i18n.T("db_confirm_prune")
	case 2:
		s.dbConfirmPrompt = i18n.T("db_confirm_export_metrics")
	case 3:
		s.dbConfirmPrompt = i18n.T("db_confirm_export_hosts")
	case 4:
		s.dbConfirmPrompt = i18n.T("db_confirm_import_hosts")
	case 5:
		s.dbConfirmAction = 0
		s.showRekeyModal = true
		s.initRekeyInputs()
	case 6:
		s.dbConfirmPrompt = i18n.T("db_confirm_reset")
	}
}

func (s *SettingsModal) executeDatabaseAction(action int) (tea.Cmd, bool) {
	if s.store == nil {
		s.errMessage = "Database connection unavailable"
		return nil, true
	}

	switch action {
	case 1: // Prune & Vacuum
		days := RetentionDays[s.retentionIndex]
		deleted, before, after, err := s.store.PruneAndVacuum(days)
		if err != nil {
			s.errMessage = err.Error()
			s.successMessage = ""
			return nil, true
		}
		s.dbStats, _ = s.store.GetDBStats()
		bStr := fmt.Sprintf("%.1f KB", float64(before)/1024.0)
		aStr := fmt.Sprintf("%.1f KB", float64(after)/1024.0)
		s.errMessage = ""
		s.successMessage = fmt.Sprintf(i18n.T("db_success_prune"), deleted, bStr, aStr)
		return nil, true

	case 2, 3: // Export Metrics CSV (2) or Hosts JSON (3)
		s.pendingExportType = action - 1
		initDir := s.getSafeInitialDir()
		s.filePicker = NewDirPickerModal(initDir, 80, 24)
		return nil, true

	case 4: // Import Hosts JSON
		s.pendingExportType = 3
		initDir := s.getSafeInitialDir()
		s.filePicker = NewFilePickerModal(initDir, 80, 24)
		return nil, true

	case 6: // Factory Reset
		s.successMessage = "⚠️ Factory Reset scheduled (skipped in development)"
		return nil, true
	}

	return nil, true
}

func (s *SettingsModal) getSafeInitialDir() string {
	initDir := "."
	if len(s.logDirPresets) > 0 && s.logDirPresets[0].Path != "" {
		initDir = s.logDirPresets[0].Path
	}
	_ = os.MkdirAll(initDir, 0755)
	if _, err := os.Stat(initDir); os.IsNotExist(err) {
		if home, hErr := os.UserHomeDir(); hErr == nil {
			initDir = home
		}
	}
	return initDir
}

func (s *SettingsModal) handleDatabaseFilePicked(pickedPath string) {
	if s.store == nil {
		s.errMessage = "Database connection unavailable"
		return
	}

	switch s.pendingExportType {
	case 1: // Export Metrics CSV
		ts := time.Now().Format("20060102_150405")
		targetFile := filepath.Join(pickedPath, fmt.Sprintf("leitstand_metrics_%s.csv", ts))
		days := RetentionDays[s.retentionIndex]
		count, err := s.store.ExportMetricsCSV(targetFile, days)
		if err != nil {
			s.errMessage = err.Error()
			s.successMessage = ""
			return
		}
		s.errMessage = ""
		s.successMessage = fmt.Sprintf(i18n.T("db_success_export_metrics"), count, filepath.Base(targetFile))

	case 2: // Export Hosts JSON
		ts := time.Now().Format("20060102_150405")
		targetFile := filepath.Join(pickedPath, fmt.Sprintf("leitstand_hosts_%s.json", ts))
		count, err := s.store.ExportHostsJSON(targetFile)
		if err != nil {
			s.errMessage = err.Error()
			s.successMessage = ""
			return
		}
		s.errMessage = ""
		s.successMessage = fmt.Sprintf(i18n.T("db_success_export_hosts"), count, filepath.Base(targetFile))

	case 3: // Import Hosts JSON
		imported, skipped, err := s.store.ImportHostsJSON(pickedPath)
		if err != nil {
			s.errMessage = err.Error()
			s.successMessage = ""
			return
		}
		s.dbStats, _ = s.store.GetDBStats()
		s.errMessage = ""
		s.successMessage = fmt.Sprintf(i18n.T("db_success_import_hosts"), imported, skipped)
	}
}
