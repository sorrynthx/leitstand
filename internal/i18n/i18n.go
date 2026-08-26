package i18n

import (
	"fmt"
	"sync"
)

type Lang string

const (
	LangEN Lang = "en"
	LangKO Lang = "ko"
	LangDE Lang = "de"
)

type LangOption struct {
	Code  Lang
	Label string
}

var (
	mu          sync.RWMutex
	currentLang = LangEN // Default to English for global showcase

	SupportedLangs = []LangOption{
		{Code: LangEN, Label: "🇺🇸 English"},
		{Code: LangKO, Label: "🇰🇷 한국어 (Korean)"},
		{Code: LangDE, Label: "🇩🇪 Deutsch (German)"},
	}

	messages = map[Lang]map[string]string{
		LangEN: {
			// App Header & Global
			"app_title":      "⚡ LEITSTAND COCKPIT",
			"demo_mode":      "DEMO MODE",
			"live_engine":    "LIVE ENGINE",
			"guard_title":    "⚠️  Terminal Window Too Small",
			"guard_current":  "Current:",
			"guard_required": "Required (Min):",
			"guard_hint":     "Please enlarge your terminal window for optimal cockpit experience.",
			"btn_save":       "Save",
			"btn_cancel":     "Cancel",
			"btn_delete":     "Delete",
			"btn_reconnect":  "Reconnect",
			"btn_exit":       "Exit",
			"btn_apply":      "Apply",
			"btn_close":      "Close",

			// Status Bar Keys
			"status_focus_console": "[Tab/c] Console",
			"status_reconnect":     "[r] Reconnect",
			"status_scroll":        "[PgUp/Dn] Scroll",
			"status_maximize":      "[Ctrl+O] Maximize",
			"status_add":           "[a] Add Host",
			"status_del":           "[d] Del",
			"status_settings":      "[p] Settings",
			"status_runbook":       "[?] Runbook",
			"status_quit":          "[q] Quit",
			"status_complete":      "[Tab] Complete",
			"status_exit_console":  "[Esc] Exit Console",
			"status_history":       "[↑/↓] History",

			// Host Explorer Pane
			"pane_host_explorer": "HOST EXPLORER",
			"no_hosts":           "No hosts registered.\nPress 'a' to add a host.",
			"host_connecting":    "connecting...",
			"host_offline":       "offline",

			// Telemetry Pane
			"pane_telemetry":   "TELEMETRY DECK",
			"select_host_hint": "Select a host from the left panel to inspect live telemetry.",
			"cpu_usage":        "CPU Usage",
			"mem_usage":        "Memory Usage",
			"disk_usage":       "Disk Usage",
			"network_io":       "Network I/O",
			"cores":            "Cores",
			"uptime":           "Uptime",

			// Console Pane
			"pane_console":          "REMOTE COMMAND CONSOLE",
			"console_no_host":       "No host selected.",
			"console_placeholder":   "Type command (e.g. docker ps, ls -la, [?] for Runbook)...",
			"console_offline_ph":    "🔴 Host offline (VPN required?). Press [r] to connect...",
			"console_connecting_ph": "⏳ Connecting to server... Please wait.",
			"console_cleared":       "✨ Console cleared.",
			"console_mode_hint":     "[Tab] Complete  [?] Runbook  [Esc] Exit  [↑/↓] History  [Ctrl+O] Fullscreen",
			"console_mode_max_hint": "[Tab] Complete  [?] Runbook  [Esc] Exit  [Ctrl+O] Exit Fullscreen",

			// In-app Editor Modal
			"editor_title":     "✏️ REMOTE FILE EDITOR",
			"editor_save_hint": "[Ctrl+S] Save to Server",
			"editor_exit_hint": "[Esc] Cancel / Exit",
			"editor_saving":    "⏳ Saving file to remote server...",
			"editor_save_ok":   "✨ File saved to remote server successfully!",

			// Modals
			"modal_add_server":    "➕ REGISTER NEW SERVER",
			"modal_delete_server": "⚠️ DELETE SERVER CONFIRMATION",
			"modal_delete_warn":   "Are you sure you want to remove host '%s' (%s)?\nAll credentials and metrics will be permanently deleted.",
			"modal_vault_title":   "🔐 MASTER VAULT AUTHENTICATION",
			"modal_vault_init":    "Enter master password to initialize vault:",
			"modal_vault_unlock":  "Enter master password to unlock vault:",
			"modal_lang_select":   "🌐 Select Language: [1] English  [2] 한국어  [3] Deutsch",
			"warn_non_ascii":      "⚠️ Non-ASCII/Korean characters detected. Please switch to English mode [한/영]!",
			"badge_caps_lock":     "⇪ Caps Lock ON",

			// Settings Modal - General Tab
			"settings_title":          "⚙️  PREFERENCES & PROFILE",
			"tab_general":             "[1] General Settings",
			"tab_about":               "[2] About & Creator",
			"settings_tab_hint":       "💡 Switch Tabs: Press [1] for General  |  Press [2] for About Creator  |  [Esc] Close",
			"settings_lang_label":     "Language:",
			"settings_interval_label": "Telemetry Monitor:",
			"settings_curr_pass":      "Current Password:",
			"settings_new_pass":       "New Password (optional):",
			"settings_confirm_pass":   "Confirm New Password:",
			"settings_saved":          "✨ Preferences updated successfully!",
			"settings_pass_mismatch":  "❌ New passwords do not match.",
			"settings_pass_changed":   "✨ Master password updated successfully!",
			"settings_auth_fail":      "❌ Incorrect current password.",
			"settings_btn_save":       "[Enter] Save & Apply",
			"settings_btn_cancel":     "[Esc] Cancel / Close",
			"settings_footer_gen":     "[Tab/↑/↓] Navigate  [◄/►] Change Option",
			"settings_footer_about":   "[1] Switch to General Settings  [Esc] Close",

			// Settings Modal - About Tab
			"about_app_name":      "⚡ LEITSTAND Cockpit v1.0.0",
			"about_tagline":       "Modern, ultra-lightweight & zero-dependency server operations cockpit.",
			"about_creator_title": "👤 Creator Profile",
			"about_creator_name":  "Kyunggon Kim (김경곤)",
			"about_vision_title":  "🎯 Vision & Story",
			"about_vision_desc":   "Engineered to provide an ultra-fast, non-intrusive server operations cockpit without heavy agent daemons. Passionate about systems architecture, cloud infrastructure, and actively seeking engineering opportunities in Germany / Europe.",
			"about_links_title":   "🔗 Connect & Links",

			// Quick Command Runbook Drawer
			"drawer_title":   "📖 COMMAND RUNBOOK & CHEAT SHEET",
			"drawer_hint":    "[↑/↓] Select  [Enter] Insert into Console  [Tab/1-5] Switch OS  [Esc] Close",
			"drawer_auto_os": "🎯 Detected Host OS:",
			"cat_resources":  "📊 Resources & Processes",
			"cat_services":   "⚙️ Services & Daemons",
			"cat_logs":       "📜 Logs & Troubleshooting",
			"cat_network":    "🌐 Network & Ports",
			"cat_disk":       "💾 Disk & Storage",
			"cat_packages":   "📦 Packages & Updates",
			"cat_security":   "🛡️ Security & Firewall",
			"cat_docker":     "🐳 Docker & Containers",

			// Runbook Item Titles & Descriptions
			"cmd_top_cpu_title":        "Top 10 CPU Consuming Processes",
			"cmd_top_cpu_desc":         "Finds active processes ranking by highest CPU consumption.",
			"cmd_top_mem_title":        "Top 10 Memory Consuming Processes",
			"cmd_top_mem_desc":         "Finds active processes ranking by highest RAM memory usage.",
			"cmd_free_title":           "Human-readable RAM & Swap Memory",
			"cmd_free_desc":            "Displays total, used, free, and cached memory statistics.",
			"cmd_ports_title":          "Active Listening Ports (ss)",
			"cmd_ports_desc":           "Lists all TCP/UDP listening ports and bound daemon processes.",
			"cmd_ip_brief_title":       "Network Interfaces & IP Brief",
			"cmd_ip_brief_desc":        "Shows summary of network interfaces with assigned IP addresses.",
			"cmd_df_title":             "Filesystem Disk Space (df)",
			"cmd_df_desc":              "Checks disk space usage with filesystem type breakdown.",
			"cmd_du_top_title":         "Top 10 Largest /var Directories",
			"cmd_du_top_desc":          "Analyzes largest directory disk hogs in /var.",
			"cmd_dmesg_title":          "Kernel Hardware & Driver Errors",
			"cmd_dmesg_desc":           "Inspects kernel ring buffer for hardware warnings and crashes.",
			"cmd_failed_svc_title":     "Failed Systemd Services",
			"cmd_failed_svc_desc":      "Lists all failed systemd units requiring immediate triage.",
			"cmd_active_svc_title":     "Running Systemd Services",
			"cmd_active_svc_desc":      "Lists all actively running systemd background services.",
			"cmd_journal_err_title":    "Recent Critical System Errors",
			"cmd_journal_err_desc":     "Filters journalctl for priority 3 (errors) in the last 50 lines.",
			"cmd_syslog_tail_title":    "Tail 50 Lines of System Syslog",
			"cmd_syslog_tail_desc":     "Reads latest entries from the standard /var/log/syslog file.",
			"cmd_ufw_title":            "UFW Firewall Rules & Status",
			"cmd_ufw_desc":             "Inspects Ubuntu Uncomplicated Firewall active policy.",
			"cmd_apt_upgradable_title": "List Upgradable APT Packages",
			"cmd_apt_upgradable_desc":  "Checks available software and security updates.",
			"cmd_reboot_req_title":     "Check if System Reboot is Required",
			"cmd_reboot_req_desc":      "Checks if kernel upgrade requires a node reboot.",
			"cmd_firewalld_title":      "Firewalld Active Rules (RHEL)",
			"cmd_firewalld_desc":       "Inspects active zones and open ports in firewalld.",
			"cmd_selinux_title":        "SELinux Enforcement Status",
			"cmd_selinux_desc":         "Displays SELinux operational mode (Enforcing/Permissive).",
			"cmd_dnf_check_title":      "Check Available DNF/RPM Updates",
			"cmd_dnf_check_desc":       "Scans repositories for pending package updates.",
			"cmd_rc_status_title":      "OpenRC Service Status (Alpine)",
			"cmd_rc_status_desc":       "Lists running and stopped services in OpenRC runlevels.",
			"cmd_rc_list_title":        "List All Available OpenRC Services",
			"cmd_rc_list_desc":         "Shows all installed service scripts on Alpine.",
			"cmd_apk_upgradable_title": "Check Outdated APK Packages",
			"cmd_apk_upgradable_desc":  "Compares installed packages with latest Alpine repos.",
			"cmd_docker_ps_title":      "List All Docker Containers",
			"cmd_docker_ps_desc":       "Displays status, names, and port mappings of all containers.",
			"cmd_docker_stats_title":   "Live Container Resource Usage",
			"cmd_docker_stats_desc":    "Snapshot of CPU, Memory, Network, and Block I/O per container.",
			"cmd_docker_df_title":      "Docker Storage Usage Breakdown",
			"cmd_docker_df_desc":       "Reclaims insight into images, containers, and volume disk space.",
			"cmd_docker_compose_title": "Docker Compose Services Status",
			"cmd_docker_compose_desc":  "Inspects containers declared in current directory compose stack.",
			"cmd_docker_logs_title":    "Tail Latest 50 Lines of Container Logs",
			"cmd_docker_logs_desc":     "Fetches trailing log lines from the most recent container.",
		},

		LangKO: {
			// App Header & Global
			"app_title":      "⚡ LEITSTAND COCKPIT",
			"demo_mode":      "데모 모드",
			"live_engine":    "라이브 엔진",
			"guard_title":    "⚠️  터미널 창 크기가 너무 작습니다",
			"guard_current":  "현재 크기:",
			"guard_required": "최소 권장 크기:",
			"guard_hint":     "최적의 콕핏 대시보드 화면을 위해 터미널 창을 조금 더 넓혀주세요.",
			"btn_save":       "저장",
			"btn_cancel":     "취소",
			"btn_delete":     "삭제",
			"btn_reconnect":  "재연결",
			"btn_exit":       "종료",
			"btn_apply":      "적용",
			"btn_close":      "닫기",

			// Status Bar Keys
			"status_focus_console": "[Tab/c] 콘솔 포커스",
			"status_reconnect":     "[r] 재연결",
			"status_scroll":        "[PgUp/Dn] 스크롤",
			"status_maximize":      "[Ctrl+O] 최대화",
			"status_add":           "[a] 서버 추가",
			"status_del":           "[d] 삭제",
			"status_settings":      "[p] 설정",
			"status_runbook":       "[?] 런북사전",
			"status_quit":          "[q] 종료",
			"status_complete":      "[Tab] 자동완성",
			"status_exit_console":  "[Esc] 콘솔 나가기",
			"status_history":       "[↑/↓] 히스토리",

			// Host Explorer Pane
			"pane_host_explorer": "서버 목록",
			"no_hosts":           "등록된 서버가 없습니다.\n'a' 키를 눌러 새 서버를 추가하세요.",
			"host_connecting":    "연결 중...",
			"host_offline":       "오프라인",

			// Telemetry Pane
			"pane_telemetry":   "실시간 텔레메트리",
			"select_host_hint": "좌측 목록에서 서버를 선택하면 텔레메트리를 모니터링합니다.",
			"cpu_usage":        "CPU 사용률",
			"mem_usage":        "메모리 사용률",
			"disk_usage":       "디스크 사용량",
			"network_io":       "네트워크 I/O",
			"cores":            "코어",
			"uptime":           "가동시간",

			// Console Pane
			"pane_console":          "원격 제어 콘솔",
			"console_no_host":       "선택된 서버가 없습니다.",
			"console_placeholder":   "명령어를 입력하세요 (예: docker ps, ls -la, [?] 런북사전)...",
			"console_offline_ph":    "🔴 서버가 오프라인입니다 (VPN 확인 필요). [r]을 눌러 연결하세요...",
			"console_connecting_ph": "⏳ 서버에 연결 중입니다... 잠시만 기다려주세요.",
			"console_cleared":       "✨ 콘솔 화면이 초기화되었습니다.",
			"console_mode_hint":     "[Tab] 자동완성  [?] 런북  [Esc] 나가기  [↑/↓] 히스토리  [Ctrl+O] 전체화면",
			"console_mode_max_hint": "[Tab] 자동완성  [?] 런북  [Esc] 나가기  [Ctrl+O] 분할화면 복귀",

			// In-app Editor Modal
			"editor_title":     "✏️ 원격 파일 편집기",
			"editor_save_hint": "[Ctrl+S] 서버에 저장",
			"editor_exit_hint": "[Esc] 취소 / 닫기",
			"editor_saving":    "⏳ 서버에 파일을 저장하는 중입니다...",
			"editor_save_ok":   "✨ 원격 서버에 파일이 성공적으로 저장되었습니다!",

			// Modals
			"modal_add_server":    "➕ 새 원격 서버 등록",
			"modal_delete_server": "⚠️ 서버 삭제 확인",
			"modal_delete_warn":   "서버 '%s' (%s)를 정말 삭제하시겠습니까?\n저장된 암호 및 텔레메트리 기록이 영구적으로 삭제됩니다.",
			"modal_vault_title":   "🔐 마스터 금고(Vault) 인증",
			"modal_vault_init":    "마스터 금고 초기화 비밀번호를 입력하세요:",
			"modal_vault_unlock":  "마스터 금고 잠금 해제 비밀번호를 입력하세요:",
			"modal_lang_select":   "🌐 언어 선택: [1] English  [2] 한국어  [3] Deutsch",
			"warn_non_ascii":      "⚠️ 한글 입력 감지됨! [한/영] 키를 눌러 영문으로 입력해 주세요.",
			"badge_caps_lock":     "⇪ Caps Lock 켜짐",

			// Settings Modal - General Tab
			"settings_title":          "⚙️  환경설정 & 개발자 프로필",
			"tab_general":             "[1] 일반 설정",
			"tab_about":               "[2] 개발자 소개 & 비전",
			"settings_tab_hint":       "💡 탭 전환: [1] 일반 설정  |  [2] 개발자 소개  |  [Esc] 닫기",
			"settings_lang_label":     "표시 언어:",
			"settings_interval_label": "서버 상태 모니터링:",
			"settings_curr_pass":      "현재 마스터 비밀번호:",
			"settings_new_pass":       "새 마스터 비밀번호 (선택):",
			"settings_confirm_pass":   "새 비밀번호 확인:",
			"settings_saved":          "✨ 설정이 성공적으로 저장되었습니다!",
			"settings_pass_mismatch":  "❌ 새 비밀번호가 일치하지 않습니다.",
			"settings_pass_changed":   "✨ 마스터 비밀번호가 성공적으로 변경되었습니다!",
			"settings_auth_fail":      "❌ 현재 비밀번호가 올바르지 않습니다.",
			"settings_btn_save":       "[Enter] 저장 및 적용",
			"settings_btn_cancel":     "[Esc] 취소 / 닫기",
			"settings_footer_gen":     "[Tab/↑/↓] 항목 이동  [◄/►] 옵션 변경",
			"settings_footer_about":   "[1] 일반 설정으로 이동  [Esc] 닫기",

			// Settings Modal - About Tab
			"about_app_name":      "⚡ LEITSTAND Cockpit v1.0.0",
			"about_tagline":       "무거운 데몬 없이 빠르고 직관적인 차세대 서버 관제/원격 제어 콕핏",
			"about_creator_title": "👤 개발자 소개",
			"about_creator_name":  "김경곤 (Kyunggon Kim)",
			"about_vision_title":  "🎯 개발 배경 & 비전",
			"about_vision_desc":   "복잡하고 무거운 모니터링 에이전트 없이도 즉각적으로 서버를 제어하고 모니터링할 수 있는 초경량 도구를 만들었습니다. 백엔드 시스템 및 인프라 자동화에 깊은 열정을 품고 있으며, 독일 및 글로벌 테크 기업에서의 소프트웨어 엔지니어링 기회를 적극적으로 모색하고 있습니다.",
			"about_links_title":   "🔗 연락처 & 소셜 링크",

			// Quick Command Runbook Drawer
			"drawer_title":   "📖 실무 핵심 명령어 런북 & 사전",
			"drawer_hint":    "[↑/↓] 명령어 선택  [Enter] 콘솔에 입력  [Tab/1-5] OS 전환  [Esc] 닫기",
			"drawer_auto_os": "🎯 감지된 서버 OS:",
			"cat_resources":  "📊 자원 점유 & 프로세스",
			"cat_services":   "⚙️ 시스템 서비스 (데몬)",
			"cat_logs":       "📜 로그 & 긴급 트러블슈팅",
			"cat_network":    "🌐 네트워크 & 열린 포트",
			"cat_disk":       "💾 디스크 & 대용량 파일",
			"cat_packages":   "📦 패키지 & 보안 업데이트",
			"cat_security":   "🛡️ 방화벽 & 보안 상태",
			"cat_docker":     "🐳 도커 & 컨테이너 관리",

			// Runbook Item Titles & Descriptions
			"cmd_top_cpu_title":        "CPU 점유율 Top 10 프로세스",
			"cmd_top_cpu_desc":         "현재 서버에서 CPU를 가장 많이 소모하는 프로세스 10개를 조회합니다.",
			"cmd_top_mem_title":        "메모리 점유율 Top 10 프로세스",
			"cmd_top_mem_desc":         "RAM 메모리를 가장 많이 점유 중인 프로세스 10개를 정렬합니다.",
			"cmd_free_title":           "실시간 RAM 및 Swap 메모리 현황",
			"cmd_free_desc":            "전체 메모리, 사용 중인 메모리, 버퍼/캐시 현황을 읽기 쉽게 표시합니다.",
			"cmd_ports_title":          "현재 열려있는 리스닝 포트 (ss)",
			"cmd_ports_desc":           "외부 연결을 수신 중인 TCP/UDP 포트와 해당 프로세스를 확인합니다.",
			"cmd_ip_brief_title":       "네트워크 인터페이스 및 IP 요약",
			"cmd_ip_brief_desc":        "서버의 네트워크 카드 목록과 할당된 IP를 한눈에 요약합니다.",
			"cmd_df_title":             "파일시스템별 디스크 잔여 용량",
			"cmd_df_desc":              "마운트된 디스크 파티션별 사용량 및 파일시스템 종류를 확인합니다.",
			"cmd_du_top_title":         "/var 디렉토리 용량 Top 10 폴더",
			"cmd_du_top_desc":          "로그나 캐시 등으로 용량을 많이 차지하는 대용량 폴더를 찾습니다.",
			"cmd_dmesg_title":          "커널 하드웨어 및 드라이버 오류",
			"cmd_dmesg_desc":           "커널 링 버퍼에서 하드웨어 오류 및 시스템 경고 로그를 점검합니다.",
			"cmd_failed_svc_title":     "비정상 종료된(Failed) 서비스 확인",
			"cmd_failed_svc_desc":      "systemd에서 오류로 멈추거나 실패한 서비스를 즉시 색출합니다.",
			"cmd_active_svc_title":     "현재 정상 가동 중인 서비스 목록",
			"cmd_active_svc_desc":      "활성화되어 실행 중인 모든 시스템 서비스를 조회합니다.",
			"cmd_journal_err_title":    "최근 시스템 에러 로그 (50줄)",
			"cmd_journal_err_desc":     "journalctl에서 에러 레벨(p 3)의 로그만 골라 50줄을 출력합니다.",
			"cmd_syslog_tail_title":    "표준 시스템 로그 (/var/log/syslog)",
			"cmd_syslog_tail_desc":     "서버의 실시간 시스템 로그 파일 최신 50줄을 확인합니다.",
			"cmd_ufw_title":            "UFW 방화벽 활성화 상태 및 규칙",
			"cmd_ufw_desc":             "우분투 기본 방화벽(UFW)의 허용/차단 포트 규칙을 점검합니다.",
			"cmd_apt_upgradable_title": "업데이트 가능한 APT 패키지 목록",
			"cmd_apt_upgradable_desc":  "설치된 패키지 중 새 버전이 나와있는 패키지를 확인합니다.",
			"cmd_reboot_req_title":     "시스템 재부팅 필요 여부 확인",
			"cmd_reboot_req_desc":      "커널 보안 업데이트 후 서버 재부팅이 필요한지 검사합니다.",
			"cmd_firewalld_title":      "Firewalld 방화벽 활성 규칙 (RHEL)",
			"cmd_firewalld_desc":       "RHEL/CentOS 계열의 방화벽 영역 및 열린 포트를 확인합니다.",
			"cmd_selinux_title":        "SELinux 보안 모드 상태 확인",
			"cmd_selinux_desc":         "SELinux가 Enforcing(차단)인지 Permissive인지 점검합니다.",
			"cmd_dnf_check_title":      "업데이트 가능한 DNF/RPM 패키지",
			"cmd_dnf_check_desc":       "RHEL 저장소에서 최신 패키지 업데이트 유무를 검색합니다.",
			"cmd_rc_status_title":      "OpenRC 서비스 상태 (Alpine)",
			"cmd_rc_status_desc":       "Alpine Linux OpenRC 런레벨의 서비스 상태를 확인합니다.",
			"cmd_rc_list_title":        "설치된 OpenRC 서비스 전체 목록",
			"cmd_rc_list_desc":         "시스템에 등록된 모든 OpenRC 스크립트를 나열합니다.",
			"cmd_apk_upgradable_title": "구버전 APK 패키지 확인",
			"cmd_apk_upgradable_desc":  "Alpine 공식 리포지토리 대비 업데이트 가능한 패키지를 찾습니다.",
			"cmd_docker_ps_title":      "전체 도커 컨테이너 상태 요약",
			"cmd_docker_ps_desc":       "실행 중/중지된 모든 컨테이너의 이름, 상태, 포트를 조회합니다.",
			"cmd_docker_stats_title":   "컨테이너별 실시간 자원 점유율",
			"cmd_docker_stats_desc":    "도커 컨테이너별 CPU, 메모리, 네트워크 트래픽을 스냅샷으로 봅니다.",
			"cmd_docker_df_title":      "도커 디스크 사용량 분석",
			"cmd_docker_df_desc":       "이미지, 볼륨, 빌드 캐시가 차지하는 디스크 공간을 점검합니다.",
			"cmd_docker_compose_title": "도커 컴포즈(Compose) 서비스 상태",
			"cmd_docker_compose_desc":  "현재 폴더의 docker-compose 서비스 구동 현황을 확인합니다.",
			"cmd_docker_logs_title":    "최근 컨테이너 로그 50줄 출력",
			"cmd_docker_logs_desc":     "가장 최근 실행된 컨테이너의 최신 로그 50줄을 가져옵니다.",
		},

		LangDE: {
			// App Header & Global
			"app_title":      "⚡ LEITSTAND COCKPIT",
			"demo_mode":      "DEMO-MODUS",
			"live_engine":    "LIVE-ENGINE",
			"guard_title":    "⚠️  Terminal-Fenster zu klein",
			"guard_current":  "Aktuell:",
			"guard_required": "Erforderlich (Min):",
			"guard_hint":     "Bitte vergrößern Sie Ihr Terminal-Fenster für die optimale Cockpit-Nutzung.",
			"btn_save":       "Speichern",
			"btn_cancel":     "Abbrechen",
			"btn_delete":     "Löschen",
			"btn_reconnect":  "Wiederverbinden",
			"btn_exit":       "Beenden",
			"btn_apply":      "Anwenden",
			"btn_close":      "Schließen",

			// Status Bar Keys
			"status_focus_console": "[Tab/c] Konsole",
			"status_reconnect":     "[r] Neu verbinden",
			"status_scroll":        "[PgUp/Dn] Scrollen",
			"status_maximize":      "[Ctrl+O] Maximieren",
			"status_add":           "[a] Server hinzufügen",
			"status_del":           "[d] Löschen",
			"status_settings":      "[p] Einstellungen",
			"status_runbook":       "[?] Runbook",
			"status_quit":          "[q] Beenden",
			"status_complete":      "[Tab] Autovervollständigung",
			"status_exit_console":  "[Esc] Konsole verlassen",
			"status_history":       "[↑/↓] Verlauf",

			// Host Explorer Pane
			"pane_host_explorer": "HOST-EXPLORER",
			"no_hosts":           "Keine Hosts registriert.\nDrücken Sie 'a', um einen Server hinzuzufügen.",
			"host_connecting":    "verbinde...",
			"host_offline":       "offline",

			// Telemetry Pane
			"pane_telemetry":   "TELEMETRIE-DECK",
			"select_host_hint": "Wählen Sie einen Host links aus, um Telemetriedaten anzuzeigen.",
			"cpu_usage":        "CPU-Auslastung",
			"mem_usage":        "Speicherauslastung",
			"disk_usage":       "Festplattennutzung",
			"network_io":       "Netzwerk-I/O",
			"cores":            "Kerne",
			"uptime":           "Betriebszeit",

			// Console Pane
			"pane_console":          "REMOTE-BEFEHLSKONSOLE",
			"console_no_host":       "Kein Host ausgewählt.",
			"console_placeholder":   "Befehl eingeben (z. B. docker ps, ls -la, [?] für Runbook)...",
			"console_offline_ph":    "🔴 Host offline (VPN erforderlich?). [r] zum Verbinden...",
			"console_connecting_ph": "⏳ Verbindung wird hergestellt... Bitte warten.",
			"console_cleared":       "✨ Konsole geleert.",
			"console_mode_hint":     "[Tab] Vervollst.  [?] Runbook  [Esc] Beenden  [↑/↓] Verlauf  [Ctrl+O] Vollbild",
			"console_mode_max_hint": "[Tab] Vervollst.  [?] Runbook  [Esc] Beenden  [Ctrl+O] Geteilt",

			// In-app Editor Modal
			"editor_title":     "✏️ REMOTE-DATEI-EDITOR",
			"editor_save_hint": "[Ctrl+S] Auf Server speichern",
			"editor_exit_hint": "[Esc] Abbrechen / Beenden",
			"editor_saving":    "⏳ Datei wird auf Server gespeichert...",
			"editor_save_ok":   "✨ Datei erfolgreich auf dem Server gespeichert!",

			// Modals
			"modal_add_server":    "➕ NEUEN SERVER REGISTRIEREN",
			"modal_delete_server": "⚠️ SERVER-LÖSCHUNG BESTÄTIGEN",
			"modal_delete_warn":   "Möchten Sie Host '%s' (%s) wirklich entfernen?\nAlle Anmeldedaten und Metriken werden dauerhaft gelöscht.",
			"modal_vault_title":   "🔐 MASTER-VAULT-AUTHENTIFIZIERUNG",
			"modal_vault_init":    "Master-Passwort zur Vault-Initialisierung eingeben:",
			"modal_vault_unlock":  "Master-Passwort zum Entsperren eingeben:",
			"modal_lang_select":   "🌐 Sprache wählen: [1] English  [2] 한국어  [3] Deutsch",
			"warn_non_ascii":      "⚠️ Nicht-ASCII/Koreanische Zeichen erkannt. Bitte [한/영] auf Englisch umstellen!",
			"badge_caps_lock":     "⇪ Caps Lock AN",

			// Settings Modal - General Tab
			"settings_title":          "⚙️  EINSTELLUNGEN & PROFIL",
			"tab_general":             "[1] Allgemein",
			"tab_about":               "[2] Über den Entwickler",
			"settings_tab_hint":       "💡 Tab wechseln: [1] Allgemein  |  [2] Über den Entwickler  |  [Esc] Schließen",
			"settings_lang_label":     "Sprache:",
			"settings_interval_label": "Serverüberwachung:",
			"settings_curr_pass":      "Aktuelles Passwort:",
			"settings_new_pass":       "Neues Passwort (optional):",
			"settings_confirm_pass":   "Neues Passwort bestätigen:",
			"settings_saved":          "✨ Einstellungen erfolgreich aktualisiert!",
			"settings_pass_mismatch":  "❌ Neue Passwörter stimmen nicht überein.",
			"settings_pass_changed":   "✨ Master-Passwort erfolgreich aktualisiert!",
			"settings_auth_fail":      "❌ Aktuelles Passwort ist falsch.",
			"settings_btn_save":       "[Enter] Speichern & Anwenden",
			"settings_btn_cancel":     "[Esc] Schließen",
			"settings_footer_gen":     "[Tab/↑/↓] Navigieren  [◄/►] Option ändern",
			"settings_footer_about":   "[1] Zu Allgemein wechseln  [Esc] Schließen",

			// Settings Modal - About Tab
			"about_app_name":      "⚡ LEITSTAND Cockpit v1.0.0",
			"about_tagline":       "Modernes, extrem leichtgewichtiges & agentenloses Server-Cockpit.",
			"about_creator_title": "👤 Entwicklerprofil",
			"about_creator_name":  "Kyunggon Kim (김경곤)",
			"about_vision_title":  "🎯 Vision & Motivation",
			"about_vision_desc":   "Entwickelt für eine nahtlose, nicht-invasive Serververwaltung ohne schwere Daemons. Leidenschaft für Systemarchitektur, Cloud-Infrastruktur und aktiv auf der Suche nach Software-Engineering-Möglichkeiten in Deutschland / Europa.",
			"about_links_title":   "🔗 Kontakt & Netzwerke",

			// Quick Command Runbook Drawer
			"drawer_title":   "📖 BEFEHLS-RUNBOOK & CHEAT SHEET",
			"drawer_hint":    "[↑/↓] Befehl wählen  [Enter] In Konsole einfügen  [Tab/1-5] OS wechseln  [Esc] Schließen",
			"drawer_auto_os": "🎯 Erkanntes Server-Betriebssystem:",
			"cat_resources":  "📊 Ressourcen & Prozesse",
			"cat_services":   "⚙️ Systemdienste",
			"cat_logs":       "📜 Protokolle & Fehlerbehebung",
			"cat_network":    "🌐 Netzwerk & Ports",
			"cat_disk":       "💾 Festplatte & Speicher",
			"cat_packages":   "📦 Pakete & Updates",
			"cat_security":   "🛡️ Sicherheit & Firewall",
			"cat_docker":     "🐳 Docker & Container",

			// Runbook Item Titles & Descriptions
			"cmd_top_cpu_title":        "Top 10 CPU-Prozesse",
			"cmd_top_cpu_desc":         "Findet aktive Prozesse mit der höchsten CPU-Auslastung.",
			"cmd_top_mem_title":        "Top 10 Speicher-Prozesse",
			"cmd_top_mem_desc":         "Sortiert Prozesse nach der höchsten RAM-Nutzung.",
			"cmd_free_title":           "RAM & Swap-Speicherübersicht",
			"cmd_free_desc":            "Zeigt Gesamt-, belegten und Cache-Speicher lesbar an.",
			"cmd_ports_title":          "Aktive Listening-Ports (ss)",
			"cmd_ports_desc":           "Listet alle lauschenden TCP/UDP-Ports und Prozesse auf.",
			"cmd_ip_brief_title":       "Netzwerkschnittstellen & IPs",
			"cmd_ip_brief_desc":        "Zeigt Zusammenfassung aller Schnittstellen und IP-Adressen.",
			"cmd_df_title":             "Festplattennutzung (df)",
			"cmd_df_desc":              "Überprüft den freien Speicherplatz nach Dateisystemtyp.",
			"cmd_du_top_title":         "Größte /var-Verzeichnisse",
			"cmd_du_top_desc":          "Findet speicherintensive Verzeichnisse unter /var.",
			"cmd_dmesg_title":          "Kernel- & Treiberfehler",
			"cmd_dmesg_desc":           "Prüft den Kernel-Ringpuffer auf Hardwarewarnungen.",
			"cmd_failed_svc_title":     "Fehlgeschlagene Systemd-Dienste",
			"cmd_failed_svc_desc":      "Listet alle fehlerhaften Systemd-Einheiten auf.",
			"cmd_active_svc_title":     "Laufende Systemd-Dienste",
			"cmd_active_svc_desc":      "Zeigt alle aktiv ausgeführten Hintergrunddienste an.",
			"cmd_journal_err_title":    "Kritische Systemfehler (50 Zeilen)",
			"cmd_journal_err_desc":     "Filtert journalctl nach Priorität 3 (Fehler).",
			"cmd_syslog_tail_title":    "Neueste Syslog-Zeilen (/var/log/syslog)",
			"cmd_syslog_tail_desc":     "Liest die neuesten 50 Zeilen des Systemprotokolls.",
			"cmd_ufw_title":            "UFW-Firewall-Status (Ubuntu)",
			"cmd_ufw_desc":             "Prüft aktive Regeln der Uncomplicated Firewall.",
			"cmd_apt_upgradable_title": "Aktualisierbare APT-Pakete",
			"cmd_apt_upgradable_desc":  "Überprüft verfügbare Software- und Sicherheitsupdates.",
			"cmd_reboot_req_title":     "Systemneustart erforderlich?",
			"cmd_reboot_req_desc":      "Prüft, ob ein Kernel-Update einen Neustart verlangt.",
			"cmd_firewalld_title":      "Firewalld-Regeln (RHEL)",
			"cmd_firewalld_desc":       "Überprüft aktive Zonen und offene Ports in firewalld.",
			"cmd_selinux_title":        "SELinux-Sicherheitsstatus",
			"cmd_selinux_desc":         "Zeigt den SELinux-Status (Enforcing/Permissive) an.",
			"cmd_dnf_check_title":      "Verfügbare DNF/RPM-Updates",
			"cmd_dnf_check_desc":       "Durchsucht Repositories nach ausstehenden Paket-Updates.",
			"cmd_rc_status_title":      "OpenRC-Dienststatus (Alpine)",
			"cmd_rc_status_desc":       "Listet Dienste in OpenRC-Runlevels auf.",
			"cmd_rc_list_title":        "Alle OpenRC-Dienste auflisten",
			"cmd_rc_list_desc":         "Zeigt alle installierten OpenRC-Skripte an.",
			"cmd_apk_upgradable_title": "Veraltete APK-Pakete prüfen",
			"cmd_apk_upgradable_desc":  "Vergleicht installierte Pakete mit Alpine-Repositories.",
			"cmd_docker_ps_title":      "Alle Docker-Container auflisten",
			"cmd_docker_ps_desc":       "Zeigt Status, Namen und Port-Mappings aller Container.",
			"cmd_docker_stats_title":   "Echtzeit-Ressourcennutzung der Container",
			"cmd_docker_stats_desc":    "Snapshot von CPU, Speicher und Netzwerk pro Container.",
			"cmd_docker_df_title":      "Docker-Speicheranalyse",
			"cmd_docker_df_desc":       "Zeigt Speicherbelegung von Images, Containern und Volumes.",
			"cmd_docker_compose_title": "Docker Compose Dienststatus",
			"cmd_docker_compose_desc":  "Prüft Compose-Container im aktuellen Verzeichnis.",
			"cmd_docker_logs_title":    "Neueste 50 Zeilen Container-Logs",
			"cmd_docker_logs_desc":     "Ruft aktuelle Protokolle des neuesten Containers ab.",
		},
	}
)

// SetLang switches the active language.
func SetLang(l Lang) {
	mu.Lock()
	defer mu.Unlock()
	if _, exists := messages[l]; exists {
		currentLang = l
	}
}

// GetLang returns the current active language.
func GetLang() Lang {
	mu.RLock()
	defer mu.RUnlock()
	return currentLang
}

// T translates a key into the current language, with fallback to English and then raw key.
func T(key string) string {
	mu.RLock()
	defer mu.RUnlock()

	if dict, exists := messages[currentLang]; exists {
		if val, found := dict[key]; found {
			return val
		}
	}

	// Fallback to English
	if dict, exists := messages[LangEN]; exists {
		if val, found := dict[key]; found {
			return val
		}
	}

	return key
}

// Tf formats a translated string.
func Tf(key string, args ...interface{}) string {
	return fmt.Sprintf(T(key), args...)
}
