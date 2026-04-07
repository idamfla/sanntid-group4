package session

import (
	"elevator_program/udp/packet"
	"fmt"
	"sync"
)

type SessionState struct {
	Seq        uint32
	PendingMsg packet.OutgoingMessage
	LastOutMsg packet.OutgoingMessage
	HasLastMsg bool
	mu         sync.Mutex
}

func NewSessionState() *SessionState {
	return &SessionState{
		Seq:        0,
		PendingMsg: packet.OutgoingMessage{},
		LastOutMsg: packet.OutgoingMessage{},
		HasLastMsg: false,
	}
}

func (ss *SessionState) PrepareSend(outMsg packet.OutgoingMessage) (seq uint32, lastOutMsg packet.OutgoingMessage) {
	ss.lock()
	defer ss.unlock()
	ss.Seq++
	ss.LastOutMsg = outMsg
	ss.HasLastMsg = true
	return ss.Seq, ss.LastOutMsg
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

func (ss *SessionState) ShouldIncrementSeq(sesID uint32, msgSeq uint32, shouldInc bool) (newSeq uint32, err error) {
	ss.lock()
	defer ss.unlock()
	expectedSeq := ss.Seq + 1

	if msgSeq <= ss.Seq {
		err := fmt.Errorf("Session %d: seq already recieved (got %d, expected %d)...\n",
			sesID, msgSeq, expectedSeq)
		fmt.Println(err)
		return msgSeq, err
	}

	if msgSeq != expectedSeq {
		err := fmt.Errorf("Session %d: seq mismatch (got %d, expected %d)...\n",
			sesID, msgSeq, expectedSeq)
		fmt.Println(err)
		return msgSeq, err
	}

	if shouldInc {
		ss.Seq = expectedSeq
	}
	return ss.Seq, nil
}

func (ss *SessionState) GetPendingMsg() packet.OutgoingMessage { // TODO this cause issues bc nil pendingMsg
	ss.lock()
	defer ss.unlock()
	return ss.PendingMsg
}

func (ss *SessionState) SetPendingMsg(pendingMsg packet.OutgoingMessage) {
	ss.lock()
	defer ss.unlock()
	ss.PendingMsg = pendingMsg
}

func (ss *SessionState) GetLastOutMsg() packet.OutgoingMessage {
	ss.lock()
	defer ss.unlock()
	return ss.LastOutMsg
}

func (ss *SessionState) ClearLastMsg() {
	ss.lock()
	defer ss.unlock()
	ss.LastOutMsg = packet.OutgoingMessage{}
	ss.HasLastMsg = false
}

func (ss *SessionState) GetHasLastMsg() bool {
	ss.lock()
	defer ss.unlock()
	return ss.HasLastMsg
}

func (ss *SessionState) lock()   { ss.mu.Lock() }
func (ss *SessionState) unlock() { ss.mu.Unlock() }
