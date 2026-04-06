package session

import (
	"elevator_program/udp"
	"elevator_program/udp/packet"
	"elevator_program/utilities"
	"net"
	"sync"
)

const (
	CHANNEL_BUF = 32
)

type SessionBehavior interface {
	HandleIncPkt(pkt packet.Packet) error
	OnSend(pktType packet.PacketType)
}

type Session struct {
	ID       uint32
	selfAddr string
	peerAddr *net.UDPAddr // addr of original sender

	seq uint32 // TODO remove ... maybe??

	// --- protocol state ---
	pendingPkt *packet.OutgoingMessage // TODO do i need if server handles the elevator tasks?
	lastOutMsg packet.OutgoingMessage
	// lastOutMsg outgoingMessage
	hasLastPkt bool

	// --- timer ---
	shutdownTimer *utilities.Timer

	// --- internal communication ---
	packetInCh    chan packet.Packet
	outgoingMsgCh chan packet.OutgoingMessage
	// outgoingMsgCh chan outgoingMessage

	// --- external systems ---
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
		// seq:                seq, // TODO have it set on init ...
		pendingPkt:    &packet.OutgoingMessage{},
		lastOutMsg:    packet.OutgoingMessage{},
		hasLastPkt:    false,
		shutdownTimer: utilities.NewTimer(),
		packetInCh:    make(chan packet.Packet, CHANNEL_BUF),
		outgoingMsgCh: make(chan packet.OutgoingMessage, CHANNEL_BUF),
		// outgoingMsgCh: make(chan outgoingMessage, CHANNEL_BUF),

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
		if ses.shutdownTimer != nil {
			ses.shutdownTimer.Stop()
		}

		// stop base session goroutines
		close(ses.stop)
		ses.wg.Wait()

		close(ses.elevDone)
		close(ses.taskReady)

		// Clear pending packet
		ses.clearPendingPkt()
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

// just the string version of the peerAddr
func (ses *Session) getPeerAddrString() string {
	return ses.GetPeerAddr().String()
}

func (ses *Session) setPendingPkt(pendingPkt *packet.OutgoingMessage) { ses.pendingPkt = pendingPkt }
func (ses *Session) clearPendingPkt()                                 { ses.pendingPkt = nil }

func (ses *Session) setHasLastPacket()   { ses.hasLastPkt = true }
func (ses *Session) clearHasLastPacket() { ses.hasLastPkt = false }

func (ses *Session) startShutdownTimer() {
	ses.shutdownTimer.Restart(udp.SHUTDOWN_TIMEOUT, func() {
		ses.requestClose()
	})
}

func (ses *Session) stopShutdownTimer() {
	ses.shutdownTimer.Stop()
}

func (ses *Session) requestClose() {
	select {
	case <-ses.stop:
	case ses.closeReq <- ses.ID:
	}
}
