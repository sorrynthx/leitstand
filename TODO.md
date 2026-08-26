# Leitstand Development Roadmap & Task Tracker

## 📌 Phase 1: Core Engine & CLI Vault (MVP)
- [x] **Step 1: Scaffolding & Go Dependencies**
  - [x] Initialize directory layout (`cmd/leitstand`, `internal/...`, `migrations/`, `configs/`)
  - [x] Install core dependencies (`bubbletea`, `lipgloss`, `cobra`, `viper`, `golang.org/x/crypto`, `modernc.org/sqlite`, `go-runewidth`)
  - [x] Verify clean compilation without CGO
- [x] **Step 2: Configuration Subsystem (`internal/config`)**
  - [x] Configuration schema, defaults (DB path, polling interval)
  - [x] Viper/Cobra integration
- [x] **Step 3: Storage & Database Layer (`internal/storage`)**
  - [x] SQLite connection setup with WAL mode and PRAGMA optimizations
  - [x] Migration runner and initial schema (`hosts`, `host_secrets`, `metrics_raw`)
- [x] **Step 4: Security & Vault Subsystem (`internal/vault`)**
  - [x] Master password handling & Argon2id KDF key derivation
  - [x] AES-256-GCM symmetric encryption/decryption with nonce & auth tag
  - [x] Unit tests for encryption round-trip and wrong password rejection
- [x] **Step 5: SSH Client & Connection Pool (`internal/ssh`)**
  - [x] Key/Password authentication handlers
  - [x] Thread-safe SSH connection lifecycle & pooling
  - [x] Remote command execution wrapper
- [x] **Step 6: Telemetry Engine (`internal/telemetry`)**
  - [x] `/proc/stat` CPU parser
  - [x] `/proc/meminfo` Memory parser
  - [x] Periodic background polling loop
- [x] **Step 7: Minimal Bubbletea TUI (`internal/tui`)**
  - [x] Basic 2-Pane UI (Host Explorer + CPU/RAM Gauges)
  - [x] CJK width-safe rendering check
  - [x] Phase 1 MVP Integration Test

---

## 📌 Phase 2: Cockpit UI & Terminal PTY
- [x] Responsive 3-Pane Cockpit layout (Host Explorer, Telemetry Deck, Remote Console)
- [x] In-TUI Modal Dialogs (Master Password setup/unlock, Host Registration Form, Delete Confirmation)
- [x] Remote Command Console with scrollable Viewport (`PageUp/PageDown`, `Ctrl+U/Ctrl+D`, `Ctrl+O` Fullscreen)
- [x] Built-in console utility handlers (`clear`, `cls`, `Ctrl+L`) & command auto-normalization (`top` batch mode)
- [x] Rich System Spec Metadata parsing (OS Distribution, Kernel, Uptime, CPU Cores)
- [x] `Tab` cycling navigation across 3 panes with active border highlights
- [ ] Interactive virtual PTY terminal integration (`creack/pty`)
- [ ] Systemd service status checker & restart triggers
- [ ] SSH Key Pair Generator & Remote Deployment (`ssh-copy-id` in-app assistant)
- [ ] Dual-pane TUI SFTP File Manager (Local ⇄ Remote file browsing & transfer)
- [ ] i18n support (`en`, `de`, `ko`)

---

## 📌 Phase 3: Telemetry Engine & Downsampling
- [ ] Time-series downsampling engine (Raw 5s -> 10m -> 1h rollups)
- [ ] Historical trend charts & exportable reports
- [ ] Batch snippet runner (multi-host execution)

---

## 📌 Phase 4 (Extension): Web Cockpit Mode (`leitstand serve --web`)
- [ ] Embedded Web UI (React/Svelte + xterm.js + Web SFTP Drag & Drop) inside single Go binary
