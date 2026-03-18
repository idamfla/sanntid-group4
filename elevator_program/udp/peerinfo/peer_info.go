package peerinfo

import (
	"elevator_program/message"
	"net"
	"time"
)

type PeerInfo struct {
	Addr      *net.UDPAddr
	LastSeen  time.Time
	IsSynced  bool
	Active    bool
	IsMaster  bool
	EMsgQueue chan message.ElevatorMessage
}

func NewPeer(addr *net.UDPAddr) *PeerInfo {
	return &PeerInfo{
		Addr:      addr,
		LastSeen:  time.Now(),
		IsSynced:  false,
		Active:    true,
		IsMaster:  false,
		EMsgQueue: make(chan message.ElevatorMessage, 10), // TODO make all channels bufferd ... how much?
	}
}

func (peer *PeerInfo) SetMaster(isMaster bool) {
	peer.IsMaster = isMaster
}

func (peer *PeerInfo) QueueMessage(msg message.ElevatorMessage) {
	select {
	case peer.EMsgQueue <- msg:
	default:
	}
}
