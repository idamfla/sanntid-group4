package session

import (
	"elevator_program/udp/packet"
	"sync"
)

type SessionState struct {
	Seq        uint32
	PendingMsg *packet.OutgoingMessage
	LastOutMsg packet.OutgoingMessage
	HasLastMsg bool
	mu         sync.Mutex
}

func NewSessionState() *SessionState {
	return &SessionState{
		Seq:        0,
		PendingMsg: &packet.OutgoingMessage{},
		LastOutMsg: packet.OutgoingMessage{},
		HasLastMsg: false,
	}
}

func (ss *SessionState) GetSeq() uint32 {
	ss.lock()
	defer ss.unlock()
	return ss.Seq
}

func (ss *SessionState) IncrementSeq() {
	ss.lock()
	defer ss.unlock()
	ss.Seq++
}

func (ss *SessionState) GetPendingMsg() packet.OutgoingMessage {
	ss.lock()
	defer ss.unlock()
	return *ss.PendingMsg
}

func (ss *SessionState) SetPendingMsg(pendingMsg *packet.OutgoingMessage) {
	ss.lock()
	defer ss.unlock()
	ss.PendingMsg = pendingMsg
}

func (ss *SessionState) ClearPendingMsg() {
	ss.lock()
	defer ss.unlock()
	ss.PendingMsg = nil
}

func (ss *SessionState) GetLastOutMsg() packet.OutgoingMessage {
	ss.lock()
	defer ss.unlock()
	return ss.LastOutMsg
}

func (ss *SessionState) SetLastOutMsg(outMsg packet.OutgoingMessage) {
	ss.lock()
	defer ss.unlock()
	ss.LastOutMsg = outMsg
}

func (ss *SessionState) GetHasLastMsg() bool {
	ss.lock()
	defer ss.unlock()
	return ss.HasLastMsg
}

func (ss *SessionState) SetHasLastMsg() {
	ss.lock()
	defer ss.unlock()
	ss.HasLastMsg = true
}

func (ss *SessionState) ClearHasLastMsg() {
	ss.lock()
	defer ss.unlock()
	ss.HasLastMsg = false
}

func (ss *SessionState) lock()   { ss.mu.Lock() }
func (ss *SessionState) unlock() { ss.mu.Unlock() }
