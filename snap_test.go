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
// Changes: Added unit tests for snap functions

package main

import (
	"testing"

	"github.com/gonutz/w32/v2"
)

func TestLeftHalf(t *testing.T) {
	display := w32.RECT{Left: 0, Top: 0, Right: 1920, Bottom: 1080}
	got := leftHalf(display, w32.RECT{})
	want := w32.RECT{Left: 0, Top: 0, Right: 960, Bottom: 1080}
	if got != want {
		t.Errorf("leftHalf() = %v, want %v", got, want)
	}
}

func TestRightHalf(t *testing.T) {
	display := w32.RECT{Left: 0, Top: 0, Right: 1920, Bottom: 1080}
	got := rightHalf(display, w32.RECT{})
	want := w32.RECT{Left: 960, Top: 0, Right: 1920, Bottom: 1080}
	if got != want {
		t.Errorf("rightHalf() = %v, want %v", got, want)
	}
}

func TestTopHalf(t *testing.T) {
	display := w32.RECT{Left: 0, Top: 0, Right: 1920, Bottom: 1080}
	got := topHalf(display, w32.RECT{})
	want := w32.RECT{Left: 0, Top: 0, Right: 1920, Bottom: 540}
	if got != want {
		t.Errorf("topHalf() = %v, want %v", got, want)
	}
}

func TestBottomHalf(t *testing.T) {
	display := w32.RECT{Left: 0, Top: 0, Right: 1920, Bottom: 1080}
	got := bottomHalf(display, w32.RECT{})
	want := w32.RECT{Left: 0, Top: 540, Right: 1920, Bottom: 1080}
	if got != want {
		t.Errorf("bottomHalf() = %v, want %v", got, want)
	}
}

func TestTopLeftHalf(t *testing.T) {
	display := w32.RECT{Left: 0, Top: 0, Right: 1920, Bottom: 1080}
	got := topLeftHalf(display, w32.RECT{})
	want := w32.RECT{Left: 0, Top: 0, Right: 960, Bottom: 540}
	if got != want {
		t.Errorf("topLeftHalf() = %v, want %v", got, want)
	}
}

func TestTopRightHalf(t *testing.T) {
	display := w32.RECT{Left: 0, Top: 0, Right: 1920, Bottom: 1080}
	got := topRightHalf(display, w32.RECT{})
	want := w32.RECT{Left: 960, Top: 0, Right: 1920, Bottom: 540}
	if got != want {
		t.Errorf("topRightHalf() = %v, want %v", got, want)
	}
}

func TestBottomLeftHalf(t *testing.T) {
	display := w32.RECT{Left: 0, Top: 0, Right: 1920, Bottom: 1080}
	got := bottomLeftHalf(display, w32.RECT{})
	want := w32.RECT{Left: 0, Top: 540, Right: 960, Bottom: 1080}
	if got != want {
		t.Errorf("bottomLeftHalf() = %v, want %v", got, want)
	}
}

func TestBottomRightHalf(t *testing.T) {
	display := w32.RECT{Left: 0, Top: 0, Right: 1920, Bottom: 1080}
	got := bottomRightHalf(display, w32.RECT{})
	want := w32.RECT{Left: 960, Top: 540, Right: 1920, Bottom: 1080}
	if got != want {
		t.Errorf("bottomRightHalf() = %v, want %v", got, want)
	}
}

func TestLeftOneThirds(t *testing.T) {
	display := w32.RECT{Left: 0, Top: 0, Right: 1920, Bottom: 1080}
	got := leftOneThirds(display, w32.RECT{})
	want := w32.RECT{Left: 0, Top: 0, Right: 640, Bottom: 1080}
	if got != want {
		t.Errorf("leftOneThirds() = %v, want %v", got, want)
	}
}

func TestCenterThird(t *testing.T) {
	display := w32.RECT{Left: 0, Top: 0, Right: 1920, Bottom: 1080}
	got := centerThird(display, w32.RECT{})
	want := w32.RECT{Left: 640, Top: 0, Right: 1280, Bottom: 1080}
	if got != want {
		t.Errorf("centerThird() = %v, want %v", got, want)
	}
}

func TestRightOneThirds(t *testing.T) {
	display := w32.RECT{Left: 0, Top: 0, Right: 1920, Bottom: 1080}
	got := rightOneThirds(display, w32.RECT{})
	want := w32.RECT{Left: 1280, Top: 0, Right: 1920, Bottom: 1080}
	if got != want {
		t.Errorf("rightOneThirds() = %v, want %v", got, want)
	}
}

func TestLeftTwoThirds(t *testing.T) {
	display := w32.RECT{Left: 0, Top: 0, Right: 1920, Bottom: 1080}
	got := leftTwoThirds(display, w32.RECT{})
	want := w32.RECT{Left: 0, Top: 0, Right: 1280, Bottom: 1080}
	if got != want {
		t.Errorf("leftTwoThirds() = %v, want %v", got, want)
	}
}

func TestRightTwoThirds(t *testing.T) {
	display := w32.RECT{Left: 0, Top: 0, Right: 1920, Bottom: 1080}
	got := rightTwoThirds(display, w32.RECT{})
	want := w32.RECT{Left: 640, Top: 0, Right: 1920, Bottom: 1080}
	if got != want {
		t.Errorf("rightTwoThirds() = %v, want %v", got, want)
	}
}

