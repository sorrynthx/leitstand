# ⚡ LEITSTAND - Project Roadmap & Master TODO

> **Modern, Agentless, High-Performance Terminal Server Cockpit & Telemetry Engine**  
> *Developed by Kyunggon Kim (김경곤 / Interpass Inc.)*

---

## 🌟 1. Completed Milestones (완료된 기능 및 아키텍처)

### 🏗️ Architecture & Refactoring (아키텍처 및 소스 품질)
- [x] **250줄 규칙 100% 준수 모듈화**: `internal/` 하위 58개 전체 소스 코드가 모두 247줄 이하로 완벽 분리 및 정리.
- [x] **Zero-CGO Pure Go**: CGO 없이 100% 순수 Go 빌드 지원 (Windows, macOS, Linux 크로스 컴파일 호환).
- [x] **무중단 SSH 커넥션 풀링 (`internal/ssh`)**: 호스트별 1개의 SSH TCP 커넥션 재사용 및 채널 멀티플렉싱.
- [x] **SFTP 채널 고갈 방지 (`internal/ssh/client.go`)**: Thread-Safe `GetSFTPClient()` 단일 SFTP 클라이언트 캐싱으로 OpenSSH `MaxSessions 10` 세션 누수 에러 영구 해결.
- [x] **명시적 접속 (Explicit Connection) 아키텍처**: 서버 목록 탐색(`↑`/`↓`) 시 자동 SSH 접속/폴링을 유예하고, `Enter` / `r` / `f` / `t` 입력 시에만 선택된 서버 세션을 수립하여 예기치 않은 접속 오버헤드 및 로그 누수 차단.

### 🌐 Global i18n & Zero Hardcoding (다국어 및 중앙 제어)
- [x] **3개 국어 실시간 다국어 엔진 (`internal/i18n`)**: 한국어(KO), 영어(EN), 독일어(DE) 100% 지원.
- [x] **115개 전체 UI 키 전수 보관 (0 Missing Keys)**: `dict_ko.go`, `dict_en.go`, `dict_de.go`로 모듈화 분리 및 하드코딩 0개 달성.
- [x] **동적 실시간 렌더링**: 설정 모달(`[p]`) 및 온보딩 모달에서 언어 변경 시 `m.updateViewportContent()` 호출로 화면 전체 즉시 재렌더링.

### 🔒 Security & Local Encrypted Vault (보관함 및 암호화)
- [x] **Argon2id KDF + AES-256-GCM 로컬 암호화 보관함 (`internal/vault`, `internal/storage`)**: SQLite 내 암호화 영구 저장.
- [x] **인공지능 비밀번호 안전 보조**: 비밀번호 입력창 실시간 Caps Lock 켜짐 감지 (`🔒 CAPS LOCK 키가 켜져 있습니다`) 및 한글/Non-ASCII 감지 경고 배지.
- [x] **SSH Private Key 인증 & 인앱 키 탐색기 (`[b]`)**: `~/.ssh/` 디렉토리 내 `.pem`, `id_rsa`, `id_ed25519` 키 파일 자동 탐색 및 암호화 보관.

### 💻 Remote Console, Shell Multiplexing & Root Elevation (콘솔 및 권한 엔진)
- [x] **멀티 탭 인터랙티브 쉘 엔진 (`internal/tui/tab.go`)**:
  - `Ctrl+N` (새 탭), `Ctrl+W` (탭 닫기), `Alt+1`~`Alt+9` (탭 전환).
  - 탭별 독립 CWD(작업 디렉토리), 히스토리(`↑/↓`), 스크롤 위치 및 미완성 입력 문구 완전 격리 보존.
- [x] **`su root` 및 `sudo` 최고 권한 승격 엔진 (`internal/tui/console_elevation.go`)**:
  - MobaXterm과 100% 동일하게 `su root` / `sudo -i` 시 root 패스워드 인증 후 `root@host:#` 세션 유지.
  - 비밀번호 기억 옵션 미체크 시에도 현 세션 동안 `sudoModeCache` / `sudoCache` 자동 보존.
- [x] **스마트 경로 자동완성 엔진 (`Tab`)**: SFTP `ReadDir` + SSH `ls` 듀얼 파이프라인, 공통 접두어 자동채움 및 다중 후보 힌트 출력.
- [x] **`top` / `htop` 라이브 스트리밍 무멈춤 아키텍처**: 모달(에디터/설정/보관함) 실행 중에도 백그라운드 스트리밍 틱 pass-through 보적으로 모달 복귀 시 `top` 무멈춤 렌더링.
- [x] **네이티브 풀스크린 PTY 터미널 (`[t]`)**: 전체화면 대화형 SSH 터미널 진입 및 `exit` 안전 복귀.

