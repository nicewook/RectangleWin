// Copyright 2022 Ahmet Alp Balkan
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// Modified by hsjeong on 2026-11-08
// Changes: New file (fork addition). Integration tests against real Win32
//
//	windows created by the test process (no other windows are touched).
//	Skip with RECTANGLEWIN_SKIP_WIN32=1.

package main

import (
	"os"
	"runtime"
	"syscall"
	"testing"
	"unsafe"

	"github.com/ahmetb/RectangleWin/w32ex"
	"github.com/gonutz/w32/v2"
)

func TestMain(m *testing.M) {
	// Same DPI setup as main(): coordinates must be physical pixels.
	w32ex.SetProcessDPIAware()
	os.Exit(m.Run())
}

var testWndProc = syscall.NewCallback(func(hwnd, msg, wParam, lParam uintptr) uintptr {
	return w32.DefWindowProc(w32.HWND(hwnd), uint32(msg), wParam, lParam)
})

const testClassName = "RectangleWinTestWindow"

var testClassRegistered bool

func pumpMessages() {
	var m w32.MSG
	for w32.PeekMessage(&m, 0, 0, 0, w32.PM_REMOVE) {
		w32.TranslateMessage(&m)
		w32.DispatchMessage(&m)
	}
}

// newTestWindow creates a visible, normal top-level window owned by the test.
func newTestWindow(t *testing.T, x, y, w, h int) w32.HWND {
	t.Helper()
	if os.Getenv("RECTANGLEWIN_SKIP_WIN32") != "" {
		t.Skip("RECTANGLEWIN_SKIP_WIN32 set")
	}
	// Each test runs in its own goroutine. A window must be driven from the
	// thread that created it, otherwise cross-thread SendMessage deadlocks
	// because nobody pumps the creating thread. Lock for the test's lifetime
	// (Cleanup runs on the same goroutine); the thread exits with it.
	runtime.LockOSThread()
	cls, _ := syscall.UTF16PtrFromString(testClassName)
	inst := w32.GetModuleHandle("")
	if !testClassRegistered {
		wc := w32.WNDCLASSEX{WndProc: testWndProc, Instance: inst, ClassName: cls}
		wc.Size = uint32(unsafe.Sizeof(wc))
		if w32.RegisterClassEx(&wc) == 0 {
			t.Skipf("RegisterClassEx failed: %d", w32.GetLastError())
		}
		testClassRegistered = true
	}
	title, _ := syscall.UTF16PtrFromString("RectangleWin test")
	hwnd := w32.CreateWindowEx(0, cls, title, w32.WS_OVERLAPPEDWINDOW, x, y, w, h, 0, 0, inst, nil)
	if hwnd == 0 {
		t.Skipf("CreateWindowEx failed: %d", w32.GetLastError())
	}
	t.Cleanup(func() {
		w32.DestroyWindow(hwnd)
		pumpMessages()
		delete(savedStates, hwnd)
	})
	w32.ShowWindow(hwnd, w32.SW_SHOWNORMAL)
	pumpMessages()
	if ok, _ := w32.DwmGetWindowAttributeEXTENDED_FRAME_BOUNDS(hwnd); !ok {
		t.Skip("DWM frame bounds unavailable (no desktop composition)")
	}
	return hwnd
}

func workAreaOf(t *testing.T, hwnd w32.HWND) w32.RECT {
	t.Helper()
	var mi w32.MONITORINFO
	if !w32.GetMonitorInfo(w32.MonitorFromWindow(hwnd, w32.MONITOR_DEFAULTTONEAREST), &mi) {
		t.Fatalf("GetMonitorInfo failed")
	}
	return mi.RcWork
}

func visibleFrame(t *testing.T, hwnd w32.HWND) w32.RECT {
	t.Helper()
	ok, f := w32.DwmGetWindowAttributeEXTENDED_FRAME_BOUNDS(hwnd)
	if !ok {
		t.Fatal("DwmGetWindowAttribute failed")
	}
	return f
}

func snapOnCurrent(t *testing.T, hwnd w32.HWND, f resizeFunc) {
	t.Helper()
	if _, err := resize(hwnd, f); err != nil {
		t.Fatalf("resize: %v", err)
	}
	pumpMessages()
}

