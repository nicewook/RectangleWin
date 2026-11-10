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

// TODO make it possible to "go generate" on Windows (https://github.com/josephspurrier/goversioninfo/issues/52).
//go:generate /bin/bash -c "go run github.com/josephspurrier/goversioninfo/cmd/goversioninfo@v1.7.0 -platform-specific -icon=assets/icon.ico - <<< '{}'"

package main

import (
	"errors"
	"fmt"
	"os"
	"os/signal"
	"runtime"
	"strings"

	"fyne.io/systray"
	"github.com/gonutz/w32/v2"

	"github.com/ahmetb/RectangleWin/w32ex"
)

// Modified by hsjeong on 2026-02-09
// Changes: Added hotkey ID allocation constants and documentation
// Modified by hsjeong on 2026-02-10
// Changes: Added toggle maximize/restore functionality (Ctrl+Alt+Enter);
//
//	Added first-run autorun enable logic
//
// Modified by hsjeong on 2026-11-08
// Changes: Removed double DPI correction (process is Per-Monitor v2 aware);
//
//	normalize maximized windows before measuring; MOD_NOREPEAT on all hotkeys;
//	window-state save/restore moved to windowstate.go; autorun path sync;
//	file logging for GUI builds
//
// Hotkey ID allocation strategy:
// IDs are organized into functional groups with gaps for future additions
// - 1-9:    Edge snaps (Halves)
// - 10-19:  Window states (Maximize, Center, Restore)
// - 20-29:  Corner snaps
// - 100+:   Reserved for future use
const (
	hotkeyIDHalvesMin  = 1
	hotkeyIDHalvesMax  = 9
	hotkeyIDStatesMin  = 10
	hotkeyIDStatesMax  = 19
	hotkeyIDCornersMin = 20
	hotkeyIDCornersMax = 29
)

func main() {
	runtime.LockOSThread() // since we bind hotkeys etc that need to dispatch their message here
	setupLogging()
	if !w32ex.SetProcessDPIAware() {
		fmt.Println("warn: failed to set DPI aware, continuing without DPI awareness")
	}

	autorun, err := SyncAutoRun()
	if err != nil {
		fmt.Printf("warn: failed to sync autorun: %v\n", err)
	}
	fmt.Printf("autorun enabled=%v\n", autorun)
	printMonitors()

	hks := defaultHotKeys()

	var failedHotKeys []HotKey
	for _, hk := range hks {
		ok, err := RegisterHotKey(hk)
		if err != nil {
			fmt.Printf("warn: %v\n", err)
			continue
		}
		if !ok {
			failedHotKeys = append(failedHotKeys, hk)
		}
	}
	if len(failedHotKeys) > 0 {
		msg := "The following hotkey(s) are in use by another process:\n\n"
		for _, hk := range failedHotKeys {
			msg += "  - " + hk.Describe() + "\n"
		}
		msg += "\nTo use these hotkeys in RectangleWin, close the other process using the key combination(s)."
		showMessageBox(msg)
	}

	exitCh := make(chan os.Signal, 1)
	signal.Notify(exitCh, os.Interrupt)
	go func() {
		<-exitCh
		fmt.Println("exit signal received")
		systray.Quit() // causes WM_CLOSE, WM_QUIT, not sure if a side-effect
	}()

	// TODO systray/systray.go already locks the OS thread in init()
	// however it's not clear if GetMessage(0,0) will continue to work
	// as we run "go initTray()" and not pin the thread that initializes the
	// tray.
	initTray()
	if err := msgLoop(); err != nil {
		fmt.Printf("fatal: message loop error: %v\n", err)
		os.Exit(1)
	}
}

// MOD_NOREPEAT: holding a combination down must not fire it repeatedly
// (repeated maximize toggles / monitor cycling).
const hotkeyMod = MOD_CONTROL | MOD_ALT | MOD_NOREPEAT

