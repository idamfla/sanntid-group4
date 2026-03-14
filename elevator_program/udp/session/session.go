package session

import (
	"elevator_program/message"
	"elevator_program/udp"
	"elevator_program/udp/packet"
	"fmt"
	"net"
	"time"
)

type PacketSender interface {
	SendReply(remoteAddr *net.UDPAddr, seq uint32, sessionID uint32, msgType packet.PacketType) error
	SendBroadcast(seq uint32, sessionID uint32, msg message.Message) error
}

type Session struct {
	ID         uint32
	addr       *net.UDPAddr // addr of original sender
	IncomingCh chan PacketContext
	pending    *packet.Packet

	// expectedSeq uint32

	IsClosing bool

	// retries  int
	// lastSeen time.Time

	closeReq chan<- uint32 // make the server/owner close this session

	timeWaitTimer *SessionTimer
	commitTimer   *SessionTimer

	// communication with the actual elevator
	elev chan<- PacketContext
	// elevDone chan struct{}

	tx PacketSender // <-- session uses this to reply
}

func NewSession(id uint32, addr *net.UDPAddr, closeReq chan<- uint32, elevator chan<- PacketContext, transmitter PacketSender) *Session {
	ses := &Session{
		ID:            id,
		addr:          addr,
		IncomingCh:    make(chan PacketContext, 10),
		pending:       &packet.Packet{},
		IsClosing:     false,
		closeReq:      closeReq,
		timeWaitTimer: NewSessionTimer(),
		commitTimer:   NewSessionTimer(),

		elev: elevator,

		tx: transmitter,
	}

	go ses.Run()
	fmt.Println("New session created:", id)
	return ses
}

// TODO make sure the sessionID is unique
// func newSessionToken() uint64 {
//     var b [8]byte
//     _, err := rand.Read(b[:])
//     if err != nil {
//         panic(err) // unlikely
//     }
//     return binary.LittleEndian.Uint64(b[:])
// }

func (ses *Session) Close() { // TODO maybe guard against closing ses.IncomingCh if already closed ...
	// optional: close IncomingCh channel if you don't plan to reuse the session
	close(ses.IncomingCh)
	ses.pending = nil

	fmt.Printf("Session %d closed\n", ses.ID)
}

func (ses *Session) Run() {
	ticker := time.NewTicker(udp.RETRY_FREQUENCY * time.Second)
	defer ticker.Stop()
	// lastSeen := ticker
	retransmissions := 0

	for {
		select {
		case pktCtx, ok := <-ses.IncomingCh:
			if !ok {
				// Channel closed, stop the session
				fmt.Printf("Session %d IncomingCh channel closed, stopping\n", ses.ID)
				return
			}
			retransmissions = 0
			ses.handlePacket(pktCtx)
		case <-ticker.C:
			// ses.retransmitt()
			retransmissions++
			if retransmissions > udp.MAX_RETRIES {
				fmt.Printf("Session %d: receiver seems dead, stopping retransmissions\n", ses.ID)
				return
			}
		}
	}
}
