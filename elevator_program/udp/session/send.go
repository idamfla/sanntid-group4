package session

import (
	"elevator_program/udp/message"
	"elevator_program/udp/packet"
	"fmt"
)

var emtpyMsg message.Message

// helper
func (ses *Session) send(pktType packet.PacketType, msg message.Message) error {
	ses.seq++
	return ses.tx.Send(
		ses.senderAddr,
		ses.seq,
		ses.ID,
		pktType,
		msg,
	)
}

func (ses *Session) QueueDataMessage(msg message.Message) {
	ses.outgoingMsgCh <- outgoingMessage{
		PktType: packet.PKT_T_Data,
		Msg:     msg,
	}
}

func (ses *Session) QueueMasterMessage(msg message.Message) {
	ses.outgoingMsgCh <- outgoingMessage{
		PktType: packet.PKT_T_SlaveReport,
		Msg:     msg,
	}
}

func (ses *Session) QueueBroadcastUpdate(msg message.Message) {
	ses.outgoingMsgCh <- outgoingMessage{
		PktType: packet.PKT_T_BroadcastUpdate,
		Msg:     msg,
	}
}

func (ses *Session) QueueStateSync() {
	ses.outgoingMsgCh <- outgoingMessage{
		PktType: packet.PKT_T_StateSync,
		Msg:     emtpyMsg,
	}
}

func (ses *Session) sendReply(pktType packet.PacketType) {
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
		ses.sendReply(packet.PKT_T_BroadcastDone)
	default:
		ses.sendReply(packet.PKT_T_Done)
	}
}

func (ses *Session) retry(pktType packet.PacketType, msg message.Message) error {
	return ses.tx.Send(
		ses.senderAddr,
		ses.seq,
		ses.ID,
		pktType,
		msg)
}

func (ses *Session) sendLoop(behavior SessionBehavior) {
	defer ses.wg.Done()

	for {
		select {
		case outPkt := <-ses.outgoingMsgCh:
			err := ses.send(outPkt.PktType, outPkt.Msg)
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
