package server

import (
	"elevator_program/udp/packet"
	"elevator_program/udp/session"
	"fmt"
	"net"
)

func (srv *Server) addSession(id uint32, sh SessionHandler) {
	srv.mu.Lock()
	defer srv.mu.Unlock()
	srv.sessions[id] = sh
}

func (srv *Server) getSession(id uint32) (SessionHandler, bool) {
	srv.mu.Lock()
	defer srv.mu.Unlock()
	sh, exists := srv.sessions[id]
	return sh, exists
}

// returns session from session map. if no hits, a new server is made. NB! this functions contains locks
func (srv *Server) getOrCreateSession(senderAddr *net.UDPAddr, sessionID uint32) SessionHandler {
	ses, exists := srv.getSession(sessionID)
	if exists {
		return ses
	}

	return srv.createSession(senderAddr, &sessionID)
}

func (srv *Server) getOrCreateWhoIsMasterSession(sessionID uint32) SessionHandler {
	bs, exists := srv.getSession(sessionID)
	if exists {
		return bs
	}

	return srv.createBroadcastSession(&sessionID, session.BS_T_WhoIsMasterBroadcast, 0)
}

func (srv *Server) deliverToSession(senderAddr *net.UDPAddr, incPkt incomingPacket) {
	sessionID := incPkt.Packet.Header.SessionID
	pktType := incPkt.Packet.Header.PktType

	var ses SessionHandler

	switch pktType {
	case packet.PKT_T_WhoIsMaster:
		ses = srv.handleWhoIsMaster(sessionID)

	case packet.PKT_T_IAmMaster:
		ses = srv.handleIAmMaster(incPkt)

	case packet.PKT_T_SyncComplete:
		ses = srv.handleSyncComplete(senderAddr, incPkt)

	default:
		if packet.IsBroadcastPkt(pktType) && srv.isSynced == false {
			fmt.Println(srv.ID, "is not synced so it can take no new updates") // TODO db
			return
		}

		ses = srv.getOrCreateSession(senderAddr, sessionID)
	}

	srv.printIncMsg(senderAddr, pktType, incPkt) // TODO DB

	ses.ReceivePacket(incPkt.Packet)
}

func (srv *Server) handleWhoIsMaster(sessionID uint32) SessionHandler {
	srv.mu.Lock()
	if srv.searchingForMaster {
		srv.mu.Unlock()
		return nil
	}
	srv.mu.Unlock()
	return srv.getOrCreateWhoIsMasterSession(sessionID)
}

func (srv *Server) handleIAmMaster(incPkt incomingPacket) SessionHandler {
	srv.mu.Lock()

	sessionID := incPkt.Packet.Header.SessionID
	peerID := incPkt.Packet.Header.SenderAddr
	peer, exists := srv.peers[peerID]

	if !exists || peer == nil {
		srv.mu.Unlock()
		return nil
	}

	oldMstr := srv.getMasterPeerLocked()

	// already know this master
	if oldMstr != nil && oldMstr.Addr.String() == peer.Addr.String() {
		srv.isSynced = true
		srv.mu.Unlock()
		return nil
	}

	// new master
	if oldMstr != nil {
		oldMstr.SetMaster(false)
	}

	peer.SetMaster(true)

	srv.isSynced = false
	peer.SetIsSynced(true)

	srv.mu.Unlock()

	srv.QueueSyncRequest()
	return srv.getOrCreateWhoIsMasterSession(sessionID)
}

func (srv *Server) handleSyncComplete(senderAddr *net.UDPAddr, incPkt incomingPacket) SessionHandler {
	srv.mu.Lock()
	sessionID := incPkt.Packet.Header.SessionID
	recipientAddr := incPkt.Packet.Payload.Addr
	selfAddr := srv.recvConn.LocalAddr().String()

	if selfAddr == recipientAddr {
		srv.isSynced = true
	}

	if peer, exists := srv.peers[recipientAddr]; exists {
		peer.SetIsSynced(true)
	}
	srv.mu.Unlock()

	return srv.getOrCreateSession(senderAddr, sessionID)
}

func (srv *Server) createSession(remoteAddr *net.UDPAddr, sessionID *uint32) *session.Session {
	var id uint32
	if sessionID != nil {
		id = *sessionID
	} else {
		id = srv.generateSessionIDLocked()
	}

	ses := session.NewSession(id, remoteAddr, srv.closeReq, srv)
	fmt.Printf("Server %s: new session: %d\n", srv.ID, id)

	srv.addSession(id, ses)
	ses.Start()
	return ses
}

func (srv *Server) createBroadcastSession(sessionID *uint32, bsType session.BroadcastSessionType, expectedResponses int) SessionHandler {
	// generate unique id
	var id uint32
	if sessionID != nil {
		id = *sessionID
	} else {
		id = srv.generateSessionIDLocked()
	}

	var bs SessionHandler

	switch bsType {
	case session.BS_T_StateBroadcast:
		bs = session.NewStateBroadcast(
			id,
			srv.recvConn.LocalAddr().String(),
			srv.broadcastAddr,
			srv.closeReq,
			srv,
			expectedResponses,
		)

	case session.BS_T_WhoIsMasterBroadcast:
		bs = session.NewWhoIsMasterBroadcast(
			id,
			srv.recvConn.LocalAddr().String(),
			srv.broadcastAddr,
			srv.closeReq,
			srv,
		)
	}

	srv.addSession(id, bs)
	bs.Start()

	return bs
}

func (srv *Server) closeSession(sessionID uint32) {
	srv.mu.Lock()
	defer srv.mu.Unlock()

	ses, exists := srv.sessions[sessionID]
	if exists {
		ses.Close()
		delete(srv.sessions, sessionID)

		// TODO remove db
		fmt.Printf("Server %s removed session: %d\n", srv.ID, sessionID)

	}
}

func (srv *Server) PrintSessions() {
	srv.mu.Lock()
	defer srv.mu.Unlock()

	fmt.Printf("%s, Active sessions (%d):\n", srv.ID, len(srv.sessions))

	for id := range srv.sessions {
		fmt.Println(" -", id)
	}
}

func (srv *Server) printIncMsg(senderAddr *net.UDPAddr, pktType packet.PacketType, incPkt incomingPacket) {
	if pktType == packet.PKT_T_MasterAck && !srv.isMaster {

	} else {
		fmt.Printf(
			`%s, Session %d:
	sent from : %s
	to        : %s
	reply sock: %s
	pktType   : %s
	payload   : %+v
`,
			srv.ID,
			incPkt.Packet.Header.SessionID,
			incPkt.Addr.String(),
			incPkt.Packet.Header.RecipientAddr,
			senderAddr.String(),
			pktType,
			incPkt.Packet.Payload,
		)
	}
}
