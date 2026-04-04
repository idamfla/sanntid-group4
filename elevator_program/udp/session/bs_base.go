package session

import (
	"elevator_program/udp"
	"elevator_program/utilities"
	"fmt"
	"sync"
)

type BaseBroadcastSession struct {
	*Session
	expectedResponses int
	responders        map[string]bool
	responseTimer     *utilities.Timer
	mu                sync.Mutex
}

func NewBaseBroadcastSession(
	id uint32,
	srv ServerAPI,
	expectedResponses int,
) *BaseBroadcastSession {
	selfAddr := srv.GetRecvString()
	addr := srv.GetBroadcastAddr()

	bbs := &BaseBroadcastSession{
		Session:           NewSession(id, addr, srv),
		expectedResponses: expectedResponses,
		responders:        make(map[string]bool),
		responseTimer:     utilities.NewTimer(),
	}
	bbs.responders[selfAddr] = true
	return bbs
}

func (bbs *BaseBroadcastSession) Start() {
	// bbs.wg.Add(2)
	// go bbs.listen(bbs)
	// go bbs.sendLoop(bbs)
	// fmt.Printf("Broadcast session %d started\n", bbs.ID)
}

func (bbs *BaseBroadcastSession) Close() {
	if bbs.responseTimer != nil {
		bbs.responseTimer.Stop()
	}

	bbs.Session.Close()
}

func (bbs *BaseBroadcastSession) addResponder(addr string) {
	bbs.mu.Lock()
	defer bbs.mu.Unlock()
	bbs.responders[addr] = true
}

func (bbs *BaseBroadcastSession) countResponders() int {
	return bbs.countTotalResponders() - 1
}

func (bbs *BaseBroadcastSession) countTotalResponders() int {
	bbs.mu.Lock()
	defer bbs.mu.Unlock()
	return len(bbs.responders)
}

func (bbs *BaseBroadcastSession) resetResponders() {
	bbs.mu.Lock()
	defer bbs.mu.Unlock()
	bbs.responders = map[string]bool{bbs.selfAddr: true}
}

func (bbs *BaseBroadcastSession) startResponseTimer() {
	bbs.responseTimer.Restart(udp.RESPONSE_TIMEOUT, func() {
		if bbs.countResponders() >= bbs.expectedResponses {
			// Ignore false timeout
			return
		}

		fmt.Printf("Peer(s) did not respond in time ... %d/%d\n", bbs.countResponders(), bbs.expectedResponses)
		bbs.queueWhoIsAliveMsg()
		bbs.stopResponseTimer()
		bbs.requestClose()
	})
}

func (bbs *BaseBroadcastSession) stopResponseTimer() {
	bbs.responseTimer.Stop()
}
