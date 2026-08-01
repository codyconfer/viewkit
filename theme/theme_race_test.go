package theme

import "testing"

func TestUseIsRaceFreeWhileRendering(t *testing.T) {
	orig := Cur()
	t.Cleanup(func() { Use(orig) })

	dark, _ := Named("solarized-dark")
	light, _ := Named("solarized-light")

	stop := make(chan struct{})
	done := make(chan struct{})
	go func() {
		defer close(done)
		for {
			select {
			case <-stop:
				return
			default:
				th := Cur()
				_ = th.Accent.Render("frame")
				_ = th.Dim.Render("msg")
				_ = StripText(th.Val.GetForeground(), "strip")
			}
		}
	}()

	for i := range 500 {
		if i%2 == 0 {
			Use(dark)
		} else {
			Use(light)
		}
	}

	close(stop)
	<-done
}

func TestCurStylesAreRaceFreeWhileUse(t *testing.T) {
	orig := Cur()
	t.Cleanup(func() { Use(orig) })

	dark, _ := Named("solarized-dark")

	stop := make(chan struct{})
	done := make(chan struct{})
	go func() {
		defer close(done)
		for {
			select {
			case <-stop:
				return
			default:
				th := Cur()
				_ = th.Dim.Render("dim")
				_ = th.Accent.Render("accent")
				_ = th.AppFrame.Render("frame")
				_ = len(th.Series)
			}
		}
	}()

	for range 500 {
		Use(dark)
	}

	close(stop)
	<-done
}
