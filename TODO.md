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
- [x] **SSH Private Key 파일 인증 & 인앱 파일 브라우저 (`internal/vault`, `internal/ssh`, `internal/tui`)**:
  - [x] 호스트 등록 모달(`[a]`)에서 `[1] 🔒 Password` ↔ `[2] 🔑 Private Key` 실시간 전환 지원
  - [x] **인앱 TUI 파일 탐색기 (`[b]` File Picker)**: `~/.ssh/` 디렉토리 자동 탐색, `.pem`/`id_rsa`/`id_ed25519`/`test_key` 등 키 파일 자동 인식 🔑 및 `Enter` 한 번으로 경로 자동 입력
  - [x] Private Key 내용과 Passphrase를 Argon2id + AES-256-GCM 마스터 금고에 안전하게 암호화 보관
  - [x] 등록 폼 상하 줄간격(Line Spacing), 입력창 너비(38), 가독성 및 안내 배지 대폭 개선
  - [x] Go 인메모리 테스트 SSH 서버 (`cmd/mockssh`) 및 자동화 유닛 테스트 완벽 검증
- [x] **호스트 정보 수정 기능 (`[e]` Edit Host Modal)**:
  - [x] 서버 목록에서 `[e]` 키를 눌러 기존 서버 이름, IP, 포트, 계정, 인증 방식(비밀번호/키), 그룹명 즉시 수정
  - [x] 수정 저장 시 금고/DB 즉시 갱신 및 백그라운드 SSH 연결 풀 자동 갱신(`🟢 Online`)
- [x] **대화형 PTY 터미널 전환 (`[t]`) & 복합 쉘 명령어 격리**:
  - [x] 단축키 `[t]` (또는 `Ctrl+T`)로 언제든지 1초 만에 전체화면 네이티브 SSH 쉘(PTY) 진입 및 `exit`으로 0.1초 복귀
  - [x] `( cd ... ) ; ( cmd )` 서브쉘 래핑으로 `&&`, `||`, 파이프(`|`), 이모지 포함 복합 명령어 100% 정상 출력
  - [x] 런북 서랍(`?`) 슬라이딩 윈도우 스크롤(▲/▼) 및 카테고리 헤더 렌더링 개선
- [x] **MobaXterm 스타일 "상태 유지 멀티 탭 쉘 엔진" (`internal/tui/tab.go`, `internal/tui/update_console.go`)**:
  - [x] 호스트별 독립된 멀티 콘솔 탭 관리자 (Multi-Tab Stateful Shell)
  - [x] 단축키: `Ctrl+N` (새 탭), `Ctrl+W` (탭 닫기), `Alt+1`~`Alt+9` (0.01초 초고속 탭 전환)
  - [x] 각 탭별 고유 CWD(작업 디렉토리), 명령어 히스토리(`↑/↓`), 뷰포트 스크롤 위치 격리 보존
  - [x] 비동기 백그라운드 실시간 스트리밍 (`tail -f`, `docker logs -f`, `journalctl -f`, `ping`) & `🔴 LIVE` 점멸 배지
  - [x] 스트리밍 중인 탭에서 `Ctrl+C` 입력 시 해당 탭 스트리밍 작업만 안전하게 중단(Cancel)
  - [x] 3개 국어(EN/KO/DE) 탭 단축키 힌트 및 상태 안내 완벽 연동
- [x] **Phase 2: SFTP 2분할 양방향 파일 매니저 & 클립보드 이동 엔진 (`internal/ssh/sftp.go`, `internal/tui/file_manager_modal.go`)**:
  - [x] **광폭 2분할 TUI 매니저 (`[f]`, `[F6]`)**: 내 로컬 PC ↔ 원격 Linux 서버 90% 광폭 뷰포트
  - [x] **클립보드 자유 파일 이동/복사 (`[x]` 잘라내기 ➔ 자유 탐색 ➔ `[p]` 붙여넣기)**:
    - `[x]` / `[Ctrl+X]` (잘라내기 / 이동 찜하기) & `[c]` / `[Ctrl+C]` (복사 찜하기)
    - 파일들을 쥔 상태로 방향키/`[Enter]`/`[Backspace]`로 눈으로 직접 확인하며 폴더 자유 탐색
    - 도착한 폴더에서 `[p]` / `[Ctrl+V]` 누르면 0.01초 만에 현재 폴더로 일괄 이동(`mv`) 또는 복사(`cp`) 완료
    - 1회 투하 후 클립보드 자동 정리 & 선명한 `✨ 완료` 녹색 배너 출력
  - [x] **다중 파일 선택 & 배치 일괄 전송**:
    - `[Space]` 다중 선택(`[*]`) 및 `[a]` 전체 선택/해제
    - `[u]` (로컬 ➔ 원격 일괄 업로드), `[d]` (원격 ➔ 로컬 일괄 다운로드)
    - 실시간 진행률 프로그레스 바, 전송 속도(MB/s), 남은 용량 표시
  - [x] **파일 매니저 내 즉석 쉘 실행 (`[:]`, `[!]`)**:
    - `ls -la`, `cd ..`, `df -h`, `cat`, `chmod` 등 현재 활성 폴더에서 즉시 명령어 실행 & 결과 팝업
    - `cd` 실행 시 탐색기 경로 자동 동기화
  - [x] **안전한 조작 & 방어 로직**:
    - 접근 권한 부족(`Permission Denied`) 폴더 진입 시 자동 롤백 및 경고 배너
    - `[Delete]` / `[Shift+X]` 누를 시 대상 항목의 실제 이름('app.py' 파일 / 'logs' 폴더)을 명시하는 안전 삭제 확인창
    - `[Esc]` / `[q]` / `[f]` 누를 시 실수 방지를 위한 안전 종료 확인 모달
    - 파일 매니저 전용 단축키 가이드 런북 (`[?]`, `[F1]`) 3개 국어(KO/EN/DE) 지원
  - [x] **인라인 파일 시스템 유틸리티**: `[n]` 새 폴더(`mkdir`), `[N]` 새 빈 파일(`touch`), `[r]` 이름 변경(`mv`), `[.]` 숨김 파일 토글

---

## 🎯 2. Upcoming Strategic Roadmap (다음 진행할 핵심 작업)

### 📜 Phase 3: 세션 감사 로그 관리 & SQLite 최적화 (Audit & Logging)
- [ ] **콘솔 세션 로그 로컬 내보내기 (`Ctrl+E`)**:
  - [ ] 콘솔 출력 및 명령어 기록을 `leitstand_<host>_<timestamp>.log` 파일로 로컬 저장
- [ ] **내부 SQLite DB 관리 (환경설정 안)**:
  - [ ] DB 백업(JSON 내보내기) 및 복원 기능
  - [ ] DB 최적화 (SQLite Vacuum) 및 마스터 금고 완전 초기화(Reset)

### 🚇 Phase 4: 고급 인프라 네트워크 & 글로벌 배포
- [ ] **SSH 포트 포워딩 / 터널링 매니저 (SSH Tunneling)**:
  - [ ] 원격 서버 내부 DB(MySQL 3306, Redis 6379)를 로컬 포트로 포워딩
- [ ] **크로스 플랫폼 빌드 파이프라인 (GitHub Actions & GoReleaser)**:
  - [ ] Windows (.exe), macOS (Apple Silicon / Intel), Linux (tar.gz) 자동 릴리즈
- [ ] **글로벌 쇼케이스 README.md & 데모 GIF 제작**
