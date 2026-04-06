//go:build windows

package ui

import (
	"fmt"
	"os"

	"golang.org/x/sys/windows"
)

func EnableVirtualTerminal() error {
	handle := windows.Handle(os.Stdout.Fd())
	var mode uint32
	if err := windows.GetConsoleMode(handle, &mode); err != nil {
		return err
	}

	mode |= windows.ENABLE_VIRTUAL_TERMINAL_PROCESSING | windows.ENABLE_PROCESSED_OUTPUT
	return windows.SetConsoleMode(handle, mode)
}

func ClearScreen() {
	fmt.Print("\x1b[H\x1b[2J")
}
