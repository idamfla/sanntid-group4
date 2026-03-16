package session

import (
	"elevator_program/udp/message"
	"elevator_program/udp/packet"
)

// Session -> Elevator
type ElevatorPacket struct {
	Packet packet.Packet
	Done   chan<- struct{}
}

// Session -> Session
type outgoingMessage struct { // TODO rename OutgoingMessage
	PktType packet.PacketType
	Msg     message.Message
	Done    chan struct{} // TODO rename, Commited, or something
}
