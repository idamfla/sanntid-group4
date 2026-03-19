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
	fmt.Println(outPkt.EMsg.ID, outPkt.PktType, "sent msg with seq", ses.seq) // TODO db remove later
	return ses.tx.Send(
		ses.peerAddr,
		ses.seq,
		ses.ID,
		outPkt.PktType,
		outPkt.EMsg,
	)
}

func (ses *Session) QueueDirectMsg(pktType packet.PacketType, eMsg message.ElevatorMessage) {
	ses.outgoingMsgCh <- outgoingMessage{
		PktType: pktType,
		EMsg:    eMsg,
	}
}

func (ses *Session) QueueBroadcastUpdateMsg(eMsg message.ElevatorMessage) {}

func (ses *Session) QueueWhoIsMasterMsg() {}

func (ses *Session) SendReply(pktType packet.PacketType) {
	select {
	case ses.outgoingMsgCh <- outgoingMessage{
		PktType: pktType,
		EMsg:    message.ElevatorMessage{},
	}:
	default:
		fmt.Println("Session", ses.ID, "outgoingMsgCh full, dropping packet")
	}
}

func (ses *Session) sendBroadcastDone() {
	ses.SendReply(packet.PKT_T_BroadcastDone)
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
		case outPkt := <-ses.outgoingMsgCh:
			err := ses.send(outPkt)
			if err != nil {
				fmt.Printf("Session %d: send error: %v\n", ses.ID, err)
			}
			behavior.OnSend(outPkt.PktType)

			// if outPkt.Done != nil {
			// 	close(outPkt.Done) // signal sender

			// }
		case <-ses.stop:
			return
		}
	}
}

// for the SessionBehavior, does nothing
func (ses *Session) OnSend(pktType packet.PacketType) {}
