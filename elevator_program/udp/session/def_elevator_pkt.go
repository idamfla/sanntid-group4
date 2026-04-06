package session

import (
	"elevator_program/message"
)

// Session -> Elevator
type ElevatorPacket struct {
	EMsg message.ElevatorMessage
	Done chan<- struct{}
}

// Session -> Session
// type outgoingMessage struct { // TODO rename OutgoingMessage
// 	PktType packet.PacketType
// 	EMsg    message.ElevatorMessage
// }
