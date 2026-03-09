package session

import (
	"elevator_program/udp"
	"elevator_program/udp/packet"
	"fmt"
	"net"
	"time"
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
		ses.tx.SendReply(replyAddr, h.Seq+1, h.SessionID, packet.PKT_T_CommitReceived)

		commitPacket := ses.pending

		go ses.commitToElevator(commitPacket, replyAddr)

	case packet.PKT_T_BroadcastCommit:
		// // clear pending

	case packet.PKT_T_CommitReceived:
		// TODO ... what to do?
		ses.commitTimer.Restart(udp.TIMEOUT*time.Second, func() {
			fmt.Println("The receiving elevator did not commit the task ...")
			// TODO what now??
			// ses.closeReq <- ses.ID
		})

	case packet.PKT_T_CommitFailed:
		// TODO fault tolerence? what to do now ...

	case packet.PKT_T_Done:
		ses.commitTimer.Stop()
		ses.closeReq <- ses.ID

	case packet.PKT_T_BroadcastDone:
		// bcDone++
		// if bcDone >= 60% of active elevators {
		// 	ses.closeReq <- ses.ID
		// }
	}
}

func (ses *Session) commitToElevator(pkt *packet.Packet, replyAddr *net.UDPAddr) {
	doneCh := make(chan struct{})

	// send to elevator
	ses.elev <- PacketContext{
		Packet: *pkt,
		Done:   doneCh,
	}

	select { // wait for completion
	case <-time.After(udp.TIMEOUT * time.Second):
		ses.timeWaitTimer.Stop()
		ses.tx.SendReply(replyAddr, pkt.Header.Seq+2, pkt.Header.SessionID, packet.PKT_T_CommitFailed)
		return
	case <-doneCh:
		fmt.Println("Elevator done commiting")
	}

	// reset pending and notify sender
	ses.pending = nil
	ses.tx.SendReply(replyAddr, pkt.Header.Seq+2, pkt.Header.SessionID, packet.PKT_T_Done)

	// start countdown to session termination
	ses.timeWaitTimer.Restart(udp.TIMEOUT*time.Second, func() {
		ses.closeReq <- ses.ID
	})
}
