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

// Modified by nicewook on 2026-01-13
// Changes: Added keyboard shortcuts list menu and dialog

package main

import (
	_ "embed"
	"fmt"

	"fyne.io/systray"
	"github.com/gonutz/w32/v2"
)

//go:embed assets/tray_icon.ico
var icon []byte

const repo = "https://github.com/nicewook/RectangleWin"

func initTray() {
	systray.Register(onReady, onExit)
}

func onReady() {
	systray.SetIcon(icon)
	systray.SetTitle("RectangleWin")
	systray.SetTooltip("RectangleWin")

	autorun, err := AutoRunEnabled()
	if err != nil {
		fmt.Printf("warn: failed to check autorun status in tray: %v\n", err)
		autorun = false // default to disabled
	}

	// About menu - opens repository
	mAbout := systray.AddMenuItem("About RectangleWin...", "")
	go func() {
		for range mAbout.ClickedCh {
			if err := w32.ShellExecute(0, "open", repo, "", "", w32.SW_SHOWNORMAL); err != nil {
				fmt.Printf("failed to launch browser: (%d), %v\n", w32.GetLastError(), err)
			}
		}
	}()

	// Keyboard shortcuts menu
	mShortcuts := systray.AddMenuItem("Keyboard Shortcuts...", "")
	go func() {
		for range mShortcuts.ClickedCh {
			showShortcutsDialog()
		}
	}()

	systray.AddSeparator()

	mAutoRun := systray.AddMenuItemCheckbox("Run on startup", "", autorun)
	go func() {
		for range mAutoRun.ClickedCh {
			if mAutoRun.Checked() {
				if err := AutoRunDisable(); err != nil {
					mAutoRun.SetTitle(err.Error())
					fmt.Printf("warn: autorun disable: %v\n", err)
					continue
				}
				fmt.Println("disabled autorun")
				mAutoRun.Uncheck()
			} else {
				if err := AutoRunEnable(); err != nil {
					mAutoRun.SetTitle(err.Error())
					fmt.Printf("warn: autorun enable: %v\n", err)
					continue
				}
				fmt.Println("enabled autorun")
				mAutoRun.Check()
			}

		}
	}()

	systray.AddSeparator()

	mQuit := systray.AddMenuItem("Exit", "")
	go func() {
		<-mQuit.ClickedCh
		fmt.Println("clicked Exit")
		systray.Quit()
	}()

	fmt.Println("tray ready")
}

func onExit() {
	fmt.Println("onExit invoked")
}

// showShortcutsDialog displays a message box with all keyboard shortcuts
func showShortcutsDialog() {
	shortcuts := `Halves
  Ctrl+Alt+Left	Left Half
  Ctrl+Alt+Right	Right Half
  Ctrl+Alt+Up	Top Half
  Ctrl+Alt+Down	Bottom Half

Maximize / Center / Restore
  Ctrl+Alt+Enter	Maximize
  Ctrl+Alt+C	Center (75%)
  Ctrl+Alt+Backspace	Restore

Corners
  Ctrl+Alt+U	Top Left
  Ctrl+Alt+I	Top Right
  Ctrl+Alt+J	Bottom Left
  Ctrl+Alt+K	Bottom Right

Thirds
  Ctrl+Alt+D	First Third
  Ctrl+Alt+F	Center Third
  Ctrl+Alt+G	Last Third
  Ctrl+Alt+E	First Two Thirds
  Ctrl+Alt+T	Last Two Thirds

Size
  Ctrl+Alt+-	Make Smaller
  Ctrl+Alt++	Make Larger`

	w32.MessageBox(0, shortcuts, "RectangleWin - Keyboard Shortcuts", w32.MB_OK|w32.MB_ICONINFORMATION)
}
