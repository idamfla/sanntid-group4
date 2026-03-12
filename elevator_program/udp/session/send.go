package session

import (
	"elevator_program/udp/message"
	"elevator_program/udp/packet"
	"fmt"
)

var emtpyMsg message.Message

// helper
func (ses *Session) send(pktType packet.PacketType, msg message.Message) error {
	ses.Seq++
	return ses.tx.Send(
		ses.senderAddr,
		ses.Seq,
		ses.ID,
		pktType,
		msg,
	)
}

func (ses *Session) SendDataMessage(msg message.Message) {
	ses.sendCh <- OutgoingPacket{
		PktType: packet.PKT_T_Data,
		Msg:     msg,
	}
}

func (ses *Session) SendMasterMessage(msg message.Message) {
	ses.sendCh <- OutgoingPacket{
		PktType: packet.PKT_T_MasterData,
		Msg:     msg,
	}
}

func (ses *Session) SendBroadcastData(msg message.Message) {
	ses.sendCh <- OutgoingPacket{
		PktType: packet.PKT_T_BroadcastData,
		Msg:     msg,
	}
}

func (ses *Session) sendReply(pktType packet.PacketType) {
	done := make(chan struct{})
	ses.sendCh <- OutgoingPacket{
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
		ses.Seq,
		ses.ID,
		pktType,
		msg)
}

func (ses *Session) SendLoop() {
	defer ses.wg.Done()

	for {
		select {
		case outPkt := <-ses.sendCh:
			err := ses.send(outPkt.PktType, outPkt.Msg)
			if err != nil {
				fmt.Printf("Session %d: send error: %v\n", ses.ID, err)
			}
			if outPkt.Done != nil {
				close(outPkt.Done) // signal sender

			}
		case <-ses.stop:
			return
		}
	}
}
