package elevator

import (
	"elevator_program/elevio"
	"fmt"
)

/*
--------------------------------
Trying to use new infrastructure
--------------------------------
*/

// TODO Chat saying this MSG_T_ naming convention is very c -style and noisy in Go
type MessageType int

const (
	MSG_T_StatusReport MessageType = iota

	MSG_T_TaskCreate   // a new task is created/published
	MSG_T_TaskAssign   // a task is assigned to you
	MSG_T_TaskDelegate // a task is assigned to another person
	MSG_T_TaskUpdate   // task changed, Don't think we need it
	MSG_T_TaskComplete // task was completed
	MSG_T_TaskRequest  // someone requests a new task
	MSG_T_LostComs     // A routine to check if you have lost communication
	MSG_T_NewToChannel // Send the latest information
)

// TODO Same thing as the comment above MSG_S_
type MessageState int

const (
	MSG_S_Sent MessageState = iota
	MSG_S_Ack
	MSG_S_Commit
	MSG_S_Applied
)

type Message struct {
	msgType        MessageType
	senderId       int
	task           elevio.ButtonEvent // Elevators current target (floor, btnType) or change current target
	btnStatus      ButtonStatus       // Type what we want the button to be: nonActive, pending, active
	elevatorStatus ElevatorsStatus
	msgState       MessageState

	// TODO temp need a com number
	comNumber int

	// TODO Maybe we need target id as well

	// Used for a full sync
	fullstate *SystemState
	// msgTimer       time.Time
	// TODO how to be able to send their chan Message as well
}

/*
TODO Needs this somewhere
ctx, cancel := context.Withcancel(context.Background())
msgCh := make(chan string)
*/

// Chat don't like these function names. Don't want any underscores
func (e *Elevator) messageHandler_slave(msg Message) {
	switch msg.msgType {
	case MSG_T_StatusReport:
	// target the updated elevator in the map and add the changes
	case MSG_T_TaskUpdate: // Maybe use TaskUpdate istead, is a more general name.
		e.setRequestStatus(msg.btnStatus, msg.task)

		if msg.btnStatus == NotActive {
			e.clearHallLamp(msg.task.Floor, msg.task.Button)
		} else {
			elevio.SetButtonLamp(msg.task.Button, msg.task.Floor, true)
		}
	case MSG_T_TaskDelegate:
		// Delegating task to another elevator, everything that is not cab should be sent on TaskCreate
		e.updateRemoteCabBtn(msg.btnStatus, msg.task, msg.senderId)

	case MSG_T_LostComs:
		e.handleLostConnection(msg.senderId)

	case MSG_T_NewToChannel:
		e.initializeFromSystemState(msg.elevatorStatus, *msg.fullstate)
	}
}

// TODO chat don't like this name either
func (e *Elevator) messageHandler_master(msg Message) {
	senderId := msg.senderId
	msg.senderId = e.id // TODO is this something we want here?

	switch msg.msgType {
	case MSG_T_StatusReport:
		// Update the information about the elevator
		e.elevatorRegistry[senderId] = msg.elevatorStatus

	case MSG_T_TaskUpdate:
		sentCommitMsg := e.addNewRequestToSystem(msg.btnStatus, msg.task, msg.msgState, msg.comNumber)

		if sentCommitMsg { // If we have sent commit we can turn on lights
			if msg.btnStatus == NotActive {
				e.clearHallLamp(msg.task.Floor, msg.task.Button)
			} else {
				elevio.SetButtonLamp(msg.task.Button, msg.task.Floor, true)
			}
		}

	case MSG_T_TaskAssign:
		msg.msgState = MSG_S_Commit
		// Send commit
		// Then we need to send to everyone that, this request is now running
		// This need to be sent to everyone except the one we are sending the assignment to
		// Maybe set on a timer so know it uses to much time

	case MSG_T_TaskRequest:
		// Scan for the next request and send it back
		msg.msgType = MSG_T_TaskAssign
		msg.msgState = MSG_S_Sent

	case MSG_T_LostComs:
		// I don't know what we should do here just try to say to the slave that master hears you

	case MSG_T_NewToChannel:
		e.registerAndSyncElevator(msg.msgState, msg.elevatorStatus, *msg.fullstate)
	}
}

