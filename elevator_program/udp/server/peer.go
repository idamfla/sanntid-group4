package server

import (
	"elevator_program/message"
	"net"
	"time"
)

type PeerView struct {
	Addr      *net.UDPAddr
	LastSeen  time.Time
	IsSynced  bool
	Active    bool
	IsMaster  bool
	EMsgQueue chan message.ElevatorMessage
}

func NewPeer(addr *net.UDPAddr) *PeerView {
	return &PeerView{
		Addr:      addr,
		LastSeen:  time.Now(),
		IsSynced:  false,
		Active:    true,
		IsMaster:  false,
		EMsgQueue: make(chan message.ElevatorMessage, CHANNEL_BUF), // TODO make all channels bufferd ... how much?
	}
}

func (peer *PeerView) SetMaster(isMaster bool) {
	peer.IsMaster = isMaster
}

func (peer *PeerView) SetIsSynced(isSynced bool) {
	peer.IsSynced = true
}

func (peer *PeerView) QueueMessage(msg message.ElevatorMessage) {
	select {
	case peer.EMsgQueue <- msg:
	default:
	}
}
