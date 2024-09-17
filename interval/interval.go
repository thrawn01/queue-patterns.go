package interval

import (
	"github.com/kapetan-io/tackle/clock"
	"sync"
	"time"
)

// Interval is a one-shot ticker.  Call `Next()` to trigger the start of the
// next interval.  Read the `C` channel for tick event.
type Interval struct {
	C     chan struct{}
	in    chan struct{}
	done  chan struct{}
	wg    sync.WaitGroup
	mutex sync.Mutex
}

// NewInterval creates a new ticker like object, however
// the `C` channel does not return the current time and
// `C` channel will only get a tick after `Next()` has
// been called.
func NewInterval(d clock.Duration) *Interval {
	i := Interval{
		C:  make(chan struct{}, 1),
		in: make(chan struct{}, 1),
	}
	i.run(d)
	return &i
}

func (i *Interval) run(d clock.Duration) {
	i.mutex.Lock()
	if i.done == nil {
		i.done = make(chan struct{})
	}
	i.mutex.Unlock()
	i.wg.Add(1)

	go func() {
		for {
			select {
			case <-i.in:
				time.Sleep(d)
				i.C <- struct{}{}
			case <-i.done:
				i.wg.Done()
				return
			}
		}
	}()
}

func (i *Interval) Stop() {
	i.mutex.Lock()
	defer i.mutex.Unlock()

	if i.done != nil {
		close(i.done)
	}
	i.wg.Wait()
	i.done = nil
}

// Next queues the next interval to run, If multiple calls to Next() are
// made before previous intervals have completed they are ignored.
func (i *Interval) Next() {
	select {
	case i.in <- struct{}{}:
	default:
	}
}
