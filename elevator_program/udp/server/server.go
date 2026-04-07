package server

import (
	"elevator_program/udp/packet"
	"fmt"
	"net"
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

	incPktCh      chan packet.Packet
	outgoingMsgCh chan packet.OutgoingMessage

	// closeReq chan uint32

	// stop      chan struct{}
	// wg        sync.WaitGroup
	// closeOnce sync.Once
	lifecycle *ServerLifecycle

	elevator *ElevatorInterface
}

func NewServer(ip string, port int, alias string, elevRecv chan ElevatorPacket) (*Server, error) {
	addr := net.UDPAddr{
		IP:   net.ParseIP(ip),
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
		incPktCh:      make(chan packet.Packet, CHANNEL_BUF),
		outgoingMsgCh: make(chan packet.OutgoingMessage, CHANNEL_BUF),
		network:       network,

		sessions:  NewSessionManager(),
		peers:     NewPeerManager(),
		lifecycle: NewServerLifecycle(),
		// closeReq: make(chan uint32, CHANNEL_BUF),
		// stop:     make(chan struct{}, 1),
		elevator: NewElevatorInterface(elevRecv),
	}

	return srv, nil
}

func (srv *Server) Start() {
	srv.WgAdd(4)
	go srv.readLoop(srv.getRecvConn())
	go srv.readLoop(srv.getBroadcastConn())
	fmt.Printf(`Server %s: listening on %s
			%s
`,
		srv.Alias, srv.GetRecvString(), srv.getBroadcastConn().LocalAddr().String(),
	)

	// srv.createMasterElectionSession()
	srv.getOrCreateMasterElectionSession()

	go srv.run()
	go srv.sendTaskLoop()
}

func (srv *Server) Close() {
	srv.lifecycle.CloseOnce.Do(func() {
		close(srv.lifecycle.Stop) // signal shutdown

		srv.network.Close()

		srv.WgWait() // wait for goroutines

		srv.sessions.Close() // TODO make function closeAllSessions

		fmt.Println(srv.Alias, "is synced:", srv.isSynced())
		srv.PrintPeers()
	})
}

func (srv *Server) GetAlias() string { return srv.Alias }
func (srv *Server) GetIdentity() packet.Identity {
	return packet.Identity{
		Identifier: srv.GetRecvString(),
		Alias:      srv.GetAlias(),
	}
}

func (srv *Server) run() {
	defer srv.WgDone()
	for {
		select {
		case <-srv.stopCh():
			return

		case id := <-srv.CloseReqCh():
			srv.closeSession(id)

		case incPkt := <-srv.incPktCh:
			srv.handleIncPkt(incPkt)

		case outMsg := <-srv.outgoingMsgCh:
			srv.WgAdd(1)
			go srv.handleOutMsg(outMsg)

		}
	}
}

func (srv *Server) updateSyncFromMsg(outMsg packet.OutgoingMessage) {
	recipientAddr := outMsg.Origin.Identifier
	if srv.GetRecvString() == recipientAddr {
		srv.setSynced()
	}
	srv.setPeerSynced(recipientAddr)
}
