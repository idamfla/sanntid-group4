package server

import (
	"crypto/rand"
	"encoding/binary"
)

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
