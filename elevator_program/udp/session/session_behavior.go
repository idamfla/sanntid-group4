package session

import "elevator_program/udp/packet"

type SessionBehavior interface {
	HandlePacket(incPkt IncomingPacket) error
	OnSend(pktType packet.PacketType)
}
