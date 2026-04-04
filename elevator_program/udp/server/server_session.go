package server

import (
	"elevator_program/udp/session"
	"fmt"
	"net"
)

func (srv *Server) getOrCreateSession(senderAddr *net.UDPAddr, sesID *uint32) SessionHandler {
	if sesID != nil {
		if *sesID == MASTER_ELECTION_SESSSION_ID {
			return srv.getOrCreateMasterElectionSession()
		}

		ses, exists := srv.getSession(*sesID)
		if exists {
			return ses
		}
	}

	return srv.createSession(senderAddr, sesID)
}

func (srv *Server) getOrCreateMasterElectionSession() SessionHandler {
	bs, exists := srv.getSession(MASTER_ELECTION_SESSSION_ID)
	if exists {
		return bs
	}

	return srv.createMasterElectionSession()
}

func (srv *Server) createSession(senderAddr *net.UDPAddr, sesID *uint32) *session.Session {
	return srv.sessions.createSession(srv, senderAddr, sesID)
}

func (srv *Server) createMasterElectionSession() SessionHandler {
	return srv.sessions.createMasterElectionSession(srv)
}

func (srv *Server) createBroadcastSession(expectedResponses int) SessionHandler {
	return srv.sessions.createBroadcastSession(srv, expectedResponses)
}

func (srv *Server) getSession(sesID uint32) (SessionHandler, bool) {
	return srv.sessions.getSession(sesID)
}

func (srv *Server) closeSession(sesID uint32) {
	srv.sessions.closeSession(sesID)
	fmt.Printf("Server %s, closed session: %d\n", srv.ID, sesID)
}

func (srv *Server) PrintSessions() {
	fmt.Printf("%s, Active sessions (%d):\n", srv.ID, srv.sessions.countSessions())

	srv.sessions.printSessions()
}
