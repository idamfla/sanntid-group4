package session

import (
	"elevator_program/message"
	"elevator_program/udp/packet"
	"fmt"
)

func (ses *Session) QueueDirectMsg(pktType packet.PacketType, eMsg message.ElevatorMessage) {
	select {
	case ses.outgoingMsgCh <- outgoingMessage{
		PktType: pktType,
		EMsg:    eMsg,
	}:
	default:
		fmt.Println("Can't queue message, sessions messageQueue is full")
	}
}

func (ses *Session) QueueStateBSUpdateMsg(pktType packet.PacketType, eMsg message.ElevatorMessage) {
}

func (ses *Session) QueueWhoIsMasterMsg() {
	ses.QueueDirectMsg(packet.PKT_T_WhoIsMaster, message.ElevatorMessage{})
}

func (ses *Session) SendReply(pktType packet.PacketType) {
	ses.QueueDirectMsg(pktType, message.ElevatorMessage{})
}

func (ses *Session) sendLoop(behavior SessionBehavior) {
	defer ses.wg.Done()

	for {
		select {
		case <-ses.stop:
			return

		case outPkt := <-ses.outgoingMsgCh:
			ses.handleOutgoing(outPkt, behavior)
		}
	}
}

func (ses *Session) handleOutgoing(outPkt outgoingMessage, behavior SessionBehavior) {
	err := ses.send(outPkt)
	if err != nil {
		fmt.Printf("Session %d: send error: %v\n", ses.ID, err)
		return
	}

	if outPkt.PktType == packet.PKT_T_IAmMaster {
		ses.setSelfAsMaster(true)
		ses.setIsSynced(true)
	}

	behavior.OnSend(outPkt.PktType)
}

// for the SessionBehavior, does nothing
func (ses *Session) OnSend(pktType packet.PacketType) {}
