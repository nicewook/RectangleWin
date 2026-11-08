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
// Changes: New file (fork addition). Saved pre-snap window state now keeps the
//
//	full WINDOWPLACEMENT (so maximized windows restore as maximized) and the
//	owning process/class, so a recycled HWND never receives another window's
//	saved position.

package main

import (
	"fmt"
	"unsafe"

	"github.com/gonutz/w32/v2"
)

// windowIdentity distinguishes a window from a later window that happens to
// receive the same (recycled) HWND.
type windowIdentity struct {
	pid   uint32
	class string
}

type savedState struct {
	placement w32.WINDOWPLACEMENT
	id        windowIdentity
	// fromSnap is false while the entry only records a maximize toggle. Such
	// an entry is dropped when the toggle completes, so a later snap records
	// the position the window actually had before that snap.
	fromSnap bool
}

// savedStates - 첫 스냅 전 창 상태 (창당 1개, 메모리에만 저장).
// Accessed only from the main (message loop) thread.
var savedStates = make(map[w32.HWND]savedState)

func identityOf(hwnd w32.HWND) (windowIdentity, bool) {
	_, pid := w32.GetWindowThreadProcessId(hwnd)
	class, ok := w32.GetClassName(hwnd)
	if !ok || pid == 0 {
		return windowIdentity{}, false
	}
	return windowIdentity{pid: uint32(pid), class: class}, true
}

func getPlacement(hwnd w32.HWND) (w32.WINDOWPLACEMENT, bool) {
	var p w32.WINDOWPLACEMENT
	p.Length = uint32(unsafe.Sizeof(p))
	if !w32.GetWindowPlacement(hwnd, &p) {
		fmt.Printf("warn: GetWindowPlacement failed:%d\n", w32.GetLastError())
		return p, false
	}
	return p, true
}

// saveStateIfAbsent stores the placement only if this window has no saved
// state yet, so Restore always returns to the state before the first snap.
// fromSnap distinguishes snaps from maximize toggles (see savedState).
func saveStateIfAbsent(hwnd w32.HWND, p w32.WINDOWPLACEMENT, fromSnap bool) {
	pruneSavedStates()
	id, ok := identityOf(hwnd)
	if !ok {
		return
	}
	if storeState(savedStates, hwnd, id, p, fromSnap) {
		fmt.Printf("> saved state for restore: %#v (showCmd=%d)\n", p.RcNormalPosition, p.ShowCmd)
	}
}

// takeSavedState returns and forgets the saved state for hwnd, but only if it
// still belongs to the same window.
func takeSavedState(hwnd w32.HWND) (w32.WINDOWPLACEMENT, bool) {
	id, ok := identityOf(hwnd)
	return takeState(savedStates, hwnd, id, ok)
}

// dropMaximizeOnlyState forgets an entry that only recorded a maximize toggle
// (called when the window is un-maximized). Snap history is kept.
func dropMaximizeOnlyState(hwnd w32.HWND) {
	dropMaximizeOnly(savedStates, hwnd)
}

func dropMaximizeOnly(m map[w32.HWND]savedState, hwnd w32.HWND) {
	if st, ok := m[hwnd]; ok && !st.fromSnap {
		delete(m, hwnd)
	}
}

// storeState records p for hwnd unless the same window already has a saved
// state. A stale entry from a different window (recycled HWND) is replaced,
// and so is a maximize-only entry when the window is now being snapped (the
// snap captures the current, maximized placement). Reports whether it stored.
func storeState(m map[w32.HWND]savedState, hwnd w32.HWND, id windowIdentity, p w32.WINDOWPLACEMENT, fromSnap bool) bool {
	if cur, exists := m[hwnd]; exists && cur.id == id && (cur.fromSnap || !fromSnap) {
		return false
	}
	// Restore must never minimize the window it brings back.
	if p.ShowCmd != w32.SW_SHOWMAXIMIZED {
		p.ShowCmd = w32.SW_SHOWNORMAL
	}
	m[hwnd] = savedState{placement: p, id: id, fromSnap: fromSnap}
	return true
}

// takeState removes the entry for hwnd and returns it if it belongs to the
// window identified by id.
func takeState(m map[w32.HWND]savedState, hwnd w32.HWND, id windowIdentity, idOK bool) (w32.WINDOWPLACEMENT, bool) {
	st, ok := m[hwnd]
	if !ok {
		return w32.WINDOWPLACEMENT{}, false
	}
	delete(m, hwnd)
	if !idOK || st.id != id {
		fmt.Println("Restore: saved state belongs to a different window, discarding")
		return w32.WINDOWPLACEMENT{}, false
	}
	return st.placement, true
}

// pruneSavedStates drops entries for windows that no longer exist.
func pruneSavedStates() {
	for h := range savedStates {
		if !w32.IsWindow(h) {
			delete(savedStates, h)
		}
	}
}
