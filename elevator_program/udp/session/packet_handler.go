package session

import (
	"elevator_program/message"
	"elevator_program/udp"
	"elevator_program/udp/packet"
	"fmt"
	"time"
)

func (ses *Session) ReceivePacket(pkt packet.Packet) {
	select {
	case ses.packetInCh <- pkt:
	default:
		fmt.Println("Session mailbox is full, could not receive new packet")
	}
}

func (ses *Session) HandlePacket(pkt packet.Packet) error { // TODO rename HandleIncPkt
	ses.stopShutdownTimer()

	h := pkt.Header

	if h.Seq != ses.seq+1 {
		err := fmt.Errorf("Session %d: seq mismatch (got %d, expected %d)...\n",
			ses.ID, h.Seq, ses.seq+1)
		fmt.Println(err)
		return err
	}

	ses.seq = pkt.Header.Seq
	// 	fmt.Printf(
	// 		`	seq : %d
	// 	pktType : %s
	// 	payload : %+v
	// `,
	// 		pkt.Header.Seq,
	// 		pkt.Header.PktType,
	// 		pkt.Payload,
	// 	)

	switch h.PktType {
	case packet.PKT_T_Heartbeat:
		fmt.Printf("%s sent %s\n", h.SenderAddr, h.PktType) // TODO remove db, although ... heatbeat should not end up here

	case packet.PKT_T_LostConn:
		fmt.Printf("%s lost connection ...", h.SenderAddr)
		// TODO what to do now?

	case packet.PKT_T_RequestTaskExecution:
		ses.handleRequestTaskExecution(pkt.Payload.EMsgType)

	// case packet.PKT_T_CatchupUpdate:
	// 	ses.handleCatchup()

	// case packet.PKT_T_Snapshot:
	// 	ses.handleSnapshot()

	// case packet.PKT_T_CatchupAck, packet.PKT_T_SnapshotAck:
	// 	ses.requestClose()

	// case packet.PKT_T_SlaveUpdate:
	// 	ses.handleSlaveUpdate(pkt.Payload)

	case packet.PKT_T_BroadcastUpdate:
		ses.handleStateBSUpdate()

	case packet.PKT_T_SyncMsg:
		ses.handleSyncMsg()
	// case packet.PKT_T_SyncComplete:
	// 	ses.handleSyncComplete()

	case packet.PKT_T_RequestTaskExecutionAck:
		ses.requestClose()
	// TODO master must give it an id, send it all important updates
	/*
		maybe send to elevator, elevator send done when it receive. the master elevator handle the request and use its server to start
		communication witht the wondering elevator on a private session another session
	*/

	case packet.PKT_T_BroadcastCommit, packet.PKT_T_SyncMsgCommit:
		// case packet.PKT_T_BroadcastCommit, packet.PKT_T_SyncCommit:
		go ses.handleStateBSCommit(h.PktType)

	case packet.PKT_T_ElevatorFailed:
		// TODO fault tolerence? what to do now ...

	}
	return nil
}

// ask elevator for sync, get PKT_T_StateSnapshot back
func (ses *Session) handleRequestTaskExecution(eMsgType message.ElevatorMessageType) {
	ses.QueueElevatorWorkTask(eMsgType, message.ElevatorMessage{})
	ses.notifyTaskReady()
	ses.queueReply(packet.PKT_T_RequestTaskExecutionAck)
	ses.startShutdownTimer()
}

// queue order of having elevator change its states, from master
func (ses *Session) QueueElevatorStateTask() {
	ses.queueElevatorTask(ses.pendingPkt.Payload, ses.elevDone)
}

// queue order of having master do some work, don't need to notify completion, just start a new session
func (ses *Session) QueueElevatorWorkTask(eMsgType message.ElevatorMessageType, eMsg message.ElevatorMessage) {
	var emsg message.ElevatorMessage
	if eMsgType == message.EMSG_T_StatusReport {
		emsg = eMsg
	} else {
		emsg = message.ElevatorMessage{
			ID:       ses.getPeerAddrString(),
			Addr:     ses.peerAddr.String(),
			EMsgType: eMsgType,
		}
	}

	ses.queueElevatorTask(emsg, nil)
}

func (ses *Session) handleStateBSUpdate() {
	ses.queueReply(packet.PKT_T_BroadcastAck)
	ses.QueueElevatorStateTask()
}

func (ses *Session) handleSyncMsg() {
	ses.QueueElevatorStateTask()
	ses.queueReply(packet.PKT_T_SyncMsgAck)
}

func (ses *Session) handleStateBSCommit(pktType packet.PacketType) {
	ses.notifyTaskReady()
	if err := ses.waitForElevatorDoneWithReply(); err != nil {
		return
	}

	switch pktType {
	case packet.PKT_T_SyncMsgCommit:
		ses.queueReply(packet.PKT_T_SyncComplete)
	case packet.PKT_T_BroadcastCommit:
		ses.queueReply(packet.PKT_T_BroadcastDone)
	}

	ses.pendingPkt = nil

	ses.startShutdownTimer()
}

func (ses *Session) notifyTaskReady() {
	select {
	case <-ses.stop:
	case ses.taskReady <- struct{}{}:
	default:
		fmt.Println("Notifications full, could not notify to elevator that task is ready")
	}
}

// --- elevator interaction ---
func (ses *Session) waitForElevatorDoneWithReply() error {
	if err := ses.waitForElevatorDone(); err != nil {
		ses.queueReply(packet.PKT_T_ElevatorFailed)
		fmt.Println(err)
		return err
	}
	return nil
}

// Send packet to elevator, block until timeout or elevator complete its task
func (ses *Session) waitForElevatorDone() error {
	timer := time.NewTimer(udp.LOCAL_COMMIT_TIMEOUT)
	defer timer.Stop()

	select { // wait for completion
	case <-ses.stop:
		return fmt.Errorf("Session stopped")

	case <-ses.elevDone:
		fmt.Println("Elevator done commiting")
		return nil

	case <-timer.C:
		return fmt.Errorf("Elevator failed to commit …")
	}
}
