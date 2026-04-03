package server

import (
	"elevator_program/message"
	"elevator_program/udp/packet"
)

type SessionHandler interface {
	Start()
	Close()
	// SendReply(pkt packet.PacketType)
	ReceivePacket(pkt packet.Packet)
	QueueWhoIsAliveMsg() // TODO this should be handled by the masterElect ses ... make this always use the same session, session 1
	QueueStateBSUpdateMsg(pktType packet.PacketType, eMsg message.ElevatorMessage)
}
