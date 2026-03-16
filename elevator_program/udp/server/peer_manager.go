package server

import (
	"elevator_program/udp/peer"
	"net"
	"time"
)

func (srv *Server) activePeerCount() int {
	srv.mu.Lock()
	defer srv.mu.Unlock()

	count := 0
	for _, p := range srv.netPeers {
		if p.Active {
			count++
		}
	}
	return count
}

func (srv *Server) nextSeq(addr *net.UDPAddr) uint32 {
	srv.mu.Lock()
	defer srv.mu.Unlock()

	key := addr.String()
	netPeer, ok := srv.netPeers[key]
	if !ok {
		netPeer = &peer.NetworkPeer{Addr: addr, Seq: 0, Active: true, LastSeen: time.Now()}
		srv.netPeers[key] = netPeer
	}

	// Assign current seq to the outgoing message
	seq := netPeer.Seq

	// Increment for next message
	netPeer.Seq++
	if netPeer.Seq >= seq { // optional wrap-around
		netPeer.Seq = 0
	}

	return seq
}
