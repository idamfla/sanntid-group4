package server

import (
	"elevator_program/udp"
	"fmt"
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
	sessions      map[uint32]*udp.Session
	mu            sync.Mutex
	closeReq      chan uint32

	stopListening chan struct{}
	wg            sync.WaitGroup

	elevator chan<- udp.ElevatorMessage
}

func NewServer(ip string, port int, id string, toElevator chan<- udp.ElevatorMessage) (*Server, error) {
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
	}

	// create broadcast-listening UDP socket
	bcConn, err := NewReusableListenUDPConn(BroadcastPort)

	sendConn, err := net.ListenUDP("udp", sendAddr)
	if err != nil {
		return nil, err
	}

	srv := &Server{
		ID:            id,
		recvConn:      recvConn,
		sendConn:      sendConn,
		broadcastConn: bcConn,
		sessions:      make(map[uint32]*udp.Session),
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
	for id := range srv.sessions {
		srv.closeSessionLocked(id)
	}
	srv.mu.Unlock()

	close(srv.closeReq)
}

// helper function, not called directly: *unsafe*
func (srv *Server) closeSessionLocked(sesID uint32) {
	ses, exists := srv.sessions[sesID]
	if exists {
		ses.Close()
		delete(srv.sessions, sesID)

		// TODO remove db
		fmt.Printf("Server %s removed session: %d\n", srv.ID, sesID)

	}
}

func (srv *Server) closeSession(sesID uint32) {
	srv.mu.Lock()
	defer srv.mu.Unlock()
	srv.closeSessionLocked(sesID)
}

func (srv *Server) PrintSessions() {
	srv.mu.Lock()
	defer srv.mu.Unlock()

	fmt.Printf("Active sessions (%d):\n", len(srv.sessions))
	// for id := range srv.sessions {
	// 	fmt.Println(" -", id)
	// }
	fmt.Printf("Server %s closed\n", srv.ID)
}
