package session

import (
	"elevator_program/message"
	"elevator_program/udp"
	"elevator_program/udp/packet"
	"fmt"
	"time"
)

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

	switch h.PktType {
	case packet.PKT_T_Heartbeat:
		fmt.Printf("%s sent %s\n", h.SenderAddr, h.PktType) // TODO remove db, although ... heatbeat should not end up here

	case packet.PKT_T_LostConn:
		fmt.Printf("%s lost connection ...", h.SenderAddr)
		// TODO what to do now?

	case packet.PKT_T_RequestTaskExecution, packet.PKT_T_BroadcastUpdate, packet.PKT_T_SyncMsg:
		ses.handleFirstIncomming(pkt)

	case packet.PKT_T_RequestTaskExecutionAck:
		ses.requestClose()

	case packet.PKT_T_BroadcastCommit, packet.PKT_T_SyncMsgCommit:
		// case packet.PKT_T_BroadcastCommit, packet.PKT_T_SyncCommit:
		go ses.handleStateBSCommit(h.PktType)

	case packet.PKT_T_ElevatorFailed:
		// TODO fault tolerence? what to do now ...

	}
	return nil
}

func (ses *Session) handleFirstIncomming(pkt packet.Packet) {
	h := pkt.Header

	ses.setPendingPkt(
		&packet.OutgoingMessage{
			Origin:  h.Origin,
			PktType: h.PktType,
			EMsg:    pkt.Payload,
		})

	switch h.PktType {
	case packet.PKT_T_RequestTaskExecution:
		ses.handleRequestTaskExecution(pkt.Payload.EMsgType)

	case packet.PKT_T_BroadcastUpdate:
		ses.handleStateBSUpdate()

	case packet.PKT_T_SyncMsg:
		ses.handleSyncMsg()
	}
}

func (ses *Session) handleRequestTaskExecution(eMsgType message.ElevatorMessageType) {
	ses.queueElevatorCommand(eMsgType)
	ses.queueReply(packet.PKT_T_RequestTaskExecutionAck)
	ses.startShutdownTimer()
}

// expects a response/completion from elevator
func (ses *Session) queueElevatorRequest() {
	ses.queueElevatorTask(ses.pendingPkt.EMsg, ses.elevDone)
}

// fire-and-forget, reponse will appear in another session
func (ses *Session) queueElevatorCommand(eMsgType message.ElevatorMessageType) {
	ses.notifyTaskReady()

	eMsg := ses.pendingPkt.EMsg
	eMsg.Addr = ses.peerAddr.String()
	eMsg.EMsgType = eMsgType

	ses.queueElevatorTask(eMsg, nil)
}

func (ses *Session) handleStateBSUpdate() {
	ses.queueReply(packet.PKT_T_BroadcastAck)
	ses.queueElevatorRequest()
}

func (ses *Session) handleSyncMsg() {
	ses.queueElevatorRequest()
	ses.queueReply(packet.PKT_T_SyncMsgAck)
}

func (ses *Session) handleStateBSCommit(pktType packet.PacketType) {
	ses.notifyTaskReady()
	if err := ses.waitForElevatorDoneWithReply(); err != nil {
		return
	}

	switch pktType {
	case packet.PKT_T_SyncMsgCommit:
		ses.queueSyncCompleteMsg(*ses.pendingPkt)
		// ses.queueReply(packet.PKT_T_SyncComplete) // TODO when sending this you need to set self as synced if the origin is the same as you
	case packet.PKT_T_BroadcastCommit:
		ses.queueReply(packet.PKT_T_BroadcastDone)
	}

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
