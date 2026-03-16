package session

import (
	"elevator_program/udp"
	"elevator_program/udp/packet"
	"fmt"
	"time"
)

func (ses *Session) ReceivePacket(pkt packet.Packet) {
	ses.packetInCh <- pkt
}

func (ses *Session) HandlePacket(pkt packet.Packet) error {
	h := pkt.Header

	if !ses.checkSequence(h.Seq) {
		fmt.Printf("order of packages is off ... got: %d, expected: %d\n", h.Seq, ses.seq+1)
		return ses.sendRetry(ses.lastOutPkt)

	}

	ses.seq = pkt.Header.Seq
	fmt.Printf(
		`	seq : %d	
	pktType : %s
	payload : %+v
`,
		pkt.Header.Seq,
		pkt.Header.PktType,
		pkt.Payload,
	)

	ses.shutdownDelayTimer.Stop()

	switch h.PktType {
	case packet.PKT_T_Heartbeat:
		fmt.Printf("%s sent %s\n", h.SenderAddr, h.PktType) // TODO remove db, although ... heatbeat should not end up here

	case packet.PKT_T_LostConn:
		fmt.Printf("%s lost connection ...", h.SenderAddr)
		// TODO what to do now?

	case packet.PKT_T_Data, packet.PKT_T_BroadcastUpdate, packet.PKT_T_SlaveReport:
		ses.handleData(&pkt, h.PktType)
		ses.QueueElevatorTask()

	case packet.PKT_T_Ack:
		ses.sendReply(packet.PKT_T_Commit)
		ses.remoteCommitTimer.Restart(udp.REMOTE_COMMIT_TIMEOUT*time.Second, func() {
			fmt.Println("The receiving elevator did not commit the task ...")
			// TODO what now??
			// ses.closeReq <- ses.ID
		})

	case packet.PKT_T_Commit, packet.PKT_T_BroadcastCommit:
		go ses.handleCommit(h.PktType)

	case packet.PKT_T_ElevatorFailed:
		// TODO fault tolerence? what to do now ...

	case packet.PKT_T_Done, packet.PKT_T_ReportAck:
		ses.remoteCommitTimer.Stop()
		ses.requestClose()

	case packet.PKT_T_RequestNewOrder:
		ses.pendingPkt = &pkt
		go ses.handleRequestNewOrder(ses.pendingPkt)
	case packet.PKT_T_StateSync:
		// TODO master must give it an id, send it all important updates
		/*
			maybe send to elevator, elevator send done when it receive. the master elevator handle the request and use its server to start
			communication witht the wondering elevator on a private session another session
		*/
	}
	return nil
}

func (ses *Session) handleData(pkt *packet.Packet, pktType packet.PacketType) {
	ses.pendingPkt = pkt
	switch pktType {
	case packet.PKT_T_BroadcastUpdate:
		ses.sendReply(packet.PKT_T_BroadcastUpdateAck)
	case packet.PKT_T_SlaveReport:
		ses.sendReply(packet.PKT_T_ReportAck)

		// select {
		// case ses.taskReady <- struct{}{}:
		// case <-ses.stop:
		// } // TODO elevator should receive and then start a broadcast session where it send the packet to everyone
		// ses.waitForElevatorDone()

		ses.scheduleSessionClose()
	default:
		ses.sendReply(packet.PKT_T_Ack)
	}
}

func (ses *Session) QueueElevatorTask() {
	ses.tx.QueueElevatorTask(*ses.pendingPkt, ses.elevDone, ses.taskReady)
}

func (ses *Session) handleCommit(pktType packet.PacketType) {
	ses.signalTaskReady()
	if err := ses.waitForElevatorDoneWithReply(); err != nil {
		return
	}

	ses.sendDoneAck(pktType)
	ses.pendingPkt = nil

	ses.scheduleSessionClose()
}

func (ses *Session) signalTaskReady() {
	select {
	case ses.taskReady <- struct{}{}:
	case <-ses.stop:
	}
}

func (ses *Session) handleRequestNewOrder(pkt *packet.Packet) { // TODO
	// if err := ses.waitForElevatorDoneWithReply(); err != nil {
	// 	return
	// }
	// ses.scheduleSessionClose()
	// ses.sendReply(packet.PKT_T_Done)
}

// --- elevator interaction
func (ses *Session) waitForElevatorDoneWithReply() error {
	if err := ses.waitForElevatorDone(); err != nil {
		ses.sendReply(packet.PKT_T_ElevatorFailed)
		fmt.Println(err)
		return err
	}
	return nil
}

// Send packet to elevator, block until timeout or elevator complete its task
func (ses *Session) waitForElevatorDone() error {
	select { // wait for completion
	case <-ses.elevDone:
		fmt.Println("Elevator done commiting")
		return nil
	case <-time.After(udp.LOCAL_COMMIT_TIMEOUT * time.Second):
		return fmt.Errorf("Elevator failed to commit …")
	case <-ses.stop:
		return fmt.Errorf("Session stopped")
	}
}

func (ses *Session) checkSequence(seq uint32) bool {
	return seq == ses.seq+1
}

// --- lifecycle / timers
func (ses *Session) scheduleSessionClose() {
	ses.shutdownDelayTimer.Restart(udp.SHUTDOWN_TIMEOUT*time.Second, func() {
		ses.closeReq <- ses.ID
	})
}

func (ses *Session) requestClose() {
	select {
	case ses.closeReq <- ses.ID:
	default:
	}
}