func TestMerge(t *testing.T) {
	horizontal := w32.RECT{Left: 0, Right: 960, Top: 0, Bottom: 1080}
	vertical := w32.RECT{Left: 0, Right: 1920, Top: 0, Bottom: 540}
	got := merge(horizontal, vertical)
	want := w32.RECT{Left: 0, Right: 960, Top: 0, Bottom: 540}
	if got != want {
		t.Errorf("merge() = %v, want %v", got, want)
	}
}

func TestMakeSmaller(t *testing.T) {
	display := w32.RECT{Left: 0, Top: 0, Right: 1920, Bottom: 1080}
	current := w32.RECT{Left: 100, Top: 100, Right: 1820, Bottom: 980}
	got := makeSmaller(display, current)
	// Should shrink by 3% of display width (1920 * 0.03 = 57.6 ≈ 57)
	// New width: 1720 - 57*2 = 1606, centered
	expectedWidth := current.Width() - (display.Width()*3/100)*2
	expectedHeight := current.Height() - (display.Width()*3/100)*2
	if got.Width() != expectedWidth {
		t.Errorf("makeSmaller() width = %d, want %d", got.Width(), expectedWidth)
	}
	if got.Height() != expectedHeight {
		t.Errorf("makeSmaller() height = %d, want %d", got.Height(), expectedHeight)
	}
	// Check that it's centered
	centerX := current.Left + current.Width()/2
	centerY := current.Top + current.Height()/2
	gotCenterX := got.Left + got.Width()/2
	gotCenterY := got.Top + got.Height()/2
	if gotCenterX != centerX || gotCenterY != centerY {
		t.Errorf("makeSmaller() not centered: got center (%d,%d), want (%d,%d)", gotCenterX, gotCenterY, centerX, centerY)
	}
}

func TestMakeSmallerMinSize(t *testing.T) {
	display := w32.RECT{Left: 0, Top: 0, Right: 1920, Bottom: 1080}
	current := w32.RECT{Left: 500, Top: 500, Right: 600, Bottom: 600} // 100x100 window
	got := makeSmaller(display, current)
	// Should not go below 100x100
	if got.Width() < 100 || got.Height() < 100 {
		t.Errorf("makeSmaller() went below minimum size: got %dx%d", got.Width(), got.Height())
	}
}

func TestMakeLarger(t *testing.T) {
	display := w32.RECT{Left: 0, Top: 0, Right: 1920, Bottom: 1080}
	current := w32.RECT{Left: 100, Top: 100, Right: 1820, Bottom: 980}
	got := makeLarger(display, current)
	// Should grow by 3% of display width (1920 * 0.03 = 57.6 ≈ 57)
	// New width: 1720 + 57*2 = 1834, centered
	expectedWidth := current.Width() + (display.Width()*3/100)*2
	expectedHeight := current.Height() + (display.Width()*3/100)*2
	if got.Width() != expectedWidth {
		t.Errorf("makeLarger() width = %d, want %d", got.Width(), expectedWidth)
	}
	if got.Height() != expectedHeight {
		t.Errorf("makeLarger() height = %d, want %d", got.Height(), expectedHeight)
	}
	// Check that it's centered
	centerX := current.Left + current.Width()/2
	centerY := current.Top + current.Height()/2
	gotCenterX := got.Left + got.Width()/2
	gotCenterY := got.Top + got.Height()/2
	if gotCenterX != centerX || gotCenterY != centerY {
		t.Errorf("makeLarger() not centered: got center (%d,%d), want (%d,%d)", gotCenterX, gotCenterY, centerX, centerY)
	}
}

func TestToLeft(t *testing.T) {
	display := w32.RECT{Left: 0, Top: 0, Right: 1920, Bottom: 1080}
	got := toLeft(display, 1, 2)
	want := w32.RECT{Left: 0, Top: 0, Right: 960, Bottom: 1080}
	if got != want {
		t.Errorf("toLeft(1,2) = %v, want %v", got, want)
	}
}

func TestToRight(t *testing.T) {
	display := w32.RECT{Left: 0, Top: 0, Right: 1920, Bottom: 1080}
	got := toRight(display, 1, 2)
	want := w32.RECT{Left: 960, Top: 0, Right: 1920, Bottom: 1080}
	if got != want {
		t.Errorf("toRight(1,2) = %v, want %v", got, want)
	}
}

func TestToTop(t *testing.T) {
	display := w32.RECT{Left: 0, Top: 0, Right: 1920, Bottom: 1080}
	got := toTop(display, 1, 2)
	want := w32.RECT{Left: 0, Top: 0, Right: 1920, Bottom: 540}
	if got != want {
		t.Errorf("toTop(1,2) = %v, want %v", got, want)
	}
}

func TestToBottom(t *testing.T) {
	display := w32.RECT{Left: 0, Top: 0, Right: 1920, Bottom: 1080}
	got := toBottom(display, 1, 2)
	want := w32.RECT{Left: 0, Top: 540, Right: 1920, Bottom: 1080}
	if got != want {
		t.Errorf("toBottom(1,2) = %v, want %v", got, want)
	}
}
