package server

import (
	"elevator_program/udp"
	"elevator_program/udp/packet"
	"elevator_program/udp/session"
	"net"
	"sync"
)

type SessionHandler interface {
	ReceivePacket(packet.Packet)
	Start()
	Close()
}

type Server struct {
	ID            string
	incPktCh      chan incomingPacket
	outgoingMsgCh chan outgoingMessage
	recvConn      *net.UDPConn
	sendConn      *net.UDPConn
	broadcastConn *net.UDPConn
	sessions      map[uint32]SessionHandler
	mu            sync.Mutex
	closeReq      chan uint32

	stop      chan struct{}
	wg        sync.WaitGroup
	closeOnce sync.Once

	activePeers int
	elevator    chan<- session.ElevatorPacket
}

func NewServer(ip string, port int, id string, numberOfPeers int, toElevator chan<- session.ElevatorPacket) (*Server, error) {
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
	bcConn, err := newReusableListenUDPConn(udp.BROADCAST_PORT)

	sendConn, err := net.ListenUDP("udp", sendAddr)
	if err != nil {
		return nil, err
	}

	srv := &Server{
		ID:            id,
		incPktCh:      make(chan incomingPacket),
		outgoingMsgCh: make(chan outgoingMessage),
		recvConn:      recvConn,
		sendConn:      sendConn,
		broadcastConn: bcConn,
		sessions:      make(map[uint32]SessionHandler),
		closeReq:      make(chan uint32),
		stop:          make(chan struct{}),
		activePeers:   numberOfPeers, // excluding oneself
		elevator:      toElevator,
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
	defer srv.wg.Done()
	for {
		select {
		case <-srv.stop:
			return

		case id := <-srv.closeReq:
			srv.closeSession(id)

		case incPkt := <-srv.incPktCh:
			srv.routeToSession(incPkt)
		case outPkt := <-srv.outgoingMsgCh:
			go srv.dispatchMessage(outPkt)
		}
	}
}

func (srv *Server) UpdateActivePeers(delta int) {
	srv.mu.Lock()
	defer srv.mu.Unlock()

	srv.activePeers += delta
	if srv.activePeers < 0 {
		srv.activePeers = 0
	}
}

func (srv *Server) Close() {
	srv.closeOnce.Do(func() {
		close(srv.stop)      // signal shutdown
		srv.recvConn.Close() // unblock ReadFromUDP
		srv.sendConn.Close()
		srv.broadcastConn.Close()
		srv.wg.Wait() // wait for goroutines

		srv.mu.Lock()
		for sesID := range srv.sessions {
			srv.closeSessionLocked(sesID)
		}
		srv.mu.Unlock()
	})
}
