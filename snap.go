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

package main

import "github.com/gonutz/w32/v2"

// TODO find a way to round up divisions consistently, otherwise we end up with off by 1px

func toLeft(d w32.RECT, mul, div int32) w32.RECT {
	return w32.RECT{
		Left:   d.Left,
		Top:    d.Top,
		Right:  d.Left + (d.Width()*mul)/div,
		Bottom: d.Top + d.Height()}
}

func toRight(d w32.RECT, mul, div int32) w32.RECT {
	return w32.RECT{
		Left:   d.Left + d.Width() - d.Width()*mul/div,
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
		Top:    d.Top + d.Height() - d.Height()*mul/div,
		Right:  d.Left + d.Width(),
		Bottom: d.Top + d.Height()}
}

func leftHalf(disp, _ w32.RECT) w32.RECT      { return toLeft(disp, 1, 2) }
func leftOneThirds(disp, _ w32.RECT) w32.RECT { return toLeft(disp, 1, 3) }
func leftTwoThirds(disp, _ w32.RECT) w32.RECT { return toLeft(disp, 2, 3) }

func topHalf(disp, _ w32.RECT) w32.RECT { return toTop(disp, 1, 2) }

func rightHalf(disp, _ w32.RECT) w32.RECT      { return toRight(disp, 1, 2) }
func rightOneThirds(disp, _ w32.RECT) w32.RECT { return toRight(disp, 1, 3) }
func rightTwoThirds(disp, _ w32.RECT) w32.RECT { return toRight(disp, 2, 3) }

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

// centerThird - 화면 중앙 1/3
func centerThird(disp, _ w32.RECT) w32.RECT {
	return w32.RECT{
		Left:   disp.Left + disp.Width()/3,
		Top:    disp.Top,
		Right:  disp.Left + disp.Width()*2/3,
		Bottom: disp.Top + disp.Height(),
	}
}

// makeSmaller - 해상도 비례 축소 (3%)
// 창을 현재 위치 중심으로 축소
func makeSmaller(disp, cur w32.RECT) w32.RECT {
	resizeStep := disp.Width() * 3 / 100 // 3%

	newWidth := cur.Width() - resizeStep*2
	newHeight := cur.Height() - resizeStep*2

	// 최소 크기 제한: 100x100
	if newWidth < 100 {
		newWidth = 100
	}
	if newHeight < 100 {
		newHeight = 100
	}

	centerX := cur.Left + cur.Width()/2
	centerY := cur.Top + cur.Height()/2

	return w32.RECT{
		Left:   centerX - newWidth/2,
		Top:    centerY - newHeight/2,
		Right:  centerX + newWidth/2,
		Bottom: centerY + newHeight/2,
	}
}

// makeLarger - 해상도 비례 확대 (3%)
// 창을 현재 위치 중심으로 확대
func makeLarger(disp, cur w32.RECT) w32.RECT {
	resizeStep := disp.Width() * 3 / 100 // 3%

	newWidth := cur.Width() + resizeStep*2
	newHeight := cur.Height() + resizeStep*2

	centerX := cur.Left + cur.Width()/2
	centerY := cur.Top + cur.Height()/2

	return w32.RECT{
		Left:   centerX - newWidth/2,
		Top:    centerY - newHeight/2,
		Right:  centerX + newWidth/2,
		Bottom: centerY + newHeight/2,
	}
}
