package elevator

import (
	// "elevator_program/elevator"
	"elevator_program/elevio"
	"elevator_program/udp/message"
	"fmt"
)

/*
--------------------------------
Trying to use new infrastructure
--------------------------------
*/

// // TODO Chat saying this MSG_T_ naming convention is very c -style and noisy in Go
// type MessageType int

// const (
// 	MSG_T_StatusReport MessageType = iota

// 	MSG_T_TaskCreate   // a new task is created/published
// 	MSG_T_TaskAssign   // a task is assigned to you
// 	MSG_T_TaskDelegate // a task is assigned to another person
// 	MSG_T_TaskUpdate   // task changed, Don't think we need it
// 	MSG_T_TaskComplete // task was completed
// 	MSG_T_TaskRequest  // someone requests a new task
// 	MSG_T_LostComs     // A routine to check if you have lost communication
// 	MSG_T_NewToChannel // Send the latest information
// )

type Protocol struct {
	// TODO temp need a place to put the ack
	ackArray map[int]int
}

/*
TODO Needs this somewhere
ctx, cancel := context.Withcancel(context.Background())
msgCh := make(chan string)
*/

// Chat don't like these function names. Don't want any underscores
func (p *Protocol) messageHandler_slave(e *Elevator, msg message.Message) {
	switch msg.MsgType {
	case message.MSG_T_StatusReport:
		// target the updated elevator in the map and add the changes
		p.applyStatusReport(e, msg)

	case message.MSG_T_TaskAssign:
		e.nextTarget = msg.Task // TODO you should not be able to loose this requests unless you get a new TaskAssign
		p.applyTaskUpdate_slave(e, msg)

	case message.MSG_T_TaskUpdate:
		p.applyTaskUpdate_slave(e, msg)

	case message.MSG_T_TaskDelegate:
		// Delegating task to another elevator, everything that is not cab should be sent on TaskCreate
		p.applyRemoteCabUpdate_slave(e, msg)

	case message.MSG_T_LostComs:
		p.applyLostComsProtocol_slave(e, msg)

	case message.MSG_T_NewToChannel:
		p.applySystemSync_slave(e, msg)
	}
}

// TODO chat don't like this name either
func (p *Protocol) messageHandler_master(e *Elevator, msg message.Message) {
	switch msg.MsgType {
	case message.MSG_T_StatusReport:
		// Update the information about the elevator
		p.applyStatusReport(e, msg)

	case message.MSG_T_TaskUpdate:
		p.addNewRequestToSystem_master(e, msg)

	case message.MSG_T_TaskAssign:
		// Then we need to send to everyone that, this request is now running
		// This need to be sent to everyone except the one we are sending the assignment to
		// Maybe set on a timer so know it uses to much time

	case message.MSG_T_TaskRequest:
		// Scan for the next request and send it back
		// msg.msgType = MSG_T_TaskAssign

	case message.MSG_T_LostComs:
		// I don't know what we should do here just try to say to the slave that master hears you

	case message.MSG_T_NewToChannel:
		p.applyRegisterAndSyncElevatorToServer(e, msg)
	}
}

func (e *Elevator) MessageHandler(msg message.Message) {
	if msg.SenderId == e.id {
		return // Ignore own messages
	}

	if e.isMaster {
		e.protocol.messageHandler_master(e, msg)
	} else {
		e.protocol.messageHandler_slave(e, msg)
	}
}

func (e *Elevator) messageListener() { // TODO Maybe need msgCh chan Message
	fmt.Println("MESSAGE LISTENER STARTED")
	for pktCtx := range e.msgRecieveCh {
		msg := pktCtx.Packet.Payload
		e.MessageHandler(msg)
		close(pktCtx.Done)
	}
}

func (s *System) setStatusReport(senderId int, targetElevator ElevatorsStatus) {
	s.Elevators[senderId] = targetElevator
}

func (s *System) setRequestStatus(status ButtonStatus, btnEvent elevio.ButtonEvent, id int) {
	f := btnEvent.Floor
	b := btnEvent.Button
	if b == elevio.BT_Cab {
		s.Elevators[id].CabRequests[f] = status
	} else {
		s.hallRequests[f][b] = status
	}
}

