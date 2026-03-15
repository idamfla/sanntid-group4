package protocol

import (
	"elevator_program/elevator"
	"elevator_program/message"
	"elevator_program/types"
	"fmt"
)

func (p *Protocol) sendListener(e *elevator.Elevator) {
	fmt.Println("MESSAGE LISTENER STARTED")
	for msg := range p.msgSendCh {
		fmt.Println("Wallah moren min")
		p.MessageHandler(e, msg)
		// pktCtx.Done <- struct{}{} // TODO Locks after the first message
	}
}

func (p *Protocol) sendProtocol(e *elevator.Elevator, msg message.Message) {
	if e.IsMaster {
		p.masterMessageHandler(e, msg)
	} else {
		p.sendMessageSlave(e, msg)
	}
}

// slave starting the session with master or someone ...
func (p *Protocol) sendMessageSlave(e *elevator.Elevator, msg message.Message) {
	// var pktType packet.PacketType

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
	// pkt := p.outgoingPacket
	// pkt.Msg = msg
	// p.msgSendCh <- pkt
	// outgoingMsg := message.Message {
	// id: e.ID,
	// ...
	// }

	// TODO function is called e.QueueMessage(remoteAddr *net.UDPAddr, pktType, msg)
}

// TODO Is this a smart way to do it, seams kind of unneccesary
func (p *Protocol) sendMessageMaster(e *elevator.Elevator, msg message.Message) {
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
	// pkt := p.outgoingPacket
	// pkt.Msg = msg
	// p.msgSendCh <- pkt
}
