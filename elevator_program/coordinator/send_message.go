package coordinator

import (
	"elevator_program/elevator"
	"elevator_program/message"
	"elevator_program/types"
	"elevator_program/udp"
	"elevator_program/udp/packet"
	"fmt"
	"time"
)

func (c *Coordinator) sendListener(e *elevator.Elevator) {
	fmt.Println("MESSAGE SENDER STARTED")
	for msg := range e.SendToProtocol {
		c.routeOutgoingMessage(e, msg)
		// pktCtx.Done <- struct{}{} // TODO Locks after the first message
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
func (c *Coordinator) sendAsSlave(msg message.ElevatorMessage) {
	// var pktType packet.PacketType
	var ip string
	// broadcastIp := udp.NtnuBroadcastIP // TODO we don't allways want broadcast ip and port, need to find the others
	// port := udp.BROADCAST_PORT
	localIP := "127.0.0.1"
	port := c.portRegistery["master"]

	msgPacket := packet.PROTO_PKT_T_BroadcastUpdate

	switch msg.MsgType {
	case types.MSG_T_StatusReport:
		ip = localIP
		port = c.portRegistery["master"]
		msgPacket = packet.PROTO_PKT_T_Data //PROTO_PKT_T_SlaveReport
		fmt.Println("Trying to send status report")

	case types.MSG_T_ButtonPress:
		ip = localIP
		port = c.portRegistery["master"]
		msgPacket = packet.PROTO_PKT_T_Data //PROTO_PKT_T_SlaveReport
		fmt.Println("Trying to send button press")

	case types.MSG_T_TaskRequest:
		ip = localIP
		port = c.portRegistery["master"]
		msgPacket = packet.PROTO_PKT_T_Data //PROTO_PKT_T_RequestNewOrder
		fmt.Println("Trying to send Task request 1")
		time.Sleep(2 * time.Second)

	case types.MSG_T_LostComs:
		// msg.MsgType = types.MSG_T_ElevatorLost
		ip = localIP //broadcastIp
		port = c.portRegistery["broadcast"]
		msgPacket = packet.PROTO_PKT_T_Data //PROTO_PKT_T_LostConn
		fmt.Println("Trying to send lost coms")

	case types.MSG_T_ElevatorLost:
		// msg.MsgType = types.MSG_T_LostComs
		ip = localIP
		port = c.portRegistery["broadcast"]
		msgPacket = packet.PROTO_PKT_T_Data //PROTO_PKT_T_BroadcastUpdate
		fmt.Println("Trying to send elevator lost")

	case types.MSG_T_NewToChannel:
		ip = localIP // broadcastIp
		port = 9000  //p.portRegistery["broadcast"]
		msgPacket = packet.PROTO_PKT_T_Data
		fmt.Println("Trying to send new to channel")
	}
	c.QueueMessage(udp.MustUDPAddr(ip, port), msgPacket, msg)
}

// TODO Is this a smart way to do it, seams kind of unneccesary
func (c *Coordinator) sendAsMaster(msg message.ElevatorMessage) {
	// port := udp.BROADCAST_PORT
	var ip string
	// broadcastIp := udp.NtnuBroadcastIP
	localIP := "127.0.0.1"
	port := 9001 //p.portRegistery["broadcast"]
	msgPacket := packet.PROTO_PKT_T_BroadcastUpdate

	switch msg.MsgType {
	case types.MSG_T_StatusReport:
		msgPacket = packet.PROTO_PKT_T_Data //PROTO_PKT_T_BroadcastUpdate
		fmt.Println("Trying to send status report")
	case types.MSG_T_ButtonPress:
		msg.MsgType = types.MSG_T_TaskUpdate // Now we send slaves to update request
		msgPacket = packet.PROTO_PKT_T_Data  //PROTO_PKT_T_BroadcastUpdate
		fmt.Println("Trying to send task update")
	case types.MSG_T_TaskRequest:
		msg.MsgType = types.MSG_T_TaskUpdate
		msgPacket = packet.PROTO_PKT_T_Data //PROTO_PKT_T_BroadcastUpdate
		fmt.Println("Trying to send task update 2")
	case types.MSG_T_NewToChannel:
		msgPacket = packet.PROTO_PKT_T_Data //PROTO_PKT_T_BroadcastUpdate
		fmt.Println("Trying to send new to channel")
	}
	ip = localIP // broadcastIp
	c.QueueMessage(udp.MustUDPAddr(ip, port), msgPacket, msg)
}
