package server

import (
	"elevator_program/udp/packet"
	"elevator_program/udp/session"
	"fmt"
	"net"
	"sync"
)

const (
	CHANNEL_BUF                 = 32
	MASTER_ELECTION_SESSSION_ID = uint32(1)
)

type Server struct {
	Alias string

	state    *ServerState
	network  *ServerNetwork
	sessions *SessionManager
	peers    *PeerManager

	incPktCh      chan incomingPacket
	outgoingMsgCh chan packet.OutgoingMessage
	// outgoingMsgCh chan outgoingMessage

	bcSeq uint32

	closeReq chan uint32

	stop      chan struct{}
	wg        sync.WaitGroup
	closeOnce sync.Once

	elevator          chan session.ElevatorPacket
	elevatorTaskQueue chan ElevatorTask
}

func NewServer(ip string, port int, alias string, toElevator chan session.ElevatorPacket) (*Server, error) {
	addr := net.UDPAddr{
		IP:   net.ParseIP(ip), // parse the string IP
		Port: port,
	}

	network, err := NewServerNetwork(&addr)
	if err != nil {
		fmt.Println("Couldn't set up network,", err)
		return nil, err
	}

	srv := &Server{
		Alias:         alias,
		state:         &ServerState{},
		incPktCh:      make(chan incomingPacket, CHANNEL_BUF),
		outgoingMsgCh: make(chan packet.OutgoingMessage, CHANNEL_BUF),
		// outgoingMsgCh: make(chan outgoingMessage, CHANNEL_BUF),
		network: network,

		sessions: NewSessionManager(),
		peers:    NewPeerManager(),
		// peers:             make(map[string]*PeerInfo),
		closeReq:          make(chan uint32, CHANNEL_BUF),
		stop:              make(chan struct{}, CHANNEL_BUF),
		elevator:          toElevator,
		elevatorTaskQueue: make(chan ElevatorTask, CHANNEL_BUF),
	}

	return srv, nil
}

func (srv *Server) Start() {
	srv.wg.Add(4)
	go srv.readLoop(srv.getRecvConn())
	go srv.readLoop(srv.getBroadcastConn())
	fmt.Printf(`Server %s: listening on %s
			%s
`,
		srv.Alias, srv.GetRecvString(), srv.getBroadcastConn().LocalAddr().String(),
	)

	srv.createMasterElectionSession()

	go srv.run()
	go srv.sendTaskLoop()
}

func (srv *Server) Close() {
	srv.closeOnce.Do(func() {
		close(srv.stop) // signal shutdown

		srv.network.Close()

		srv.wg.Wait() // wait for goroutines

		srv.sessions.Close() // TODO make function closeAllSessions

		fmt.Println(srv.Alias, "is synced:", srv.isSynced())
		srv.PrintPeers()
	})
}

func (srv *Server) GetAlias() string { return srv.Alias }

func (srv *Server) GetCloseReqCh() chan uint32 {
	return srv.closeReq
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
			srv.handleIncPkt(incPkt)

		case outMsg := <-srv.outgoingMsgCh:
			srv.wg.Add(1)
			go srv.handleOutPkt(outMsg)

		}
	}
}
