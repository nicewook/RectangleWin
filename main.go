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
//go:generate /bin/bash -c "go run github.com/josephspurrier/goversioninfo/cmd/goversioninfo@latest -arm -64 -icon=assets/icon.ico - <<< '{}'"

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

// savedStates - 스냅 전 창 상태 저장 (창당 1개, 메모리에만 저장)
var savedStates = make(map[w32.HWND]w32.RECT)

// Modified by hsjeong on 2026-02-09
// Changes: Added hotkey ID allocation constants and documentation
// Hotkey ID allocation strategy:
// IDs are organized into functional groups with gaps for future additions
// - 1-9:    Edge snaps (Halves)
// - 10-19:  Window states (Maximize, Center, Restore)
// - 20-29:  Corner snaps
// - 30-39:  Third snaps
// - 40-49:  Size adjustments
// - 100+:   Reserved for future use
const (
	hotkeyIDHalvesMin  = 1
	hotkeyIDHalvesMax  = 9
	hotkeyIDStatesMin  = 10
	hotkeyIDStatesMax  = 19
	hotkeyIDCornersMin = 20
	hotkeyIDCornersMax = 29
	hotkeyIDThirdsMin  = 30
	hotkeyIDThirdsMax  = 39
	hotkeyIDSizeMin    = 40
	hotkeyIDSizeMax    = 49
)

