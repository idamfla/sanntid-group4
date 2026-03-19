package server

import (
	"context"
	"elevator_program/udp/packet"
	"errors"
	"fmt"
	"net"
	"syscall"
	"golang.org/x/sys/unix"
)

func (srv *Server) readLoop(conn *net.UDPConn) {
	defer srv.wg.Done()
	buf := make([]byte, 2048)

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

		srv.incPktCh <- incomingPacket{
			Packet: pkt,
			Addr:   addr,
		}
	}
}

// Helper
func newReusableListenUDPConn(port int) (*net.UDPConn, error) {
	lc := net.ListenConfig{
		Control: func(network, address string, c syscall.RawConn) error {
			var sockErr error

			err := c.Control(func(fd uintptr) {
				if err := unix.SetsockoptInt(int(fd), unix.SOL_SOCKET, unix.SO_REUSEADDR, 1); err != nil {
					sockErr = err
					return
				}
				if err := unix.SetsockoptInt(int(fd), unix.SOL_SOCKET, unix.SO_REUSEPORT, 1); err != nil {
					sockErr = err
					return
				}
			})
			if err != nil {
				return err
			}
			return sockErr
		},
	}


	pc, err := lc.ListenPacket(context.Background(), "udp4", fmt.Sprintf(":%d", port))
	if err != nil {
		return nil, err
	}

	return pc.(*net.UDPConn), nil
}

func (srv *Server) routeInkPkt(incPkt incomingPacket) {
	senderAddr, err := srv.resolveSenderAddr(incPkt.Packet.Header.SenderAddr)
	if err != nil {
		return
	}

	if incPkt.Packet.Header.PktType == packet.PKT_T_CatchupDone {
		// srv.handleSyncRequest(senderAddr)
		srv.isSynced = true
	}

	srv.registerOrUpdatePeer(senderAddr, false)

	srv.deliverToSession(senderAddr, incPkt)
}

// func (srv *Server) handleSyncRequest(addr *net.UDPAddr) {
// 	srv.registerOrUpdatePeer(addr, true)
// 	fmt.Println("StateSync packet: peer registration and sync only")
// }

func (srv *Server) resolveSenderAddr(replyAddr string) (*net.UDPAddr, error) {
	udpAddr, err := net.ResolveUDPAddr("udp", replyAddr)
	if err != nil {
		fmt.Printf("Invalid reply address %s\n", replyAddr)
		return nil, err
	}

	if replyAddr == srv.recvConn.LocalAddr().String() {
		err := fmt.Errorf("Server is receiving message from itself")
		return nil, err
	}

	return udpAddr, nil
}
