# RectangleWin 기능 개선 최종 스펙

## 1. 프로젝트 개요

### 1.1 기본 정보
- **프로젝트 이름**: RectangleWin
- **프로젝트 상태**: 새 프로젝트 (기존 사용자 호환성 고려 불필요)
- **목표**: macOS Rectangle 앱의 Windows 클론, 17개 단축키 완전 구현

### 1.2 주요 변경사항 요약
| 항목 | 변경 전 | 변경 후 |
|------|---------|---------|
| Center | 크기 유지 중앙 배치 | 75% 크기로 리사이즈 후 중앙 배치 |
| Size 조정 | 50px 고정 | 화면 너비의 3% (해상도 비례) |
| Restore | SW_RESTORE만 | 스냅 전 상태 복원 포함 (통합) |
| Multi-Display | 미지원 | Halves/Thirds 방향별 모니터 이동 |
| UI | 콘솔만 | 시스템 트레이 아이콘 추가 |
| 미사용 코드 | 유지 | 삭제 |

---

## 2. 단축키 스펙 (17개)

### 2.1 Halves (4개)
| 단축키 | 기능 | Multi-Display |
|--------|------|---------------|
| CTRL+ALT+LEFT | Left Half | 왼쪽 방향 이동, 순환 |
| CTRL+ALT+RIGHT | Right Half | 오른쪽 방향 이동, 순환 |
| CTRL+ALT+UP | Top Half | 이동 없음 |
| CTRL+ALT+DOWN | Bottom Half | 이동 없음 |

### 2.2 Maximize / Center / Restore (3개)
| 단축키 | 기능 | 동작 | Multi-Display |
|--------|------|------|---------------|
| CTRL+ALT+ENTER | Maximize | 창 최대화 | 이동 없음 |
| CTRL+ALT+C | Center | 75% 크기로 중앙 배치 | 이동 없음 |
| CTRL+ALT+BACKSPACE | Restore | 통합 복원 (아래 상세) | 이동 없음 |

### 2.3 Corners (4개)
| 단축키 | 기능 | Multi-Display |
|--------|------|---------------|
| CTRL+ALT+U | Top Left (1/4) | 이동 없음 |
| CTRL+ALT+I | Top Right (1/4) | 이동 없음 |
| CTRL+ALT+J | Bottom Left (1/4) | 이동 없음 |
| CTRL+ALT+K | Bottom Right (1/4) | 이동 없음 |

### 2.4 Thirds (5개)
| 단축키 | 기능 | Multi-Display |
|--------|------|---------------|
| CTRL+ALT+D | First Third (Left 1/3) | 왼쪽 방향 이동, 순환 |
| CTRL+ALT+F | Center Third | 이동 없음 |
| CTRL+ALT+G | Last Third (Right 1/3) | 오른쪽 방향 이동, 순환 |
| CTRL+ALT+E | First Two Thirds (Left 2/3) | 왼쪽 방향 이동, 순환 |
| CTRL+ALT+T | Last Two Thirds (Right 2/3) | 오른쪽 방향 이동, 순환 |

### 2.5 Size (2개)
| 단축키 | 기능 | Multi-Display |
|--------|------|---------------|
| CTRL+ALT+- | Make Smaller | 이동 없음 |
| CTRL+ALT++ | Make Larger | 이동 없음 |

---

## 3. 기능 상세 스펙

### 3.1 Center 기능 (CTRL+ALT+C)

**동작**: 창을 화면의 75% (3/4) 크기로 리사이즈하고 화면 중앙에 배치

```
계산식:
- 너비 = 화면 너비 × 0.75
- 높이 = 화면 높이 × 0.75
- X 위치 = (화면 너비 - 창 너비) / 2
- Y 위치 = (화면 높이 - 창 높이) / 2
```

### 3.2 Make Smaller/Larger 기능 (CTRL+ALT+-/+)

**동작**: 창을 현재 위치 중심으로 확대/축소

**크기 조정량**:
- **비율**: 화면 너비의 3%
- **예시**: 1920px 화면 → 약 58px, 3840px(4K) 화면 → 약 115px

**제한**:
- **최소 크기**: 100px × 100px
- **최대 크기**: 제한 없음

```
계산식 (축소 기준):
resizeStep = 화면 너비 × 0.03

newWidth = 현재 너비 - (resizeStep × 2)
newHeight = 현재 높이 - (resizeStep × 2)

if newWidth < 100: newWidth = 100
if newHeight < 100: newHeight = 100

centerX = 현재 Left + 현재 너비 / 2
centerY = 현재 Top + 현재 높이 / 2

결과:
Left = centerX - newWidth / 2
Top = centerY - newHeight / 2
Right = centerX + newWidth / 2
Bottom = centerY + newHeight / 2
```

### 3.3 Restore 기능 (CTRL+ALT+BACKSPACE)

**동작**: 통합 복원
1. 창이 Maximize 상태 → `SW_RESTORE` 호출 (Windows 기본 복원)
2. 창이 스냅 상태 → 저장된 원래 위치/크기로 복원

**스냅 전 상태 저장 규칙**:
| 항목 | 값 |
|------|-----|
| 저장 위치 | 메모리 (프로그램 종료 시 소멸) |
| 저장 개수 | 창당 1개 |
| 저장 시점 | 첫 스냅 시에만 (저장된 상태가 없을 때만) |
| Restore 후 | 저장된 상태 삭제 (다시 스냅해야 저장됨) |

