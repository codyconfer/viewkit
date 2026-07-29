package clipboard

import (
	"bytes"
	"encoding/base64"
	"errors"
	"io"
	"os"
	"os/exec"
	"slices"
	"strings"
	"testing"
)

type capture struct {
	name  string
	stdin string
	args  []string

	path   []string
	proc   string
	runErr error

	tty     bytes.Buffer
	ttyErr  error
	ttyOpen int
	closes  int
}

type fakeTTY struct{ c *capture }

func (f fakeTTY) Write(p []byte) (int, error) { return f.c.tty.Write(p) }

func (f fakeTTY) Close() error { f.c.closes++; return nil }

func install(t *testing.T, targetOS string, onPath ...string) *capture {
	t.Helper()
	c := &capture{path: onPath}
	origOS, origRun, origLook, origTTY, origProc := goos, run, lookPath, ttyWriter, procVersion
	t.Cleanup(func() {
		goos, run, lookPath, ttyWriter, procVersion = origOS, origRun, origLook, origTTY, origProc
	})
	goos = targetOS
	run = func(name string, stdin string, args ...string) error {
		c.name, c.stdin, c.args = name, stdin, args
		return c.runErr
	}
	lookPath = func(name string) (string, error) {
		if slices.Contains(c.path, name) {
			return "/usr/bin/" + name, nil
		}
		return "", exec.ErrNotFound
	}
	ttyWriter = func() (io.WriteCloser, error) {
		c.ttyOpen++
		if c.ttyErr != nil {
			return nil, c.ttyErr
		}
		return fakeTTY{c: c}, nil
	}
	procVersion = func() string { return c.proc }
	t.Setenv("WAYLAND_DISPLAY", "")
	t.Setenv("DISPLAY", "")
	return c
}

const wslProc = "Linux version 5.15.0-1054-microsoft (Microsoft@Microsoft.com)"

func TestCopySelectsTool(t *testing.T) {
	tests := []struct {
		name     string
		goos     string
		wayland  string
		display  string
		proc     string
		path     []string
		wantName string
		wantArgs []string
	}{
		{
			name:     "wayland prefers wl-copy",
			goos:     "linux",
			wayland:  "wayland-0",
			display:  ":0",
			path:     []string{"wl-copy", "xclip", "pbcopy"},
			wantName: "wl-copy",
		},
		{
			name:     "wayland without wl-copy falls through to x11",
			goos:     "linux",
			wayland:  "wayland-0",
			display:  ":0",
			path:     []string{"xclip"},
			wantName: "xclip",
			wantArgs: []string{"-selection", "clipboard"},
		},
		{
			name:     "x11 prefers xclip",
			goos:     "linux",
			display:  ":0",
			path:     []string{"xclip", "xsel"},
			wantName: "xclip",
			wantArgs: []string{"-selection", "clipboard"},
		},
		{
			name:     "x11 falls back to xsel when xclip missing",
			goos:     "linux",
			display:  ":0",
			path:     []string{"xsel"},
			wantName: "xsel",
			wantArgs: []string{"--clipboard", "--input"},
		},
		{
			name:     "darwin uses pbcopy",
			goos:     "darwin",
			path:     []string{"pbcopy"},
			wantName: "pbcopy",
		},
		{
			name:     "windows uses clip",
			goos:     "windows",
			path:     []string{"clip", "clip.exe"},
			wantName: "clip",
		},
		{
			name:     "wsl uses clip.exe",
			goos:     "linux",
			proc:     wslProc,
			path:     []string{"clip.exe"},
			wantName: "clip.exe",
		},
		{
			name:     "wayland beats native tool",
			goos:     "darwin",
			wayland:  "wayland-1",
			path:     []string{"wl-copy", "pbcopy"},
			wantName: "wl-copy",
		},
	}

	const text = "copy me"
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := install(t, tt.goos, tt.path...)
			c.proc = tt.proc
			t.Setenv("WAYLAND_DISPLAY", tt.wayland)
			t.Setenv("DISPLAY", tt.display)

			if err := Copy(text); err != nil {
				t.Fatalf("Copy: %v", err)
			}
			if c.name != tt.wantName {
				t.Errorf("command = %q, want %q", c.name, tt.wantName)
			}
			if !slices.Equal(c.args, tt.wantArgs) {
				t.Errorf("args = %q, want %q", c.args, tt.wantArgs)
			}
			if c.stdin != text {
				t.Errorf("stdin = %q, want %q", c.stdin, text)
			}
			if c.ttyOpen != 0 {
				t.Errorf("tty opened %d times, want 0", c.ttyOpen)
			}
			if !Available() {
				t.Error("Available() = false, want true")
			}
		})
	}
}

