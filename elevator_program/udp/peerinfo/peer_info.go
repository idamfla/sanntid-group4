package peerinfo

import (
	"elevator_program/message"
	"net"
	"time"
)

const (
	CHANNEL_BUF = 32
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
		EMsgQueue: make(chan message.ElevatorMessage, CHANNEL_BUF),
	}
}

func (peer *PeerInfo) SetMaster(isMaster bool) {
	peer.IsMaster = isMaster
}

func (peer *PeerInfo) SetIsSynced(isSynced bool) {
	peer.IsSynced = true
}

func (peer *PeerInfo) QueueMessage(msg message.ElevatorMessage) {
	select {
	case peer.EMsgQueue <- msg:
	default:
	}
}
