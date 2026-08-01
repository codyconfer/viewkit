// Package browser opens URLs in the user's default web browser.
package browser

import (
	"os/exec"
	"runtime"
)

var run = func(name string, args ...string) error {
	return exec.Command(name, args...).Start()
}

// Open launches the platform's URL opener for url and reports failure to start it.
func Open(url string) error {
	switch runtime.GOOS {
	case "darwin":
		return run("open", url)
	case "windows":
		return run("rundll32", "url.dll,FileProtocolHandler", url)
	default:
		return run("xdg-open", url)
	}
}
