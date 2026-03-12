package server

import (
	"crypto/rand"
	"elevator_program/udp/session"
	"encoding/binary"
	"fmt"
	"net"
)

func (srv *Server) registerNewSession(id uint32, addr *net.UDPAddr) *session.Session {
	ses := session.NewSession(id, addr, srv.closeReq, srv.elevator, srv)
	srv.sessions[id] = ses
	ses.Start()
	return ses
}

// returns session from session map. if no hits, a new server is made
func (srv *Server) getOrCreateSession(senderAddr *net.UDPAddr, sessionID uint32) *session.Session {
	srv.mu.Lock()
	defer srv.mu.Unlock()

	if ses, exists := srv.sessions[sessionID]; exists {
		return ses
	}

	return srv.registerNewSession(sessionID, senderAddr)
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

// region Session ID
func generateSessionID() uint32 {
	var b [4]byte
	if _, err := rand.Read(b[:]); err != nil {
		panic(err) // extremely unlikely but correct
	}
	return binary.LittleEndian.Uint32(b[:])
}

func (srv *Server) createSession(remoteAddr *net.UDPAddr) *session.Session {
	srv.mu.Lock()
	defer srv.mu.Unlock()

	// generate unique id
	var id uint32
	for {
		id = generateSessionID()

		if _, exists := srv.sessions[id]; !exists {
			break
		}
	}

	return srv.registerNewSession(id, remoteAddr)
}

// endregion
