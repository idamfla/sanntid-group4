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
	quorumPercent   = 50
	percentDivisor  = 100
)

type Server struct {
	ID              string
	incomingPackets chan session.IncomingPacket
	recvConn        *net.UDPConn
	sendConn        *net.UDPConn
	broadcastConn   *net.UDPConn
	sessions        map[uint32]*session.Session
	mu              sync.Mutex
	closeReq        chan uint32

	stopListening chan struct{}
	wg            sync.WaitGroup

	activePeers int
	elevator    chan<- session.ElevatorPacket
}

func NewServer(ip string, port int, id string, numberOfElevators int, toElevator chan<- session.ElevatorPacket) (*Server, error) {
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
	bcConn, err := newReusableListenUDPConn(BroadcastPort)

	sendConn, err := net.ListenUDP("udp", sendAddr)
	if err != nil {
		return nil, err
	}

	srv := &Server{
		ID:              id,
		incomingPackets: make(chan session.IncomingPacket),
		recvConn:        recvConn,
		sendConn:        sendConn,
		broadcastConn:   bcConn,
		sessions:        make(map[uint32]*session.Session),
		closeReq:        make(chan uint32),
		stopListening:   make(chan struct{}),
		activePeers:     numberOfElevators - 1, // excluding oneself
		elevator:        toElevator,
	}

	return srv, nil
}

func (srv *Server) Start() {
	srv.wg.Add(3)
	go srv.readLoop(srv.recvConn)
	go srv.readLoop(srv.broadcastConn)

	go srv.run()
}

func (srv *Server) run() {
	for {
		select {
		case <-srv.stopListening:
			return

		case id := <-srv.closeReq:
			srv.closeSession(id)

		case incPkt := <-srv.incomingPackets:
			srv.routeToSession(incPkt)
		}
	}
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