// defaultHotKeys is the single source of truth for all registered hotkeys.
func defaultHotKeys() []HotKey {
	return []HotKey{
		// ===== Halves (ID: 1-4) =====
		// Left/Right Half support multi-display movement
		{id: 1, mod: hotkeyMod, vk: w32.VK_LEFT, name: "Left Half", callback: multiDisplayResize(SnapLeftHalf, "Left Half")},
		{id: 2, mod: hotkeyMod, vk: w32.VK_RIGHT, name: "Right Half", callback: multiDisplayResize(SnapRightHalf, "Right Half")},
		{id: 3, mod: hotkeyMod, vk: w32.VK_UP, name: "Top Half", callback: simpleResize(topHalf, "Top Half")},
		{id: 4, mod: hotkeyMod, vk: w32.VK_DOWN, name: "Bottom Half", callback: simpleResize(bottomHalf, "Bottom Half")},

		// ===== Maximize / Center / Restore (ID: 10-12) =====
		{id: 10, mod: hotkeyMod, vk: w32.VK_RETURN /*Enter*/, name: "Maximize", callback: simpleAction(toggleMaximize, "Toggle Maximize")},
		{id: 11, mod: hotkeyMod, vk: 'C', name: "Center", callback: simpleResize(center, "Center")},
		{id: 12, mod: hotkeyMod, vk: w32.VK_BACK /*Backspace*/, name: "Restore", callback: simpleAction(restore, "Restore")},

		// ===== Corners (ID: 20-23) =====
		{id: 20, mod: hotkeyMod, vk: 'U', name: "Top Left", callback: simpleResize(topLeftHalf, "Top Left")},
		{id: 21, mod: hotkeyMod, vk: 'I', name: "Top Right", callback: simpleResize(topRightHalf, "Top Right")},
		{id: 22, mod: hotkeyMod, vk: 'J', name: "Bottom Left", callback: simpleResize(bottomLeftHalf, "Bottom Left")},
		{id: 23, mod: hotkeyMod, vk: 'K', name: "Bottom Right", callback: simpleResize(bottomRightHalf, "Bottom Right")},
	}
}

// simpleResize snaps the foreground window on its current monitor.
func simpleResize(f resizeFunc, name string) func() {
	return func() {
		fmt.Printf("Hotkey: %s\n", name)
		hwnd := w32.GetForegroundWindow()
		if hwnd == 0 {
			fmt.Println("warn: foreground window is NULL")
			return
		}
		if _, err := resize(hwnd, f); err != nil {
			fmt.Printf("warn: resize: %v\n", err)
		}
	}
}

// multiDisplayResize snaps the foreground window, moving it to the adjacent
// monitor when it is already at the requested position.
func multiDisplayResize(pos SnapPosition, name string) func() {
	return func() {
		fmt.Printf("Hotkey: %s\n", name)
		hwnd := w32.GetForegroundWindow()
		if hwnd == 0 {
			fmt.Println("warn: foreground window is NULL")
			return
		}
		if _, err := resizeWithMultiDisplay(hwnd, pos); err != nil {
			fmt.Printf("warn: resize: %v\n", err)
		}
	}
}

// simpleAction wraps maximize/restore style actions.
func simpleAction(action func() error, name string) func() {
	return func() {
		fmt.Printf("Hotkey: %s\n", name)
		if err := action(); err != nil {
			fmt.Printf("warn: %s: %v\n", strings.ToLower(name), err)
		}
	}
}

func showMessageBox(text string) {
	w32.MessageBox(w32.GetActiveWindow(), text, "RectangleWin", w32.MB_ICONWARNING|w32.MB_OK)
}

type resizeFunc func(disp, cur w32.RECT) w32.RECT

// center - 창을 화면의 75% 크기로 리사이즈하고 중앙에 배치
func center(disp, _ w32.RECT) w32.RECT {
	width := disp.Width() * 3 / 4   // 75%
	height := disp.Height() * 3 / 4 // 75%
	return w32.RECT{
		Left:   disp.Left + (disp.Width()-width)/2,
		Top:    disp.Top + (disp.Height()-height)/2,
		Right:  disp.Left + (disp.Width()+width)/2,
		Bottom: disp.Top + (disp.Height()+height)/2,
	}
}

