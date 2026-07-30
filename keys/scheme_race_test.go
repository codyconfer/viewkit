package keys

import (
	"sync"
	"testing"
)

func TestUseIsRaceFreeWhileReadingBindings(t *testing.T) {
	orig := Cur()
	t.Cleanup(func() { Use(orig) })

	alt := Default().With(Binding{Keys: []string{"x"}, Action: Open})

	const iterations = 20000
	start := make(chan struct{})
	var wg sync.WaitGroup
	wg.Add(2)

	go func() {
		defer wg.Done()
		<-start
		for range iterations {
			_ = Cur().Binding(Up)
			_ = Cur().Binding(Open)
		}
	}()

	go func() {
		defer wg.Done()
		<-start
		for i := range iterations {
			if i%2 == 0 {
				Use(alt)
			} else {
				Use(orig)
			}
		}
	}()

	close(start)
	wg.Wait()
}
