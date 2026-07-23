package browser

import "testing"

func TestOpenPassesURL(t *testing.T) {
	var gotName string
	var gotArgs []string
	orig := run
	run = func(name string, args ...string) error {
		gotName, gotArgs = name, args
		return nil
	}
	defer func() { run = orig }()

	if err := Open("https://example.com/pr/1"); err != nil {
		t.Fatalf("Open: %v", err)
	}
	if gotName == "" {
		t.Fatal("expected a command to be invoked")
	}
	found := false
	for _, a := range gotArgs {
		if a == "https://example.com/pr/1" {
			found = true
		}
	}
	if !found {
		t.Fatalf("url not passed to opener: %s %v", gotName, gotArgs)
	}
}
