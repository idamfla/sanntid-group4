package server

import (
	"net"
	"time"
)

type PeerInfo struct {
	Addr     *net.UDPAddr
	LastSeen time.Time
	Active   bool
	IsMaster bool
}

func NewPeer(addr *net.UDPAddr) *PeerInfo {
	return &PeerInfo{
		Addr:     addr,
		LastSeen: time.Now(),
		Active:   true,
		IsMaster: false,
	}
}