**상태 저장 데이터 구조**:
```go
type WindowState struct {
    Left   int32
    Top    int32
    Right  int32
    Bottom int32
}

var savedStates = make(map[w32.HWND]WindowState)
```

**Restore 로직**:
```
1. hwnd = GetForegroundWindow()
2. if IsZoomed(hwnd):  // 최대화 상태인지 확인
       ShowWindow(hwnd, SW_RESTORE)
       return
3. if savedStates[hwnd] exists:
       SetWindowPos(hwnd, savedStates[hwnd])
       delete(savedStates, hwnd)
```

---

## 4. Multi-Display 스펙

### 4.1 지원 범위

**이동 지원 기능** (6개):
- Left Half, Right Half
- First Third, Last Third
- First Two Thirds, Last Two Thirds

**이동 미지원 기능** (11개):
- Top Half, Bottom Half (수평 배치에서 위/아래 개념 모호)
- Center Third (방향성 없음)
- Corners 전체 (U/I/J/K)
- Center, Maximize, Restore
- Make Smaller/Larger

### 4.2 모니터 배치 가정

- **지원**: 좌/우 수평 배치만 고려
- **미지원**: 세로(위/아래) 배치

### 4.3 이동 방향 규칙

| 기능 | 이동 방향 |
|------|-----------|
| Left Half | 왼쪽으로 이동 |
| Right Half | 오른쪽으로 이동 |
| First Third | 왼쪽으로 이동 |
| Last Third | 오른쪽으로 이동 |
| First Two Thirds | 왼쪽으로 이동 |
| Last Two Thirds | 오른쪽으로 이동 |

### 4.4 이동 로직 상세

**Left Half 연속 누름 예시** (2개 모니터, 왼쪽=모니터1, 오른쪽=모니터2):

```
시작: 모니터2 Right Half
  ↓ CTRL+ALT+LEFT
모니터2 Left Half
  ↓ CTRL+ALT+LEFT
모니터1 Right Half (에지 맞춤)
  ↓ CTRL+ALT+LEFT
모니터1 Left Half
  ↓ CTRL+ALT+LEFT
모니터2 Right Half (순환, 에지 맞춤)
  ↓ (반복)
```

**Right Half 연속 누름** (반대 방향):
```
시작: 모니터1 Left Half
  ↓ CTRL+ALT+RIGHT
모니터1 Right Half
  ↓ CTRL+ALT+RIGHT
모니터2 Left Half (에지 맞춤)
  ↓ CTRL+ALT+RIGHT
모니터2 Right Half
  ↓ CTRL+ALT+RIGHT
모니터1 Left Half (순환, 에지 맞춤)
```

### 4.5 에지 맞춤 규칙

모니터 이동 시 스냅 위치 변환:

| 현재 위치 | 이동 방향 | 결과 위치 |
|-----------|-----------|-----------|
| Left Half | 왼쪽 모니터로 | Right Half |
| Right Half | 오른쪽 모니터로 | Left Half |
| First Third | 왼쪽 모니터로 | Last Third |
| Last Third | 오른쪽 모니터로 | First Third |
| First Two Thirds | 왼쪽 모니터로 | Last Two Thirds |
| Last Two Thirds | 오른쪽 모니터로 | First Two Thirds |

### 4.6 겹침 창 처리

창이 두 모니터에 걸쳐 있을 때:
- **기준**: 창 면적이 더 많이 걸친 모니터

```
계산식:
monitor1_overlap = 모니터1과 창의 교차 영역 면적
monitor2_overlap = 모니터2와 창의 교차 영역 면적

기준 모니터 = max(monitor1_overlap, monitor2_overlap) 쪽
```

### 4.7 화면 경계 처리

모니터 이동 시 창이 대상 모니터보다 클 경우:
- **처리**: 대상 모니터 크기에 맞춤 (클리핑)

```
if 창 너비 > 대상 모니터 너비:
    창 너비 = 대상 모니터 너비
if 창 높이 > 대상 모니터 높이:
    창 높이 = 대상 모니터 높이
```

---

## 5. 시스템 트레이 스펙

### 5.1 기본 동작

| 항목 | 동작 |
|------|------|
| 트레이 아이콘 | 시스템 트레이에 표시 |
| 좌클릭 | 아무 동작 없음 |
| 더블클릭 | 아무 동작 없음 |
| 우클릭 | 컨텍스트 메뉴 표시 |
| 툴팁 | "RectangleWin" |

### 5.2 컨텍스트 메뉴 구조

```
┌─────────────────────────────┐
│ About RectangleWin...       │
├─────────────────────────────┤
│ 단축키 목록...              │ → 단축키 목록 창 표시
├─────────────────────────────┤
│ ✓ Windows 시작 시 실행     │ → 토글 (체크 표시)
├─────────────────────────────┤
│ Exit                        │
└─────────────────────────────┘
```

### 5.3 시작 프로그램 등록

| 항목 | 값 |
|------|-----|
| 기본값 | ON (체크됨) |
| 등록 위치 | `HKCU\Software\Microsoft\Windows\CurrentVersion\Run` |
| 키 이름 | `RectangleWin` |
| 값 | 실행 파일 전체 경로 |

### 5.4 단축키 목록 창

단축키 목록 메뉴 클릭 시 표시되는 창:

