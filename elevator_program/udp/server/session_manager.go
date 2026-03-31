package server

import (
	"elevator_program/udp/packet"
	"elevator_program/udp/session"
	"fmt"
	"net"
)

// returns session from session map. if no hits, a new server is made
func (srv *Server) getOrCreateSession(senderAddr *net.UDPAddr, sessionID uint32) SessionHandler {
	srv.mu.Lock()
	ses, exists := srv.sessions[sessionID]
	srv.mu.Unlock()

	if exists {
		return ses
	}

	return srv.createSession(senderAddr, &sessionID)
}

func (srv *Server) getOrCreateWhoIsMasterSession(sessionID uint32) SessionHandler {
	srv.mu.Lock()
	bs, exists := srv.sessions[sessionID]
	srv.mu.Unlock()

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
		srv.mu.Lock()
		if srv.searchingForMaster {
			srv.mu.Unlock()
			return
		}
		srv.mu.Unlock()
		ses = srv.getOrCreateWhoIsMasterSession(sessionID)

	case packet.PKT_T_IAmMaster:
		srv.mu.Lock()

		key := incPkt.Packet.Header.SenderAddr
		peer, exists := srv.peers[key]

		if !exists || peer == nil {
			srv.mu.Unlock()
			return
		}

		oldMstr := srv.getMasterPeerLocked()

		// already know this master
		if oldMstr != nil && oldMstr.Addr.String() == peer.Addr.String() {
			srv.isSynced = true
			srv.mu.Unlock()
			return
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
		ses = srv.getOrCreateWhoIsMasterSession(sessionID)

		// if !srv.IsMaster() {
		// 	fmt.Println(srv.ID, "from", incPkt.Packet.Header.SenderAddr, incPkt.Packet.Header.PktType, srv.ID, "is not master") // TODO db remove later
		// }

	default:
		if pktType == packet.PKT_T_SyncComplete {
			srv.mu.Lock()
			recipientAddr := incPkt.Packet.Payload.Addr
			selfAddr := srv.recvConn.LocalAddr().String()

			if selfAddr == recipientAddr {
				srv.isSynced = true
			}

			if peer, exists := srv.peers[recipientAddr]; exists {
				peer.SetIsSynced(true)
			}
			srv.mu.Unlock()
			if packet.IsBroadcastPkt(pktType) && srv.isSynced == false {
				fmt.Println(srv.ID, "is not synced so it can take no new updates") // TODO db
				return
			}
		}

		ses = srv.getOrCreateSession(senderAddr, sessionID)
	}

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

	ses.ReceivePacket(incPkt.Packet)
}

// helper function, not called directly: *unsafe*
func (srv *Server) createSession(remoteAddr *net.UDPAddr, sessionID *uint32) *session.Session {
	var id uint32
	if sessionID != nil {
		id = *sessionID
	} else {
		id = srv.generateSessionIDLocked()
	}

	ses := session.NewSession(id, remoteAddr, srv.closeReq, srv)
	fmt.Printf("Server %s: new session: %d\n", srv.ID, id)

	srv.mu.Lock()
	srv.sessions[ses.ID] = ses
	srv.mu.Unlock()
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

	srv.mu.Lock()
	srv.sessions[id] = bs
	srv.mu.Unlock()
	bs.Start()

	return bs
}

func (srv *Server) closeSessionLocked(sesID uint32) {
	ses, exists := srv.sessions[sesID]
	if exists {
		ses.Close()
		delete(srv.sessions, sesID)

		// TODO remove db
		fmt.Printf("Server %s removed session: %d\n", srv.ID, sesID)

	}
}

func (srv *Server) closeSession(sesID uint32) {
	srv.mu.Lock()
	defer srv.mu.Unlock()
	srv.closeSessionLocked(sesID)
}

func (srv *Server) PrintSessions() {
	srv.mu.Lock()
	defer srv.mu.Unlock()

	fmt.Printf("%s, Active sessions (%d):\n", srv.ID, len(srv.sessions))

	for id := range srv.sessions {
		fmt.Println(" -", id)
	}
}
