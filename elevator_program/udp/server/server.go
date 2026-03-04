package server

import (
	"elevator_program/udp"
	"fmt"
	"net"
	"sync"
)

const (
	NtnuBroadcastIP = "10.100.23.255"
	Group4IP        = "10.100.23.15"
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
}

func NewServer(ip string, port int, id string) (*Server, error) {
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
	// TODO for running test when only using one pc
	idOffset := int(int(id[0] - 'a')) // 'a' -> 0, 'b' -> 1, 'c' -> 2
	// listen on broadcast

	bcConn, bcErr := net.ListenUDP("udp4", &net.UDPAddr{
		IP:   net.IPv4zero,             // listen on all interfaces
		Port: BroadcastPort + idOffset, // the port you want to receive on
	})
	if bcErr != nil {
		return nil, err
	}

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
