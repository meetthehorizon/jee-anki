//go:build linux

package terminal

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"golang.org/x/term"
)

type terminalLauncher struct {
	name string
	args func(exe string) []string
}

// terminals in priority order
var terminals = []terminalLauncher{
	{
		name: "xdg-terminal-exec",
		args: func(exe string) []string { return []string{exe} },
	},
	{
		name: "konsole",
		args: func(exe string) []string { return []string{"-p", "tabtitle=jee-anki", "-e", exe} },
	},
	{
		name: "gnome-terminal",
		args: func(exe string) []string { return []string{"--title=jee-anki", "--", exe} },
	},
	{
		name: "alacritty",
		args: func(exe string) []string { return []string{"--title", "jee-anki", "-e", exe} },
	},
	{
		name: "kitty",
		args: func(exe string) []string { return []string{"--title", "jee-anki", exe} },
	},
	{
		name: "foot",
		args: func(exe string) []string { return []string{"--title", "jee-anki", exe} },
	},
	{
		name: "ptyxis",
		args: func(exe string) []string { return []string{"--", exe} },
	},
	{
		name: "xfce4-terminal",
		args: func(exe string) []string { return []string{"--title=jee-anki", "-e", exe} },
	},
	{
		name: "mate-terminal",
		args: func(exe string) []string { return []string{"--title=jee-anki", "-e", exe} },
	},
	{
		name: "tilix",
		args: func(exe string) []string { return []string{"-e", exe} },
	},
	{
		name: "terminator",
		args: func(exe string) []string { return []string{"-e", exe} },
	},
	{
		name: "x-terminal-emulator",
		args: func(exe string) []string { return []string{"-e", exe} },
	},
	{
		name: "xterm",
		args: func(exe string) []string { return []string{"-title", "jee-anki", "-e", exe} },
	},
}

// EnsureTerminal checks if the application is running inside an interactive terminal.
// If double-clicked from a GUI file manager (no TTY attached), it spawns a terminal emulator
// running this executable and terminates the GUI parent process.
func EnsureTerminal() {
	// If already running inside a spawned terminal, do nothing
	if os.Getenv("JEE_ANKI_IN_TERMINAL") == "1" {
		return
	}

	// Check if stdin and stdout are terminals
	isTTY := term.IsTerminal(int(os.Stdin.Fd())) && term.IsTerminal(int(os.Stdout.Fd()))
	if isTTY {
		return // Running interactively in a terminal already
	}

	// We are running without a TTY (likely double-clicked from Dolphin, Nautilus, etc.)
	exePath, err := os.Executable()
	if err != nil {
		exePath = os.Args[0]
	}
	exePath, err = filepath.EvalSymlinks(exePath)
	if err != nil {
		exePath, _ = os.Executable()
	}
	exePath, err = filepath.Abs(exePath)
	if err != nil {
		exePath, _ = os.Executable()
	}

	cwd, err := os.Getwd()
	if err != nil {
		cwd = filepath.Dir(exePath)
	}

	// Look for an installed terminal emulator
	for _, t := range terminals {
		termBin, err := exec.LookPath(t.name)
		if err != nil {
			continue
		}

		args := t.args(exePath)
		cmd := exec.Command(termBin, args...)
		cmd.Dir = cwd
		cmd.Env = append(os.Environ(), "JEE_ANKI_IN_TERMINAL=1")

		if err := cmd.Start(); err == nil {
			// Successfully launched in a new terminal window; exit GUI parent process
			os.Exit(0)
		}
	}

	// If no terminal emulator could be launched, alert the user via graphical dialog if possible
	alertNoTerminal(cwd)
	os.Exit(1)
}

func alertNoTerminal(dir string) {
	msg := fmt.Sprintf("jee-anki requires a terminal emulator to display its interactive menu.\n\nPlease open your terminal, navigate to:\n%s\nand run: ./jee-anki-linux-amd64", dir)

	// Try kdialog (KDE)
	if kdialog, err := exec.LookPath("kdialog"); err == nil {
		_ = exec.Command(kdialog, "--sorry", msg).Run()
		return
	}

	// Try zenity (GNOME / GTK)
	if zenity, err := exec.LookPath("zenity"); err == nil {
		_ = exec.Command(zenity, "--error", "--text", msg).Run()
		return
	}

	// Try notify-send
	if notify, err := exec.LookPath("notify-send"); err == nil {
		_ = exec.Command(notify, "jee-anki", msg).Run()
		return
	}
}
