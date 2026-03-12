package session

import "time"

type SessionTimer struct {
	stop chan struct{}
}

func NewSessionTimer() *SessionTimer {
	return &SessionTimer{stop: make(chan struct{})}
}

func (t *SessionTimer) Stop() {
	select {
	case <-t.stop:
		// already stopped
	default:
		close(t.stop)
	}
}

func (t *SessionTimer) Restart(duration time.Duration, callback func()) {
	t.Stop() // stop previous timer

	t.stop = make(chan struct{})
	go func() {
		select {
		case <-time.After(duration):
			callback() // time expired
		case <-t.stop:
			// timer cancelled
		}
	}()
}