func main() {
	runtime.LockOSThread() // since we bind hotkeys etc that need to dispatch their message here
	if !w32ex.SetProcessDPIAware() {
		fmt.Println("warn: failed to set DPI aware, continuing without DPI awareness")
	}

	autorun, err := AutoRunEnabled()
	if err != nil {
		fmt.Printf("warn: failed to check autorun status: %v\n", err)
		autorun = false // default to disabled
	}
	fmt.Printf("autorun enabled=%v\n", autorun)
	printMonitors()

	// Simple resize helper - no cycling, just direct snap
	simpleResize := func(f resizeFunc, name string) func() {
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

	// Multi-display resize helper - supports moving between monitors
	multiDisplayResize := func(pos SnapPosition, name string) func() {
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

	// Simple action helper - for maximize/restore
	simpleAction := func(action func() error, name string) func() {
		return func() {
			fmt.Printf("Hotkey: %s\n", name)
			if err := action(); err != nil {
				fmt.Printf("warn: %s: %v\n", strings.ToLower(name), err)
			}
		}
	}

	hks := []HotKey{
		// ===== Halves (ID: 1-4) =====
		// Left/Right Half support multi-display movement
		{id: 1, mod: MOD_CONTROL | MOD_ALT, vk: w32.VK_LEFT, callback: multiDisplayResize(SnapLeftHalf, "Left Half")},
		{id: 2, mod: MOD_CONTROL | MOD_ALT, vk: w32.VK_RIGHT, callback: multiDisplayResize(SnapRightHalf, "Right Half")},
		{id: 3, mod: MOD_CONTROL | MOD_ALT, vk: w32.VK_UP, callback: simpleResize(topHalf, "Top Half")},
		{id: 4, mod: MOD_CONTROL | MOD_ALT, vk: w32.VK_DOWN, callback: simpleResize(bottomHalf, "Bottom Half")},

		// ===== Maximize / Center / Restore (ID: 10-12) =====
		{id: 10, mod: MOD_CONTROL | MOD_ALT, vk: w32.VK_RETURN /*Enter*/, callback: simpleAction(maximize, "Maximize")},
		{id: 11, mod: MOD_CONTROL | MOD_ALT, vk: 'C', callback: simpleResize(center, "Center")},
		{id: 12, mod: MOD_CONTROL | MOD_ALT, vk: w32.VK_BACK /*Backspace*/, callback: simpleAction(restore, "Restore")},

		// ===== Corners (ID: 20-23) =====
		{id: 20, mod: MOD_CONTROL | MOD_ALT, vk: 'U', callback: simpleResize(topLeftHalf, "Top Left")},
		{id: 21, mod: MOD_CONTROL | MOD_ALT, vk: 'I', callback: simpleResize(topRightHalf, "Top Right")},
		{id: 22, mod: MOD_CONTROL | MOD_ALT, vk: 'J', callback: simpleResize(bottomLeftHalf, "Bottom Left")},
		{id: 23, mod: MOD_CONTROL | MOD_ALT, vk: 'K', callback: simpleResize(bottomRightHalf, "Bottom Right")},

		// ===== Thirds (ID: 30-34) =====
		// First/Last Third and First/Last Two Thirds support multi-display movement
		{id: 30, mod: MOD_CONTROL | MOD_ALT, vk: 'D', callback: multiDisplayResize(SnapFirstThird, "First Third")},
		{id: 31, mod: MOD_CONTROL | MOD_ALT, vk: 'F', callback: simpleResize(centerThird, "Center Third")},
		{id: 32, mod: MOD_CONTROL | MOD_ALT, vk: 'G', callback: multiDisplayResize(SnapLastThird, "Last Third")},
		{id: 33, mod: MOD_CONTROL | MOD_ALT, vk: 'E', callback: multiDisplayResize(SnapFirstTwoThirds, "First Two Thirds")},
		{id: 34, mod: MOD_CONTROL | MOD_ALT, vk: 'T', callback: multiDisplayResize(SnapLastTwoThirds, "Last Two Thirds")},

		// ===== Size (ID: 40-41) =====
		{id: 40, mod: MOD_CONTROL | MOD_ALT, vk: w32.VK_OEM_MINUS /*-*/, callback: simpleResize(makeSmaller, "Make Smaller")},
		{id: 41, mod: MOD_CONTROL | MOD_ALT, vk: w32.VK_OEM_PLUS /*+*/, callback: simpleResize(makeLarger, "Make Larger")},
	}

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

func resize(hwnd w32.HWND, f resizeFunc) (bool, error) {
	if !isZonableWindow(hwnd) {
		fmt.Printf("warn: non-zonable window: %s\n", w32.GetWindowText(hwnd))
		return false, nil
	}
	rect := w32.GetWindowRect(hwnd)
	mon := w32.MonitorFromWindow(hwnd, w32.MONITOR_DEFAULTTONEAREST)
	hdc := w32.GetDC(hwnd)
	displayDPI := w32.GetDeviceCaps(hdc, w32.LOGPIXELSY)
	if !w32.ReleaseDC(hwnd, hdc) {
		return false, fmt.Errorf("failed to ReleaseDC:%d", w32.GetLastError())
	}
	var monInfo w32.MONITORINFO
	if !w32.GetMonitorInfo(mon, &monInfo) {
		return false, fmt.Errorf("failed to GetMonitorInfo:%d", w32.GetLastError())
	}

	ok, frame := w32.DwmGetWindowAttributeEXTENDED_FRAME_BOUNDS(hwnd)
	if !ok {
		return false, fmt.Errorf("failed to DwmGetWindowAttributeEXTENDED_FRAME_BOUNDS:%d", w32.GetLastError())
	}
	windowDPI := w32ex.GetDpiForWindow(hwnd)
	resizedFrame := resizeForDpi(frame, int32(windowDPI), int32(displayDPI))

	fmt.Printf("> window: 0x%x %#v (w:%d,h:%d) mon=0x%X(@ display DPI:%d)\n", hwnd, rect, rect.Width(), rect.Height(), mon, displayDPI)
	fmt.Printf("> DWM frame:        %#v (W:%d,H:%d) @ window DPI=%v\n", frame, frame.Width(), frame.Height(), windowDPI)
	fmt.Printf("> DPI-less frame:   %#v (W:%d,H:%d)\n", resizedFrame, resizedFrame.Width(), resizedFrame.Height())

	// calculate how many extra pixels go to win10 invisible borders
	lExtra := resizedFrame.Left - rect.Left
	rExtra := -resizedFrame.Right + rect.Right
	tExtra := resizedFrame.Top - rect.Top
	bExtra := -resizedFrame.Bottom + rect.Bottom

	newPos := f(monInfo.RcWork, resizedFrame)

	// adjust offsets based on invisible borders
	newPos.Left -= lExtra
	newPos.Top -= tExtra
	newPos.Right += rExtra
	newPos.Bottom += bExtra

	if sameRect(rect, &newPos) {
		fmt.Println("no resize")
		return false, nil
	}

	// 첫 스냅 시에만 현재 상태 저장 (저장된 상태가 없을 때만)
	if _, exists := savedStates[hwnd]; !exists {
		savedStates[hwnd] = *rect
		fmt.Printf("> saved state for restore: %#v\n", *rect)
	}

	fmt.Printf("> resizing to: %#v (W:%d,H:%d)\n", newPos, newPos.Width(), newPos.Height())
	if !w32.ShowWindow(hwnd, w32.SW_SHOWNORMAL) { // normalize window first if it's set to SW_SHOWMAXIMIZE (and therefore stays maximized)
		return false, fmt.Errorf("failed to normalize window ShowWindow:%d", w32.GetLastError())
	}
	if !w32.SetWindowPos(hwnd, 0, int(newPos.Left), int(newPos.Top), int(newPos.Width()), int(newPos.Height()), w32.SWP_NOZORDER|w32.SWP_NOACTIVATE) {
		return false, fmt.Errorf("failed to SetWindowPos:%d", w32.GetLastError())
	}
	rect = w32.GetWindowRect(hwnd)
	fmt.Printf("> post-resize: %#v(W:%d,H:%d)\n", rect, rect.Width(), rect.Height())
	return true, nil
}

// resizeWithMultiDisplay handles snap with multi-display support
func resizeWithMultiDisplay(hwnd w32.HWND, pos SnapPosition) (bool, error) {
	if !isZonableWindow(hwnd) {
		fmt.Printf("warn: non-zonable window: %s\n", w32.GetWindowText(hwnd))
		return false, nil
	}

	rect := w32.GetWindowRect(hwnd)
	hdc := w32.GetDC(hwnd)
	displayDPI := w32.GetDeviceCaps(hdc, w32.LOGPIXELSY)
	if !w32.ReleaseDC(hwnd, hdc) {
		return false, fmt.Errorf("failed to ReleaseDC:%d", w32.GetLastError())
	}

	ok, frame := w32.DwmGetWindowAttributeEXTENDED_FRAME_BOUNDS(hwnd)
	if !ok {
		return false, fmt.Errorf("failed to DwmGetWindowAttributeEXTENDED_FRAME_BOUNDS:%d", w32.GetLastError())
	}
	windowDPI := w32ex.GetDpiForWindow(hwnd)
	resizedFrame := resizeForDpi(frame, int32(windowDPI), int32(displayDPI))

	// Get target monitor and snap function from multi-display logic
	targetWork, snapFunc, proceed := multiDisplaySnap(hwnd, pos, resizedFrame)
	if !proceed {
		fmt.Println("no movement (single monitor or already at position)")
		return false, nil
	}

	fmt.Printf("> window: 0x%x %#v (w:%d,h:%d) displayDPI:%d\n", hwnd, rect, rect.Width(), rect.Height(), displayDPI)
	fmt.Printf("> DWM frame:        %#v (W:%d,H:%d) @ window DPI=%v\n", frame, frame.Width(), frame.Height(), windowDPI)
	fmt.Printf("> target monitor work: %#v\n", targetWork)

	// calculate how many extra pixels go to win10 invisible borders
	lExtra := resizedFrame.Left - rect.Left
	rExtra := -resizedFrame.Right + rect.Right
	tExtra := resizedFrame.Top - rect.Top
	bExtra := -resizedFrame.Bottom + rect.Bottom

	newPos := snapFunc(targetWork, resizedFrame)

	// adjust offsets based on invisible borders
	newPos.Left -= lExtra
	newPos.Top -= tExtra
	newPos.Right += rExtra
	newPos.Bottom += bExtra

	if sameRect(rect, &newPos) {
		fmt.Println("no resize")
		return false, nil
	}

	// 첫 스냅 시에만 현재 상태 저장 (저장된 상태가 없을 때만)
	if _, exists := savedStates[hwnd]; !exists {
		savedStates[hwnd] = *rect
		fmt.Printf("> saved state for restore: %#v\n", *rect)
	}

	fmt.Printf("> resizing to: %#v (W:%d,H:%d)\n", newPos, newPos.Width(), newPos.Height())
	if !w32.ShowWindow(hwnd, w32.SW_SHOWNORMAL) {
		return false, fmt.Errorf("failed to normalize window ShowWindow:%d", w32.GetLastError())
	}
	if !w32.SetWindowPos(hwnd, 0, int(newPos.Left), int(newPos.Top), int(newPos.Width()), int(newPos.Height()), w32.SWP_NOZORDER|w32.SWP_NOACTIVATE) {
		return false, fmt.Errorf("failed to SetWindowPos:%d", w32.GetLastError())
	}
	rect = w32.GetWindowRect(hwnd)
	fmt.Printf("> post-resize: %#v(W:%d,H:%d)\n", rect, rect.Width(), rect.Height())
	return true, nil
}

func maximize() error {
	hwnd := w32.GetForegroundWindow()
	if !isZonableWindow(hwnd) {
		return errors.New("foreground window is not zonable")
	}
	if !w32.ShowWindow(hwnd, w32.SW_MAXIMIZE) {
		return fmt.Errorf("failed to ShowWindow:%d", w32.GetLastError())
	}
	return nil
}

// restore - 통합 복원 함수
// 1. 최대화 상태 → SW_RESTORE
// 2. 스냅 상태 → 저장된 원래 위치로 복원
func restore() error {
	hwnd := w32.GetForegroundWindow()
	if !isZonableWindow(hwnd) {
		return errors.New("foreground window is not zonable")
	}

	// 1. 최대화 상태 확인
	if w32ex.IsZoomed(hwnd) {
		fmt.Println("Restore: window is maximized, calling SW_RESTORE")
		if !w32.ShowWindow(hwnd, w32.SW_RESTORE) {
			return fmt.Errorf("failed to ShowWindow(SW_RESTORE):%d", w32.GetLastError())
		}
		return nil
	}

	// 2. 저장된 상태가 있으면 복원
	if state, ok := savedStates[hwnd]; ok {
		fmt.Printf("Restore: restoring to saved state %#v\n", state)
		if !w32.SetWindowPos(hwnd, 0, int(state.Left), int(state.Top),
			int(state.Width()), int(state.Height()),
			w32.SWP_NOZORDER|w32.SWP_NOACTIVATE) {
			return fmt.Errorf("failed to SetWindowPos:%d", w32.GetLastError())
		}
		delete(savedStates, hwnd)
		return nil
	}

	// 3. 저장된 상태가 없으면 SW_RESTORE 시도 (최소화 등 다른 상태 복원)
	fmt.Println("Restore: no saved state, calling SW_RESTORE")
	if !w32.ShowWindow(hwnd, w32.SW_RESTORE) {
		return fmt.Errorf("failed to ShowWindow(SW_RESTORE):%d", w32.GetLastError())
	}
	return nil
}

func resizeForDpi(src w32.RECT, from, to int32) w32.RECT {
	return w32.RECT{
		Left:   src.Left * to / from,
		Right:  src.Right * to / from,
		Top:    src.Top * to / from,
		Bottom: src.Bottom * to / from,
	}
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
