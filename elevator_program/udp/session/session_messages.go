package session

import (
	"elevator_program/message"
	"elevator_program/udp/packet"
	"net"
)

type IncomingPacket struct {
	Addr   *net.UDPAddr
	Packet packet.Packet
}

type ElevatorPacket struct {
	Packet packet.Packet
	Done   chan<- struct{}
}

type OutgoingPacket struct { // TODO rename OutgoingMessage
	PktType packet.PacketType
	Msg     message.Message
	Done    chan struct{} // TODO rename, Commited, or something
}
