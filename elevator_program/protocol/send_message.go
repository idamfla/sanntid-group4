package protocol

import (
	"elevator_program/elevator"
	"elevator_program/message"
	"elevator_program/types"
	"elevator_program/udp"
	"elevator_program/udp/packet"
	"fmt"
)

func (p *Protocol) sendListener(e *elevator.Elevator) {
	fmt.Println("MESSAGE SENDER STARTED")
	for msg := range e.SendToProtocol {
		p.sendProtocol(e, msg)
		// pktCtx.Done <- struct{}{} // TODO Locks after the first message
	}
}

func (p *Protocol) sendProtocol(e *elevator.Elevator, msg message.Message) {
	if e.IsMaster {
		p.SendMessageMaster(msg)
	} else {
		p.SendMessageSlave(e, msg)
	}
}

// slave starting the session with master or someone ...
func (p *Protocol) SendMessageSlave(e *elevator.Elevator, msg message.Message) {
	// var pktType packet.PacketType
	var ip string
	broadcastIp := udp.NtnuBroadcastIP // TODO we don't allways want broadcast ip and port, need to find the others
	// port := udp.BROADCAST_PORT
	localIP := "127.0.0.1"
	port := p.portRegistery["master"]

	msgPacket := packet.PROTO_PKT_T_BroadcastData

	switch msg.MsgType {
	case types.MSG_T_StatusReport:
		msg.MsgType = types.MSG_T_StatusReport // TODO Don't think we actually need to define it again, but nice to be sure
		msg.Id = e.Id
		msg.Elevators[e.Id] = e.System.Elevators[e.Id]
		ip = localIP
		port = p.portRegistery["master"]
		msgPacket = packet.PROTO_PKT_T_SlaveReport

	case types.MSG_T_ButtonPress:
		msg.MsgType = types.MSG_T_ButtonPress
		ip = localIP
		port = p.portRegistery["master"]
		msgPacket = packet.PROTO_PKT_T_SlaveReport

	case types.MSG_T_TaskRequest:
		msg.MsgType = types.MSG_T_TaskRequest
		msg.Id = e.Id
		msg.Elevators[e.Id] = e.System.Elevators[e.Id]
		ip = localIP
		port = p.portRegistery["master"]
		msgPacket = packet.PROTO_PKT_T_RequestNewOrder

	case types.MSG_T_LostComs:
		msg.MsgType = types.MSG_T_ElevatorLost
		ip = broadcastIp
		port = p.portRegistery["broadcast"]
		msgPacket = packet.PROTO_PKT_T_Data //PROTO_PKT_T_LostConn

	case types.MSG_T_ElevatorLost:
		msg.MsgType = types.MSG_T_LostComs
		msg.Id = e.Id
		ip = localIP
		port = p.portRegistery["broadcast"]
		msgPacket = packet.PROTO_PKT_T_BroadcastData

	case types.MSG_T_NewToChannel:
		msg.MsgType = types.MSG_T_NewToChannel
		msg.Ip = e.Ip
		ip = broadcastIp
		port = 3000 //p.portRegistery["broadcast"]
		msgPacket = packet.PROTO_PKT_T_Data
	}
	p.QueueMessage(udp.MustUDPAddr(ip, port), msgPacket, msg)
}

// TODO Is this a smart way to do it, seams kind of unneccesary
func (p *Protocol) SendMessageMaster(msg message.Message) {
	// port := udp.BROADCAST_PORT
	var ip string
	broadcastIp := udp.NtnuBroadcastIP
	// localIP := "127.0.0.1"
	port := p.portRegistery["broadcast"]
	msgPacket := packet.PROTO_PKT_T_BroadcastData

	switch msg.MsgType {
	case types.MSG_T_StatusReport:
		msg.MsgType = types.MSG_T_StatusReport
		msgPacket = packet.PROTO_PKT_T_BroadcastData
	case types.MSG_T_ButtonPress:
		msg.MsgType = types.MSG_T_TaskUpdate // Now we send slaves to update request
		msgPacket = packet.PROTO_PKT_T_BroadcastData
	case types.MSG_T_TaskRequest:
		msg.MsgType = types.MSG_T_TaskUpdate
		msgPacket = packet.PROTO_PKT_T_BroadcastData
	case types.MSG_T_NewToChannel:
		msg.MsgType = types.MSG_T_NewToChannel
		msgPacket = packet.PROTO_PKT_T_Data
	}
	ip = broadcastIp
	p.QueueMessage(udp.MustUDPAddr(ip, port), msgPacket, msg)
}
