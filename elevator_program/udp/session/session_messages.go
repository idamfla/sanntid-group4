package session

import (
	"elevator_program/message"
	"elevator_program/udp/packet"
)

type ElevatorPacket struct {
	EMsg message.ElevatorMessage
	Done chan<- struct{}
}

type outgoingMessage struct {
	PktType packet.PacketType
	EMsg    message.ElevatorMessage
	Done    chan struct{}
}
