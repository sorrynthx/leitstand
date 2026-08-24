````markdown
# Leitstand: Architectural Blueprint & Implementation Plan

> **Leitstand** (`leitstand` / [ˈlaɪt.ʃtant]): A lightweight, agentless server control center and telemetry cockpit written in Go.

---

## 1. Project Overview & Vision

* **Target Users**: Backend/DevOps/SRE engineers managing multiple Linux/Unix remote servers across various environments.
* **Core Philosophy**:
  * **Zero-Agent**: No daemons or agents installed on target servers. Pure SSH protocol and POSIX-compliant data extraction.
  * **Zero-Knowledge Local Vault**: Server credentials and private keys are encrypted locally using master-key derivation.
  * **Single Binary & Ultra-Low Footprint**: Cross-platform (macOS Apple Silicon/Intel, Windows, Linux) with pure Go dependencies (no CGO).
  * **Native TUI Cockpit**: Sub-millisecond responsive Terminal UI with interactive PTY session multiplexing and real-time telemetry.

---

## 2. Problem Statement & Solution

| Problem (Legacy Tools / MobaXterm / Heavy GUIs) | Leitstand Solution |
|---|---|
| License restrictions on session counts (free-tier limits) | Unlimited hosts, groups, and tags |
| High RAM/CPU overhead (Electron/heavy GUI engines) | Pure Go TUI (~15MB–30MB RAM, < 0.05s startup) |
| Mandatory agent installation on target machines | Pure SSH polling via POSIX commands (`/proc/*`, `systemctl`) |
| Platform lock-in (Windows-centric MobaXterm) | Cross-platform native binary (macOS, Windows, Linux) |
| Insecure/plaintext credential storage | Argon2id + AES-256-GCM encrypted local SQLite storage |

---

## 3. Technical Architecture

### 3.1. Security & Vault Subsystem

* **Key Derivation Function (KDF)**: `Argon2id`
  * Memory: 64 MB, Iterations: 3, Parallelism: 4
  * Derives a 256-bit encryption key from a single user master password.
* **Symmetric Encryption**: `AES-256-GCM`
  * Encrypts connection secrets (IP, Port, Username, SSH Passphrase, Private Keys).
  * Nonce generated via `crypto/rand` per record.
* **Memory Safety**: Decrypted key byte slices are zeroed (`memclr`) after SSH handshake.

### 3.2. Agentless Telemetry Pipeline

* **Connection Multiplexing**: A shared SSH client connection per host (`golang.org/x/crypto/ssh`).
  * **Channel 1 (Background)**: Non-interactive exec session running batched POSIX metric scripts at configurable intervals (e.g., every 5s).
  * **Channel 2 (Interactive)**: On-demand PTY allocation for full interactive shell sessions.
* **OS Strategy Pattern**:
  * Auto-detects OS/Distro (`/etc/os-release`, `uname -s`).
  * Parses `/proc/stat` (CPU), `/proc/meminfo` (RAM), `df -k` (Disk), `/proc/net/dev` (Network I/O), and `systemctl list-units` (Services).

### 3.3. Embedded Time-Series Storage (Pure-Go SQLite)

