package glyph

import "testing"

func TestRegisterNamed(t *testing.T) {
	Register("plugin.demo", Variants{Nerd: "N", Uni: "U", ASCII: "A"})
	v, ok := Named("plugin.demo")
	if !ok || v.Nerd != "N" {
		t.Fatalf("Lookup = %+v ok=%v", v, ok)
	}
	SetMode(ModeNone)
	if got := ResolveID("plugin.demo"); got != "A" {
		t.Fatalf("ResolveID ascii = %q", got)
	}
	SetMode(ModeNerd)
}

func TestBuiltinBrandIDs(t *testing.T) {
	SetMode(ModeNerd)
	for _, id := range []string{"github", "GitHub", "slack", "google"} {
		if got := ResolveID(id); got == "" {
			t.Fatalf("ResolveID(%q) empty", id)
		}
	}
	if ResolveID("github") != GitHub() {
		t.Fatalf("github = %q, want GitHub()", ResolveID("github"))
	}
	if ResolveID("slack") != Slack() {
		t.Fatalf("slack = %q, want Slack()", ResolveID("slack"))
	}
	if ResolveID("google") != Google() {
		t.Fatalf("google = %q, want Google()", ResolveID("google"))
	}
}

func TestNormalizeID(t *testing.T) {
	if got := NormalizeID("  GitHub "); got != "github" {
		t.Fatalf("NormalizeID = %q", got)
	}
	Register("MiXeD.ID", Variants{Nerd: "X", Uni: "x", ASCII: "x"})
	if _, ok := Named("mixed.id"); !ok {
		t.Fatal("Register should normalize id")
	}
}

func TestBuildStatusStripKeepsSeverity(t *testing.T) {
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
	if strip.Right[0].Severity != SeverityPositive || strip.Right[1].Severity != SeverityWarning || strip.Right[2].Severity != SeverityNegative {
		t.Fatalf("severities = %+v", strip.Right)
	}
}
