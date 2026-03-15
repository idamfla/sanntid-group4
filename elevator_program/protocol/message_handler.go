package protocol

import (
	"elevator_program/elevator"
	"elevator_program/message"
	"elevator_program/types"
	"fmt"
)

// TODO Ida thinks it could be a better way to structure it

/*
TODO Needs this somewhere
ctx, cancel := context.Withcancel(context.Background())
msgCh := make(chan string)
*/

// Chat don't like these function names. Don't want any underscores
func (p *Protocol) slaveMessageHandler(e *elevator.Elevator, msg message.Message) {
	switch msg.MsgType {
	case types.MSG_T_StatusReport:
		e.System.SetStatusReport(msg.Id, msg.Elevators[msg.Id])

	case types.MSG_T_TaskUpdate:
		if e.Id == msg.Id && msg.BtnStatus == types.Running { // Assign new task
			e.SetRequestAsTarget(*msg.Task)
		} else { // Just update system
			e.System.SetRequestStatus(msg.Id, msg.BtnStatus, *msg.Task) // TODO Should I use my own ID or msg Id??
		}
		e.UpdateBtnLamp(msg.BtnStatus, msg.Task.Floor, msg.Task.Button)

	case types.MSG_T_LostComs:
		e.HandleLostConnection(msg.Id)

	case types.MSG_T_NewToChannel:
		e.System.InitializeFromSystemState(msg)
		e.SetConnectionState(msg)
	}
}

// TODO chat don't like this name either
func (p *Protocol) masterMessageHandler(e *elevator.Elevator, msg message.Message) {
	switch msg.MsgType {
	case types.MSG_T_StatusReport:
		e.System.SetStatusReport(msg.Id, msg.Elevators[msg.Id])
		// TODO Send broadcast of status report
		msg.MsgType = types.MSG_T_StatusReport // TODO probably don't need but nice to be safe
		p.outgoingPacket.Msg = msg
		p.msgSendCh <- p.outgoingPacket // TODO Is this right?

		// TODO Should i create a place for the slave to send button updates, and master sending the update on another chanel
		// Would create a better skille between those messages that just needs to be commited and those you need to find out if someone else is better
	case types.MSG_T_ButtonPress:
		// TODO Lets hope that we only get commit messages, or else we need to count ack
		taskElevatorId, _, _ := e.ClosestToTarget(e.System.Elevators, *msg.Task)
		if taskElevatorId != -1 {
			// Someone has a better task to do, we need to broadcast task_Update
		} // If it is not the case we just need to broadcast the change
		e.System.SetRequestStatus(msg.Id, msg.BtnStatus, *msg.Task) // TODO These shouldn't be uppdated before everyone is ready
		e.UpdateBtnLamp(msg.BtnStatus, msg.Task.Floor, msg.Task.Button)

	case types.MSG_T_TaskUpdate:
		// TODO Should just send commit and the change light if necessary

	case types.MSG_T_TaskRequest:
		// Scan for the next request and send it back
		cabRequestsTemp := msg.Elevators[msg.Id].CabRequests
		currentFloor := msg.Elevators[msg.Id].CurrentFloor
		direction := msg.Elevators[msg.Id].Direction
		task := e.ComputeNewTarget(currentFloor, cabRequestsTemp, direction)
		fmt.Println("Need to send new task: ", task)
		// Broadcast new assignment if we found a new task
		msg.Task = &task
		msg.BtnStatus = types.Running
		msg.MsgType = types.MSG_T_TaskUpdate
		p.outgoingPacket.Msg = msg
		p.msgSendCh <- p.outgoingPacket

	case types.MSG_T_LostComs:
		// I don't know what we should do here just try to say to the slave that master hears you

	case types.MSG_T_NewToChannel:
		e.System.RegisterAndSyncElevator(msg, e.IpRegistery)
		// Broadcast statusReport
		id := e.IpRegistery[msg.Ip]
		msg.MsgType = types.MSG_T_StatusReport
		msg.Id = id
		msg.Elevators[id] = e.System.Elevators[id] // TODO Could cause panic if msg.Elevators is not initialized
		p.outgoingPacket.Msg = msg
		p.msgSendCh <- p.outgoingPacket
	}
}

func (p *Protocol) MessageHandler(e *elevator.Elevator, msg message.Message) {
	if e.IsMaster {
		p.masterMessageHandler(e, msg)
	} else {
		p.slaveMessageHandler(e, msg)
	}
}

func (p *Protocol) MessageListener(e *elevator.Elevator) {
	fmt.Println("MESSAGE LISTENER STARTED")
	for pktCtx := range p.msgRecieveCh {
		fmt.Println("Wallah broren min")
		msg := pktCtx.Packet.Payload
		p.MessageHandler(e, msg)
		// pktCtx.Done <- struct{}{} // TODO Locks after the first message
	}
}

/*
------------------------------------------------------------------------------
Applying protocol functions which is ment to split between the different roles
------------------------------------------------------------------------------
*/

// func (p Protocol) applyStatusReport(e *elevator.Elevator, msg message.Message) {
// 	e.System.SetStatusReport(msg.Id, msg.Elevators[msg.Id])
// }

// func (p Protocol) applyTaskUpdate_slave(e *elevator.Elevator, msg message.Message) {
// 	e.System.SetRequestStatus(msg.Id, msg.BtnStatus, *msg.Task) // TODO Should I use my own ID or msg Id??
// 	e.UpdateBtnLamp(msg.BtnStatus, msg.Task.Floor, msg.Task.Button)
// }

// func (p Protocol) applyRemoteCabUpdate_slave(e *elevator.Elevator, msg message.Message) {
// 	e.System.UpdateRemoteCabBtn(msg.Id, msg.BtnStatus, msg.Task.Floor)
// }

// func (p Protocol) applyLostComsProtocol_slave(e *elevator.Elevator, msg message.Message) {
// 	e.HandleLostConnection(msg.Id)
// }

// func (p Protocol) applySystemSync_slave(e *elevator.Elevator, msg message.Message) {
// 	e.System.InitializeFromSystemState(msg)
// 	e.SetConnectionState(msg)
// }

// func (p *Protocol) addNewRequestToSystem_master(e *elevator.Elevator, msg message.Message) {
// 	// TODO Lets hope that we only get commit messages, or else we need to count ack
// 	taskElevatorId, _, _ := e.ClosestToTarget(e.System.Elevators, *msg.Task)
// 	if taskElevatorId != -1 {

// 	}
// 	e.System.SetRequestStatus(msg.Id, msg.BtnStatus, *msg.Task)
// 	e.UpdateBtnLamp(msg.BtnStatus, msg.Task.Floor, msg.Task.Button)
// }

// func (p *Protocol) applyRegisterAndSyncElevatorToServer(e *elevator.Elevator, msg message.Message) {
// 	e.System.RegisterAndSyncElevator(msg, e.IpRegistery)
// }
