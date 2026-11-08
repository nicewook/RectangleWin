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
// Modified by hsjeong on 2026-11-08
// Changes: Removed tests for deleted thirds/size functions

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

func TestMerge(t *testing.T) {
	horizontal := w32.RECT{Left: 0, Right: 960, Top: 0, Bottom: 1080}
	vertical := w32.RECT{Left: 0, Right: 1920, Top: 0, Bottom: 540}
	got := merge(horizontal, vertical)
	want := w32.RECT{Left: 0, Right: 960, Top: 0, Bottom: 540}
	if got != want {
		t.Errorf("merge() = %v, want %v", got, want)
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
