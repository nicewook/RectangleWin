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
// Changes: New file (fork addition). Tests for hotkey table, shortcuts dialog,
//
//	autorun decisions, multi-display targeting, window filtering, saved
//	window state and snap tiling.

package main

import (
	"fmt"
	"strings"
	"testing"

	"github.com/gonutz/w32/v2"
)

// ---------- hotkey table ----------

func TestHotKeyTable(t *testing.T) {
	hks := defaultHotKeys()
	if len(hks) != 11 {
		t.Fatalf("got %d hotkeys, want 11", len(hks))
	}
	ids := map[int]bool{}
	combos := map[string]bool{}
	for _, h := range hks {
		if ids[h.id] {
			t.Errorf("duplicate id %d", h.id)
		}
		ids[h.id] = true
		combo := fmt.Sprintf("%x/%x", h.mod&^MOD_NOREPEAT, h.vk)
		if combos[combo] {
			t.Errorf("duplicate key combination %s (%s)", h.Shortcut(), h.name)
		}
		combos[combo] = true
		if h.mod&MOD_NOREPEAT == 0 {
			t.Errorf("%s: missing MOD_NOREPEAT", h.name)
		}
		if h.mod&^MOD_NOREPEAT != MOD_CONTROL|MOD_ALT {
			t.Errorf("%s: modifiers = %#x, want Ctrl+Alt", h.name, h.mod)
		}
		if h.callback == nil || h.name == "" {
			t.Errorf("id %d: missing callback or name", h.id)
		}
		if h.id < 1 || h.id >= 100 {
			t.Errorf("%s: id %d outside 1-99", h.name, h.id)
		}
		inGroup := (h.id >= hotkeyIDHalvesMin && h.id <= hotkeyIDHalvesMax) ||
			(h.id >= hotkeyIDStatesMin && h.id <= hotkeyIDStatesMax) ||
			(h.id >= hotkeyIDCornersMin && h.id <= hotkeyIDCornersMax)
		if !inGroup {
			t.Errorf("%s: id %d not in any group", h.name, h.id)
		}
	}
}

func TestHotKeyShortcutLabel(t *testing.T) {
	for _, tc := range []struct {
		h    HotKey
		want string
	}{
		{HotKey{mod: hotkeyMod, vk: w32.VK_LEFT}, "Ctrl+Alt+Left"},
		{HotKey{mod: hotkeyMod, vk: w32.VK_BACK}, "Ctrl+Alt+Backspace"},
		{HotKey{mod: hotkeyMod, vk: w32.VK_RETURN}, "Ctrl+Alt+Enter"},
		{HotKey{mod: hotkeyMod, vk: 'K'}, "Ctrl+Alt+K"},
		{HotKey{mod: MOD_WIN | MOD_SHIFT, vk: '7'}, "Win+Shift+7"},
		{HotKey{mod: MOD_ALT, vk: 0x70}, "Alt+0x70"},
	} {
		if got := tc.h.Shortcut(); got != tc.want {
			t.Errorf("Shortcut() = %q, want %q", got, tc.want)
		}
	}
}

func TestHotKeyDescribeOmitsNoRepeat(t *testing.T) {
	got := HotKey{mod: hotkeyMod, vk: 'C'}.Describe()
	if got != "Ctrl + Alt + C key" {
		t.Errorf("Describe() = %q", got)
	}
}

func TestShortcutsTextListsEveryHotKey(t *testing.T) {
	hks := defaultHotKeys()
	text := shortcutsText(hks)
	for _, h := range hks {
		line := h.Shortcut() + "\t" + h.name
		if !strings.Contains(text, line) {
			t.Errorf("dialog is missing %q", line)
		}
	}
	for _, title := range []string{"Halves", "Maximize / Center / Restore", "Corners"} {
		if !strings.Contains(text, title+"\n") {
			t.Errorf("dialog is missing group %q", title)
		}
	}
	if strings.Contains(text, "Third") || strings.Contains(text, "Larger") {
		t.Error("dialog lists unimplemented features")
	}
}

