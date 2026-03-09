package session

import (
	"elevator_program/udp/packet"
	"fmt"
	"net"
)

func (ses *Session) handlePacket(pktCtx PacketContext) {
	pkt := pktCtx.Packet
	h := pkt.Header

	replyAddr, err := net.ResolveUDPAddr("udp", pkt.Header.SenderAddr)
	if err != nil {
		fmt.Printf("Session %d: invalid reply address %s\n", ses.ID, pkt.Header.SenderAddr)
		return
	}

	switch h.PktType {
	case packet.PKT_T_Data:
		ses.pending = &pkt
		ses.tx.SendReply(replyAddr, h.Seq+1, h.SessionID, packet.PKT_T_Ack)

	case packet.PKT_T_BroadcastData:
		ses.pending = &pkt
		ses.tx.SendReply(replyAddr, h.Seq+1, h.SessionID, packet.PKT_T_BroadcastAck)

	case packet.PKT_T_MasterData:
		// ses.tx.SendReply(replyAddr, h.Seq+1, h.SessionID, packet.PKT_T_Done)
		// go ses.startTimeWaitTimer()

	case packet.PKT_T_Ack:
		ses.tx.SendReply(replyAddr, h.Seq+1, h.SessionID, packet.PKT_T_Commit)

	case packet.PKT_T_BroadcastAck:

	case packet.PKT_T_Commit:
		// clear pending
		ses.elev <- PacketContext{
			Packet: *ses.pending,
			Done:   ses.elevDone,
		}

		// TODO make elevator send to channel ses.elev.CommitDone (rename later) when it has completed the task
		// elevator send "elev.Done <- struct{}{}" when Done
		for range ses.elevDone {
			ses.pending = nil
			ses.tx.SendReply(replyAddr, h.Seq+1, h.SessionID, packet.PKT_T_Done)
			go ses.startTimeWaitTimer()
		}

	case packet.PKT_T_BroadcastCommit:
		// // clear pending
		// ses.pending = ses.pending[:0]
		// ses.commitCh <- pkt.Payload

	case packet.PKT_T_Done:
		ses.closeReq <- ses.ID

	case packet.PKT_T_BroadcastDone:
		// bcDone++
		// if bcDone >= 60% of active elevators {
		// 	ses.closeReq <- ses.ID
		// }
	}
}
