package session

import (
	"elevator_program/udp/packet"
	"net"
)

type PacketContext struct {
	Packet packet.Packet
	Addr   *net.UDPAddr
	Done   chan<- struct{}
}
