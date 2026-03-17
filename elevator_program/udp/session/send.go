package session

import (
	"elevator_program/udp/message"
	"elevator_program/udp/packet"
	"fmt"
)

var emtpyMsg message.Message

// helper
func (ses *Session) send(outPkt outgoingMessage) error {
	ses.seq++
	ses.lastOutPkt = &outPkt
	return ses.tx.Send(
		ses.peerAddr,
		ses.seq,
		ses.ID,
		outPkt.PktType,
		outPkt.Msg,
	)
}

func (ses *Session) QueueSlaveUpdate(msg message.Message) {
	ses.outgoingMsgCh <- outgoingMessage{
		PktType: packet.PKT_T_SlaveUpdate,
		Msg:     msg,
	}
}

func (ses *Session) QueueBroadcastUpdate(msg message.Message) {
	ses.outgoingMsgCh <- outgoingMessage{
		PktType: packet.PKT_T_BroadcastUpdate,
		Msg:     msg,
	}
}

func (ses *Session) QueueWhoIsMasterMsg() {
	ses.outgoingMsgCh <- outgoingMessage{
		PktType: packet.PKT_T_WhoIsMaster,
		Msg:     emtpyMsg,
	}
}

// func (ses *Session) QueueStateSync() {
// 	ses.outgoingMsgCh <- outgoingMessage{
// 		PktType: packet.PKT_T_StateSync,
// 		Msg:     emtpyMsg,
// 	}
// }

func (ses *Session) SendReply(pktType packet.PacketType) {
	done := make(chan struct{})
	ses.outgoingMsgCh <- outgoingMessage{
		PktType: pktType,
		Msg:     emtpyMsg,
		Done:    done, // new field in Outgoing
	}
	<-done // wait until SendLoop actually sends it
}

func (ses *Session) sendDoneAck(pktType packet.PacketType) {
	switch pktType {
	case packet.PKT_T_BroadcastCommit:
		ses.SendReply(packet.PKT_T_BroadcastDone)
	default:
	}
}

func (ses *Session) sendRetry(outPkt outgoingMessage) error {
	return ses.tx.Send(
		ses.peerAddr,
		ses.seq,
		ses.ID,
		outPkt.PktType,
		outPkt.Msg)
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

			if outPkt.Done != nil {
				close(outPkt.Done) // signal sender

			}
		case <-ses.stop:
			return
		}
	}
}

// for the SessionBehavior, does nothing
func (ses *Session) OnSend(pktType packet.PacketType) {}
