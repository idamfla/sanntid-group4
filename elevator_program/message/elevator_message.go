package message

import (
	"elevator_program/elevio"
	"elevator_program/types"
)

type ElevatorMessageType int

const (
	EMSG_T_StatusReport ElevatorMessageType = iota

	EMSG_T_TaskCreate  // a new task is created/published
	EMSG_T_ButtonPress // Slave notices a new button press
	// EMSG_T_TaskAssign   // a task is assigned to you
	// EMSG_T_TaskDelegate // a task is assigned to another person
	EMSG_T_TaskUpdate   // task changed, Don't think we need it
	EMSG_T_TaskComplete // task was completed
	EMSG_T_TaskRequest  // someone requests a new task
	EMSG_T_LostComs     // A routine to check if you have lost communication
	EMSG_T_ElevatorLost // An elevator has lost coms, you need to send your connection to master status
	EMSG_T_NewToChannel // Send the latest information
)

type ElevatorMessage struct {
	EMsgType ElevatorMessageType

	ID   string
	Addr string

	ActivePeers int
	Task        elevio.ButtonEvent
	BtnStatus   types.ButtonStatus

	HallRequests [][2]types.ButtonStatus
	Elevators    map[string]types.ElevatorsStatus
}
