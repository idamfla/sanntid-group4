package session

import (
	"elevator_program/message"
)

// Session -> Elevator
type ElevatorPacket struct {
	EMsg message.ElevatorMessage
	Done chan<- struct{}
}
