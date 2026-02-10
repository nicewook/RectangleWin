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

// Modified by hsjeong on 2026-02-09
// Changes: Added Windows 10+ DPI awareness context support with graceful fallback

package w32ex

import (
	"syscall"
	"unsafe"

	"github.com/gonutz/w32/v2"
)

const (
	GA_PARENT    = 1
	GA_ROOT      = 2
	GA_ROOTOWNER = 3
)

// DPI_AWARENESS_CONTEXT values for SetProcessDpiAwarenessContext
const (
	DPI_AWARENESS_CONTEXT_UNAWARE              = -1
	DPI_AWARENESS_CONTEXT_SYSTEM_AWARE         = -2
	DPI_AWARENESS_CONTEXT_PER_MONITOR_AWARE    = -3
	DPI_AWARENESS_CONTEXT_PER_MONITOR_AWARE_V2 = -4
)

var user32 = syscall.NewLazyDLL("user32.dll")

func RegisterHotKey(hwnd w32.HWND, id, mod, vk int) bool {
	r1, _, _ := user32.NewProc("RegisterHotKey").Call(uintptr(hwnd), uintptr(id), uintptr(mod), uintptr(vk))
	return r1 != 0
}

func GetDpiForWindow(hwnd w32.HWND) int32 {
	r1, _, _ := user32.NewProc("GetDpiForWindow").Call(uintptr(hwnd))
	return int32(r1)
}

func GetWindowModuleFileName(hwnd w32.HWND) string {
	var path [32768]uint16
	ret, _, _ := user32.NewProc("GetWindowModuleFileNameW").Call(
		uintptr(hwnd),
		uintptr(unsafe.Pointer(&path[0])),
		uintptr(len(path)),
	)
	if ret == 0 {
		return ""
	}
	return syscall.UTF16ToString(path[:])
}

func GetAncestor(hwnd w32.HWND, gaFlags uint) w32.HWND {
	r1, _, _ := user32.NewProc("GetAncestor").Call(uintptr(hwnd), uintptr(gaFlags))
	return w32.HWND(r1)
}

func GetShellWindow() (hwnd w32.HWND) {
	r1, _, _ := user32.NewProc("GetShellWindow").Call()
	return w32.HWND(r1)
}

// setProcessDpiAwarenessContext attempts to set DPI awareness using Windows 10+ API
func setProcessDpiAwarenessContext(value int32) bool {
	proc := user32.NewProc("SetProcessDpiAwarenessContext")
	if proc.Find() != nil {
		return false // Function not available (Windows < 1703)
	}
	r1, _, _ := proc.Call(uintptr(value))
	return r1 != 0
}

// SetProcessDPIAware sets the process as DPI-aware with fallback support
// Tries Windows 10+ APIs first, then falls back to legacy API
func SetProcessDPIAware() bool {
	// Try Per-Monitor Aware V2 (Windows 10 1703+)
	if setProcessDpiAwarenessContext(DPI_AWARENESS_CONTEXT_PER_MONITOR_AWARE_V2) {
		return true
	}
	// Try Per-Monitor Aware (Windows 8.1+)
	if setProcessDpiAwarenessContext(DPI_AWARENESS_CONTEXT_PER_MONITOR_AWARE) {
		return true
	}
	// Fall back to legacy API (Windows Vista+)
	r1, _, _ := user32.NewProc("SetProcessDPIAware").Call()
	return r1 != 0
}

// IsZoomed - 창이 최대화 상태인지 확인
func IsZoomed(hwnd w32.HWND) bool {
	r1, _, _ := user32.NewProc("IsZoomed").Call(uintptr(hwnd))
	return r1 != 0
}
