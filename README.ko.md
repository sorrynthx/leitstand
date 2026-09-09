# ⚡ LEITSTAND (라이트슈탄트)

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

> **현대적이고 가벼운 무의존성 터미널 서버 콕핏 & SFTP 듀얼 패널 파일 매니저**  
> *개발자, DevOps 및 시스템 엔지니어를 위해 설계된 에이전트리스(Agentless) 서버 제어 센터.*

---

## 🌟 개요 (Overview)

**LEITSTAND** (*독일어로 "통제소 / 관제 센터"*)는 **Go 언어**와 **Charmbracelet Lipgloss & Bubbletea** 생태계로 제작된 고성능 터미널 사용자 인터페이스(TUI) 관제 콕핏입니다.

대상 서버에 **어떠한 에이전트나 백그라운드 데몬, 추가 런타임 설치도 필요 없이**, 표준 SSH 프로토콜만으로 즉시 연결되는 **빠르고 가벼운 단일 바이너리 키보드 중심의 관제 환경**을 제공하여 무거운 데스크톱 GUI 툴을 대체합니다.

---

## 🚀 주요 핵심 기능 (Key Highlights)

### 1. 🖥️ 3-Pane 통합 라이브 콕핏
- **호스트 탐색기 (좌측)**: 서버 목록, 실시간 연결 헬스 배지(`🟢 Online`, `🔴 Offline`), 사용자 정의 그룹핑 및 단일 TCP 소켓 기반 SSH 연결 다중화(Multiplexing).
- **텔레메트리 덱 (우측 상단)**: 원격 서버의 **CPU 점유율**, **메모리 사용량**, **디스크 용량**, **실시간 네트워크 I/O 속도**(B/s, KB/s, MB/s)를 DB 부하 없이 온디맨드로 폴링하는 초경량 실시간 게이지.
- **상태 보존형 원격 콘솔 (우측 하단)**: 작업 디렉토리 추적(`CWD`), 명령어 히스토리(`↑/↓`), 구문 강조를 지원하는 고성능 원격 셸.

### 2. 📑 멀티 탭 상태 보존 셸
- **서버별 독립 멀티 탭**: 서버당 여러 개의 터미널 탭을 생성하고 전환 (`Ctrl+T` / `Ctrl+N`, `Ctrl+W`, `Alt+1`~`Alt+9`).
- **독립적 상태 보존**: 각 탭마다 고유한 작업 디렉토리(`CWD`), 히스토리 스택, 스크롤 뷰포트를 격리 보존.
- **비동기 실시간 스트리밍**: 로그 모니터링(`tail -f`, `docker logs -f`, `ping`) 실행 시 백그라운드에서 `🔴 LIVE` 깜빡임 배지와 함께 스트리밍되며, `Ctrl+C`로 탭 연결 해제 없이 해당 스트림만 안전하게 중단.

### 3. 🛡️ SSH 베스천 / 점프 호스트 터널링
- **폐쇄망 프라이빗 서브넷 접근**: 사내 방화벽 내부의 폐쇄망 서버도 표준 SSH ProxyJump 베스천 호스트를 경유하여 원클릭으로 직접 터널링 연결.
- **자동 호스트 핑거프린트 검증**: 터널 생성 시 실시간 SSH 키 교환 및 엄격한 핑거프린트 무결성 검증.

### 4. 📂 SFTP 듀얼 패널 파일 매니저 & 클립보드 이동 (`[f]`)
- **90% 와이드 듀얼 패널 뷰**: 로컬 PC (좌측) ↔ 원격 서버 (우측) 간 실시간 포커스 전환 (`[Tab]`).
- **신속한 탐색 및 검색**: 고속 페이지 스크롤 (`[PgUp]`/`[PgDn]`, `[Home]`/`[End]`) 및 실시간 파일명 필터 검색 (`[/]`).
- **클립보드 잘라내기 & 붙여넣기 이동 (`[x]` ➔ 자유 이동 ➔ `[p]`)**:
  - **`[x]` / `[Ctrl+X]` (잘라내기 / 이동 대기)**: 파일에 `[✂]` 배지가 부여되며 클립보드에 등록.
  - **자유 탐색**: 방향키, `[Enter]`(폴더 진입), `[Backspace]`(상위 폴더 이동)를 사용해 원하는 대상 폴더로 자유롭게 이동.
  - **`[p]` / `[Ctrl+V]` (붙여넣기 / 드롭)**: 현재 보고 있는 폴더로 즉시 디스크 레벨 `os.Rename` 또는 SFTP 원격 이동 수행!
  - **`[c]` / `[Ctrl+C]` (복사)**: 원하는 대상 폴더로 복제 대기.
