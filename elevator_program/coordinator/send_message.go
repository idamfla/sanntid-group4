package coordinator

import (
	"elevator_program/elevator"
	"elevator_program/message"
	"elevator_program/udp/packet"
	"fmt"
	"net"
)

func (c *Coordinator) sendListener(e *elevator.Elevator) {
	fmt.Println("MESSAGE SENDER STARTED")
	for msg := range e.SendToCoordinator {
		c.routeOutgoingMessage(e, msg)
	}
}

func (c *Coordinator) routeOutgoingMessage(e *elevator.Elevator, msg message.ElevatorMessage) {
	if c.Server.IsMaster() {
		c.sendAsMaster(msg)
	} else {
		c.sendAsSlave(msg)
	}
}

func (c *Coordinator) sendAsSlave(eMsg message.ElevatorMessage) {
	msgPacket := packet.PROTO_PKT_T_BroadcastUpdate

	switch eMsg.EMsgType {
	case message.EMSG_T_StatusReport:
		msgPacket = packet.PROTO_PKT_T_SlaveUpdate

	case message.EMSG_T_ButtonPress:
		msgPacket = packet.PROTO_PKT_T_SlaveUpdate

	case message.EMSG_T_TaskRequest:
		msgPacket = packet.PROTO_PKT_T_RequestTaskExecution

	case message.EMSG_T_IAmMaster: // TODO fix stuff here
		eMsg.EMsgType = message.EMSG_T_NewToChannel
		// msgPacket = packet.PROTO_PKT_T_WhoIsAlive // TODO ask ida how election works now
	}
	c.QueueMessage(nil, msgPacket, eMsg)
}

func (c *Coordinator) sendAsMaster(eMsg message.ElevatorMessage) {
	msgPacket := packet.PROTO_PKT_T_BroadcastUpdate
	var addr *net.UDPAddr

	switch eMsg.EMsgType {
	case message.EMSG_T_StatusReport:
		eMsg.EMsgType = message.EMSG_T_StatusReportBroadcast
		msgPacket = packet.PROTO_PKT_T_BroadcastUpdate
		addr = nil

	case message.EMSG_T_ButtonPress:
		eMsg.EMsgType = message.EMSG_T_TaskUpdate
		msgPacket = packet.PROTO_PKT_T_BroadcastUpdate
		addr = nil

	case message.EMSG_T_TaskRequest:
		eMsg.EMsgType = message.EMSG_T_TaskUpdate
		msgPacket = packet.PROTO_PKT_T_BroadcastUpdate
		addr = nil

	case message.EMSG_T_NewToChannel:
		eMsg.EMsgType = message.EMSG_T_SyncSystem
		msgPacket = packet.PROTO_PKT_T_Snapshot

		// udpAddr, err := udp.StringAddrToUDPAddr(eMsg.Addr)
		// if err != nil {
		// 	return
		// }
		// addr = udpAddr
	}

	c.QueueMessage(addr, msgPacket, eMsg)
}
