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
// Changes: New file (fork addition). Tests for border compensation, center,
//
//	multi-display navigation and autorun path handling.

package main

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/gonutz/w32/v2"
)

// qhd is a 2560x1440 work area with a 48px taskbar at the bottom.
var qhd = w32.RECT{Left: 0, Top: 0, Right: 2560, Bottom: 1392}

func TestSnapRectCompensatesInvisibleBorders(t *testing.T) {
	// Typical Win10/11 window: 7px invisible border left/right/bottom, none on top.
	rect := w32.RECT{Left: 293, Top: 100, Right: 1307, Bottom: 907}
	frame := w32.RECT{Left: 300, Top: 100, Right: 1300, Bottom: 900}

	got := snapRect(rect, frame, qhd, leftHalf)
	want := w32.RECT{Left: -7, Top: 0, Right: 1287, Bottom: 1399}
	if got != want {
		t.Fatalf("snapRect(leftHalf) = %v, want %v", got, want)
	}
	// The visible frame must land exactly on the left half.
	visible := w32.RECT{
		Left:   got.Left + (frame.Left - rect.Left),
		Top:    got.Top + (frame.Top - rect.Top),
		Right:  got.Right - (rect.Right - frame.Right),
		Bottom: got.Bottom - (rect.Bottom - frame.Bottom),
	}
	if visible != leftHalf(qhd, frame) {
		t.Errorf("visible frame = %v, want %v", visible, leftHalf(qhd, frame))
	}
}

// Regression for the removed resizeForDpi: a large, high-DPI window must never
// be scaled; the result depends only on the border widths.
func TestSnapRectIsDPIIndependent(t *testing.T) {
	// 150% scaling: borders are ~11px.
	rect := w32.RECT{Left: 989, Top: 200, Right: 2571, Bottom: 1411}
	frame := w32.RECT{Left: 1000, Top: 200, Right: 2560, Bottom: 1400}

	for name, f := range map[string]resizeFunc{
		"left": leftHalf, "right": rightHalf, "top": topHalf, "bottom": bottomHalf,
		"topLeft": topLeftHalf, "bottomRight": bottomRightHalf, "center": center,
	} {
		got := snapRect(rect, frame, qhd, f)
		target := f(qhd, frame)
		if got.Width() != target.Width()+22 || got.Height() != target.Height()+11 {
			t.Errorf("%s: size %dx%d, want %dx%d", name, got.Width(), got.Height(), target.Width()+22, target.Height()+11)
		}
		if got.Left != target.Left-11 || got.Top != target.Top {
			t.Errorf("%s: origin (%d,%d), want (%d,%d)", name, got.Left, got.Top, target.Left-11, target.Top)
		}
	}
}

func TestSnapRectNegativeMonitor(t *testing.T) {
	// Secondary monitor to the left of the primary.
	work := w32.RECT{Left: -1920, Top: 0, Right: 0, Bottom: 1040}
	rect := w32.RECT{Left: -1500, Top: 100, Right: -500, Bottom: 800}
	got := snapRect(rect, rect, work, rightHalf)
	want := w32.RECT{Left: -960, Top: 0, Right: 0, Bottom: 1040}
	if got != want {
		t.Errorf("snapRect(rightHalf) = %v, want %v", got, want)
	}
}

func TestCenter(t *testing.T) {
	got := center(qhd, w32.RECT{})
	want := w32.RECT{Left: 320, Top: 174, Right: 2240, Bottom: 1218}
	if got != want {
		t.Errorf("center() = %v, want %v", got, want)
	}
	off := w32.RECT{Left: -1920, Top: 100, Right: 0, Bottom: 1180}
	got = center(off, w32.RECT{})
	if got.Left != -1680 || got.Top != 235 || got.Width() != 1440 || got.Height() != 810 {
		t.Errorf("center(offset) = %v", got)
	}
}

func TestIsAtSnapPosition(t *testing.T) {
	left := leftHalf(qhd, w32.RECT{})
	if !isAtSnapPosition(left, qhd, SnapLeftHalf) {
		t.Error("exact left half not detected")
	}
	nearly := left
	nearly.Right += 10
	if !isAtSnapPosition(nearly, qhd, SnapLeftHalf) {
		t.Error("left half within tolerance not detected")
	}
	nearly.Right += 1
	if isAtSnapPosition(nearly, qhd, SnapLeftHalf) {
		t.Error("left half beyond tolerance detected")
	}
	if isAtSnapPosition(left, qhd, SnapRightHalf) {
		t.Error("left half detected as right half")
	}
}