func TestWin32TestWindowIsZonable(t *testing.T) {
	hwnd := newTestWindow(t, 100, 100, 800, 600)
	if !isZonableWindow(hwnd) {
		t.Fatal("a normal overlapped window must be zonable")
	}
	if isZonableWindow(0) || isZonableWindow(w32.GetDesktopWindow()) {
		t.Error("NULL/desktop window must not be zonable")
	}
}

// The visible frame must land exactly on the layout target, regardless of the
// window's DPI (regression for the removed double DPI correction).
func TestWin32SnapPlacesVisibleFrameExactly(t *testing.T) {
	hwnd := newTestWindow(t, 120, 140, 900, 650)
	work := workAreaOf(t, hwnd)
	for name, f := range map[string]resizeFunc{
		"leftHalf": leftHalf, "rightHalf": rightHalf, "topHalf": topHalf, "bottomHalf": bottomHalf,
		"topLeft": topLeftHalf, "topRight": topRightHalf, "bottomLeft": bottomLeftHalf, "bottomRight": bottomRightHalf,
		"center": center,
	} {
		snapOnCurrent(t, hwnd, f)
		want := f(work, w32.RECT{})
		if got := visibleFrame(t, hwnd); got != want {
			t.Errorf("%s: visible frame %v, want %v (window DPI %d)", name, got, want, w32ex.GetDpiForWindow(hwnd))
		}
	}
}

func TestWin32SnapIsIdempotent(t *testing.T) {
	hwnd := newTestWindow(t, 150, 150, 700, 500)
	snapOnCurrent(t, hwnd, leftHalf)
	moved, err := resize(hwnd, leftHalf)
	if err != nil || moved {
		t.Errorf("second identical snap: moved=%v err=%v, want no-op", moved, err)
	}
}

func TestWin32SnapMaximizedWindow(t *testing.T) {
	hwnd := newTestWindow(t, 200, 200, 800, 600)
	w32.ShowWindow(hwnd, w32.SW_MAXIMIZE)
	pumpMessages()
	if !w32ex.IsZoomed(hwnd) {
		t.Skip("could not maximize test window")
	}

	snapOnCurrent(t, hwnd, leftHalf)
	if w32ex.IsZoomed(hwnd) {
		t.Fatal("window still maximized after snap")
	}
	work := workAreaOf(t, hwnd)
	if got, want := visibleFrame(t, hwnd), leftHalf(work, w32.RECT{}); got != want {
		t.Errorf("visible frame %v, want %v (top edge must not be cut off)", got, want)
	}

	// Restore brings back the maximized state, not a screen-sized normal window.
	if err := restoreWindow(hwnd); err != nil {
		t.Fatal(err)
	}
	pumpMessages()
	if !w32ex.IsZoomed(hwnd) {
		t.Error("restore after snapping a maximized window must maximize it again")
	}
}

// Codex review: a maximized window whose normal bounds already equal the
// target used to hit the no-op path and lose its maximized state.
func TestWin32SnapMaximizedWindowAlreadyAtTargetNormalBounds(t *testing.T) {
	hwnd := newTestWindow(t, 100, 100, 600, 400)
	snapOnCurrent(t, hwnd, topHalf) // normal bounds = top half
	delete(savedStates, hwnd)       // pretend this is the user's own layout
	w32.ShowWindow(hwnd, w32.SW_MAXIMIZE)
	pumpMessages()
	if !w32ex.IsZoomed(hwnd) {
		t.Skip("could not maximize test window")
	}

	moved, err := resize(hwnd, topHalf)
	if err != nil {
		t.Fatal(err)
	}
	pumpMessages()
	if !moved || w32ex.IsZoomed(hwnd) {
		t.Fatalf("moved=%v zoomed=%v, want un-maximized top half", moved, w32ex.IsZoomed(hwnd))
	}
	if err := restoreWindow(hwnd); err != nil {
		t.Fatal(err)
	}
	pumpMessages()
	if !w32ex.IsZoomed(hwnd) {
		t.Error("restore must bring back the maximized state")
	}
}