```
┌─────────────────────────────────────────┐
│ RectangleWin 단축키 목록                │
├─────────────────────────────────────────┤
│ Halves                                  │
│   CTRL+ALT+LEFT      Left Half          │
│   CTRL+ALT+RIGHT     Right Half         │
│   CTRL+ALT+UP        Top Half           │
│   CTRL+ALT+DOWN      Bottom Half        │
│                                         │
│ Maximize / Center / Restore             │
│   CTRL+ALT+ENTER     Maximize           │
│   CTRL+ALT+C         Center             │
│   CTRL+ALT+BACKSPACE Restore            │
│                                         │
│ Corners                                 │
│   CTRL+ALT+U         Top Left           │
│   CTRL+ALT+I         Top Right          │
│   CTRL+ALT+J         Bottom Left        │
│   CTRL+ALT+K         Bottom Right       │
│                                         │
│ Thirds                                  │
│   CTRL+ALT+D         First Third        │
│   CTRL+ALT+F         Center Third       │
│   CTRL+ALT+G         Last Third         │
│   CTRL+ALT+E         First Two Thirds   │
│   CTRL+ALT+T         Last Two Thirds    │
│                                         │
│ Size                                    │
│   CTRL+ALT+-         Make Smaller       │
│   CTRL+ALT++         Make Larger        │
├─────────────────────────────────────────┤
│                              [ 닫기 ]   │
└─────────────────────────────────────────┘
```

### 5.5 콘솔 숨김

**빌드 옵션**:
```bash
go build -ldflags "-H=windowsgui" -o RectangleWin.exe .
```

---

## 6. 에러 처리 및 로깅

### 6.1 단축키 충돌 처리

다른 프로그램이 단축키를 이미 사용 중일 때:

**처리 방식**: 메시지 박스로 사용자에게 알림

```
┌─────────────────────────────────────────┐
│ RectangleWin                      [X]   │
├─────────────────────────────────────────┤
│ ⚠ 다음 단축키를 등록할 수 없습니다:     │
│                                         │
│   CTRL+ALT+C (다른 프로그램이 사용 중)  │
│                                         │
│ 나머지 단축키는 정상 동작합니다.        │
│                                         │
│                              [ 확인 ]   │
└─────────────────────────────────────────┘
```

**동작**:
- 충돌하는 단축키만 목록에 표시
- 충돌하지 않는 단축키는 정상 등록 및 동작
- 프로그램은 계속 실행됨

### 6.2 로깅

| 항목 | 값 |
|------|-----|
| 로깅 방식 | 콘솔 출력 (fmt.Println/Printf) |
| 로그 파일 | 없음 |
| 표시 조건 | 콘솔 숨김 상태에서는 보이지 않음 |

---

## 7. 코드 변경 계획

### 7.1 snap.go 변경

**삭제할 함수들**:
```go
// 사용하지 않는 Thirds 관련 함수 삭제
- topTwoThirds
- topOneThirds
- bottomTwoThirds
- bottomOneThirds
- topLeftTwoThirds
- topLeftOneThirds
- topRightTwoThirds
- topRightOneThirds
- bottomLeftTwoThirds
- bottomLeftOneThirds
- bottomRightTwoThirds
- bottomRightOneThirds
```

**수정할 함수**:
```go
// center 함수 - 75% 크기로 변경
func center(disp, _ w32.RECT) w32.RECT {
    width := disp.Width() * 3 / 4
    height := disp.Height() * 3 / 4
    return w32.RECT{
        Left:   disp.Left + (disp.Width()-width)/2,
        Top:    disp.Top + (disp.Height()-height)/2,
        Right:  disp.Left + (disp.Width()+width)/2,
        Bottom: disp.Top + (disp.Height()+height)/2,
    }
}
```

**추가할 함수**:
```go
// centerThird - 화면 중앙 1/3
func centerThird(disp, _ w32.RECT) w32.RECT {
    return w32.RECT{
        Left:   disp.Left + disp.Width()/3,
        Top:    disp.Top,
        Right:  disp.Left + disp.Width()*2/3,
        Bottom: disp.Top + disp.Height(),
    }
}

// makeSmaller - 해상도 비례 축소 (3%)
func makeSmaller(disp, cur w32.RECT) w32.RECT {
    resizeStep := disp.Width() * 3 / 100  // 3%

    newWidth := cur.Width() - resizeStep*2
    newHeight := cur.Height() - resizeStep*2

    if newWidth < 100 { newWidth = 100 }
    if newHeight < 100 { newHeight = 100 }

    centerX := cur.Left + cur.Width()/2
    centerY := cur.Top + cur.Height()/2

    return w32.RECT{
        Left:   centerX - newWidth/2,
        Top:    centerY - newHeight/2,
        Right:  centerX + newWidth/2,
        Bottom: centerY + newHeight/2,
    }
}

// makeLarger - 해상도 비례 확대 (3%)
func makeLarger(disp, cur w32.RECT) w32.RECT {
    resizeStep := disp.Width() * 3 / 100  // 3%

    newWidth := cur.Width() + resizeStep*2
    newHeight := cur.Height() + resizeStep*2

    centerX := cur.Left + cur.Width()/2
    centerY := cur.Top + cur.Height()/2

    return w32.RECT{
        Left:   centerX - newWidth/2,
        Top:    centerY - newHeight/2,
        Right:  centerX + newWidth/2,
        Bottom: centerY + newHeight/2,
    }
}
```

