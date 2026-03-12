package server

import (
	"crypto/rand"
	"elevator_program/udp/session"
	"encoding/binary"
	"fmt"
	"net"
)

func generateSessionID() uint32 {
	var b [4]byte
	if _, err := rand.Read(b[:]); err != nil {
		panic(err) // extremely unlikely but correct
	}
	return binary.LittleEndian.Uint32(b[:])
}

// helper function, not called directly: *unsafe*
func (srv *Server) newSessionIDLocked() uint32 {
	for {
		id := generateSessionID()

		if _, exists := srv.sessions[id]; !exists {
			return id
		}
	}
}

// retrieves the session with the correct session id from the servers sessionMap and returns it. if there are no hits, a new session will be created
func (srv *Server) getOrCreateSession(senderAddr *net.UDPAddr, sessionID uint32) *session.Session {
	srv.mu.Lock()
	defer srv.mu.Unlock()

	ses, exists := srv.sessions[sessionID]
	if !exists {
		ses = session.NewSession(sessionID, senderAddr, srv.closeReq, srv.elevator, srv)
		srv.sessions[sessionID] = ses
		ses.Start()
	}

	return ses
}

func (srv *Server) createSession(remoteAddr *net.UDPAddr) *session.Session {
	srv.mu.Lock()
	defer srv.mu.Unlock()

	id := srv.newSessionIDLocked()

	ses := session.NewSession(id, remoteAddr, srv.closeReq, srv.elevator, srv)
	srv.sessions[id] = ses
	ses.Start()

	return ses
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

	fmt.Printf("Active sessions (%d):\n", len(srv.sessions))

	for id := range srv.sessions {
		fmt.Println(" -", id)
	}
}
