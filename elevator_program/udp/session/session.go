package session

import (
	"elevator_program/message"
	"elevator_program/udp/packet"
	"elevator_program/udp/peerinfo"
	"elevator_program/udp/timer"
	"fmt"
	"net"
	"sync"
)

type PacketSender interface {
	Send(remoteAddr *net.UDPAddr, seq uint32, sessionID uint32, msgType packet.PacketType, eMsg message.ElevatorMessage) error
	QueueElevatorTask(eMsg message.ElevatorMessage, elevDone chan<- struct{}, taskReady <-chan struct{})
	IsMaster() bool
	GetMasterPeer() *peerinfo.PeerInfo
	StartPeerCatchup(peerAddr *net.UDPAddr)
}

type Session struct {
	ID       uint32
	peerAddr *net.UDPAddr // addr of original sender
	peerID   string

	seq uint32 // TODO remove

	// --- protocol state ---
	pendingPkt *packet.Packet // TODO do i need if server handles the elevator tasks?
	lastOutPkt outgoingMessage
	hasLastPkt bool

	// --- internal communication ---
	packetInCh    chan packet.Packet
	outgoingMsgCh chan outgoingMessage

	// --- timers/lifecycle ---
	shutdownDelayTimer *timer.Timer
	remoteCommitTimer  *timer.Timer

	// --- external systems ---
	// elev     chan<- ElevatorPacket // TODO remove when server handles elevator communication
	elevDone  chan struct{}
	taskReady chan struct{}
	tx        PacketSender // <-- session uses this to reply

	// --- session control ---
	closeReq  chan<- uint32 // make the server/owner close this session
	stop      chan struct{}
	wg        sync.WaitGroup
	closeOnce sync.Once
}

func NewSession(id uint32,
	peerAddr *net.UDPAddr,
	closeReq chan<- uint32,
	transmitter PacketSender,
) *Session {
	ses := &Session{
		ID:       id,
		peerAddr: peerAddr,
		peerID:   peerAddr.String(),
		// seq:                seq, // TODO have it set on init ...
		pendingPkt:         &packet.Packet{},
		lastOutPkt:         outgoingMessage{},
		hasLastPkt:         false,
		packetInCh:         make(chan packet.Packet, 32),
		outgoingMsgCh:      make(chan outgoingMessage, 32),
		remoteCommitTimer:  timer.NewTimer(),
		shutdownDelayTimer: timer.NewTimer(),

		elevDone:  make(chan struct{}),
		taskReady: make(chan struct{}),
		tx:        transmitter,

		stop:     make(chan struct{}),
		closeReq: closeReq,
	}

	return ses
}

func (ses *Session) Start() {
	ses.wg.Add(2)
	go ses.listen(ses)
	go ses.sendLoop(ses)
	// fmt.Printf("Session %d started\n", ses.ID)
}

func (ses *Session) Close() {
	ses.closeOnce.Do(func() {
		// Stop normal session timers
		ses.remoteCommitTimer.Stop()
		ses.shutdownDelayTimer.Stop()

		// Stop base session goroutines
		close(ses.stop)
		ses.wg.Wait()

		// Close channels
		close(ses.packetInCh)
		close(ses.outgoingMsgCh)
		close(ses.elevDone)
		close(ses.taskReady)

		// Clear pending packet
		ses.pendingPkt = nil

		fmt.Printf("Session %d closed\n", ses.ID)
	})
}