- **실시간 프로그레스 일괄 전송**: 다중 선택(`[Space]`, `[a]`) 후 일괄 **업로드(`[u]`)** 또는 **다운로드(`[d]`)** 시 청크 스트리밍 진행률 바와 MB/s 전송 속도 측정기 표시.
- **매니저 내 인스턴트 셸 실행 (`[:]` / `[!]`)**:
  - 현재 위치한 디렉토리에서 바로 셸 명령(`ls -la`, `chmod`, `tar -xvf`) 실행 가능.
  - `cd` 명령어 입력 시 파일 매니저의 디렉토리 목록이 즉시 동기화!
- **안전장치**:
  - **권한 부족 보호**: 접근 권한이 없는 폴더 진입 시 경고 배너 표시 및 자동 롤백.
  - **명확한 삭제 확인 모달**: 파일명과 타입을 명시적으로 재확인 (`🗑️ 'production.db' (file) 영구 삭제하시겠습니까? [y/n]`).
  - **안전 종료 모달**: `[Esc]`, `[q]`, `[f]` 키 실수로 인한 오작동 방지.

### 5. 🔐 로컬 암호화 마스터 볼트 & 보안 관리 (`[p]` 탭 4)
- **Argon2id KDF + AES-256-GCM**: 서버 비밀번호, sudo 비밀번호, SSH 프라이빗 키 전체를 표준 암호화 알고리즘으로 안전하게 봉인.
- **SSH 프라이빗 키 매니저 (`[b]` 파일 탐색기)**: `~/.ssh/` 내의 `id_rsa`, `id_ed25519`, `.pem` 키를 직접 탐색하고 등록.
- **마스터 비밀번호 재암호화 (Rekeying)**: 설정 탭 `[4]`에서 기존 비밀번호를 변경하면서 저장된 모든 호스트 인증 정보를 안전하게 일괄 재암호화.
- **2단계 공장 초기화(Factory Reset) 보호**: 데이터 영구 삭제 전 마스터 비밀번호를 재검증하는 2FA 모달로 실수 방지.
- **Caps Lock 감지 배지**: 오타 방지를 위한 실시간 `[🔒 CAPS LOCK ON]` 경고 배너.

### 6. 🗄️ 데이터베이스 유지보수 & 감사 관리 (`[p]` 탭 4)
- **실시간 DB 진단**: SQLite 용량, 등록 서버 수, 텔레메트리 히스토리 레코드 수 실시간 확인.
- **보관 기간 설정 & 디스크 정리 (Vacuum)**: 만료된 메트릭 데이터 일괄 정리(7일, 14일, 30일) 후 `VACUUM`을 통해 즉각적인 디스크 용량 회수.
- **감사 데이터 내보내기 / 가져오기**:
  - **메트릭 CSV 익스포트**: 과거 CPU/RAM/Disk/Net 지표를 타임스탬프 CSV로 추출.
  - **호스트 JSON 백업 및 복원**: 등록된 서버 목록을 JSON으로 백업하고 중복 없이 안전하게 복원.

### 7. 🤖 AI 터미널 코파일럿 & 자율 진단 (`[F4]`)
- **다중 공급자 지원**: 초고속 무료 LPU 엔진인 **Groq Cloud API**, 로컬 오프라인 LLM인 **Ollama**, 그리고 **OpenAI** 지원 (공급자별 독립 프로필 캐싱).
- **상황 인지형 장애 진단**: 실시간 메트릭(CPU/RAM/Disk), 리눅스 배포판, 현재 경로, 직전 실패 명령어의 종료 코드 및 stderr를 AI 프롬프트에 자동으로 주입.
- **2단계 직관적 UX**: 장애 원인을 1~2문장으로 명쾌하게 설명하고, 즉각 조치할 수 있는 단일 실행 명령어를 제안.
- **원클릭 실행 & 탭 수정**: 제안된 명령어를 `[Enter]`로 즉시 실행하거나, `[Tab]`으로 콘솔 입력창에 복사하여 검토 후 실행.
- **Bash 스타일 히스토리 탐색 (`↑ / ↓`)**: 이전 AI 질의 및 제안 명령어를 방향키로 다시 호출.
- **2중 안전 가드레일**: 리눅스 운영 범위를 벗어난 질의 차단 및 파괴적 명령어(`rm -rf`, `shutdown` 등)의 클라이언트 단 하드코딩 차단.

