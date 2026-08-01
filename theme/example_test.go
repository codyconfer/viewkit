package theme_test

import "github.com/codyconfer/viewkit/theme"

func ExampleNamed() {
	th := theme.Default()
	if named, ok := theme.Named("monokai"); ok {
		th = named
	}

	body := th.Accent.Render("ready")
	_ = th.Screen(body, theme.MinScreenWidth, theme.MinBodyHeight)
}