func TestRegisterHotKeyRejectsInvalidIDs(t *testing.T) {
	for _, id := range []int{0, -1, 100, 1000} {
		ok, err := RegisterHotKey(HotKey{id: id, mod: hotkeyMod, vk: 'Z'})
		if ok || err == nil {
			t.Errorf("RegisterHotKey(id=%d) = %v, %v; want error", id, ok, err)
		}
	}
}

func TestRunHotkeyRecoversPanic(t *testing.T) {
	ran := false
	runHotkey(&HotKey{id: 1, callback: func() { ran = true; panic("boom") }})
	if !ran {
		t.Fatal("callback did not run")
	}
	// reaching here means the panic did not propagate
}

// ---------- autorun decisions ----------

func TestDecideAutoRun(t *testing.T) {
	const self = `C:\Apps\RectangleWin-x64-v2.0.0.exe`
	quoted := autoRunCommand(self)
	for _, tc := range []struct {
		name          string
		value         string
		exists        bool
		configured    bool
		transient     bool
		wantAction    autoRunAction
		wantEnabledAt bool
	}{
		{"first run", "", false, false, false, autoRunEnable, true},
		{"user disabled", "", false, true, false, autoRunKeep, false},
		{"legacy empty marker", "", true, false, false, autoRunMigrate, false},
		{"legacy whitespace marker", "  ", true, true, false, autoRunMigrate, false},
		{"up to date", quoted, true, true, false, autoRunKeep, true},
		{"same path unquoted (old format)", self, true, false, false, autoRunNormalize, true},
		{"same path different case", `"c:\apps\rectanglewin-x64-v2.0.0.EXE"`, true, true, false, autoRunNormalize, true},
		{"previous version", `"C:\Apps\RectangleWin-x64-v1.0.0.exe"`, true, true, false, autoRunRepoint, true},
		{"relative path from old os.Args[0]", `.\RectangleWin.exe`, true, false, false, autoRunRepoint, true},
		{"go run build must not repoint", `"C:\Apps\RectangleWin.exe"`, true, true, true, autoRunKeep, true},
		{"first run from go run still enables", "", false, false, true, autoRunEnable, true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			action, enabled := decideAutoRun(tc.value, tc.exists, tc.configured, self, tc.transient)
			if action != tc.wantAction || enabled != tc.wantEnabledAt {
				t.Errorf("decideAutoRun = (%d, %v), want (%d, %v)", action, enabled, tc.wantAction, tc.wantEnabledAt)
			}
		})
	}
}

// ---------- multi-display targeting ----------

