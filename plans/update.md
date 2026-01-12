# RectangleWin 의존성 업데이트 상세 스펙

## 개요

이 문서는 RectangleWin 프로젝트의 Go 및 라이브러리 버전 업데이트에 대한 상세 스펙입니다.

---

## 1. 업데이트 범위

### 1.1 버전 변경 요약

| 구분 | 현재 버전 | 목표 버전 |
|------|----------|----------|
| Go (go.mod) | 1.17 | 1.25 |
| Go (CI) | 1.20 | 1.25 |
| gonutz/w32/v2 | v2.2.2 | v2.11.1 (latest) |
| golang.org/x/sys | v0.0.0-20211216... | latest |
| getlantern/systray | v1.1.0 | fyne.io/systray (마이그레이션) |
| actions/checkout | @master | v4 |
| actions/setup-go | v4 | v6 |
| goreleaser-action | v2 | v6 |
| goreleaser | v2.12.0 | ~> v2 (latest v2.x) |

### 1.2 추가 변경사항

- README.md: Go 버전 요구사항 1.17+ → 1.25+ 업데이트
- .goreleaser.yaml: 필요시 스키마 업데이트

---

## 2. 기술적 결정사항

### 2.1 업그레이드 전략
- **직행 업그레이드**: Go 1.17 → 1.25로 한 번에 업그레이드 (중간 버전 거치지 않음)
- **이유**: 빠른 진행, 문제 발생 시 직접 수정

### 2.2 systray 라이브러리 마이그레이션
- **결정**: `github.com/getlantern/systray` → `fyne.io/systray`로 마이그레이션
- **이유**: getlantern/systray는 2021년 이후 업데이트 없음, Fyne 팀이 활발히 관리 중
- **사전 작업**: 마이그레이션 전 fyne.io/systray API 문서 검토 필요
- **주의사항**: import 경로 변경 필요

### 2.3 빌드 환경
- **CGO**: CGO_ENABLED=0으로 빌드 (변경 없음)
- **대상 OS**: Windows 10 이상만 지원
- **빌드 환경**: WSL Ubuntu

### 2.4 롤백 전략
- **방법**: git revert로 되돌리기
- **전제조건**: 각 PR이 독립적으로 revert 가능하도록 분리

---

## 3. PR 구조 및 순서

### PR #1: CI/CD 업데이트 (먼저 진행)

**목적**: 새로운 Go 버전으로 빌드할 수 있도록 CI 환경 준비

**변경 파일**:
- `.github/workflows/ci.yml`

**변경 내용**:
```yaml
# Before
- uses: actions/checkout@master
- uses: actions/setup-go@v4
  with:
    go-version: "1.20"
- uses: goreleaser/goreleaser-action@v2
  with:
    version: v2.12.0

# After
- uses: actions/checkout@v4
- uses: actions/setup-go@v6
  with:
    go-version: "1.25"
- uses: goreleaser/goreleaser-action@v6
  with:
    version: "~> v2"
```

**테스트**: CI 파이프라인 통과 확인

---

### PR #2: Go 및 라이브러리 업데이트

**목적**: Go 버전 및 핵심 라이브러리 업데이트

**변경 파일**:
- `go.mod`
- `go.sum`
- `README.md`
- `.goreleaser.yaml` (필요시)

**작업 순서**:
1. w32 라이브러리 CHANGELOG 확인 (Breaking Changes 검토)
2. `go mod edit -go=1.25`
3. `go get github.com/gonutz/w32/v2@latest`
4. `go get golang.org/x/sys@latest`
5. `go mod tidy`
6. 빌드 테스트: `GOOS=windows go build -ldflags -H=windowsgui .`
7. 컴파일 에러 발생 시 직접 수정
8. README.md Go 버전 요구사항 업데이트

**테스트**:
- 로컬 빌드 성공
- Windows에서 exe 실행 확인

---

### PR #3: systray 마이그레이션

**목적**: 유지보수되는 systray 포크로 마이그레이션

**사전 작업**:
- fyne.io/systray API 문서 검토
- API 호환성 확인

**변경 파일**:
- `go.mod`
- `go.sum`
- systray import가 있는 모든 .go 파일

