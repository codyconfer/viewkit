// Package clipboard copies text to the system clipboard.
//
// It prefers a native helper for the current session — wl-copy under Wayland,
// xclip or xsel under X11, pbcopy on macOS, clip on Windows and clip.exe under
// WSL — and falls back to the OSC 52 terminal escape sequence when none is
// found. The fallback is written to the controlling terminal rather than
// standard output, so redirecting a program's output to a file is never
// corrupted by the escape sequence.
package clipboard

import (
	"encoding/base64"
	"fmt"
	"io"
	"os"
	"os/exec"
	"runtime"
	"strings"
)

var goos = runtime.GOOS

var run = func(name string, stdin string, args ...string) error {
	cmd := exec.Command(name, args...)
	cmd.Stdin = strings.NewReader(stdin)
	return cmd.Run()
}

var lookPath = exec.LookPath

var ttyWriter = func() (io.WriteCloser, error) {
	return os.OpenFile("/dev/tty", os.O_WRONLY, 0)
}

var procVersion = func() string {
	b, err := os.ReadFile("/proc/version")
	if err != nil {
		return ""
	}
	return string(b)
}

type tool struct {
	name string
	args []string
}

// Copy writes s to the system clipboard.
//
// It uses the first native helper available for the current session. When
// there is none it writes an OSC 52 escape sequence to the controlling
// terminal, which lets a terminal emulator forward the text to the clipboard
// of whichever machine the user is sitting at. Copy reports an error if no
// mechanism works.
func Copy(s string) error {
	if t, ok := resolve(); ok {
		if err := run(t.name, s, t.args...); err != nil {
			return fmt.Errorf("clipboard: %s: %w", t.name, err)
		}
		return nil
	}
	if goos == "windows" {
		return fmt.Errorf("clipboard: clip not found on PATH")
	}
	w, err := ttyWriter()
	if err != nil {
		return fmt.Errorf("clipboard: open tty: %w", err)
	}
	defer w.Close()
	if _, err := io.WriteString(w, osc52(s)); err != nil {
		return fmt.Errorf("clipboard: write tty: %w", err)
	}
	return nil
}

// Available reports whether a clipboard mechanism was found. It is true when a
// native helper is on PATH, or when the OSC 52 fallback can reach the
// controlling terminal.
func Available() bool {
	if _, ok := resolve(); ok {
		return true
	}
	if goos == "windows" {
		return false
	}
	w, err := ttyWriter()
	if err != nil {
		return false
	}
	_ = w.Close()
	return true
}

func resolve() (tool, bool) {
	if os.Getenv("WAYLAND_DISPLAY") != "" {
		if found("wl-copy") {
			return tool{name: "wl-copy"}, true
		}
	}
	if os.Getenv("DISPLAY") != "" {
		if found("xclip") {
			return tool{name: "xclip", args: []string{"-selection", "clipboard"}}, true
		}
		if found("xsel") {
			return tool{name: "xsel", args: []string{"--clipboard", "--input"}}, true
		}
	}
	switch {
	case goos == "darwin":
		if found("pbcopy") {
			return tool{name: "pbcopy"}, true
		}
	case goos == "windows":
		if found("clip") {
			return tool{name: "clip"}, true
		}
	case wsl():
		if found("clip.exe") {
			return tool{name: "clip.exe"}, true
		}
	}
	return tool{}, false
}

func found(name string) bool {
	_, err := lookPath(name)
	return err == nil
}

func wsl() bool {
	return strings.Contains(strings.ToLower(procVersion()), "microsoft")
}

func osc52(s string) string {
	return "\x1b]52;c;" + base64.StdEncoding.EncodeToString([]byte(s)) + "\x07"
}
