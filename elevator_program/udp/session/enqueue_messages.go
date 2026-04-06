package session

import (
	"elevator_program/udp/packet"
	"fmt"
)

func (ses *Session) QueueDirectMsg(pktType packet.PacketType, outMsg packet.OutgoingMessage) {
	select {
	case ses.outgoingMsgCh <- packet.OutgoingMessage{
		Origin:  outMsg.Origin,
		PktType: pktType,
		EMsg:    outMsg.EMsg,
	}:
	default:
		fmt.Println("Can't queue message, sessions messageQueue is full")
	}
}

func (ses *Session) queueReply(pktType packet.PacketType) {
	ses.QueueDirectMsg(pktType, ses.lastOutMsg)
}
