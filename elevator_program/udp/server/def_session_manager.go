package server

import (
	"crypto/rand"
	"elevator_program/udp/session"
	"encoding/binary"
	"fmt"
	"net"
	"sync"
)

type SessionManager struct {
	sessions map[uint32]SessionHandler
	mu       sync.Mutex
}

func NewSessionManager() *SessionManager {
	return &SessionManager{sessions: make(map[uint32]SessionHandler)}
}

func (sm *SessionManager) Close() {
	sm.lock()
	ids := make([]uint32, 0, len(sm.sessions))

	for id := range sm.sessions {
		ids = append(ids, id)
	}

	sm.unlock()

	for _, id := range ids {
		sm.closeSession(id)
	}
}

func (sm *SessionManager) createSession(srv *Server, remoteAddr *net.UDPAddr, sesID *uint32) *session.Session {
	sm.lock()

	if srv.isLocalAddr(remoteAddr) {
		err := fmt.Errorf("Tried to send to oneself %s", remoteAddr.String())
		fmt.Println(err)

		sm.unlock()
		return nil
	}

	var id uint32
	if sesID != nil {
		id = *sesID
	} else {
		id = sm.generateSessionIDUnsafe()
	}

	ses := session.NewSession(id, remoteAddr, srv)
	sm.addSession(ses)
	fmt.Printf("Server %s: new session: %d\n", srv.GetAlias(), id)

	sm.unlock()

	ses.Start()
	return ses
}

func (srv *Server) isLocalAddr(addr *net.UDPAddr) bool {
	local := srv.getRecvUDPAddr()
	return addr.IP.Equal(local.IP) && addr.Port == local.Port
}

func (sm *SessionManager) createBroadcastSession(srv *Server, expectedResponses int) SessionHandler {
	sm.lock()

	id := sm.generateSessionIDUnsafe()
	sbs := session.NewStateBroadcast(id, srv, expectedResponses)
	sm.addSession(sbs)
	fmt.Printf("Server %s: new StateBroadcast session: %d\n", srv.GetAlias(), id)

	sm.unlock()

	sbs.Start()

	return sbs
}
func (sm *SessionManager) createMasterElectionSession(srv *Server) SessionHandler {
	sm.lock()

	id := MASTER_ELECTION_SESSSION_ID
	ws := session.NewWhoIsAliveBroadcast(id, srv)
	sm.addSession(ws)
	fmt.Printf("Server %s: new WhoIsAlive session: %d\n", srv.GetAlias(), id)

	sm.unlock()

	ws.Start()

	return ws
}

func (sm *SessionManager) getSession(sesID uint32) (SessionHandler, bool) {
	sm.mu.Lock()
	defer sm.mu.Unlock()
	sh, exists := sm.sessions[sesID]
	return sh, exists
}

func (sm *SessionManager) closeSession(sesID uint32) {
	sm.lock()

	ses, exists := sm.sessions[sesID]

	if exists {
		delete(sm.sessions, sesID)
	}

	sm.unlock()

	if exists {
		ses.Close()
	}
}

// --- unsafe ---

func (sm *SessionManager) addSession(sh SessionHandler) {
	sm.sessions[sh.GetID()] = sh
}

// generate unique id
func (sm *SessionManager) generateSessionIDUnsafe() uint32 {
	for {
		id, err := generateID()
		if err != nil {
			continue
		}

		if _, exists := sm.sessions[id]; !exists {
			return id
		}
	}

}

func generateID() (uint32, error) {
	var b [4]byte
	_, err := rand.Read(b[:])
	if err != nil {
		return 0, err
	}
	return binary.LittleEndian.Uint32(b[:]), nil
}

// --- helpers ---

func (sm *SessionManager) countSessions() int {
	sm.lock()
	defer sm.unlock()
	return len(sm.sessions)
}

func (sm *SessionManager) printSessions() {
	sm.lock()
	defer sm.unlock()

	for id := range sm.sessions {
		fmt.Println(" -", id)
	}
}

func (sm *SessionManager) lock()   { sm.mu.Lock() }
func (sm *SessionManager) unlock() { sm.mu.Unlock() }
