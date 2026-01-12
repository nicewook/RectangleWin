# GitHub Actions CI 워크플로우 설명

이 문서는 `.github/workflows/ci.yml` 파일의 라인별 상세 설명입니다.

---

## 1. 라이선스 헤더 (Lines 1-13)

```yaml
# Copyright 2022 Ahmet Alp Balkan
#
# Licensed under the Apache License, Version 2.0 (the "License");
# ...
```

- Apache License 2.0 라이선스 고지
- 원작자: Ahmet Alp Balkan
- 모든 소스 파일에 동일한 라이선스 헤더 적용

---

## 2. 워크플로우 기본 설정 (Lines 15-18)

```yaml
name: RectangleWin
on:
  push:
  pull_request:
```

| 라인 | 설명 |
|------|------|
| `name: RectangleWin` | GitHub Actions UI에 표시되는 워크플로우 이름 |
| `on:` | 워크플로우 트리거 조건 정의 |
| `push:` | 모든 브랜치에 push 시 실행 (필터 없음) |
| `pull_request:` | PR 생성/업데이트 시 실행 |

**트리거 시점:**
- 코드가 push될 때 (모든 브랜치)
- PR이 열리거나 업데이트될 때

---

## 3. 작업 정의 (Lines 19-21)

```yaml
jobs:
  ci:
    runs-on: ubuntu-latest
```

| 라인 | 설명 |
|------|------|
| `jobs:` | 실행할 작업들 정의 시작 |
| `ci:` | 작업 ID (이름은 자유롭게 지정 가능) |
| `runs-on: ubuntu-latest` | Ubuntu 최신 LTS 버전에서 실행 |

**실행 환경:**
- GitHub가 제공하는 Ubuntu 가상 머신
- 매 실행마다 깨끗한 환경에서 시작

---

## 4. Checkout 단계 (Lines 23-24)

```yaml
    - name: Checkout
      uses: actions/checkout@v4
```

| 항목 | 설명 |
|------|------|
| **목적** | 리포지토리 코드를 runner에 복제 |
| **Action** | `actions/checkout@v4` (GitHub 공식) |
| **동작** | `git clone` + `git checkout` 수행 |

**왜 필요한가?**
- GitHub Actions runner는 빈 환경에서 시작
- 코드가 없으면 빌드/테스트 불가능

---

## 5. Go 설치 (Lines 25-28)

```yaml
    - name: Setup Go
      uses: actions/setup-go@v6
      with:
        go-version: "1.25"
```

| 항목 | 설명 |
|------|------|
| **목적** | Go 언어 환경 설치 |
| **Action** | `actions/setup-go@v6` (GitHub 공식) |
| **버전** | Go 1.25 설치 |

**기능:**
- 지정된 Go 버전 다운로드 및 설치
- `PATH` 환경변수에 Go 추가
- `GOPATH`, `GOMODCACHE` 등 환경 설정

---

## 6. 캐시 경로 추출 (Lines 29-32)

```yaml
    - id: go-cache-paths
      run: |
        echo "go-build=$(go env GOCACHE)" >> $GITHUB_OUTPUT
        echo "go-mod=$(go env GOMODCACHE)" >> $GITHUB_OUTPUT
```

| 항목 | 설명 |
|------|------|
| **목적** | Go 캐시 디렉토리 경로를 변수로 저장 |
| `id: go-cache-paths` | 이 단계를 참조하기 위한 ID |
| `go env GOCACHE` | Go 빌드 캐시 경로 (컴파일 결과물) |
| `go env GOMODCACHE` | Go 모듈 캐시 경로 (다운로드된 의존성) |
| `$GITHUB_OUTPUT` | 다른 단계에서 사용할 출력 값 저장 |

**출력 예시:**
```
go-build=/home/runner/.cache/go-build
go-mod=/home/runner/go/pkg/mod
```

---

## 7. 빌드 캐시 설정 (Lines 33-39)

```yaml
    - name: go build cache
      uses: actions/cache@v4
      with:
        key: ${{ runner.os }}-go-build-${{ hashFiles('**/go.sum') }}
        path: |
          ${{ steps.go-cache-paths.outputs.go-build }}
          ${{ steps.go-cache-paths.outputs.go-mod }}
```

| 항목 | 설명 |
|------|------|
| **목적** | 빌드 속도 향상을 위한 캐싱 |
| **Action** | `actions/cache@v4` |
| `key` | 캐시 식별자 (OS + go.sum 해시) |
| `path` | 캐시할 디렉토리들 |

**캐시 키 구성:**
- `runner.os`: 운영체제 (Linux)
- `hashFiles('**/go.sum')`: go.sum 파일의 해시값

**동작 원리:**
1. 캐시 키로 기존 캐시 검색
2. 있으면 → 캐시 복원 (빠름)
3. 없으면 → 새로 빌드 후 캐시 저장

**효과:**
- 의존성 다운로드 시간 절약
- 컴파일 시간 단축

---

## 8. 코드 포맷 검사 (Lines 40-41)

```yaml
    - name: Ensure gofmt
      run: test -z "$(gofmt -s -d .)"
```

