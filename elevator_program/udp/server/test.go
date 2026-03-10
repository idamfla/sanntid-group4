package server

import (
	"elevator_program/udp/udp_message"
	"fmt"
	"net"
)

func (srv *Server) SendSession(sessionID uint32, remoteIP string, remotePort int, msg udp_message.Message) error {
	addr := fmt.Sprintf("%s:%d", remoteIP, remotePort)
	remoteAddr, err := net.ResolveUDPAddr("udp", addr)
	if err != nil {
		return fmt.Errorf("resolve UDP addr: %w", err)
	}

	seq := uint32(1) // could be incremented per session

	return srv.SendMessage(remoteAddr, seq, sessionID, msg)
}
