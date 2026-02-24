package udp

import (
	"fmt"
	"net"
)

type Session struct {
	id   uint32
	addr *net.UDPAddr
	conn *net.UDPConn

	incoming chan Packet

	// retries  int
	// lastSeen time.Time
	// unacked  map[uint32]Packet
	closeReq chan<- uint32
}

func NewSession(id uint32, addr *net.UDPAddr, conn *net.UDPConn, closeReq chan<- uint32) *Session {
	ses := &Session{
		id:       id,
		addr:     addr,
		conn:     conn,
		incoming: make(chan Packet, 10),
		closeReq: closeReq,
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
	for packet := range ses.incoming {
		ses.handlePacket(packet)
	}
	fmt.Printf("Session %d stopped\n", ses.id)
}

func (ses *Session) handlePacket(pck Packet) {
	fmt.Printf("Session %d received: %+v\n",
		ses.id,
		pck.Payload,
	)

	// Send ACK back
	ack := Packet{
		Header: Header{
			Seq:       pck.Header.Seq,
			SessionID: pck.Header.SessionID,
			MsgType:   MSG_T_Ack,
		},
		Payload: Message{Content: "ACK"},
	}

	data := encodePacket(ack)
	ses.conn.WriteToUDP(data, ses.addr)

	if pck.Header.MsgType == MSG_T_Done {
		ses.closeReq <- ses.id
		return
	}
}
