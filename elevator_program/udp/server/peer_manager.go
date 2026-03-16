package server

import (
	"elevator_program/udp/peer_info"
	"net"
	"time"
)

func (srv *Server) activePeerCount() int {
	srv.mu.Lock()
	defer srv.mu.Unlock()

	count := 0
	for _, p := range srv.peers {
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
	peer, ok := srv.peers[key]
	if !ok {
		peer = &peer_info.PeerInfo{Addr: addr, Seq: 0, Active: true, LastSeen: time.Now()}
		srv.peers[key] = peer
	}

	// Assign current seq to the outgoing message
	seq := peer.Seq

	// Increment for next message
	peer.Seq++
	if peer.Seq >= seq { // optional wrap-around
		peer.Seq = 0
	}

	return seq
}