### 8. 🌐 완벽한 다국어 지원 (i18n)
- 전체 UI 뷰, 팝업 모달, 런북, 에러 배너에 걸쳐 **한국어**, **English**, **Deutsch (독일어)** 100% 네이티브 지원.

### 9. 📖 모듈형 런북 & 단축키 치트시트 (`[?]`)
- **[1] ⌨️ 단축키 가이드**: 메인 콕핏, 원격 콘솔, SFTP 파일 매니저 전체 단축키 내장 안내.
- **[2]~[6] OS별 진단 런북**: Common Linux, Ubuntu, RHEL/Rocky, Alpine, Docker 환경에 최적화된 즉시 실행 점검 명령어 수록.

### 10. 🧪 오프라인 데모 모드 (`--demo`)
- `leitstand --demo` 명령어로 실제 SSH 연결 없이도 현실적인 가상 서버와 메트릭이 시뮬레이션되는 데모 환경 즉시 실행.

---

## ⌨️ 키보드 단축키 치트시트

### 🖥️ 메인 콕핏
| 단축키 | 기능 설명 |
|---|---|
| **`[↑/↓]`**, **`[j/k]`** | 서버 목록 탐색 |
| **`[Enter]`** | 서버 선택 및 원격 콘솔 활성화 |
| **`[Tab]`** | 서버 목록 ↔ 원격 콘솔 포커스 전환 |
| **`[a]`** / **`[e]`** / **`[x]`** | 서버 추가 / 서버 수정 / 서버 삭제 |
| **`[f]`**, **`[F6]`** | SFTP 듀얼 패널 파일 매니저 열기 |
| **`[F4]`** | AI 터미널 코파일럿 & 자율 진단 열기 |
| **`[t]`** | 풀스크린 인터랙티브 PTY 터미널 실행 |
| **`[Ctrl+O]`** | 풀스크린 콘솔 토글 |
| **`[?]`**, **`[Ctrl+K]`** | 빠른 명령어 런북 & 단축키 가이드 열기 |
| **`[p]`**, **`[,]`** | 환경설정 및 크리에이터 프로필 모달 열기 |
| **`[Ctrl+T]`** / **`[Ctrl+N]`** | 신규 콘솔 탭 생성 |
| **`[Alt+1]` ~ `[Alt+9]`** | 해당 탭으로 즉시 전환 |
| **`[Ctrl+W]`** | 현재 활성 콘솔 탭 닫기 |
| **`[Ctrl+C]`** | 활성 탭의 실행 중인 스트림 취소 |
| **`[q]`**, **`[Esc]`** | 프로그램 종료 |

---

### 📂 SFTP 듀얼 패널 파일 매니저 (`[f]`)
| 단축키 | 기능 설명 |
|---|---|
| **`[Tab]`**, **`[◄/►]`** | 활성 패널 전환 (로컬 PC ↔ 원격 서버) |
| **`[↑/↓]`**, **`[j/k]`** | 파일 / 폴더 목록 탐색 (1행씩) |
| **`[PgUp/PgDn]`**, **`[Ctrl+U/D]`** | 빠른 페이지 스크롤 (화면 단위 건너뛰기) |
| **`[Home/End]`**, **`[g/G]`** | 목록 최상단 / 최하단으로 즉시 이동 |
| **`[/]`** | 실시간 파일명 검색 / 필터링 |
| **`[Enter]`** | 디렉토리 열기 (권한 오류 시 자동 롤백) |
| **`[Backspace]`** | 상위 디렉토리로 이동 (`..`) |
| **`[Space]`** | 다중 선택 토글 배지 (`[*]`) |
| **`[a]`** | 전체 파일 선택 / 선택 해제 |
| **`[x]`**, **`[Ctrl+X]`** | **잘라내기 (이동 대기)**: 원하는 폴더로 이동 후 `[p]` 입력 |
| **`[c]`**, **`[Ctrl+C]`** | **복사 (복제 대기)**: 원하는 폴더로 이동 후 `[p]` 입력 |
| **`[p]`**, **`[Ctrl+V]`** | **붙여넣기 (드롭)**: 대기 중인 항목을 현재 폴더로 이동/복사 |
| **`[u]`** | 선택한 파일 업로드 (로컬 ➔ 원격) |
| **`[d]`** | 선택한 파일 다운로드 (원격 ➔ 로컬) |
| **`[:]`**, **`[!]`** | 현재 폴더 경로에서 인스턴트 셸 명령 실행 |
| **`[n]`** / **`[N]`** | 새 폴더 생성 (`mkdir`) / 새 파일 생성 (`touch`) |
| **`[r]`** | 선택한 파일 / 폴더 이름 변경 |
| **`[Delete]`**, **`[Shift+X]`** | 선택한 항목 영구 삭제 (확인 팝업) |
| **`[.]`** | 숨김 파일 표시 토글 (`.env`, `.*`) |
| **`[F5]`** | 디렉토리 목록 새로고침 |
| **`[?]`**, **`[F1]`** | 파일 매니저 전용 도움말 열기 |
| **`[Esc]`**, **`[q]`** | 클립보드 비우기 / 안전 종료 확인 팝업 |

