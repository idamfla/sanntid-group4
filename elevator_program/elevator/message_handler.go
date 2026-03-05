package elevator

import (
	// "elevator_program/elevator"
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

type Message struct { // TODO Have to make everyone big letters for testing maybe change back
	msgType        MessageType
	senderId       int
	task           elevio.ButtonEvent // Elevators current target (floor, btnType) or change current target
	btnStatus      ButtonStatus       // Type what we want the button to be: nonActive, pending, active
	elevatorStatus ElevatorsStatus

	// TODO temp need a com number
	comNumber int

	// TODO Maybe we need target id as well

	// Used for a full sync
	fullstate *System
	// msgTimer       time.Time
	// TODO how to be able to send their chan Message as well
}

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
func (p *Protocol) messageHandler_slave(e *Elevator, msg Message) {
	switch msg.msgType {
	case MSG_T_StatusReport:
		// target the updated elevator in the map and add the changes
		p.applyStatusReport(e, msg)

	case MSG_T_TaskUpdate:
		p.applyTaskUpdate_slave(e, msg)

	case MSG_T_TaskDelegate:
		// Delegating task to another elevator, everything that is not cab should be sent on TaskCreate
		p.applyRemoteCabUpdate_slave(e, msg)

	case MSG_T_LostComs:
		p.applyLostComsProtocol_slave(e, msg)

	case MSG_T_NewToChannel:
		p.applySystemSync_slave(e, msg)
	}
}

// TODO chat don't like this name either
func (p *Protocol) messageHandler_master(e *Elevator, msg Message) {
	switch msg.msgType {
	case MSG_T_StatusReport:
		// Update the information about the elevator
		p.applyStatusReport(e, msg)

	case MSG_T_TaskUpdate:
		p.addNewRequestToSystem_master(e, msg)

	case MSG_T_TaskAssign:
		// Then we need to send to everyone that, this request is now running
		// This need to be sent to everyone except the one we are sending the assignment to
		// Maybe set on a timer so know it uses to much time

	case MSG_T_TaskRequest:
		// Scan for the next request and send it back
		msg.msgType = MSG_T_TaskAssign

	case MSG_T_LostComs:
		// I don't know what we should do here just try to say to the slave that master hears you

	case MSG_T_NewToChannel:
		p.applyRegisterAndSyncElevatorToServer(e, msg)
	}
}

func (e *Elevator) MessageHandler(msg Message) {
	if msg.senderId == e.id {
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
	for msg := range e.msgRecieveCh {
		e.MessageHandler(msg)
	}
}

func (s *System) setStatusReport(senderId int, targetElevator ElevatorsStatus) {
	s.Elevators[senderId] = targetElevator
}

func (s *System) setRequestStatus(status ButtonStatus, btnEvent elevio.ButtonEvent, id int) {
	f := btnEvent.Floor
	b := btnEvent.Button

	if b == elevio.BT_Cab {
		s.Elevators[id].cabRequests[f] = status
	} else {
		s.hallRequests[f][b] = status
	}
}

func (s *System) updateRemoteCabBtn(status ButtonStatus, btnEvent elevio.ButtonEvent, id int) {
	s.Elevators[id].cabRequests[btnEvent.Floor] = status
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
	newMessage := Message{
		senderId: e.id,
	}
	// TODO Should we send a init pos?
	senderId, ok := e.ipToId[targetElevator.ip]
	if ok {

		newMessage.elevatorStatus = s.Elevators[senderId]
	} else {
		// TODO Do the master have itself in the elevatorRegistery?
		senderId = len(e.elevatorRegistry) + 1
		newElevator := ElevatorsStatus{
			id: senderId,
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

func (e Elevator) updateBtnLamp(msg Message) {
	if msg.btnStatus == NotActive {
		e.clearHallLamp(msg.task.Floor, msg.task.Button)
	} else {
		elevio.SetButtonLamp(msg.task.Button, msg.task.Floor, true)
	}
}

func (e *Elevator) setConnectionState(self ElevatorsStatus) {
	e.id = self.id
	e.isMaster = false
	e.connectedToMaster = true
	// e.ipToId Need to know the ip/id to the others
}

/*
------------------------------------------------------------------------------
Applying protocol functions which is ment to split between the different roles
------------------------------------------------------------------------------
*/

func (p Protocol) applyStatusReport(e *Elevator, msg Message) {
	e.system.setStatusReport(msg.senderId, msg.elevatorStatus)
}

func (p Protocol) applyTaskUpdate_slave(e *Elevator, msg Message) {
	e.system.setRequestStatus(msg.btnStatus, msg.task, e.id)
	e.updateBtnLamp(msg)
}

func (p Protocol) applyRemoteCabUpdate_slave(e *Elevator, msg Message) {
	e.system.updateRemoteCabBtn(msg.btnStatus, msg.task, e.id)
}

func (p Protocol) applyLostComsProtocol_slave(e *Elevator, msg Message) {
	e.handleLostConnection(msg.senderId)
}

func (p Protocol) applySystemSync_slave(e *Elevator, msg Message) {
	e.system.initializeFromSystemState(*msg.fullstate)
	e.setConnectionState(msg.elevatorStatus)
}

func (p *Protocol) addNewRequestToSystem_master(e *Elevator, msg Message) {
	if p.ackArray[msg.comNumber] == len(e.elevatorRegistry) { // TODO is this the right length??
		// Send commit message
		e.system.setRequestStatus(msg.btnStatus, msg.task, e.id)
		e.updateBtnLamp(msg)
	} else {
		// TODO Maybe need to check that it is a unique elevator and not the same
		p.ackArray[msg.comNumber] += 1
	}
}

func (p *Protocol) applyRegisterAndSyncElevatorToServer(e *Elevator, msg Message) {
	e.system.registerAndSyncElevator(*e, msg.elevatorStatus)
}

func (e Elevator) InitMsg() Message {
	msg := Message{
		msgType:  MSG_T_TaskUpdate,
		senderId: 2,
		task: elevio.ButtonEvent{
			Floor:  2,
			Button: elevio.BT_HallUp,
		},
		btnStatus: Pending,
	}
	return msg
}
func TestMsgHandler(numFloors int) {
	// Create system
	system := &System{
		hallRequests: make([][2]ButtonStatus, numFloors),
		Elevators:    make(map[int]ElevatorsStatus),
	}

	// Create protocol
	protocol := &Protocol{
		ackArray: make(map[int]int),
	}

	// Create elevator
	e := &Elevator{
		id:           1,
		isMaster:     false,
		system:       *system,
		protocol:     protocol,
		msgRecieveCh: make(chan Message, 10),
	}

	// Fake elevator status
	system.Elevators[1] = ElevatorsStatus{
		id:          1,
		cabRequests: make([]ButtonStatus, numFloors),
	}
	system.Elevators[2] = ElevatorsStatus{
		id:          2,
		cabRequests: make([]ButtonStatus, numFloors),
	}

	fmt.Println("Initial system state:")
	fmt.Println(system)

	// Create test message
	msg := e.InitMsg()

	fmt.Println("\nSending message:", msg.msgType)

	e.MessageHandler(msg)

	fmt.Println("\nSystem state after message:")
	fmt.Println(system)
}
