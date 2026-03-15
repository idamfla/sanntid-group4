package session

import "elevator_program/udp/packet"

type SessionBehavior interface {
	HandlePacket(pkt packet.Packet) error
	OnSend(pktType packet.PacketType)
}
