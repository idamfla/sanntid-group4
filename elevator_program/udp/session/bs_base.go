package session

import (
	"net"
	"sync"
)

type BroadcastSessionType int

const (
	BS_T_StateBroadcast BroadcastSessionType = iota
	BS_T_WhoIsMasterBroadcast
)

type BaseBroadcastSession struct {
	*Session
	selfAddr          string
	expectedResponses int
	responders        map[string]bool
	mu                sync.Mutex
}

func NewBaseBroadcastSession(
	id uint32,
	selfAddr string,
	addr *net.UDPAddr,
	closeReq chan<- uint32,
	tx PacketSender,
	expectedResponses int,
) *BaseBroadcastSession {
	bbs := &BaseBroadcastSession{
		Session:           NewSession(id, addr, closeReq, tx),
		selfAddr:          selfAddr,
		expectedResponses: expectedResponses,
		responders:        make(map[string]bool),
	}
	bbs.responders[selfAddr] = true
	return bbs
}

func (bbs *BaseBroadcastSession) addResponder(addr string) {
	bbs.mu.Lock()
	defer bbs.mu.Unlock()
	bbs.responders[addr] = true
}

func (bbs *BaseBroadcastSession) countResponders() int {
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
