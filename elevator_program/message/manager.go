package message

import (
	"elevator_program/elevator"
	"elevator_program/elevio"
)

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

// Uses to add a new elevator to our system
type SystemState struct {
	hallRequests [][2]elevator.ButtonStatus
	Elevators    map[int]elevator.ElevatorsStatus
}

type Message struct {
	msgType        MessageType
	senderId       int
	task           elevio.ButtonEvent    // Elevators current target (floor, btnType) or change current target
	btnStatus      elevator.ButtonStatus // Type what we want the button to be: nonActive, pending, active
	elevatorStatus elevator.ElevatorsStatus
	msgState       MessageState

	// TODO Maybe we need target id as well

	// Used for a full sync
	fullstate *SystemState
	// msgTimer       time.Time
	// TODO how to be able to send their chan Message as well
}
