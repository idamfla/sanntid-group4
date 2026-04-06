package packet

import (
	"elevator_program/message"
)

type OutgoingMessage struct {
	Origin  Identity // id and addr of the initater/original sender
	PktType PacketType
	EMsg    message.ElevatorMessage // the actual message that the elevator will interpret
}
