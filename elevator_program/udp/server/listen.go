package server

import (
	"context"
	"elevator_program/udp/packet"
	"errors"
	"fmt"
	"net"
	"syscall"
)

func (srv *Server) readLoop(conn *net.UDPConn) {
	defer srv.wg.Done()
	buf := make([]byte, 2048)

	for {
		pkt, addr, continueLoop := srv.readPacket(conn, buf)
		if !continueLoop {
			return
		}
		if addr == nil {
			continue
		}

		srv.receivePacket(pkt, addr)

	}
}

func (srv *Server) readPacket(conn *net.UDPConn, buf []byte) (pkt packet.Packet, addr *net.UDPAddr, continueLoop bool) {
	n, addr, err := conn.ReadFromUDP(buf)
	if err != nil {
		// normal shutdown, just exit the loop
		if errors.Is(err, net.ErrClosed) {
			return packet.Packet{}, nil, false
		}

		fmt.Println("Read error:", err)
		return packet.Packet{}, nil, true
	}

	pkt, err = packet.DecodePacket(buf, n)
	if err != nil {
		fmt.Println("Decode error: ", err)
		return packet.Packet{}, nil, true
	}

	return pkt, addr, true
}

func (srv *Server) receivePacket(pkt packet.Packet, addr *net.UDPAddr) {
	select {
	case <-srv.stop:
		return
	case srv.incPktCh <- pkt:
	default:
		fmt.Println("Server mailbox is full, could not receive new packet")
	}
}

func (srv *Server) handleIncPkt(pkt packet.Packet) { // TODO rename handleIncPkt, should this be with session?
	senderAddr, err := srv.resolveSenderAddr(pkt.Header.SenderAddr)
	if err != nil {
		return
	}

	srv.updatePeer(senderAddr, pkt.Header.Origin)

	srv.deliverToSession(senderAddr, pkt)
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

func (srv *Server) resolveSenderAddr(replyAddr string) (*net.UDPAddr, error) {
	udpAddr, err := net.ResolveUDPAddr("udp", replyAddr)
	if err != nil {
		fmt.Printf("Invalid reply address %s\n", replyAddr)
		return nil, err
	}

	if replyAddr == srv.GetRecvString() {
		err := fmt.Errorf("Server is receiving message from itself")
		return nil, err
	}

	return udpAddr, nil
}
