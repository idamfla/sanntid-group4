package protocol

import (
	"elevator_program/elevator"
	"elevator_program/message"
	"elevator_program/types"
	"fmt"
)

// TODO Ida thinks it could be a better way to structure it

/*
--------------------------------
Trying to use new infrastructure
--------------------------------
*/

/*
TODO Needs this somewhere
ctx, cancel := context.Withcancel(context.Background())
msgCh := make(chan string)
*/

// Chat don't like these function names. Don't want any underscores
func (p *Protocol) slaveMessageHandler(e *elevator.Elevator, msg message.Message) {
	switch msg.MsgType {
	case types.MSG_T_StatusReport:
		// target the updated elevator in the map and add the changes
		p.applyStatusReport(e, msg)

	case types.MSG_T_TaskAssign:
		// e.nextTarget = msg.Task // TODO you should not be able to loose this requests unless you get a new TaskAssign
		p.applyTaskUpdate_slave(e, msg)

	case types.MSG_T_TaskUpdate:
		p.applyTaskUpdate_slave(e, msg)

	case types.MSG_T_TaskDelegate:
		// Delegating task to another elevator, everything that is not cab should be sent on TaskCreate
		p.applyRemoteCabUpdate_slave(e, msg)

	case types.MSG_T_LostComs:
		p.applyLostComsProtocol_slave(e, msg)

	case types.MSG_T_NewToChannel:
		p.applySystemSync_slave(e, msg)
	}
}

// TODO chat don't like this name either
func (p *Protocol) masterMessageHandler(e *elevator.Elevator, msg message.Message) {
	switch msg.MsgType {
	case types.MSG_T_StatusReport:
		// Update the information about the elevator
		p.applyStatusReport(e, msg)

	case types.MSG_T_TaskUpdate:
		p.addNewRequestToSystem_master(e, msg)

	// case message.MSG_T_TaskAssign:
	// Then we need to send to everyone that, this request is now running
	// This need to be sent to everyone except the one we are sending the assignment to
	// Maybe set on a timer so know it uses to much time

	case types.MSG_T_TaskRequest:
		// Scan for the next request and send it back
		cabRequestsTemp := msg.Elevators[msg.Id].CabRequests
		currentFloor := msg.Elevators[msg.Id].CurrentFloor
		direction := msg.Elevators[msg.Id].Direction
		task := e.ComputeNewTarget(currentFloor, cabRequestsTemp, direction)
		fmt.Println("Need to send new task: ", task)
		// Need to send assign to the elevator
		// And send task uppdate to the other

	case types.MSG_T_LostComs:
		// I don't know what we should do here just try to say to the slave that master hears you

	case types.MSG_T_NewToChannel:
		p.applyRegisterAndSyncElevatorToServer(e, msg)
	}
}

func (p *Protocol) MessageHandler(e *elevator.Elevator, msg message.Message) {
	// if msg.SenderId == e.id {
	// 	return // Ignore own messages
	// }

	if e.IsMaster {
		p.masterMessageHandler(e, msg)
	} else {
		p.slaveMessageHandler(e, msg)
	}
}

func (p *Protocol) messageListener(e *elevator.Elevator) {
	fmt.Println("MESSAGE LISTENER STARTED")
	// for pktCtx := range e.MsgRecieveCh { // TODO fix channels and server
	// 	msg := pktCtx.Packet.Payload
	// 	p.MessageHandler(e, msg)
	// 	pktCtx.Done <- struct{}{}
	// }
}

/*
------------------------------------------------------------------------------
Applying protocol functions which is ment to split between the different roles
------------------------------------------------------------------------------
*/

func (p Protocol) applyStatusReport(e *elevator.Elevator, msg message.Message) {
	e.System.SetStatusReport(msg.Id, msg.Elevators[msg.Id])
}

func (p Protocol) applyTaskUpdate_slave(e *elevator.Elevator, msg message.Message) {
	e.System.SetRequestStatus(msg.Id, msg.BtnStatus, *msg.Task) // TODO Should I use my own ID or msg Id??
	e.UpdateBtnLamp(msg.BtnStatus, msg.Task.Floor, msg.Task.Button)
}

func (p Protocol) applyRemoteCabUpdate_slave(e *elevator.Elevator, msg message.Message) {
	e.System.UpdateRemoteCabBtn(msg.Id, msg.BtnStatus, msg.Task.Floor)
}

func (p Protocol) applyLostComsProtocol_slave(e *elevator.Elevator, msg message.Message) {
	e.HandleLostConnection(msg.Id)
}

func (p Protocol) applySystemSync_slave(e *elevator.Elevator, msg message.Message) {
	e.System.InitializeFromSystemState(msg)
	e.SetConnectionState(msg)
}

func (p *Protocol) addNewRequestToSystem_master(e *elevator.Elevator, msg message.Message) {
	// TODO Lets hope that we only get commit messages, or else we need to count ack
	// TODO Maybe need to check that it is a unique elevator and not the same
	// p.ackArray[msg.comNumber] += 1
	// if p.ackArray[msg.comNumber] == (len(e.system.Elevators) - 1) {
	// Send commit message
	e.System.SetRequestStatus(msg.Id, msg.BtnStatus, *msg.Task)
	e.UpdateBtnLamp(msg.BtnStatus, msg.Task.Floor, msg.Task.Button)
	// Then we need to close the msg
	// }
}

func (p *Protocol) applyRegisterAndSyncElevatorToServer(e *elevator.Elevator, msg message.Message) {
	e.System.RegisterAndSyncElevator(msg, e.IpRegistery)
}
