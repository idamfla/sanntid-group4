package server

import (
	"elevator_program/message"
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

	if (forceSync || isNew || wasRevived) && srv.IsMaster() {
		fmt.Println("sync peer")
	}
}

func (srv *Server) GetMasterPeer() *PeerInfo {
	srv.mu.Lock()
	defer srv.mu.Unlock()

	return srv.getMasterPeerLocked()
}

func (srv *Server) getMasterPeerLocked() *PeerInfo {
	for _, p := range srv.peers {
		if p.IsMaster {
			return p
		}
	}
	return nil
}

func (srv *Server) getPeerCount() int {
	count := 0
	srv.mu.Lock()
	defer srv.mu.Unlock()

	for _, peer := range srv.peers {
		if peer.Active {
			count++
		}
	}
	return count
}

func (srv *Server) flushPeerPendingMsg(peer *PeerInfo) { // TODO should this be flushed to a catchup session, have it wait for the ack before sending next thing, then lastly have the flush done?
	defer srv.wg.Done()

	for {
		select {
		case <-srv.stop:
			return

		case msg := <-peer.EMsgQueue:
			srv.QueueMessage(peer.Addr, packet.PROTO_PKT_T_CatchupUpdate, msg)

		default:
			fmt.Println("Flush DONE")
			srv.QueueMessage(nil, packet.PROTO_PKT_T_SyncComplete, message.ElevatorMessage{
				EMsgType: message.EMSG_T_SyncedElevator, // TODO created a new messagetype for syncing elevators
				ID:       peer.Addr.String(),
				Addr:     peer.Addr.String(),
			})
			srv.mu.Lock()
			peer.IsSynced = true
			srv.mu.Unlock()
			return
		}
	}
}

func (srv *Server) StartPeerCatchup(peerAddr *net.UDPAddr) {
	peer, isNew := srv.getOrCreatePeer(peerAddr)
	if isNew {
		return
	}
	srv.wg.Add(1)
	go srv.flushPeerPendingMsg(peer)
}

func (srv *Server) setMasterPeer(peerID string, isMaster bool) {
	srv.mu.Lock()
	defer srv.mu.Unlock()
	if peer, exists := srv.peers[peerID]; exists {
		peer.SetMaster(isMaster)
	}

	if isMaster {
		srv.clearMasterSearch()
	}

}

func (srv *Server) clearAllPeerActive() {
	srv.mu.Lock()
	defer srv.mu.Unlock()
	for key, peer := range srv.peers {
		if peer == nil {
			fmt.Printf("%s -> nil\n", key)
			continue
		}

		peer.ClearMaster()
	}
}

func (srv *Server) PrintPeers() {
	srv.mu.Lock()
	defer srv.mu.Unlock()

	msg := "=== Peers map ===\n"

	for key, peer := range srv.peers {
		if peer == nil {
			msg += fmt.Sprintf("%s -> nil\n", key)
			continue
		}
		msg += fmt.Sprintf(`%s -> Addr: %v
	IsMaster: %v
	Active:   %v
	IsSynced: %v
	LastSeen: %v
`,
			key, peer.Addr, peer.IsMaster, peer.Active, peer.IsSynced, peer.LastSeen)
	}

	msg += `=================
	`
	fmt.Println(msg)
}