### 7.2 main.go 변경

**삭제할 코드**:
```go
// 사이클 인프라 전체 삭제
- var lastResized w32.HWND
- edgeFuncs 배열
- edgeFuncTurn 변수
- cornerFuncs 배열
- cornerFuncTurn 변수
- cycleFuncs 클로저
- cycleEdgeFuncs 변수
- cycleCornerFuncs 변수
- toggleAlwaysOnTop 함수
- resize 함수 내 lastResized = hwnd
```

**추가할 코드**:
```go
// 스냅 전 상태 저장용 맵
var savedStates = make(map[w32.HWND]w32.RECT)

// restore 함수 - 통합 복원
func restore() error {
    hwnd := w32.GetForegroundWindow()
    if !isZonableWindow(hwnd) {
        return errors.New("foreground window is not zonable")
    }

    // 1. 최대화 상태 확인
    if w32.IsZoomed(hwnd) {
        if !w32.ShowWindow(hwnd, w32.SW_RESTORE) {
            return fmt.Errorf("failed to ShowWindow(SW_RESTORE):%d", w32.GetLastError())
        }
        return nil
    }

    // 2. 저장된 상태가 있으면 복원
    if state, ok := savedStates[hwnd]; ok {
        if !w32.SetWindowPos(hwnd, 0, int(state.Left), int(state.Top),
            int(state.Width()), int(state.Height()),
            w32.SWP_NOZORDER|w32.SWP_NOACTIVATE) {
            return fmt.Errorf("failed to SetWindowPos:%d", w32.GetLastError())
        }
        delete(savedStates, hwnd)
        return nil
    }

    // 3. 저장된 상태가 없으면 SW_RESTORE 시도
    if !w32.ShowWindow(hwnd, w32.SW_RESTORE) {
        return fmt.Errorf("failed to ShowWindow(SW_RESTORE):%d", w32.GetLastError())
    }
    return nil
}

// resize 함수 수정 - 첫 스냅 시 상태 저장
func resize(hwnd w32.HWND, snapFunc SnapFunc) (w32.RECT, error) {
    // ... 기존 코드 ...

    // 첫 스냅 시에만 상태 저장
    if _, exists := savedStates[hwnd]; !exists {
        savedStates[hwnd] = currentRect  // 현재 위치 저장
    }

    // ... 나머지 코드 ...
}
```

### 7.3 새 파일 추가

**tray.go** - 시스템 트레이 관련:
```go
package main

// 시스템 트레이 아이콘 및 메뉴 구현
// - 아이콘 표시
// - 우클릭 메뉴
// - 시작 프로그램 등록/해제
// - 단축키 목록 창
```

**multimon.go** - Multi-Display 관련:
```go
package main

// 다중 모니터 지원
// - 모니터 목록 열거
// - 현재 창이 속한 모니터 판별
// - 다음/이전 모니터 계산
// - 에지 맞춤 변환
```

---

## 8. Hotkey ID 체계

| ID 범위 | 카테고리 |
|---------|----------|
| 1-4 | Halves (LEFT/RIGHT/UP/DOWN) |
| 10-12 | Maximize, Center, Restore |
| 20-23 | Corners (U/I/J/K) |
| 30-34 | Thirds (D/F/G/E/T) |
| 40-41 | Size (-/+) |

---

## 9. 검증 체크리스트

### 9.1 기본 기능 테스트
- [ ] CTRL+ALT+LEFT → 왼쪽 반
- [ ] CTRL+ALT+RIGHT → 오른쪽 반
- [ ] CTRL+ALT+UP → 위쪽 반
- [ ] CTRL+ALT+DOWN → 아래쪽 반
- [ ] CTRL+ALT+ENTER → 최대화
- [ ] CTRL+ALT+C → 75% 크기로 중앙 배치
- [ ] CTRL+ALT+BACKSPACE → 복원 (Maximize/스냅 상태에서)
- [ ] CTRL+ALT+U → 좌상단 1/4
- [ ] CTRL+ALT+I → 우상단 1/4
- [ ] CTRL+ALT+J → 좌하단 1/4
- [ ] CTRL+ALT+K → 우하단 1/4
- [ ] CTRL+ALT+D → 왼쪽 1/3
- [ ] CTRL+ALT+F → 중앙 1/3
- [ ] CTRL+ALT+G → 오른쪽 1/3
- [ ] CTRL+ALT+E → 왼쪽 2/3
- [ ] CTRL+ALT+T → 오른쪽 2/3
- [ ] CTRL+ALT+- → 창 축소 (해상도 비례)
- [ ] CTRL+ALT++ → 창 확대 (해상도 비례)

### 9.2 Multi-Display 테스트
- [ ] Left Half 연속 → 왼쪽 방향으로 이동, 순환
- [ ] Right Half 연속 → 오른쪽 방향으로 이동, 순환
- [ ] First Third 연속 → 왼쪽 방향으로 이동, 순환
- [ ] Last Third 연속 → 오른쪽 방향으로 이동, 순환
- [ ] First Two Thirds 연속 → 왼쪽 방향으로 이동, 순환
- [ ] Last Two Thirds 연속 → 오른쪽 방향으로 이동, 순환
- [ ] 모니터 이동 시 에지 맞춤 동작 확인
- [ ] 창 크기가 대상 모니터보다 클 때 맞춤 동작 확인

