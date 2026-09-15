//go:build !linux

package terminal

// EnsureTerminal is a no-op on platforms like Windows where the OS automatically allocates a console window.
func EnsureTerminal() {
}
