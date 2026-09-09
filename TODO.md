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
- [x] **`[?]` 런북 내 `[7] ⭐️ Custom (내 명령어)` 탭 신설**:
  - 프로젝트 및 실무 전용 자주 쓰는 명령어들을 모아두는 나만의 런북 보관함.
  - 인앱 추가(`[a]`), 수정(`[e]`), 삭제(`[d]`) 및 방향키 이동 후 `Enter` 즉시 콘솔 입력 연동 완료.
- [x] **하이브리드 런북 아키텍처 (Built-in + SQLite Overlay)**:
  - 기존 OS별 기본 내장 런북(불변의 안정성) + 로컬 SQLite `custom_commands` 테이블의 사용자 명령어 투명 결합 완료.
  - 250줄 규칙 100% 준수 모듈 분리(`drawer_custom.go`, `drawer_custom_view.go`, `drawer_view_list.go`) 및 한/영/독 i18n 100% 일치.
- [x] **팀 런북 JSON Export & Import 파이프라인**:
  - 내가 작성한 명령어들을 `runbook_export_<timestamp>.json`으로 안전 내보내기(`[x]`).
  - `SetEscapeHTML(false)` 적용으로 `&, <, >, awk, 따옴표` 등 쉘 특수문자 왜곡/깨짐 0% 무결성 보장.
  - 팀원이 공유해 준 공통 런북 JSON 파일을 불러와 중복 없이 내 보관함에 안전 병합(`[i]`).
  - 특수문자 전용 단위 테스트(`TestStorageCustomCommandsExportImport`) 100% 통과.
- [x] **스마트 우선순위 탭 포커스 (Plan A)**:
  - 등록된 커스텀 명령어가 1개 이상 존재할 경우 런북 실행 시 `[7] ⭐️ Custom` 탭으로 즉시 자동 진입.
  - 커스텀 명령어가 0개일 경우 기존 호스트 OS 감지 탭(Ubuntu/RHEL/Alpine/Common)으로 자동 진입하여 편의성 극대화.

#### 🤖 Phase 5: AI 터미널 코파일럿 & 자율 진단 엔진 (AI Terminal Copilot - 100% 완료)
- [x] **인앱 인라인 AI 코파일럿 UI (`[F4]`)**:
  - 하단 인라인 AI 대화창 토글 및 렌더링 (`[F4]`, `[Esc]` 닫기 시 터미널 포커스 자동 복구).
  - 마크다운 스타일링 및 추천 명령어(`[Enter] 즉시 실행`, `[Tab] 입력창 복사`, `[s] 런북 저장`) 버튼 UI.
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
- [x] **클라우드 LLM API 연동 및 초고속 Groq LPU 무료 연동 파이프라인**:
  - Groq Cloud API(`https://api.groq.com/openai/v1`, `llama-3.3-70b-versatile`) 및 OpenAI 호환 엔드포인트 연동.
  - HTTP 401(인증 실패), 429(Rate Limit), 404(모델 부재) 정밀 에러 포맷팅 및 한국어 사용자 안내.
  - 환경설정(`[p] ➔ 5번 AI 탭`) Provider 프리셋(`groq`, `ollama`, `openai`, `custom`) 좌우 전환 시 엔드포인트/모델/플레이스홀더 자동 채움 및 무료 키 발급 힌트 배너 연동.
- [x] **서버 실시간 텔레메트리 & 최근 로그 컨텍스트 자동 주입 (Context-Aware Prompting)**:
  - AI 코파일럿(`[F4]`) 호출 시 현재 타깃 서버의 순간 CPU(%), RAM(사용량/총량/비율), 루트 디스크(사용량/총량/비율) 텔레메트리를 프롬프트에 자동 주입.
  - AI 인라인 타이틀바에 서버 실시간 자원 상태 배지(`• CPU: ... | RAM: ... | Disk: ...`) 동적 렌더링.
  - 직전 터미널 명령어, 종료 코드 및 에러 출력과 결합하여 맞춤형 장애 진단 명령어 추천 파이프라인 완성.
