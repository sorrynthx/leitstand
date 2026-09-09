# Leitstand System Architecture

`Leitstand` (라이트슈탄트) is a modern, lightweight, agentless server control center, telemetry cockpit, and AI troubleshooting assistant built in Go using the Charmbracelet Bubbletea & Lipgloss ecosystem.

---

## 🏛️ Architecture Overview

```text
┌──────────────────────────────────────────────────────────────────────────────────────────────────┐
│                                     Leitstand TUI Frontend                                       │
│          (Bubbletea Model-View-Update + Lipgloss Styling + Multi-Tab Shell Engine)               │
├──────────────────────┬───────────────────────────────┬───────────────────────────────────────────┤
│    Host Explorer     │        Telemetry Deck         │          Remote Command Console           │
│    (Server List)     │    (CPU/RAM/Disk/Net Gauges)  │   (Multi-Tab, Live Stream, Root Elevation)│
├──────────────────────┴───────────────────────────────┴───────────────────────────────────────────┤
│                                 Overlay & Specialized Modules                                    │
│  • SFTP Dual-Pane File Manager & In-App Text Editor ([f] / [F6])                                 │
│  • AI Terminal Copilot & Autonomous Diagnostics ([F4]) (Groq / Ollama / OpenAI)                  │
│  • SSH Port Forwarding & Tunnel Manager ([T] / [F7])                                             │
│  • OS-Aware Runbook & Team Custom Commands ([?] / [7] ⭐️)                                        │
│  • Local Master Vault & Database Maintenance Modal ([p])                                         │
├──────────────────────────────────────────────────────────────────────────────────────────────────┤
│                                    Core Application Layer                                        │
├───────────────────────────────┬──────────────────────────────────┬───────────────────────────────┤
│       SSH Client Pool         │       Telemetry Collector        │       AI Copilot Engine       │
│  (TCP Socket / Host Muxing)   │ (Agentless /proc Delta Polling)  │ (Streaming SSE / Safety Guard)│
├───────────────────────────────┼──────────────────────────────────┼───────────────────────────────┤
│      Master Vault Engine      │        SSH Tunnel Manager        │     Demo Simulator Layer      │
│ (Argon2id KDF + AES-256-GCM)  │ (Local Port ➔ Remote SSH Socket) │  (Realistic Mock & Showcase)  │
├───────────────────────────────┴──────────────────────────────────┴───────────────────────────────┤
│                                 Local SQLite Storage & Config                                    │
│       (Hosts, Encrypted Vault, Telemetry Snapshots, Custom Commands, Tunnels, AI Chats)          │
└──────────────────────────────────────────────────────────────────────────────────────────────────┘
```

---

## 📦 Package Structure

```text
leitstand/
├── cmd/
│   ├── leitstand/          # Main CLI entrypoint (root command, flags, --demo)
│   └── mockssh/            # Mock SSH testing server
├── docs/                   # Architecture & design documentation
├── internal/
│   ├── ai/                 # AI Terminal Copilot (Groq, Ollama, OpenAI clients, SSE streaming, safety guard)
│   ├── app/                # Application bootstrapper and dependency coordinator
│   ├── config/             # YAML configuration parser and runtime overrides
│   ├── highlight/          # Syntax highlighting for JSON, YAML, SQL, Shell, Markdown
│   ├── i18n/               # Tri-lingual localization engine (EN, KO, DE) with 100% dictionary parity
│   ├── logger/             # Structured rotation file logger
│   ├── profile/            # Obfuscated AES-256-GCM encrypted creator profile (Easter egg)
│   ├── quickcmd/           # OS-specific runbooks (Ubuntu, RHEL, Alpine, Docker, Common) & 7-category troubleshooting catalog
│   ├── sessionlog/         # Plain-text ANSI clean session export pipeline (Ctrl+E)
│   ├── ssh/                # SSH connection pool, channel multiplexer, interactive PTY, SFTP client cache, local port forwarder
│   ├── storage/            # Pure-Go SQLite persistence, migrations, encryption, audit export, and maintenance
│   ├── telemetry/          # Agentless Linux /proc delta collector (CPU, Mem, Disk, Net)
│   ├── tui/                # Presentation layer (Bubbletea models, views, update event dispatchers, modals)
│   └── vault/              # AES-256-GCM encrypted credential vault with Argon2id key derivation
└── tests/                  # Integration & end-to-end test suites
```

### Key Modules Breakdown

- **`internal/ai`**: 
  - `client.go`, `groq.go`, `ollama.go`, `openai.go`: Pluggable LLM backends with streaming responses.
  - `ai_safety.go`: Proactive command interception blocking destructive calls (`rm -rf /`, `shutdown`, `reboot`, `swapoff`).
- **`internal/ssh`**:
  - `pool.go`: Reusable SSH client pool (1 TCP socket per host, multiplexed channels).
  - `client.go`: Thread-safe `GetSFTPClient()` caching to eliminate OpenSSH `MaxSessions 10` exhaustion.
  - `tunnel.go`, `tunnel_manager.go`: Local port forwarding via SSH channel dialing (`net.Listener` ➔ `ssh.Channel`).
- **`internal/storage`**:
  - `database.go`: Modern SQLite database with automatic table migrations.
  - `maintenance.go`: Database health statistics, `PruneAndVacuum`, CSV metrics export, JSON hosts backup/restore, and complete `FactoryReset`.
  - `vault.go`: Master password verification sentinel and Argon2id salt storage.
  - `custom_cmd.go`, `tunnels.go`, `ai_chats.go`: Repositories for runbooks, tunnels, and AI chat histories.
