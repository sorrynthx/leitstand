# ⚡ LEITSTAND - Project Roadmap & Master TODO

> **Modern, Agentless, High-Performance Terminal Server Cockpit & Telemetry Engine**  
> *Developed by Kyunggon Kim (김경곤 / Interpass Inc.)*

---

## 🌟 1. Completed Milestones (완료된 기능 및 아키텍처)

### 🏗️ Architecture & Refactoring (아키텍처 및 소스 품질)
- [x] **250줄 규칙 100% 준수 모듈화**: 모든 로직 및 UI 소스 코드는 250줄 이하로 엄격 분리 유지. *(단, 향후 명령어 카탈로그 및 다국어 번역의 지속적인 확장이 발생하는 순수 데이터 정의 파일인 `internal/quickcmd/tab_*.go` 및 `internal/i18n/dict_*.go`는 250줄 제한에서 명시적 예외로 관리)*
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
- [x] **SFTP 파일 업로드/다운로드 교차 검증**:
  - `sftp.go` & `sftp_transfer.go` 분리 후 로컬 ↔ 원격 대용량 파일 전송(`[F5]`) 및 삭제 정밀 테스트 완료.
- [x] **SFTP 파일/폴더 조작 기능 실물 검증**:
  - 새 폴더 생성(`[n]`), 빈 파일 생성(`[N/t]`), 이름 변경(`[r]`), 삭제(`[Delete]`), 숨김파일 토글(`[.]`) 실서버 동작 확인 완료.
  - 대량 파일 디렉토리 고속 탐색을 위한 `PgUp` / `PgDn` (10개 단위 점프), `Home` / `End` 즉시 이동 지원 탑재.
- [x] **클립보드 잘라내기/복사/붙여넣기 테스트**:
  - `[x]`(잘라내기) / `[c]`(복사) ➔ 경로 이동 ➔ `[p]` 또는 `[v]`(붙여넣기) 단축키 표준화 적용 완료 및 실물 검증 완료.
- [x] **`GetSFTPClient()` 커넥션 안정성 & Keep-Alive 적용**:
  - OpenSSH 표준 30초 주기 Keep-Alive 핑 엔진 탑재로 유휴(Idle) 세션 강제 차단 원천 방어.
  - 단일 SFTP 캐싱(`GetSFTPClient`) 실연동으로 폴더 이동 시 세션 재생성 렉 및 `MaxSessions 10` 고갈 방지.
  - 네트워크 단선/지연 시 `ResetSFTPClient()` 투명 1회 자동 재연결(Auto-Retry) 파이프라인 탑재.

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
- [x] **콘솔 세션 로그 로컬 저장 (`Ctrl+E`)**:
  - ANSI 제어 문자 자동 정제(Clean Plain Text) 및 감사 헤더 삽입 파이프라인(`sessionlog`).
  - 콘솔 탭에서 `Ctrl+E` 입력 시 `session_<host>_<timestamp>.log` 즉시 생성 및 성공 토스트 배너 연동.
  - 환경설정(`[p]`) 내 `[3] 📜 Logs` 전용 탭 개편 (OS 기본 문서 폴더, 로컬 폴더, 홈 폴더, 사용자 지정 경로 프리셋 선택 및 SQLite 영구 저장).