**작업 순서**:
1. import 경로 변경: `github.com/getlantern/systray` → `fyne.io/systray`
2. `go mod tidy`
3. 빌드 테스트
4. API 차이로 인한 코드 수정 (필요시)

**테스트**:
- 로컬 빌드 성공
- Windows에서 시스템 트레이 아이콘 동작 확인
- 메뉴 기능 동작 확인

---

## 4. 업데이트 전 동작 확인 절차

### 4.1 로컬 빌드 테스트

```bash
cd /home/nicewook/dev/myproject/RectangleWin

# go.mod 정리
go mod tidy

# 리소스 생성 (아이콘, 버전 정보)
go generate

# Windows용 빌드 (WSL/Linux에서)
GOOS=windows go build -ldflags -H=windowsgui .

# 빌드 결과 확인
ls -la RectangleWin.exe
```

### 4.2 코드 품질 검사

```bash
# gofmt 검사
gofmt -s -d .

# go mod 정리 상태 확인
go mod tidy && git diff --exit-code go.mod go.sum
```

### 4.3 Windows에서 실행 테스트

#### Step 1: 기존 RectangleWin 종료
- **방법 1:** 시스템 트레이 아이콘 우클릭 → "Quit" 선택
- **방법 2:** 작업 관리자(`Ctrl+Shift+Esc`) → RectangleWin 프로세스 종료

#### Step 2: WSL에서 빌드한 exe 파일 실행

**방법 1: Windows 탐색기에서 직접 접근**
```
\\wsl$\Ubuntu\home\nicewook\dev\myproject\RectangleWin\RectangleWin.exe
```

**방법 2: Windows 경로로 복사 후 실행**
```bash
cp RectangleWin.exe /mnt/c/Users/<사용자명>/Desktop/
```

#### Step 3: 기능 확인
- 시스템 트레이 아이콘이 표시되는지
- 윈도우 스냅 기능이 동작하는지 (Win + 방향키)
- "Always on Top" 기능이 동작하는지

---

## 5. 테스트 체크리스트

### 5.1 각 PR별 테스트

**PR #1 (CI/CD)**:
- [ ] GitHub Actions 워크플로우 통과
- [ ] gofmt 검사 통과
- [ ] go mod tidy 검사 통과

**PR #2 (Go/라이브러리)**:
- [ ] `go generate` 성공
- [ ] `go build` 성공
- [ ] `gofmt` 통과
- [ ] Windows에서 exe 실행

**PR #3 (systray)**:
- [ ] `go build` 성공
- [ ] 시스템 트레이 아이콘 표시
- [ ] 트레이 메뉴 동작
- [ ] Quit 메뉴 동작

### 5.2 최종 통합 테스트

- [ ] 윈도우 스냅 기능 (Win + 방향키)
- [ ] "Always on Top" 기능
- [ ] 시스템 트레이 아이콘 표시
- [ ] 트레이 메뉴에서 Quit 동작

---

## 6. 위험 요소 및 대응

| 위험 요소 | 가능성 | 대응 방안 |
|----------|--------|----------|
| w32 API Breaking Changes | 중간 | CHANGELOG 사전 검토, 컴파일 에러 시 직접 수정 |
| systray API 차이 | 낮음 | 사전 문서 검토, 테스트로 확인 |
| Go 언어 자체 변경 | 낮음 | Go는 하위 호환성 유지가 원칙 |
| goreleaser 스키마 변경 | 낮음 | CI 실패 시 .goreleaser.yaml 수정 |

---

## 7. 수정 대상 파일 목록

```
.github/workflows/ci.yml    # PR #1
go.mod                      # PR #2, #3
go.sum                      # PR #2, #3
README.md                   # PR #2
.goreleaser.yaml            # PR #2 (필요시)
*.go (systray import)       # PR #3
```

---

## 8. 참고 자료

- [Go Release History](https://go.dev/doc/devel/release)
- [Go 1.25 Release Notes](https://go.dev/doc/go1.25)
- [gonutz/w32 GitHub](https://github.com/gonutz/w32)
- [golang.org/x/sys](https://pkg.go.dev/golang.org/x/sys)
- [fyne.io/systray](https://pkg.go.dev/fyne.io/systray)
- [goreleaser-action](https://github.com/goreleaser/goreleaser-action)
- [actions/setup-go](https://github.com/actions/setup-go)
