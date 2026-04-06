package server

import (
	"elevator_program/udp/packet"
)

type SessionHandler interface {
	Start()
	Close()
	GetID() uint32
	ReceivePacket(pkt packet.Packet)
	QueueDirectMsg(pktType packet.PacketType, outMsg packet.OutgoingMessage) // TODO this should be handled by the masterElect ses ... make this always use the same session, session 1
}
