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
	RemoteAddr *net.UDPAddr // addr of receiver
	PktType    packet.PacketType
	EMsg       message.ElevatorMessage // the actual message that the elevator will interpret
}