- [x] **커스텀 런북 저장 연계 (Phase 4-1 연계 완성)**:
  - AI가 진단 후 추천한 유용한 명령어를 `[Ctrl+S]` 단축키 1클릭으로 커스텀 런북(`custom_commands`)에 안전 저장.
  - 텍스트 입력창 타이핑 중 실수 저장 방지 및 중복 명령어 자동 감지/안내(`ℹ️ 이미 런북에 등록되어 있는 명령어입니다.`).
  - 질문 내용이 제목(Title)으로, AI 진단 요약이 설명(Description)으로 자동 매핑.
- [x] **앱 종료 안전 확인 다이얼로그 (Graceful Quit Confirmation)**:
  - `q` 또는 `Ctrl+C` 입력 시 즉시 종료되지 않고 `[y/Enter] 종료`, `[n/Esc] 취소` 2단계 확인 팝업 탑재.

### ⏳ Phase 5-1: 명령어 실시간 로딩 인디케이터 & 내장 런북 실무 대폭 강화 (100% 완료)
- [x] **명령어 비동기 실행 실시간 로딩 인디케이터 & 경과 시간 타이머 (`internal/tui`)**:
  - `Enter` 실행 즉시 터미널 뷰포트 및 상태바에 100ms 틱 회전 스피너(`⠋⠙⠹⠸⠼⠴⠦⠧⠇⠏`) 애니메이션 작동.
  - 0.1초 단위 실시간 경과 시간(`⏱️ 1.2s...`) 카운팅 및 장시간 명령(`du`, `find`, `curl`, `apt`) 실행 시 시각적 안도감 제공.
  - 실행 완료 시 `[⏱️ 2.4s 소요]` 메타데이터 자동 삽입 및 초록색 성공 상태바 배너 동적 연동.
- [x] **공통 및 배포판별 실무 필수 런북 대폭 보강 (`internal/quickcmd/tab_*.go`)**:
  - **자원 & 긴급 장애**: `OOM-Killer` 처형 기록 검색, `좀비 프로세스` 색출, `스레드 다배체 Top 5`.
  - **스토리지 & 디스크**: `50MB+ 대용량 파일 Top 10`, `삭제되었으나 프로세스가 잡고 있는 유령 파일(lsof +L1)`, `/var` 디렉토리 용량 분석.
  - **네트워크 & 웹**: `ESTABLISHED 접속 IP 순위 Top 10`, `curl 응답 지연 구간(DNS/TCP/TLS/TTFB) 분해`, `SSL/TLS 인증서 만료일`.
  - **보안 & 감사**: `실시간 접속자(w)`, `최근 로그인 이력(last 10)`, `SSH 공격 실패 IP 순위 Top 10`.
  - **Docker 컨테이너 실무**: `비정상 종료(Exited) 컨테이너 확인`, `최근 에러 로그 필터`, `컨테이너 내부 IP 조회`, `시스템/볼륨 안전 정리(prune)`.
  - **배포판별 시스템 도구**: `부팅 지연 서비스 분석(systemd-analyze blame)`, `최근 1시간 시스템 저널`, `패키지 캐시 청소(clean)`.
- [x] **전역 다국어 사전(KO, EN, DE) 100% 동기화 (0 Missing Keys)**:
  - 7대 런북 카테고리 헤더 및 모든 신규/기존 명령어 번역 전수 등록 및 `TestDictionaryParity` 통과.

### 🍺 Phase 5-2: 이스터에그 암호화 프로필 & 50:50 맥주 아스키 아트 About 개편 (100% 완료)
- [x] **난독화 분산 키 기반 이스터에그 암호화 프로필 파이프라인 (`internal/profile`)**:
  - 공개 깃허브 크롤러/스팸 봇 방지를 위해 `profile.enc` Base64 AES-256-GCM 암호화 파일 분리.
  - 코드 상 키 변수명 위장(`layoutGlyphMetrics*`) 및 바이트 분산 난독화 적용.
