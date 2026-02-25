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
	incoming chan incommingPacket

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
		incoming: make(chan incommingPacket, 10),
		pending:  make([]Packet, 0),
		closeReq: closeReq,

		sender: sndr,
	}

	go ses.Run()

	fmt.Println("New session created:", id)

	return ses
}

func (ses *Session) Close() {
	// optional: close incoming channel if you don't plan to reuse the session
	close(ses.incoming)

	fmt.Printf("Session %d closed\n", ses.id)
}

func (ses *Session) startTimeWaitTimer() {
	time.Sleep(5 * time.Second) // example
	ses.closeReq <- ses.id
}

func (ses *Session) Run() {
	// lastSeen := ticker
	// retransmitt := 0

	for incPck := range ses.incoming {
		// switch case
		// <-ticker
		// retransmitt every 2 sec and reset timer
		// retransmitt >= 5
		// <-ses.incomming
		// retransmit = 0
		ses.handlePacket(incPck)
	}
	fmt.Printf("Session %d stopped\n", ses.id)
}

func (ses *Session) handlePacket(incPck incommingPacket) {
	pck := incPck.packet
	addr := incPck.addr // <-- source addr

	replyAddr, _ := net.ResolveUDPAddr("udp", pck.Header.SenderAddr)

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
		ses.pending = ses.pending[:0]
		ses.sender.SendReply(replyAddr, pck, MSG_T_Done)
		go ses.startTimeWaitTimer()

	case MSG_T_Done:
		ses.closeReq <- ses.id // TODO need to find a way to remove the session from the one that is commiting without it makint the other one send a 'Done' afterwards
	}
}

// TODO
/*
cant seem to get both servers to close their session, from both ends
need to clean up code
*/
