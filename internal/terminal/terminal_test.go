//go:build linux

package terminal

import (
	"os"
	"testing"
)

func TestEnsureTerminal_InsideTerminal(t *testing.T) {
	// If JEE_ANKI_IN_TERMINAL is set, EnsureTerminal returns immediately without exit
	os.Setenv("JEE_ANKI_IN_TERMINAL", "1")
	defer os.Unsetenv("JEE_ANKI_IN_TERMINAL")

	EnsureTerminal()
}