- [x] **언어별(KO, EN, DE) 맞춤 50:50 분할 레이아웃 (`settings_modal_view_about.go`)**:
  - 상단: 시원한 거품이 넘치는 독일 바이젠 맥주(Bier Mug) ASCII 아트 렌더링.
  - 좌측 (50%): 삶과 사람에 대한 이야기 (암으로 세상을 떠난 이에 대한 추모, 후회 없는 삶, 'No Problem', 거절의 의미, Retry.Day).
  - 우측 (50%): 기술 철학(Zero-Agent, Pure Go, Zero-Knowledge), 공통 링크(retry.day, GitHub, LinkedIn), 및 언어별 맞춤 메시지:
    - 한국어: `🤝 문제 해결 & 팀 동료 (도메인의 본질적 문제를 깊이 파악하고 함께 푸는 동료)`
    - 영어: `🤝 Problem Solver & Trusted Teammate`
    - 독일어: `🇩🇪 Karriere & Vor-Ort-Team in Deutschland (학센과 바이젠 맥주를 함께 즐길 현지 팀 이직 제안 환영 🍺)`
- [x] **250줄 모듈화 규칙 100% 준수**:
  - `types.go`(29줄), `cipher.go`(65줄), `profile.go`(76줄), `profile_test.go`(25줄), `settings_modal_view_about.go`(123줄).

### 🛡️ Phase 5-3: 공장 초기화(Factory Reset) 2차 패스워드 검증 및 핫 리로드 (100% 완료)
- [x] **공장 초기화 시 마스터 비밀번호 2차 검증 모달 (`settings_modal_reset.go`)**:
  - `[p]` ➔ 4번 Database ➔ `[6] Factory Reset` 시 단순 Enter 실수 방지를 위한 전용 팝업 모달 탑재.
  - 마스터 비밀번호 입력 대조 검증 및 불일치 시 붉은색 경고 차단 (`⚠️ 마스터 비밀번호가 올바르지 않습니다`).
  - Caps Lock 감지 배지 및 `[Esc]` 안전 취소 지원.
- [x] **초기화 집행 시 `vault_meta` 완전 삭제 및 초기 비밀번호 설정 화면 전환**:
  - `vault_meta` 테이블까지 영구 삭제하여 완전 무결한 클린 초기 상태 복원.
  - 인메모리 호스트/탭/세션/SSH풀/터널 즉시 파기 및 앱 최초 실행 마스터 비밀번호 생성(`VaultModalInit`) 화면으로 즉시 전환.
- [x] **JSON 호스트 임포트 핫 리로드**:
  - 4번 탭에서 호스트 JSON 복원 시 메인 콕핏 복귀와 동시에 인메모리 호스트 목록 즉시 갱신(`m.loadHostsCmd()`).
- [x] **250줄 엄격 모듈화 & 다국어 Parity 100% 통과**:
  - `settings_modal_reset.go`(95줄), `update_modal_settings.go`(54줄) 분리 및 `dict_*.go` 전수 동기화.

---

## 🎬 3. Video Showcase & Social Media Launch Checklist (영상 촬영 및 SNS 릴리즈 기획)

### 📹 Video Showcase Guidelines (영상 촬영 가이드)
- [ ] **영상 규격**: 60초 ~ 90초 내외 (LinkedIn & Threads 숏폼 포맷 최적화, 1080p 60fps).
- [ ] **Scene 1 (00:00~00:08) - Intro & Vault Unlock**:
  - 터미널에서 `./leitstand` 실행 ➔ 암호화 보관함 언락 (Caps Lock 배지 노출).
  - *"Zero-Agent, 100% Pure Go 터미널 콕핏"*
