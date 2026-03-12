package server

import (
	"elevator_program/udp/session"
	"net"
	"sync"
)

const (
	// Group4IP        = "10.100.23.15"
	NtnuBroadcastIP = "10.100.23.255"
	HomeBroadcastIP = "192.168.50.255"
	BroadcastPort   = 3000
)

type Server struct {
	ID            string
	recvConn      *net.UDPConn
	sendConn      *net.UDPConn
	broadcastConn *net.UDPConn
	sessions      map[uint32]*session.Session
	mu            sync.Mutex
	closeReq      chan uint32

	stopListening chan struct{}
	wg            sync.WaitGroup

	elevator chan<- session.ElevatorPacket
}

func NewServer(ip string, port int, id string, toElevator chan<- session.ElevatorPacket) (*Server, error) {
	addr := net.UDPAddr{
		IP:   net.ParseIP(ip), // parse the string IP
		Port: port,
	}

	// make sockets
	recvConn, err := net.ListenUDP("udp", &addr)
	if err != nil {
		return nil, err
	}

	// create a local UDP socket for sending (unconnected)
	sendAddr := &net.UDPAddr{
		IP:   net.ParseIP("0.0.0.0"), // binds to any local IP
		Port: 0,                      // 0 = let OS pick a free port
		// Port: port - 10, // TODO magic number, just testing sending from a choosen port
	}

	// create broadcast-listening UDP socket
	bcConn, err := newReusableListenUDPConn(BroadcastPort)

	sendConn, err := net.ListenUDP("udp", sendAddr)
	if err != nil {
		return nil, err
	}

	srv := &Server{
		ID:            id,
		recvConn:      recvConn,
		sendConn:      sendConn,
		broadcastConn: bcConn,
		sessions:      make(map[uint32]*session.Session),
		closeReq:      make(chan uint32),
		stopListening: make(chan struct{}),
		elevator:      toElevator,
	}

	return srv, nil
}

// TODO freezes if you close when there are sessions that are half-closed? might be that it just takes time ...
func (srv *Server) Close() {
	close(srv.stopListening) // signal shutdown
	srv.recvConn.Close()     // unblock ReadFromUDP
	srv.sendConn.Close()
	srv.broadcastConn.Close()
	srv.wg.Wait() // wait for goroutines

	srv.mu.Lock()
	for sesID := range srv.sessions {
		srv.closeSessionLocked(sesID)
	}
	srv.mu.Unlock()

	close(srv.closeReq)
}
