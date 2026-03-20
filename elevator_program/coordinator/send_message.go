package coordinator

import (
	"elevator_program/elevator"
	"elevator_program/message"
	"elevator_program/udp"
	"elevator_program/udp/packet"
	"fmt"
)

func (c *Coordinator) sendListener(e *elevator.Elevator) {
	fmt.Println("MESSAGE SENDER STARTED")
	for msg := range e.SendToCoordinator {
		c.routeOutgoingMessage(e, msg)
	}
}

func (c *Coordinator) routeOutgoingMessage(e *elevator.Elevator, msg message.ElevatorMessage) {
	if e.IsMaster {
		c.sendAsMaster(msg)
	} else {
		c.sendAsSlave(msg)
	}
}

// slave starting the session with master or someone ...
func (c *Coordinator) sendAsSlave(eMsg message.ElevatorMessage) {
	msgPacket := packet.PROTO_PKT_T_BroadcastUpdate

	switch eMsg.EMsgType {
	case message.EMSG_T_StatusReport:
		msgPacket = packet.PROTO_PKT_T_SlaveUpdate
		fmt.Println("Trying to send status report")

	case message.EMSG_T_ButtonPress:
		msgPacket = packet.PROTO_PKT_T_SlaveUpdate
		fmt.Println("Trying to send button press")

	case message.EMSG_T_TaskRequest:
		msgPacket = packet.PROTO_PKT_T_RequestTaskExecution
		fmt.Println("Trying to send Task request 1")

	case message.EMSG_T_NewToChannel:
		msgPacket = packet.PROTO_PKT_T_WhoIsMaster
		fmt.Println("Trying to send new to channel")
	}
	c.QueueMessage(nil, msgPacket, eMsg)
}

// TODO Is this a smart way to do it, seams kind of unneccesary
func (c *Coordinator) sendAsMaster(eMsg message.ElevatorMessage) {
	msgPacket := packet.PROTO_PKT_T_BroadcastUpdate
	var ip string

	if eMsg.EMsgType == message.EMSG_T_NewToChannel {
		ip = eMsg.Addr
		msgPacket = packet.PROTO_PKT_T_Snapshot
		fmt.Println("Trying to send new to channel")

		addr, err := udp.StringAddrToUDPAddr(ip)
		if err != nil {
			fmt.Println("Could not find addr")
			return
		}

		c.QueueMessage(addr, msgPacket, eMsg)
		return
	}

	switch eMsg.EMsgType {
	case message.EMSG_T_StatusReport:
		eMsg.EMsgType = message.EMSG_T_StatusReportBroadcast
		msgPacket = packet.PROTO_PKT_T_BroadcastUpdate
		fmt.Println("Trying to send status report")
	case message.EMSG_T_ButtonPress:
		eMsg.EMsgType = message.EMSG_T_TaskUpdate
		msgPacket = packet.PROTO_PKT_T_BroadcastUpdate
		fmt.Println("Trying to send task update")
	case message.EMSG_T_TaskRequest:
		eMsg.EMsgType = message.EMSG_T_TaskUpdate
		msgPacket = packet.PROTO_PKT_T_BroadcastUpdate
		fmt.Println("Trying to send task update 2")
	}

	c.QueueMessage(nil, msgPacket, eMsg)
}
