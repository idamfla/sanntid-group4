package elevator

import (
	// "elevator_program/elevator"
	"elevator_program/elevio"
	"elevator_program/message"
	"fmt"
)

/*
--------------------------------
Trying to use new infrastructure
--------------------------------
*/

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
		cabRequestsTemp := make([]ButtonStatus, len(msg.CabRequests))
		for i, req := range msg.CabRequests {
			cabRequestsTemp[i] = ButtonStatus(req) // Explicit conversion
		}
		task := e.computeNewTarget(msg.CurrentFloor, cabRequestsTemp, msg.Direction)
		// Need to send assign to the elevator
		// And send task uppdate to the other

	case message.MSG_T_LostComs:
		// I don't know what we should do here just try to say to the slave that master hears you

	case message.MSG_T_NewToChannel:
		p.applyRegisterAndSyncElevatorToServer(e, msg)
	}
}

func (e *Elevator) MessageHandler(msg message.Message) {
	// if msg.SenderId == e.id {
	// 	return // Ignore own messages
	// }

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

func (s *System) setStatusReport(msg message.Message) {
	// Convert CabRequests from Message.CabRequests to ElevatorsStatus.CabRequests
	cabRequests := make([]ButtonStatus, len(msg.CabRequests))
	for i, req := range msg.CabRequests {
		cabRequests[i] = ButtonStatus(req) // Explicit conversion
	}

	s.Elevators[msg.Id] = ElevatorsStatus{
		Id:           msg.Id,
		Ip:           msg.Ip,
		CurrentFloor: msg.CurrentFloor,
		CabRequests:  cabRequests,
		Target:       msg.Task,
		Direction:    msg.Direction,
		State:        ElevatorState(msg.Direction),
	}
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

func (s *System) initializeFromSystemState(msg message.Message) {
	// Ensure the hallRequests slice is properly initialized
	s.hallRequests = make([][2]ButtonStatus, len(msg.HallRequests))

	// Deep copy each element from msg.HallRequests to s.hallRequests
	for i := range msg.HallRequests {
		s.hallRequests[i][0] = ButtonStatus(msg.HallRequests[i][0])
		s.hallRequests[i][1] = ButtonStatus(msg.HallRequests[i][1])
	}
	// s.Elevators = state.Elevators // TODO Fix this one
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

func (s *System) registerAndSyncElevator(e Elevator, msg message.Message) {
	// TODO Hmm this is wierd, do we even want Message to be in elevator package??
	newMessage := message.Message{}

	// TODO Should we send a init pos?
	senderId, ok := e.ipToId[msg.Ip]
	if ok {
		foundElevator := s.Elevators[senderId]

		newMessage.Id = senderId
		newMessage.Ip = msg.Ip
		newMessage.CurrentFloor = foundElevator.CurrentFloor

		cabRequests := make([]message.ButtonStatus, len(foundElevator.CabRequests))
		for i, req := range foundElevator.CabRequests {
			cabRequests[i] = message.ButtonStatus(req) // Explicit conversion
		}
		newMessage.CabRequests = cabRequests
		newMessage.Task = foundElevator.Target
		newMessage.Direction = foundElevator.Direction
		newMessage.State = message.ElevatorState(foundElevator.State)
	} else {
		// TODO Do the master have itself in the elevatorRegistery?
		senderId = len(e.elevatorRegistry) + 1
		newElevator := ElevatorsStatus{
			Id: senderId,
			// Hope everything else is already configured
		}
		s.Elevators[senderId] = newElevator
		newMessage.Id = newElevator.Id
	}
	// Ensure the hallRequests slice is properly initialized
	newMessage.HallRequests = make([][2]message.ButtonStatus, len(s.hallRequests))

	// Deep copy each element from msg.HallRequests to s.hallRequests
	for i := range s.hallRequests {
		newMessage.HallRequests[i][0] = message.ButtonStatus(s.hallRequests[i][0])
		newMessage.HallRequests[i][1] = message.ButtonStatus(s.hallRequests[i][1])
	}

	// Need to figure out what to do with Elevator map
	// for id, currentElevator := range s.Elevators {
	// 	if newMessage.senderId == id {
	// 		continue
	// 	}
	// 	// TODO We have the ip in the elevatorsStatus struct but maybe we need to send a map of them as well
	// 	newMessage.fullstate.Elevators[id] = currentElevator
	// }
}

func (e Elevator) updateBtnLamp(msg message.Message) {
	if ButtonStatus(msg.BtnStatus) == NotActive {
		e.clearHallLamp(msg.Task.Floor, msg.Task.Button)
	} else {
		elevio.SetButtonLamp(msg.Task.Button, msg.Task.Floor, true)
	}
}

func (e *Elevator) setConnectionState(msg message.Message) {
	e.id = msg.Id
	e.isMaster = false
	e.connectedToMaster = true
	e.elevatorState = ES_Idle // TODO Do I need this one here?
	// e.ipToId Need to know the ip/id to the others
}

/*
------------------------------------------------------------------------------
Applying protocol functions which is ment to split between the different roles
------------------------------------------------------------------------------
*/

func (p Protocol) applyStatusReport(e *Elevator, msg message.Message) {
	e.system.setStatusReport(msg)
}

func (p Protocol) applyTaskUpdate_slave(e *Elevator, msg message.Message) {
	e.system.setRequestStatus(ButtonStatus(msg.BtnStatus), msg.Task, e.id)
	e.updateBtnLamp(msg)
}

func (p Protocol) applyRemoteCabUpdate_slave(e *Elevator, msg message.Message) {
	e.system.updateRemoteCabBtn(ButtonStatus(msg.BtnStatus), msg.Task, msg.Id)
}

func (p Protocol) applyLostComsProtocol_slave(e *Elevator, msg message.Message) {
	e.handleLostConnection(msg.Id)
}

func (p Protocol) applySystemSync_slave(e *Elevator, msg message.Message) {
	e.system.initializeFromSystemState(msg)
	e.setConnectionState(msg)
}

func (p *Protocol) addNewRequestToSystem_master(e *Elevator, msg message.Message) {
	// TODO Lets hope that we only get commit messages, or else we need to count ack
	// TODO Maybe need to check that it is a unique elevator and not the same
	// p.ackArray[msg.comNumber] += 1
	// if p.ackArray[msg.comNumber] == (len(e.system.Elevators) - 1) {
	// Send commit message
	e.system.setRequestStatus(ButtonStatus(msg.BtnStatus), msg.Task, e.id)
	e.updateBtnLamp(msg)
	// Then we need to close the msg
	// }
}

func (p *Protocol) applyRegisterAndSyncElevatorToServer(e *Elevator, msg message.Message) {
	e.system.registerAndSyncElevator(*e, msg)
}
