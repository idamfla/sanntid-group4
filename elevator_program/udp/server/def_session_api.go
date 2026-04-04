package server

import (
	"elevator_program/message"
	"elevator_program/udp/packet"
)

type SessionHandler interface {
	Start()
	Close()
	GetID() uint32
	ReceivePacket(pkt packet.Packet)
	QueueDirectMsg(pktType packet.PacketType, eMsg message.ElevatorMessage) // TODO this should be handled by the masterElect ses ... make this always use the same session, session 1
	QueueStateBSUpdateMsg(pktType packet.PacketType, eMsg message.ElevatorMessage)
}

func (srv *Server) QueueWhoIsAliveMsg() {
	ws := srv.getOrCreateMasterElectionSession()
	ws.QueueDirectMsg(packet.PKT_T_WhoIsAlive, message.ElevatorMessage{})
}