func TestChooseSnapTarget(t *testing.T) {
	left := MonitorInfo{Handle: 1, Work: w32.RECT{Left: -1920, Top: 0, Right: 0, Bottom: 1040}}
	mid := MonitorInfo{Handle: 2, Work: w32.RECT{Left: 0, Top: 0, Right: 2560, Bottom: 1392}}
	right := MonitorInfo{Handle: 3, Work: w32.RECT{Left: 2560, Top: 0, Right: 4480, Bottom: 1040}}
	all := []MonitorInfo{left, mid, right}
	floating := w32.RECT{Left: 500, Top: 300, Right: 1500, Bottom: 900}
	atLeft := leftHalf(mid.Work, w32.RECT{})
	atRight := rightHalf(mid.Work, w32.RECT{})

	type want struct {
		work    w32.RECT
		layout  resizeFunc
		proceed bool
	}
	for _, tc := range []struct {
		name     string
		pos      SnapPosition
		window   w32.RECT
		monitors []MonitorInfo
		idx      int
		want     want
	}{
		{"not yet at left half: snap on current", SnapLeftHalf, floating, all, 1, want{mid.Work, leftHalf, true}},
		{"at left half: move to left monitor, right-aligned", SnapLeftHalf, atLeft, all, 1, want{left.Work, rightHalf, true}},
		{"at right half: move to right monitor, left-aligned", SnapRightHalf, atRight, all, 1, want{right.Work, leftHalf, true}},
		{"at left half on leftmost: wrap to rightmost", SnapLeftHalf, leftHalf(left.Work, w32.RECT{}), all, 0, want{right.Work, rightHalf, true}},
		{"single monitor already there: no-op", SnapLeftHalf, atLeft, []MonitorInfo{mid}, 0, want{mid.Work, leftHalf, false}},
		{"monitor not found: re-snap on current", SnapRightHalf, atRight, all, -1, want{mid.Work, rightHalf, true}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			cur := mid.Work
			if tc.idx == 0 && len(tc.monitors) > 1 {
				cur = left.Work
			}
			work, layout, proceed := chooseSnapTarget(tc.pos, tc.window, cur, tc.monitors, tc.idx, true)
			if proceed != tc.want.proceed || work != tc.want.work {
				t.Fatalf("got (%v, %v), want (%v, %v)", work, proceed, tc.want.work, tc.want.proceed)
			}
			probe := w32.RECT{Left: 0, Top: 0, Right: 1000, Bottom: 1000}
			if layout(probe, probe) != tc.want.layout(probe, probe) {
				t.Errorf("wrong layout function")
			}
		})
	}

	if _, _, ok := chooseSnapTarget(SnapPosition(99), floating, mid.Work, all, 1, true); ok {
		t.Error("unknown SnapPosition must not proceed")
	}

	// Window was maximized and its normal bounds equal the right half: it
	// must snap on the current monitor instead of jumping to the next one.
	work, layout, ok := chooseSnapTarget(SnapRightHalf, atRight, mid.Work, all, 1, false)
	probe := w32.RECT{Left: 0, Top: 0, Right: 1000, Bottom: 1000}
	if !ok || work != mid.Work || layout(probe, probe) != rightHalf(probe, probe) {
		t.Errorf("maximized window cycled monitors: work=%v ok=%v", work, ok)
	}
}

// ---------- window filtering ----------

func TestIsZonableStyle(t *testing.T) {
	const normal = w32.WS_OVERLAPPEDWINDOW | w32.WS_VISIBLE
	for _, tc := range []struct {
		name           string
		style, exStyle uint32
		want           bool
	}{
		{"normal app window", normal, 0, true},
		{"popup with min/max buttons (e.g. some editors)", w32.WS_POPUP | w32.WS_THICKFRAME | w32.WS_MAXIMIZEBOX, 0, true},
		{"popup with thick frame, no buttons", w32.WS_POPUP | w32.WS_THICKFRAME, 0, false},
		{"child window", normal | w32.WS_CHILD, 0, false},
		{"disabled (modal owner)", normal | w32.WS_DISABLED, 0, false},
		{"tool window", normal, w32.WS_EX_TOOLWINDOW, false},
		{"no-activate window", normal, w32.WS_EX_NOACTIVATE, false},
	} {
		if got := isZonableStyle(tc.style, tc.exStyle); got != tc.want {
			t.Errorf("%s: isZonableStyle = %v, want %v", tc.name, got, tc.want)
		}
	}
}

func TestIsSystemClassName(t *testing.T) {
	for _, c := range []string{"Shell_TrayWnd", "shell_traywnd", "Progman", "WorkerW", "SysListView32", "Shell_SecondaryTrayWnd"} {
		if !isSystemClassName(c) {
			t.Errorf("%q should be a system class", c)
		}
	}
	for _, c := range []string{"Chrome_WidgetWin_1", "Notepad", "CabinetWClass", ""} {
		if isSystemClassName(c) {
			t.Errorf("%q should not be a system class", c)
		}
	}
}

// ---------- saved window state ----------

func placement(showCmd uint32, l, t, r, b int32) w32.WINDOWPLACEMENT {
	return w32.WINDOWPLACEMENT{ShowCmd: showCmd, RcNormalPosition: w32.RECT{Left: l, Top: t, Right: r, Bottom: b}}
}

