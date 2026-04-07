package session

import (
	"elevator_program/message"
	"elevator_program/udp/packet"
	"fmt"
)

func (ses *Session) HandleIncPkt(pkt packet.Packet) error { // TODO rename HandleIncPkt
	ses.stopShutdownTimer()

	h := pkt.Header
	_, err := ses.shouldIncrementSeq(h.Seq, true)
	if err != nil {
		return err
	}

	ses.clearLastMsg()

	switch h.PktType {
	case packet.PKT_T_Heartbeat:
		fmt.Printf("%s sent %s\n", h.SenderAddr, h.PktType) // TODO remove db, although ... heatbeat should not end up here

	case packet.PKT_T_LostConn:
		fmt.Printf("%s lost connection ...", h.SenderAddr)
		// TODO what to do now?

	case packet.PKT_T_RequestTaskExecution:
		ses.handleRequestTaskExecution(pkt.Payload.EMsgType)

	case packet.PKT_T_BroadcastUpdate:
		ses.handleStateBSUpdate()

	case packet.PKT_T_SyncMsg:
		ses.handleSyncMsg()

	case packet.PKT_T_RequestTaskExecutionAck:
		ses.requestClose()

	case packet.PKT_T_BroadcastCommit, packet.PKT_T_SyncMsgCommit:
		go ses.handleStateBSCommit(h.PktType)

	case packet.PKT_T_ElevatorFailed:
		fmt.Printf("Server %s (addr %s) reported that the elevator failed, restart?", h.Origin.Alias, h.Origin.Identifier)
		ses.requestClose()
	}

	return nil
}

func (ses *Session) handleRequestTaskExecution(eMsgType message.ElevatorMessageType) {
	ses.queueElevatorCommand(eMsgType)
	ses.queueReply(packet.PKT_T_RequestTaskExecutionAck)
	ses.startShutdownTimer()
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

	err := ses.waitForElevatorDone()
	if err != nil {
		fmt.Println(err)
		switch err {
		case ErrSessionStopped:
			fmt.Println("Session stopped")
		case ErrElevatorTimeout:
			fmt.Println("Elevator failed to commit …")
			ses.queueReply(packet.PKT_T_ElevatorFailed)
		}

	} else {
		switch pktType {
		case packet.PKT_T_SyncMsgCommit:
			ses.queueSyncCompleteMsg(ses.getPendingMsg())

		case packet.PKT_T_BroadcastCommit:
			ses.queueReply(packet.PKT_T_BroadcastDone)
		}
	}

	ses.startShutdownTimer()
}

func (ses *Session) notifyTaskReady() {
	select {
	case <-ses.stopCh():
	case ses.taskReady <- struct{}{}:
	default:
		fmt.Println("Notifications full, could not notify to elevator that task is ready")
	}
}
