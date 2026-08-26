# ⚡ LEITSTAND - Project Roadmap & TODO

> **Modern, Ultra-lightweight & Zero-dependency Terminal Server Cockpit**  
> *Developed by Kyunggon Kim (김경곤)*

---

## 🌟 1. Completed Milestones (완료된 기능들)

- [x] **Core Architecture & TUI Cockpit**:
  - [x] Charmbracelet Lipgloss & Bubbletea 기반 3분할 콕핏 (Host Explorer / Telemetry / Remote Console)
  - [x] 무중단 SSH 연결 풀링 (`internal/ssh`) & 비동기 텔레메트리 파이프라인
  - [x] 데모 모드 (`--demo`) 지원 (오프라인/쇼케이스용 가상 메트릭 생성)
- [x] **Security & Master Vault**:
  - [x] Argon2id KDF + AES-256-GCM 마스터 금고 (`internal/vault`)
  - [x] 서버 비밀번호 및 시크릿 영구 암호화 보관
  - [x] 마스터 비밀번호 재암호화(Rekeying) 기능
- [x] **Smart Terminal UX & Remote Operations**:
  - [x] 원격 파일/경로 탭 자동완성 (`Tab` in Console)
  - [x] 인앱 파일 편집기 (`edit <file>` / `vi <file>` ➡️ Textarea Modal) & 원격 저장 (`Ctrl+S`)
  - [x] Chroma 터미널 문법 하이라이팅 (`cat`, `tail`, `head`)
  - [x] 원격 전체화면 콘솔 (`Ctrl+O`) & 복귀 (`Esc`)
- [x] **Global i18n & Preferences (`internal/i18n`)**:
  - [x] 3개 국어 실시간 다국어 지원 (**English** [기본값], **한국어**, **Deutsch**)
  - [x] 최초 온보딩 금고 초기화 언어 선택 바 (`[F1]~[F3]`)
  - [x] 풀스크린 환경설정 모달 (`[p]` / `[,]`): 언어 변경, 폴링 주기 변경, 마스터 비밀번호 변경
  - [x] 다국어 지원 **About Creator & Project** 프로필 탭 (비전, 독일 진출 비전, 소셜 링크)
  - [x] 설정값 내부 SQLite DB (`app_settings`) 영구 저장 및 자동 복원
- [x] **Telemetry & Network Rate Optimization**:
  - [x] 텔레메트리 DB 저장 제거 (Zero DB 오버헤드 / 메모리 실시간 뷰)
  - [x] 선택된 활성 서버 1대만 온디맨드 0.1초 즉시 폴링 (네트워크 부하 0.01% 미만)
  - [x] 텔레메트리 주기 설정 (`5s`, `10s`, `30s`, `60s`, `Off`) ➡️ `Off` 시 콘솔 풀사이즈 확장
  - [x] 실시간 초당 네트워크 전송 속도 계산 (B/s, KB/s, MB/s 자동 단위 변환)
- [x] **Safety & Input Assistance**:
  - [x] 비밀번호 입력 화면(로그인, 설정, 서버 등록) 실시간 **Caps Lock 켜짐 (`⇪ Caps Lock ON`)** 감지 배지
  - [x] 한글/비-ASCII 입력 시 실시간 경고 배지 (`⚠️ 한글 입력 감지됨! [한/영] 전환 필요`)
- [x] **OS-Aware Quick Command Runbook (`internal/quickcmd`)**:
  - [x] 접속된 서버 OS(Ubuntu/Debian, RHEL/Rocky, Alpine, Docker, Common) 자동 감지 및 기본 탭 포커스
  - [x] 단축키 (`[?]`, `[Ctrl+K]`, `[F5]`) 지원 (전체화면 콘솔 및 일반 콘솔 어디서든 호출 가능)
  - [x] 선택 시 콘솔 입력창에 안전하게 자동 입력 (`Auto-fill & Review`) 후 사용자 실행

---

## 🎯 2. Upcoming Roadmap (남은 작업 TODO)

### 🔑 Phase 1: 실무 인증 & 파일 전송 (High Priority)
- [ ] **SSH Private Key 파일 인증 지원**:
  - [ ] 호스트 등록 모달(`[a]`)에서 `[Password]` ↔ `[Private Key (.pem, id_rsa, id_ed25519)]` 선택
  - [ ] 로컬 키 파일 경로 브라우징 및 키 Passphrase 지원
  - [ ] Private Key 내용 마스터 금고에 암호화 보관
- [ ] **SFTP 양방향 파일 전송 매니저 (File Manager)**:
  - [ ] 로컬 파일 ↔ 원격 서버 파일 2분할 브라우저 TUI 모달
  - [ ] 다중 파일(Multi-file) 선택 지원
  - [ ] 업로드(`[u]`) / 다운로드(`[d]`) 프로그레스 바 및 전송 속도 표시

### 📜 Phase 2: 로그 관리 & 감사 리포트 (Audit & Logging)
- [ ] **콘솔 세션 로그 저장 및 내보내기 (Export)**:
  - [ ] 콘솔 출력 및 명령어 기록을 `leitstand_<host>_<timestamp>.log` 파일로 로컬 저장 (`Ctrl+E`)
  - [ ] 세션 로그 보관 기간 및 정리 정책 설정
- [ ] **내부 SQLite DB 관리 (환경설정 안)**:
  - [ ] DB 백업(JSON 내보내기) 및 복원 기능
  - [ ] DB 최적화 (SQLite Vacuum)
  - [ ] 마스터 금고 및 데이터 완전 초기화(Reset)

### 🚇 Phase 3: 고급 인프라 네트워크 기능
- [ ] **SSH 포트 포워딩 / 터널링 매니저 (SSH Tunneling)**:
  - [ ] 원격 서버 내부 DB(MySQL 3306, Redis 6379)를 로컬 포트(`localhost:xxxx`)로 포워딩
  - [ ] 활성 터널 상태 모니터링 및 원클릭 온/오프

### 📦 Phase 4: 글로벌 배포 & 쇼케이스 (Global Release)
- [ ] **크로스 플랫폼 빌드 파이프라인 (GitHub Actions & GoReleaser)**:
  - [ ] Windows (.exe), macOS (Apple Silicon / Intel), Linux (tar.gz) 자동 빌드 및 릴리즈
- [ ] **글로벌 쇼케이스 README.md & 데모 GIF 제작**:
  - [ ] 영문/독일어 공식 소개 문서, GIF 터미널 애니메이션, 뱃지 추가
  - [ ] LinkedIn, Threads, Reddit, HackerNews 릴리즈 포스팅 준비
