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

func (ses *Session) queueReply(pktType packet.PacketType) {
	ses.QueueDirectMsg(pktType, message.ElevatorMessage{})
}
