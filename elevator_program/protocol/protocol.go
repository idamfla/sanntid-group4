package protocol

import (
	"elevator_program/udp/packet"
	"elevator_program/udp/server"
	"elevator_program/udp/session"
)

type Protocol struct {
	Server         *server.Server             // TODO be carefull with pass by value functions, locks
	msgRecieveCh   chan session.PacketContext // Update the channel type
	msgSendCh      chan session.PacketContext
	outgoingPacket session.PacketContext
}

func (p *Protocol) InitProtocol(ip string, port int, id string) {
	p.msgRecieveCh = make(chan session.PacketContext, 10) // Match the expected type
	p.msgSendCh = make(chan session.PacketContext, 10)
	p.Server, _ = server.NewServer(ip, port, id, p.msgRecieveCh)
	p.outgoingPacket = session.PacketContext{
		Packet: packet.Packet{
			Header: packet.Header{ // TODO Just write something here i guess
				Seq: 1023,
			},
		},
	}
}
