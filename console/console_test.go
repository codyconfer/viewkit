package console

import (
	"bytes"
	"io"
	"os"
	"testing"
)

func TestDetectEnvironment(t *testing.T) {
	tests := []struct {
		name string
		tty  bool
		env  map[string]string
		want Capabilities
	}{
		{"redirected ignores overrides", false, map[string]string{"LANG": "en_US.UTF-8", "VIEWKIT_NERD_FONT": "true"}, Capabilities{Powerline: Unsupported, NerdFont: Unsupported}},
		{"dumb ignores overrides", true, map[string]string{"TERM": "DuMb", "VIEWKIT_POWERLINE": "1"}, Capabilities{Terminal: true, Powerline: Unsupported, NerdFont: Unsupported}},
		{"no locale", true, nil, Capabilities{Terminal: true, ANSI: true}},
		{"utf8 unknown fonts", true, map[string]string{"LANG": "en_US.UTF-8", "TERM": "xterm-256color"}, Capabilities{Terminal: true, ANSI: true, Unicode: true}},
		{"ctype wins over lang", true, map[string]string{"LC_CTYPE": "C", "LANG": "en_US.UTF-8"}, Capabilities{Terminal: true, ANSI: true}},
		{"all wins over ctype", true, map[string]string{"LC_ALL": "C.utf8", "LC_CTYPE": "C"}, Capabilities{Terminal: true, ANSI: true, Unicode: true}},
		{"posix locale", true, map[string]string{"LC_ALL": "POSIX"}, Capabilities{Terminal: true, ANSI: true}},
		{"powerline is not nerd", true, map[string]string{"VIEWKIT_POWERLINE": "true"}, Capabilities{Terminal: true, ANSI: true, Unicode: true, Powerline: Supported}},
		{"nerd includes powerline", true, map[string]string{"VIEWKIT_NERD_FONT": "1"}, Capabilities{Terminal: true, ANSI: true, Unicode: true, Powerline: Supported, NerdFont: Supported}},
		{"explicit powerline disable", true, map[string]string{"VIEWKIT_NERD_FONT": "true", "VIEWKIT_POWERLINE": "false"}, Capabilities{Terminal: true, ANSI: true, Unicode: true, Powerline: Unsupported, NerdFont: Supported}},
		{"disabled fonts", true, map[string]string{"LANG": "C.UTF-8", "VIEWKIT_NERD_FONT": "0", "VIEWKIT_POWERLINE": "false"}, Capabilities{Terminal: true, ANSI: true, Unicode: true, Powerline: Unsupported, NerdFont: Unsupported}},
		{"invalid and auto unknown", true, map[string]string{"VIEWKIT_NERD_FONT": "auto", "VIEWKIT_POWERLINE": "maybe"}, Capabilities{Terminal: true, ANSI: true}},
		{"terminal brand proves no fonts", true, map[string]string{"TERM_PROGRAM": "WezTerm", "WT_SESSION": "session", "LANG": "C.UTF-8"}, Capabilities{Terminal: true, ANSI: true, Unicode: true}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := detect(tt.tty, func(key string) string { return tt.env[key] })
			if got != tt.want {
				t.Errorf("detect() = %+v, want %+v", got, tt.want)
			}
		})
	}
}

func TestDetectNonTerminals(t *testing.T) {
	t.Setenv("VIEWKIT_NERD_FONT", "true")
	f, err := os.CreateTemp(t.TempDir(), "output")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = f.Close() })
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = r.Close(); _ = w.Close() })
	var nilFile *os.File
	for _, writer := range []io.Writer{nil, nilFile, &bytes.Buffer{}, f, w} {
		if got := Detect(writer); got != (Capabilities{Powerline: Unsupported, NerdFont: Unsupported}) {
			t.Errorf("Detect(%T) = %+v, want plain output", writer, got)
		}
	}
}

func TestFontSupport(t *testing.T) {
	for value, want := range map[string]Support{
		"": Unknown, "auto": Unknown, "yes": Unknown,
		" true ": Supported, "TRUE": Supported, "1": Supported,
		" false ": Unsupported, "FALSE": Unsupported, "0": Unsupported,
	} {
		if got := fontSupport(value); got != want {
			t.Errorf("fontSupport(%q) = %v, want %v", value, got, want)
		}
	}
}
