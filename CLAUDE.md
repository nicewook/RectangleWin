# RectangleWin

Windows용 핫키 기반 윈도우 스냅 및 리사이징 유틸리티. macOS Rectangle.app/Spectacle.app의 Windows 재구현.

## 프로젝트 구조

```
RectangleWin/
├── main.go          # 진입점, 핫키 등록, 스냅/최대화/복원 로직
├── snap.go          # 스냅 레이아웃 함수 (halves, corners)
├── multimon.go      # 좌/우 반복 시 인접 모니터로 이동
├── windowstate.go   # 복원용 창 상태(WINDOWPLACEMENT) 저장
├── logging.go       # GUI 실행 시 파일 로그 (%LOCALAPPDATA%\RectangleWin)
├── hotkey.go        # 핫키 등록 및 메시지 루프
├── tray.go          # 시스템 트레이 아이콘 관리
├── autorun.go       # Windows 시작프로그램 등록 (레지스트리)
├── monitor.go       # 멀티 모니터 정보 조회
├── systemwindow.go  # 시스템 윈도우 필터링
├── keymap.go        # 가상 키코드 → 문자열 매핑
├── w32ex/           # 확장 Windows API 바인딩
│   └── functions.go # GetDpiForWindow, SetProcessDPIAware 등
└── assets/          # 아이콘 파일들 (ico, png)
```

## 기술 스택

- **언어**: Go 1.27+
- **타겟 OS**: Windows only
- **주요 의존성**:
  - `fyne.io/systray` - 시스템 트레이
  - `github.com/gonutz/w32/v2` - Windows API 바인딩
  - `golang.org/x/sys/windows` - Windows 시스템 콜

## 빌드

```bash
# 아이콘 리소스 생성 (Linux/macOS에서 실행)
go generate

# Windows 바이너리 빌드 (GUI 모드, 콘솔 창 없음)
GOOS=windows go build -ldflags -H=windowsgui .
```

## 주요 기능

모든 단축키는 Ctrl+Alt 조합이며 MOD_NOREPEAT로 등록한다.

- **절반 스냅**: Ctrl+Alt+←/→/↑/↓ (←/→는 이미 그 위치면 인접 모니터로 이동)
- **코너 스냅**: Ctrl+Alt+U/I/J/K
- **중앙 배치(75%)**: Ctrl+Alt+C
- **최대화 토글**: Ctrl+Alt+Enter
- **복원**: Ctrl+Alt+Backspace (첫 스냅 전 상태로)
- thirds·크기 조절·Always On Top은 미구현

## 코드 컨벤션

- Windows API 호출은 주로 `github.com/gonutz/w32/v2` 사용
- 추가 API가 필요하면 `w32ex/` 패키지에 추가
- 메인 스레드 고정 필수 (`runtime.LockOSThread()`)
- DPI 인식 처리 필수 (`SetProcessDPIAware` → Per-Monitor v2). PMv2에서는 GetWindowRect·DWM 프레임이 모두 물리 픽셀이므로 **좌표에 DPI 배율을 곱하지 말 것**
- `syscall.NewCallback`은 호출마다 만들지 말 것 (약 2000개 한도에서 프로세스 종료)
- 테스트: `go vet ./... && go test ./...` (Windows에서). 좌표 계산은 순수 함수로 두고 `logic_test.go`·`behavior_test.go`에 표 테스트 추가. 실제 창을 쓰는 통합 테스트는 `win32_integration_test.go` (`RECTANGLEWIN_SKIP_WIN32=1`로 생략)

## 개발 시 주의사항

- WSL에서는 Windows API 관련 코드 빌드/테스트 불가 (GOOS=windows 크로스 컴파일만 가능)
- 핫키 충돌 시 다른 프로그램이 해당 키 조합 사용 중일 수 있음
- 시스템 트레이 초기화는 별도 goroutine에서 수행
- **커밋 전 gofmt 필수**: CI에서 gofmt 검사를 수행하므로, 커밋 전에 반드시 `gofmt -s -w .` 실행

## 라이선스 (Apache 2.0)

이 프로젝트는 [ahmetb/RectangleWin](https://github.com/ahmetb/RectangleWin)의 포크입니다.

**코드 수정 시 필수 사항:**
- 파일 상단의 기존 저작권 문구(`Copyright 2022 Ahmet Alp Balkan`) 유지
- 수정한 파일에 수정 사실 명시 (주석으로 날짜, 수정자, 변경 내용 기입)
  ```go
  // Modified by [이름] on [날짜]
  // Changes: [변경 내용 요약]
  ```
- `LICENSE` 파일 삭제 금지
