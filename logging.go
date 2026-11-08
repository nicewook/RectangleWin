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
// Changes: New file (fork addition). GUI builds (-H=windowsgui) have no
//
//	console, so diagnostics were lost. When stdout is unusable, log to
//	%LOCALAPPDATA%\RectangleWin\RectangleWin.log (truncated at each start).

package main

import (
	"os"
	"path/filepath"
)

// setupLogging redirects stdout/stderr to a log file when the process has no
// usable stdout (normal tray launch). Shell redirection keeps working as is.
func setupLogging() {
	if os.Stdout != nil {
		if _, err := os.Stdout.Stat(); err == nil {
			return
		}
	}
	dir, err := os.UserCacheDir() // %LOCALAPPDATA%
	if err != nil {
		return
	}
	dir = filepath.Join(dir, "RectangleWin")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return
	}
	f, err := os.Create(filepath.Join(dir, "RectangleWin.log"))
	if err != nil {
		return
	}
	os.Stdout = f
	os.Stderr = f
}
