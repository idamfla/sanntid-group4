package session

import (
	"elevator_program/message"
	"elevator_program/udp/packet"
	"fmt"
)

var emtpyEMsg message.ElevatorMessage

// helper
func (ses *Session) send(outPkt outgoingMessage) error {
	ses.seq++
	ses.lastOutPkt = &outPkt
	return ses.tx.Send(
		ses.peerAddr,
		ses.seq,
		ses.ID,
		outPkt.PktType,
		outPkt.EMsg,
	)
}

func (ses *Session) QueueSlaveUpdate(eMsg message.ElevatorMessage) {
	ses.outgoingMsgCh <- outgoingMessage{
		PktType: packet.PKT_T_SlaveUpdate,
		EMsg:    eMsg,
	}
}

func (ses *Session) QueueBroadcastUpdate(eMsg message.ElevatorMessage) {
	ses.outgoingMsgCh <- outgoingMessage{
		PktType: packet.PKT_T_BroadcastUpdate,
		EMsg:    eMsg,
	}
}

func (ses *Session) QueueWhoIsMasterMsg() {
	ses.outgoingMsgCh <- outgoingMessage{
		PktType: packet.PKT_T_WhoIsMaster,
		EMsg:    emtpyEMsg,
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
		EMsg:    emtpyEMsg,
		Done:    done, // new field in Outgoing
	}
	<-done // wait until SendLoop actually sends it
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
