package browser

import (
	"os/exec"
	"runtime"
)

var run = func(name string, args ...string) error {
	return exec.Command(name, args...).Start()
}

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
