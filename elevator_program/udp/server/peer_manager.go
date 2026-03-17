package server

import (
	"elevator_program/udp/packet"
	"fmt"
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

func (srv *Server) getOrCreatePeer(addr *net.UDPAddr) (*PeerInfo, bool) {
	srv.mu.Lock()
	defer srv.mu.Unlock()

	key := addr.String()
	peer, exists := srv.peers[key]
	if !exists {
		peer = NewPeer(addr)
		srv.peers[key] = peer
		fmt.Printf("Server %s: new peer made: %s\n", srv.ID, key)
		return peer, true
	}

	peer.LastSeen = time.Now()
	return peer, false
}

func (srv *Server) registerOrUpdatePeer(addr *net.UDPAddr, forceSync bool) {
	peer, isNew := srv.getOrCreatePeer(addr)

	wasRevived := !isNew && !peer.Active
	if wasRevived {
		peer.Active = true
		peer.LastSeen = time.Now()
	}

	if (forceSync || isNew || wasRevived) && srv.isMaster {
		fmt.Println("sync peer")
		// TODO request elevator to sync this peer
		// go srv.syncPeer(addr.String()) // TODO this is handled by elevator
	}
}

// func (srv *Server) syncPeer(key string) {
// 	srv.mu.Lock()
// 	peer := srv.peers[key]
// 	srv.mu.Unlock()

// 	srv.QueueMessage(peer.Addr, packet.PROTO_PKT_T_Data, message.Message{Content: "information to sync ..."}) // TODO what this should actually contain, should it be another type? sync this content kind of, snapshot or something
// }

func (srv *Server) isMasterKnown() bool {
	for _, p := range srv.peers {
		if p.IsMaster {
			return true
		}
	}
	return false
}

func (srv *Server) getAliveUnsyncedPeers() []*PeerInfo {
	srv.mu.Lock()
	peers := make([]*PeerInfo, 0, len(srv.peers))
	for _, p := range srv.peers {
		if p.Active && !p.IsSynced {
			peers = append(peers, p)
		}
	}
	srv.mu.Unlock()

	return peers
}

func (srv *Server) flushPeerPendingMsg(peer *PeerInfo) {
	done := false
	for !done {
		select {
		case msg := <-peer.msgQueue:
			srv.QueueMessage(peer.Addr, packet.PROTO_PKT_T_StateSnapshot, msg)
		default:
			done = true // no more messages
		}
	}
}