| 항목 | 설명 |
|------|------|
| **목적** | Go 코드 포맷팅 규칙 준수 확인 |
| `gofmt -s -d .` | 현재 디렉토리의 모든 Go 파일 검사 |
| `-s` | 코드 단순화 (simplify) |
| `-d` | diff 형식으로 출력 |
| `test -z "..."` | 출력이 비어있으면 성공 (exit 0) |

**실패 조건:**
- 포맷팅되지 않은 코드가 있으면 diff 출력 → 테스트 실패

**해결 방법:**
```bash
gofmt -s -w .  # 자동 수정
```

---

## 9. go.mod 정리 상태 확인 (Lines 42-43)

```yaml
    - name: go.mod is tidied
      run: go mod tidy && git diff --no-patch --exit-code
```

| 항목 | 설명 |
|------|------|
| **목적** | go.mod/go.sum이 정리된 상태인지 확인 |
| `go mod tidy` | 불필요한 의존성 제거, 누락된 의존성 추가 |
| `git diff --no-patch --exit-code` | 파일 변경 있으면 실패 |

**검사 로직:**
1. `go mod tidy` 실행
2. 파일이 변경되었는지 확인
3. 변경됨 → 커밋 전에 `go mod tidy`를 안 했다는 의미 → 실패

**해결 방법:**
```bash
go mod tidy
git add go.mod go.sum
git commit --amend
```

---

## 10. 리소스 생성 (Lines 44-45)

```yaml
    - name: go generate (Binary Version Information and Icon)
      run: go generate
```

| 항목 | 설명 |
|------|------|
| **목적** | 빌드에 필요한 리소스 파일 생성 |
| `go generate` | `//go:generate` 주석이 있는 코드 실행 |

**이 프로젝트에서 생성하는 것:**
- Windows 실행 파일 버전 정보 (versioninfo)
- 아이콘 리소스 (.syso 파일)

**관련 코드 (main.go):**
```go
//go:generate goversioninfo -icon=assets/icon.ico
```

---

## 11. 스냅샷 빌드 (Lines 46-54)

```yaml
    - name: Build-only (GoReleaser)
      if: "!startsWith(github.ref, 'refs/tags/')"
      uses: goreleaser/goreleaser-action@v6
      with:
        distribution: goreleaser
        version: "~> v2"
        args: release --snapshot
      env:
        GORELEASER_SKIP_PUBLISH: true
```

| 항목 | 설명 |
|------|------|
| **목적** | 일반 push/PR에서 빌드 테스트 |
| `if: "!startsWith(...)"` | 태그가 **아닐 때**만 실행 |
| `--snapshot` | 버전 없이 테스트 빌드 |
| `GORELEASER_SKIP_PUBLISH: true` | GitHub Release에 게시하지 않음 |

**조건 분석:**
- `github.ref`: 현재 브랜치/태그 참조
- `refs/tags/v1.0.0` 형태면 태그
- `refs/heads/main` 형태면 브랜치

**스냅샷 빌드:**
- 실제 릴리스 없이 빌드 프로세스만 검증
- 빌드 실패를 미리 발견

---

## 12. 릴리스 빌드 (Lines 55-64)

```yaml
    - name: Publish release (GoReleaser)
      if: startsWith(github.ref, 'refs/tags/')
      uses: goreleaser/goreleaser-action@v6
      env:
        GITHUB_TOKEN: ${{ secrets.GITHUB_TOKEN }}
        GORELEASER_SKIP_PUBLISH: true
      with:
        distribution: goreleaser
        version: "~> v2"
        args: release
```

| 항목 | 설명 |
|------|------|
| **목적** | 태그 push 시 실제 릴리스 빌드 |
| `if: startsWith(...)` | 태그일 때**만** 실행 |
| `GITHUB_TOKEN` | GitHub API 인증 (자동 제공) |
| `args: release` | 실제 릴리스 수행 |

**실행 조건:**
```bash
git tag v1.0.0
git push origin v1.0.0  # 이때 실행됨
```

**참고:** `GORELEASER_SKIP_PUBLISH: true`가 설정되어 있어 실제로는 GitHub Releases에 게시되지 않음 (테스트 목적인 듯)

---

## 워크플로우 흐름도

```
push/PR 발생
    │
    ▼
┌─────────────────────────────────────┐
│  1. Checkout (코드 복제)             │
│  2. Setup Go (Go 1.25 설치)          │
│  3. 캐시 경로 추출                    │
│  4. 캐시 복원/저장                    │
│  5. gofmt 검사                       │
│  6. go mod tidy 검사                 │
│  7. go generate (리소스 생성)         │
└─────────────────────────────────────┘
    │
    ▼
┌─────────────────┐     ┌─────────────────┐
│  태그 아님?      │ YES │  스냅샷 빌드     │
│  (일반 push/PR) │────▶│  (테스트 목적)   │
└─────────────────┘     └─────────────────┘
    │ NO (태그)
    ▼
┌─────────────────┐
│  릴리스 빌드     │
│  (실제 배포)     │
└─────────────────┘
```

---

## 관련 파일

- `.github/workflows/ci.yml` - 이 문서에서 설명하는 파일
- `.goreleaser.yaml` - GoReleaser 빌드 설정
- `go.mod` / `go.sum` - Go 모듈 의존성
- `main.go` - `//go:generate` 지시문 포함
