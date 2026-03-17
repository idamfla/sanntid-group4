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

	if h.Seq != ses.seq+1 {
		fmt.Printf("Session %d: seq mismatch (got %d, expected %d), retrying last packet\n",
			ses.ID, h.Seq, ses.seq+1)
		return ses.sendRetry(*ses.lastOutPkt)

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

	case packet.PKT_T_SlaveUpdate, packet.PKT_T_SyncRequest:
		ses.handleWorkerRequest(&pkt, h.PktType)

	case packet.PKT_T_StateSnapshot:
		ses.handleSnapshot()

	case packet.PKT_T_BroadcastUpdate:
		ses.SendReply(packet.PKT_T_BroadcastAck)
		ses.QueueElevatorTask()

	case packet.PKT_T_SyncAck, packet.PKT_T_SlaveUpdateAck, packet.PKT_T_SnapshotAck:
		ses.remoteCommitTimer.Stop()
		ses.requestClose()
	// TODO master must give it an id, send it all important updates
	/*
		maybe send to elevator, elevator send done when it receive. the master elevator handle the request and use its server to start
		communication witht the wondering elevator on a private session another session
	*/

	case packet.PKT_T_BroadcastCommit:
		go ses.handleBroadcastCommit()

	case packet.PKT_T_ElevatorFailed:
		// TODO fault tolerence? what to do now ...

	}
	return nil
}

func (ses *Session) handleWorkerRequest(pkt *packet.Packet, pktType packet.PacketType) {
	ses.pendingPkt = pkt
	switch pktType {
	case packet.PKT_T_SyncRequest:
		ses.SendReply(packet.PKT_T_SlaveUpdateAck)

	case packet.PKT_T_SlaveUpdate:
		ses.SendReply(packet.PKT_T_SlaveUpdateAck)
	}

	ses.QueueElevatorTask()
	ses.notifyTaskReady()
	ses.scheduleSessionClose()
}

func (ses *Session) handleSnapshot() {
	ses.SendReply(packet.PKT_T_SnapshotAck)
	ses.QueueElevatorTask()
	ses.notifyTaskReady()
	ses.scheduleSessionClose()
}

func (ses *Session) QueueElevatorTask() {
	ses.tx.QueueElevatorTask(*ses.pendingPkt, ses.elevDone, ses.taskReady)
}

func (ses *Session) handleBroadcastCommit() {
	ses.notifyTaskReady()
	if err := ses.waitForElevatorDoneWithReply(); err != nil {
		return
	}

	ses.sendBroadcastDone()
	ses.pendingPkt = nil

	ses.scheduleSessionClose()
}

func (ses *Session) notifyTaskReady() {
	select {
	case ses.taskReady <- struct{}{}:
	case <-ses.stop:
	}
}

// --- elevator interaction
func (ses *Session) waitForElevatorDoneWithReply() error {
	if err := ses.waitForElevatorDone(); err != nil {
		ses.SendReply(packet.PKT_T_ElevatorFailed)
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
	case <-time.After(udp.LOCAL_COMMIT_TIMEOUT):
		return fmt.Errorf("Elevator failed to commit …")
	case <-ses.stop:
		return fmt.Errorf("Session stopped")
	}
}

// --- lifecycle / timers
func (ses *Session) scheduleSessionClose() {
	ses.shutdownDelayTimer.Restart(udp.SHUTDOWN_TIMEOUT, func() {
		ses.closeReq <- ses.ID
	})
}

func (ses *Session) requestClose() {
	select {
	case ses.closeReq <- ses.ID:
	default:
	}
}
