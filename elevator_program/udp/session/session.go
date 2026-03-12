package session

import (
	"elevator_program/udp/message"
	"elevator_program/udp/packet"
	"fmt"
	"net"
	"sync"
)

const (
	INIT_SEQ_NUMBER = 0
)

type PacketSender interface {
	Send(remoteAddr *net.UDPAddr, seq uint32, sessionID uint32, msgType packet.PacketType, msg message.Message) error
	// SendBroadcast(seq uint32, sessionID uint32, msg message.Message) error
}

type Session struct {
	ID         uint32
	senderAddr *net.UDPAddr // addr of original sender

	Seq uint32

	// --- protocol state ---
	pendingPkt *packet.Packet
	// IsClosing bool // TODO make into enum

	// retries  int
	// lastSeen time.Time

	// --- internal communication ---
	RecvCh chan IncomingPacket
	sendCh chan OutgoingPacket

	// --- timers/lifecycle ---
	timeWaitTimer *SessionTimer
	commitTimer   *SessionTimer

	// --- external systems ---
	elev chan<- ElevatorPacket
	tx   PacketSender // <-- session uses this to reply

	closeReq chan<- uint32 // make the server/owner close this session
	wg       sync.WaitGroup
	stop     chan struct{}

	closeOnce sync.Once
}

func NewSession(id uint32,
	addr *net.UDPAddr,
	closeReq chan<- uint32,
	elevator chan<- ElevatorPacket,
	transmitter PacketSender,
) *Session {
	ses := &Session{
		ID:         id,
		senderAddr: addr,
		Seq:        INIT_SEQ_NUMBER,
		pendingPkt: &packet.Packet{},
		// IsClosing:     false,
		RecvCh:        make(chan IncomingPacket, 32),
		sendCh:        make(chan OutgoingPacket, 32),
		timeWaitTimer: NewSessionTimer(),
		commitTimer:   NewSessionTimer(),

		elev: elevator,
		tx:   transmitter,

		stop:     make(chan struct{}),
		closeReq: closeReq,
	}

	// go ses.Run()
	fmt.Println("New session created:", id)

	return ses
}

func (ses *Session) Close() {
	ses.closeOnce.Do(func() {
		close(ses.stop)
		ses.wg.Wait()
		close(ses.RecvCh)
		close(ses.sendCh)
		close(ses.closeReq)
		ses.pendingPkt = nil

		fmt.Printf("Session %d closed\n", ses.ID)
	})
}

// TODO this is not working ....
func (ses *Session) Start() {
	ses.wg.Add(2)
	go ses.Listen()
	go ses.SendLoop()
	fmt.Printf("Session %d started\n", ses.ID)
}
