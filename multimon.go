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
// Changes: Multi-display left/right movement (fork addition); deterministic monitor sort

package main

import (
	"sort"

	"github.com/gonutz/w32/v2"
)

// MonitorInfo holds monitor handle and its work area
type MonitorInfo struct {
	Handle w32.HMONITOR
	Work   w32.RECT
}

// getMonitorList returns all monitors sorted by X position (left to right)
func getMonitorList() []MonitorInfo {
	var monitors []MonitorInfo
	EnumMonitors(func(h w32.HMONITOR) bool {
		var info w32.MONITORINFO
		if w32.GetMonitorInfo(h, &info) {
			monitors = append(monitors, MonitorInfo{
				Handle: h,
				Work:   info.RcWork,
			})
		}
		return true
	})
	sortMonitors(monitors)
	return monitors
}

// sortMonitors orders monitors left to right (top to bottom on ties) so the
// order does not depend on the OS enumeration order.
func sortMonitors(monitors []MonitorInfo) {
	sort.SliceStable(monitors, func(i, j int) bool {
		if monitors[i].Work.Left != monitors[j].Work.Left {
			return monitors[i].Work.Left < monitors[j].Work.Left
		}
		return monitors[i].Work.Top < monitors[j].Work.Top
	})
}

// findMonitorIndex returns the index of the monitor in the sorted list
func findMonitorIndex(monitors []MonitorInfo, current w32.HMONITOR) int {
	for i, m := range monitors {
		if m.Handle == current {
			return i
		}
	}
	return -1
}

// getLeftMonitor returns the monitor to the left, with wrap-around
func getLeftMonitor(monitors []MonitorInfo, currentIdx int) (MonitorInfo, bool) {
	if len(monitors) <= 1 {
		return MonitorInfo{}, false
	}
	newIdx := currentIdx - 1
	if newIdx < 0 {
		newIdx = len(monitors) - 1 // wrap to rightmost
	}
	return monitors[newIdx], true
}

// getRightMonitor returns the monitor to the right, with wrap-around
func getRightMonitor(monitors []MonitorInfo, currentIdx int) (MonitorInfo, bool) {
	if len(monitors) <= 1 {
		return MonitorInfo{}, false
	}
	newIdx := currentIdx + 1
	if newIdx >= len(monitors) {
		newIdx = 0 // wrap to leftmost
	}
	return monitors[newIdx], true
}

// MoveDirection indicates the direction of multi-display movement
type MoveDirection int

const (
	MoveNone  MoveDirection = iota
	MoveLeft                // Move to left monitor
	MoveRight               // Move to right monitor
)

// SnapPosition represents window snap positions
type SnapPosition int

const (
	SnapLeftHalf SnapPosition = iota
	SnapRightHalf
)

// snapPositionInfo holds snap function and its multi-display behavior
type snapPositionInfo struct {
	snapFunc      resizeFunc
	moveDirection MoveDirection
	edgeAligned   SnapPosition // position after moving to another monitor
}

// snapPositions defines the behavior for each snap position
var snapPositions = map[SnapPosition]snapPositionInfo{
	SnapLeftHalf:  {leftHalf, MoveLeft, SnapRightHalf},
	SnapRightHalf: {rightHalf, MoveRight, SnapLeftHalf},
}

// isAtSnapPosition checks if window is already at the given snap position
func isAtSnapPosition(windowRect, monitorWork w32.RECT, pos SnapPosition) bool {
	expected := snapPositions[pos].snapFunc(monitorWork, windowRect)

	// Allow 10px tolerance for comparison (DPI and rounding issues)
	tolerance := int32(10)
	return abs32(windowRect.Left-expected.Left) <= tolerance &&
		abs32(windowRect.Right-expected.Right) <= tolerance &&
		abs32(windowRect.Top-expected.Top) <= tolerance &&
		abs32(windowRect.Bottom-expected.Bottom) <= tolerance
}

func abs32(x int32) int32 {
	if x < 0 {
		return -x
	}
	return x
}

// chooseSnapTarget is the pure decision behind multiDisplaySnap.
// monitors must be sorted (sortMonitors); currentIdx is the index of the
// window's monitor in it (-1 if unknown, e.g. the topology just changed).
// allowCycle=false forces a snap on the current monitor (e.g. the window was
// maximized, so its normal bounds say nothing about where it is shown).
func chooseSnapTarget(pos SnapPosition, windowRect, currentWork w32.RECT, monitors []MonitorInfo, currentIdx int, allowCycle bool) (w32.RECT, resizeFunc, bool) {
	info, exists := snapPositions[pos]
	if !exists {
		return w32.RECT{}, nil, false
	}
	if !allowCycle || info.moveDirection == MoveNone || !isAtSnapPosition(windowRect, currentWork, pos) {
		return currentWork, info.snapFunc, true
	}
	if currentIdx < 0 {
		return currentWork, info.snapFunc, true
	}

	var target MonitorInfo
	var ok bool
	switch info.moveDirection {
	case MoveLeft:
		target, ok = getLeftMonitor(monitors, currentIdx)
	case MoveRight:
		target, ok = getRightMonitor(monitors, currentIdx)
	}
	if !ok {
		return currentWork, info.snapFunc, false // single monitor: already there
	}
	return target.Work, snapPositions[info.edgeAligned].snapFunc, true
}

// multiDisplaySnap handles snap with multi-display support
// Returns: target monitor work area, snap function to use, whether to proceed
func multiDisplaySnap(hwnd w32.HWND, pos SnapPosition, windowRect w32.RECT, allowCycle bool) (w32.RECT, resizeFunc, bool) {
	currentMon := w32.MonitorFromWindow(hwnd, w32.MONITOR_DEFAULTTONEAREST)
	var monInfo w32.MONITORINFO
	if !w32.GetMonitorInfo(currentMon, &monInfo) {
		return w32.RECT{}, nil, false
	}
	var monitors []MonitorInfo
	idx := -1
	if allowCycle && isAtSnapPosition(windowRect, monInfo.RcWork, pos) {
		monitors = getMonitorList()
		idx = findMonitorIndex(monitors, currentMon)
	}
	return chooseSnapTarget(pos, windowRect, monInfo.RcWork, monitors, idx, allowCycle)
}
