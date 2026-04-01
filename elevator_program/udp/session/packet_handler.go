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

func (ses *Session) HandlePacket(pkt packet.Packet) error {
	h := pkt.Header

	if h.Seq != ses.seq+1 {
		err := fmt.Errorf("Session %d: seq mismatch (got %d, expected %d)...\n",
			ses.ID, h.Seq, ses.seq+1)
		fmt.Println(err)
		return err

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

	switch h.PktType {
	case packet.PKT_T_Heartbeat:
		fmt.Printf("%s sent %s\n", h.SenderAddr, h.PktType)

	case packet.PKT_T_LostConn:
		fmt.Printf("%s lost connection ...", h.SenderAddr)

	case packet.PKT_T_RequestTaskExecution:
		ses.handleRequestTaskExecution(pkt.Payload.EMsgType)

	case packet.PKT_T_CatchupUpdate:
		ses.handleCatchup()

	case packet.PKT_T_Snapshot:
		ses.handleSnapshot()

	case packet.PKT_T_CatchupAck, packet.PKT_T_SnapshotAck:
		ses.requestClose()

	case packet.PKT_T_SlaveUpdate:
		ses.handleSlaveUpdate(pkt.Payload)

	case packet.PKT_T_BroadcastUpdate, packet.PKT_T_SyncComplete:
		ses.handleStateBSUpdate(h.PktType)

	case packet.PKT_T_RequestTaskExecutionAck, packet.PKT_T_SlaveUpdateAck:
		ses.requestClose()

	case packet.PKT_T_BroadcastCommit, packet.PKT_T_SyncCommit:
		go ses.handleStateBSCommit(h.PktType)

	case packet.PKT_T_ElevatorFailed:
	}
	return nil
}

func (ses *Session) handleRequestTaskExecution(eMsgType message.ElevatorMessageType) {
	ses.QueueElevatorWorkTask(eMsgType, message.ElevatorMessage{})
	fmt.Println("Halla balla")
	ses.notifyTaskReady()
	fmt.Println("Mordecay")
	ses.SendReply(packet.PKT_T_RequestTaskExecutionAck)
	ses.scheduleSessionClose()
}

func (ses *Session) handleSnapshot() {
	ses.QueueElevatorStateTask()
	ses.notifyTaskReady()
	ses.SendReply(packet.PKT_T_SnapshotAck)

	ses.scheduleSessionClose()
}

func (ses *Session) handleCatchup() {
	ses.QueueElevatorStateTask()
	ses.notifyTaskReady()
	ses.SendReply(packet.PKT_T_CatchupAck)

	ses.scheduleSessionClose()
}

func (ses *Session) handleSlaveUpdate(eMsg message.ElevatorMessage) {
	ses.QueueElevatorWorkTask(eMsg.EMsgType, eMsg) // TODO slave update is not just statusReport
	fmt.Println("YOYYOYOYOYYYOfrignrigne \n\n\n\n\n\n\n", eMsg)
	ses.notifyTaskReady()
	ses.SendReply(packet.PKT_T_SlaveUpdateAck)
	ses.scheduleSessionClose()
}

func (ses *Session) QueueElevatorStateTask() {
	ses.tx.QueueElevatorTask(ses.pendingPkt.Payload, ses.elevDone, ses.taskReady)
}

func (ses *Session) QueueElevatorWorkTask(eMsgType message.ElevatorMessageType, eMsg message.ElevatorMessage) {
	var emsg message.ElevatorMessage
	if eMsgType == message.EMSG_T_StatusReport {
		emsg = eMsg
	} else if eMsgType == message.EMSG_T_ButtonPress {
		emsg = eMsg
	} else {
		emsg = message.ElevatorMessage{
			ID:       ses.peerID,
			Addr:     ses.peerAddr.String(),
			EMsgType: eMsgType,
		}
	}

	ses.tx.QueueElevatorTask(emsg, nil, ses.taskReady)
}

func (ses *Session) handleStateBSUpdate(pktType packet.PacketType) {
	switch pktType {
	case packet.PKT_T_BroadcastUpdate:
		ses.SendReply(packet.PKT_T_BroadcastAck)
	case packet.PKT_T_SyncComplete:
		ses.SendReply(packet.PKT_T_SyncAck)
	}

	ses.QueueElevatorStateTask()
}

func (ses *Session) handleStateBSCommit(pktType packet.PacketType) {
	ses.notifyTaskReady()
	if err := ses.waitForElevatorDoneWithReply(); err != nil {
		return
	}

	switch pktType {
	case packet.PKT_T_SyncCommit:
		ses.SendReply(packet.PKT_T_SyncDone)
	case packet.PKT_T_BroadcastCommit:
		ses.SendReply(packet.PKT_T_BroadcastDone)
	}

	ses.pendingPkt = nil

	ses.scheduleSessionClose()
}

func (ses *Session) notifyTaskReady() {
	select {
	case <-ses.stop:
		fmt.Println("Im djfjrrf")
	case ses.taskReady <- struct{}{}:
		fmt.Println("freddy fazbear")
	default:
		fmt.Println("Notifications full, could not notify to elevator that task is ready")
	}
}

func (ses *Session) waitForElevatorDoneWithReply() error {
	if err := ses.waitForElevatorDone(); err != nil {
		ses.SendReply(packet.PKT_T_ElevatorFailed)
		fmt.Println(err)
		return err
	}
	return nil
}

func (ses *Session) waitForElevatorDone() error {
	timer := time.NewTimer(udp.LOCAL_COMMIT_TIMEOUT)
	defer timer.Stop()

	select {
	case <-ses.stop:
		return fmt.Errorf("Session stopped")

	case <-ses.elevDone:
		fmt.Println("Elevator done commiting")
		return nil

	case <-timer.C:
		return fmt.Errorf("Elevator failed to commit …")
	}
}

func (ses *Session) scheduleSessionClose() {
	go func() {
		select {
		case <-ses.stop:
			return
		case <-time.After(udp.SHUTDOWN_TIMEOUT):
			ses.requestClose()
		}
	}()
}

func (ses *Session) requestClose() {
	select {
	case <-ses.stop:
	case ses.closeReq <- ses.ID:
	}
}
