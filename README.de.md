# ⚡ LEITSTAND

<p align="center">
  <a href="README.md"><b>English</b></a> •
  <a href="README.ko.md"><b>한국어</b></a> •
  <a href="README.de.md"><b>Deutsch</b></a>
</p>

<p align="center">
  <img src="https://img.shields.io/badge/Go-1.22+-00ADD8?style=flat&logo=go" alt="Go Version">
  <img src="https://img.shields.io/badge/Arch-Zero--CGO_Pure_Go-blue" alt="Architecture">
  <img src="https://img.shields.io/badge/Platform-Linux%20%7C%20macOS%20%7C%20Windows-lightgrey" alt="Platform">
  <img src="https://img.shields.io/badge/License-MIT-green" alt="License">
</p>

> **Modernes, leichtgewichtiges & abhängigkeitsfreies Terminal-Server-Cockpit & SFTP Dual-Pane-Dateimanager**  
> *Eine agentenlose Server-Leitstelle für Entwickler, DevOps und Systemingenieure.*

---

## 🌟 Überblick (Overview)

**LEITSTAND** ist ein hochperformantes Terminal-User-Interface (TUI)-Cockpit, geschrieben in **Go** und basierend auf dem **Charmbracelet Lipgloss & Bubbletea**-Ökosystem.

Es ersetzt überladene Desktop-Tools durch ein **schnelles, tastaturgesteuertes Single-Binary-Cockpit**, das sich über Standard-SSH mit jedem entfernten Linux/Unix-Server verbindet – **ohne dass Agenten, Hintergrund-Daemons oder zusätzliche Abhängigkeiten auf den Zielsystemen erforderlich sind**.

---

## 🚀 Wichtigste Funktionen (Key Highlights)

### 1. 🖥️ Einheitliches 3-Pane-Live-Cockpit
- **Host-Explorer (Links)**: Multi-Server-Flottenliste mit sofortigen Verbindungsstatus-Badges (`🟢 Online`, `🔴 Offline`), benutzerdefinierten Gruppen und SSH-Verbindungsmultiplexing.
- **Telemetrie-Deck (Oben rechts)**: Echtzeit-Anzeigen für **CPU-Auslastung**, **Speicherbelegung**, **Festplattenplatz** und **Live-Netzwerk-I/O-Geschwindigkeit** (B/s, KB/s, MB/s), bedarfsgesteuert ohne Datenbank-Overhead abgefragt.
- **Zustandsbehaftete Remote-Konsole (Unten rechts)**: Vollwertige Remote-Shell mit Arbeitsverzeichnistracking (`CWD`), Befehlshistorie und Syntaxhervorhebung.

### 2. 📑 Multi-Tab-Shell mit Statuserhalt
- **Host-isolierte Multi-Tabs**: Öffnen und Verwalten mehrerer Terminal-Tabs pro Server (`Ctrl+T` / `Ctrl+N`, `Ctrl+W`, `Alt+1`~`Alt+9`).
- **Unabhängige Zustandserhaltung**: Jeder Tab behält sein eigenes Arbeitsverzeichnis (`CWD`), seine Befehlshistorie (`↑/↓`) und seinen Scroll-Viewport bei.
- **Asynchrones Live-Streaming**: Ausführung von Streaming-Befehlen (`tail -f /var/log/syslog`, `docker logs -f`, `ping`) im Hintergrund mit blinkenden `🔴 LIVE`-Badges. Das Abbrechen eines Streams (`Ctrl+C`) beendet nur diesen spezifischen Job, ohne den Tab zu trennen.

### 3. 🛡️ SSH Bastion / Jump-Host-Tunneling
- **Zugriff auf private Subnetze**: Nahtloser Zugriff auf isolierte Server hinter Unternehmens-Firewalls über standardmäßige SSH-ProxyJump-Bastion-Hosts.
- **Automatische Host-Fingerprint-Verifikation**: Echtzeit-SSH-Schlüsselaustausch und strikte Fingerprint-Prüfung zur Gewährleistung manipulationssicherer Tunnel.

