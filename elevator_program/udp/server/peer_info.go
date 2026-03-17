package server

import (
	"elevator_program/udp/message"
	"net"
	"time"
)

type PeerInfo struct {
	Addr     *net.UDPAddr
	LastSeen time.Time
	IsSynced bool
	Active   bool
	IsMaster bool
	msgQueue chan message.Message
}

func NewPeer(addr *net.UDPAddr) *PeerInfo {
	return &PeerInfo{
		Addr:     addr,
		LastSeen: time.Now(),
		IsSynced: false,
		Active:   true,
		IsMaster: false,
		msgQueue: make(chan message.Message),
	}
}

func (peer *PeerInfo) QueueMessage(msg message.Message) {
	select {
	case peer.msgQueue <- msg:
	default:
	}
}
