package tui

import (
	"fmt"
	"leitstand/internal/i18n"
	"strconv"
)

func (m *Model) applyAndPersistSettings(res SettingsResult) {
	i18n.SetLang(res.Lang)
	if m.cfg != nil {
		m.cfg.Telemetry.PollingInterval = res.Interval
		m.cfg.Telemetry.CPUThreshold = res.CPUThresh
		m.cfg.Telemetry.RAMThreshold = res.RAMThresh
		m.cfg.Telemetry.DiskThreshold = res.DiskThresh
		if res.LogDir != "" {
			m.cfg.Logging.SessionLogDir = res.LogDir
		}
		if res.AIProvider != "" {
			m.cfg.AI.Provider = res.AIProvider
		}
		if res.AIEndpoint != "" {
			m.cfg.AI.Endpoint = res.AIEndpoint
		}
		if res.AIModel != "" {
			m.cfg.AI.Model = res.AIModel
		}
		if res.AIRetention > 0 {
			m.cfg.AI.RetentionDays = res.AIRetention
		}
		if res.AIMaxHistory > 0 {
			m.cfg.AI.MaxHistory = res.AIMaxHistory
		}
	}
	if m.store != nil {
		_ = m.store.SetSetting("language", string(res.Lang))
		_ = m.store.SetSetting("polling_interval", res.Interval.String())
		_ = m.store.SetSetting("cpu_threshold", fmt.Sprintf("%.0f", res.CPUThresh))
		_ = m.store.SetSetting("ram_threshold", fmt.Sprintf("%.0f", res.RAMThresh))
		_ = m.store.SetSetting("disk_threshold", fmt.Sprintf("%.0f", res.DiskThresh))
		if res.LogDir != "" {
			_ = m.store.SetSetting("session_log_dir", res.LogDir)
		}
		if res.AIProvider != "" {
			_ = m.store.SetSetting("ai_provider", res.AIProvider)
		}
		if res.AIEndpoint != "" {
			_ = m.store.SetSetting("ai_endpoint", res.AIEndpoint)
		}
		if res.AIKey != "" {
			_ = m.store.SetSetting("ai_api_key", res.AIKey)
		}
		if res.AIModel != "" {
			_ = m.store.SetSetting("ai_model", res.AIModel)
		}
		if res.AIRetention > 0 {
			_ = m.store.SetSetting("ai_retention_days", strconv.Itoa(res.AIRetention))
		}
		if res.AIMaxHistory > 0 {
			_ = m.store.SetSetting("ai_max_history", strconv.Itoa(res.AIMaxHistory))
		}
	}
}
