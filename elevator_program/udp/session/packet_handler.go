package session

import (
	"elevator_program/udp"
	"elevator_program/udp/packet"
	"fmt"
	"net"
	"time"
)

func (ses *Session) handlePacket(incPkt IncomingPacket) error {
	pkt := incPkt.Packet
	h := pkt.Header

	if err := ses.resolveSenderAddr(h.SenderAddr); err != nil {
		return err
	}

	if !ses.checkSequence(h.Seq) {
		fmt.Printf("order of packages is off ... got: %d, expected: %d\n", h.Seq, ses.Seq+1)
		// return ses.sendRetry()

	}

	ses.Seq = pkt.Header.Seq
	fmt.Printf(
		`	seq : %d	
	pktType : %s
	payload : %+v
`,
		pkt.Header.Seq,
		pkt.Header.PktType,
		incPkt.Packet.Payload,
	)

	switch h.PktType {
	case packet.PKT_T_Heartbeat:
		fmt.Printf("%s sent %s\n", h.SenderAddr, h.PktType) // TODO remove db

	case packet.PKT_T_Data, packet.PKT_T_BroadcastData, packet.PKT_T_MasterData:
		ses.handleData(&pkt, h.PktType)

	case packet.PKT_T_Ack:
		ses.sendReply(packet.PKT_T_Commit)
		ses.commitTimer.Restart(udp.REMOTE_COMMIT_TIMEOUT*time.Second, func() {
			fmt.Println("The receiving elevator did not commit the task ...")
			// TODO what now??
			// ses.closeReq <- ses.ID
		})

	case packet.PKT_T_BroadcastAck:
		ses.sendReply(packet.PKT_T_BroadcastCommit)

	case packet.PKT_T_MasterAck:
		ses.timeWaitTimer.Restart(udp.LOCAL_COMMIT_TIMEOUT*time.Second, func() {
			ses.closeReq <- ses.ID
		})

	case packet.PKT_T_Commit, packet.PKT_T_BroadcastCommit:
		commitPacket := ses.pendingPkt
		go ses.commitToElevator(commitPacket, h.PktType)

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
	return nil
}

// func (ses *Session) sendRetry(outPkt OutgoingPacket) {
// 	// return ses.tx.SendReply(ses.senderAddr, ses.seq, ses.id, incPkt)
// }

func (ses *Session) handleData(pkt *packet.Packet, pktType packet.PacketType) {
	ses.pendingPkt = pkt
	switch pktType {
	case packet.PKT_T_BroadcastData:
		ses.sendReply(packet.PKT_T_BroadcastAck)

	case packet.PKT_T_MasterData:
		ses.sendReply(packet.PKT_T_MasterAck)
		// start broadcast
	default:
		ses.sendReply(packet.PKT_T_Ack)
	}
}

func (ses *Session) commitToElevator(pkt *packet.Packet, pktType packet.PacketType) {
	doneCh := make(chan struct{})

	// send to elevator
	ses.elev <- ElevatorPacket{
		Packet: *pkt,
		Done:   doneCh,
	}

	select { // wait for completion
	case <-time.After(udp.LOCAL_COMMIT_TIMEOUT * time.Second):
		ses.timeWaitTimer.Stop()
		ses.sendReply(packet.PKT_T_CommitFailed)
		fmt.Println("Elevator failed to commit ...")
		return
	case <-doneCh:
		fmt.Println("Elevator done commiting")
	}

	// reset pendingPkt and notify sender
	switch pktType {
	case packet.PKT_T_BroadcastCommit:
		ses.sendReply(packet.PKT_T_BroadcastDone)
	default:
		ses.sendReply(packet.PKT_T_Done)
	}
	ses.pendingPkt = nil

	// start countdown to session termination
	ses.timeWaitTimer.Restart(udp.LOCAL_COMMIT_TIMEOUT*time.Second, func() {
		ses.closeReq <- ses.ID
	})
}

func (ses *Session) resolveSenderAddr(addr string) error {
	if ses.senderAddr == nil {
		udpAddr, err := net.ResolveUDPAddr("udp", addr)
		if err != nil {
			fmt.Printf("Session %d: invalid reply address %s\n", ses.ID, addr)
			return err
		}
		ses.senderAddr = udpAddr
	}
	return nil
}

func (ses *Session) checkSequence(seq uint32) bool {
	return seq == ses.Seq+1
}
