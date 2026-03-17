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
		fmt.Println("Wallah moren min")
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
	// ip := udp.NtnuBroadcastIP // TODO we don't allways want broadcast ip and port, need to find the others
	// port := udp.BROADCAST_PORT
	localIP := "127.0.0.1"
	port := 9000

	if e.Id == "1" {
		port = 9001
	}

	switch msg.MsgType {
	case types.MSG_T_StatusReport:
		msg.MsgType = types.MSG_T_StatusReport // TODO Don't think we actually need to define it again, but nice to be sure
		msg.Id = e.Id
		msg.Elevators[e.Id] = e.System.Elevators[e.Id]

	case types.MSG_T_ButtonPress:
		msg.MsgType = types.MSG_T_ButtonPress

	case types.MSG_T_TaskRequest:
		msg.MsgType = types.MSG_T_TaskRequest
		msg.Id = e.Id
		msg.Elevators[e.Id] = e.System.Elevators[e.Id]

	case types.MSG_T_ElevatorLost:
		msg.MsgType = types.MSG_T_ElevatorLost
		msg.Id = e.Id

	case types.MSG_T_NewToChannel:
		msg.MsgType = types.MSG_T_NewToChannel
		msg.Ip = e.Ip
	}
	p.QueueMessage(udp.MustUDPAddr(localIP, port), packet.PROTO_PKT_T_Data, msg)
}

// TODO Is this a smart way to do it, seams kind of unneccesary
func (p *Protocol) SendMessageMaster(msg message.Message) {
	// ip := udp.NtnuBroadcastIP
	// port := udp.BROADCAST_PORT
	localIP := "127.0.0.1"
	port := 9001

	switch msg.MsgType {
	case types.MSG_T_StatusReport:
		msg.MsgType = types.MSG_T_StatusReport
	case types.MSG_T_ButtonPress:
		msg.MsgType = types.MSG_T_TaskUpdate // Now we send slaves to update request
	case types.MSG_T_TaskRequest:
		msg.MsgType = types.MSG_T_TaskUpdate
	case types.MSG_T_NewToChannel:
		msg.MsgType = types.MSG_T_NewToChannel
	}
	p.QueueMessage(udp.MustUDPAddr(localIP, port), packet.PROTO_PKT_T_Data, msg)
}