---

## 📦 다운로드 & 즉시 실행 (Go 설치 불필요)

Go 언어나 별도의 개발 도구를 설치할 필요 없이, [`dist/`](dist/) 폴더에 미리 빌드된 단일 실행 압축 파일을 내려받아 즉시 실행할 수 있습니다:

| 플랫폼 (OS) | 아키텍처 | 다운로드 패키지 | 실행 방법 |
|---|---|---|---|
| **Windows** | x86_64 (`amd64`) | [**leitstand-v1.0.0-windows-amd64.zip**](dist/leitstand-v1.0.0-windows-amd64.zip) | 압축 해제 후 `leitstand-windows-amd64.exe` 더블 클릭 또는 터미널 실행 |
| **Linux** | x86_64 (`amd64`) | [**leitstand-v1.0.0-linux-amd64.tar.gz**](dist/leitstand-v1.0.0-linux-amd64.tar.gz) ([`.zip`](dist/leitstand-v1.0.0-linux-amd64.zip)) | `tar -xvf ... && ./leitstand-linux-amd64` |
| **macOS** | Apple Silicon (`arm64`) | [**leitstand-v1.0.0-darwin-arm64.tar.gz**](dist/leitstand-v1.0.0-darwin-arm64.tar.gz) ([`.zip`](dist/leitstand-v1.0.0-darwin-arm64.zip)) | `tar -xvf ... && ./leitstand-darwin-arm64` |

> [!TIP]
> 실제 서버 연결 없이 가상 콕핏을 체험하려면 데모 모드로 실행해 보세요:  
> `./leitstand --demo`

---

## 🛠️ 설치 및 크로스 플랫폼 빌드 가이드

### 사전 요구사항
- **Go 1.22 이상** 설치.
- 최신 ANSI/UTF-8 지원 터미널 (Windows Terminal, iTerm2, Alacritty, Kitty, GNOME Terminal 등).

### 1. 현재 운영체제용 기본 빌드
```bash
# 저장소 복제
git clone https://github.com/sorrynthx/leitstand.git
cd leitstand

# 단위 테스트 실행
go test -v ./...

# 실행 바이너리 빌드
go build -o bin/leitstand ./cmd/leitstand
```

### 2. 크로스 플랫폼 컴파일 (Zero-CGO Pure Go)
LEITSTAND는 순수 Go 기반 SQLite 드라이버를 탑재하여 외부 C 라이브러리 의존성이 전혀 없으므로(`CGO_ENABLED=0`), 별도의 크로스 툴체인 설치 없이 어디서나 다른 OS용 바이너리를 즉시 생성할 수 있습니다:

```bash
# 리눅스 (Linux amd64)
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o bin/leitstand-linux-amd64 ./cmd/leitstand

# 맥 (macOS Apple Silicon arm64)
CGO_ENABLED=0 GOOS=darwin GOARCH=arm64 go build -o bin/leitstand-darwin-arm64 ./cmd/leitstand

# 윈도우 (Windows amd64)
CGO_ENABLED=0 GOOS=windows GOARCH=amd64 go build -o bin/leitstand-windows-amd64.exe ./cmd/leitstand
```

### 3. 실행 방법
```bash
# 일반 모드
./bin/leitstand

# 오프라인 데모 모드 (실제 서버 없이 즉시 시연)
./bin/leitstand --demo
```

---

## 🏛️ 시스템 아키텍처

자세한 내부 동작 원리 및 시퀀스 다이어그램은 [docs/ARCHITECTURE.md](docs/ARCHITECTURE.md)에서 확인하실 수 있습니다.

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

## 👤 개발자 & 비전

- **제작자**: **김경곤 (Kyunggon Kim)**
- **비전**: 개발자와 엔지니어가 분산 서버 인프라를 지연 없이 직관적이고 경쾌하게 제어할 수 있는 의존성 없는 최고의 도구를 만듭니다.
- **GitHub 저장소**: [github.com/sorrynthx/leitstand](https://github.com/sorrynthx/leitstand)
- **라이선스**: MIT License

