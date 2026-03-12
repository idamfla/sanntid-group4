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
	incPktChan := make(chan session.IncomingPacket, 32)

	// UDP reader goroutines
	srv.wg.Add(1)
	go srv.readLoop(srv.recvConn, incPktChan)

	srv.wg.Add(1)
	go srv.readLoop(srv.broadcastConn, incPktChan)

	// Main event loop
	for {
		select {
		case <-srv.stopListening:
			return

		case id := <-srv.closeReq:
			srv.closeSession(id)

		case incPkt := <-incPktChan:
			srv.dispatchToSession(incPkt)
		}
	}
}

func (srv *Server) dispatchToSession(incPkt session.IncomingPacket) {
	sessionID := incPkt.Packet.Header.SessionID
	senderAddr, err := net.ResolveUDPAddr("udp", incPkt.Packet.Header.SenderAddr)
	if err != nil {
		fmt.Printf("Invalid reply address %s\n", incPkt.Packet.Header.SenderAddr)
		return
	}

	ses := srv.getOrCreateSession(senderAddr, sessionID)

	fmt.Printf(
		`%s, Session %d:
	sent from : %s
	to        : %s
	reply sock: %s
	pktType   : %s
`,
		srv.ID,
		incPkt.Packet.Header.SessionID,
		incPkt.Addr.String(),
		incPkt.Packet.Header.RecipientAddr,
		senderAddr.String(),
		incPkt.Packet.Header.PktType,
	)

	ses.RecvCh <- incPkt
}

func (srv *Server) readLoop(conn *net.UDPConn, out chan<- session.IncomingPacket) {
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

		out <- session.IncomingPacket{
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
