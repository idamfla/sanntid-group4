package server

import (
	"context"
	"elevator_program/udp"
	"errors"
	"fmt"
	"net"
	"syscall"
)

func (srv *Server) Listen() {
	packetChan := make(chan udp.IncomingPacket, 32)

	// UDP reader goroutines
	srv.wg.Add(1)
	go srv.readLoop(srv.recvConn, packetChan)

	srv.wg.Add(1)
	go srv.readLoop(srv.broadcastConn, packetChan)

	// Main event loop
	for {
		select {
		case <-srv.stopListening:
			return

		case id := <-srv.closeReq:
			srv.closeSession(id)

		case inc := <-packetChan:
			srv.handleIncoming(inc)
		}
	}
}

func (srv *Server) handleIncoming(incPck udp.IncomingPacket) {
	id := incPck.Packet.Header.SessionID

	srv.mu.Lock()
	ses, exists := srv.sessions[id]

	if !exists {
		ses = udp.NewSession(id, incPck.Addr, srv.closeReq, srv.elevator, srv)
		srv.sessions[id] = ses
	}
	srv.mu.Unlock()

	fmt.Printf(
		`%s, Session %d:
	sent from : %s
	to        : %s
	reply sock: %s
	msgType   : %s
	msg: %+v
`,
		srv.ID,
		incPck.Packet.Header.SessionID,
		incPck.Addr.String(),
		incPck.Packet.Header.RecipientAddr,
		incPck.Packet.Header.SenderAddr,
		incPck.Packet.Header.MsgType,
		incPck.Packet.Payload,
	)

	ses.Incoming <- incPck
}

func (srv *Server) readLoop(conn *net.UDPConn, out chan<- udp.IncomingPacket) {
	defer srv.wg.Done()
	buf := make([]byte, 2048)

	fmt.Println(srv.ID, "listening on", conn.LocalAddr().String())

	for {
		n, addr, err := conn.ReadFromUDP(buf)
		if err != nil {

			// normal shutdown, just exit the loop
			if errors.Is(err, net.ErrClosed) {
				return
			}

			fmt.Println("Read error:", err)
			continue
		}

		pck, err := udp.DecodePacket(buf, n)
		if err != nil {
			fmt.Println("Decode error: ", err)
			continue
		}

		out <- udp.IncomingPacket{
			Packet: pck,
			Addr:   addr,
		}
	}
}

// Helper
func NewReusableListenUDPConn(port int) (*net.UDPConn, error) {
	lc := net.ListenConfig{
		Control: func(network, address string, c syscall.RawConn) error {
			return c.Control(func(fd uintptr) {
				syscall.SetsockoptInt(int(fd), syscall.SOL_SOCKET, syscall.SO_REUSEADDR, 1)
			})
		},
	}

	pc, err := lc.ListenPacket(context.Background(), "udp4", fmt.Sprintf(":%d", port))
	if err != nil {
		return nil, err
	}

	return pc.(*net.UDPConn), nil
}
