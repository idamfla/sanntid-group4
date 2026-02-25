package elevator

import (
	"elevator_program/elevio"
	"elevator_program/network"
	"fmt"
)

type MessageType int

const (
	MSG_T_StatusReport MessageType = iota

	MSG_T_TaskCreate   // a new task is created/published
	MSG_T_TaskAssign   // a task is assigned to you
	MSG_T_TaskDelegate // a task is assigned to another person
	MSG_T_TaskUpdate   // task changed, Don't think we need it
	MSG_T_TaskComplete // task was completed
	MSG_T_TaskRequest  // someone requests a new task

	MSG_T_Broadcast // Probably don't need
	MSG_T_DirectMsg // Probably don't need
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
	msgType           MessageType
	id                int
	task              elevio.ButtonEvent // Elevators current target (floor, btnType) or change current target
	btnStatus         ButtonStatus       // Type what we want the button to be: nonActive, pending, active
	elevatorStatus    ElevatorsStatus
	transactionNumber int // To show which number it is in the whole msg system
	msgState          MessageState
	// msgTimer       time.Time
	// TODO how to be able to send their chan Message as well
}

/*
Needs this somewhere
ctx, cancel := context.Withcancel(context.Background())
msgCh := make(chan string)
*/

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
		e.elevatorRegistry[id].cabRequests[f] = !e.elevatorRegistry[id].cabRequests[f] // Make it the opposite of what it was

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

	}

	// msg.msgState = MSG_S_Applied
	e.msgSendCh <- msg
}

// To remember which messages we are waiting for
var openMsgThreads = make(map[int]Message) // Maybe we don't need this one
var msgNumber = 0
var isSender = make(map[int]bool)

func (e *Elevator) messageHandler_master(msg Message) {
	// switch msg.msgState {
	// case MSG_S_Sent:
	// 	switch msg.msgType {
	// 	case MSG_T_TaskRequest:
	// 		// task = compute new task for that specific elevator
	// 		// msg.chan <- Message{Type: MSG_T_TaskAssign, Task: task, msgState: MSG_S_Sent}
	// 	}
	// case MSG_S_Ack:
	// 	switch msg.msgType {
	// 	case MSG_T_StatusReport:
	// 		// commit the change
	// 		// broadcast that all should commit the change
	// 	case MSG_T_TaskCreate:
	// 		// tell all to commit the new task
	// 	case MSG_T_TaskAssign:
	// 		// send a delegate message
	// 		// mark task as being served
	// 		// give task
	// 	case MSG_T_TaskDelegate:
	// 		// tell all to mark a task as being served
	// 	// case MSG_T_TaskUpdate:
	// 	// 	// TODO not sure we need this? re-release a task is the same as making it anew
	// 	case MSG_T_TaskComplete:
	// 		// tell all to cross off task

	// 	case MSG_T_TaskRequest:
	// 		// all it takes to commit the task

	// 	}
	// }

	senderId := msg.id
	msg.id = e.id

	switch msg.msgType {
	case MSG_T_StatusReport:

	case MSG_T_TaskCreate:
		if isSender[msg.transactionNumber] {
			// Maybe have a counter of number of ack we have received for this mission
			// We have received a ack now we need them to commit
			msg.msgState = MSG_S_Commit
			for id, _ := range e.elevatorRegistry {
				network.Trancive(msg, e.ports[id], "unicom", "udp4")
			}
			// Need to delete the msg thread when the other elevators have received
		} else {
			// Need them first to be warned, when we receive ack we will send commit
			msgNumber++ // Maybe we need this
			isSender[msg.transactionNumber] = true
			msg.msgState = MSG_S_Sent
			for id, _ := range e.elevatorRegistry {
				network.Trancive(msg, e.ports[id], "unicom", "udp4")
			}
		}

	case MSG_T_TaskAssign:
		msg.msgState = MSG_S_Commit
		network.Trancive(msg, e.ports[senderId], "unicom", "udp4") // Send commit msg to target, received ack

	case MSG_T_TaskComplete:
		if isSender[msg.transactionNumber] {
			// Maybe have a counter of number of ack we have received for this mission
			// We have received a ack now we need them to commit
			msg.msgState = MSG_S_Commit
			for id, _ := range e.elevatorRegistry {
				network.Trancive(msg, e.ports[id], "unicom", "udp4")
			}
			// Need to delete the msg thread when the other elevators have received
		} else {
			msgNumber++ // Maybe we need this
			isSender[msg.transactionNumber] = true
			msg.msgState = MSG_S_Sent
			for id, _ := range e.elevatorRegistry {
				network.Trancive(msg, e.ports[id], "unicom", "udp4")
			}
		}

	case MSG_T_TaskRequest:
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
