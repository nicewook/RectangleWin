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
// Changes: Added hotkey ID validation to RegisterHotKey function
// Modified by hsjeong on 2026-11-08
// Changes: A panicking hotkey callback or an unknown hotkey id no longer stops
//
//	the message loop (which silently killed all hotkeys)

package main

import (
	"fmt"
	"strings"

	"github.com/ahmetb/RectangleWin/w32ex"
	"github.com/gonutz/w32/v2"
)

var (
	hotkeyRegistrations = make(map[int]*HotKey)
)

type HotKey struct {
	id, mod, vk int
	name        string // shown in the shortcuts dialog
	callback    func()
}

func (h HotKey) String() string { return fmt.Sprintf("mod=0x%x,vk=%d", h.mod, h.vk) }

func (h HotKey) Describe() string {
	var out string
	if h.mod&MOD_WIN == MOD_WIN {
		out += modKeyNames[MOD_WIN] + " + "
	}
	if h.mod&MOD_CONTROL == MOD_CONTROL {
		out += modKeyNames[MOD_CONTROL] + " + "
	}
	if h.mod&MOD_ALT == MOD_ALT {
		out += modKeyNames[MOD_ALT] + " + "
	}
	if h.mod&MOD_SHIFT == MOD_SHIFT {
		out += modKeyNames[MOD_SHIFT] + " + "
	}
	if v, ok := keyNames[h.vk]; ok {
		out += v
	} else {
		out += fmt.Sprintf("UNKNOWN KEY(0x%x)", h.vk)
	}
	return out
}

var shortKeyNames = map[int]string{
	0x08: "Backspace", 0x0D: "Enter", 0x25: "Left", 0x26: "Up", 0x27: "Right", 0x28: "Down",
}

// Shortcut returns a compact label such as "Ctrl+Alt+Left".
func (h HotKey) Shortcut() string {
	var parts []string
	for _, m := range []int{MOD_WIN, MOD_CONTROL, MOD_ALT, MOD_SHIFT} {
		if h.mod&m == m {
			parts = append(parts, modKeyNames[m])
		}
	}
	switch name, ok := shortKeyNames[h.vk]; {
	case ok:
		parts = append(parts, name)
	case (h.vk >= '0' && h.vk <= '9') || (h.vk >= 'A' && h.vk <= 'Z'):
		parts = append(parts, string(rune(h.vk)))
	default:
		parts = append(parts, fmt.Sprintf("0x%X", h.vk))
	}
	return strings.Join(parts, "+")
}

func RegisterHotKey(h HotKey) (bool, error) {
	if _, ok := hotkeyRegistrations[h.id]; ok {
		return false, fmt.Errorf("hotkey id %d already registered", h.id)
	}

	// Validate ID is within acceptable range
	if h.id < 1 || h.id >= 100 {
		return false, fmt.Errorf("hotkey id %d is outside valid range (1-99)", h.id)
	}

	ok := w32ex.RegisterHotKey(0, h.id, h.mod, h.vk)
	if ok {
		hotkeyRegistrations[h.id] = &h
	}
	return ok, nil
}

func msgLoop() error {
	defer fmt.Println("event loop finished")
	for {
		var m w32.MSG
		c := w32.GetMessage(&m, 0, 0, 0)
		if c == -1 {
			return fmt.Errorf("GetMessage failed: %d", c)
		} else if c == 0 {
			// WM_QUIT received
			return nil
		}
		if m.Message == w32.WM_HOTKEY {
			h, ok := hotkeyRegistrations[int(m.WParam)]
			if !ok {
				fmt.Printf("warn: hotkey without callback: %#v\n", m)
				continue
			}
			fmt.Printf("trace: hotkey id=%d (%s)\n", m.WParam, h)
			runHotkey(h)
		} else {
			fmt.Printf("unhandled message received:0x%x %d\n", m.Message, m.Message)
			w32.TranslateMessage(&m)
			w32.DispatchMessage(&m)
		}
	}
}

// runHotkey invokes the callback, turning a panic into a log line so one bad
// window cannot take down the whole tool.
func runHotkey(h *HotKey) {
	defer func() {
		if r := recover(); r != nil {
			fmt.Printf("error: hotkey %s panicked: %v\n", h, r)
		}
	}()
	h.callback()
}
