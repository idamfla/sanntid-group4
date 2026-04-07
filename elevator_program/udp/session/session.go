package session

import (
	"elevator_program/udp/packet"
	"net"
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

	state *SessionState

	// --- internal communication ---
	packetInCh    chan packet.Packet
	outgoingMsgCh chan packet.OutgoingMessage
	// outgoingMsgCh chan outgoingMessage

	// --- external systems ---
	elevDone  chan struct{}
	taskReady chan struct{}

	lifecycle *SessionLifecycle
	srv       ServerAPI // <-- session uses this to reply
}

func NewSession(id uint32,
	peerAddr *net.UDPAddr,
	srv ServerAPI,
) *Session {
	ses := &Session{
		ID:       id,
		selfAddr: srv.GetRecvString(),
		peerAddr: peerAddr,

		state:         NewSessionState(),
		packetInCh:    make(chan packet.Packet, CHANNEL_BUF),
		outgoingMsgCh: make(chan packet.OutgoingMessage, CHANNEL_BUF),

		elevDone:  make(chan struct{}, 1),
		taskReady: make(chan struct{}, 1),

		lifecycle: NewSessionLifecycle(srv.CloseReqCh()),
		srv:       srv,
	}

	return ses
}

func (ses *Session) Start() {
	ses.WgAdd(2)
	go ses.listen(ses)
	go ses.sendLoop(ses)
}

func (ses *Session) Close() {
	ses.lifecycle.CloseOnce.Do(func() {
		ses.stopShutdownTimer()

		// stop base session goroutines
		close(ses.lifecycle.Stop)
		ses.WgWait()

		close(ses.elevDone)
		close(ses.taskReady)
	})
}

func (ses *Session) GetID() uint32 {
	return ses.ID
}

func (ses *Session) GetPeerAddr() *net.UDPAddr {
	return ses.peerAddr
}

// just the string version of the peerAddr
func (ses *Session) getPeerAddrString() string {
	return ses.GetPeerAddr().String()
}
