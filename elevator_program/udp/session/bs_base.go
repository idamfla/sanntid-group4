package session

import (
	"elevator_program/udp"
	"fmt"
	"sync"
)

type BroadcastSessionType int

const (
	BS_T_StateBroadcast BroadcastSessionType = iota
	BS_T_WhoIsMasterBroadcast
)

type BaseBroadcastSession struct {
	*Session
	// selfAddr          string
	expectedResponses int
	responders        map[string]bool
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

func (bsType BroadcastSessionType) String() string {
	switch bsType {
	case BS_T_StateBroadcast:
		return "State Broadcast"
	case BS_T_WhoIsMasterBroadcast:
		return "Who Is Master Broadcast"
	default:
		return "unknown"
	}
}

func (bbs *BaseBroadcastSession) startResponseTimer() {
	bbs.responseTimer.Restart(udp.RESPONSE_TIMEOUT, func() {
		fmt.Printf("Peer(s) did not respond in time ... %d/%d\n", bbs.countResponders(), bbs.expectedResponses)
		bbs.queueWhoIsAliveMsg()
		bbs.stopResponseTimer()
		bbs.requestClose()
	})
}
