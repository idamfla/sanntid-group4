package message

import "elevator_program/elevio"

/*
Needs this somewhere
ctx, cancel := context.Withcancel(context.Background())
msgCh := make(chan string)
*/

// Chat don't like these function names. Don't want any underscores
func (e *Elevator) messageHandler_slave(msg Message) {
	// if msg.msgState != MSG_S_Commit {
	// 	return
	// }

	switch msg.msgType {
	case MSG_T_StatusReport:
		// target the updated elevator in the map and add the changes

	case MSG_T_TaskCreate: // Maybe use TaskUpdate istead, is a more general name.
		f := msg.task.Floor
		b := msg.task.Button

		if b == elevio.BT_Cab {
			e.cabRequests[f] = true
		} else {
			e.hallRequests[f][b] = msg.btnStatus
		}
		elevio.SetButtonLamp(b, f, true)

	case MSG_T_TaskAssign:
		// Could be merged with TaskCreate, but maybe not smart
		f := msg.task.Floor
		b := msg.task.Button
		e.nextTarget = msg.task

		if b == elevio.BT_Cab {
			e.cabRequests[f] = true
		} else {
			e.hallRequests[f][b] = Running
		}
		elevio.SetButtonLamp(b, f, true)

	case MSG_T_TaskDelegate:
		// Delegating task to another elevator, everything that is not cab should be sent on TaskCreate
		f := msg.task.Floor
		id := msg.elevatorStatus.id
		e.elevatorRegistry[id].cabRequests[f] = true // Could not turn it to the oposite since it will break consistency

	case MSG_T_TaskComplete:
		f := msg.task.Floor
		b := msg.task.Button

		if b == elevio.BT_Cab {
			e.cabRequests[f] = false
		} else {
			e.hallRequests[f][b] = NotActive
		}
		e.clearCurrentFloor(f, b)

	case MSG_T_TaskRequest:
		// will just be replied to with a task, probably don't need this one on slave just send to TaskAssign

	case MSG_T_LostComs:
		// We need a way to know who initiatet the msg:
		// if we initiated the msg{
		if msg.senderId == e.ipToId["Master"] {
			e.connectedToMaster = true
			// Make the msg resolved
		} else {
			// Need to count if any elevator har connection to master
			// Need to count how many does not have connection
			// if you are the problem, restart or complete cab then restart
			// If you are not the problem, select a new master
		}

	case MSG_T_NewToChannel:
		transferElevator := msg.elevatorStatus
		e.cabRequests = transferElevator.cabRequests
		e.id = transferElevator.id
		e.isMaster = false
		e.connectedToMaster = true
		// e.ipToId Need to know the ip/id to the others
		e.hallRequests = msg.fullstate.hallRequests
		e.elevatorRegistry = msg.fullstate.Elevators
	}

	// msg.msgState = MSG_S_Applied
	e.msgSendCh <- msg
}
