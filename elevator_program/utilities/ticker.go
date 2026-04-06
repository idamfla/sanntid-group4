package utilities

import "time"

func ResetTicker(t *time.Ticker, d time.Duration) {
	select {
	case <-t.C:
	default:
	}
	t.Reset(d)
}
