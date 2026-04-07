package session

import "elevator_program/udp/packet"

func (ses *Session) prepareSend(outMsg packet.OutgoingMessage) (seq uint32, lastMsg packet.OutgoingMessage) {
	return ses.state.PrepareSend(outMsg)
}

func (ses *Session) getSeq() uint32 { return ses.state.GetSeq() }
func (ses *Session) shouldIncrementSeq(msgSeq uint32, shouldInc bool) (newSeq uint32, err error) {
	return ses.state.ShouldIncrementSeq(ses.ID, msgSeq, shouldInc)
}

func (ses *Session) getPendingMsg() packet.OutgoingMessage       { return ses.state.GetPendingMsg() }
func (ses *Session) setPendingMsg(penMsg packet.OutgoingMessage) { ses.state.SetPendingMsg(penMsg) }

func (ses *Session) getLastOutMsg() packet.OutgoingMessage { return ses.state.GetLastOutMsg() }
func (ses *Session) hasLastMsg() bool                      { return ses.state.GetHasLastMsg() }
func (ses *Session) clearLastMsg()                         { ses.state.ClearLastMsg() }
