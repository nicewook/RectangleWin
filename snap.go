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

// Modified by hsjeong on 2026-02-10
// Changes: Removed thirds/size helpers (feature dropped); halves and corners only
// Modified by hsjeong on 2026-11-08
// Changes: Right/bottom snaps start where left/top end (no 1px gap on odd sizes)

package main

import "github.com/gonutz/w32/v2"

// toRight/toBottom start exactly where toLeft/toTop(div-mul, div) end, so
// complementary snaps (e.g. left + right half) tile without a 1px gap on odd
// work-area sizes.

func toLeft(d w32.RECT, mul, div int32) w32.RECT {
	return w32.RECT{
		Left:   d.Left,
		Top:    d.Top,
		Right:  d.Left + (d.Width()*mul)/div,
		Bottom: d.Top + d.Height()}
}

func toRight(d w32.RECT, mul, div int32) w32.RECT {
	return w32.RECT{
		Left:   d.Left + d.Width()*(div-mul)/div,
		Top:    d.Top,
		Right:  d.Left + d.Width(),
		Bottom: d.Top + d.Height()}
}

func toTop(d w32.RECT, mul, div int32) w32.RECT {
	return w32.RECT{
		Left:   d.Left,
		Top:    d.Top,
		Right:  d.Left + d.Width(),
		Bottom: d.Top + d.Height()*mul/div}
}

func toBottom(d w32.RECT, mul, div int32) w32.RECT {
	return w32.RECT{
		Left:   d.Left,
		Top:    d.Top + d.Height()*(div-mul)/div,
		Right:  d.Left + d.Width(),
		Bottom: d.Top + d.Height()}
}

func leftHalf(disp, _ w32.RECT) w32.RECT { return toLeft(disp, 1, 2) }

func topHalf(disp, _ w32.RECT) w32.RECT { return toTop(disp, 1, 2) }

func rightHalf(disp, _ w32.RECT) w32.RECT { return toRight(disp, 1, 2) }

func bottomHalf(disp, _ w32.RECT) w32.RECT { return toBottom(disp, 1, 2) }

func topLeftHalf(disp, _ w32.RECT) w32.RECT { return merge(toLeft(disp, 1, 2), toTop(disp, 1, 2)) }

func topRightHalf(disp, _ w32.RECT) w32.RECT { return merge(toRight(disp, 1, 2), toTop(disp, 1, 2)) }

func bottomLeftHalf(disp, _ w32.RECT) w32.RECT {
	return merge(toLeft(disp, 1, 2), toBottom(disp, 1, 2))
}

func bottomRightHalf(disp, _ w32.RECT) w32.RECT {
	return merge(toRight(disp, 1, 2), toBottom(disp, 1, 2))
}

func merge(x, y w32.RECT) w32.RECT {
	return w32.RECT{Left: x.Left, Right: x.Right, Top: y.Top, Bottom: y.Bottom}
}
