package server

import (
	"elevator_program/udp/packet"
)

type SessionHandler interface {
	Start()
	Close()
	GetID() uint32
	ReceivePacket(pkt packet.Packet)
	QueueDirectMsg(pktType packet.PacketType, outMsg packet.OutgoingMessage)
}
