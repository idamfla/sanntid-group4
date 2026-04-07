package server

import (
	"context"
	"elevator_program/udp"
	"fmt"
	"net"
	"sync"
	"syscall"
)

type ServerNetwork struct {
	recvConn      *net.UDPConn
	sendConn      *net.UDPConn
	broadcastConn *net.UDPConn // Listening conn
	broadcastAddr *net.UDPAddr // Broadcast sending addr
	closeOnce     sync.Once
}

func NewServerNetwork(addr *net.UDPAddr) (*ServerNetwork, error) {
	recvConn, err := net.ListenUDP("udp", addr)
	if err != nil {
		return nil, err
	}

	// create a local UDP socket for sending (unconnected)
	sendAddr := &net.UDPAddr{
		IP:   net.ParseIP("0.0.0.0"), // binds to any local IP
		Port: 0,                      // 0 = let OS pick a free port
	}

	sendConn, err := net.ListenUDP("udp", sendAddr)
	if err != nil {
		return nil, err
	}

	// create broadcast-listening UDP socket
	bcConn, err := newReusableListenUDPConn(udp.BROADCAST_PORT)
	if err != nil {
		return nil, err
	}

	bcAddr := &net.UDPAddr{
		IP:   net.ParseIP(udp.BroadcastIP),
		Port: udp.BROADCAST_PORT,
	}

	return &ServerNetwork{
		recvConn:      recvConn,
		sendConn:      sendConn,
		broadcastConn: bcConn,
		broadcastAddr: bcAddr,
	}, nil
}

func (sn *ServerNetwork) Close() {
	sn.closeOnce.Do(func() {
		sn.recvConn.Close() // unblock ReadFromUDP
		sn.sendConn.Close()
		sn.broadcastConn.Close()
	})
}

func (sn *ServerNetwork) GetRecvConn() *net.UDPConn {
	return sn.recvConn
}

func (sn *ServerNetwork) GetSendConn() *net.UDPConn {
	return sn.sendConn
}

func (sn *ServerNetwork) GetBroadcastConn() *net.UDPConn {
	return sn.broadcastConn
}

func (sn *ServerNetwork) GetBroadcastAddr() *net.UDPAddr {
	return sn.broadcastAddr
}

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
