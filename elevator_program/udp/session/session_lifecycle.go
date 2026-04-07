package session

import (
	"elevator_program/udp"
	"elevator_program/utilities"
	"sync"
)

type SessionLifecycle struct {
	ShutdownTimer *utilities.Timer
	CloseReq      chan<- uint32 // make the server/owner close this session
	Stop          chan struct{}
	Wg            sync.WaitGroup
	CloseOnce     sync.Once
}

func NewSessionLifecycle(closeReq chan<- uint32) *SessionLifecycle {
	return &SessionLifecycle{
		ShutdownTimer: utilities.NewTimer(),
		CloseReq:      closeReq,
		Stop:          make(chan struct{}, CHANNEL_BUF),
	}
}

func (ses *Session) WgAdd(value int) { ses.lifecycle.Wg.Add(value) }
func (ses *Session) WgWait()         { ses.lifecycle.Wg.Wait() }
func (ses *Session) WgDone()         { ses.lifecycle.Wg.Done() }

func (ses *Session) startShutdownTimer() {
	ses.lifecycle.ShutdownTimer.Restart(udp.SHUTDOWN_TIMEOUT, func() {
		ses.requestClose()
	})
}

func (ses *Session) stopShutdownTimer() {
	ses.lifecycle.ShutdownTimer.Stop()
}

func (ses *Session) requestClose() {
	select {
	case <-ses.stopCh():
	case ses.lifecycle.CloseReq <- ses.ID:
	}
}

func (ses *Session) stopCh() <-chan struct{} { return ses.lifecycle.Stop }