### 9.3 Restore 테스트
- [ ] Maximize 상태에서 Restore → 이전 크기로 복원
- [ ] 스냅 상태에서 Restore → 스냅 전 원래 위치로 복원
- [ ] 연속 스냅 후 Restore → 최초 스냅 전 상태로 복원
- [ ] Restore 후 다시 스냅 → 새로운 상태 저장

### 9.4 트레이 테스트
- [ ] 트레이 아이콘 표시됨
- [ ] 우클릭 → 메뉴 표시
- [ ] About 메뉴 동작
- [ ] 단축키 목록 메뉴 → 창 표시
- [ ] 시작 프로그램 토글 동작
- [ ] Exit 메뉴 → 프로그램 종료
- [ ] 콘솔 창 숨겨짐 확인

### 9.5 에러 처리 테스트
- [ ] 단축키 충돌 시 메시지 박스 표시
- [ ] 충돌 외 단축키 정상 동작

### 9.6 제거 확인
- [ ] ALT+WIN+C → 작동 안함
- [ ] ALT+WIN+A → 작동 안함 (Always On Top 제거됨)
- [ ] CTRL+ALT+WIN+Arrow → 작동 안함

---

## 10. 구현 순서

> **진행 규칙**: 각 Phase는 PR을 생성하여 리뷰/승인 후 main에 머지합니다.
> 다음 Phase는 이전 Phase PR이 머지된 후 시작합니다.

---

### Phase 1: 코드 정리 [x]

**브랜치**: `feature/phase1-cleanup`
**PR 제목**: `refactor: Phase 1 - 미사용 코드 정리`

#### 작업 항목
- [x] snap.go에서 사용하지 않는 함수들 삭제
  - topTwoThirds, topOneThirds, bottomTwoThirds, bottomOneThirds
  - topLeftTwoThirds, topLeftOneThirds, topRightTwoThirds, topRightOneThirds
  - bottomLeftTwoThirds, bottomLeftOneThirds, bottomRightTwoThirds, bottomRightOneThirds
- [x] main.go에서 사이클 인프라 삭제
  - lastResized 변수
  - edgeFuncs, edgeFuncTurn 변수
  - cornerFuncs, cornerFuncTurn 변수
  - cycleFuncs 클로저
  - cycleEdgeFuncs, cycleCornerFuncs 변수
- [x] toggleAlwaysOnTop 함수 삭제
- [x] 관련 단축키 제거 (ALT+WIN+C, ALT+WIN+A, CTRL+ALT+WIN+Arrow)

#### 구현 확인 방법
```bash
# 1. 빌드 성공 확인
GOOS=windows GOARCH=amd64 go build -o RectangleWin.exe .

# 2. 삭제된 함수가 없는지 확인
grep -r "topTwoThirds\|bottomTwoThirds\|toggleAlwaysOnTop\|cycleFuncs" *.go
# 결과: 없어야 함

# 3. 남아있는 함수 목록 확인 (snap.go)
grep "^func " snap.go
# 예상: leftHalf, rightHalf, topHalf, bottomHalf,
#       topLeftHalf, topRightHalf, bottomLeftHalf, bottomRightHalf,
#       leftOneThirds, leftTwoThirds, rightOneThirds, rightTwoThirds
```

#### PR 머지 조건
- [x] 빌드 성공 (에러 없음)
- [x] 삭제 대상 코드가 모두 제거됨
- [ ] 기존 기능 동작 확인 (수동 테스트)
  - CTRL+ALT+Arrow (4방향) 동작
  - CTRL+ALT+U/I/J/K 동작
- [ ] 코드 리뷰 승인

---

### Phase 2: 기본 기능 수정 [x]

**브랜치**: `feature/phase2-core-functions`
**PR 제목**: `feat: Phase 2 - Center/Size/Restore 기능 구현`

#### 작업 항목
- [x] center 함수를 75% 크기로 수정
- [x] makeSmaller 함수 해상도 비례(3%)로 수정
- [x] makeLarger 함수 해상도 비례(3%)로 수정
- [x] centerThird 함수 추가
- [x] savedStates 맵 추가 (스냅 전 상태 저장용)
- [x] resize 함수에 상태 저장 로직 추가 (첫 스냅 시에만)
- [x] restore 함수 구현 (통합 복원)

#### 구현 확인 방법
```bash
# 1. 빌드 성공 확인
GOOS=windows GOARCH=amd64 go build -o RectangleWin.exe .

# 2. 새 함수 존재 확인
grep "func centerThird\|func restore\|savedStates" *.go
# 결과: 모두 존재해야 함
```

**수동 테스트 체크리스트**:
| 테스트 | 기대 결과 | 확인 |
|--------|-----------|------|
| CTRL+ALT+C | 창이 화면 75% 크기로 중앙 배치 | [ ] |
| CTRL+ALT+- (5회 연속) | 창이 점점 작아짐 (해상도 비례) | [ ] |
| CTRL+ALT++ (5회 연속) | 창이 점점 커짐 (해상도 비례) | [ ] |
| CTRL+ALT+- (최소 크기 도달) | 100x100 이하로 줄어들지 않음 | [ ] |
| CTRL+ALT+F | 창이 화면 중앙 1/3에 배치 | [ ] |
| CTRL+ALT+ENTER → CTRL+ALT+BACKSPACE | 최대화 후 복원 | [ ] |
| CTRL+ALT+LEFT → CTRL+ALT+BACKSPACE | 스냅 후 원래 위치로 복원 | [ ] |
| 연속 스냅(LEFT→UP→RIGHT) → CTRL+ALT+BACKSPACE | 최초 스냅 전 위치로 복원 | [ ] |

