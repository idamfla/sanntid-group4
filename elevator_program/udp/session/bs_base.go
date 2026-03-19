package session

import (
	"elevator_program/udp"
	"elevator_program/udp/timer"
	"fmt"
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
	broadcastAckTimer *timer.Timer
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
		broadcastAckTimer: timer.NewTimer(),
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
	if bbs.broadcastAckTimer != nil {
		bbs.broadcastAckTimer.Stop()
	}

	bbs.Session.Close()
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

func (bbs *BaseBroadcastSession) startAckTimer() {
	bbs.broadcastAckTimer.Restart(udp.BROADCAST_ACK_TIMEOUT, func() {
		fmt.Println("Not enough elevators received the data in time ...")
		// TODO what now??
		// ses.closeReq <- ses.ID
		bbs.requestClose()
	})
}

func (bbs *BaseBroadcastSession) stopAckTimer() {
	bbs.broadcastAckTimer.Stop()
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