func TestCopyFallsBackToOSC52(t *testing.T) {
	tests := []struct {
		name string
		text string
	}{
		{name: "ascii", text: "hello world"},
		{name: "empty", text: ""},
		{name: "unicode and newlines", text: "line one\nline two ✓"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := install(t, "linux")

			stdout, restore := captureStdout(t)
			err := Copy(tt.text)
			out := restore()
			if err != nil {
				t.Fatalf("Copy: %v", err)
			}

			want := "\x1b]52;c;" + base64.StdEncoding.EncodeToString([]byte(tt.text)) + "\x07"
			if got := c.tty.String(); got != want {
				t.Errorf("tty payload = %q, want %q", got, want)
			}
			if c.ttyOpen != 1 {
				t.Errorf("tty opened %d times, want 1", c.ttyOpen)
			}
			if c.closes != 1 {
				t.Errorf("tty closed %d times, want 1", c.closes)
			}
			if c.name != "" {
				t.Errorf("unexpected command %q", c.name)
			}
			if out != "" {
				t.Errorf("wrote %q to stdout, want nothing (%s)", out, stdout)
			}
		})
	}
}

func TestCopyNoToolAndNoTTY(t *testing.T) {
	c := install(t, "linux")
	c.ttyErr = errors.New("no tty")

	if err := Copy("x"); err == nil {
		t.Fatal("Copy: expected an error when no tool and no tty")
	}
	if Available() {
		t.Error("Available() = true, want false")
	}
}

func TestCopyWindowsWithoutClip(t *testing.T) {
	c := install(t, "windows")

	err := Copy("x")
	if err == nil {
		t.Fatal("Copy: expected an error on windows without clip")
	}
	if !strings.Contains(err.Error(), "clip") {
		t.Errorf("error = %v, want it to mention clip", err)
	}
	if c.ttyOpen != 0 {
		t.Errorf("tty opened %d times, want 0 on windows", c.ttyOpen)
	}
	if Available() {
		t.Error("Available() = true, want false on windows without clip")
	}
}

func TestCopyWrapsToolError(t *testing.T) {
	c := install(t, "darwin", "pbcopy")
	sentinel := errors.New("boom")
	c.runErr = sentinel

	err := Copy("x")
	if !errors.Is(err, sentinel) {
		t.Fatalf("Copy error = %v, want it to wrap %v", err, sentinel)
	}
	if !strings.Contains(err.Error(), "pbcopy") {
		t.Errorf("error = %v, want it to name pbcopy", err)
	}
}

func TestWSLDetectionIsCaseInsensitiveAndScoped(t *testing.T) {
	tests := []struct {
		name string
		goos string
		proc string
		want bool
	}{
		{name: "lowercase microsoft", goos: "linux", proc: "Linux version 5.15 microsoft-standard", want: true},
		{name: "mixed case Microsoft", goos: "linux", proc: wslProc, want: true},
		{name: "plain linux", goos: "linux", proc: "Linux version 6.9.3-arch1-1", want: false},
		{name: "darwin ignores proc version", goos: "darwin", proc: wslProc, want: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := install(t, tt.goos, "clip.exe")
			c.proc = tt.proc

			if err := Copy("x"); err != nil {
				t.Fatalf("Copy: %v", err)
			}
			if got := c.name == "clip.exe"; got != tt.want {
				t.Errorf("used clip.exe = %v, want %v (command %q)", got, tt.want, c.name)
			}
		})
	}
}

func TestAvailable(t *testing.T) {
	tests := []struct {
		name    string
		goos    string
		display string
		path    []string
		ttyErr  error
		want    bool
	}{
		{name: "tool on path", goos: "linux", display: ":0", path: []string{"xclip"}, want: true},
		{name: "no tool but tty works", goos: "linux", want: true},
		{name: "no tool and no tty", goos: "linux", ttyErr: errors.New("nope"), want: false},
		{name: "windows without clip", goos: "windows", want: false},
		{name: "windows with clip", goos: "windows", path: []string{"clip"}, want: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := install(t, tt.goos, tt.path...)
			c.ttyErr = tt.ttyErr
			t.Setenv("DISPLAY", tt.display)

			if got := Available(); got != tt.want {
				t.Errorf("Available() = %v, want %v", got, tt.want)
			}
			if c.name != "" {
				t.Errorf("Available ran a command (%q), want none", c.name)
			}
		})
	}
}

func captureStdout(t *testing.T) (string, func() string) {
	t.Helper()
	path := t.TempDir() + "/stdout"
	f, err := os.Create(path)
	if err != nil {
		t.Fatalf("create stdout file: %v", err)
	}
	orig := os.Stdout
	os.Stdout = f
	return path, func() string {
		os.Stdout = orig
		_ = f.Close()
		b, err := os.ReadFile(path)
		if err != nil {
			t.Fatalf("read stdout file: %v", err)
		}
		return string(b)
	}
}