#### PR 머지 조건
- [x] 빌드 성공
- [ ] 모든 수동 테스트 통과
- [ ] 코드 리뷰 승인

---

### Phase 3: 단축키 재정의 [x]

**브랜치**: `feature/phase3-hotkeys`
**PR 제목**: `feat: Phase 3 - 18개 단축키 재정의`

#### 작업 항목
- [x] 기존 hks 슬라이스 삭제
- [x] 새 hks 슬라이스 작성 (18개)
  - Halves: CTRL+ALT+LEFT/RIGHT/UP/DOWN (4개)
  - Maximize/Center/Restore: CTRL+ALT+ENTER/C/BACKSPACE (3개)
  - Corners: CTRL+ALT+U/I/J/K (4개)
  - Thirds: CTRL+ALT+D/F/G/E/T (5개)
  - Size: CTRL+ALT+-/+ (2개)
- [x] Hotkey ID 체계 적용 (섹션 8 참조)
- [x] 단축키 등록 코드 정리

#### 구현 확인 방법
```bash
# 1. 빌드 성공 확인
GOOS=windows GOARCH=amd64 go build -o RectangleWin.exe .

# 2. 단축키 개수 확인
grep -E "^\s+\{id:" main.go | wc -l
# 결과: 18
```

**전체 단축키 동작 테스트**:
| 카테고리 | 단축키 | 기능 | 확인 |
|----------|--------|------|------|
| Halves | CTRL+ALT+LEFT | Left Half | [ ] |
| Halves | CTRL+ALT+RIGHT | Right Half | [ ] |
| Halves | CTRL+ALT+UP | Top Half | [ ] |
| Halves | CTRL+ALT+DOWN | Bottom Half | [ ] |
| Max/Ctr/Rst | CTRL+ALT+ENTER | Maximize | [ ] |
| Max/Ctr/Rst | CTRL+ALT+C | Center (75%) | [ ] |
| Max/Ctr/Rst | CTRL+ALT+BACKSPACE | Restore | [ ] |
| Corners | CTRL+ALT+U | Top Left | [ ] |
| Corners | CTRL+ALT+I | Top Right | [ ] |
| Corners | CTRL+ALT+J | Bottom Left | [ ] |
| Corners | CTRL+ALT+K | Bottom Right | [ ] |
| Thirds | CTRL+ALT+D | First Third | [ ] |
| Thirds | CTRL+ALT+F | Center Third | [ ] |
| Thirds | CTRL+ALT+G | Last Third | [ ] |
| Thirds | CTRL+ALT+E | First Two Thirds | [ ] |
| Thirds | CTRL+ALT+T | Last Two Thirds | [ ] |
| Size | CTRL+ALT+- | Make Smaller | [ ] |
| Size | CTRL+ALT++ | Make Larger | [ ] |

#### PR 머지 조건
- [x] 빌드 성공
- [ ] 18개 단축키 모두 동작 확인
- [ ] 제거된 단축키 비동작 확인 (ALT+WIN+C, ALT+WIN+A 등)
- [ ] 코드 리뷰 승인

---

### Phase 4: Multi-Display 구현 [x]

**브랜치**: `feature/phase4-multi-display`
**PR 제목**: `feat: Phase 4 - 다중 모니터 지원`

#### 작업 항목
- [x] multimon.go 파일 생성
- [x] 모니터 열거 함수 구현 (EnumDisplayMonitors)
- [x] 현재 창이 속한 모니터 판별 함수 구현
- [x] 다음/이전 모니터 계산 함수 구현 (순환 포함)
- [x] 에지 맞춤 변환 함수 구현
- [x] 화면 경계 처리 (클리핑) 구현
- [x] 이동 지원 단축키에 Multi-Display 로직 적용
  - Left Half, Right Half
  - First Third, Last Third
  - First Two Thirds, Last Two Thirds

#### 구현 확인 방법
```bash
# 1. 빌드 성공 확인
GOOS=windows GOARCH=amd64 go build -o RectangleWin.exe .

# 2. multimon.go 파일 존재 확인
ls -la multimon.go

# 3. 필수 함수 존재 확인
grep "func.*Monitor\|func.*monitor" multimon.go
```

**다중 모니터 테스트** (2개 모니터 환경 필요):
| 시나리오 | 시작 위치 | 동작 | 기대 결과 | 확인 |
|----------|-----------|------|-----------|------|
| Left Half 연속 | 모니터2 중앙 | CTRL+ALT+LEFT x4 | 모니터2 Left → 모니터1 Right → 모니터1 Left → 모니터2 Right (순환) | [ ] |
| Right Half 연속 | 모니터1 중앙 | CTRL+ALT+RIGHT x4 | 모니터1 Right → 모니터2 Left → 모니터2 Right → 모니터1 Left (순환) | [ ] |
| First Third 연속 | 모니터2 중앙 | CTRL+ALT+D x4 | 모니터2 First → 모니터1 Last → 모니터1 First → 모니터2 Last (순환) | [ ] |
| Last Third 연속 | 모니터1 중앙 | CTRL+ALT+G x4 | 동일 패턴 반대 방향 | [ ] |
| 이동 미지원 | 어디서든 | CTRL+ALT+UP | 현재 모니터 Top Half (이동 없음) | [ ] |
| 이동 미지원 | 어디서든 | CTRL+ALT+U | 현재 모니터 Top Left (이동 없음) | [ ] |