// snapRect computes the outer window rectangle (what SetWindowPos expects)
// that places the window's visible frame at f(work, frame).
//
// rect is the window rect (GetWindowRect) and frame is the visible frame
// (DWMWA_EXTENDED_FRAME_BOUNDS). The process is Per-Monitor v2 DPI aware, so
// both are in physical pixels and their difference is exactly the invisible
// resize border; no DPI scaling must be applied.
func snapRect(rect, frame, work w32.RECT, f resizeFunc) w32.RECT {
	newPos := f(work, frame)
	newPos.Left -= frame.Left - rect.Left
	newPos.Top -= frame.Top - rect.Top
	newPos.Right += rect.Right - frame.Right
	newPos.Bottom += rect.Bottom - frame.Bottom
	return newPos
}

// targetChooser decides where to snap given the window's visible frame.
// wasMaximized reports that the window was maximized before snapWindow
// normalized it (its frame is then the normal, not the visible, position).
// It returns the work area to snap into, the layout function and whether to
// proceed at all.
type targetChooser func(hwnd w32.HWND, frame w32.RECT, wasMaximized bool) (w32.RECT, resizeFunc, bool)

func resize(hwnd w32.HWND, f resizeFunc) (bool, error) {
	return snapWindow(hwnd, func(hwnd w32.HWND, _ w32.RECT, _ bool) (w32.RECT, resizeFunc, bool) {
		mon := w32.MonitorFromWindow(hwnd, w32.MONITOR_DEFAULTTONEAREST)
		var monInfo w32.MONITORINFO
		if !w32.GetMonitorInfo(mon, &monInfo) {
			fmt.Printf("warn: failed to GetMonitorInfo:%d\n", w32.GetLastError())
			return w32.RECT{}, nil, false
		}
		return monInfo.RcWork, f, true
	})
}

// resizeWithMultiDisplay handles snap with multi-display support
func resizeWithMultiDisplay(hwnd w32.HWND, pos SnapPosition) (bool, error) {
	return snapWindow(hwnd, func(hwnd w32.HWND, frame w32.RECT, wasMaximized bool) (w32.RECT, resizeFunc, bool) {
		// A maximized window is not "at" a half position even if its normal
		// bounds are, so it must snap on its current monitor, not cycle.
		work, f, ok := multiDisplaySnap(hwnd, pos, frame, !wasMaximized)
		if !ok {
			fmt.Println("no movement (single monitor or already at position)")
		}
		return work, f, ok
	})
}

func snapWindow(hwnd w32.HWND, choose targetChooser) (bool, error) {
	if !isZonableWindow(hwnd) {
		fmt.Printf("warn: non-zonable window: %s\n", w32.GetWindowText(hwnd))
		return false, nil
	}

	// Capture the placement before touching the window so that Restore can
	// bring a maximized window back as maximized.
	placement, havePlacement := getPlacement(hwnd)

	// A maximized window reports maximized bounds and borders; normalize it
	// first and measure the normal window instead.
	wasMaximized := w32ex.IsZoomed(hwnd)
	if wasMaximized {
		w32.ShowWindow(hwnd, w32.SW_SHOWNORMAL) // return value is previous visibility, not success
	}

	rect := w32.GetWindowRect(hwnd)
	if rect == nil {
		return false, fmt.Errorf("failed to GetWindowRect:%d", w32.GetLastError())
	}
	ok, frame := w32.DwmGetWindowAttributeEXTENDED_FRAME_BOUNDS(hwnd)
	if !ok {
		// No DWM frame info: treat the window rect as the visible frame.
		fmt.Printf("warn: DwmGetWindowAttribute failed (%d), assuming no invisible borders\n", w32.GetLastError())
		frame = *rect
	}

	work, f, proceed := choose(hwnd, frame, wasMaximized)
	if !proceed {
		return false, nil
	}
	newPos := snapRect(*rect, frame, work, f)

	fmt.Printf("> window: 0x%x %#v (w:%d,h:%d)\n", hwnd, *rect, rect.Width(), rect.Height())
	fmt.Printf("> DWM frame:   %#v (W:%d,H:%d)\n", frame, frame.Width(), frame.Height())
	fmt.Printf("> work area:   %#v\n", work)

	// Save before the no-op check: if normalization alone already put the
	// window at the target, it still changed state (maximized -> snapped) and
	// Restore must be able to bring the maximized state back.
	if havePlacement && (wasMaximized || !sameRect(rect, &newPos)) {
		saveStateIfAbsent(hwnd, placement, true)
	}

	if sameRect(rect, &newPos) {
		fmt.Println("no resize")
		return wasMaximized, nil
	}

	fmt.Printf("> resizing to: %#v (W:%d,H:%d)\n", newPos, newPos.Width(), newPos.Height())
	if !w32.SetWindowPos(hwnd, 0, int(newPos.Left), int(newPos.Top), int(newPos.Width()), int(newPos.Height()), w32.SWP_NOZORDER|w32.SWP_NOACTIVATE) {
		return false, fmt.Errorf("failed to SetWindowPos:%d", w32.GetLastError())
	}
	if post := w32.GetWindowRect(hwnd); post != nil {
		fmt.Printf("> post-resize: %#v(W:%d,H:%d)\n", *post, post.Width(), post.Height())
	}
	return true, nil
}