func (s *System) updateRemoteCabBtn(status ButtonStatus, btnEvent elevio.ButtonEvent, id int) {
	s.Elevators[id].CabRequests[btnEvent.Floor] = status
}

func (s *System) initializeFromSystemState(state System) {
	s.hallRequests = state.hallRequests
	s.Elevators = state.Elevators
}

// TODO This function is wierd, either we need to have it as e or something else if it is msg sending
func (e *Elevator) handleLostConnection(senderId int) {
	// We need a way to know who initiatet the msg:
	// if we initiated the msg{
	if senderId == e.ipToId["Master"] {
		e.connectedToMaster = true
		// Make the msg resolved
	} else {
		// Need to count if any elevator har connection to master
		// Need to count how many does not have connection
		// if you are the problem, restart or complete cab then restart
		// If you are not the problem, select a new master
	}
}

func (s *System) registerAndSyncElevator(e Elevator, targetElevator ElevatorsStatus) {
	// TODO Hmm this is wierd, do we even want Message to be in elevator package??
	newMessage := message.Message{
		senderId: e.id,
	}
	// TODO Should we send a init pos?
	senderId, ok := e.ipToId[targetElevator.Ip]
	if ok {

		newMessage.elevatorStatus = s.Elevators[senderId]
	} else {
		// TODO Do the master have itself in the elevatorRegistery?
		senderId = len(e.elevatorRegistry) + 1
		newElevator := ElevatorsStatus{
			Id: senderId,
			// Hope everything else is already configured
		}
		s.Elevators[senderId] = newElevator
		newMessage.elevatorStatus = newElevator
	}
	s.hallRequests = e.hallRequests
	for id, currentElevator := range s.Elevators {
		if newMessage.senderId == id {
			continue
		}
		// TODO We have the ip in the elevatorsStatus struct but maybe we need to send a map of them as well
		newMessage.fullstate.Elevators[id] = currentElevator
	}
}

func (e Elevator) updateBtnLamp(msg message.Message) {
	if msg.btnStatus == NotActive {
		e.clearHallLamp(msg.task.Floor, msg.task.Button)
	} else {
		elevio.SetButtonLamp(msg.task.Button, msg.task.Floor, true)
	}
}

func (e *Elevator) setConnectionState(self ElevatorsStatus) {
	e.id = self.Id
	e.isMaster = false
	e.connectedToMaster = true
	e.elevatorState = ES_Idle
	// e.ipToId Need to know the ip/id to the others
}

/*
------------------------------------------------------------------------------
Applying protocol functions which is ment to split between the different roles
------------------------------------------------------------------------------
*/

func (p Protocol) applyStatusReport(e *Elevator, msg message.Message) {
	e.system.setStatusReport(msg.senderId, msg.elevatorStatus)
}

func (p Protocol) applyTaskUpdate_slave(e *Elevator, msg message.Message) {
	e.system.setRequestStatus(msg.btnStatus, msg.task, e.id)
	e.updateBtnLamp(msg)
}

func (p Protocol) applyRemoteCabUpdate_slave(e *Elevator, msg message.Message) {
	e.system.updateRemoteCabBtn(msg.btnStatus, msg.task, msg.idToElevatorMission)
}

func (p Protocol) applyLostComsProtocol_slave(e *Elevator, msg message.Message) {
	e.handleLostConnection(msg.senderId)
}

func (p Protocol) applySystemSync_slave(e *Elevator, msg message.Message) {
	e.system.initializeFromSystemState(*msg.fullstate)
	e.setConnectionState(msg.elevatorStatus)
}

func (p *Protocol) addNewRequestToSystem_master(e *Elevator, msg message.Message) {
	// TODO Maybe need to check that it is a unique elevator and not the same
	p.ackArray[msg.comNumber] += 1
	if p.ackArray[msg.comNumber] == (len(e.system.Elevators) - 1) {
		// Send commit message
		e.system.setRequestStatus(msg.btnStatus, msg.task, e.id)
		e.updateBtnLamp(msg)
		// Then we need to close the msg
	}
}

func (p *Protocol) applyRegisterAndSyncElevatorToServer(e *Elevator, msg message.Message) {
	e.system.registerAndSyncElevator(*e, msg.elevatorStatus)
}
