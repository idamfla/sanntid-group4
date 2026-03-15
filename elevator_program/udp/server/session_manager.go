package server

import (
	"crypto/rand"
	"elevator_program/udp/session"
	"encoding/binary"
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

func generateSessionID() uint32 {
	var b [4]byte
	if _, err := rand.Read(b[:]); err != nil {
		panic(err) // TODO i dont want panic, just something that makes it try again
	}
	return binary.LittleEndian.Uint32(b[:])
}

// generates a unique session id,the called must mutex lock srv
func (srv *Server) generateSessionIDLocked() uint32 {
	var id uint32
	for {
		id = generateSessionID()

		if _, exists := srv.sessions[id]; !exists {
			break
		}
	}

	return id
}

func (srv *Server) createSession(remoteAddr *net.UDPAddr, sessionID *uint32) *session.Session {
	srv.mu.Lock()
	defer srv.mu.Unlock()

	var id uint32
	if sessionID != nil {
		id = *sessionID
	} else {
		id = srv.generateSessionIDLocked()
	}

	ses := session.NewSession(id, remoteAddr, srv.closeReq, srv.elevator, srv)
	srv.sessions[ses.ID] = ses
	ses.Start()
	return ses
}

func (srv *Server) createBroadcastSession(remoteAddr *net.UDPAddr, expectedResponses int) *session.BroadcastSession {
	srv.mu.Lock()
	defer srv.mu.Unlock()

	// generate unique id
	id := srv.generateSessionIDLocked()

	bs := session.NewBroadcastSession(
		id,
		remoteAddr,
		srv.closeReq,
		srv.elevator,
		srv,
		expectedResponses,
	)

	// store it in sessions map (so server tracks it)
	srv.sessions[bs.Session.ID] = bs
	bs.Start()

	return bs
}
