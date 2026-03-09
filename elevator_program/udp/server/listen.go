package server

import (
	"context"
	"elevator_program/udp/packet"
	"elevator_program/udp/session"
	"errors"
	"fmt"
	"net"
	"syscall"
)

func (srv *Server) Listen() {
	packetChan := make(chan session.PacketContext, 32)

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

		case pktCtx := <-packetChan:
			srv.dispatchToSession(pktCtx)
		}
	}
}

func (srv *Server) dispatchToSession(pktCtx session.PacketContext) {
	sessionID := pktCtx.Packet.Header.SessionID
	senderAddr, err := net.ResolveUDPAddr("udp", pktCtx.Packet.Header.SenderAddr)
	if err != nil {
		fmt.Printf("Invalid reply address %s\n", pktCtx.Packet.Header.SenderAddr)
		return
	}

	ses := srv.getOrCreateSession(senderAddr, sessionID)

	fmt.Printf(
		`%s, Session %d:
	sent from : %s
	to        : %s
	reply sock: %s
	pktType   : %s
	payload: %+v
`,
		srv.ID,
		pktCtx.Packet.Header.SessionID,
		pktCtx.Addr.String(),
		pktCtx.Packet.Header.RecipientAddr,
		senderAddr.String(),
		pktCtx.Packet.Header.PktType,
		pktCtx.Packet.Payload,
	)

	ses.IncomingCh <- pktCtx
}

// retrieves the session with the correct session id from the servers sessionMap and returns it. if there are no hits, a new session will be created
func (srv *Server) getOrCreateSession(senderAddr *net.UDPAddr, sessionID uint32) *session.Session {
	srv.mu.Lock()
	ses, exists := srv.sessions[sessionID]

	if !exists {
		ses = session.NewSession(sessionID, senderAddr, srv.closeReq, srv.elevator, srv)
		srv.sessions[sessionID] = ses
	}
	srv.mu.Unlock()
	return ses
}

func (srv *Server) readLoop(conn *net.UDPConn, out chan<- session.PacketContext) {
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

		pkt, err := packet.DecodePacket(buf, n)
		if err != nil {
			fmt.Println("Decode error: ", err)
			continue
		}

		out <- session.PacketContext{
			Packet: pkt,
			Addr:   addr,
		}
	}
}

// Helper
func newReusableListenUDPConn(port int) (*net.UDPConn, error) {
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