func (e *Elevator) messageHandler(msg Message) {
	if msg.senderId == e.id {
		return // Ignore own messages
	}

	if e.isMaster {
		e.messageHandler_master(msg)
	} else {
		if msg.msgState == MSG_S_Sent {
			// Send ack back to master/coordinator
			msg.msgState = MSG_S_Ack
			e.msgSendCh <- msg
			// Return early: we don't commit/apply yet
			return
		}

		e.messageHandler_slave(msg)
	}
}

func (e *Elevator) messageListener(msgCh chan Message) {
	fmt.Println("MESSAGE LISTENER STARTED")
	for msg := range e.msgRecieveCh {
		e.messageHandler(msg)
	}
}

func (e *Elevator) setRequestStatus(status ButtonStatus, btnEvent elevio.ButtonEvent) {
	f := btnEvent.Floor
	b := btnEvent.Button

	if b == elevio.BT_Cab {
		e.cabRequests[f] = status
	} else {
		e.hallRequests[f][b] = status
	}
}

func (e *Elevator) updateRemoteCabBtn(status ButtonStatus, btnEvent elevio.ButtonEvent, id int) {
	e.elevatorRegistry[id].cabRequests[btnEvent.Floor] = status
}

func (e *Elevator) initializeFromSystemState(self ElevatorsStatus, state SystemState) {
	e.cabRequests = self.cabRequests
	e.id = self.id
	e.isMaster = false
	e.connectedToMaster = true
	// e.ipToId Need to know the ip/id to the others
	e.hallRequests = state.hallRequests
	e.elevatorRegistry = state.Elevators
}

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

func (e *Elevator) addNewRequestToSystem(btnStatus ButtonStatus, task elevio.ButtonEvent, msgState MessageState, comNumber int) bool {
	if msgState == MSG_S_Sent {
		// Send prepare to commit for this request to every elevator
	} else {
		if e.ackArray[comNumber] == len(e.elevatorRegistry) {
			// Send commit message
			e.setRequestStatus(btnStatus, task)
			return true
		} else {
			// TODO Maybe need to check that it is a unique elevator and not the same
			e.ackArray[comNumber] += 1
		}
	}
	return false
}

func (e *Elevator) registerAndSyncElevator(msgState MessageState, targetElevator ElevatorsStatus, fullstate SystemState) {
	// TODO Should we send a init pos?
	msgState = MSG_S_Commit
	senderId, ok := e.ipToId[targetElevator.ip]
	if ok {
		targetElevator = e.elevatorRegistry[senderId]
	} else {
		// TODO Do the master have itself in the elevatorRegistery?
		senderId = len(e.elevatorRegistry) + 1
		newElevator := ElevatorsStatus{
			id: senderId,
			// Hope everything else is already configured
		}
		e.elevatorRegistry[senderId] = newElevator
		targetElevator = newElevator
	}
	fullstate.hallRequests = e.hallRequests
	for id, _ := range e.elevatorRegistry {
		if senderId == id {
			continue
		}
		// TODO We have the ip in the elevatorsStatus struct but maybe we need to send a map of them as well
		fullstate.Elevators[id] = e.elevatorRegistry[id]
	}
}

// Can't just change targetElevator, it won't affect anything outside of this function
// A better structure to giveAllInfoToElev()
// func (e *Elevator) registerAndSyncElevator(ip string) {

// 	id, exists := e.ipToId[ip]
// 	if !exists {
// 		id = e.allocateNewElevatorID()
// 		e.elevatorRegistry[id] = ElevatorsStatus{id: id}
// 		e.ipToId[ip] = id
// 	}

// 	e.sendFullStateTo(id)
// }