- [x] **내부 SQLite DB 관리기 및 보안 유지보수 (환경설정 `[4]` 탭 일원화)**:
  - 텔레메트리 메트릭 자동 보관 주기 조절 (7일 권장 / 14일 / 30일) 및 상단 뱃지 렌더링.
  - `[1] 🧹 메트릭 정리 & DB 디스크 압축 (Prune & VACUUM)` 1클릭 최적화 및 전후 용량 변화 배너.
  - `[2] 📊 텔레메트리 메트릭 CSV 보고서 백업 추출` (디렉토리 자동생성 및 안전 GUI 탐색기 연동).
  - `[3] 📤 서버 목록 백업 (Export JSON)` 및 `[4] 📥 서버 목록 복원 (Import JSON)` (중복 방지 안전 병합).
  - `[5] 🔑 마스터 비밀번호 변경 (Rekey Vault)` 및 전체 서버 크리덴셜 안전 재암호화 (일반설정 중복 제거 후 4번 탭 일원화).
  - `[6] ⚠️ 보관함 및 설정 초기화 (Factory Reset)` 안전 확인 팝업 탑재. *(※ 실서버 데이터 삭제 테스트는 Phase 4 포트포워딩 개발 후 진행)*
  - 탭별 포커스 네비게이션 엔진(`settings_modal_nav.go`) 신설로 1번 탭 유령 필드 버그 및 4번 탭 이동 충돌 완벽 해결.

### 🚇 Phase 4: SSH 포트 포워딩 & 터널링 매니저 (SSH Tunneling)
- [x] **SSH 로컬 포트 포워딩 엔진 (`internal/ssh/tunnel.go`, `internal/ssh/tunnel_manager.go`)**:
  - 원격 서버 내부 사설 DB(MySQL 3306, PostgreSQL 5432, Redis 6379) 또는 비공개 웹 포트(8080, 5678, 11434)를 내 PC 로컬 포트로 암호화 터널링.
  - `net.Listen` 로컬 소켓 바인딩 및 SSH 채널 `Dial` 양방향 고속 스트리밍(`io.Copy`).
  - 활성 커넥션 수(atomic counter) 추적 및 리소스 누수 없는 안전한 수명주기 해제.
- [x] **터널링 관리 모달 UI (`internal/tui/tunnel_modal*.go`)**:
  - `T` (`Shift+T`) 및 `F7` 전역 단축키 호출 지원 (콘솔 입력 중에도 `F7` 전역 감지).
  - 로컬 바인딩 포트 및 원격 대상 포트 입력/추가/삭제 폼.
  - 활성 터널 실시간 상태 표시 (`🟢 ON / 🔴 OFF`) 및 `Space` / `Enter` 원클릭 시작/중지 토글.
  - 상단 헤더(`LIVE ENGINE` 옆) 활성 터널 카운트 & 포트 요약 초록색 뱃지 실시간 동적 연동.
  - 하단 상태바 `[T/F7] 터널` 키 가이드 상시 노출.
  - SQLite 영구 저장(`ssh_tunnels`)으로 앱 재실행 시에도 등록된 터널링 규칙 보존.
- [x] **실서버 및 Docker 컨테이너 실물 검증**:
  - 테스트 서버 192.168.14.119 내 Docker 컨테이너(n8n 5678, Ollama 11434, MySQL 3306) 구동 후 로컬 PC 브라우저/도구 접속 100% 동작 확인 완료.
  - 2단계 삭제 확인 다이얼로그(`[d]` -> `[Enter/y]` 확인, `[Esc/n]` 취소) 탑재.
  - 전역 다국어(KO, EN, DE) 전수 점검 및 AST 정밀 스캔을 통한 잔존 하드코딩 0개(100% Zero-Hardcoding) 달성.

### ⭐️ Phase 4-1: 커스텀 런북 & 팀 런북 JSON 확장 (Custom Runbooks & JSON Extension)
- [ ] **`[?]` 런북 내 `[7] ⭐️ Custom (내 명령어)` 탭 신설**:
  - 프로젝트 및 실무 전용 자주 쓰는 명령어들을 모아두는 나만의 런북 보관함.
  - 인앱 추가(`[a]`), 수정(`[e]`), 삭제(`[d]`) 및 방향키 이동 후 `Enter` 즉시 콘솔 입력.
- [ ] **하이브리드 런북 아키텍처 (Built-in + SQLite Overlay)**:
  - 기존 OS별 기본 내장 런북(불변의 안정성) + 로컬 SQLite `custom_commands` 테이블의 사용자 명령어 투명 결합.