// Codex review: maximized window with normal bounds = right half must snap on
// its own monitor, not cycle (single-monitor safe version of the check).
func TestWin32MaximizedWindowSnapsOnCurrentMonitor(t *testing.T) {
	hwnd := newTestWindow(t, 100, 100, 600, 400)
	snapOnCurrent(t, hwnd, rightHalf)
	delete(savedStates, hwnd)
	start := w32.MonitorFromWindow(hwnd, w32.MONITOR_DEFAULTTONEAREST)
	w32.ShowWindow(hwnd, w32.SW_MAXIMIZE)
	pumpMessages()
	if !w32ex.IsZoomed(hwnd) {
		t.Skip("could not maximize test window")
	}

	moved, err := resizeWithMultiDisplay(hwnd, SnapRightHalf)
	if err != nil {
		t.Fatal(err)
	}
	pumpMessages()
	if !moved || w32ex.IsZoomed(hwnd) {
		t.Fatalf("moved=%v zoomed=%v, want un-maximized right half", moved, w32ex.IsZoomed(hwnd))
	}
	if w32.MonitorFromWindow(hwnd, w32.MONITOR_DEFAULTTONEAREST) != start {
		t.Error("maximized window jumped to another monitor")
	}
	if got, want := visibleFrame(t, hwnd), rightHalf(workAreaOf(t, hwnd), w32.RECT{}); got != want {
		t.Errorf("frame %v, want right half %v", got, want)
	}
}

func TestWin32RestoreReturnsToPositionBeforeFirstSnap(t *testing.T) {
	hwnd := newTestWindow(t, 210, 230, 820, 610)
	orig := *w32.GetWindowRect(hwnd)

	snapOnCurrent(t, hwnd, leftHalf)
	snapOnCurrent(t, hwnd, bottomRightHalf) // second snap must not overwrite the saved state

	if err := restoreWindow(hwnd); err != nil {
		t.Fatal(err)
	}
	pumpMessages()
	if got := *w32.GetWindowRect(hwnd); got != orig {
		t.Errorf("restored to %v, want original %v", got, orig)
	}
	if _, ok := savedStates[hwnd]; ok {
		t.Error("saved state must be consumed by restore")
	}
}

// original → left snap → maximize → un-maximize → Restore must return to the
// original position (the toggle used to overwrite it).
func TestWin32ToggleMaximizeKeepsOriginalPosition(t *testing.T) {
	hwnd := newTestWindow(t, 260, 180, 760, 540)
	orig := *w32.GetWindowRect(hwnd)
	work := workAreaOf(t, hwnd)

	snapOnCurrent(t, hwnd, leftHalf)

	if err := toggleMaximizeWindow(hwnd); err != nil {
		t.Fatal(err)
	}
	pumpMessages()
	if !w32ex.IsZoomed(hwnd) {
		t.Skip("could not maximize test window")
	}
	if err := toggleMaximizeWindow(hwnd); err != nil {
		t.Fatal(err)
	}
	pumpMessages()
	if w32ex.IsZoomed(hwnd) {
		t.Fatal("second toggle must un-maximize")
	}
	if got, want := visibleFrame(t, hwnd), leftHalf(work, w32.RECT{}); got != want {
		t.Errorf("after toggle: frame %v, want left half %v", got, want)
	}

	if err := restoreWindow(hwnd); err != nil {
		t.Fatal(err)
	}
	pumpMessages()
	if got := *w32.GetWindowRect(hwnd); got != orig {
		t.Errorf("restore: %v, want original %v", got, orig)
	}
}

func TestWin32RestoreIgnoresStateOfDestroyedWindow(t *testing.T) {
	a := newTestWindow(t, 100, 100, 600, 400)
	saveStateIfAbsent(a, placement(w32.SW_SHOWNORMAL, 1, 2, 300, 300), true)
	w32.DestroyWindow(a)
	pumpMessages()

	b := newTestWindow(t, 300, 300, 500, 400)
	saveStateIfAbsent(b, placement(w32.SW_SHOWNORMAL, 0, 0, 100, 100), true) // triggers pruning
	if _, ok := savedStates[a]; ok && a != b {
		t.Error("state of destroyed window was not pruned")
	}
}

