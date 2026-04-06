package packet

import (
	"elevator_program/message"
	"net"
)

type OutgoingMessage struct {
	Origin     Identity     // id and addr of the initater/original sender
	RemoteAddr *net.UDPAddr // addr of receiver
	PktType    PacketType
	EMsg       message.ElevatorMessage // the actual message that the elevator will interpret
}