* **Driver**: `modernc.org/sqlite` (No CGO, cross-compile ready).
* **PRAGMA Optimizations**:

  ```sql
  PRAGMA journal_mode = WAL;
  PRAGMA synchronous = NORMAL;
  PRAGMA cache_size = -64000;
  PRAGMA temp_store = MEMORY;
````

* **Data Lifecycle & Downsampling Tiering**:

  1. **Tier 1 (Raw Data, 5–10s ticks)**: Retained for **7 days** (real-time troubleshooting).
  2. **Tier 2 (10-minute Aggregates)**: Retained for **30 days** (monthly trends).
  3. **Tier 3 (1-hour Aggregates)**: Retained for **up to 1 year** (annual reporting).

* **Storage Budget**: ~30MB–60MB SQLite DB size for 10 hosts over 1 year.

### 3.4. TUI Cockpit & Layout Engine

* **Framework**: `charmbracelet/bubbletea` + `lipgloss`

* **Minimum Resolution Guard**: `100 cols × 28 rows`

  * Displays a resizing hint when the viewport falls below standard dimensions.

* **3-Pane Adaptive Layout Grid**:

  * **Pane 1 (Left, ~25–30% Width)**: Host Explorer (Grouped tree view with live ping indicators).
  * **Pane 2 (Right Top, ~45% Height)**: Live Telemetry Deck (CPU/RAM/Disk gauges, mini sparklines, systemd status).
  * **Pane 3 (Right Bottom, ~55% Height)**: Interactive Remote PTY / Snippet Console.
  * **Full-screen Toggle (`Enter` / `F`)**: Expands PTY or Metric views to full viewport.

### 3.5. CJK Width Safe Rendering & i18n

* **Cell-Width Normalization**: Use `lipgloss.Width()` and `mattn/go-runewidth` to calculate visual column width instead of byte/rune count, preventing border deformation across mixed languages.

* **Localization (`embed.FS`)**:

  * `en` (Default / Global)
  * `de` (German Engineering Locale: *Leitstand, Verbindung, Systemstatus*)
  * `ko` (Korean Locale)

---

## 4. Tech Stack & Dependencies

| **Layer**          | **Technology**                             | **Purpose**                                      |
| ------------------ | ------------------------------------------ | ------------------------------------------------ |
| **Core Language**  | Go 1.23+                                   | High concurrency, fast I/O, cross-compilation    |
| **CLI & Config**   | `spf13/cobra`, `spf13/viper`               | CLI commands, flags, and configuration loading   |
| **TUI & Styling**  | `charmbracelet/bubbletea`, `lipgloss`      | Elm-architecture TUI and responsive layout       |
| **SSH & Terminal** | `golang.org/x/crypto/ssh`, `creack/pty`    | SSH connection pooling and virtual PTY emulation |
| **Embedded DB**    | `modernc.org/sqlite`                       | Zero-install embedded SQL engine (pure Go)       |
| **Crypto & Vault** | `golang.org/x/crypto/argon2`, `crypto/aes` | Master password KDF and AES-GCM vault storage    |
| **Width Calc**     | `mattn/go-runewidth`                       | CJK and East Asian character display alignment   |

---

## 5. Phased Implementation Roadmap

```text
Phase 1: Core Engine & CLI Vault (MVP)
├── [x] Project scaffolding, Cobra CLI setup, config structure
├── [ ] Argon2id + AES-256-GCM Vault storage on SQLite
├── [ ] SSH Connection Pool & basic Linux metric parser (/proc/stat, meminfo)
└── [ ] Minimal Bubbletea 2-Pane UI (Host List + Live CPU/RAM gauge)

Phase 2: Cockpit UI & Terminal PTY
├── [ ] Responsive 3-Pane Cockpit with Lipgloss (Width-safe rendering)
├── [ ] Interactive PTY terminal integration inside Bubbletea
├── [ ] Systemd service status check & quick restart triggers
└── [ ] i18n support (en, de, ko) via embed.FS