### 4. 📂 SFTP Dual-Pane-Dateimanager & Zwischenablage-Bewegung (`[f]`)
- **90 % breite Dual-Pane-Ansicht**: Lokaler PC (Links) ↔ Remote-Server (Rechts) mit sofortigem Fokuswechsel (`[Tab]`).
- **Schnelle Navigation & Suche**: Rasantes Scrollen (`[PgUp]`/`[PgDn]`, `[Home]`/`[End]`) und Dateinamensfilter in Echtzeit (`[/]`).
- **Ausschneiden & Einfügen via Zwischenablage (`[x]` ➔ Freie Navigation ➔ `[p]`)**:
  - **`[x]` / `[Ctrl+X]` (Ausschneiden / Für Verschiebung markieren)**: Dateien mit `[✂]`-Badge in die Zwischenablage legen.
  - **Freie Navigation**: Beliebig mit Pfeiltasten, `[Enter]` (Ordner öffnen) und `[Backspace]` (Übergeordneter Ordner) navigieren.
  - **`[p]` / `[Ctrl+V]` (Einfügen / Ablegen)**: Sofortiges Verschieben auf Festplattenebene (`os.Rename`) oder SFTP-Move in den aktuell angezeigten Ordner!
  - **`[c]` / `[Ctrl+C]` (Kopieren)**: Dateien zum Duplizieren vormerken.
- **Batch-Transfer mit Live-Fortschritt**: Mehrfachauswahl (`[Space]`, `[a]`) und Stapel-**Upload (`[u]`)** oder **-Download (`[d]`)** mit Chunk-Streaming, Fortschrittsbalken und MB/s-Durchsatzanzeige.
- **Sofortige Shell-Ausführung im Dateimanager (`[:]` / `[!]`)**:
  - Shell-Befehle (`ls -la`, `cd ..`, `df -h`, `chmod`, `tar -xvf`) direkt im aktiven Verzeichnis ausführen.
  - `cd`-Befehle synchronisieren automatisch die Verzeichnisansicht!
- **Sicherheitsmechanismen**:
  - **Zugriffsschutz**: Automatischer Rollback mit Warnbanner beim Versuch, nicht berechtigte Ordner zu betreten.
  - **Explizite Löschbestätigung**: Anzeige des genauen Dateinamens und -typs (`🗑️ 'production.db' (Datei) endgültig löschen? [y/n]`).
  - **Beendigungs-Schutzmodal**: Verhindert versehentliches Schließen bei `[Esc]`, `[q]` oder `[f]`.

### 5. 🔐 Lokaler verschlüsselter Master-Tresor & Sicherheitswartung (`[p]` Tab 4)
- **Argon2id KDF + AES-256-GCM**: Industriestandard-Verschlüsselung für alle Server-Passwörter, Sudo-Passwörter und privaten SSH-Schlüssel.
- **SSH-Schlüssel-Manager (`[b]` Dateiauswahl)**: Integrierter Explorer zum Laden von `id_rsa`, `id_ed25519` und `.pem`-Schlüsseln aus `~/.ssh/`.
- **Master-Passwort-Neuschlüsselung (Rekeying)**: Reibungslose Neuverschlüsselung aller gespeicherten Anmeldedaten mit einem neuen Master-Passwort in Tab `[4]`.
- **Zwei-Faktor-Werkseinstellungs-Schutz (Factory Reset)**: Sicherheitsabfrage mit erneuter Master-Passworteingabe vor dem Löschen der Daten.
- **Feststelltasten-Erkennung**: Echtzeit-Warnbanner `[🔒 CAPS LOCK ON]` zur Vermeidung von Tippfehlern.

### 6. 🗄️ Datenbankwartung & Audit-Management (`[p]` Tab 4)
- **Echtzeit-DB-Diagnose**: Live-Dateigröße, registrierte Host-Anzahl und Telemetrie-Snapshot-Statistiken.
- **Konfigurierbare Aufbewahrung & Disk-Vacuum**: Veraltete Metriken per 1-Klick bereinigen (7 Tage, 14 Tage, 30 Tage) mit sofortigem SQLite-`VACUUM`.
- **Audit-Export / -Import**:
  - **Metriken-CSV-Export**: Historische CPU-, RAM-, Festplatten- und Netzwerkdaten in CSV-Dateien exportieren.
  - **Host-JSON-Sicherung & Wiederherstellung**: Hosts in portables JSON sichern und ohne Duplikate wiederherstellen.

