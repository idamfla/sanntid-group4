package utilities

import (
	"sync"
	"time"
)

type Timer struct {
	mu    sync.Mutex
	timer *time.Timer
}

func NewTimer() *Timer {
	return &Timer{}
}

func (t *Timer) Stop() {
	t.mu.Lock()
	defer t.mu.Unlock()
	if t.timer != nil {
		t.timer.Stop()
		t.timer = nil
	}
}

func (t *Timer) Restart(duration time.Duration, callback func()) {
	t.Stop() // stop previous timer

	t.mu.Lock()
	t.timer = time.AfterFunc(duration, callback)
	t.mu.Unlock()
}
