package server

import (
	"elevator_program/message"
	"elevator_program/udp"
	"elevator_program/udp/packet"
	"elevator_program/udp/peerinfo"
	"elevator_program/udp/session"
	"fmt"
	"net"
	"sync"
)

const (
	CHANNEL_BUF = 32
)

type SessionHandler interface {
	Start()
	Close()
	SendReply(pkt packet.PacketType)
	ReceivePacket(pkt packet.Packet)
	QueueWhoIsMasterMsg()
	QueueBroadcastUpdateMsg(pktType packet.PacketType, eMsg message.ElevatorMessage)
}

type Server struct {
	ID                 string
	isMaster           bool
	searchingForMaster bool
	isSynced           bool
	incPktCh           chan incomingPacket
	outgoingMsgCh      chan outgoingMessage
	recvConn           *net.UDPConn
	sendConn           *net.UDPConn
	broadcastConn      *net.UDPConn // Listening conn
	broadcastAddr      *net.UDPAddr // Broadcast sending addr
	sessions           map[uint32]SessionHandler
	peers              map[string]*peerinfo.PeerInfo
	bcSeq              uint32
	mu                 sync.Mutex
	closeReq           chan uint32

	stop      chan struct{}
	wg        sync.WaitGroup
	closeOnce sync.Once

	elevator          chan session.ElevatorPacket
	elevatorTaskQueue chan ElevatorTask
}

func NewServer(ip string, port int, id string, toElevator chan session.ElevatorPacket) (*Server, error) { // TODO isMaster is default false, set by election or something
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
	if err != nil {
		return nil, err
	}

	bcAddr := &net.UDPAddr{
		// IP: net.ParseIP("127.0.0.1"),
		IP:   net.ParseIP(udp.BroadcastIP),
		Port: udp.BROADCAST_PORT,
	}

	srv := &Server{
		ID:                 id,
		isMaster:           false,
		searchingForMaster: false,
		isSynced:           true,
		incPktCh:           make(chan incomingPacket, CHANNEL_BUF),
		outgoingMsgCh:      make(chan outgoingMessage, CHANNEL_BUF),
		recvConn:           recvConn,
		sendConn:           sendConn,
		broadcastConn:      bcConn,
		broadcastAddr:      bcAddr,
		sessions:           make(map[uint32]SessionHandler),
		peers:              make(map[string]*peerinfo.PeerInfo),
		closeReq:           make(chan uint32, CHANNEL_BUF),
		stop:               make(chan struct{}, CHANNEL_BUF),
		elevator:           toElevator,
		elevatorTaskQueue:  make(chan ElevatorTask, CHANNEL_BUF),
	}

	return srv, nil
}

func (srv *Server) Start() {
	srv.wg.Add(4)
	go srv.readLoop(srv.recvConn)
	go srv.readLoop(srv.broadcastConn)
	fmt.Printf(`Server %s: listening on %s
			%s
`,
		srv.ID, srv.recvConn.LocalAddr().String(), srv.broadcastConn.LocalAddr().String(),
	)

	go srv.run()
	go srv.sendTaskLoop()
}

func (srv *Server) run() {
	defer srv.wg.Done()
	for {
		select {
		case id := <-srv.closeReq:
			srv.closeSession(id)

		case incPkt := <-srv.incPktCh:
			srv.routeInkPkt(incPkt)

		case outMsg := <-srv.outgoingMsgCh:
			srv.wg.Add(1)
			go srv.dispatchMessage(outMsg)

		case <-srv.stop:
			return
		}
	}
}

func (srv *Server) Close() {
	srv.closeOnce.Do(func() {
		close(srv.stop) // signal shutdown

		srv.recvConn.Close() // unblock ReadFromUDP
		srv.sendConn.Close()
		srv.broadcastConn.Close()

		close(srv.elevatorTaskQueue)

		srv.wg.Wait() // wait for goroutines

		srv.mu.Lock()
		for sesID := range srv.sessions {
			srv.closeSessionLocked(sesID)
		}
		srv.mu.Unlock()

		fmt.Println(srv.ID, "is synced:", srv.isSynced)
		srv.PrintPeers()
	})
}

func (srv *Server) IsMaster() bool {
	srv.mu.Lock()
	defer srv.mu.Unlock()
	return srv.isMaster
}

func (srv *Server) setSelfAsMaster(isMaster bool) {
	srv.mu.Lock()
	defer srv.mu.Unlock()
	srv.isMaster = isMaster

	if isMaster {
		srv.searchingForMaster = false
	}
}