### 7. 🤖 AI Terminal-Copilot & Autonome Diagnose (`[F4]`)
- **Multi-Provider-Unterstützung**: Ultraschnelle **Groq Cloud API** (kostenlose LPU), lokales **Ollama** (offline/privat) und **OpenAI** mit isoliertem Caching.
- **Kontextsensitive Diagnose**: Automatische Übergabe von Echtzeit-Telemetrie, OS-Distribution, Arbeitsverzeichnis und Exit-Code/Stderr an das LLM.
- **Zweistufige UX**: Prägnante Erklärung der Ursache in 1–2 Sätzen und Bereitstellung optimaler Einzeiler-Befehle zur Behebung.
- **1-Klick-Ausführung & Tab-Bearbeitung**: Vorgeschlagene Befehle mit `[Enter]` sofort ausführen oder mit `[Tab]` zur Prüfung in die Konsole übernehmen.
- **Bash-Historie (`↑ / ↓`)**: Vorherige KI-Anfragen und Befehle mit den Pfeiltasten durchblättern.
- **Sicherheits-Leitplanken**: Strikt auf Linux-Ops beschränkte Prompts und clientseitige Blockade destruktiver Befehle (`rm -rf`, `shutdown`).

### 8. 🌐 Mehrsprachige Lokalisierung (i18n)
- Vollständige native Unterstützung für **Deutsch**, **English** und **한국어 (Koreanisch)** in allen Ansichten, Dialogen und Runbooks.

### 9. 📖 Modulare Runbooks & Tastaturkürzel-Übersicht (`[?]`)
- **[1] ⌨️ Shortcuts Guide**: Vollständige Tastenkürzel-Referenz für Explorer, Konsole und Dateimanager.
- **[2]~[6] OS Runbooks**: Sofort einsatzbereite Diagnose-Befehle für Common Linux, Ubuntu, RHEL/Rocky, Alpine und Docker.

### 10. 🧪 Offline-Demomodus (`--demo`)
- Starten mit `leitstand --demo` für eine voll funktionsfähige Mock-Umgebung mit realistischer Telemetrie ohne echte SSH-Server.

---

## ⌨️ Tastenkürzel-Übersicht (Cheat Sheet)

### 🖥️ Hauptcockpit
| Tastenkürzel | Aktion |
|---|---|
| **`[↑/↓]`**, **`[j/k]`** | Serverliste navigieren |
| **`[Enter]`** | Server auswählen & Remote-Konsole öffnen |
| **`[Tab]`** | Fokus zwischen Serverliste und Konsole wechseln |
| **`[a]`** / **`[e]`** / **`[x]`** | Server hinzufügen / bearbeiten / löschen |
| **`[f]`**, **`[F6]`** | SFTP Dual-Pane-Dateimanager öffnen |
| **`[F4]`** | AI Terminal-Copilot & Diagnose öffnen |
| **`[t]`** | Vollbild-interaktives PTY-Terminal starten |
| **`[Ctrl+O]`** | Vollbild-Konsole umschalten |
| **`[?]`**, **`[Ctrl+K]`** | Schnelle Befehls-Runbooks & Hilfe anzeigen |
| **`[p]`**, **`[,]`** | Einstellungen & Entwicklerprofil öffnen |
| **`[Ctrl+T]`** / **`[Ctrl+N]`** | Neuen Konsolen-Tab erstellen |
| **`[Alt+1]` ~ `[Alt+9]`** | Direkt zu Tab wechseln |
| **`[Ctrl+W]`** | Aktiven Konsolen-Tab schließen |
| **`[Ctrl+C]`** | Laufenden Stream im aktiven Tab abbrechen |
| **`[q]`**, **`[Esc]`** | Anwendung beenden |

