package udp

import (
	"fmt"
	"net"
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

func (ses *Session) Run() {
	for incPck := range ses.incoming {
		ses.handlePacket(incPck)
	}
	fmt.Printf("Session %d stopped\n", ses.id)
}

func (ses *Session) handlePacket(incPck incommingPacket) {
	pck := incPck.packet
	addr := incPck.packet.Header.ReplyAddr // <-- source addr

	fmt.Printf(
		"Session %d received %+v, reply to %s\n",
		ses.id,
		pck.Payload,
		addr.String(),
	)

	switch pck.Header.MsgType {
	case MSG_T_Data:
		ses.pending = append(ses.pending, pck)
		ses.sender.SendReply(addr, pck, MSG_T_Ack)

	case MSG_T_Ack:
		ses.sender.SendReply(addr, pck, MSG_T_Commit)

	case MSG_T_Commit:
		ses.pending = ses.pending[:0]
		ses.sender.SendReply(addr, pck, MSG_T_Done)
		ses.closeReq <- ses.id

	case MSG_T_Done:
		ses.closeReq <- ses.id
	}
}