func TestSavedStateKeepsFirstSnap(t *testing.T) {
	m := map[w32.HWND]savedState{}
	id := windowIdentity{pid: 10, class: "Notepad"}
	orig := placement(w32.SW_SHOWNORMAL, 100, 100, 900, 700)

	if !storeState(m, 1, id, orig, true) {
		t.Fatal("first store should succeed")
	}
	// second snap / maximize toggle must not overwrite the original
	if storeState(m, 1, id, placement(w32.SW_SHOWNORMAL, 0, 0, 1280, 1392), true) {
		t.Fatal("second store must be ignored")
	}
	got, ok := takeState(m, 1, id, true)
	if !ok || got.RcNormalPosition != orig.RcNormalPosition {
		t.Fatalf("take = %v %v, want original", got.RcNormalPosition, ok)
	}
	if _, ok := takeState(m, 1, id, true); ok {
		t.Error("state must be consumed by take")
	}
}

func TestSavedStateKeepsMaximized(t *testing.T) {
	m := map[w32.HWND]savedState{}
	id := windowIdentity{pid: 10, class: "Notepad"}
	storeState(m, 1, id, placement(w32.SW_SHOWMAXIMIZED, 100, 100, 900, 700), true)
	got, _ := takeState(m, 1, id, true)
	if got.ShowCmd != w32.SW_SHOWMAXIMIZED {
		t.Errorf("ShowCmd = %d, want SW_SHOWMAXIMIZED", got.ShowCmd)
	}
}

func TestSavedStateNeverRestoresMinimized(t *testing.T) {
	for _, cmd := range []uint32{w32.SW_SHOWMINIMIZED, w32.SW_MINIMIZE, w32.SW_HIDE, w32.SW_SHOWMINNOACTIVE} {
		m := map[w32.HWND]savedState{}
		id := windowIdentity{pid: 1, class: "X"}
		storeState(m, 1, id, placement(cmd, 0, 0, 10, 10), true)
		got, _ := takeState(m, 1, id, true)
		if got.ShowCmd != w32.SW_SHOWNORMAL {
			t.Errorf("ShowCmd %d stored as %d, want SW_SHOWNORMAL", cmd, got.ShowCmd)
		}
	}
}

func TestSavedStateRecycledHWND(t *testing.T) {
	m := map[w32.HWND]savedState{}
	oldWin := windowIdentity{pid: 10, class: "Notepad"}
	newWin := windowIdentity{pid: 22, class: "Chrome_WidgetWin_1"}
	storeState(m, 7, oldWin, placement(w32.SW_SHOWNORMAL, 1, 2, 3, 4), true)

	// A different window got HWND 7: restore must not use the old position...
	if _, ok := takeState(m, 7, newWin, true); ok {
		t.Fatal("state of a different window must not be restored")
	}
	// ...and the stale entry is gone.
	if _, exists := m[7]; exists {
		t.Error("stale entry not removed")
	}

	// Saving for the new window replaces a stale entry instead of being ignored.
	storeState(m, 7, oldWin, placement(w32.SW_SHOWNORMAL, 1, 2, 3, 4), true)
	fresh := placement(w32.SW_SHOWNORMAL, 50, 50, 60, 60)
	if !storeState(m, 7, newWin, fresh, true) {
		t.Fatal("store for a new window on a recycled HWND must replace the stale entry")
	}
	if got, ok := takeState(m, 7, newWin, true); !ok || got.RcNormalPosition != fresh.RcNormalPosition {
		t.Errorf("take = %v %v, want fresh", got.RcNormalPosition, ok)
	}
}

func TestSavedStateUnknownIdentity(t *testing.T) {
	m := map[w32.HWND]savedState{}
	id := windowIdentity{pid: 10, class: "Notepad"}
	storeState(m, 1, id, placement(w32.SW_SHOWNORMAL, 1, 2, 3, 4), true)
	if _, ok := takeState(m, 1, windowIdentity{}, false); ok {
		t.Error("must not restore when the window identity cannot be read")
	}
}

// ---------- snap tiling ----------