---

### 📂 SFTP Dual-Pane-Dateimanager (`[f]`)
| Tastenkürzel | Aktion |
|---|---|
| **`[Tab]`**, **`[◄/►]`** | Aktives Panel wechseln (Lokaler PC ↔ Remote-Server) |
| **`[↑/↓]`**, **`[j/k]`** | Datei- / Ordnerliste navigieren (1 Zeile) |
| **`[PgUp/PgDn]`**, **`[Ctrl+U/D]`** | Schneller Seitensprung (Ganzen Bildschirm überspringen) |
| **`[Home/End]`**, **`[g/G]`** | An den Anfang / ans Ende der Liste springen |
| **`[/]`** | Dateinamen-Echtzeitsuche / Filter |
| **`[Enter]`** | Verzeichnis öffnen (mit Auto-Rollback-Schutz) |
| **`[Backspace]`** | In übergeordnetes Verzeichnis wechseln (`..`) |
| **`[Space]`** | Mehrfachauswahl-Markierung umschalten (`[*]`) |
| **`[a]`** | Alle Dateien auswählen / abwählen |
| **`[x]`**, **`[Ctrl+X]`** | **Ausschneiden**: Zielordner ansteuern und `[p]` drücken |
| **`[c]`**, **`[Ctrl+C]`** | **Kopieren**: Zielordner ansteuern und `[p]` drücken |
| **`[p]`**, **`[Ctrl+V]`** | **Einfügen**: Markierte Dateien im aktuellen Ordner ablegen |
| **`[u]`** | Ausgewählte Dateien hochladen (Lokal ➔ Remote) |
| **`[d]`** | Ausgewählte Dateien herunterladen (Remote ➔ Lokal) |
| **`[:]`**, **`[!]`** | Shell-Befehl im aktuellen Verzeichnis ausführen |
| **`[n]`** / **`[N]`** | Neuen Ordner (`mkdir`) / Neue Datei (`touch`) erstellen |
| **`[r]`** | Datei / Ordner umbenennen |
| **`[Delete]`**, **`[Shift+X]`** | Ausgewählte Elemente löschen (mit Bestätigung) |
| **`[.]`** | Versteckte Dateien ein-/ausblenden (`.env`, `.*`) |
| **`[F5]`** | Verzeichnisansicht aktualisieren |
| **`[?]`**, **`[F1]`** | Dateimanager-Hilfe öffnen |
| **`[Esc]`**, **`[q]`** | Zwischenablage leeren / Sicher beenden |

---

## 📦 Download & Schnellstart (Keine Go-Installation erforderlich)

Sie müssen weder Go noch externe Build-Tools installieren. Laden Sie einfach das vorkompilierte Single-Binary-Paket aus dem Ordner [`dist/`](dist/) herunter und starten Sie LEITSTAND sofort:

| Plattform | Architektur | Download-Archiv | Kurzanleitung |
|---|---|---|---|
| **Windows** | x86_64 (`amd64`) | [**leitstand-v1.0.0-windows-amd64.zip**](dist/leitstand-v1.0.0-windows-amd64.zip) | Entpacken und `leitstand-windows-amd64.exe` ausführen |
| **Linux** | x86_64 (`amd64`) | [**leitstand-v1.0.0-linux-amd64.tar.gz**](dist/leitstand-v1.0.0-linux-amd64.tar.gz) ([`.zip`](dist/leitstand-v1.0.0-linux-amd64.zip)) | `tar -xvf ... && ./leitstand-linux-amd64` |
| **macOS** | Apple Silicon (`arm64`) | [**leitstand-v1.0.0-darwin-arm64.tar.gz**](dist/leitstand-v1.0.0-darwin-arm64.tar.gz) ([`.zip`](dist/leitstand-v1.0.0-darwin-arm64.zip)) | `tar -xvf ... && ./leitstand-darwin-arm64` |

> [!TIP]
> Testen Sie die TUI ohne echte Server über den integrierten Offline-Demomodus:  
> `./leitstand --demo`

