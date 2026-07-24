package layout

import (
	"regexp"
	"strings"
	"testing"
)

var tbAnsi = regexp.MustCompile("\x1b\\[[0-9;]*m")

func TestTitledBoxShowsTitleAndBody(t *testing.T) {
	out := tbAnsi.ReplaceAllString(NewFrame(40).TitledBox("STATUS", "line one", "line two"), "")
	for _, want := range []string{"STATUS", "line one", "line two", "╭", "╰"} {
		if !strings.Contains(out, want) {
			t.Errorf("titled box missing %q:\n%s", want, out)
		}
	}
}

func TestTitledBoxIconRendersIcon(t *testing.T) {
	out := tbAnsi.ReplaceAllString(NewFrame(40).TitledBoxIcon("»", "AUTH"), "")
	if !strings.Contains(out, "»") || !strings.Contains(out, "AUTH") {
		t.Errorf("titled box icon missing icon/title:\n%s", out)
	}
}