Phase 3: Telemetry Engine & Automation
├── [ ] Time-series downsampling engine (Raw -> 10m -> 1h rollups)
├── [ ] Historical trend charts & exportable HTML/Markdown reports
├── [ ] Custom one-click snippet manager (Batch remote command runner)
└── [ ] Multi-platform packaging (Homebrew, Scoop, WinGet, AUR)
```

---

## 6. Portfolio & Career Positioning (DACH / Global Focus)

* **Key Narrative**: "Engineered a zero-agent, memory-safe infrastructure cockpit in Go, replacing heavy legacy terminal multiplexers with a single-binary, encrypted terminal manager."

### Highlighted Competencies

* Advanced Go concurrency patterns (`goroutines`, `channels`, `worker pools`, `sync.Pool`).
* Low-level systems engineering (SSH protocol internals, POSIX metrics, PTY allocation).
* Robust data security architecture (Zero-knowledge local vault using Argon2id/AES-GCM).
* Clean architectural design (Strategy pattern, Elm TUI architecture, pure Go storage engine).

---

## Phase 1 — Recommended Go Project Structure

```text
leitstand/
├── cmd/
│   └── leitstand/
│       └── main.go
│
├── internal/
│   ├── app/
│   │   └── app.go
│   │
│   ├── config/
│   │   ├── config.go
│   │   └── defaults.go
│   │
│   ├── vault/
│   │   ├── vault.go
│   │   ├── crypto.go
│   │   ├── kdf.go
│   │   └── store.go
│   │
│   ├── ssh/
│   │   ├── client.go
│   │   ├── pool.go
│   │   └── auth.go
│   │
│   ├── telemetry/
│   │   ├── collector.go
│   │   ├── parser.go
│   │   ├── cpu.go
│   │   └── memory.go
│   │
│   ├── storage/
│   │   ├── database.go
│   │   ├── migrations.go
│   │   └── metrics.go
│   │
│   └── tui/
│       ├── model.go
│       ├── update.go
│       ├── view.go
│       └── styles.go
│
├── migrations/
│   └── 001_initial.sql
│
├── configs/
│   └── config.example.yaml
│
├── testdata/
│   ├── proc_stat.txt
│   └── meminfo.txt
│
├── go.mod
├── go.sum
├── Makefile
├── README.md
└── LICENSE
```

### Phase 1 Dependency Baseline

```go
module github.com/<your-github-username>/leitstand

go 1.23

require (
    github.com/charmbracelet/bubbletea v1.3.10
    github.com/charmbracelet/lipgloss v1.1.0
    github.com/mattn/go-runewidth v0.0.16
    github.com/spf13/cobra v1.8.1
    github.com/spf13/viper v1.19.0
    golang.org/x/crypto v0.31.0
    modernc.org/sqlite v1.34.5
)
```

> **Note:** `creack/pty`는 Phase 2의 실제 interactive PTY 구현 시 추가하는 편이 깔끔합니다. 또한 실제 프로젝트 초기화 시점에는 각 의존성의 최신 호환 버전을 `go get`으로 확정하고 `go mod tidy`로 `go.sum`을 생성하는 것을 권장합니다.

---

## Phase 1 Implementation Order

```text
1. Repository + Go module
        │
        ▼
2. Config subsystem
        │
        ▼
3. SQLite database initialization
        │
        ▼
4. Argon2id key derivation
        │
        ▼
5. AES-256-GCM Vault
        │
        ▼
6. SSH authentication + connection pool
        │
        ▼
7. /proc/stat + /proc/meminfo parser
        │
        ▼
8. Telemetry polling loop
        │
        ▼
9. Bubbletea MVP
        │
        ▼
10. Integration tests
```

### Recommended Phase 1 Package Responsibilities

```text
internal/config
    └── Application configuration only.
        No SSH, DB, or TUI logic.

internal/vault
    └── Master-password handling
    └── Argon2id
    └── AES-256-GCM
    └── Encrypted secret persistence

internal/storage
    └── SQLite connection
    └── Schema/migrations
    └── Host + metric persistence

internal/ssh
    └── SSH authentication
    └── Connection lifecycle
    └── Session creation
    └── Remote command execution

internal/telemetry
    └── Remote metric collection
    └── POSIX/Linux parsing
    └── Normalized metric models

internal/tui
    └── Bubbletea Model
    └── Input handling
    └── Rendering
    └── Layout calculation

internal/app
    └── Wires all subsystems together.
    └── Owns application lifecycle.
```

---

## Suggested Initial Domain Models

```go
type Host struct {
    ID        int64
    Name      string
    Address   string
    Port      int
    Username  string
    Group     string
    CreatedAt time.Time
    UpdatedAt time.Time
}

type HostSecret struct {
    HostID       int64
    AuthMethod   string
    Password     []byte
    PrivateKey   []byte
    Passphrase   []byte
}

type Metrics struct {
    Timestamp   time.Time
    CPUPercent  float64
    MemoryTotal uint64
    MemoryUsed  uint64
    DiskUsed    uint64
    DiskTotal   uint64
    NetRxBytes  uint64
    NetTxBytes  uint64
}
```

---

## Vault Design

```text
Master Password
       │
       ▼
   Argon2id
       │
       │  256-bit key
       ▼