**단일 모니터 테스트**:
| 시나리오 | 동작 | 기대 결과 | 확인 |
|----------|------|-----------|------|
| Left Half 연속 | CTRL+ALT+LEFT x3 | Left Half 유지 (변화 없음) | [ ] |
| Right Half 연속 | CTRL+ALT+RIGHT x3 | Right Half 유지 (변화 없음) | [ ] |

#### PR 머지 조건
- [x] 빌드 성공
- [x] 다중 모니터 환경에서 모든 테스트 통과
- [x] 단일 모니터 환경에서 정상 동작
- [x] 코드 리뷰 승인

---

### Phase 5: 시스템 트레이 구현 [~]

**브랜치**: `feature/phase5-system-tray`
**PR 제목**: `feat: Phase 5 - 시스템 트레이 및 UI`

#### 작업 항목
- [x] tray.go 파일 생성
- [x] 트레이 아이콘 리소스 준비 (icon.ico)
- [x] 트레이 아이콘 표시 구현
- [~] 컨텍스트 메뉴 구현
  - [x] About RectangleWin... (현재 "Documentation"으로 구현됨)
  - [ ] 단축키 목록... ← **미구현**
  - [x] Windows 시작 시 실행 (토글)
  - [x] Exit
- [ ] 단축키 목록 창 구현 ← **미구현**
- [x] 시작 프로그램 등록/해제 구현 (레지스트리)
- [x] 빌드 스크립트에 -H=windowsgui 추가 (CLAUDE.md에 문서화됨)
- [ ] Makefile 또는 build.bat 업데이트

#### 구현 확인 방법
```bash
# 1. GUI 빌드 성공 확인
GOOS=windows GOARCH=amd64 go build -ldflags "-H=windowsgui" -o RectangleWin.exe .

# 2. tray.go 파일 존재 확인
ls -la tray.go

# 3. 레지스트리 관련 함수 확인
grep "OpenKey\|SetStringValue\|DeleteValue" *.go
```

**시스템 트레이 테스트**:
| 테스트 항목 | 기대 결과 | 확인 |
|-------------|-----------|------|
| 프로그램 시작 | 콘솔 창 없이 시작, 트레이 아이콘 표시 | [ ] |
| 트레이 아이콘 hover | "RectangleWin" 툴팁 표시 | [ ] |
| 트레이 좌클릭 | 아무 동작 없음 | [ ] |
| 트레이 우클릭 | 컨텍스트 메뉴 표시 | [ ] |
| About 메뉴 클릭 | About 다이얼로그 표시 | [ ] |
| 단축키 목록 클릭 | 단축키 목록 창 표시 | [ ] |
| 단축키 목록 닫기 버튼 | 창 닫힘 | [ ] |
| 시작 프로그램 토글 ON | 체크 표시, 레지스트리에 등록 | [ ] |
| 시작 프로그램 토글 OFF | 체크 해제, 레지스트리에서 제거 | [ ] |
| Exit 클릭 | 프로그램 종료, 트레이 아이콘 제거 | [ ] |

**레지스트리 확인** (PowerShell):
```powershell
# 등록 확인
Get-ItemProperty -Path "HKCU:\Software\Microsoft\Windows\CurrentVersion\Run" -Name "RectangleWin"

# 제거 확인
Get-ItemProperty -Path "HKCU:\Software\Microsoft\Windows\CurrentVersion\Run" -Name "RectangleWin"
# 결과: 에러 (존재하지 않음)
```

#### PR 머지 조건
- [ ] GUI 빌드 성공 (콘솔 숨김)
- [ ] 트레이 아이콘 정상 표시
- [ ] 모든 메뉴 동작 확인
- [ ] 시작 프로그램 등록/해제 동작 확인
- [ ] 코드 리뷰 승인

---

### Phase 6: 에러 처리 [ ]

**브랜치**: `feature/phase6-error-handling`
**PR 제목**: `feat: Phase 6 - 단축키 충돌 처리`

#### 작업 항목
- [ ] 단축키 등록 실패 감지 로직 추가
- [ ] 충돌 단축키 목록 수집
- [ ] 메시지 박스 표시 구현
- [ ] 충돌 외 단축키 정상 등록 보장

#### 구현 확인 방법
```bash
# 1. 빌드 성공 확인
GOOS=windows GOARCH=amd64 go build -ldflags "-H=windowsgui" -o RectangleWin.exe .
```

**에러 처리 테스트**:
| 테스트 시나리오 | 방법 | 기대 결과 | 확인 |
|-----------------|------|-----------|------|
| 정상 시작 | 단축키 충돌 없이 시작 | 메시지 박스 없이 트레이에 등록 | [ ] |
| 단축키 충돌 | 다른 프로그램에서 CTRL+ALT+C 사용 후 RectangleWin 시작 | 충돌 메시지 박스 표시, 나머지 단축키 동작 | [ ] |
| 여러 단축키 충돌 | 여러 단축키 충돌 상황 생성 | 모든 충돌 단축키 목록이 메시지에 표시 | [ ] |

