package session

import (
	"elevator_program/udp/packet"
	"fmt"
)

func (ses *Session) QueueDirectMsg(pktType packet.PacketType, outMsg packet.OutgoingMessage) {
	// func (ses *Session) QueueDirectMsg(pktType packet.PacketType, eMsg message.ElevatorMessage) {
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
	// func (ses *Session) queueReply(pktType packet.PacketType, eMsg message.ElevatorMessage) {
	ses.QueueDirectMsg(pktType, ses.lastOutPkt)
	// ses.QueueDirectMsg(pktType, eMsg)
	// ses.QueueDirectMsg(pktType, message.ElevatorMessage{ID: ses.getSrvID()})
}