- **`internal/tui`**:
  - **Main Cockpit**: `view.go`, `view_hostlist.go`, `view_console_pane.go`, `view_telemetry_deck.go`.
  - **Multi-Tab Shell**: `tab.go`, `tab_ops.go`, `console_runner_*.go` with 100ms Braille spinner (`⠋⠙⠹...`) and elapsed timer (`⏱️ 1.2s`).
  - **SFTP Manager**: `file_manager_*.go` with 2-pane view, clipboard staging (`[x]➔[p]`), and in-manager shell.
  - **Modals**:
    - `vault_modal_*.go`: Master password initialization and unlock with Caps Lock detector.
    - `tunnel_modal_*.go`: SSH port forwarding configuration and live ON/OFF toggling.
    - `aicopilot_modal_*.go`: Inline AI assistant with context-aware prompt injection and 1-click execution (`Enter`) / runbook save (`Ctrl+S`).
    - `settings_modal_*.go`: Tabbed preferences (General, Telemetry, Logs, Database & Security, AI Copilot, About).
    - `settings_modal_reset.go`: 2FA master password verification popup for Factory Reset.

---

## 🔀 Key Data & Execution Flows

### 1. Dual-Pane SFTP Clipboard (Cut / Copy / Paste) Flow
```text
[Select Files via Space]
        │
        ├── Press [x] (Cut)  ──► Stage paths in Clipboard (isCut = true, [✂] badge)
        └── Press [c] (Copy) ──► Stage paths in Clipboard (isCut = false, [📋] badge)
                                          │
                        [Freely Navigate Folders with Arrows/Enter/Backspace]
                                          │
                                   Press [p] (Paste)
                                          │
              ┌───────────────────────────┴───────────────────────────┐
              ▼                                                       ▼
      [Same Environment]                                      [Cross-Environment]
(Local ➔ Local or Remote ➔ Remote)                     (Local ➔ Remote or Remote ➔ Local)
              │                                                       │
  Instant os.Rename / SFTP Rename                       Background Async SFTP Pipeline
   (0.01s disk-level move/copy)                          (Chunked stream + live progress bar)
              │                                                       │
              └───────────────────────────┬───────────────────────────┘
                                          │
                      [Clear Clipboard + Show '✨ Success' Banner]
                                          │
                      [Auto Refresh File Manager Listings]
```

### 2. Context-Aware AI Copilot Flow (`[F4]`)
```text
[User presses F4 in Cockpit]
        │
        ▼
[Capture Real-Time Context]
  • Current Server Telemetry: CPU (%), RAM (Used/Total), Disk (Used/Total)
  • Host OS Distro & Working Directory (CWD)
  • Last executed shell command, Exit Code, and Stderr output
        │
        ▼
[Stream Query to LLM (Groq / Ollama / OpenAI)] ──► Real-Time SSE Token Streaming
        │
        ▼
[Safety Inspection via ai_safety.go]
  • Destructive commands (reboot, rm -rf, etc.) blocked with warning
        │
        ▼
[User Action]
  ├── Press [Enter]   ──► Transfer suggested command to active shell & execute
  ├── Press [Tab]     ──► Copy suggested command to input prompt for review
  └── Press [Ctrl+S]  ──► 1-Click Save to Custom Runbook (⭐️)
```

### 3. SSH Local Port Forwarding (Tunneling) Flow (`[T] / [F7]`)
```text
[User activates Tunnel: Local 3306 ➔ Remote 127.0.0.1:3306]
        │
        ▼
[Bind Local TCP Listener: 127.0.0.1:3306]
        │
        ▼
[Incoming Client Connection (e.g., MySQL Workbench, DBeaver)]
        │
        ▼
[Multiplexed SSH Channel via Client Pool: client.Dial("tcp", "127.0.0.1:3306")]
        │
        ▼
[Bidirectional Zero-Copy Streaming: io.Copy(sshChan, localConn) & io.Copy(localConn, sshChan)]
        │
        ▼
[Live Status Indicator & Active Connection Counter updated in Cockpit Header]
```

### 4. Factory Reset 2FA Authorization Flow (`[p] ➔ [4] ➔ [6]`)
```text
[User selects [6] Factory Reset in Settings Database Tab]
        │
        ▼
[Display Warning Modal with Password Input (settings_modal_reset.go)]
        │
        ▼
[User types Master Password & presses Enter]
        │
        ├── Password Invalid  ──► Show Red Warning Banner ("Master password incorrect")
        │
        └── Password Verified
                │
                ▼
        [Execute Storage.FactoryReset()]
          • Wipe: hosts, metrics_raw, metrics_hourly, custom_commands,
                  ssh_tunnels, ai_chat_history, app_settings, vault_meta
          • SQLite VACUUM to reclaim disk
                │
                ▼
        [Purge In-Memory State]
          • Close all SSH connections & Tunnels
          • Purge hosts, metrics, tabs, and lock Vault
                │
                ▼
        [Switch directly to Initial Master Password Setup Modal (VaultModalInit)]
```

---

## 🛡️ Software Engineering Rules & Standards

1. **Strict 250-Line Limit**: All business logic and UI presentation source files must not exceed 250 lines to ensure high cohesion and modularity (*Pure data catalogs such as `dict_*.go` and `quickcmd/tab_*.go` are explicitly exempted*).
2. **0% CGO (Pure Go)**: Cross-compiles seamlessly to Windows, macOS, and Linux without external C library dependencies.
3. **100% Dictionary Parity**: Zero hardcoded strings. Every UI label, placeholder, and error message is synchronized across English, Korean, and German.
