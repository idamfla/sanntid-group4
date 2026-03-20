package server

import (
	"crypto/rand"
	"encoding/binary"
)

func generateSessionID() (uint32, error) {
	var b [4]byte
	_, err := rand.Read(b[:])
	if err != nil {
		return 0, err
	}
	return binary.LittleEndian.Uint32(b[:]), nil
}

func (srv *Server) generateSessionIDLocked() uint32 {
	for {
		id, err := generateSessionID()
		if err != nil {
			continue
		}

		if _, exists := srv.sessions[id]; !exists {
			return id
		}
	}

}