func mon(h uintptr, left, top int32) MonitorInfo {
	return MonitorInfo{Handle: w32.HMONITOR(h), Work: w32.RECT{Left: left, Top: top, Right: left + 100, Bottom: top + 100}}
}

func TestSortMonitorsDeterministic(t *testing.T) {
	a := []MonitorInfo{mon(3, 1920, 0), mon(1, -1920, 0), mon(4, 0, 1080), mon(2, 0, 0)}
	b := []MonitorInfo{mon(2, 0, 0), mon(4, 0, 1080), mon(1, -1920, 0), mon(3, 1920, 0)}
	sortMonitors(a)
	sortMonitors(b)
	for i := range a {
		if a[i] != b[i] {
			t.Fatalf("order depends on input: %v vs %v", a, b)
		}
	}
	wantHandles := []uintptr{1, 2, 4, 3}
	for i, h := range wantHandles {
		if a[i].Handle != w32.HMONITOR(h) {
			t.Errorf("position %d: got monitor %d, want %d", i, a[i].Handle, h)
		}
	}
}

func TestMonitorNavigation(t *testing.T) {
	ms := []MonitorInfo{mon(1, -1920, 0), mon(2, 0, 0), mon(3, 1920, 0)}

	if got := findMonitorIndex(ms, w32.HMONITOR(2)); got != 1 {
		t.Errorf("findMonitorIndex = %d, want 1", got)
	}
	if got := findMonitorIndex(ms, w32.HMONITOR(9)); got != -1 {
		t.Errorf("findMonitorIndex(unknown) = %d, want -1", got)
	}
	if m, ok := getLeftMonitor(ms, 0); !ok || m.Handle != 3 {
		t.Errorf("left of leftmost should wrap to rightmost, got %v %v", m.Handle, ok)
	}
	if m, ok := getRightMonitor(ms, 2); !ok || m.Handle != 1 {
		t.Errorf("right of rightmost should wrap to leftmost, got %v %v", m.Handle, ok)
	}
	if m, ok := getRightMonitor(ms, 0); !ok || m.Handle != 2 {
		t.Errorf("right of first = %v %v, want 2", m.Handle, ok)
	}
	if _, ok := getLeftMonitor(ms[:1], 0); ok {
		t.Error("single monitor must not move")
	}
	if _, ok := getRightMonitor(nil, 0); ok {
		t.Error("no monitors must not move")
	}
}

func TestParseAutoRunCommand(t *testing.T) {
	for in, want := range map[string]string{
		`"C:\Program Files\RectangleWin\RectangleWin.exe"`:        `C:\Program Files\RectangleWin\RectangleWin.exe`,
		`"C:\Program Files\RectangleWin\RectangleWin.exe" --flag`: `C:\Program Files\RectangleWin\RectangleWin.exe`,
		`C:\tools\RectangleWin.exe`:                               `C:\tools\RectangleWin.exe`,
		`  C:\tools\RectangleWin.exe  `:                           `C:\tools\RectangleWin.exe`,
		`"C:\unterminated.exe`:                                    `C:\unterminated.exe`,
		``:                                                        ``,
	} {
		if got := parseAutoRunCommand(in); got != want {
			t.Errorf("parseAutoRunCommand(%q) = %q, want %q", in, got, want)
		}
	}
}

func TestAutoRunCommandRoundTrip(t *testing.T) {
	p := `C:\Users\me\My Apps\RectangleWin-x64-v1.2.3.exe`
	if got := parseAutoRunCommand(autoRunCommand(p)); got != p {
		t.Errorf("round trip = %q, want %q", got, p)
	}
}

func TestSamePath(t *testing.T) {
	if !samePath(`C:\Tools\RectangleWin.exe`, `c:\tools\.\RECTANGLEWIN.EXE`) {
		t.Error("paths differing in case/dot should match")
	}
	if samePath(`C:\Tools\RectangleWin-v1.exe`, `C:\Tools\RectangleWin-v2.exe`) {
		t.Error("different versions must not match")
	}
}

func TestIsTransientPath(t *testing.T) {
	if !isTransientPath(filepath.Join(os.TempDir(), "go-build123", "exe", "RectangleWin.exe")) {
		t.Error("go run binary under TEMP should be transient")
	}
	if isTransientPath(`C:\Tools\RectangleWin.exe`) {
		t.Error("regular install path must not be transient")
	}
}