// Regression: each EnumMonitors call used to leak a syscall callback and the
// runtime aborts the process after ~2000 of them.
func TestWin32EnumMonitorsDoesNotLeakCallbacks(t *testing.T) {
	for i := 0; i < 5000; i++ {
		if len(getMonitorList()) == 0 {
			t.Fatalf("call %d: no monitors", i)
		}
	}
}

func TestWin32MonitorListMatchesSystem(t *testing.T) {
	ms := getMonitorList()
	if n := w32.GetSystemMetrics(w32.SM_CMONITORS); len(ms) != n {
		t.Errorf("getMonitorList: %d monitors, SM_CMONITORS = %d", len(ms), n)
	}
	for i := 1; i < len(ms); i++ {
		if ms[i-1].Work.Left > ms[i].Work.Left {
			t.Errorf("monitors not sorted left to right: %v", ms)
		}
	}
}

func TestWin32MultiDisplayMove(t *testing.T) {
	if w32.GetSystemMetrics(w32.SM_CMONITORS) < 2 {
		t.Skip("needs at least two monitors")
	}
	hwnd := newTestWindow(t, 100, 100, 800, 600)
	start := w32.MonitorFromWindow(hwnd, w32.MONITOR_DEFAULTTONEAREST)
	for i := 0; i < 2; i++ {
		if _, err := resizeWithMultiDisplay(hwnd, SnapRightHalf); err != nil {
			t.Fatal(err)
		}
		pumpMessages()
	}
	if w32.MonitorFromWindow(hwnd, w32.MONITOR_DEFAULTTONEAREST) == start {
		t.Error("second Right Half did not move the window to another monitor")
	}
	work := workAreaOf(t, hwnd)
	if got, want := visibleFrame(t, hwnd), leftHalf(work, w32.RECT{}); got != want {
		t.Errorf("after moving: frame %v, want edge-aligned left half %v", got, want)
	}
}

// Codex review (round 2): maximize -> un-maximize -> move by hand -> snap ->
// Restore must return to the hand-moved position, not the pre-maximize one.
func TestWin32CompletedMaximizeToggleLeavesNoHistory(t *testing.T) {
	hwnd := newTestWindow(t, 150, 160, 700, 500)
	for i := 0; i < 2; i++ {
		if err := toggleMaximizeWindow(hwnd); err != nil {
			t.Fatal(err)
		}
		pumpMessages()
	}
	if w32ex.IsZoomed(hwnd) {
		t.Skip("could not toggle maximize")
	}
	w32.SetWindowPos(hwnd, 0, 400, 300, 640, 480, w32.SWP_NOZORDER|w32.SWP_NOACTIVATE)
	pumpMessages()
	moved := *w32.GetWindowRect(hwnd)

	snapOnCurrent(t, hwnd, leftHalf)
	if err := restoreWindow(hwnd); err != nil {
		t.Fatal(err)
	}
	pumpMessages()
	if got := *w32.GetWindowRect(hwnd); got != moved {
		t.Errorf("restored to %v, want hand-moved position %v", got, moved)
	}
}

// maximize -> un-maximize -> move by hand -> Restore (no snap) must not jump
// back to the pre-maximize position.
func TestWin32RestoreAfterCompletedToggleDoesNotJump(t *testing.T) {
	hwnd := newTestWindow(t, 150, 160, 700, 500)
	for i := 0; i < 2; i++ {
		if err := toggleMaximizeWindow(hwnd); err != nil {
			t.Fatal(err)
		}
		pumpMessages()
	}
	if w32ex.IsZoomed(hwnd) {
		t.Skip("could not toggle maximize")
	}
	w32.SetWindowPos(hwnd, 0, 420, 310, 640, 480, w32.SWP_NOZORDER|w32.SWP_NOACTIVATE)
	pumpMessages()
	moved := *w32.GetWindowRect(hwnd)

	if err := restoreWindow(hwnd); err != nil {
		t.Fatal(err)
	}
	pumpMessages()
	if got := *w32.GetWindowRect(hwnd); got != moved {
		t.Errorf("restore jumped to %v, want to stay at %v", got, moved)
	}
}