func TestHalvesTileWithoutGap(t *testing.T) {
	for _, work := range []w32.RECT{
		{Left: 0, Top: 0, Right: 2560, Bottom: 1392},
		{Left: 0, Top: 0, Right: 2561, Bottom: 1393}, // odd sizes
		{Left: -1921, Top: -7, Right: 0, Bottom: 1033},
		{Left: 3, Top: 5, Right: 1027, Bottom: 773},
	} {
		l, r := leftHalf(work, work), rightHalf(work, work)
		if l.Right != r.Left || l.Left != work.Left || r.Right != work.Right {
			t.Errorf("%v: left %v / right %v do not tile", work, l, r)
		}
		tp, bt := topHalf(work, work), bottomHalf(work, work)
		if tp.Bottom != bt.Top || tp.Top != work.Top || bt.Bottom != work.Bottom {
			t.Errorf("%v: top %v / bottom %v do not tile", work, tp, bt)
		}
		tl, tr := topLeftHalf(work, work), topRightHalf(work, work)
		bl, br := bottomLeftHalf(work, work), bottomRightHalf(work, work)
		if tl.Right != tr.Left || bl.Right != br.Left || tl.Bottom != bl.Top || tr.Bottom != br.Top {
			t.Errorf("%v: corners do not tile: %v %v %v %v", work, tl, tr, bl, br)
		}
		area := func(r w32.RECT) int64 { return int64(r.Width()) * int64(r.Height()) }
		if area(tl)+area(tr)+area(bl)+area(br) != area(work) {
			t.Errorf("%v: corners do not cover the work area exactly", work)
		}
	}
}

func TestCenterStaysInsideWorkArea(t *testing.T) {
	for _, work := range []w32.RECT{
		{Left: 0, Top: 0, Right: 2560, Bottom: 1392},
		{Left: -1920, Top: -200, Right: 0, Bottom: 880},
		{Left: 0, Top: 0, Right: 801, Bottom: 601},
	} {
		c := center(work, w32.RECT{})
		if c.Left < work.Left || c.Top < work.Top || c.Right > work.Right || c.Bottom > work.Bottom {
			t.Errorf("center(%v) = %v leaves the work area", work, c)
		}
		if dl, dr := c.Left-work.Left, work.Right-c.Right; dl-dr > 1 || dr-dl > 1 {
			t.Errorf("center(%v) = %v not horizontally centered (%d vs %d)", work, c, dl, dr)
		}
	}
}

// Codex review (round 2): a completed maximize toggle must not leave history
// behind, but a toggle must never discard snap history.
func TestMaximizeOnlyStateIsDropped(t *testing.T) {
	id := windowIdentity{pid: 10, class: "Notepad"}
	beforeMax := placement(w32.SW_SHOWNORMAL, 100, 100, 900, 700)

	m := map[w32.HWND]savedState{}
	storeState(m, 1, id, beforeMax, false) // maximize
	dropMaximizeOnly(m, 1)                 // un-maximize
	if _, ok := m[1]; ok {
		t.Fatal("maximize-only entry survived the completed toggle")
	}

	// snap -> maximize -> un-maximize keeps the snap history
	m = map[w32.HWND]savedState{}
	storeState(m, 1, id, beforeMax, true)
	if storeState(m, 1, id, placement(w32.SW_SHOWNORMAL, 0, 0, 1280, 1392), false) {
		t.Error("maximize toggle overwrote snap history")
	}
	dropMaximizeOnly(m, 1)
	if got, ok := takeState(m, 1, id, true); !ok || got.RcNormalPosition != beforeMax.RcNormalPosition {
		t.Errorf("snap history lost: %v %v", got.RcNormalPosition, ok)
	}

	// maximize -> snap: the snap replaces the maximize-only entry (it captures
	// the maximized placement) and the entry becomes permanent
	m = map[w32.HWND]savedState{}
	storeState(m, 1, id, beforeMax, false)
	maxed := placement(w32.SW_SHOWMAXIMIZED, 100, 100, 900, 700)
	if !storeState(m, 1, id, maxed, true) {
		t.Fatal("snap must replace a maximize-only entry")
	}
	dropMaximizeOnly(m, 1)
	if got, ok := takeState(m, 1, id, true); !ok || got.ShowCmd != w32.SW_SHOWMAXIMIZED {
		t.Errorf("want maximized snap history, got %v %v", got.ShowCmd, ok)
	}
}
