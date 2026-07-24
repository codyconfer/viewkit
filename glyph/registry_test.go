package glyph

import "testing"

func TestRegisterLookup(t *testing.T) {
	Register("plugin.demo", Variants{Nerd: "N", Uni: "U", ASCII: "A"})
	v, ok := Lookup("plugin.demo")
	if !ok || v.Nerd != "N" {
		t.Fatalf("Lookup = %+v ok=%v", v, ok)
	}
	SetMode(ModeNone)
	if got := ResolveID("plugin.demo"); got != "A" {
		t.Fatalf("ResolveID ascii = %q", got)
	}
	SetMode(ModeNerd)
}

func TestBuildStatusStripKeepsTone(t *testing.T) {
	strip := BuildStatusStrip("##", "work", []string{"k8s/prod"}, []StatusContribution{
		{Status: func() (string, Severity) { return "●", SeverityPositive }},
		{Status: func() (string, Severity) { return "⚠", SeverityWarning }},
		{Status: func() (string, Severity) { return "x", SeverityNegative }},
	})
	if len(strip.Left) != 3 || strip.Left[2] != "k8s/prod" {
		t.Fatalf("left = %v", strip.Left)
	}
	if len(strip.Right) != 3 {
		t.Fatalf("right = %v", strip.Right)
	}
	if strip.Right[0].Tone != SeverityPositive || strip.Right[1].Tone != SeverityWarning || strip.Right[2].Tone != SeverityNegative {
		t.Fatalf("tones = %+v", strip.Right)
	}
}
