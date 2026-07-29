package term

import (
	"bytes"
	"io"
	"os"
	"testing"
)

type fakeFd struct {
	io.Writer
	fd uintptr
}

func (f fakeFd) Fd() uintptr { return f.fd }

func TestIsTerminalRejectsWritersWithoutFd(t *testing.T) {
	var buf bytes.Buffer
	if IsTerminal(&buf) {
		t.Error("bytes.Buffer reported as a terminal")
	}
	if IsTerminal(io.Discard) {
		t.Error("io.Discard reported as a terminal")
	}
}

func TestIsTerminalRejectsBogusFileDescriptor(t *testing.T) {
	cases := []struct {
		name string
		fd   uintptr
	}{
		{"unopened high fd", 1 << 20},
		{"max fd", ^uintptr(0)},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			w := fakeFd{Writer: io.Discard, fd: c.fd}
			if IsTerminal(w) {
				t.Errorf("fd %d reported as a terminal", c.fd)
			}
		})
	}
}

func TestIsTerminalRejectsRegularFile(t *testing.T) {
	f, err := os.CreateTemp(t.TempDir(), "tty")
	if err != nil {
		t.Fatalf("CreateTemp: %v", err)
	}
	defer f.Close()

	if IsTerminal(f) {
		t.Error("temp file reported as a terminal")
	}
}

func TestIsTerminalNilWriterDoesNotPanic(t *testing.T) {
	defer func() {
		if r := recover(); r != nil {
			t.Fatalf("IsTerminal(nil) panicked: %v", r)
		}
	}()
	if IsTerminal(nil) {
		t.Error("nil writer reported as a terminal")
	}

	var typed *os.File
	if IsTerminal(typed) {
		t.Error("typed-nil *os.File reported as a terminal")
	}
}
