package server

import (
	"elevator_program/udp"
	"elevator_program/udp/packet"
	"elevator_program/udp/peer_info"
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
	broadcastAddr *net.UDPAddr
	sessions      map[uint32]SessionHandler
	peers         map[string]*peer_info.PeerInfo
	bcSeq         uint32
	mu            sync.Mutex
	closeReq      chan uint32

	stop      chan struct{}
	wg        sync.WaitGroup
	closeOnce sync.Once

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
	}

	sendConn, err := net.ListenUDP("udp", sendAddr)
	if err != nil {
		return nil, err
	}

	// create broadcast-listening UDP socket
	bcConn, err := newReusableListenUDPConn(udp.BROADCAST_PORT)

	bcAddr := &net.UDPAddr{
		// IP: net.ParseIP("127.0.0.1"),
		IP:   net.ParseIP(udp.BroadcastIP),
		Port: udp.BROADCAST_PORT,
	}

	srv := &Server{
		ID:            id,
		incPktCh:      make(chan incomingPacket),
		outgoingMsgCh: make(chan outgoingMessage),
		recvConn:      recvConn,
		sendConn:      sendConn,
		broadcastConn: bcConn,
		broadcastAddr: bcAddr,
		sessions:      make(map[uint32]SessionHandler),
		peers:         make(map[string]*peer_info.PeerInfo),
		// bcSeq: 0,
		closeReq: make(chan uint32),
		stop:     make(chan struct{}),
		elevator: toElevator,
	}

	// TODO add initial msg
	// srv.QueueMessage(
	// 	nil,
	// 	packet.PROTO_PKT_T_StateSync,
	// 	message.Message{Content: "I am new"},
	// )

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
		case outMsg := <-srv.outgoingMsgCh:
			srv.wg.Add(1)
			go srv.dispatchMessage(outMsg)
		}
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

func (srv *Server) getBroadcastSeq() uint32 {
	srv.mu.Lock()
	defer srv.mu.Unlock()
	return srv.bcSeq
}
