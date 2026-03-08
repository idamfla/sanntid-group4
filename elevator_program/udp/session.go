package udp

import (
	"fmt"
	"net"
	"time"
)

type PacketSender interface {
	SendReply(remoteAddr *net.UDPAddr, seq uint32, sessionID uint32, msgType MessageType) error
	SendBroadcast(seq uint32, sessionID uint32, msg Message) error
}

type Session struct {
	id       uint32
	addr     *net.UDPAddr // addr of original sender
	Incoming chan IncomingPacket

	// retries  int
	// lastSeen time.Time
	pending  Message
	closeReq chan<- uint32 // make the server/owner close this session

	// communication with the actual elevator
	elev     chan<- ElevatorMessage
	elevDone chan Message

	tx PacketSender // <-- session uses this to reply
}

func NewSession(id uint32, addr *net.UDPAddr, closeReq chan<- uint32, elevator chan<- ElevatorMessage, transmitter PacketSender) *Session {
	ses := &Session{
		id:       id,
		addr:     addr,
		Incoming: make(chan IncomingPacket, 10),
		pending:  Message{},
		closeReq: closeReq,

		elev: elevator,

		tx: transmitter,
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
	h := pck.Header

	replyAddr, err := net.ResolveUDPAddr("udp", pck.Header.SenderAddr)
	if err != nil {
		fmt.Printf("Session %d: invalid reply address %s\n", ses.id, pck.Header.SenderAddr)
		return
	}

	switch h.MsgType {
	case MSG_T_Data:
		ses.pending = pck.Payload
		ses.tx.SendReply(replyAddr, h.Seq+1, h.SessionID, MSG_T_Ack)

	case MSG_T_BroadcastData:
		ses.pending = pck.Payload
		ses.tx.SendReply(replyAddr, h.Seq+1, h.SessionID, MSG_T_BroadcastAck)

	case MSG_T_MasterData:
		// ses.tx.SendReply(replyAddr, h.Seq+1, h.SessionID, MSG_T_Done)
		// go ses.startTimeWaitTimer()

	case MSG_T_Ack:
		ses.tx.SendReply(replyAddr, h.Seq+1, h.SessionID, MSG_T_Commit)

	case MSG_T_BroadcastAck:

	case MSG_T_Commit:
		// clear pending
		ses.elev <- ElevatorMessage{
			SessionID: ses.id,
			Message:   ses.pending,
			Done:      ses.elevDone,
		}

		// TODO make elevator send to channel ses.elev.CommitDone (rename later) when it has completed the task
		ses.pending = Message{}
		ses.tx.SendReply(replyAddr, h.Seq+1, h.SessionID, MSG_T_Done)
		go ses.startTimeWaitTimer()

	case MSG_T_BroadcastCommit:
		// // clear pending
		// ses.pending = ses.pending[:0]
		// ses.commitCh <- pck.Payload

	case MSG_T_Done:
		ses.closeReq <- ses.id

	case MSG_T_BroadcastDone:
		// bcDone++
		// if bcDone >= 60% of active elevators {
		// 	ses.closeReq <- ses.id
		// }
	}
}
