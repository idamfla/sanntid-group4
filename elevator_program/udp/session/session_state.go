package session

import "elevator_program/udp/packet"

func (ses *Session) getSeq() uint32 { return ses.state.GetSeq() }
func (ses *Session) incrementSeq()  { ses.state.IncrementSeq() }

func (ses *Session) getPendingMsg() packet.OutgoingMessage { return ses.state.GetPendingMsg() }

func (ses *Session) setPendingMsg(pendingMsg *packet.OutgoingMessage) {
	ses.state.SetPendingMsg(pendingMsg)
}

func (ses *Session) clearPendingMsg() { ses.state.ClearPendingMsg() }

func (ses *Session) getLastOutMsg() packet.OutgoingMessage { return ses.state.GetLastOutMsg() }
func (ses *Session) hasLastMsg() bool                      { return ses.state.GetHasLastMsg() }

func (ses *Session) setLastOutMsg(outMsg packet.OutgoingMessage) {
	ses.state.SetLastOutMsg(outMsg)
	ses.state.SetHasLastMsg()
}

func (ses *Session) clearLastMsg() {
	ses.state.SetLastOutMsg(packet.OutgoingMessage{})
	ses.state.ClearHasLastMsg()
}
