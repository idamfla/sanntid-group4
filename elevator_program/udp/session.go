package udp

import (
	"fmt"
	"net"
	"time"
)

type Sender interface {
	SendReply(remoteAddr *net.UDPAddr, pck Packet, msgType MessageType) error
}

type Session struct {
	id       uint32
	addr     *net.UDPAddr // addr of original sender
	Incoming chan IncomingPacket

	// retries  int
	// lastSeen time.Time
	pending  []Packet
	closeReq chan<- uint32

	sender Sender // <-- session uses this to reply
}

func NewSession(id uint32, addr *net.UDPAddr, closeReq chan<- uint32, sndr Sender) *Session {
	ses := &Session{
		id:       id,
		addr:     addr,
		Incoming: make(chan IncomingPacket, 10),
		pending:  make([]Packet, 0),
		closeReq: closeReq,

		sender: sndr,
	}

	go ses.Run()
	fmt.Println("New session created:", id)
	return ses
}

func (ses *Session) Close() { // TODO maybe guard against closing ses.incoming if already closed ...
	// optional: close incoming channel if you don't plan to reuse the session
	close(ses.Incoming)

	fmt.Printf("Session %d closed\n", ses.id)
}

// startTimeWaitTimer closes the session after a delay
func (ses *Session) startTimeWaitTimer() {
	time.Sleep(5 * time.Second)
	ses.closeReq <- ses.id
}

func (ses *Session) Run() {
	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()
	// lastSeen := ticker
	retransmissions := 0

	for {
		select {
		case incPck, ok := <-ses.Incoming:
			if !ok {
				// Channel closed, stop the session
				fmt.Printf("Session %d incoming channel closed, stopping\n", ses.id)
				return
			}
			retransmissions = 0
			ses.handlePacket(incPck)
		case <-ticker.C:
			// ses.retransmitt()
			retransmissions++
			if retransmissions > 5 {
				fmt.Printf("Session %d: receiver seems dead, stopping retransmissions\n", ses.id)
				return
			}
		}
	}
}

func (ses *Session) handlePacket(incPck IncomingPacket) {
	pck := incPck.Packet
	addr := incPck.Addr // <-- source addr

	replyAddr, err := net.ResolveUDPAddr("udp", pck.Header.SenderAddr)
	if err != nil {
		fmt.Printf("Session %d: invalid reply address %s\n", ses.id, pck.Header.SenderAddr)
		return
	}

	fmt.Printf(
		"Session %d received from %s: %+v, reply to %s\n",
		ses.id,
		addr.String(),
		pck.Payload,
		replyAddr.String(),
	)

	switch pck.Header.MsgType {
	case MSG_T_Data:
		ses.pending = append(ses.pending, pck)
		ses.sender.SendReply(replyAddr, pck, MSG_T_Ack)

	case MSG_T_Ack:
		ses.sender.SendReply(replyAddr, pck, MSG_T_Commit)

	case MSG_T_Commit:
		// clear pending
		ses.pending = ses.pending[:0]
		ses.sender.SendReply(replyAddr, pck, MSG_T_Done)
		go ses.startTimeWaitTimer()

	case MSG_T_Done:
		// tell session initator/manager to remove this session
		ses.closeReq <- ses.id
	}
}
