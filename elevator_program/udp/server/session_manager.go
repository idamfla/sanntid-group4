package server

import (
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

func (srv *Server) deliverToSession(senderAddr *net.UDPAddr, incPkt incomingPacket) {
	sessionID := incPkt.Packet.Header.SessionID
	ses := srv.getOrCreateSession(senderAddr, sessionID)

	fmt.Printf(
		`%s, Session %d:
	sent from : %s
	to        : %s
	reply sock: %s
	pktType   : %s
`,
		srv.ID,
		incPkt.Packet.Header.SessionID,
		incPkt.Addr.String(),
		incPkt.Packet.Header.RecipientAddr,
		senderAddr.String(),
		incPkt.Packet.Header.PktType,
	)

	ses.ReceivePacket(incPkt.Packet)
}

// helper function, not called directly: *unsafe*
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

func (srv *Server) createBroadcastSession(remoteAddr *net.UDPAddr, expectedResponses int) *session.BroadcastSession {
	// generate unique id
	id := srv.generateSessionIDLocked()

	bs := session.NewBroadcastSession(
		id,
		remoteAddr,
		srv.closeReq,
		srv,
		expectedResponses,
	)
	fmt.Printf("Server %s: new broadcast session: %d\n", srv.ID, id)

	// store it in sessions map (so server tracks it)
	srv.mu.Lock()
	srv.sessions[bs.Session.ID] = bs
	srv.mu.Unlock()
	bs.Start()

	return bs
}
