# Leitstand System Architecture

`Leitstand` (라이트슈탄트) is a lightweight, agentless server control center and telemetry cockpit built in Go using the Bubbletea TUI framework.

---

## 🏛️ Architecture Overview

```text
┌─────────────────────────────────────────────────────────────────────────────┐
│                          Leitstand TUI Frontend                             │
│       (Bubbletea Model-View-Update + Lipgloss Sizing + Multi-Tab)           │
├───────────────────┬─────────────────────────┬───────────────────────────────┤
│   Host Explorer   │     Telemetry Deck      │     Remote Command Console    │
│   (Server List)   │ (CPU/RAM/Disk/Net Gauges)│  (Multi-Tab, Stream, Root#)   │
└─────────┬─────────┴────────────┬────────────┴───────────────┬───────────────┘
          │                      │                            │
┌─────────▼──────────────────────▼────────────────────────────▼───────────────┐
│                           Core Application Layer                            │
├────────────────────────────────┬────────────────────────────────────────────┤
│       SSH Client Pool          │            Telemetry Collector             │
│ (Single TCP Socket/Host Mux)   │   (Agentless /proc & Command Polling)      │
├────────────────────────────────┼────────────────────────────────────────────┤
│       Master Vault Engine      │            Demo Simulator Layer            │
│  (AES-256-GCM + Argon2id Key)  │  (Realistic Mock Engine & Video Demo Mode) │
├────────────────────────────────┴────────────────────────────────────────────┤
│                      Local SQLite Storage & Config                          │
│               (Hosts, Credentials, Metric Snapshots, i18n)                  │
└─────────────────────────────────────────────────────────────────────────────┘
```

---

## 📦 Package Structure

- **`cmd/leitstand`**: CLI entry point and root command definitions.
- **`internal/app`**: Application lifecycle bootstrapper and dependency coordinator.
- **`internal/config`**: YAML configuration parser and environment overrides.
- **`internal/highlight`**: Syntax highlighting for JSON, YAML, SQL, and Shell.
- **`internal/i18n`**: Multi-language localization dictionary (`en`, `ko`, `de`).
- **`internal/logger`**: Structured rotation logging.
- **`internal/quickcmd`**: Built-in Runbook dictionary and category tags.
- **`internal/ssh`**: SSH connection pool, channel multiplexer, interactive PTY runner, SFTP file transfer engine (`sftp.go`).
- **`internal/storage`**: SQLite database models, migrations, and CRUD repositories.
- **`internal/telemetry`**: Agentless remote metrics collector and parser.
- **`internal/tui`**: Terminal UI presentation, multi-tab state engine, and event handlers:
  - `model.go`: TUI state struct, bubbletea model initialization.
  - `file_manager_modal.go`: Dual-pane (Local ↔ Remote) SFTP file manager with multi-selection, cut/copy/paste clipboard, inline shell execution, permission rollback, and live progress bar.
  - `view.go`: Master layout composition, header, and status bar.
  - `view_hostlist.go`: Server list rendering and connection badges.
  - `view_telemetry.go`: CPU, Memory, Disk, and Network telemetry gauges.
  - `view_console.go`: Multi-tab remote console and sliding window tab bar.
  - `view_modals.go`: Overlay modals (Vault, Sudo, Editor, Settings, FilePicker).
  - `update.go`: Central message dispatcher and cursor clamping.
  - `update_console.go`: Remote command execution, streaming pipeline, tab autocomplete.
  - `update_hostlist.go`: Server explorer navigation, CRUD shortcuts.
  - `update_drawer.go`: Runbook drawer filtering and search.
  - `update_modals.go`: Modal interactive input processing.
  - `update_sftp.go`: Background batch file upload/download transfer queue, disk-level rename/copy, and inline shell dispatch.
  - `demo_simulator.go`: Self-contained demo simulation layer for testing and video capture.
- **`internal/vault`**: AES-256-GCM encrypted credential vault with Argon2id key derivation.

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

### 2. Telemetry & Multi-Tab Command Execution
- **Telemetry**: Agentless zero-overhead polling of `/proc/stat`, `/proc/meminfo`, `/proc/net/dev`, and `df -h` on the single selected host.
- **Multi-Tab Shell**: Stateful isolated working directory (`CWD`), command history buffer, and non-blocking streaming execution (`tail -f`, `docker logs -f`) with live blink badges.