---

## 🛠️ Installation & Cross-Platform-Build

### Voraussetzungen
- **Go 1.22+** installiert.
- Ein beliebiges modernes ANSI/UTF-8-Terminal (Windows Terminal, iTerm2, Alacritty, Kitty, GNOME Terminal usw.).

### 1. Für das aktuelle Betriebssystem kompilieren
```bash
# Repository klonen
git clone https://github.com/sorrynthx/leitstand.git
cd leitstand

# Unit-Tests ausführen
go test -v ./...

# Binary kompilieren
go build -o bin/leitstand ./cmd/leitstand
```

### 2. Plattformübergreifende Kompilierung (Zero-CGO Pure Go)
Da LEITSTAND eine CGO-freie SQLite-Engine nutzt, können Binärdateien für jedes OS ohne externe Toolchains erstellt werden:

```bash
# Linux (amd64)
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o bin/leitstand-linux-amd64 ./cmd/leitstand

# macOS Apple Silicon (arm64)
CGO_ENABLED=0 GOOS=darwin GOARCH=arm64 go build -o bin/leitstand-darwin-arm64 ./cmd/leitstand

# Windows (amd64)
CGO_ENABLED=0 GOOS=windows GOARCH=amd64 go build -o bin/leitstand-windows-amd64.exe ./cmd/leitstand
```

### 3. Starten
```bash
# Normaler Betrieb
./bin/leitstand

# Offline-Demomodus (sofortige Präsentation ohne Server)
./bin/leitstand --demo
```

---

## 🏛️ Systemarchitektur

Detaillierte Ablaufdiagramme und technische Spezifikationen finden Sie unter [docs/ARCHITECTURE.md](docs/ARCHITECTURE.md).

```text
┌─────────────────────────────────────────────────────────────────────────────┐
│                          Leitstand TUI Frontend                             │
│       (Bubbletea Model-View-Update + Lipgloss Styling + Multi-Tab)          │
├───────────────────┬─────────────────────────┬───────────────────────────────┤
│   Host Explorer   │     Telemetry Deck      │     Remote Command Console    │
│   (Server List)   │ (CPU/RAM/Disk/Net Gauges)│  (Multi-Tab, Stream, Root#)   │
├───────────────────┴─────────────────────────┴───────────────────────────────┤
│              SFTP Dual-Pane File Manager & Clipboard Engine                 │
│      (Local ↔ Remote, Cut/Paste [x]➔[p], Instant Shell [:], Progress)       │
├─────────────────────────────────────────────────────────────────────────────┤
│                           Core Application Layer                            │
├────────────────────────────────┬────────────────────────────────────────────┤
│       SSH Client Pool          │            Telemetry Collector             │
│ (Single TCP Socket/Host Mux)   │   (Agentless /proc & Command Polling)      │
├────────────────────────────────┼────────────────────────────────────────────┤
│       Master Vault Engine      │            Demo Simulator Layer            │
│  (AES-256-GCM + Argon2id Key)  │  (Realistic Mock Engine & Video Demo Mode) │
├────────────────────────────────┼────────────────────────────────────────────┤
│      AI Copilot Engine         │            Bastion Jump Tunnel             │
│ (Groq Cloud / Ollama / OpenAI) │      (SSH ProxyJump Direct-TCP Tunnel)     │
├────────────────────────────────┴────────────────────────────────────────────┤
│                      Local SQLite Storage & Config                          │
│               (Hosts, Credentials, Metric Snapshots, i18n)                  │
└─────────────────────────────────────────────────────────────────────────────┘
```

---

## 👤 Entwickler & Vision

- **Urheber**: **Kyunggon Kim (김경곤)**
- **Vision**: Bereitstellung reaktionsschneller, abhängigkeitsfreier Entwickler-Infrastruktur-Tools, mit denen verteilte Serverflotten mit maximaler Geschwindigkeit und Transparenz gesteuert werden können.
- **GitHub Repository**: [github.com/sorrynthx/leitstand](https://github.com/sorrynthx/leitstand)
- **Lizenz**: MIT License

