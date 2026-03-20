package session

import (
	"elevator_program/message"
	"elevator_program/udp/packet"
	"fmt"
)

// helper
func (ses *Session) send(outPkt outgoingMessage) error {
	ses.seq++
	ses.lastOutPkt = outPkt
	return ses.tx.Send(
		ses.peerAddr,
		ses.seq,
		ses.ID,
		outPkt.PktType,
		outPkt.EMsg,
	)
}

func (ses *Session) QueueDirectMsg(pktType packet.PacketType, eMsg message.ElevatorMessage) {
	select {
	case ses.outgoingMsgCh <- outgoingMessage{
		PktType: pktType,
		EMsg:    eMsg,
	}:
	default:
		fmt.Println("Can't queue message, sessions messageQueue is full")
	}
}

func (ses *Session) QueueStateBSUpdateMsg(pktType packet.PacketType, eMsg message.ElevatorMessage) {
}

func (ses *Session) QueueWhoIsMasterMsg() {
	ses.QueueDirectMsg(packet.PKT_T_WhoIsMaster, message.ElevatorMessage{})
}

func (ses *Session) SendReply(pktType packet.PacketType) {
	ses.QueueDirectMsg(pktType, message.ElevatorMessage{})
}

func (ses *Session) sendRetry(outPkt outgoingMessage) error {
	return ses.tx.Send(
		ses.peerAddr,
		ses.seq,
		ses.ID,
		outPkt.PktType,
		outPkt.EMsg)
}

func (ses *Session) sendLoop(behavior SessionBehavior) {
	defer ses.wg.Done()

	for {
		select {
		case <-ses.stop:
			return

		case outPkt := <-ses.outgoingMsgCh:
			err := ses.send(outPkt)
			if err != nil {
				fmt.Printf("Session %d: send error: %v\n", ses.ID, err)
			}
			behavior.OnSend(outPkt.PktType)

			// if outPkt.Done != nil {
			// 	close(outPkt.Done) // signal sender

			// }
		}
	}
}

// for the SessionBehavior, does nothing
func (ses *Session) OnSend(pktType packet.PacketType) {
	ses.startResponseTimer()
}
