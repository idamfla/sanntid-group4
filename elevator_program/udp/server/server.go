package server

import (
	"elevator_program/message"
	"elevator_program/udp/packet"
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
	QueueStateBSUpdateMsg(pktType packet.PacketType, eMsg message.ElevatorMessage)
}

type Server struct {
	ID string

	state   *ServerState
	network *ServerNetwork

	incPktCh      chan incomingPacket
	outgoingMsgCh chan outgoingMessage

	sessions map[uint32]SessionHandler
	peers    map[string]*PeerInfo
	bcSeq    uint32
	mu       sync.Mutex
	closeReq chan uint32

	stop      chan struct{}
	wg        sync.WaitGroup
	closeOnce sync.Once

	elevator          chan session.ElevatorPacket
	elevatorTaskQueue chan ElevatorTask
}

func NewServer(ip string, port int, id string, toElevator chan session.ElevatorPacket) (*Server, error) {
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
		ID:                id,
		state:             &ServerState{},
		incPktCh:          make(chan incomingPacket, CHANNEL_BUF),
		outgoingMsgCh:     make(chan outgoingMessage, CHANNEL_BUF),
		network:           network,
		sessions:          make(map[uint32]SessionHandler),
		peers:             make(map[string]*PeerInfo),
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
		srv.ID, srv.getRecvString(), srv.getBroadcastConn().LocalAddr().String(),
	)

	go srv.run()
	go srv.sendTaskLoop()
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
			srv.routeIncPkt(incPkt)

		case outMsg := <-srv.outgoingMsgCh:
			srv.wg.Add(1)
			go srv.dispatchMessage(outMsg)

		}
	}
}

func (srv *Server) Close() {
	srv.closeOnce.Do(func() {
		close(srv.stop) // signal shutdown

		srv.network.Close()

		srv.wg.Wait() // wait for goroutines

		for sesID := range srv.sessions {
			srv.closeSession(sesID)
		}

		fmt.Println(srv.ID, "is synced:", srv.isSynced())
		srv.PrintPeers()
	})
}
