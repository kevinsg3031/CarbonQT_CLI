//go:build !windows

package ui

import "fmt"

func EnableVirtualTerminal() error {
	return nil
}

func ClearScreen() {
	fmt.Print("\x1b[H\x1b[2J")
}
