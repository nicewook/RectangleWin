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
// Changes: AutoRunDisable stores empty string instead of deleting; added AutoRunConfigured for first-run detection
// Modified by hsjeong on 2026-11-08
// Changes: Register the absolute, quoted executable path (os.Executable) instead
//
//	of os.Args[0]; keep the Run entry pointing at the running binary after an
//	upgrade; disable now deletes the Run value and remembers the choice under
//	HKCU\Software\RectangleWin (legacy empty-string values are migrated).

package main

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"golang.org/x/sys/windows/registry"
)

const (
	AutoRunName = `RectangleWin`
	regKey      = `SOFTWARE\Microsoft\Windows\CurrentVersion\Run`

	settingsKey            = `SOFTWARE\RectangleWin`
	autoRunConfiguredValue = `AutoRunConfigured`
)

// selfPath returns the absolute path of the running executable.
func selfPath() (string, error) {
	p, err := os.Executable()
	if err != nil {
		return "", err
	}
	if r, err := filepath.EvalSymlinks(p); err == nil {
		p = r
	}
	return p, nil
}

// autoRunCommand builds the Run value for an executable path. The path is
// quoted so that paths containing spaces are unambiguous.
func autoRunCommand(path string) string {
	return `"` + path + `"`
}

// parseAutoRunCommand extracts the executable path from a Run value
// (quoted or unquoted, optionally followed by arguments when quoted).
func parseAutoRunCommand(v string) string {
	v = strings.TrimSpace(v)
	if strings.HasPrefix(v, `"`) {
		if i := strings.Index(v[1:], `"`); i >= 0 {
			return v[1 : 1+i]
		}
		return strings.Trim(v, `"`)
	}
	return v
}

func samePath(a, b string) bool {
	return strings.EqualFold(filepath.Clean(a), filepath.Clean(b))
}

// isTransientPath reports whether path looks like a throwaway build
// (go run / go test put binaries under the temp dir); such paths must not be
// registered for autorun.
func isTransientPath(path string) bool {
	tmp := filepath.Clean(os.TempDir())
	rel, err := filepath.Rel(tmp, filepath.Clean(path))
	return err == nil && !strings.HasPrefix(rel, "..")
}

// readRunValue returns the Run value and whether it exists.
func readRunValue() (string, bool, error) {
	rk, err := registry.OpenKey(registry.CURRENT_USER, regKey, registry.QUERY_VALUE)
	if err != nil {
		return "", false, err
	}
	defer rk.Close()

	v, _, err := rk.GetStringValue(AutoRunName)
	if errors.Is(err, registry.ErrNotExist) {
		return "", false, nil
	}
	if errors.Is(err, registry.ErrUnexpectedType) {
		return "", true, nil
	}
	if err != nil {
		return "", false, err
	}
	return v, true, nil
}

func autoRunConfigured() bool {
	k, err := registry.OpenKey(registry.CURRENT_USER, settingsKey, registry.QUERY_VALUE)
	if err != nil {
		return false
	}
	defer k.Close()
	_, _, err = k.GetIntegerValue(autoRunConfiguredValue)
	return err == nil
}

func markAutoRunConfigured() error {
	k, _, err := registry.CreateKey(registry.CURRENT_USER, settingsKey, registry.SET_VALUE)
	if err != nil {
		return err
	}
	defer k.Close()
	return k.SetDWordValue(autoRunConfiguredValue, 1)
}

// AutoRunEnabled reports whether a RectangleWin Run entry is registered.
func AutoRunEnabled() (bool, error) {
	v, exists, err := readRunValue()
	if err != nil {
		return false, err
	}
	return exists && strings.TrimSpace(v) != "", nil
}

func AutoRunEnable() error {
	path, err := selfPath()
	if err != nil {
		return err
	}
	rk, err := registry.OpenKey(registry.CURRENT_USER, regKey, registry.SET_VALUE)
	if err != nil {
		return err
	}
	defer rk.Close()
	if err := rk.SetStringValue(AutoRunName, autoRunCommand(path)); err != nil {
		return err
	}
	return markAutoRunConfigured()
}

func AutoRunDisable() error {
	rk, err := registry.OpenKey(registry.CURRENT_USER, regKey, registry.SET_VALUE)
	if err != nil {
		return err
	}
	defer rk.Close()
	if err := rk.DeleteValue(AutoRunName); err != nil && !errors.Is(err, registry.ErrNotExist) {
		return err
	}
	return markAutoRunConfigured()
}

type autoRunAction int

const (
	autoRunKeep      autoRunAction = iota // nothing to do
	autoRunEnable                         // first run: register
	autoRunMigrate                        // legacy empty value: delete + remember "disabled"
	autoRunNormalize                      // same binary, rewrite quoting
	autoRunRepoint                        // enabled for another binary: point at this one
)

// decideAutoRun is the pure decision behind SyncAutoRun. It returns the
// action to take and whether autorun is enabled afterwards.
func decideAutoRun(value string, exists, configured bool, self string, selfTransient bool) (autoRunAction, bool) {
	switch {
	case !exists && !configured:
		return autoRunEnable, true
	case !exists:
		return autoRunKeep, false
	case strings.TrimSpace(value) == "":
		return autoRunMigrate, false
	case samePath(parseAutoRunCommand(value), self):
		if value != autoRunCommand(self) {
			return autoRunNormalize, true
		}
		return autoRunKeep, true
	case selfTransient:
		return autoRunKeep, true
	default:
		return autoRunRepoint, true
	}
}

// SyncAutoRun is called once at startup and returns whether autorun is
// enabled afterwards (see decideAutoRun for the rules).
func SyncAutoRun() (bool, error) {
	v, exists, err := readRunValue()
	if err != nil {
		return false, err
	}
	self, err := selfPath()
	if err != nil {
		return exists && strings.TrimSpace(v) != "", err
	}
	action, enabled := decideAutoRun(v, exists, autoRunConfigured(), self, isTransientPath(self))
	switch action {
	case autoRunEnable:
		fmt.Println("autorun: enabling on first run")
		err = AutoRunEnable()
	case autoRunMigrate:
		err = AutoRunDisable()
	case autoRunNormalize:
		err = AutoRunEnable()
	case autoRunRepoint:
		fmt.Printf("autorun: repointing from %q to %q\n", v, self)
		err = AutoRunEnable()
	}
	if err != nil && action == autoRunEnable {
		enabled = false
	}
	return enabled, err
}
