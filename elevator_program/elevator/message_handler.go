package elevator

import (
	"elevator_program/elevio"
	"fmt"
)

type MessageType int

const (
	MSG_T_StatusReport MessageType = iota

	MSG_T_TaskCreate   // a new task is created/published
	MSG_T_TaskAssign   // a task is assigned to you
	MSG_T_TaskDelegate // a task is assigned to another person
	MSG_T_TaskUpdate   // task changed
	MSG_T_TaskComplete // task was completed
	MSG_T_TaskRequest  // someone requests a new task

	MSG_T_Broadcast
	MSG_T_DirectMsg
	MSG_T_LostComs
	MSG_T_NewToChannel
)

type MessageState int

const (
	MSG_S_Sent MessageState = iota
	MSG_S_Ack
	MSG_S_Commit
	MSG_S_Applied
)

type Message struct {
	msgType        MessageType
	id             int
	task           elevio.ButtonEvent // Elevators current target (floor, btnType) or change current target
	elevatorStatus ElevatorsStatus
	msgState       MessageState
	// msgTimer       time.Time
	// TODO how to be able to send their chan Message as well
}

/*
Needs this somewhere
ctx, cancel := context.Withcancel(context.Background())
msgCh := make(chan string)
*/

func (e *Elevator) messageHandler_slave(msg Message) {
	if msg.msgState != MSG_S_Commit {
		return
	}

	switch msg.msgType {
	case MSG_T_StatusReport:
		// target the updated elevator in the map and add the changes
	case MSG_T_TaskCreate:
		f := msg.task.Floor
		b := msg.task.Button

		if b == elevio.BT_Cab {
			e.cabRequests[f] = true
		} else {
			e.hallRequests[f][b] = true
		}
	case MSG_T_TaskAssign:
		e.nextTarget = msg.task
	case MSG_T_TaskDelegate:
		f := msg.task.Floor
		b := msg.task.Button

		// TODO need to have it as 2d array of custom requestState instead of just bool
		if b == elevio.BT_Cab {
			e.cabRequests[f] = true
		} else {
			e.hallRequests[f][b] = true
		}
	case MSG_T_TaskUpdate:
		// TODO not sure we need this? re-release a task is the same as making it anew
	case MSG_T_TaskComplete:
		e.clearHallRequest(msg.task.Floor, msg.task.Button)
		e.clearHallLamp(msg.task.Floor, msg.task.Button)

	case MSG_T_TaskRequest:
		// will just be replied to with a task

	}

	msg.msgState = MSG_S_Applied
	e.msgSendCh <- msg
}

func (e *Elevator) messageHandler_master(msg Message) {
	switch msg.msgState {
	case MSG_S_Sent:
		switch msg.msgType {
		case MSG_T_TaskRequest:
			// task = compute new task for that specific elevator
			// msg.chan <- Message{Type: MSG_T_TaskAssign, Task: task, msgState: MSG_S_Sent}
		}
	case MSG_S_Ack:
		switch msg.msgType {
		case MSG_T_StatusReport:
			// commit the change
			// broadcast that all should commit the change
		case MSG_T_TaskCreate:
			// tell all to commit the new task
		case MSG_T_TaskAssign:
			// send a delegate message
			// mark task as being served
			// give task
		case MSG_T_TaskDelegate:
			// tell all to mark a task as being served
		case MSG_T_TaskUpdate:
			// TODO not sure we need this? re-release a task is the same as making it anew
		case MSG_T_TaskComplete:
			// tell all to cross off task

		case MSG_T_TaskRequest:
			// all it takes to commit the task

		}
	}

	msg.msgState = MSG_S_Applied
	e.msgSendCh <- msg
}

func (e *Elevator) messageHandler(msg Message) {
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

// func (e *Elevator) messageHandler(msgCh chan Message) {
// 	for {
// 		select {
// 		case msg := <-msgCh:
// 			fmt.Println("Received msg")

// 			id_num, err := strconv.Atoi(msg.Id) // Should do this somewhere where it does not happen when sending ip
// 			if err != nil {
// 				fmt.Println("Error when converting to int")
// 				return
// 			}

// 			switch msg.MsgType {
// 			case 0:
// 				// broadcast basic
// 				// If there is a new request, need to add it to button map, turn on lights, send ack
// 			case 1:
// 				// direct msg
// 				if id_num == 0 {
// 					// Need to be sure that the overwritten target dose'nt get lost
// 					e.nextTarget = msg.NextTarget
// 					// Need to send ack to master
// 				} else {
// 					// Master need to add request to que and inform other elevators
// 				}
// 			case 2:
// 				// lost coms
// 				// Meesage your communication status with the master
// 			case 3:
// 				// New to chanel
// 				// Check if the IP has been here before, if this is the case, send back the id and button map
// 				// and all the cab calls
// 			}

// 		case <-time.After(2 * time.Second):
// 			fmt.Println("Maybe lost comunication")
// 			//Broadcast that you have lost communication, figure out how to restart yourself or other
// 		}
// 	}
// }