- [ ] **팀 런북 JSON Export & Import 파이프라인**:
  - 내가 작성한 명령어들을 `my_runbook.json`으로 내보내기(Export).
  - 팀원이 공유해 준 공통 런북 JSON 파일을 불러와 중복 없이 내 보관함에 안전 병합(Import).

### 🤖 Phase 5: AI 터미널 코파일럿 & 자율 진단 엔진 (AI Terminal Copilot - 50% 진행)
- [x] **인앱 인라인 AI 코파일럿 UI (`[F4]`)**:
  - 하단 인라인 AI 대화창 토글 및 렌더링 (`[F4]`, `[Esc]` 닫기 시 터미널 포커스 자동 복구).
  - 마크다운 스타일링 및 추천 명령어(`[Enter] 즉시 실행`, `[Tab] 입력창 복사`) 버튼 UI.
- [x] **로컬 Ollama 스트리밍 엔진 & 최근 컨텍스트 주입 (`internal/ai`)**:
  - 로컬 Ollama REST API(`http://localhost:11434/api/chat`) 순수 Go `net/http` 실시간 SSE 스트리밍.
  - 대화 내역(`ai_chat_history`) SQLite 영구 저장 및 최근 2턴 컨텍스트 격리 주입 (과거 명령 혼동 차단).
- [x] **치명적 위험 명령 원천 차단 안전 가드 (`ai_safety.go`)**:
  - 시스템 재부팅, 종료, 포맷, `swapoff`, 대상 경로 없는 bare `rm` (`rm`, `rm -f`) 자동 실행 원천 차단.
- [x] **AI 추천 명령어 원클릭 주입 & 터미널 포커스 인계 루프**:
  - `[Enter]` 실행 직후 원격 터미널(`PaneConsole`) 입력창으로 포커스 자동 인계 및 명령 실행.
- [x] **시스템 프롬프트 외부 파일 분리 (`~/.leitstand/copilot_system_prompt.txt`)**:
  - 코드 재빌드 없이 사용자가 직접 수정 가능한 프롬프트 템플릿 파일 로더 및 자동 생성기.
- [x] **단축키 모드 한글 IME 및 독일어 특수문자 안내 엔진 (`update_ime.go`)**:
  - 탐색 모드에서 한글(자모/음절) 및 독일어 Umlaut/특수문자(`ä, ö, ü, ß`) 감지 시 상태바 전환 안내 경고 출력.
- [ ] **클라우드 LLM API 키 연동 및 외부 API 테스트 (Next)**:
  - 외부 API 대응: OpenAI / Claude / Gemini API 키 입력 지원 및 실서버 진단 테스트.
- [ ] **서버 텔레메트리 & 최근 로그 컨텍스트 자동 주입 (Context-Aware Prompting)**:
  - 질문 전송 시 현재 서버 OS, CWD, CPU/RAM/디스크 사용량, 직전 터미널 에러를 시스템 프롬프트에 자동 첨부.
- [ ] **커스텀 런북 저장 연계**:
  - AI가 추천한 유용한 명령어를 커스텀 런북(`custom_commands`)으로 원클릭 저장.

### 📦 Phase 6: 크로스 플랫폼 자동 빌드 & 글로벌 배포 (Distribution & GoReleaser)
- [ ] **크로스 플랫폼 자동 빌드 파이프라인 (GoReleaser)**:
  - Windows (.exe), macOS (Apple Silicon M-series / Intel), Linux x86_64 바이너리 패키징.
  - GitHub Release 자동 연동 및 단일 실행 파일 릴리스.
- [ ] **Phase 3-2 공장 초기화 (Factory Reset) 최종 실물 검증**:
  - 최종 릴리스 전, 4번 탭 `[6] Factory Reset` 2단계 확인 팝업 및 SQLite DB 완전 초기화 최종 점검.

---

*Last Updated: 2026-09-04 (Phase 5 AI Copilot 50% Milestone & IME Navigation Guard Completed)*