// restore - 통합 복원 함수
// 1. 최대화 상태 → SW_RESTORE (저장된 상태는 유지)
// 2. 저장된 상태 있음 → 첫 스냅 전 상태(위치·최대화 여부)로 복원
// 3. 그 외 → SW_RESTORE
func restore() error {
	return restoreWindow(w32.GetForegroundWindow())
}

func restoreWindow(hwnd w32.HWND) error {
	if !isZonableWindow(hwnd) {
		return errors.New("window is not zonable")
	}

	if w32ex.IsZoomed(hwnd) {
		fmt.Println("Restore: window is maximized, calling SW_RESTORE")
		w32.ShowWindow(hwnd, w32.SW_RESTORE)
		dropMaximizeOnlyState(hwnd)
		return nil
	}

	if placement, ok := takeSavedState(hwnd); ok {
		fmt.Printf("Restore: restoring saved placement %#v (showCmd=%d)\n", placement.RcNormalPosition, placement.ShowCmd)
		if !w32.SetWindowPlacement(hwnd, &placement) {
			return fmt.Errorf("failed to SetWindowPlacement:%d", w32.GetLastError())
		}
		return nil
	}

	fmt.Println("Restore: no saved state, calling SW_RESTORE")
	w32.ShowWindow(hwnd, w32.SW_RESTORE)
	return nil
}

// toggleMaximize - 최대화/복원 토글
// 최대화 상태면 SW_RESTORE(최대화 직전 위치로), 아니면 최대화한다.
// 최대화할 때는 저장된 상태가 없을 때만 저장한다(첫 스냅 전 위치를 덮어쓰지 않음).
func toggleMaximize() error {
	return toggleMaximizeWindow(w32.GetForegroundWindow())
}

func toggleMaximizeWindow(hwnd w32.HWND) error {
	if !isZonableWindow(hwnd) {
		return errors.New("window is not zonable")
	}

	if w32ex.IsZoomed(hwnd) {
		w32.ShowWindow(hwnd, w32.SW_RESTORE)
		dropMaximizeOnlyState(hwnd)
		return nil
	}
	if placement, ok := getPlacement(hwnd); ok {
		saveStateIfAbsent(hwnd, placement, false)
	}
	w32.ShowWindow(hwnd, w32.SW_MAXIMIZE)
	return nil
}

// Modified by hsjeong on 2026-02-09
// Changes: Replaced reflect.DeepEqual with direct field comparison
func sameRect(a, b *w32.RECT) bool {
	if a == nil || b == nil {
		return false
	}
	return a.Left == b.Left && a.Top == b.Top &&
		a.Right == b.Right && a.Bottom == b.Bottom
}
