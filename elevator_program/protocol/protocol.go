package protocol

import (
	"elevator_program/message"
	"elevator_program/udp/server"
	"elevator_program/udp/session"
	"sync"
)

type Protocol struct {
	Server       *server.Server              // TODO be carefull with pass by value functions, locks
	msgRecieveCh chan session.ElevatorPacket // Update the channel type, wait should this one be IncomingPacket, do i need to debug and encode this one?
	msgSendCh    chan message.Message
	wg           sync.WaitGroup
	activePeers  int
}

func (p *Protocol) InitProtocol() { // TODO how can I allways now excactly how many elevators we are going to use
	p.msgRecieveCh = make(chan session.ElevatorPacket, 10) // Match the expected type
	p.msgSendCh = make(chan message.Message, 10)
	p.activePeers = 1
	// p.Server, _ = server.NewServer(ip, port, id, numElevators, p.msgRecieveCh)
}