**충돌 시뮬레이션 방법**:
1. AutoHotkey 또는 다른 프로그램으로 CTRL+ALT+C 등록
2. RectangleWin 시작
3. 메시지 박스 확인
4. 다른 단축키 동작 확인

#### PR 머지 조건
- [ ] 빌드 성공
- [ ] 단축키 충돌 시 메시지 박스 정상 표시
- [ ] 충돌 외 단축키 정상 동작
- [ ] 코드 리뷰 승인

---

### Phase 7: 최종 테스트 및 문서화 [ ]

**브랜치**: `feature/phase7-final`
**PR 제목**: `docs: Phase 7 - 최종 검증 및 README 업데이트`

#### 작업 항목
- [ ] 전체 기능 통합 테스트
- [ ] README.md 업데이트 (단축키 목록, 설치 방법)
- [ ] CHANGELOG.md 작성 (선택)
- [ ] 릴리즈 빌드 생성

#### 구현 확인 방법

**전체 기능 체크리스트** (섹션 9 참조):

**9.1 기본 기능 테스트**:
- [ ] CTRL+ALT+LEFT → 왼쪽 반
- [ ] CTRL+ALT+RIGHT → 오른쪽 반
- [ ] CTRL+ALT+UP → 위쪽 반
- [ ] CTRL+ALT+DOWN → 아래쪽 반
- [ ] CTRL+ALT+ENTER → 최대화
- [ ] CTRL+ALT+C → 75% 크기로 중앙 배치
- [ ] CTRL+ALT+BACKSPACE → 복원
- [ ] CTRL+ALT+U → 좌상단 1/4
- [ ] CTRL+ALT+I → 우상단 1/4
- [ ] CTRL+ALT+J → 좌하단 1/4
- [ ] CTRL+ALT+K → 우하단 1/4
- [ ] CTRL+ALT+D → 왼쪽 1/3
- [ ] CTRL+ALT+F → 중앙 1/3
- [ ] CTRL+ALT+G → 오른쪽 1/3
- [ ] CTRL+ALT+E → 왼쪽 2/3
- [ ] CTRL+ALT+T → 오른쪽 2/3
- [ ] CTRL+ALT+- → 창 축소
- [ ] CTRL+ALT++ → 창 확대

**9.2 Multi-Display 테스트**:
- [ ] Left Half 연속 → 왼쪽 방향 이동, 순환
- [ ] Right Half 연속 → 오른쪽 방향 이동, 순환
- [ ] First Third/Last Third 이동 테스트
- [ ] First Two Thirds/Last Two Thirds 이동 테스트
- [ ] 에지 맞춤 동작 확인

**9.3 Restore 테스트**:
- [ ] Maximize 상태에서 Restore
- [ ] 스냅 상태에서 Restore
- [ ] 연속 스냅 후 Restore

**9.4 트레이 테스트**:
- [ ] 트레이 아이콘 표시
- [ ] 메뉴 동작
- [ ] 시작 프로그램 토글

**9.5 에러 처리 테스트**:
- [ ] 단축키 충돌 시 메시지 박스

**9.6 제거 확인**:
- [ ] ALT+WIN+C → 작동 안함
- [ ] ALT+WIN+A → 작동 안함

#### PR 머지 조건
- [ ] 모든 테스트 항목 통과
- [ ] README.md 업데이트 완료
- [ ] 릴리즈 빌드 생성 및 테스트
- [ ] 코드 리뷰 승인

---

### 진행 상황 요약

| Phase | 설명 | 상태 | PR |
|-------|------|------|-----|
| Phase 1 | 코드 정리 | [x] 완료 | [#4](https://github.com/nicewook/RectangleWin/pull/4) |
| Phase 2 | 기본 기능 수정 | [x] 완료 | [#7](https://github.com/nicewook/RectangleWin/pull/7) |
| Phase 3 | 단축키 재정의 | [x] 완료 | [#8](https://github.com/nicewook/RectangleWin/pull/8) |
| Phase 4 | Multi-Display | [x] 완료 | [#10](https://github.com/nicewook/RectangleWin/pull/10) |
| Phase 5 | 시스템 트레이 | [~] 일부 구현 | - |
| Phase 6 | 에러 처리 | [ ] 대기 | - |
| Phase 7 | 최종 테스트 | [ ] 대기 | - |

---

## 11. Virtual Key Codes 참조

### w32 라이브러리 상수
| 키 | 상수 | 값 |
|----|------|-----|
| LEFT | `w32.VK_LEFT` | 0x25 |
| RIGHT | `w32.VK_RIGHT` | 0x27 |
| UP | `w32.VK_UP` | 0x26 |
| DOWN | `w32.VK_DOWN` | 0x28 |
| ENTER | `w32.VK_RETURN` | 0x0D |
| BACKSPACE | `w32.VK_BACK` | 0x08 |
| - | `w32.VK_OEM_MINUS` | 0xBD |
| + | `w32.VK_OEM_PLUS` | 0xBB |

### 문자 키 (hex 값)
| 키 | 값 |
|----|-----|
| C | 0x43 |
| U | 0x55 |
| I | 0x49 |
| J | 0x4A |
| K | 0x4B |
| D | 0x44 |
| F | 0x46 |
| G | 0x47 |
| E | 0x45 |
| T | 0x54 |
