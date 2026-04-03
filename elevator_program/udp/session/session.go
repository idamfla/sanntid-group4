package session

import (
	"elevator_program/udp"
	"elevator_program/udp/packet"
	"elevator_program/utilities"
	"fmt"
	"net"
	"sync"
)

const (
	CHANNEL_BUF = 32
)

type SessionBehavior interface {
	HandlePacket(pkt packet.Packet) error
	OnSend(pktType packet.PacketType)
}

type Session struct {
	ID       uint32
	selfAddr string
	peerAddr *net.UDPAddr // addr of original sender
	peerID   string       // TODO have a function for this instead ...

	seq uint32 // TODO remove ... maybe??

	// --- protocol state ---
	pendingPkt *packet.Packet // TODO do i need if server handles the elevator tasks?
	lastOutPkt outgoingMessage
	hasLastPkt bool

	// --- internal communication ---
	packetInCh    chan packet.Packet
	outgoingMsgCh chan outgoingMessage

	// --- lifesycle ---
	responseTimer *utilities.Timer

	// --- external systems ---
	// elev     chan<- ElevatorPacket // TODO remove when server handles elevator communication
	elevDone  chan struct{}
	taskReady chan struct{}
	srv       ServerAPI // <-- session uses this to reply

	// --- session control ---
	closeReq  chan<- uint32 // make the server/owner close this session
	stop      chan struct{}
	wg        sync.WaitGroup
	closeOnce sync.Once
}

func NewSession(id uint32,
	peerAddr *net.UDPAddr,
	srv ServerAPI,
) *Session {
	ses := &Session{
		ID:       id,
		selfAddr: srv.GetRecvString(),
		peerAddr: peerAddr,
		peerID:   peerAddr.String(),
		// seq:                seq, // TODO have it set on init ...
		pendingPkt:    &packet.Packet{},
		lastOutPkt:    outgoingMessage{},
		hasLastPkt:    false,
		packetInCh:    make(chan packet.Packet, CHANNEL_BUF),
		outgoingMsgCh: make(chan outgoingMessage, CHANNEL_BUF),

		responseTimer: utilities.NewTimer(),

		elevDone:  make(chan struct{}, 1),
		taskReady: make(chan struct{}, 1),
		srv:       srv,

		stop:     make(chan struct{}, CHANNEL_BUF),
		closeReq: srv.GetCloseReqCh(),
	}

	return ses
}

func (ses *Session) Start() {
	ses.wg.Add(2)
	go ses.listen(ses)
	go ses.sendLoop(ses)
}

func (ses *Session) Close() {
	ses.closeOnce.Do(func() {
		ses.stopResponseTimer()

		// stop base session goroutines
		close(ses.stop)
		ses.wg.Wait()

		close(ses.elevDone)
		close(ses.taskReady)

		// Clear pending packet
		ses.pendingPkt = nil
	})
}

func (ses *Session) GetID() uint32 {
	return ses.ID
}

func (ses *Session) GetSeq() uint32 {
	return ses.seq
}

func (ses *Session) GetPeerAddr() *net.UDPAddr {
	return ses.peerAddr
}

func (ses *Session) startResponseTimer() {
	ses.responseTimer.Restart(udp.RESPONSE_TIMEOUT, func() {
		fmt.Println("Peer did not respond in time ...")
		ses.QueueWhoIsAliveMsg()
		ses.stopResponseTimer()
		ses.requestClose()
	})
}

func (ses *Session) stopResponseTimer() {
	ses.responseTimer.Stop()
}
