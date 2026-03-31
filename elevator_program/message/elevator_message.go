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

// DTOs, data transfer objects

// Hopefully a better struct
type ElevatorMessage struct {
	EMsgType ElevatorMessageType

	ID   string
	Addr string

	// Elevator state reporting
	// Status *types.ElevatorsStatus // TODO Don't think we need this one, only need to use Elevator map
	ActivePeers int
	// Task / button updates
	Task      elevio.ButtonEvent // TODO do we want it as a pointer? Gives us the option to not send Task on every message
	BtnStatus types.ButtonStatus

	// System synchronization
	HallRequests [][2]types.ButtonStatus
	Elevators    map[string]types.ElevatorsStatus
}

func (eMsgType ElevatorMessageType) String() string {
	switch eMsgType {
	case EMSG_T_StatusReport:
		return "Status Report"
	case EMSG_T_TaskCreate:
		return "Task Create"
	case EMSG_T_ButtonPress:
		return "Button Press"
	// EMSG_T_TaskAssign   // a task is assigned to you
	// EMSG_T_TaskDelegate // a task is assigned to another person
	case EMSG_T_TaskUpdate:
		return "Task Update"
	case EMSG_T_TaskComplete:
		return "Task Complete"
	case EMSG_T_TaskRequest:
		return "Task Request"
	case EMSG_T_LostComs:
		return "Lost Conn" // "Coms" or "Conn"?
	case EMSG_T_ElevatorLost:
		return "Elevator Lost"
	case EMSG_T_NewToChannel:
		return "New to Channel"
	default:
		return "unknown"
	}
}
