package server

import (
	"elevator_program/message"
	"elevator_program/udp/packet"
	"net"
)

type incomingPacket struct {
	Addr   *net.UDPAddr
	Packet packet.Packet
}

type outgoingMessage struct {
	RemoteAddr *net.UDPAddr
	PktType    packet.PacketType
	EMsg       message.ElevatorMessage
}
