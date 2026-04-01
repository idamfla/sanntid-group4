package session

import (
	"elevator_program/message"
	"elevator_program/udp/packet"
)

// Session -> Elevator
type ElevatorPacket struct {
	EMsg message.ElevatorMessage
	Done chan<- struct{}
}

// Session -> Session
type outgoingMessage struct { // TODO rename OutgoingMessage
	PktType packet.PacketType
	EMsg    message.ElevatorMessage
	Done    chan struct{} // TODO rename, Commited, or something. could be ready or something ...
}