- [ ] **Scene 2 (00:08~00:20) - Live Cockpit & Telemetry**:
  - 방향키 탐색 후 `Enter` 즉시 연결 ➔ `F5` 텔레메트리 덱 토글 (실시간 CPU/RAM/Disk/Net 게이지).
  - *"원격 서버 실시간 자원 상태 한눈에 파악"*
- [ ] **Scene 3 (00:20~00:35) - Multi-Tab Shell & Spinner**:
  - `Ctrl+N` 새 탭 생성 ➔ 대용량 검색 or `du` 실행 ➔ **100ms 틱 브레일 스피너 + 경과 시간(`⏱️ 1.4s...`)**.
  - *"독립 세션 멀티탭 + 비동기 실행 스피너"*
- [ ] **Scene 4 (00:35~00:50) - SFTP Dual-Pane & In-App Editor**:
  - `[f]` 키로 90% 2분할 SFTP 진입 ➔ 원격 설정파일 `Enter`로 열어 수정 후 `F2` 저장 (선명한 성공 배너).
  - *"별도 도구(FileZilla 등) 없는 인앱 파일 탐색 & 실시간 편집"*
- [ ] **Scene 5 (00:50~01:05) - AI Terminal Copilot**:
  - `[F4]` AI 코파일럿 호출 ➔ 장애 진단 질문 ➔ 추천 명령어 `[Enter]` 즉시 콘솔 실행 ➔ `[Ctrl+S]`로 런북 원클릭 저장.
  - *"서버 텔레메트리 연동 AI 자율 진단 & 나만의 런북 보관"*
- [ ] **Scene 6 (01:05~01:15) - SSH Port Forwarding & Runbooks**:
  - `[?]` 런북 7대 카탈로그 ➔ `[T]` 원클릭 SSH 터널링 (사설 DB/Docker 로컬 바인딩).
  - *"클릭 한 번으로 끝나는 사설 포트포워딩"*
- [ ] **Scene 7 (01:15~01:25) - Outro & German Beer ASCII**:
  - 설정(`[p]`) ➔ 6번 탭 전환 ➔ **시원한 바이젠 맥주 ASCII 아트** + Retry.Day 및 개발자 철학 안내.
  - *"독일 현지 팀 이직 제안 및 피드백 환영 🍺"*

### 📱 Social Media Posting Checklist (SNS 업로드 준비)
- [ ] **LinkedIn 포스팅**:
  - 기술 스택(Pure Go, TUI, Bubbletea, SSH Mux) 및 아키텍처 중심 릴리즈 노트 공유.
  - 링크: GitHub 저장소 + Retry.Day 블로그 링크.
- [ ] **Threads 포스팅**:
  - 가볍고 임팩트 있는 비디오 클립 + 핵심 기능 요약 (에이전트 0%, 맥주 아스키).

---

## 📦 4. Distribution & Release (배포 파이프라인)
- [x] **크로스 플랫폼 무의존성(Zero-CGO Pure Go) 단일 바이너리 빌드**:
  - Windows (`leitstand-windows-amd64.exe`), Linux (`leitstand-linux-amd64`), macOS (`leitstand-darwin-arm64`) 컴파일 완료.
- [x] **즉시 실행용 배포 아카이브 패키징 (`dist/`)**:
  - `leitstand-v1.0.0-windows-amd64.zip` (압축률 50%, 약 14.2MB)
  - `leitstand-v1.0.0-linux-amd64.tar.gz` & `.zip` (약 14.0MB)
  - `leitstand-v1.0.0-darwin-arm64.tar.gz` & `.zip` (약 13.5MB)
  - `dist/README.md` 다운로드 및 즉시 실행 가이드 탑재.
  - 3개 국어 README(`README.md`, `README.ko.md`, `README.de.md`)에 OS별 원클릭 다운로드 표 반영.
- [ ] **GitHub Release 연동**: 저장소 릴리즈 시 `dist/` 아카이브 바이너리 에셋 등록.

---

*Last Updated: 2026-09-09 (Phase 5-3 Factory Reset 2FA Completed, Video & SNS Launch Planning Finalized)*


