package deck

import (
	"context"
	"fmt"
	"os"
	"strings"
	"testing"
	"time"

	"golang.org/x/sys/unix"
)

// Exercise the real Tea terminal lifecycle, not only Update/View. A real PTY
// catches capability detection, raw mode, final row accounting, and teardown.
func TestInlineTerminalPrompt(t *testing.T) {
	for _, cancel := range []bool{false, true} {
		t.Run(fmt.Sprint(cancel), func(t *testing.T) {
			t.Setenv("TERM", "xterm-256color")
			master, slave := inlinePTY(t)
			defer master.Close()
			defer slave.Close()
			chunks := make(chan string, 100)
			readDone := make(chan struct{})
			go func() {
				defer close(readDone)
				buf := make([]byte, 4096)
				for {
					n, err := master.Read(buf)
					if n > 0 {
						chunks <- string(buf[:n])
					}
					if err != nil {
						return
					}
				}
			}()
			defer func() {
				master.Close()
				select {
				case <-readDone:
				case <-time.After(time.Second):
					t.Error("PTY reader did not stop")
				}
			}()
			u := NewInline(InlineSpec{Input: slave, Output: slave, Width: 40})
			if !u.interactive {
				t.Fatal("PTY not detected as interactive")
			}
			if err := u.Panel("Setup", "instructions"); err != nil {
				t.Fatal(err)
			}
			ctx, stop := context.WithTimeout(context.Background(), 5*time.Second)
			defer stop()
			type response struct {
				value string
				err   error
			}
			result := make(chan response, 1)
			go func() {
				value, err := u.Ask(ctx, "Setup", "Value", "")
				result <- response{value, err}
			}()
			var output strings.Builder
			for !strings.Contains(output.String(), "Value:") {
				select {
				case chunk := <-chunks:
					output.WriteString(chunk)
				case <-ctx.Done():
					t.Fatal("prompt did not render", output.String())
				}
			}
			input := "answer\r"
			if cancel {
				input = "\x03"
			}
			if _, err := master.WriteString(input); err != nil {
				t.Fatal(err)
			}
			var got response
			select {
			case got = <-result:
			case <-ctx.Done():
				t.Fatal("prompt did not finish", output.String())
			}
			if cancel {
				if got.err != ErrCancelled {
					t.Fatalf("cancel error=%v", got.err)
				}
			} else if got.err != nil || got.value != "answer" {
				t.Fatalf("response=%q err=%v", got.value, got.err)
			}
			if err := u.Panel("Next", "next body"); err != nil {
				t.Fatal(err)
			}
			for !strings.Contains(output.String(), "next body") {
				select {
				case chunk := <-chunks:
					output.WriteString(chunk)
				case <-ctx.Done():
					t.Fatal("next panel did not render", output.String())
				}
			}
			text := output.String()
			if strings.Count(text, "\x1b[J") < 2 || strings.Contains(text, "\x1b[?1049h") {
				t.Fatalf("expected two inline replacements without alternate screen: %q", text)
			}
		})
	}
}

func inlinePTY(t *testing.T) (*os.File, *os.File) {
	t.Helper()
	fd, err := unix.Open("/dev/ptmx", unix.O_RDWR|unix.O_NOCTTY|unix.O_CLOEXEC|unix.O_NONBLOCK, 0)
	if err != nil {
		t.Skip("PTY unavailable:", err)
	}
	master := os.NewFile(uintptr(fd), "ptmx")
	if err := unix.IoctlSetPointerInt(fd, unix.TIOCSPTLCK, 0); err != nil {
		master.Close()
		t.Fatal(err)
	}
	number, err := unix.IoctlGetInt(fd, unix.TIOCGPTN)
	if err != nil {
		master.Close()
		t.Fatal(err)
	}
	slave, err := os.OpenFile(fmt.Sprintf("/dev/pts/%d", number), os.O_RDWR, 0)
	if err != nil {
		master.Close()
		t.Fatal(err)
	}
	if err := unix.IoctlSetWinsize(int(slave.Fd()), unix.TIOCSWINSZ, &unix.Winsize{Row: 30, Col: 60}); err != nil {
		master.Close()
		slave.Close()
		t.Fatal(err)
	}
	return master, slave
}
