package server

import (
	"fmt"
	"net"
)

func (srv *Server) SendSession(sessionID uint32, remoteIP string, remotePort int, message string) error {
	addr := fmt.Sprintf("%s:%d", remoteIP, remotePort)
	remoteAddr, err := net.ResolveUDPAddr("udp", addr)
	if err != nil {
		return fmt.Errorf("resolve UDP addr: %w", err)
	}

	seq := uint32(1) // could be improved per session

	return srv.SendMessage(remoteAddr, seq, sessionID, message)
}
