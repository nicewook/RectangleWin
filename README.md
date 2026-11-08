# RectangleWin

> **Fork Information:** This project is a fork of [ahmetb/RectangleWin](https://github.com/ahmetb/RectangleWin).
>
> **License:** Licensed under the [Apache License 2.0](./LICENSE).

A minimalistic Windows rewrite of macOS
[Rectangle.app](https://rectangleapp.com)/[Spectacle.app](https://www.spectacleapp.com/).
([Why?](#why))

A hotkey-oriented window snapping and resizing tool for Windows.

This animation (from the upstream project) illustrates how RectangleWin helps
move windows to edges and corners only using hotkeys. This fork supports halves,
corners, center, maximize and restore; thirds cycling is not implemented:

![RectangleWin demo](./assets/RectangleWin-demo.gif)

## Install

1. Go to [Releases](https://github.com/nicewook/RectangleWin/releases) and
   download the suitable binary for your architecture (typically x64).

2. Launch the `.exe` file. Now the program icon should be visible on system
   tray!

3. "Run on startup" is enabled automatically on first launch (uncheck it in the
   tray menu to disable). When you launch a newer version, the startup entry is
   updated to point to it.

Diagnostic logs are written to `%LOCALAPPDATA%\RectangleWin\RectangleWin.log`.

## Keyboard Bindings

- **Snap to edges** (left/right/top/bottom halves):
  - <kbd>Ctrl</kbd> + <kbd>Alt</kbd> + <kbd>&larr;</kbd>: Left Half
  - <kbd>Ctrl</kbd> + <kbd>Alt</kbd> + <kbd>&rarr;</kbd>: Right Half
  - <kbd>Ctrl</kbd> + <kbd>Alt</kbd> + <kbd>&uarr;</kbd>: Top Half
  - <kbd>Ctrl</kbd> + <kbd>Alt</kbd> + <kbd>&darr;</kbd>: Bottom Half

- **Corner snapping** (quadrants):
  - <kbd>Ctrl</kbd> + <kbd>Alt</kbd> + <kbd>U</kbd>: Top Left
  - <kbd>Ctrl</kbd> + <kbd>Alt</kbd> + <kbd>I</kbd>: Top Right
  - <kbd>Ctrl</kbd> + <kbd>Alt</kbd> + <kbd>J</kbd>: Bottom Left
  - <kbd>Ctrl</kbd> + <kbd>Alt</kbd> + <kbd>K</kbd>: Bottom Right

- **Center window**: <kbd>Ctrl</kbd>+<kbd>Alt</kbd>+<kbd>C</kbd>

- **Maximize/Restore (toggle)**: <kbd>Ctrl</kbd>+<kbd>Alt</kbd>+<kbd>Enter</kbd>
  - Press to maximize, press again to restore to previous size/position

- **Restore**: <kbd>Ctrl</kbd>+<kbd>Alt</kbd>+<kbd>Backspace</kbd>
  - Restores maximized or snapped windows to their original position
    (a window that was maximized before snapping is maximized again)

- **Multi-display**: pressing Left/Right Half again moves the window to the
  adjacent display.

> **Note:** See [docs/rectangle_windows_shortcuts.md](./docs/rectangle_windows_shortcuts.md) for the complete list of keyboard shortcuts.

## Why?

It seems that no window snapping utility for Windows is capable of letting
user snap windows to edges or corners in {half, two-thirds, one-third} sizes
using configurable **shortcut keys**, and center windows in a screen like
Rectangle.app does, so I wrote this small utility for myself.

I've tried the native Windows shortcuts and PowerToys FancyZones and they
are not supporting corners, alternating between half and one/two thirds, and
are not offering enough hotkey support.

## Roadmap

- Configurable shortcuts: I don't need these and it will likely require a pop-up
  UI, so I will probably not get to this.

## Development (Install from source)

With Go 1.27+ installed, clone this repository and run:

```sh
go generate
GOOS=windows go build -ldflags -H=windowsgui .
```

The `RectangleWin.exe` will be available in the same directory.

## License

This project is distributed as-is under the Apache 2.0 license.
See [LICENSE](./LICENSE).

If you see bugs, please open issues. I can't promise any fixes.