### 📂 SFTP File Manager & In-App Text Editor (파일 매니저 및 에디터)
- [x] **90% 광폭 2분할 TUI 파일 매니저 (`[f]`, `[F6]`)**: 로컬 ↔ 원격 2분할 탐색, `PageUp/PageDown`, `/` 파일 검색 필터.
- [x] **클립보드 잘라내기/복사/붙여넣기 (`[x]` 잘라내기 ➔ 폴더 이동 ➔ `[p]` 붙여넣기)**: 일괄 이동(`mv`) 및 복사(`cp`).
- [x] **다중 선택 & 배치 전송**: `Space` 다중 선택, `[u]` 업로드, `[d]` 다운로드, `[n]` mkdir, `[N]` touch, `[r]` rename.
- [x] **내장 텍스트 에디터 (`EditorModal`) & 만능 저장 단축키**:
  - 윈도우 OS 가로채기 방지 만능 저장 키 **`F2`**, **`Alt+S`**, **`Ctrl+S`** 지원.
  - 저장 성공 시 타임스탬프가 표기된 **`✅ File saved successfully at 16:11:50! (Remote file updated)` 선명한 초록색 성공 배너** 출력.
  - 권한 부족 보호 파일 편집 시 Root 승격 세션이면 Root Fallback 쓰기 (`cat << 'EOF' > ...`) 자동 작동.

---

## 🎯 2. Upcoming Priority Tasks (다음 진행할 핵심 개발 리스트)

### 📂 Phase 3-0: SFTP 파일 매니저 리팩토링 후 전체 기능 점검 & 통합 테스트 (SFTP Validation)
- [ ] **SFTP 파일 업로드/다운로드 교차 검증**:
  - `sftp.go` & `sftp_transfer.go` 분리 후 로컬 ↔ 원격 파일 전송(`[F5]`) 및 삭제(`[F8/d]`) 정밀 테스트.
- [ ] **SFTP 파일/폴더 조작 기능 실물 검증**:
  - 새 폴더 생성(`[n]`), 빈 파일 생성(`[N]`), 이름 변경(`[r]`), 삭제(`[Delete]`), 숨김파일 토글(`[.]`) 동작 확인.
- [ ] **클립보드 잘라내기/복사/붙여넣기 테스트**:
  - `[x]`(잘라내기) / `[c]`(복사) ➔ 경로 이동 ➔ `[p]`(붙여넣기) 배치 처리 정상 여부 점검.
- [ ] **`GetSFTPClient()` 단일 SFTP 커넥션 안정성 확인**:
  - 파일 전송 도중 세션이 끊기거나 SSH 세션 타임아웃 없이 안정적으로 유지되는지 확인.

### 📊 Phase 3-1: 텔레메트리 (Telemetry) 성능 측정 및 시각화 고도화
- [x] **실시간 메트릭 수집기 (Telemetry Collector) 튜닝**:
  - `/proc/stat` 0.1초 이중 샘플링 델타 파이프라인(`ParseDualProcStat`) 구현으로 `top`/`htop`과 100% 동일한 순간 CPU 점유율 측정.
  - `/proc/meminfo`, `df -k /`, `/proc/net/dev` 파싱 정밀화.
- [x] **텔레메트리 패널 (`view_hostlist.go`) 초슬림 시각화**:
  - `F5` 전역 단축키 수용 (우측 콘솔 입력 중에도 전역 토글).
  - 우측 인터랙티브 터미널 패널 100% 온전 보존.
  - 초슬림 2줄 수직 스택 프로그래스 바 및 Uptime/용량 단축 렌더링.
- [x] **서버 자원 위험 임계치 알림 및 환경설정(`[p]`) 연동**:
  - 환경설정 모달(`[p]`) 내 `[2] 📊 Telemetry` 전용 탭 개편 (수집 주기 & CPU/RAM/Disk 경고 임계치 수치 조절).
  - 사용자 지정 임계치(Configured Thresholds) 초과 시 서버 탐색기 및 텔레메트리 콕핏에 `🔥 OVERLOAD` / `🔥 DANGER` / `⚠️ HIGH` 배지 실시간 동적 발동.
  - 숫자가 아닌 입력 시 강력한 예외 검증(Validation) 및 SQLite DB (`app_settings`) 영구 저장 연동.

### 📜 Phase 3-2: 세션 감사 로그 및 SQLite 관리 (Audit & Maintenance)
- [ ] **콘솔 세션 로그 로컬 저장 (`Ctrl+E`)**:
  - 콘솔 출력 및 명령어 수행 기록을 `logs/leitstand_<host>_<timestamp>.log` 파일로 백업.
- [ ] **내부 SQLite DB 관리기 (환경설정 안)**:
  - DB 백업(JSON 내보내기) 및 복원.
  - DB 최적화 (SQLite Vacuum) 및 마스터 보관함 초기화 (Reset).

### 🚇 Phase 4: SSH 터널링 및 글로벌 배포 (Tunneling & Distribution)
- [ ] **SSH 포트 포워딩 / 터널링 매니저 (SSH Tunneling)**:
  - 원격 서버 내부 DB(MySQL 3306, Redis 6379)를 로컬 포트로 포워딩.
- [ ] **크로스 플랫폼 자동 빌드 파이프라인 (GoReleaser)**:
  - Windows (.exe), macOS (Apple Silicon / Intel), Linux 바이너리 패키징.

---

*Last Updated: 2026-09-02*
