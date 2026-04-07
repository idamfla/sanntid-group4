package server

import (
	"fmt"
	"net"
)

func (srv *Server) getOrCreateSession(senderAddr *net.UDPAddr, sesID *uint32) SessionHandler {
	return srv.sessions.GetOrCreateSession(srv, senderAddr, sesID)
}

func (srv *Server) getOrCreateMasterElectionSession() SessionHandler {
	return srv.sessions.GetOrCreateMasterElectionSession(srv)
}

func (srv *Server) createBroadcastSession(expectedResponses int) SessionHandler {
	return srv.sessions.CreateBroadcastSession(srv, expectedResponses)
}

func (srv *Server) getSession(sesID uint32) (SessionHandler, bool) {
	return srv.sessions.GetSession(sesID)
}

func (srv *Server) closeSession(sesID uint32) {
	srv.sessions.closeSession(sesID)
	fmt.Printf("Server %s, closed session: %d\n", srv.GetAlias(), sesID)
}

func (srv *Server) PrintSessions() {
	fmt.Printf("%s, Active sessions (%d):\n", srv.GetAlias(), srv.sessions.CountSessions())

	srv.sessions.PrintSessions()
}