┌───────────────────────┐
│ AES-256-GCM           │
│                       │
│ nonce + ciphertext    │
│ + authentication tag  │
└───────────────────────┘
       │
       ▼
     SQLite
```

### Important Vault Rules

```text
DO:
- Generate a unique random salt for the vault.
- Generate a fresh 96-bit nonce for every AES-GCM encryption.
- Store ciphertext and nonce separately.
- Authenticate encrypted records with GCM.
- Keep the master password out of persistent storage.
- Zero sensitive byte slices when their lifetime ends.
- Fail closed when authentication/tag verification fails.

DO NOT:
- Store the master password.
- Reuse AES-GCM nonces.
- Store plaintext private keys in SQLite.
- Log passwords, private keys, passphrases, or decrypted secrets.
- Derive the encryption key directly from the password without Argon2id.
```

---

## Minimal SQLite Schema

```sql
CREATE TABLE hosts (
    id          INTEGER PRIMARY KEY AUTOINCREMENT,
    name        TEXT NOT NULL,
    address     TEXT NOT NULL,
    port        INTEGER NOT NULL DEFAULT 22,
    username    TEXT NOT NULL,
    group_name  TEXT NOT NULL DEFAULT '',
    created_at  INTEGER NOT NULL,
    updated_at  INTEGER NOT NULL
);

CREATE TABLE host_secrets (
    host_id       INTEGER PRIMARY KEY,
    auth_method   TEXT NOT NULL,
    nonce         BLOB NOT NULL,
    ciphertext    BLOB NOT NULL,
    FOREIGN KEY (host_id) REFERENCES hosts(id) ON DELETE CASCADE
);

CREATE TABLE metrics_raw (
    host_id       INTEGER NOT NULL,
    timestamp     INTEGER NOT NULL,
    cpu_percent   REAL NOT NULL,
    memory_total  INTEGER NOT NULL,
    memory_used   INTEGER NOT NULL,
    disk_used     INTEGER NOT NULL,
    disk_total    INTEGER NOT NULL,
    net_rx_bytes  INTEGER NOT NULL,
    net_tx_bytes  INTEGER NOT NULL,
    PRIMARY KEY (host_id, timestamp),
    FOREIGN KEY (host_id) REFERENCES hosts(id) ON DELETE CASCADE
);

CREATE INDEX idx_metrics_raw_timestamp
ON metrics_raw(timestamp);

CREATE INDEX idx_metrics_raw_host_timestamp
ON metrics_raw(host_id, timestamp);
```

---

## Phase 1 Acceptance Criteria

```text
[ ] `leitstand` starts without external runtime dependencies.

[ ] First launch creates the local SQLite database.

[ ] User can initialize a master password.

[ ] Vault encryption/decryption passes round-trip tests.

[ ] Incorrect master password fails authentication.

[ ] SSH credentials are never stored in plaintext.

[ ] Leitstand can register at least one remote Linux host.

[ ] Leitstand establishes an SSH connection without installing an agent.

[ ] CPU usage can be collected from /proc/stat.

[ ] Memory usage can be collected from /proc/meminfo.

[ ] Metrics are displayed in a minimal Bubbletea UI.

[ ] SSH/network failures do not crash the TUI.

[ ] Connection timeout and retry behavior are bounded.

[ ] `go test ./...` passes.

[ ] `go vet ./...` passes.

[ ] Cross-compilation works without CGO.
```

---

## Portfolio-Grade Definition of Done

The Phase 1 MVP should demonstrate this complete vertical slice:

```text
User
 │
 │ master password
 ▼
Leitstand CLI
 │
 ├── Argon2id
 │      │
 │      ▼
 │   AES-256-GCM
 │      │
 │      ▼
 │   SQLite Vault
 │
 └── Host Manager
        │
        ▼
   SSH Connection Pool
        │
        ▼
   Remote Linux Host
        │
        ├── /proc/stat
        └── /proc/meminfo
        │
        ▼
   Telemetry Collector
        │
        ▼
   Bubbletea TUI
        │
        ├── Host List
        ├── CPU
        └── Memory
```
